package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/v1/health", app.healthHandler)
	router.HandlerFunc(http.MethodGet, "/v1/redirect", app.redirectHandler)

	return router
}
