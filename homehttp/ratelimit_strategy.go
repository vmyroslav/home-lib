package homehttp

import (
	"context"
	"net/http"
	"time"
)

// RateLimitStrategy defines how rate limiting should be applied to requests.
type RateLimitStrategy interface {
	// Apply applies rate limiting before the request.
	// Returns an error if the rate limit is exceeded and behavior is set to error.
	Apply(ctx context.Context, host string) error

	// Observe allows the strategy to observe responses (for adaptive rate limiting).
	Observe(resp *http.Response)
}

// noRateLimitStrategy is a strategy that does not apply any rate limiting.
type noRateLimitStrategy struct{}

// Apply always returns nil (no rate limiting).
func (noRateLimitStrategy) Apply(context.Context, string) error {
	return nil
}

// Observe does nothing.
func (noRateLimitStrategy) Observe(*http.Response) {}

// NoRateLimitStrategy returns a strategy that does not apply any rate limiting.
func NoRateLimitStrategy() RateLimitStrategy {
	return noRateLimitStrategy{}
}

// rateLimitStrategy is a concrete implementation that wraps a scoped rate limiter.
type rateLimitStrategy struct {
	adaptive *AdaptiveRateLimiter // optional, only set if adaptive is enabled
	limiter  *ScopedRateLimiter
	behavior RateLimitBehavior
}

// Apply applies rate limiting based on the configured behavior.
// When adaptive limiting is enabled, the adaptive limiter wraps the base limiter
// and handles both adaptive backoff (from server responses) and base rate limiting.
func (s *rateLimitStrategy) Apply(ctx context.Context, host string) error {
	if s.limiter == nil {
		return nil
	}

	// If adaptive limiter is configured, check it first
	var allowed bool
	if s.adaptive != nil {
		allowed = s.adaptive.Allow(ctx)
	}

	switch s.behavior {
	case RateLimitBehaviorWait:
		// block until rate limit allows the request
		if s.adaptive != nil {
			return s.adaptive.Wait(ctx)
		}

		return s.limiter.Wait(ctx, host)
	case RateLimitBehaviorError:
		if s.adaptive != nil && !allowed {
			return ErrRateLimitExceeded
		}

		if !s.limiter.Allow(ctx, host) {
			return ErrRateLimitExceeded
		}
	}

	return nil
}

// Observe allows adaptive rate limiting to observe responses.
func (s *rateLimitStrategy) Observe(resp *http.Response) {
	if s.adaptive != nil {
		s.adaptive.ObserveResponse(resp)
	}
}

// RateLimitOption configures rate limit strategy behavior.
type RateLimitOption func(*rateLimitConfig)

type rateLimitConfig struct {
	scope    RateLimitScope
	behavior RateLimitBehavior
	adaptive bool
}

// WithScope sets the rate limiting scope (client or host).
func WithScope(scope RateLimitScope) RateLimitOption {
	return func(cfg *rateLimitConfig) {
		cfg.scope = scope
	}
}

// WithBehavior sets the rate limiting behavior (wait or error).
func WithBehavior(behavior RateLimitBehavior) RateLimitOption {
	return func(cfg *rateLimitConfig) {
		cfg.behavior = behavior
	}
}

// WithAdaptive enables adaptive rate limiting based on server responses.
func WithAdaptive() RateLimitOption {
	return func(cfg *rateLimitConfig) {
		cfg.adaptive = true
	}
}

// TokenBucketRateLimit creates a rate limit strategy using the token bucket algorithm.
// rate is the number of requests per second, burst is the maximum burst size.
func TokenBucketRateLimit(rate float64, burst int, opts ...RateLimitOption) RateLimitStrategy {
	cfg := &rateLimitConfig{
		scope:    RateLimitScopeClient,
		behavior: RateLimitBehaviorWait,
		adaptive: false,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return buildRateLimitStrategy(
		func() RateLimiter { return NewTokenBucketRateLimiter(rate, burst) },
		cfg,
	)
}

// FixedWindowRateLimit creates a rate limit strategy using a fixed window counter.
// limit is the maximum number of requests per window, window is the time window duration.
func FixedWindowRateLimit(limit int, window time.Duration, opts ...RateLimitOption) RateLimitStrategy {
	cfg := &rateLimitConfig{
		scope:    RateLimitScopeClient,
		behavior: RateLimitBehaviorWait,
		adaptive: false,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return buildRateLimitStrategy(
		func() RateLimiter { return NewFixedWindowRateLimiter(limit, window) },
		cfg,
	)
}

// CustomRateLimit creates a rate limit strategy with a custom limiter.
func CustomRateLimit(limiter RateLimiter, opts ...RateLimitOption) RateLimitStrategy {
	cfg := &rateLimitConfig{
		scope:    RateLimitScopeClient,
		behavior: RateLimitBehaviorWait,
		adaptive: false,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return buildRateLimitStrategy(
		func() RateLimiter { return limiter },
		cfg,
	)
}

// buildRateLimitStrategy constructs a rate limit strategy from a limiter factory and config.
func buildRateLimitStrategy(factory RateLimiterFactory, cfg *rateLimitConfig) RateLimitStrategy {
	var limiter RateLimiter

	var scopedLimiter *ScopedRateLimiter

	if cfg.scope == RateLimitScopeClient {
		limiter = factory()
	}

	if cfg.adaptive {
		if cfg.scope == RateLimitScopeClient {
			adaptiveLimiter := NewAdaptiveRateLimiter(limiter)
			scopedLimiter = NewScopedRateLimiter(cfg.scope, adaptiveLimiter, nil)

			return &rateLimitStrategy{
				limiter:  scopedLimiter,
				behavior: cfg.behavior,
				adaptive: adaptiveLimiter,
			}
		}

		// for per-host scope, wrap the factory
		adaptiveFactory := func() RateLimiter {
			return NewAdaptiveRateLimiter(factory())
		}
		scopedLimiter = NewScopedRateLimiter(cfg.scope, nil, adaptiveFactory)

		// Note: For per-host adaptive, we can't track per-host adaptive limiters
		// this is a known limitation - adaptive works best with per-client scope
		return &rateLimitStrategy{
			limiter:  scopedLimiter,
			behavior: cfg.behavior,
			adaptive: nil, // can't observe per-host
		}
	}

	// no adaptive wrapping
	if cfg.scope == RateLimitScopeClient {
		scopedLimiter = NewScopedRateLimiter(cfg.scope, limiter, nil)
	} else {
		scopedLimiter = NewScopedRateLimiter(cfg.scope, nil, factory)
	}

	return &rateLimitStrategy{
		limiter:  scopedLimiter,
		behavior: cfg.behavior,
		adaptive: nil,
	}
}
