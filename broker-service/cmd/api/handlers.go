package main

import (
	"net/http"
)

func (app *Config) HealthCheck(w http.ResponseWriter, r *http.Request) {
	payload := jsonReponse{
		Error:   false,
		Message: "Hit the broker again! pow",
	}
	app.writeJSON(w, http.StatusOK, payload)
}
