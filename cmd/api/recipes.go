package main

import (
	"encoding/json"
	"errors"
	"fmt"
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
		CoffeeId *int64          `json:"coffee_id,omitempty"`
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
			app.unknownMethodResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	v := validator.New()
	if data.ValidateRecipe(v, recipe); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	var coffee *data.Coffee

	if recipe.CoffeeID != nil {
		coffee, err = app.models.Coffees.GetOne(*recipe.CoffeeID, recipe.UserID)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.unknownCoffeeResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// pre-sign the image url if we actually have a coffee
		if coffee != nil && coffee.Info.Img != "" {
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
		}
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
		Coffee:    coffee,
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

	fullRecipes := make([]*data.FullRecipe, 0, len(recipes))

	for _, recipe := range recipes {
		method, err := app.models.Methods.GetOne(recipe.MethodID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// Coffee is optional
		var coffee *data.Coffee

		if recipe.CoffeeID != nil {
			coffee, err = app.models.Coffees.GetOne(*recipe.CoffeeID, recipe.UserID)
			if err != nil {
				switch {
				case errors.Is(err, data.ErrRecordNotFound):
					app.unknownCoffeeResponse(w, r)
				default:
					app.serverErrorResponse(w, r, err)
				}
				return
			}

			// Pre-sign image URL only if we actually have a coffee and an image
			if coffee != nil && coffee.Info.Img != "" {
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
			}
		}

		fullRecipe := &data.FullRecipe{
			ID:        recipe.ID,
			UserID:    recipe.UserID,
			Method:    *method,
			Coffee:    coffee,
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

func (app *application) updateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	recipe, err := app.models.Recipes.Get(id, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.unknownCoffeeResponse(w, r)
		default:
			fmt.Println(err)
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		MethodID *int64           `json:"method_id"`
		CoffeeID NullInt64        `json:"coffee_id"`
		Info     *json.RawMessage `json:"info"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.MethodID != nil {
		recipe.MethodID = *input.MethodID
	}

	if input.CoffeeID.Present {
		recipe.CoffeeID = input.CoffeeID.Value
	}

	if input.Info != nil {
		recipe.Info = *input.Info
	}

	method, err := app.models.Methods.GetOne(recipe.MethodID)
	if err != nil {
		app.unknownMethodResponse(w, r)
		return
	}

	var coffee *data.Coffee

	if recipe.CoffeeID != nil {
		coffee, err = app.models.Coffees.GetOne(*recipe.CoffeeID, recipe.UserID)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.unknownCoffeeResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		if coffee != nil && coffee.Info.Img != "" {
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
		}
	}

	v := validator.New()
	if data.ValidateRecipe(v, recipe); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Recipes.Update(recipe)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	fullRecipe := &data.FullRecipe{
		ID:        recipe.ID,
		UserID:    recipe.UserID,
		Method:    *method,
		Coffee:    coffee,
		Info:      recipe.Info,
		CreatedAt: recipe.CreatedAt,
		Version:   recipe.Version,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"recipe": fullRecipe}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteRecipeHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Recipes.Delete(id, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.unknownCoffeeResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "recipe successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
