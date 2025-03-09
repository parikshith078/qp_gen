package main

import (
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
	})

	// Protected routes

	return mux
}
