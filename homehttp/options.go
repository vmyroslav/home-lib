package homehttp

import (
	"context"
	"time"

	"github.com/rs/zerolog"

	"github.com/vmyroslav/home-lib/homeconfig"
)

// ClientOption configures the client.
type ClientOption = homeconfig.Option[clientConfig]

// WithAppName sets the user agent for the client.
func WithAppName(appName string) ClientOption {
	return homeconfig.OptionFunc[clientConfig](func(c *clientConfig) {
		c.AppName = appName
	})
}

// WithTimeout sets the timeout for the client.
func WithTimeout(timeout time.Duration) ClientOption {
	return homeconfig.OptionFunc[clientConfig](func(c *clientConfig) {
		c.Timeout = timeout
	})
}

// WithLogger sets the logger for the client.
func WithLogger(log *zerolog.Logger) ClientOption {
	return homeconfig.OptionFunc[clientConfig](func(c *clientConfig) {
		c.Logger = log
	})
}

// WithTokenProvider sets the token provider for the client.
// The token provider is used to set the Authorization header.
func WithTokenProvider(tp TokenProvider) ClientOption {
	return homeconfig.OptionFunc[clientConfig](func(c *clientConfig) {
		c.TransportMiddlewares = append(c.TransportMiddlewares, clientAuthorizationToken(tp))
	})
}

// WithAuthorizationToken sets the token for the client.
func WithAuthorizationToken(t Token) ClientOption {
	return homeconfig.OptionFunc[clientConfig](func(c *clientConfig) {
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
	return homeconfig.OptionFunc[clientConfig](func(c *clientConfig) {
		c.TransportMiddlewares = append(c.TransportMiddlewares, clientAuthorizationToken(basicAuthorization(username, password)))
	})
}

// WithHeader sets a default header for the client.
func WithHeader(key, value string) ClientOption {
	return homeconfig.OptionFunc[clientConfig](func(c *clientConfig) {
		c.TransportMiddlewares = append(c.TransportMiddlewares, clientHeader(key, value))
	})
}

// WithRetryStrategy returns a ClientOption that adds a RetryMiddleware to the client's transport middlewares.
func WithRetryStrategy(strategy RetryStrategy) ClientOption {
	return homeconfig.OptionFunc[clientConfig](func(c *clientConfig) {
		c.Retryer = strategy
	})
}

func WithMaxRetries(maxRetries int) ClientOption {
	return homeconfig.OptionFunc[clientConfig](func(c *clientConfig) {
		c.MaxRetries = maxRetries
	})
}

func WithBackoffStrategy(strategy BackoffStrategy) ClientOption {
	return homeconfig.OptionFunc[clientConfig](func(c *clientConfig) {
		c.Backoff = strategy
	})
}

func WithConstantBackoff(t time.Duration) ClientOption {
	return WithBackoffStrategy(ConstantBackoff(t))
}

// WithRetryWaitTimes sets the minimum and maximum wait times for retries.
func WithRetryWaitTimes(minWait, maxWait time.Duration) ClientOption {
	return homeconfig.OptionFunc[clientConfig](func(c *clientConfig) {
		c.MinRetryWait = minWait
		c.MaxRetryWait = maxWait
	})
}

// WithMinRetryWait sets the minimum wait time between retries.
func WithMinRetryWait(minWait time.Duration) ClientOption {
	return homeconfig.OptionFunc[clientConfig](func(c *clientConfig) {
		c.MinRetryWait = minWait
	})
}

// WithMaxRetryWait sets the maximum wait time between retries.
func WithMaxRetryWait(maxWait time.Duration) ClientOption {
	return homeconfig.OptionFunc[clientConfig](func(c *clientConfig) {
		c.MaxRetryWait = maxWait
	})
}
