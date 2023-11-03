package auth_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/adiazny/strong/internal/pkg/auth"
	"golang.org/x/oauth2"
)

func TestNewProvider(t *testing.T) {
	type args struct {
		service      auth.ServiceID
		tokenPath    string
		clientID     string
		clientSecret string
		redirectURL  string
	}
	tests := []struct {
		name    string
		args    args
		want    *auth.Provider
		wantErr bool
	}{
		{
			name:    "success gdrive",
			args:    args{service: auth.GDriveService, tokenPath: "gdrive/storage.json", clientID: "1234567890", clientSecret: "gdrive-client-secret", redirectURL: "http://localhost:4001/v1/redirect"},
			wantErr: false,
		},
		{
			name:    "success strava",
			args:    args{service: auth.GDriveService, tokenPath: "strava/storage.json", clientID: "1234567890", clientSecret: "strava-client-secret", redirectURL: "http://localhost:4001/v1/redirect"},
			wantErr: false,
		},
		{
			name:    "empty client id",
			args:    args{service: auth.GDriveService, tokenPath: "strava/storage.json", clientID: "", clientSecret: "strava-client-secret", redirectURL: "http://localhost:4001/v1/redirect"},
			wantErr: true,
		},
		{
			name:    "empty client secret",
			args:    args{service: auth.GDriveService, tokenPath: "strava/storage.json", clientID: "1234567890", clientSecret: "", redirectURL: "http://localhost:4001/v1/redirect"},
			wantErr: true,
		},
		{
			name:    "empty redirect url",
			args:    args{service: auth.GDriveService, tokenPath: "strava/storage.json", clientID: "1234567890", clientSecret: "strava-client-secret", redirectURL: ""},
			wantErr: true,
		},
		{
			name:    "invalid service id",
			args:    args{service: 999, tokenPath: "strava/storage.json", clientID: "1234567890", clientSecret: "strava-client-secret", redirectURL: "http://localhost:4001/v1/redirect"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := auth.NewProvider(tt.args.service, tt.args.tokenPath, tt.args.clientID, tt.args.clientSecret, tt.args.redirectURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if (got == nil) != tt.wantErr {
				t.Errorf("error NewProvider() is %v", got)
			}
		})
	}
}

func TestProvider_TokenNotPresent(t *testing.T) {
	tests := []struct {
		name      string
		tokenPath string
		want      bool
	}{
		{
			name:      "success",
			tokenPath: "testdata/storage.json",
			want:      true,
		},
		{
			name:      "fail",
			tokenPath: "testdata/does-not-exist.json",
			want:      true,
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := createProvider(t, tt.tokenPath)

			if got := p.TokenNotPresent(); got != tt.want {
				t.Errorf("Provider.TokenNotPresent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_FileTokens(t *testing.T) {
	tests := []struct {
		name    string
		want    *oauth2.Token
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "success",
			want: &oauth2.Token{
				AccessToken:  "1234567890",
				TokenType:    "Bearer",
				RefreshToken: "0987654321",
				Expiry:       parseTime("2023-10-31T16:56:05.570863-04:00"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := createProvider(t, "")

			got, err := p.FileTokens()
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.FileTokens() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Provider.FileTokens() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_StoreToken(t *testing.T) {
	type args struct {
		token *oauth2.Token
	}
	tests := []struct {
		name    string
		p       *auth.Provider
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := tt.p.StoreToken(tt.args.token); (err != nil) != tt.wantErr {
				t.Errorf("Provider.StoreToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProvider_AuthCodeURL(t *testing.T) {
	tests := []struct {
		name  string
		state string
		want  string
	}{
		// TODO: Add test cases.
		{
			name:  "success",
			state: "gdrive-state",
			want:  "https://accounts.google.com/o/oauth2/auth?client_id=12345&redirect_uri=http%3A%2F%2Flocalhost%3A4001%2Fv1%2Fredirect&response_type=code&scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fdrive.readonly&state=gdrive-state",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := createProvider(t, "")

			if got := p.AuthCodeURL(tt.state); got != tt.want {
				t.Errorf("Provider.AuthCodeURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_HttpClient(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		p       *auth.Provider
		args    args
		want    *http.Client
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.p.HttpClient(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.HttpClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Provider.HttpClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_Exchange(t *testing.T) {
	type args struct {
		ctx  context.Context
		code string
	}
	tests := []struct {
		name    string
		p       *auth.Provider
		args    args
		want    *oauth2.Token
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.p.Exchange(tt.args.ctx, tt.args.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("Provider.Exchange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Provider.Exchange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createProvider(t *testing.T, tokenPath string) *auth.Provider {
	t.Helper()

	p, err := auth.NewProvider(auth.GDriveService, tokenPath, "12345", "secret", "http://localhost:4001/v1/redirect")
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	if tokenPath == "" {
		p.TokenPath = "testdata/storage.json"
	}

	return p
}

func parseTime(timeString string) time.Time {
	t, _ := time.Parse(time.RFC3339, timeString)

	return t
}
