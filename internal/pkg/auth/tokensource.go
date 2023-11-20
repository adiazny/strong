package auth

import (
	"context"

	"golang.org/x/oauth2"
)

type storageTokenSource struct {
	*Provider
	oauth2.TokenSource
}

// Token satisfies the oauth2.TokenSource interface
func (s *storageTokenSource) Token() (*oauth2.Token, error) {
	if token, err := s.Provider.Storage.GetToken(); err == nil && token.Valid() {
		return token, err
	}

	token, err := s.TokenSource.Token()
	if err != nil {
		return token, err
	}

	if err := s.Provider.Storage.SetToken(token); err != nil {
		return nil, err
	}

	return token, nil
}

// StorageTokenSource will be used by provider.TokenSource method
func StorageTokenSource(ctx context.Context, provider *Provider, t *oauth2.Token) oauth2.TokenSource {
	if t == nil || !t.Valid() {
		if tok, err := provider.Storage.GetToken(); err == nil {
			t = tok
		}
	}

	ts := provider.Config.TokenSource(ctx, t)
	return &storageTokenSource{provider, ts}
}
