package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {
	route := httprouter.New()

	route.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	route.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	route.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)

	return route
}
