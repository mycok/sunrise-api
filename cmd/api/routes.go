package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requiresPermission("movies:write", app.createMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requiresPermission("movies:read", app.listMoviesHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requiresPermission("movies:read", app.showMovieHandler))
	router.HandlerFunc(http.MethodPut, "/v1/movies/:id", app.requiresPermission("movies:write", app.replaceMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requiresPermission("movies:write", app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requiresPermission("movies:write", app.deleteMovieHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.RegisterUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	router.HandlerFunc(http.MethodPost, "/v1/permissions", app.addPermissionForUserHandler)

	router.HandlerFunc(http.MethodGet, "/debug/metrics", app.requiresPermission("metrics:view", expvar.Handler().ServeHTTP))

	return app.metrics(app.recoverFromPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
