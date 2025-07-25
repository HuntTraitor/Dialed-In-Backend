package main

import (
	"bytes"
	"errors"
	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"github.com/hunttraitor/dialed-in-backend/internal/s3"
	"github.com/hunttraitor/dialed-in-backend/internal/validator"
	"io"
	"mime/multipart"
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
		if coffee.Info.Img != "" {
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

	err = app.writeJSON(w, http.StatusOK, envelope{"coffees": coffees}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createCoffeeHandler(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	var input struct {
		Name         string   `form:"name"`
		Roaster      string   `form:"roaster"`
		Region       string   `form:"region"`
		Process      string   `form:"process"`
		Description  string   `form:"description"`
		Decaf        bool     `form:"decaf"`
		OriginType   string   `form:"origin_type"`
		TastingNotes []string `form:"tasting_notes"`
		Rating       int      `form:"rating"`
		RoastLevel   string   `form:"roast_level"`
		Cost         float64  `form:"cost"`
		Image        []byte   `form:"image"`
	}

	// limit 10mb
	err := app.readMultipart(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// validate the input
	v := validator.New()

	// extract image from form
	var (
		img     multipart.File
		handler *multipart.FileHeader
	)
	img, handler, err = r.FormFile("img")

	// An error other than missing file occurred
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		app.badRequestResponse(w, r, err)
		return
	}

	coffee := &data.Coffee{
		UserID: int(user.ID),
		Info: data.CoffeeInfo{
			Name:         input.Name,
			Roaster:      input.Roaster,
			Region:       input.Region,
			Process:      input.Process,
			Description:  input.Description,
			Decaf:        input.Decaf,
			OriginType:   input.OriginType,
			TastingNotes: input.TastingNotes,
			Rating:       input.Rating,
			RoastLevel:   input.RoastLevel,
			Cost:         float64(input.Cost),
			// Img will be set from `img` or base64/image path
		},
	}

	if data.ValidateCoffee(v, coffee); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	if img != nil {
		defer img.Close()

		var buf bytes.Buffer
		_, err = io.Copy(&buf, img)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		fileName, err := s3.UploadToS3(
			s3.WithClient(app.s3.Client),
			s3.WithFile(buf),
			s3.WithFileType(handler.Header),
			s3.WithBucket(app.config.s3.bucket),
			s3.WithFilePath("coffees/"))
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		coffee.Info.Img = fileName
	}

	err = app.models.Coffees.Insert(coffee)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if coffee.Info.Img != "" {
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
	if coffee.Info.Img != "" {
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

	var input struct {
		Name         *string   `form:"name"`
		Roaster      *string   `form:"roaster"`
		Region       *string   `form:"region"`
		Process      *string   `form:"process"`
		Description  *string   `form:"description"`
		Decaf        *bool     `form:"decaf"`
		OriginType   *string   `form:"origin_type"`
		TastingNotes *[]string `form:"tasting_notes"`
		Rating       *int      `form:"rating"`
		RoastLevel   *string   `form:"roast_level"`
		Cost         *float64  `form:"cost"`
		Image        []byte    `form:"image"`
	}

	err = app.readMultipart(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Name != nil {
		coffee.Info.Name = *input.Name
	}
	if input.Roaster != nil {
		coffee.Info.Roaster = *input.Roaster
	}
	if input.Region != nil {
		coffee.Info.Region = *input.Region
	}
	if input.Process != nil {
		coffee.Info.Process = *input.Process
	}
	if input.Description != nil {
		coffee.Info.Description = *input.Description
	}
	if input.Decaf != nil {
		coffee.Info.Decaf = *input.Decaf
	}
	if input.OriginType != nil {
		coffee.Info.OriginType = *input.OriginType
	}
	if input.TastingNotes != nil {
		coffee.Info.TastingNotes = *input.TastingNotes
	}
	if input.Rating != nil {
		coffee.Info.Rating = *input.Rating
	}
	if input.RoastLevel != nil {
		coffee.Info.RoastLevel = *input.RoastLevel
	}
	if input.Cost != nil {
		coffee.Info.Cost = *input.Cost
	}

	// check if an image is uploaded and if so replace the image
	imgFile, header, err := r.FormFile("img")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		// Real error, not just missing file
		app.badRequestResponse(w, r, err)
		return
	}
	if err == nil {
		defer imgFile.Close()

		var buf bytes.Buffer
		_, err = io.Copy(&buf, imgFile)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		var fileName string
		fileName, err = s3.UploadToS3(
			s3.WithClient(app.s3.Client),
			s3.WithFile(buf),
			s3.WithFileType(header.Header),
			s3.WithFilePath("coffees/"),
			s3.WithOldFileName(coffee.Info.Img),
			s3.WithBucket(app.config.s3.bucket),
		)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		coffee.Info.Img = fileName
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
	if coffee.Info.Img != "" {
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

	deleted, err := s3.DeleteFromS3(
		s3.WithDeleteClient(app.s3.Client),
		s3.WithDeleteBucket(app.config.s3.bucket),
		s3.WithDeleteFilePath("coffees/"+coffee.Info.Img),
	)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !deleted {
		app.logger.Info("S3 image deletion failed for " + coffee.Info.Img)
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
