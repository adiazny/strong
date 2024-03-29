package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) healthHandler(w http.ResponseWriter, r *http.Request) {
	health := struct {
		Status  string
		Version string
	}{
		Status:  "available",
		Version: version,
	}

	data, err := json.Marshal(health)
	if err != nil {
		http.Error(w, "error marshling health data to json", http.StatusInternalServerError)
	}

	w.Write(data)
}
