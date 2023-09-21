package web

import (
	"errors"
	"net/http"
)

const code = "code"

// ParseRequest takes in an http.Request and returns the OAuth Code parameter and an error.
func ParseRequest(req *http.Request) (string, error) {
	urlValues := req.URL.Query()

	values, ok := urlValues[code]
	if !ok {
		return "", errors.New("error code parameter not found in request query")
	}

	if len(values) > 1 {
		return "", errors.New("error too many values for the code parameter")
	}

	if values[0] == "" {
		return "", errors.New("error code value cannot be empty")
	}

	return values[0], nil
}
