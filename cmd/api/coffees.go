package main

import (
	"bytes"
	"errors"
	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"github.com/hunttraitor/dialed-in-backend/internal/s3"
	"github.com/hunttraitor/dialed-in-backend/internal/validator"
	"io"
	"net/http"
	"time"
)

func (app *application) listCoffeesHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	coffees, err := app.models.Coffees.GetAllForUser(user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	for _, coffee := range coffees {
		// pre-sign the image url
		var imgURL string
		imgURL, err = s3.PreSignURL(app.s3.Presigner, app.config.s3.bucket, "coffees/"+coffee.Img, time.Hour*24)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		coffee.Img = imgURL
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"coffees": coffees}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createCoffeeHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	// limit 10mb
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// extract image from form
	img, handler, err := r.FormFile("img")
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	defer img.Close()

	// convert image to byte buffer
	var buf bytes.Buffer
	_, err = io.Copy(&buf, img)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// upload byte buffer to s3
	fileName, err := s3.UploadToS3(app.s3.Client, buf, handler.Header, "coffees/", app.config.s3.bucket)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	coffee := &data.Coffee{
		Name:        r.Form.Get("name"),
		Region:      r.Form.Get("region"),
		Process:     r.Form.Get("process"),
		Img:         fileName,
		Description: r.Form.Get("description"),
	}

	// validate the input
	v := validator.New()
	if data.ValidateCoffee(v, coffee); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	coffee, err = app.models.Coffees.Insert(user.ID, coffee)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Pre-sign the URL to send back to the client
	imgURL, err := s3.PreSignURL(app.s3.Presigner, app.config.s3.bucket, "coffees/"+coffee.Img, time.Hour*24)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	coffee.Img = imgURL

	err = app.writeJSON(w, http.StatusCreated, envelope{"coffee": coffee}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getCoffeeHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	coffee, err := app.models.Coffees.GetOne(id, user.ID)
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
	imgURL, err := s3.PreSignURL(app.s3.Presigner, app.config.s3.bucket, "coffees/"+coffee.Img, time.Hour*24)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	coffee.Img = imgURL

	err = app.writeJSON(w, http.StatusOK, envelope{"coffee": coffee}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateCoffeeHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// get the coffee we need to update
	coffee, err := app.models.Coffees.GetOne(id, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Extract form values as pointers using the helper
	name := GetOptionalString(r, "name")
	region := GetOptionalString(r, "region")
	process := GetOptionalString(r, "process")
	description := GetOptionalString(r, "description")

	// update the fields
	if name != nil {
		coffee.Name = *name
	}
	if region != nil {
		coffee.Region = *region
	}
	if process != nil {
		coffee.Process = *process
	}
	if description != nil {
		coffee.Img = *description
	}

	// check if an image is uploaded and if so replace the image
	imgFile, header, err := r.FormFile("img")
	if err == nil {
		// convert image to byte buffer
		var buf bytes.Buffer
		_, err = io.Copy(&buf, imgFile)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// upload byte buffer to s3
		var fileName string
		fileName, err = s3.UploadToS3(app.s3.Client, buf, header.Header, "coffees/", app.config.s3.bucket)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		coffee.Img = fileName
	}

	// validate the new coffee struct
	v := validator.New()
	if data.ValidateCoffee(v, coffee); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// update the new coffee model in the database
	err = app.models.Coffees.Update(coffee)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// PreSign the URL to send back to the client
	imgURL, err := s3.PreSignURL(app.s3.Presigner, app.config.s3.bucket, "coffees/"+coffee.Img, time.Hour*24)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	coffee.Img = imgURL

	// write the new coffee model to the response
	err = app.writeJSON(w, http.StatusOK, envelope{"coffee": coffee}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteCoffeeHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Coffees.Delete(id, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "coffee successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
