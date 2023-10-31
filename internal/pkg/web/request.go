package web

import (
	"errors"
	"net/http"
)

const code = "code"
const state = "state"

// ParseRequest takes in an http.Request and returns the OAuth code and state parameter and an error.
func ParseRequest(req *http.Request) (string, string, error) {
	urlValues := req.URL.Query()

	codes, ok := urlValues[code]
	if !ok {
		return "", "", errors.New("error code parameter not found in request query")
	}

	states, ok := urlValues[state]
	if !ok {
		return "", "", errors.New("error code parameter not found in request query")
	}

	if len(codes) > 1 {
		return "", "", errors.New("error too many values for the code parameter")
	}

	if codes[0] == "" {
		return "", "", errors.New("error code value cannot be empty")
	}

	if len(states) > 1 {
		return "", "", errors.New("error too many values for the state parameter")
	}

	if states[0] == "" {
		return "", "", errors.New("error state value cannot be empty")
	}

	return codes[0], states[0], nil
}
