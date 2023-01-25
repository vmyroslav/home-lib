package homehttp

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

// ClientOption configures the client.
type ClientOption interface {
	apply(c *clientConfig)
}

type clientOptionFn func(c *clientConfig)

func (f clientOptionFn) apply(c *clientConfig) {
	f(c)
}

// WithAppName sets the user agent for the client.
func WithAppName(appName string) ClientOption {
	return clientOptionFn(func(c *clientConfig) {
		c.AppName = appName
	})
}

// WithTimeout sets the timeout for the client.
func WithTimeout(timeout time.Duration) ClientOption {
	return clientOptionFn(func(c *clientConfig) {
		c.Timeout = timeout
	})
}

// WithLogger sets the logger for the client.
func WithLogger(log *zerolog.Logger) ClientOption {
	return clientOptionFn(func(c *clientConfig) {
		c.Logger = log
	})
}

// WithTokenProvider sets the token provider for the client.
// The token provider is used to set the Authorization header.
func WithTokenProvider(tp TokenProvider) ClientOption {
	return clientOptionFn(func(c *clientConfig) {
		c.TransportMiddlewares = append(c.TransportMiddlewares, clientAuthorizationToken(tp))
	})
}

// WithAuthorizationToken sets the token for the client.
func WithAuthorizationToken(t Token) ClientOption {
	return clientOptionFn(func(c *clientConfig) {
		c.TransportMiddlewares = append(
			c.TransportMiddlewares,
			clientAuthorizationToken(TokenProviderFunc(func(context.Context) (Token, error) {
				return t, nil
			})),
		)
	})
}

// WithBasicAuth sets the basic auth token for the client.
func WithBasicAuth(username, password string) ClientOption {
	return clientOptionFn(func(c *clientConfig) {
		c.TransportMiddlewares = append(c.TransportMiddlewares, clientAuthorizationToken(basicAuthorization(username, password)))
	})
}

// WithHeader sets a default header for the client.
func WithHeader(key, value string) ClientOption {
	return clientOptionFn(func(c *clientConfig) {
		c.TransportMiddlewares = append(c.TransportMiddlewares, clientHeader(key, value))
	})
}
