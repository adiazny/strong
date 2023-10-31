package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v2"
)

const (
	GDriveService = iota
	StravaService

	defaultStravaRedirectURL = "http://localhost:4001/v1/redirect"
	stravaAuthorizeURL       = "https://www.strava.com/oauth/authorize"
	stravaTokenURL           = "https://www.strava.com/oauth/token"
	stravaScopes             = "activity:write,activity:read"
	gdriveAuthorizeURL       = "https://accounts.google.com/o/oauth2/auth"
	gdriveTokenURL           = "https://oauth2.googleapis.com/token"
)

type Provider struct {
	oauth *oauth2.Config
}

func NewProvider(service int, client, secrect, redirect string) (*Provider, error) {
	if client == "" {
		return nil, errors.New("client id is required")
	}

	if secrect == "" {
		return nil, errors.New("client secret is required")
	}

	if redirect == "" {
		return nil, errors.New("redirect url is required")
	}

	var p Provider

	switch service {
	case GDriveService:
		p.oauth = &oauth2.Config{
			ClientID:     client,
			ClientSecret: secrect,
			Endpoint: oauth2.Endpoint{
				AuthURL:  gdriveAuthorizeURL,
				TokenURL: gdriveTokenURL,
			},
			Scopes:      []string{drive.DriveReadonlyScope},
			RedirectURL: redirect,
		}
	case StravaService:
		p.oauth = &oauth2.Config{
			ClientID:     client,
			ClientSecret: secrect,
			Endpoint: oauth2.Endpoint{
				AuthURL:  stravaAuthorizeURL,
				TokenURL: stravaTokenURL,
			},
			Scopes:      []string{stravaScopes},
			RedirectURL: redirect,
		}
	default:
		return nil, errors.New("service not found")
	}

	return &p, nil
}

func (p *Provider) Exists(fileName string) bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	path := path.Join(homeDir, fileName)

	_, err = os.Stat(path)
	return err == nil
}

func (p *Provider) FileTokens(fileName string) (*oauth2.Token, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	tok := &oauth2.Token{}

	err = json.NewDecoder(f).Decode(tok)
	if err != nil {
		return nil, err
	}

	return tok, nil
}

func (p *Provider) StoreToken(path string, token *oauth2.Token) error {
	err := os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		return fmt.Errorf("unable to create directory: %v", err)
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)

	if err != nil {
		return fmt.Errorf("unable to store tokens: %v", err)
	}

	defer f.Close()

	json.NewEncoder(f).Encode(token)

	return nil
}

func (p *Provider) AuthCodeURL(state string) string {
	return p.oauth.AuthCodeURL(state)
}

func (p *Provider) HttpClient(ctx context.Context, path string) (*http.Client, error) {
	tokens, err := p.FileTokens(path)
	if err != nil {
		return nil, err
	}

	return p.oauth.Client(ctx, tokens), nil
}
