package main

import "net/http"

func (app *application) listCoffeesHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	coffees, err := app.models.Coffees.GetAllForUser(user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"coffees": coffees}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
