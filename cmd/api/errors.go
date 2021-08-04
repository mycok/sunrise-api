package main

import (
	"fmt"
	"net/http"
)

func (app *application) logError(r *http.Request, err error) {
	app.logger.Println(err)
}

func (app *application) errResponse(wr http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{"error": message}

	err := app.writeJSON(wr, status, env, nil)
	if err != nil {
		app.logError(r, err)
		wr.WriteHeader(500)
	}
}

func (app *application) serverErrorResponse(wr http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)
	message := "the server encountered a problem and could not process your request"

	app.errResponse(wr, r, http.StatusInternalServerError, message)
}

func (app *application) notFoundResponse(wr http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"

	app.errResponse(wr, r, http.StatusNotFound, message)
}

func (app *application) methodNotAllowedResponse(wr http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)

	app.errResponse(wr, r, http.StatusMethodNotAllowed, message)
}

func (app *application) badRequestResponse(wr http.ResponseWriter, r *http.Request, err error) {
	app.errResponse(wr, r, http.StatusBadRequest, err.Error())
}

func (app *application) failedValidationResponse(wr http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errResponse(wr, r, http.StatusUnprocessableEntity, errors)
}

func (app *application) editConflictResponse(wr http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.errResponse(wr, r, http.StatusConflict, message)
}