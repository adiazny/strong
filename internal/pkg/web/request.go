package web

import (
	"fmt"
	"net/http"
	"net/url"
)

const (
	code         = "code"
	state        = "state"
	maxValueSize = 1
)

// ParseRequest takes in an http.Request and returns the OAuth code and state parameter and an error.
func ParseRequest(req *http.Request) (string, string, error) {
	urlValues := req.URL.Query()

	codes, err := extractURLValues(code, urlValues)
	if err != nil {
		return "", "", err
	}

	states, err := extractURLValues(state, urlValues)
	if err != nil {
		return "", "", err
	}

	return codes[0], states[0], nil
}

func extractURLValues(key string, urlValues url.Values) ([]string, error) {
	values, ok := urlValues[key]
	if !ok {
		return nil, fmt.Errorf("error '%s' url parameter not found in http request query", key)
	}

	if len(values) > maxValueSize {
		return nil, fmt.Errorf("error too many values for the '%s' url parameter", key)
	}

	if values[0] == "" {
		return nil, fmt.Errorf("error '%s' url parameter value cannot be empty", key)
	}

	return values, nil
}
