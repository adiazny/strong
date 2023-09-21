package web_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/adiazny/strong/internal/pkg/web"
)

func TestParseRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *http.Request
		want    string
		wantErr bool
	}{
		{
			name:    "success",
			req:     newRequest("code=1234567890"),
			want:    "1234567890",
			wantErr: false,
		},
		{
			name:    "empty code value",
			req:     newRequest("code="),
			wantErr: true,
		},
		{
			name:    "code param not present",
			req:     newRequest(""),
			wantErr: true,
		},
		{
			name:    "code param multiple values",
			req:     newRequest("code=1234567890&code=5647382910"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := web.ParseRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
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
