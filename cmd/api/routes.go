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
	router.Use(app.recoverPanic, app.logRequest, app.rateLimit, app.authenticate)

	router.NotFound(app.notFoundResponse)
	router.MethodNotAllowed(app.methodNotAllowedResponse)

	// Health Check Routes
	router.Route("/v1/healthcheck", app.loadHealthCheckRoutes)

	// User Routes
	router.Route("/v1/users", app.loadUserRoutes)

	// Token Routes
	router.Route("/v1/tokens", app.loadTokenRoutes)

	// Method Routes
	router.Route("/v1/methods", app.loadMethodRoutes)

	// Coffee Routes
	router.Route("/v1/coffees", app.loadCoffeeRoutes)

	router.Route("/debug", app.loadDebugRoutes)

	return router
}

func (app *application) loadHealthCheckRoutes(router chi.Router) {
	router.Get("/", app.healthcheckHandler)
}

func (app *application) loadUserRoutes(router chi.Router) {
	router.Post("/", app.registerUserHandler)
	router.Put("/activated", app.activateUserHandler)
	router.With(app.requireAuthenticatedUser).Get("/verify", app.verifyUserHandler)
	router.Put("/password", app.updateUserPasswordHandler)
}

func (app *application) loadTokenRoutes(router chi.Router) {
	router.Post("/authentication", app.createAuthenticationTokenHandler)
	router.Post("/password-reset", app.createPasswordResetTokenHandler)
}

func (app *application) loadDebugRoutes(router chi.Router) {
	router.Get("/vars", func(w http.ResponseWriter, r *http.Request) {
		expvar.Handler().ServeHTTP(w, r)
	})
}

func (app *application) loadMethodRoutes(router chi.Router) {
	router.Get("/", app.listMethodsHandler)
}

func (app *application) loadCoffeeRoutes(router chi.Router) {
	router.With(app.requireAuthenticatedUser).Get("/", app.listCoffeesHandler)
	router.With(app.requireAuthenticatedUser).Post("/", app.createCoffeeHandler)
	router.With(app.requireAuthenticatedUser).Get("/{id}", app.getCoffeeHandler)
	router.With(app.requireAuthenticatedUser).Patch("/{id}", app.updateCoffeeHandler)
	router.With(app.requireAuthenticatedUser).Delete("/{id}", app.deleteCoffeeHandler)
}
