package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := chi.NewRouter()

	// Middleware
	router.Use(app.recoverPanic, app.rateLimit)

	router.NotFound(app.notFoundResponse)
	router.MethodNotAllowed(app.methodNotAllowedResponse)

	// Health Check Routes
	router.Route("/v1/healthcheck", app.loadHealthCheckRoutes)

	// User Routes
	router.Route("/v1/users", app.loadUserRoutes)

	return router
}

func (app *application) loadHealthCheckRoutes(router chi.Router) {
	router.Get("/", app.healthcheckHandler)
}

func (app *application) loadUserRoutes(router chi.Router) {
	router.Post("/", app.registerUserHandler)
}
