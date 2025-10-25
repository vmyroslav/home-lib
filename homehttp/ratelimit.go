package homehttp

import (
	"context"
	"errors"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// ErrRateLimitExceeded is returned when the rate limit is exceeded and blocking is disabled.
var ErrRateLimitExceeded = errors.New("rate limit exceeded")

// RateLimiter defines the interface for rate limiting strategies.
type RateLimiter interface {
	// Allow checks if a request is allowed without blocking.
	// Returns true if the request can proceed immediately.
	Allow(ctx context.Context) bool

	// Wait blocks until the request can proceed or the context is canceled.
	// Returns an error if the context is canceled.
	Wait(ctx context.Context) error
}

// NoRateLimit is a rate limiter that never limits requests.
type NoRateLimit struct{}

// Allow always returns true.
func (NoRateLimit) Allow(context.Context) bool {
	return true
}

// Wait always returns nil immediately.
func (NoRateLimit) Wait(context.Context) error {
	return nil
}

// TokenBucketRateLimiter implements rate limiting using the token bucket algorithm.
type TokenBucketRateLimiter struct {
	limiter *rate.Limiter
}

// NewTokenBucketRateLimiter creates a new token bucket rate limiter.
// ratePerSecond is the number of requests per second, burst is the maximum burst size.
//
// Parameters:
//   - ratePerSecond: should be positive. Use rate.Inf for no limit.
//   - burst: should be >= 1. A burst of 0 means no requests can ever succeed.
func NewTokenBucketRateLimiter(ratePerSecond float64, burst int) *TokenBucketRateLimiter {
	return &TokenBucketRateLimiter{
		limiter: rate.NewLimiter(rate.Limit(ratePerSecond), burst),
	}
}

// Allow checks if a request is allowed without blocking.
func (tb *TokenBucketRateLimiter) Allow(_ context.Context) bool {
	return tb.limiter.Allow()
}

// Wait blocks until a token is available or the context is canceled.
func (tb *TokenBucketRateLimiter) Wait(ctx context.Context) error {
	return tb.limiter.Wait(ctx)
}

// FixedWindowRateLimiter implements rate limiting using a fixed window counter.
// It allows a fixed number of requests per time window.
type FixedWindowRateLimiter struct {
	windowStart time.Time
	limit       int
	window      time.Duration
	count       int
	mu          sync.Mutex
}

// NewFixedWindowRateLimiter creates a new fixed window rate limiter.
// limit is the maximum number of requests per window, window is the time window duration.
//
// Parameters:
//   - limit: maximum requests per window. Should be >= 1. A limit of 0 blocks all requests.
//   - window: time window duration. Should be positive. Invalid values are clamped to minimum.
func NewFixedWindowRateLimiter(limit int, window time.Duration) *FixedWindowRateLimiter {
	// clamp to sensible minimums to avoid division by zero or unexpected behavior
	if limit < 0 {
		limit = 0
	}

	if window <= 0 {
		window = time.Nanosecond // clamp to minimum valid window to avoid division by zero
	}

	return &FixedWindowRateLimiter{
		limit:       limit,
		window:      window,
		count:       0,
		windowStart: time.Now(),
	}
}

// Allow checks if a request is allowed without blocking.
func (fw *FixedWindowRateLimiter) Allow(_ context.Context) bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	fw.maybeResetWindow()

	if fw.count < fw.limit {
		fw.count++

		return true
	}

	return false
}

// Wait blocks until the next window or the context is canceled.
func (fw *FixedWindowRateLimiter) Wait(ctx context.Context) error {
	for {
		fw.mu.Lock()
		fw.maybeResetWindow()

		if fw.count < fw.limit {
			fw.count++
			fw.mu.Unlock()

			return nil
		}

		waitTime := fw.window - time.Since(fw.windowStart)
		fw.mu.Unlock()

		// ensure wait time is non-negative
		if waitTime < 0 {
			waitTime = 0
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// loop back to check if window has reset
		}
	}
}

// maybeResetWindow resets the window if it has expired.
func (fw *FixedWindowRateLimiter) maybeResetWindow() {
	now := time.Now()

	if now.Sub(fw.windowStart) >= fw.window {
		fw.count = 0
		fw.windowStart = now
	}
}
