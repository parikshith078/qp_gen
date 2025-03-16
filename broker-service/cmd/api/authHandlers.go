package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/parikshith078/qp_gen/broker/internal/db/sqlc"
)

func (app *Config) Register(w http.ResponseWriter, r *http.Request) {
	// TODO: update validation tags for password & username
	reqBody := struct {
		Name     string `json:"name" validate:"required,min=2,max=100"`
		Email    string `json:"email" validate:"required,email,max=255"`
		Username string `json:"username" validate:"required,alphanum,min=3,max=30"`
		Password string `json:"password" validate:"required,min=8"`
	}{}

	// Decode & validate request body
	err := app.readJSON(w, r, &reqBody)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// Hash the password
	passwordHash, err := HashPassword(reqBody.Password)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	// Create a context with timeout for database operations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if email already exists
	_, err = app.Db.GetUserByEmail(ctx, reqBody.Email)
	if err == nil {
		app.errorJSON(w, fmt.Errorf("user with email %s already exists", reqBody.Email), http.StatusBadRequest)
		return
	}

	// Check if username already exists
	_, err = app.Db.GetUserByUsername(ctx, reqBody.Username)
	if err == nil {
		app.errorJSON(w, fmt.Errorf("user with username %s already exists", reqBody.Username), http.StatusBadRequest)
		return
	}

	// Create new user
	user, err := app.Db.CreateUser(ctx, sqlc.CreateUserParams{
		Name:         reqBody.Name,
		Email:        reqBody.Email,
		Username:     reqBody.Username,
		PasswordHash: passwordHash,
	})
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	payload := jsonReponse{
		Error:   false,
		Message: fmt.Sprintf("User created successfully with id %v", user.ID),
		Data:    user,
	}
	app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) Login(w http.ResponseWriter, r *http.Request) {
	reqBody := struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}{}

	err := app.readJSON(w, r, &reqBody)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// Create a new context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// get user by email
	user, err := app.Db.GetUserByEmail(ctx, reqBody.Email)
	if err != nil {
		app.errorJSON(w, ErrInvalidCredentials, http.StatusUnauthorized)
		return
	}

	// validate password
	err = VerifyPassword(reqBody.Password, user.PasswordHash)
	if err != nil {
		app.errorJSON(w, ErrInvalidCredentials, http.StatusUnauthorized)
		return
	}

	// Generate tokens
	sessionToken, err := GenerateSessionToken()
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	csrfToken, err := GenerateSessionToken()
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	experiesTime := time.Now().Add(7 * 24 * time.Hour) // expires after one week

	// Set cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  experiesTime,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		Expires:  experiesTime,
		HttpOnly: false, // need to be accessible to client-side
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	// Create a new context for database operations
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer dbCancel()

	// store session token in db
	sessionDoc, err := app.Db.CreateSessionToken(dbCtx, sqlc.CreateSessionTokenParams{
		UserID:    user.ID,
		Token:     sessionToken,
		ExpiresAt: experiesTime,
	})
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	// Create CSRF token with a new context
	csrfCtx, csrfCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer csrfCancel()

	_, err = app.Db.CreateCsrfToken(csrfCtx, sqlc.CreateCsrfTokenParams{
		SessionID: sessionDoc.ID,
		Token:     csrfToken,
		ExpiresAt: experiesTime,
	})
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	// Update user's last activity in a separate goroutine
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel() // Ensure context is canceled to prevent resource leaks

		err := app.Db.UpdateUserLastActivity(ctx, user.ID)
		if err != nil {
			// Log the error but don't affect the main request flow
			app.logger.Printf("Failed to update last activity for user %s: %v", user.ID, err)
		}
	}()

	// return success message
	app.writeJSON(w, http.StatusOK, jsonReponse{
		Error:   false,
		Message: "Login successful",
		Data:    user,
	})
}

// TODO: logout
func (app *Config) Logout(w http.ResponseWriter, r *http.Request) {
	st, err := r.Cookie("session_token")
	if err != nil || st.Value == "" {
		app.errorJSON(w, ErrInvalidSession, http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// no need to delete csrf as it will cascade
	err = app.Db.DeleteSessionToken(ctx, st.Value)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		app.errorJSON(w, fmt.Errorf("failed to logout user"), http.StatusInternalServerError)
		return
	}

	// Clear cookies by setting them to expire immediately
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: false,
	})

	app.writeJSON(w, http.StatusOK, jsonReponse{
		Error:   false,
		Message: "logout successful",
	})
}
