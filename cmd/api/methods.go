package main

import (
	"errors"
	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"net/http"
)

func (app *application) listMethodsHandler(w http.ResponseWriter, r *http.Request) {
	methods, err := app.models.Methods.GetAll()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"methods": methods}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getMethodHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	method, err := app.models.Methods.GetOne(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"method": method}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
