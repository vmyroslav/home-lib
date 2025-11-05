package homehttp

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// HTTP headers for rate limiting
const (
	// X-RateLimit-* headers
	headerXRateLimitLimit     = "X-RateLimit-Limit"
	headerXRateLimitRemaining = "X-RateLimit-Remaining"
	headerXRateLimitReset     = "X-RateLimit-Reset"

	// RateLimit-* headers (IETF draft standard)
	headerRateLimitLimit     = "RateLimit-Limit"
	headerRateLimitRemaining = "RateLimit-Remaining"
	headerRateLimitReset     = "RateLimit-Reset"

	// Standard HTTP headers
	headerRetryAfter = "Retry-After"
)

// AdaptiveRateLimiter wraps a rate limiter and dynamically adjusts limits based on API responses.
// It monitors rate limit headers and 429 responses to optimize throughput while respecting server limits.
type AdaptiveRateLimiter struct {
	backoffUntil      time.Time
	base              RateLimiter
	lastObservedLimit int
	mu                sync.RWMutex
}

// NewAdaptiveRateLimiter creates a new adaptive rate limiter that wraps the base limiter.
func NewAdaptiveRateLimiter(base RateLimiter) *AdaptiveRateLimiter {
	return &AdaptiveRateLimiter{
		base: base,
	}
}

// Allow checks if a request is allowed, considering both base limiter and adaptive backoff.
func (a *AdaptiveRateLimiter) Allow(ctx context.Context) bool {
	a.mu.RLock()
	backoffUntil := a.backoffUntil
	a.mu.RUnlock()

	// check if we're in adaptive backoff period
	if time.Now().Before(backoffUntil) {
		return false
	}

	return a.base.Allow(ctx)
}

// Wait blocks until a token is available, considering both base limiter and adaptive backoff.
func (a *AdaptiveRateLimiter) Wait(ctx context.Context) error {
	a.mu.RLock()
	backoffUntil := a.backoffUntil
	a.mu.RUnlock()

	// wait for adaptive backoff to expire
	if time.Now().Before(backoffUntil) {
		waitDuration := time.Until(backoffUntil)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitDuration):
			// continue to base limiter
		}
	}

	return a.base.Wait(ctx)
}

// ObserveResponse analyzes the HTTP response for rate limit information and adjusts accordingly.
func (a *AdaptiveRateLimiter) ObserveResponse(resp *http.Response) {
	if resp == nil {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// handle 429 Too Many Requests
	if resp.StatusCode == http.StatusTooManyRequests {
		a.handle429Response(resp) // still parse headers even on 429
	}

	// parse rate limit headers to understand current limits
	a.parseRateLimitHeaders(resp)
}

// handle429Response processes a 429 response and applies appropriate backoff.
func (a *AdaptiveRateLimiter) handle429Response(resp *http.Response) {
	if retryAfter := resp.Header.Get(headerRetryAfter); retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			// retry-After in seconds
			a.backoffUntil = time.Now().Add(time.Duration(seconds) * time.Second)

			return
		}

		// try parsing as HTTP date
		if t, err := http.ParseTime(retryAfter); err == nil {
			a.backoffUntil = t

			return
		}
	}

	if reset := resp.Header.Get(headerXRateLimitReset); reset != "" {
		if timestamp, err := strconv.ParseInt(reset, 10, 64); err == nil {
			a.backoffUntil = time.Unix(timestamp, 0)

			return
		}
	}

	if reset := resp.Header.Get(headerRateLimitReset); reset != "" {
		if timestamp, err := strconv.ParseInt(reset, 10, 64); err == nil {
			a.backoffUntil = time.Unix(timestamp, 0)

			return
		}
	}

	// default backoff if no headers available: 60 seconds
	a.backoffUntil = time.Now().Add(60 * time.Second)
}

// parseRateLimitHeaders extracts rate limit information from response headers.
func (a *AdaptiveRateLimiter) parseRateLimitHeaders(resp *http.Response) {
	// X-RateLimit-Limit
	if limit := resp.Header.Get(headerXRateLimitLimit); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			a.lastObservedLimit = l
		}
	}

	// RateLimit-Limit (IETF draft standard)
	if limit := resp.Header.Get(headerRateLimitLimit); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			a.lastObservedLimit = l
		}
	}

	a.checkRemainingAndSetBackoff(resp)
}

// checkRemainingAndSetBackoff checks if X-RateLimit-Remaining is zero and sets backoff if needed.
func (a *AdaptiveRateLimiter) checkRemainingAndSetBackoff(resp *http.Response) {
	remaining := resp.Header.Get(headerXRateLimitRemaining)
	if remaining == "" {
		return
	}

	r, err := strconv.Atoi(remaining)
	if err != nil || r != 0 {
		return
	}

	// No requests remaining, check reset time
	reset := resp.Header.Get(headerXRateLimitReset)
	if reset == "" {
		return
	}

	timestamp, err := strconv.ParseInt(reset, 10, 64)
	if err == nil {
		a.backoffUntil = time.Unix(timestamp, 0)
	}
}

// GetLastObservedLimit returns the last observed rate limit from API headers.
// Returns 0 if no limit has been observed yet.
func (a *AdaptiveRateLimiter) GetLastObservedLimit() int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.lastObservedLimit
}
