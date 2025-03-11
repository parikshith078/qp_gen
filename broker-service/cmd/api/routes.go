package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (app *Config) routes() *chi.Mux {
	mux := chi.NewRouter()
	// make it open for all
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	mux.Use(middleware.Heartbeat("/ping"))

	// Public routes
	mux.Group(func(r chi.Router) {
		r.Get("/health", app.HealthCheck)
		r.Post("/register", app.Register)
		r.Post("/login", app.Login)
	})

	// Protected routesut
	mux.Group(func(r chi.Router) {
		r.Use(app.AuthMiddleware)
		r.Post("/logout", app.Logout)
		r.Post("/test", func(w http.ResponseWriter, r *http.Request) {
			id, err := GetUserIDFromContext(r.Context())
			if err != nil {
				app.errorJSON(w, err, http.StatusUnauthorized)
				return
			}
			log.Print("UserID: ", id)
			user, err := app.Db.GetUserByID(r.Context(), id)
			if err != nil {
				app.errorJSON(w, err, http.StatusInternalServerError)
				return
			}
			app.writeJSON(w, http.StatusOK, &jsonReponse{
				Message: "got it",
				Data:    user,
			})
		})
	})

	return mux
}
