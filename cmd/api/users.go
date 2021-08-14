package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/mycok/sunrise-api/internal/data"
	"github.com/mycok/sunrise-api/internal/validator"
)

func (app *application) RegisterUserHandler(rw http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(rw, r, &input)
	if err != nil {
		app.badRequestResponse(rw, r, err)

		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(rw, r, err)

		return
	}

	v := validator.New()

	// Validate the user struct and return the error messages to the client if any of
	// the checks fail.
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(rw, r, v.Errors)

		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		// If we get a ErrDuplicateEmail error, use the v.AddError() method to manually // add a message to the validator instance, and then call our
		// failedValidationResponse() helper.
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email already exists")
			app.failedValidationResponse(rw, r, v.Errors)
		default:
			app.serverErrorResponse(rw, r, err)
		}

		return
	}

	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(rw, r, err)

		return
	}

	// Use the background helper to execute an anonymous function that sends
	// emails by calling mailer.Send() method
	app.background(func() {
		// We create a map to hold various pieces of data to pass to our templates
		data := map[string]interface{}{
			"activationToken": token.PlainText,
			"userID": user.ID,
		}

		err = app.mailer.Send(user.Email, "user_welcome.go.tmpl", data)
		if err != nil {
			// If there is an error sending the email then we use the 
			// app.logger.PrintError() helper to manage it, instead of the
			// app.serverErrorResponse() helper.
			app.logger.PrintError(err, nil)
		}
	})

	// Send the client a 202 Accepted status code to indicate that the request has been
	// accepted for processing but the processing has not yet been completed
	err = app.writeJSON(rw, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(rw, r, err)
	}
}
