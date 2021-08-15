package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/mycok/sunrise-api/internal/data"
	"github.com/mycok/sunrise-api/internal/validator"
)

func (app *application) createAuthenticationTokenHandler(rw http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(rw, r, &input)
	if err != nil {
		app.badRequestResponse(rw, r, err)

		return
	}

	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePassword(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(rw, r, v.Errors)

		return
	}

	// Lookup the user record based on the email address. If no matching user was
	// found, then we call the app.invalidCredentialsResponse() helper to send a 401 
	// Unauthorized response to the client (we will create this helper in a moment).
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(rw, r)
		default:
			app.serverErrorResponse(rw, r, err)
		}

		return
	}

	// Check if the provided password matches the actual password for the user.
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(rw, r, err)

		return
	}

	// If the passwords don't match, then we call the app.invalidCredentialsResponse() 
	// helper again and return.
	if !match {
		app.invalidCredentialsResponse(rw, r)

		return
	}

	// Otherwise, if the password is correct, we generate a new token with a 24-hour
	// expiry time and the scope 'authentication'.
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(rw, r, err)

		return
	}

	// Encode the token to JSON and send it in the response along with a 201 Created
	// status code.
	app.writeJSON(rw, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(rw, r, err)
	}
}