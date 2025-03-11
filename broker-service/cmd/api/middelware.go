package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

type contextKey string

const userIDKey contextKey = "user_id"

// TODO: Auth middleware
func (app *Config) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session token from cookie
		st, err := r.Cookie("session_token")
		if err != nil || st.Value == "" {
			app.errorJSON(w, ErrInvalidSession, http.StatusUnauthorized)
			return
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		// Validate session token with db
		sessionToken, err := app.Db.GetSessionTokenByToken(ctx, st.Value)
		if err != nil {
			app.errorJSON(w, ErrInvalidSession, http.StatusUnauthorized)
			return
		}

		// Check if session token is expired
		if time.Now().After(sessionToken.ExpiresAt) {
			app.errorJSON(w, ErrSessionExpired, http.StatusUnauthorized)
			return
		}

		// Get and validate CSRF token
		ct := r.Header.Get("X-CSRF-Token")
		if ct == "" {
			app.errorJSON(w, ErrMissingCSRF, http.StatusUnauthorized)
			return
		}

		csrfToken, err := app.Db.GetCsrfTokenBySessionID(ctx, sessionToken.ID)
		if err != nil {
			app.errorJSON(w, ErrInvalidSession, http.StatusUnauthorized)
			return
		}

		// Check if CSRF token is expired
		if time.Now().After(csrfToken.ExpiresAt) {
			app.errorJSON(w, ErrSessionExpired, http.StatusUnauthorized)
			return
		}

		// Validate CSRF token match
		if csrfToken.Token != ct {
			app.errorJSON(w, ErrInvalidCSRF, http.StatusUnauthorized)
			return
		}

		// Add user ID to request context for downstream handlers
		// Use the timeout context we created earlier as the parent
		log.Print("userid in middleware: ", sessionToken.UserID)
		r = r.WithContext(context.WithValue(ctx, userIDKey, sessionToken.UserID))

		// Update last activity timestamp
		go func() {
			updateCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = app.Db.UpdateUserLastActivity(updateCtx, sessionToken.UserID)
		}()

		next.ServeHTTP(w, r)
	})
}
