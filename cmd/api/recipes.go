package main

import (
	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"github.com/hunttraitor/dialed-in-backend/internal/validator"
	"net/http"
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
