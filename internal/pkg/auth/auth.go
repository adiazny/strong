package auth

import (
	"context"
	"errors"
	"net/http"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v2"
)

const (
	GDriveService serviceID = iota
	StravaService

	defaultStravaRedirectURL = "http://localhost:4001/v1/redirect"
	stravaAuthorizeURL       = "https://www.strava.com/oauth/authorize"
	stravaTokenURL           = "https://www.strava.com/oauth/token"
	stravaScopes             = "activity:write,activity:read"
	gdriveAuthorizeURL       = "https://accounts.google.com/o/oauth2/auth"
	gdriveTokenURL           = "https://oauth2.googleapis.com/token"
)

type serviceID int

type Provider struct {
	Config  *oauth2.Config
	Storage Storage
}

func NewProvider(service serviceID, tokenPath, client, secrect, redirect string, store Storage) (*Provider, error) {
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
		p.Config = &oauth2.Config{
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
		p.Config = &oauth2.Config{
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

	p.Storage = store

	return &p, nil
}

func (p *Provider) AuthCodeURL(state string) string {
	return p.Config.AuthCodeURL(state)
}

// Exchange stores a token after retrieval
func (p *Provider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := p.Config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	if err := p.Storage.SetToken(token); err != nil {
		return nil, err
	}

	return token, nil
}

// TokenSource can be passed a token which
// is stored, or when a new one is retrieved,
// that's stored
func (p *Provider) TokenSource(ctx context.Context, t *oauth2.Token) oauth2.TokenSource {
	return StorageTokenSource(ctx, p, t)
}

// Client is attached to the TokenSource
func (p *Provider) Client(ctx context.Context, t *oauth2.Token) *http.Client {
	return oauth2.NewClient(ctx, p.TokenSource(ctx, t))
}
