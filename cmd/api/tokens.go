package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"github.com/hunttraitor/dialed-in-backend/internal/validator"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlainText(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	token, err := app.models.Tokens.New(user.ID, 350*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createPasswordResetTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateEmail(v, input.Email)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.logger.Error(fmt.Sprintf(
				"User with email %s is not associated with this service",
				input.Email,
			))
			env := envelope{"message": "If an account exists for that email, a reset link has been sent."}
			err = app.writeJSON(w, http.StatusCreated, env, nil)
			if err != nil {
				app.serverErrorResponse(w, r, err)
			}
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	token, err := app.models.Tokens.New(user.ID, 45*time.Minute, data.ScopePasswordReset)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Email the user
	app.background(func() {
		params := map[string]any{
			"passwordResetToken": token.Plaintext,
			"userName":           user.Name,
		}
		err := app.mailer.Send(user.Email, "token_password_reset.tmpl", params)
		if err != nil {
			app.logger.Error(err.Error())
		}
	})

	env := envelope{"message": "If an account exists for that email, a reset link has been sent."}

	err = app.writeJSON(w, http.StatusCreated, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
