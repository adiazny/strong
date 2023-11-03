package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v2"
)

const (
	GDriveService ServiceID = iota
	StravaService

	defaultStravaRedirectURL = "http://localhost:4001/v1/redirect"
	stravaAuthorizeURL       = "https://www.strava.com/oauth/authorize"
	stravaTokenURL           = "https://www.strava.com/oauth/token"
	stravaScopes             = "activity:write,activity:read"
	gdriveAuthorizeURL       = "https://accounts.google.com/o/oauth2/auth"
	gdriveTokenURL           = "https://oauth2.googleapis.com/token"
)

type ServiceID int

type Provider struct {
	oauth     *oauth2.Config
	TokenPath string
}

func NewProvider(service ServiceID, tokenPath, clientID, clientSecret, redirectURL string) (*Provider, error) {
	if clientID == "" {
		return nil, errors.New("client id is required")
	}

	if clientSecret == "" {
		return nil, errors.New("client secret is required")
	}

	if redirectURL == "" {
		return nil, errors.New("redirect url is required")
	}

	var p Provider

	switch service {
	case GDriveService:
		p.oauth = &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  gdriveAuthorizeURL,
				TokenURL: gdriveTokenURL,
			},
			Scopes:      []string{drive.DriveReadonlyScope},
			RedirectURL: redirectURL,
		}
	case StravaService:
		p.oauth = &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  stravaAuthorizeURL,
				TokenURL: stravaTokenURL,
			},
			Scopes:      []string{stravaScopes},
			RedirectURL: redirectURL,
		}
	default:
		return nil, errors.New("service id for auth provider not found")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error returning the user's home directory %w", err)
	}

	p.TokenPath = filepath.Join(homeDir, tokenPath)

	return &p, nil
}

func (p *Provider) TokenNotPresent() bool {
	_, err := os.Stat(p.TokenPath)
	return err != nil
}

func (p *Provider) FileTokens() (*oauth2.Token, error) {
	file, err := os.Open(p.TokenPath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	token := &oauth2.Token{}

	err = json.NewDecoder(file).Decode(token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (p *Provider) StoreToken(token *oauth2.Token) error {
	err := os.MkdirAll(filepath.Dir(p.TokenPath), 0700)
	if err != nil {
		return fmt.Errorf("unable to create directory: %w", err)
	}

	f, err := os.OpenFile(p.TokenPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to store tokens: %w", err)
	}

	defer f.Close()

	json.NewEncoder(f).Encode(token)

	return nil
}

func (p *Provider) AuthCodeURL(state string) string {
	return p.oauth.AuthCodeURL(state)
}

func (p *Provider) HttpClient(ctx context.Context) (*http.Client, error) {
	tokens, err := p.FileTokens()
	if err != nil {
		return nil, err
	}

	return p.oauth.Client(ctx, tokens), nil
}

func (p *Provider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	tokens, err := p.oauth.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}