package main

import (
	"errors"
	"fmt"
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

	coffee, err := app.models.Coffees.GetOne(recipe.CoffeeID, recipe.UserID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	method, err := app.models.Methods.GetOne(recipe.MethodID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.models.Recipes.Insert(recipe)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	fullRecipe := &data.FullRecipe{
		ID:        recipe.ID,
		UserID:    recipe.UserID,
		Method:    *method,
		Coffee:    *coffee,
		Info:      recipe.Info,
		CreatedAt: recipe.CreatedAt,
		Version:   recipe.Version,
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"recipe": fullRecipe}, nil)
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

	// Parse coffee_id
	strCoffeeId := qs.Get("coffee_id")
	if strCoffeeId != "" {
		coffeeId, err := strconv.Atoi(strCoffeeId)
		if err != nil {
			app.badRequestResponse(w, r, fmt.Errorf("invalid coffee_id: %w", err))
			return
		}
		input.CoffeeId = int64(coffeeId)
	}

	// Parse method_id
	strMethodId := qs.Get("method_id")
	if strMethodId != "" {
		methodId, err := strconv.Atoi(strMethodId)
		if err != nil {
			app.badRequestResponse(w, r, fmt.Errorf("invalid method_id: %w", err))
			return
		}
		input.MethodId = int64(methodId)
	}

	recipes, err := app.models.Recipes.GetAllForUser(user.ID, input.CoffeeId, input.MethodId)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	var fullRecipes []*data.FullRecipe

	for _, recipe := range recipes {
		coffee, err := app.models.Coffees.GetOne(recipe.CoffeeID, recipe.UserID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		method, err := app.models.Methods.GetOne(recipe.MethodID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		fullRecipe := &data.FullRecipe{
			ID:        recipe.ID,
			UserID:    recipe.UserID,
			Method:    *method,
			Coffee:    *coffee,
			Info:      recipe.Info,
			CreatedAt: recipe.CreatedAt,
			Version:   recipe.Version,
		}

		fullRecipes = append(fullRecipes, fullRecipe)
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"recipes": fullRecipes}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
