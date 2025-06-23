package homehttp

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"
)

// Token is a struct that holds the authorization token information.
type Token struct {
	AccessToken string
	ExpiresAt   time.Time
	Type        string
}

// IsValid checks if the token is valid.
func (t Token) IsValid() bool {
	if t.AccessToken == "" {
		return false
	}

	if t.ExpiresAt.IsZero() {
		return true
	} // if ExpiresAt is zero time, token doesn't expire (e.g., basic auth)

	return time.Now().Before(t.ExpiresAt)
}

// TokenProvider is a token provider.
type TokenProvider interface {
	GetToken(ctx context.Context) (Token, error)
}

// TokenProviderFunc is a token provider function.
type TokenProviderFunc func(ctx context.Context) (Token, error)

func (f TokenProviderFunc) GetToken(ctx context.Context) (Token, error) {
	return f(ctx)
}

// basicAuthorization provides basic authorization TokenProvider.
func basicAuthorization(username, password string) TokenProvider {
	token := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))

	return TokenProviderFunc(func(context.Context) (Token, error) {
		return Token{
			AccessToken: token,
			ExpiresAt:   time.Time{},
			Type:        "Basic",
		}, nil
	})
}
