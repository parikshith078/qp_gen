package main

import (
	"encoding/json"
	"net/http"
)

type jsonReponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonReponse{
		Error:   false,
		Message: "Hit the broker",
	}
	out, _ := json.MarshalIndent(payload, "", "\t")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(out)
}
