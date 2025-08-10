package main

import (
	"encoding/json"
	"errors"
	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"github.com/hunttraitor/dialed-in-backend/internal/s3"
	"github.com/hunttraitor/dialed-in-backend/internal/validator"
	"net/http"
	"time"
)

func (app *application) createRecipeHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	var input struct {
		MethodId int64           `json:"method_id"`
		CoffeeId int64           `json:"coffee_id"`
		Info     json.RawMessage `json:"info"`
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

	v := validator.New()
	if data.ValidateRecipe(v, recipe, method); !v.Valid() {
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

	// pre-sign the image url
	imgURL, err := s3.PreSignURL(
		s3.WithPresigner(app.s3.Presigner),
		s3.WithPresignBucket(app.config.s3.bucket),
		s3.WithPresignFilePath("coffees/"+coffee.Info.Img),
		s3.WithPresignExpiration(time.Hour*24),
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	coffee.Info.Img = imgURL

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

	qs := r.URL.Query()

	recipes, err := app.models.Recipes.GetAllForUser(user.ID, qs)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	fullRecipes := []*data.FullRecipe{}

	for _, recipe := range recipes {
		coffee, err := app.models.Coffees.GetOne(recipe.CoffeeID, recipe.UserID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// pre-sign the image url
		imgURL, err := s3.PreSignURL(
			s3.WithPresigner(app.s3.Presigner),
			s3.WithPresignBucket(app.config.s3.bucket),
			s3.WithPresignFilePath("coffees/"+coffee.Info.Img),
			s3.WithPresignExpiration(time.Hour*24),
		)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		coffee.Info.Img = imgURL

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
