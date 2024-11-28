package main

import (
	"expvar"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := chi.NewRouter()

	// if metrics are enabled
	if app.config.metrics {
		router.Use(app.metrics)
	}

	// Middleware
	router.Use(app.recoverPanic, app.rateLimit, app.authenticate)

	router.NotFound(app.notFoundResponse)
	router.MethodNotAllowed(app.methodNotAllowedResponse)

	// Health Check Routes
	router.Route("/v1/healthcheck", app.loadHealthCheckRoutes)

	// User Routes
	router.Route("/v1/users", app.loadUserRoutes)

	// Token Routes
	router.Route("/v1/tokens", app.loadTokenRoutes)

	router.Route("/debug", app.loadDebugRoutes)

	return router
}

func (app *application) loadHealthCheckRoutes(router chi.Router) {
	router.Get("/", app.healthcheckHandler)
}

func (app *application) loadUserRoutes(router chi.Router) {
	router.Post("/", app.registerUserHandler)
	router.Put("/activated", app.activateUserHandler)
}

func (app *application) loadTokenRoutes(router chi.Router) {
	router.Post("/authentication", app.createAuthenticationTokenHandler)
}

func (app *application) loadDebugRoutes(router chi.Router) {
	router.Get("/vars", func(w http.ResponseWriter, r *http.Request) {
		expvar.Handler().ServeHTTP(w, r)
	})
}
