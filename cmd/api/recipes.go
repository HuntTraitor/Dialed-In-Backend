package main

import (
	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"github.com/hunttraitor/dialed-in-backend/internal/validator"
	"net/http"
	"strconv"
)

func (app *application) createRecipeHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	var input struct {
		MethodId int64           `json:"method_id"`
		CoffeeId int64           `json:"coffee_id"`
		Info     data.RecipeInfo `json:"info"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	recipe := &data.Recipe{
		UserID:   user.ID,
		MethodID: input.MethodId,
		CoffeeID: input.CoffeeId,
		Info:     input.Info,
	}

	v := validator.New()
	if data.ValidateRecipe(v, recipe); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Recipes.Insert(recipe)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusCreated, envelope{"recipe": recipe}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listRecipesHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	var input struct {
		CoffeeId int64 `json:"coffee_id"`
		MethodId int64 `json:"method_id"`
	}

	qs := r.URL.Query()

	// get the query parameters coffee_id and method_id
	strCoffeeId, err := strconv.Atoi(qs.Get("coffee_id"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	input.CoffeeId = int64(strCoffeeId)
	strMethodId, err := strconv.Atoi(qs.Get("method_id"))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	input.MethodId = int64(strMethodId)

	recipes, err := app.models.Recipes.GetAllForUser(user.ID, input.CoffeeId, input.MethodId)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"recipes": recipes}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
