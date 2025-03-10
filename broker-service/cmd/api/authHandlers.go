package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/parikshith078/qp_gen/broker/internal/db/sqlc"
)

// TODO: register
func (app *Config) Register(w http.ResponseWriter, r *http.Request) {
	// TODO: update validation tags for password & username
	reqBody := struct {
		Name     string `json:"name" validate:"required"`
		Email    string `json:"email" validate:"required,email"`
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
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
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
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

// TODO: login
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
	// get user by email
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	user, err := app.Db.GetUserByEmail(ctx, reqBody.Email)
	if err != nil {
		app.errorJSON(w, InvalidCredsError, http.StatusUnauthorized)
		return
	}
	// validate password
	err = VerifyPassword(reqBody.Password, user.PasswordHash)
	if err != nil {
		app.errorJSON(w, InvalidCredsError, http.StatusUnauthorized)
		return
	}

	// set cooke
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
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  experiesTime,
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		Expires:  experiesTime,
		HttpOnly: false, // need to accessable to client-side
	})

	// store session token in db
	sessionDoc, err := app.Db.CreateSessionToken(ctx, sqlc.CreateSessionTokenParams{
		UserID: user.ID, Token: sessionToken, ExpiresAt: experiesTime,
	})
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	_, err = app.Db.CreateCsrfToken(ctx, sqlc.CreateCsrfTokenParams{
		SessionID: sessionDoc.ID, Token: csrfToken, ExpiresAt: experiesTime,
	})
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	// return success message
	app.writeJSON(w, http.StatusOK, jsonReponse{
		Error:   false,
		Message: "Login successful",
		Data:    user,
	})
}

// TODO: logout
func (app *Config) Logout(w http.ResponseWriter, r *http.Request) {

}
