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

// TODO: logout
