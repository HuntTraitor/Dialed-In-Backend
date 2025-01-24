package main

import (
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
