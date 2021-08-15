package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requiresActivatedUser(app.createMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requiresActivatedUser(app.listMoviesHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requiresActivatedUser(app.showMovieHandler))
	router.HandlerFunc(http.MethodPut, "/v1/movies/:id", app.requiresActivatedUser(app.replaceMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requiresActivatedUser(app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requiresActivatedUser(app.deleteMovieHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.RegisterUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	return app.recoverFromPanic(app.rateLimit(app.authenticate(router)))
}
