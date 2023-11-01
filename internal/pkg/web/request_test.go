package web_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/adiazny/strong/internal/pkg/web"
)

func TestParseRequest(t *testing.T) {
	tests := []struct {
		name      string
		req       *http.Request
		wantCode  string
		wantState string
		wantErr   bool
	}{
		{
			name:      "success gdrive request",
			req:       newRequest("state=gdrive-state&code=1234567890"),
			wantCode:  "1234567890",
			wantState: "gdrive-state",
			wantErr:   false,
		},
		{
			name:    "empty code param value",
			req:     newRequest("state=gdrive-state&code="),
			wantErr: true,
		},
		{
			name:    "empty state param value",
			req:     newRequest("state=&code=1234567890"),
			wantErr: true,
		},
		{
			name:    "code param not present",
			req:     newRequest("state=gdrive-state"),
			wantErr: true,
		},
		{
			name:    "state param not present",
			req:     newRequest("code=1234567890"),
			wantErr: true,
		},
		{
			name:    "multiple code params",
			req:     newRequest("code=1234567890&code=5647382910"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			code, state, err := web.ParseRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if code != tt.wantCode {
				t.Errorf("got %v, want %v", code, tt.wantCode)
			}

			if state != tt.wantState {
				t.Errorf("got %v, want %v", code, tt.wantState)
			}
		})
	}
}

func newRequest(codeParam string) *http.Request {
	var url string

	if codeParam == "" {
		url = "localhost.com/callback"
	}

	url = fmt.Sprintf("localhost.com/callback?%s", codeParam)

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	return req
}
