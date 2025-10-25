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

// WithRateLimitStrategy configures rate limiting with a custom strategy.
func WithRateLimitStrategy(strategy RateLimitStrategy) ClientOption {
	return homeconfig.OptionFunc[clientConfig](func(c *clientConfig) {
		c.RateLimitStrategy = strategy
	})
}

// WithTokenBucketRateLimit configures rate limiting using the token bucket algorithm.
// rate is the number of requests per second, burst is the maximum burst size.
//
// Options can be provided to customize behavior:
//   - WithScope(RateLimitScopeHost) - apply rate limiting per host
//   - WithBehavior(RateLimitBehaviorError) - fail fast instead of blocking
//   - WithAdaptive() - enable adaptive rate limiting based on server responses
func WithTokenBucketRateLimit(rate float64, burst int, opts ...RateLimitOption) ClientOption {
	return homeconfig.OptionFunc[clientConfig](func(c *clientConfig) {
		c.RateLimitStrategy = TokenBucketRateLimit(rate, burst, opts...)
	})
}

// WithFixedWindowRateLimit configures rate limiting using a fixed window counter.
// limit is the maximum number of requests per window, window is the time window duration.
//
// Options can be provided to customize behavior:
//   - WithScope(RateLimitScopeHost) - apply rate limiting per host
//   - WithBehavior(RateLimitBehaviorError) - fail fast instead of blocking
//   - WithAdaptive() - enable adaptive rate limiting based on server responses
func WithFixedWindowRateLimit(limit int, window time.Duration, opts ...RateLimitOption) ClientOption {
	return homeconfig.OptionFunc[clientConfig](func(c *clientConfig) {
		c.RateLimitStrategy = FixedWindowRateLimit(limit, window, opts...)
	})
}

// WithPerHostTokenBucketRateLimit configures per-host rate limiting using the token bucket algorithm.
// It applies RateLimitScopeHost by default.
// Each unique host will have its own independent rate limiter.
//
// rate is the number of requests per second, burst is the maximum burst size.
//
// Additional options can be provided:
//   - WithBehavior(RateLimitBehaviorError) - fail fast instead of blocking
//   - WithAdaptive() - enable adaptive rate limiting based on server responses
func WithPerHostTokenBucketRateLimit(rate float64, burst int, opts ...RateLimitOption) ClientOption {
	allOpts := append([]RateLimitOption{WithScope(RateLimitScopeHost)}, opts...)
	return WithTokenBucketRateLimit(rate, burst, allOpts...)
}

// WithPerHostFixedWindowRateLimit configures per-host rate limiting using a fixed window counter.
// It applies RateLimitScopeHost by default.
// Each unique host will have its own independent rate limiter.
//
// limit is the maximum number of requests per window, window is the time window duration.
//
// Additional options can be provided:
//   - WithBehavior(RateLimitBehaviorError) - fail fast instead of blocking
//   - WithAdaptive() - enable adaptive rate limiting based on server responses
func WithPerHostFixedWindowRateLimit(limit int, window time.Duration, opts ...RateLimitOption) ClientOption {
	// Prepend WithScope(RateLimitScopeHost) to the options
	allOpts := append([]RateLimitOption{WithScope(RateLimitScopeHost)}, opts...)
	return WithFixedWindowRateLimit(limit, window, allOpts...)
}

// WithoutRateLimit disables rate limiting for the client.
func WithoutRateLimit() ClientOption {
	return homeconfig.OptionFunc[clientConfig](func(c *clientConfig) {
		c.RateLimitStrategy = NoRateLimitStrategy()
	})
}
