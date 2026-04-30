package main

import (
	"expvar"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (app *application) routes() http.Handler {
	router := chi.NewRouter()

	router.Use(app.recoverPanic)
	router.Handle("/metrics", promhttp.Handler())

	// API routes - logged and rate limited
	router.Group(func(r chi.Router) {
		r.Use(app.authenticate)
		r.Use(app.logRequest, app.rateLimit)
		r.Use(app.metrics)

		r.Route("/v1/healthcheck", app.loadHealthCheckRoutes)
		r.Route("/v1/users", app.loadUserRoutes)
		r.Route("/v1/tokens", app.loadTokenRoutes)
		r.Route("/v1/methods", app.loadMethodRoutes)
		r.Route("/v1/coffees", app.loadCoffeeRoutes)
		r.Route("/v1/recipes", app.loadRecipeRoutes)
		r.Route("/v1/grinders", app.loadGrinderRoutes)
		r.Route("/debug", app.loadDebugRoutes)
	})

	router.NotFound(app.notFoundResponse)
	router.MethodNotAllowed(app.methodNotAllowedResponse)

	// Frontend - no logging, no rate limit
	if app.config.env == "development" {
		router.Handle("/*", app.devProxyHandler())
	} else {
		subFS, err := fs.Sub(app.frontendFS, "dist")
		if err != nil {
			panic(err)
		}
		router.Handle("/*", app.spaHandler(http.FS(subFS)))
	}

	return router
}

func (app *application) loadHealthCheckRoutes(router chi.Router) {
	router.Get("/", app.healthcheckHandler)
}

func (app *application) spaHandler(fsys http.FileSystem) http.HandlerFunc {
	fileServer := http.FileServer(fsys)
	return func(w http.ResponseWriter, r *http.Request) {
		f, err := fsys.Open(r.URL.Path)
		if err != nil {
			if os.IsNotExist(err) {
				r.URL.Path = "/"
			}
		} else {
			f.Close()
		}
		fileServer.ServeHTTP(w, r)
	}
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
	router.Get("/{id}", app.getMethodHandler)
}

func (app *application) loadCoffeeRoutes(router chi.Router) {
	router.With(app.requireAuthenticatedUser).Get("/", app.listCoffeesHandler)
	router.With(app.requireAuthenticatedUser).Post("/", app.createCoffeeHandler)
	router.With(app.requireAuthenticatedUser).Get("/{id}", app.getCoffeeHandler)
	router.With(app.requireAuthenticatedUser).Patch("/{id}", app.updateCoffeeHandler)
	router.With(app.requireAuthenticatedUser).Delete("/{id}", app.deleteCoffeeHandler)
}

func (app *application) loadRecipeRoutes(router chi.Router) {
	router.With(app.requireAuthenticatedUser).Post("/", app.createRecipeHandler)
	router.With(app.requireAuthenticatedUser).Get("/", app.listRecipesHandler)
	router.With(app.requireAuthenticatedUser).Get("/{id}", app.getOneRecipeHandler)
	router.With(app.requireAuthenticatedUser).Patch("/{id}", app.updateRecipeHandler)
	router.With(app.requireAuthenticatedUser).Delete("/{id}", app.deleteRecipeHandler)
}

func (app *application) loadGrinderRoutes(router chi.Router) {
	router.With(app.requireAuthenticatedUser).Get("/", app.listGrindersHandler)
	router.With(app.requireAuthenticatedUser).Post("/", app.createGrinderHandler)
	router.With(app.requireAuthenticatedUser).Get("/{id}", app.getGrinderHandler)
	router.With(app.requireAuthenticatedUser).Patch("/{id}", app.updateGrinderHandler)
	router.With(app.requireAuthenticatedUser).Delete("/{id}", app.deleteGrinderHandler)
}

func (app *application) devProxyHandler() http.HandlerFunc {
	target, _ := url.Parse("http://frontend:5173")
	proxy := httputil.NewSingleHostReverseProxy(target)
	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}
