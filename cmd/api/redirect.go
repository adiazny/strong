package main

import (
	"fmt"
	"net/http"

	"github.com/adiazny/strong/internal/pkg/web"
)

func (app *application) redirectHandler(w http.ResponseWriter, r *http.Request) {
	app.log.Printf("redirect handler triggered for %s", r.URL.String())

	code, state, err := web.ParseRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	switch state {
	case "gdrive-state":
		tokens, err := app.gdriveAuthProvider.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		err = app.gdriveAuthProvider.StoreToken(tokens)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		app.log.Printf("storing tokens to path %s", app.gdriveAuthProvider.TokenPath)
	case "strava-state":
		tokens, err := app.stravaAuthProvider.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		err = app.stravaAuthProvider.StoreToken(tokens)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		app.log.Printf("storing tokens to path %s", app.stravaAuthProvider.TokenPath)
	}

	fmt.Fprint(w, "redirect successful")
}
