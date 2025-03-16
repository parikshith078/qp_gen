package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type jsonReponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (app *Config) readJSON(_ http.ResponseWriter, r *http.Request, data any) error {
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		switch {
		case err == io.EOF:
			return errors.New("request body is empty")
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("request body contains invalid JSON")
		default:
			return fmt.Errorf("error parsing JSON: %v", err)
		}
	}

	err = validate.Struct(data)
	if err != nil {
		return err
	}
	return nil
}

func (aap *Config) writeJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	out, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}
	return nil
}

func (app *Config) errorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload jsonReponse
	payload.Error = true
	payload.Message = err.Error()

	return app.writeJSON(w, statusCode, payload)
}
