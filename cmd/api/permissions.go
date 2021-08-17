package main

import (
	"net/http"

	"github.com/mycok/sunrise-api/internal/validator"
)

func (app *application) addPermissionForUserHandler(rw http.ResponseWriter, r *http.Request) {
	var input struct {
		Permission string `json:"permission"`
	}

	err := app.readJSON(rw, r, &input)
	if err != nil {
		app.badRequestResponse(rw, r, err)

		return
	}

	v := validator.New()
	if v.Check(input.Permission != "", "permission", "must be provided"); !v.Valid() {
		app.failedValidationResponse(rw, r, v.Errors)

		return
	}

	user := app.contextGetUser(r)

	err = app.models.Permissions.AddForUser(user.ID, input.Permission)
	if err != nil {
		app.serverErrorResponse(rw, r, err)

		return
	}

	err = app.writeJSON(rw, http.StatusOK, envelope{"message": "permission successfully added"}, nil)
	if err != nil {
		app.serverErrorResponse(rw, r, err)
	}
}
