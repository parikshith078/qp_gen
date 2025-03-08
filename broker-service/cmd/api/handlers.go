package main

import (
	"log"
	"net/http"
)

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonReponse{
		Error:   false,
		Message: "Hit the broker again! pow",
	}
	app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := app.Db.ListUsers(r.Context())
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	payload := jsonReponse{
		Error:   false,
		Message: "success",
		Data:    users,
	}
	app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) CreateUser(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Email string `json:"email"`
	}{}
	err := app.readJSON(w, r, &data)
	log.Print(data)
	if err != nil {
		return
	}
	users, err := app.Db.CreateUser(r.Context(), data.Email)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	payload := jsonReponse{
		Error:   false,
		Message: "User created successfully!",
		Data:    users,
	}
	app.writeJSON(w, http.StatusCreated, payload)
}
