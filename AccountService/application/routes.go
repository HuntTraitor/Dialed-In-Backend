package application

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/hunttraitor/dialed-in-backend/handler"
	"github.com/hunttraitor/dialed-in-backend/repository/account"
	"net/http"
)

func (a *App) loadRoutes() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Route("/accounts", a.loadAccountRoutes)
	a.router = router
}

func (a *App) loadAccountRoutes(router chi.Router) {
	accountHandler := &handler.Account{
		Db: &account.Db{
			Pool: a.db,
		},
	}

	router.Post("/", accountHandler.Create)
	router.Get("/", accountHandler.List)
	router.Get("/{id}", accountHandler.GetById)
	router.Put("/{id}", accountHandler.Update)
	router.Delete("/{id}", accountHandler.Delete)
}
