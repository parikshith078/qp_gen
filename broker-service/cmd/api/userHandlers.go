package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (app *Config) GetUserByID(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userid")
	// Convert string to UUID
	userID, err := uuid.Parse(userIDStr)
	print(userID.String(), userIDStr)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	user, err := app.Db.GetUserByID(ctx, userID)
	if err != nil {
		// Check if it's a "no rows" error
		if err.Error() == "no rows in result set" {
			app.logger.Printf("User not found: %s", userID)
			app.errorJSON(w, fmt.Errorf("user not found"), http.StatusNotFound)
			return
		}
		// Log the error for debugging
		app.logger.Printf("Error fetching user %s: %v", userID, err)
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	app.writeJSON(w, http.StatusOK, jsonReponse{
		Error:   false,
		Data:    user,
		Message: "success",
	})
}
