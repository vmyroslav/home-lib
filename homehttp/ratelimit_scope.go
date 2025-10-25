package homehttp

import (
	"context"
	"sync"
)

// RateLimitScope defines the scope of rate limiting.
type RateLimitScope int

const (
	// RateLimitScopeClient applies rate limiting per client instance.
	// Each Client instance has its own independent rate limiter.
	RateLimitScopeClient RateLimitScope = iota

	// RateLimitScopeHost applies rate limiting per host/domain.
	// Requests to the same host share the same rate limiter across client instances.
	RateLimitScopeHost
)

// RateLimiterFactory is a function that creates a new RateLimiter instance.
type RateLimiterFactory func() RateLimiter

// PerHostRateLimiter manages rate limiters on a per-host basis.
// It creates and caches a separate rate limiter for each unique host.
type PerHostRateLimiter struct {
	factory  RateLimiterFactory
	limiters sync.Map // map[string]RateLimiter
}

// NewPerHostRateLimiter creates a new per-host rate limiter manager.
// The factory function is called to create a new limiter for each unique host.
func NewPerHostRateLimiter(factory RateLimiterFactory) *PerHostRateLimiter {
	return &PerHostRateLimiter{
		factory: factory,
	}
}

// Allow checks if a request to the specified host is allowed without blocking.
func (p *PerHostRateLimiter) Allow(ctx context.Context, host string) bool {
	limiter := p.getLimiterForHost(host)

	return limiter.Allow(ctx)
}

// Wait blocks until a request to the specified host can proceed or the context is canceled.
func (p *PerHostRateLimiter) Wait(ctx context.Context, host string) error {
	limiter := p.getLimiterForHost(host)

	return limiter.Wait(ctx)
}

// getLimiterForHost returns the rate limiter for the specified host.
// If no limiter exists for the host, a new one is created using the factory.
func (p *PerHostRateLimiter) getLimiterForHost(host string) RateLimiter {
	if limiter, ok := p.limiters.Load(host); ok {
		rl, ok := limiter.(RateLimiter)
		if !ok {
			panic("ratelimit: stored value is not a RateLimiter")
		}

		return rl
	}

	newLimiter := p.factory()

	actual, _ := p.limiters.LoadOrStore(host, newLimiter)

	rl, ok := actual.(RateLimiter)
	if !ok {
		panic("ratelimit: stored value is not a RateLimiter")
	}

	return rl
}

// ScopedRateLimiter wraps rate limiting with scope awareness.
// It supports both per-client and per-host rate limiting.
type ScopedRateLimiter struct {
	clientLimit RateLimiter
	hostLimit   *PerHostRateLimiter
	scope       RateLimitScope
}

// NewScopedRateLimiter creates a new scoped rate limiter.
// For per-client scope, provide a limiter instance.
// For per-host scope, provide a factory function.
func NewScopedRateLimiter(scope RateLimitScope, limiter RateLimiter, factory RateLimiterFactory) *ScopedRateLimiter {
	sl := &ScopedRateLimiter{
		scope: scope,
	}

	switch scope {
	case RateLimitScopeClient:
		sl.clientLimit = limiter
	case RateLimitScopeHost:
		sl.hostLimit = NewPerHostRateLimiter(factory)
	}

	return sl
}

// Allow checks if a request to the specified host is allowed without blocking.
func (s *ScopedRateLimiter) Allow(ctx context.Context, host string) bool {
	switch s.scope {
	case RateLimitScopeClient:
		if s.clientLimit == nil {
			return true
		}

		return s.clientLimit.Allow(ctx)
	case RateLimitScopeHost:
		if s.hostLimit == nil {
			return true
		}

		return s.hostLimit.Allow(ctx, host)
	default:
		return true
	}
}

// Wait blocks until a request to the specified host can proceed or the context is canceled.
func (s *ScopedRateLimiter) Wait(ctx context.Context, host string) error {
	switch s.scope {
	case RateLimitScopeClient:
		if s.clientLimit == nil {
			return nil
		}

		return s.clientLimit.Wait(ctx)
	case RateLimitScopeHost:
		if s.hostLimit == nil {
			return nil
		}

		return s.hostLimit.Wait(ctx, host)
	default:
		return nil
	}
}
