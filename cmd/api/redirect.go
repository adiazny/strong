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
		_, err := app.gdriveAuthProvider.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	case "strava-state":
		_, err := app.stravaAuthProvider.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	fmt.Fprint(w, "redirect successful")
}
