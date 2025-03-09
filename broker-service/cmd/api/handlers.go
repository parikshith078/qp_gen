package main

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/parikshith078/qp_gen/broker/internal/db/sqlc"
)

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonReponse{
		Error:   false,
		Message: "Hit the broker again! pow",
	}
	app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) CreateUser(w http.ResponseWriter, r *http.Request) {
	data := sqlc.CreateUserParams{
		Name:         pgtype.Text{String: "dat", Valid: true},
		Email:        "da",
		Username:     "cdd",
		PasswordHash: "tewww",
	}
	// err := app.readJSON(w, r, &data)
	// log.Print(data)
	// if err != nil {
	// 	return
	// }
	user, err := app.Db.CreateUser(r.Context(), data)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	payload := jsonReponse{
		Error:   false,
		Message: "User created successfully!",
		Data:    user,
	}
	app.writeJSON(w, http.StatusCreated, payload)
}
