package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	route := httprouter.New()
	route.NotFound = http.HandlerFunc(app.notFoundResponse)
	route.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	route.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	route.HandlerFunc(http.MethodGet, "/v1/movies", app.listMovieHandler)
	route.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	route.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	route.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.updateMovieHandler)
	route.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)
	route.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	route.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	route.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	return app.recoverPanic(app.rateLimit(app.authenticate(route)))
}
