package homehttp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoRateLimit(t *testing.T) {
	t.Parallel()

	limiter := NoRateLimit{}
	ctx := context.Background()

	assert.True(t, limiter.Allow(ctx))
	assert.True(t, limiter.Allow(ctx))
	assert.True(t, limiter.Allow(ctx))

	assert.NoError(t, limiter.Wait(ctx))
}

func TestTokenBucketRateLimiter_Allow(t *testing.T) {
	t.Parallel()

	t.Run("allows up to burst capacity", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(1, 3)
		ctx := t.Context()

		assert.True(t, limiter.Allow(ctx))
		assert.True(t, limiter.Allow(ctx))
		assert.True(t, limiter.Allow(ctx))

		// 4th request should be denied (no tokens left)
		assert.False(t, limiter.Allow(ctx))
	})

	t.Run("refills tokens over time", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(10, 2)
		ctx := context.Background()

		// Consume all tokens
		assert.True(t, limiter.Allow(ctx))
		assert.True(t, limiter.Allow(ctx))
		assert.False(t, limiter.Allow(ctx))

		// Wait for refill (100ms should give us 1 token at 10 req/s)
		time.Sleep(110 * time.Millisecond)

		// Should allow one more request
		assert.True(t, limiter.Allow(ctx))
		assert.False(t, limiter.Allow(ctx))
	})
}

func TestTokenBucketRateLimiter_Wait(t *testing.T) {
	t.Parallel()

	t.Run("waits for token availability", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(10, 1) // 10 req/s, burst of 1
		ctx := context.Background()

		// Consume the token
		assert.True(t, limiter.Allow(ctx))

		// Wait should block until token is available
		start := time.Now()
		err := limiter.Wait(ctx)
		elapsed := time.Since(start)

		require.NoError(t, err)
		// Should wait approximately 100ms (1/10 second)
		assert.GreaterOrEqual(t, elapsed, 90*time.Millisecond, "expected to wait at least 90ms, got %v", elapsed)
		assert.Less(t, elapsed, 200*time.Millisecond, "expected to wait less than 200ms, got %v", elapsed)
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(1, 1) // 1 req/s, burst of 1

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		// Consume the token
		assert.True(t, limiter.Allow(ctx))

		// Wait should return error when context times out
		// The stdlib rate.Limiter returns its own error message but respects context
		err := limiter.Wait(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline")
	})
}

func TestTokenBucketRateLimiter_Concurrent(t *testing.T) {
	t.Parallel()

	limiter := NewTokenBucketRateLimiter(100, 10) // 100 req/s, burst of 10
	ctx := context.Background()

	var (
		allowed           atomic.Int32
		denied            atomic.Int32
		wg                sync.WaitGroup
		numWorkers        = 20
		requestsPerWorker = 5
	)

	// Launch concurrent workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for j := 0; j < requestsPerWorker; j++ {
				if limiter.Allow(ctx) {
					allowed.Add(1)
				} else {
					denied.Add(1)
				}
			}
		}()
	}

	wg.Wait()

	// Should have allowed exactly 10 requests (burst capacity)
	assert.Equal(t, int32(10), allowed.Load())
	assert.Equal(t, int32(90), denied.Load())
}

func TestFixedWindowRateLimiter_Allow(t *testing.T) {
	t.Parallel()

	t.Run("allows up to limit per window", func(t *testing.T) {
		limiter := NewFixedWindowRateLimiter(3, time.Second) // 3 req/s
		ctx := context.Background()

		// Should allow first 3 requests
		assert.True(t, limiter.Allow(ctx))
		assert.True(t, limiter.Allow(ctx))
		assert.True(t, limiter.Allow(ctx))

		// 4th request should be denied
		assert.False(t, limiter.Allow(ctx))
	})

	t.Run("resets window after duration", func(t *testing.T) {
		limiter := NewFixedWindowRateLimiter(2, 100*time.Millisecond) // 2 req per 100ms
		ctx := context.Background()

		// Consume limit
		assert.True(t, limiter.Allow(ctx))
		assert.True(t, limiter.Allow(ctx))
		assert.False(t, limiter.Allow(ctx))

		// Wait for window to reset
		time.Sleep(110 * time.Millisecond)

		// Should allow requests again
		assert.True(t, limiter.Allow(ctx))
		assert.True(t, limiter.Allow(ctx))
		assert.False(t, limiter.Allow(ctx))
	})
}

func TestFixedWindowRateLimiter_Wait(t *testing.T) {
	t.Parallel()

	t.Run("waits for next window", func(t *testing.T) {
		limiter := NewFixedWindowRateLimiter(1, 100*time.Millisecond) // 1 req per 100ms
		ctx := context.Background()

		// Consume the limit
		assert.True(t, limiter.Allow(ctx))

		// Wait should block until next window
		start := time.Now()
		err := limiter.Wait(ctx)
		elapsed := time.Since(start)

		require.NoError(t, err)
		// Should wait for window reset
		assert.GreaterOrEqual(t, elapsed, 90*time.Millisecond, "expected to wait at least 90ms, got %v", elapsed)
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		limiter := NewFixedWindowRateLimiter(1, time.Second)

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		// Consume the limit
		assert.True(t, limiter.Allow(ctx))

		// Wait should return error when context is canceled
		err := limiter.Wait(ctx)
		require.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
	})
}

func TestPerHostRateLimiter(t *testing.T) {
	t.Parallel()

	t.Run("creates separate limiters per host", func(t *testing.T) {
		factory := func() RateLimiter {
			return NewTokenBucketRateLimiter(10, 2) // 10 req/s, burst of 2
		}

		hostLimiter := NewPerHostRateLimiter(factory)
		ctx := context.Background()

		// Consume limit for host1
		assert.True(t, hostLimiter.Allow(ctx, "host1.com"))
		assert.True(t, hostLimiter.Allow(ctx, "host1.com"))
		assert.False(t, hostLimiter.Allow(ctx, "host1.com"))

		// host2 should have its own limit
		assert.True(t, hostLimiter.Allow(ctx, "host2.com"))
		assert.True(t, hostLimiter.Allow(ctx, "host2.com"))
		assert.False(t, hostLimiter.Allow(ctx, "host2.com"))
	})

	t.Run("reuses limiter for same host", func(t *testing.T) {
		factory := func() RateLimiter {
			return NewTokenBucketRateLimiter(10, 1) // 10 req/s, burst of 1
		}

		hostLimiter := NewPerHostRateLimiter(factory)
		ctx := context.Background()

		// Consume limit for host
		assert.True(t, hostLimiter.Allow(ctx, "host.com"))
		assert.False(t, hostLimiter.Allow(ctx, "host.com"))

		// Same host should still be limited
		assert.False(t, hostLimiter.Allow(ctx, "host.com"))
	})
}

func TestScopedRateLimiter(t *testing.T) {
	t.Parallel()

	t.Run("client scope", func(t *testing.T) {
		limiter := NewTokenBucketRateLimiter(10, 2)
		scoped := NewScopedRateLimiter(RateLimitScopeClient, limiter, nil)
		ctx := context.Background()

		// Should use same limiter regardless of host
		assert.True(t, scoped.Allow(ctx, "host1.com"))
		assert.True(t, scoped.Allow(ctx, "host2.com"))
		assert.False(t, scoped.Allow(ctx, "host3.com"))
	})

	t.Run("host scope", func(t *testing.T) {
		factory := func() RateLimiter {
			return NewTokenBucketRateLimiter(10, 2)
		}
		scoped := NewScopedRateLimiter(RateLimitScopeHost, nil, factory)
		ctx := context.Background()

		// Each host should have independent limit
		assert.True(t, scoped.Allow(ctx, "host1.com"))
		assert.True(t, scoped.Allow(ctx, "host1.com"))
		assert.False(t, scoped.Allow(ctx, "host1.com"))

		assert.True(t, scoped.Allow(ctx, "host2.com"))
		assert.True(t, scoped.Allow(ctx, "host2.com"))
		assert.False(t, scoped.Allow(ctx, "host2.com"))
	})
}

func TestAdaptiveRateLimiter_429Response(t *testing.T) {
	t.Parallel()

	t.Run("respects Retry-After header in seconds", func(t *testing.T) {
		base := NoRateLimit{}
		adaptive := NewAdaptiveRateLimiter(base)
		ctx := context.Background()

		// Simulate 429 response with Retry-After
		resp := &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header: http.Header{
				headerRetryAfter: []string{"1"}, // 1 second
			},
		}

		adaptive.ObserveResponse(resp)

		// Should be blocked immediately after 429
		assert.False(t, adaptive.Allow(ctx))

		// Wait for backoff to expire
		time.Sleep(1100 * time.Millisecond)

		// Should allow again
		assert.True(t, adaptive.Allow(ctx))
	})

	t.Run("respects X-RateLimit-Reset header", func(t *testing.T) {
		base := NoRateLimit{}
		adaptive := NewAdaptiveRateLimiter(base)
		ctx := context.Background()

		// Set reset time to 1 second in the future
		resetTime := time.Now().Add(time.Second).Unix()

		resp := &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header: http.Header{
				headerXRateLimitReset: []string{strconv.FormatInt(resetTime, 10)},
			},
		}

		adaptive.ObserveResponse(resp)

		// Should be blocked
		assert.False(t, adaptive.Allow(ctx))
	})

	t.Run("uses default backoff when no headers present", func(t *testing.T) {
		base := NoRateLimit{}
		adaptive := NewAdaptiveRateLimiter(base)
		ctx := context.Background()

		resp := &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header:     http.Header{},
		}

		adaptive.ObserveResponse(resp)

		// Should be blocked
		assert.False(t, adaptive.Allow(ctx))
	})
}

func TestAdaptiveRateLimiter_ParseHeaders(t *testing.T) {
	t.Parallel()

	t.Run("parses X-RateLimit-Limit header", func(t *testing.T) {
		base := NoRateLimit{}
		adaptive := NewAdaptiveRateLimiter(base)

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
		}
		resp.Header.Set(headerXRateLimitLimit, "100")

		adaptive.ObserveResponse(resp)

		assert.Equal(t, 100, adaptive.GetLastObservedLimit())
	})

	t.Run("handles X-RateLimit-Remaining zero", func(t *testing.T) {
		base := NoRateLimit{}
		adaptive := NewAdaptiveRateLimiter(base)
		ctx := context.Background()

		resetTime := time.Now().Add(time.Second).Unix()

		resp := &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
		}
		resp.Header.Set(headerXRateLimitRemaining, "0")
		resp.Header.Set(headerXRateLimitReset, strconv.FormatInt(resetTime, 10))

		adaptive.ObserveResponse(resp)

		// Should be blocked when remaining is 0
		assert.False(t, adaptive.Allow(ctx))
	})
}

func TestClientWithRateLimit_Integration(t *testing.T) {
	t.Parallel()

	t.Run("token bucket rate limiting", func(t *testing.T) {
		var requestCount atomic.Int32

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			requestCount.Add(1)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(
			WithTokenBucketRateLimit(10, 3, WithBehavior(RateLimitBehaviorError)),
		)

		ctx := context.Background()

		// Should allow burst of 3 requests
		for i := 0; i < 3; i++ {
			resp, err := client.DoJSON(ctx, http.MethodGet, server.URL, nil)
			require.NoError(t, err)
			require.NotNil(t, resp)
			resp.Body.Close()
		}

		// 4th request should fail with rate limit error
		resp, err := client.DoJSON(ctx, http.MethodGet, server.URL, nil)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrRateLimitExceeded)
		require.Nil(t, resp)

		// Only 3 requests should have reached the server
		assert.Equal(t, int32(3), requestCount.Load())
	})

	t.Run("rate limiting with wait behavior", func(t *testing.T) {
		var requestCount atomic.Int32

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			requestCount.Add(1)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(
			WithTokenBucketRateLimit(5, 2, WithBehavior(RateLimitBehaviorWait)),
		)

		ctx := context.Background()

		start := time.Now()

		// Make 3 requests (2 immediate, 1 should wait)
		for i := 0; i < 3; i++ {
			resp, err := client.DoJSON(ctx, http.MethodGet, server.URL, nil)
			require.NoError(t, err)
			require.NotNil(t, resp)
			resp.Body.Close()
		}

		elapsed := time.Since(start)

		// Should have taken at least 200ms (waiting for 1 token at 5 req/s)
		assert.GreaterOrEqual(t, elapsed, 180*time.Millisecond, "expected to wait at least 180ms, got %v", elapsed)
		assert.Equal(t, int32(3), requestCount.Load())
	})

	t.Run("per-host rate limiting", func(t *testing.T) {
		var server1Count, server2Count atomic.Int32

		server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			server1Count.Add(1)
			w.WriteHeader(http.StatusOK)
		}))
		defer server1.Close()

		server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			server2Count.Add(1)
			w.WriteHeader(http.StatusOK)
		}))
		defer server2.Close()

		client := NewClient(
			WithTokenBucketRateLimit(10, 2, WithScope(RateLimitScopeHost), WithBehavior(RateLimitBehaviorError)),
		)

		ctx := context.Background()

		// Server 1: burst of 2
		for i := 0; i < 2; i++ {
			resp, err := client.DoJSON(ctx, http.MethodGet, server1.URL, nil)
			require.NoError(t, err)
			resp.Body.Close()
		}

		// Server 2: should have independent burst of 2
		for i := 0; i < 2; i++ {
			resp, err := client.DoJSON(ctx, http.MethodGet, server2.URL, nil)
			require.NoError(t, err)
			resp.Body.Close()
		}

		// Both servers should have received 2 requests
		assert.Equal(t, int32(2), server1Count.Load())
		assert.Equal(t, int32(2), server2Count.Load())

		// 3rd request to server1 should fail
		resp, err := client.DoJSON(ctx, http.MethodGet, server1.URL, nil)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrRateLimitExceeded)
		require.Nil(t, resp)
	})

	t.Run("adaptive rate limiting with 429", func(t *testing.T) {
		var (
			requestCount atomic.Int32
			return429    atomic.Bool
		)

		return429.Store(true)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			requestCount.Add(1)

			if return429.Load() {
				w.Header().Set(headerRetryAfter, "1")
				w.WriteHeader(http.StatusTooManyRequests)

				return
			}

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(
			WithTokenBucketRateLimit(100, 10, WithBehavior(RateLimitBehaviorError), WithAdaptive()),
		)

		ctx := context.Background()

		// First request gets 429
		resp, _ := client.DoJSON(ctx, http.MethodGet, server.URL, nil)
		if resp != nil {
			resp.Body.Close()
		}

		// Subsequent requests should be blocked by adaptive limiter
		_, err := client.DoJSON(ctx, http.MethodGet, server.URL, nil)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrRateLimitExceeded)

		// Stop returning 429
		return429.Store(false)

		// Wait for backoff to expire
		time.Sleep(1100 * time.Millisecond)

		// Should work now
		resp, err = client.DoJSON(ctx, http.MethodGet, server.URL, nil)
		require.NoError(t, err)
		require.NotNil(t, resp)
		resp.Body.Close()
	})

	t.Run("respects 429 with Retry-After header in seconds", func(t *testing.T) {
		var requestCount atomic.Int32

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			count := requestCount.Add(1)

			if count == 1 {
				// First request returns 429 with Retry-After
				w.Header().Set(headerRetryAfter, "1")
				w.WriteHeader(http.StatusTooManyRequests)

				return
			}

			// Subsequent requests succeed
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(
			WithTokenBucketRateLimit(100, 10, WithBehavior(RateLimitBehaviorWait), WithAdaptive()),
		)

		ctx := context.Background()

		// First request gets 429
		resp, _ := client.DoJSON(ctx, http.MethodGet, server.URL, nil)
		if resp != nil {
			resp.Body.Close()
		}

		// Immediately try again - should be blocked by adaptive limiter
		start := time.Now()
		resp, err := client.DoJSON(ctx, http.MethodGet, server.URL, nil)
		elapsed := time.Since(start)

		require.NoError(t, err)
		require.NotNil(t, resp)
		resp.Body.Close()

		// Should have waited approximately 1 second as specified in Retry-After
		assert.GreaterOrEqual(t, elapsed, 950*time.Millisecond, "should wait for Retry-After duration")
		assert.LessOrEqual(t, elapsed, 1200*time.Millisecond, "should not wait much longer than Retry-After")

		// Should have made exactly 2 requests
		assert.Equal(t, int32(2), requestCount.Load())
	})

	t.Run("blocks subsequent requests after 429 until retry time passes", func(t *testing.T) {
		ctx := context.Background()

		// Create adaptive limiter and simulate a 429 response
		baseLimiter := NewTokenBucketRateLimiter(100, 10)
		adaptiveLimiter := NewAdaptiveRateLimiter(baseLimiter)

		// Simulate 429 response with 1 second retry
		resp429 := &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Header:     make(http.Header),
		}
		resp429.Header.Set(headerRetryAfter, "1")
		adaptiveLimiter.ObserveResponse(resp429)

		// Immediately after observing 429, requests should be blocked
		allowed := adaptiveLimiter.Allow(ctx)
		assert.False(t, allowed, "should block immediately after 429")

		// After retry time passes, should allow again
		time.Sleep(1100 * time.Millisecond)

		allowed = adaptiveLimiter.Allow(ctx)
		assert.True(t, allowed, "should allow after retry time passes")
	})

	t.Run("respects X-RateLimit-Reset header on 429", func(t *testing.T) {
		var requestCount atomic.Int32

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			count := requestCount.Add(1)

			if count == 1 {
				// Return 429 with X-RateLimit-Reset (Unix timestamp)
				// Use 2 seconds to ensure enough margin for processing
				// Round to nearest second to avoid truncation issues
				resetTime := time.Now().Add(2 * time.Second).Unix()
				w.Header().Set(headerXRateLimitReset, strconv.FormatInt(resetTime, 10))
				w.WriteHeader(http.StatusTooManyRequests)

				return
			}

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(
			WithTokenBucketRateLimit(100, 10, WithBehavior(RateLimitBehaviorWait), WithAdaptive()),
		)

		ctx := context.Background()

		// First request gets 429
		resp, _ := client.DoJSON(ctx, http.MethodGet, server.URL, nil)
		if resp != nil {
			resp.Body.Close()
		}

		// Second request should wait
		start := time.Now()
		resp, err := client.DoJSON(ctx, http.MethodGet, server.URL, nil)
		elapsed := time.Since(start)

		require.NoError(t, err)
		require.NotNil(t, resp)
		resp.Body.Close()

		// Should wait at least 1 second (accounting for Unix second truncation and processing time)
		// We set 2s but due to truncation and timing, expect at least 900ms
		assert.GreaterOrEqual(t, elapsed, 900*time.Millisecond, "should wait for reset time")
	})

	t.Run("parses X-RateLimit-Limit from successful responses", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set(headerXRateLimitLimit, "1000")
			w.Header().Set(headerXRateLimitRemaining, "999")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Create adaptive limiter directly for inspection
		baseLimiter := NewTokenBucketRateLimiter(100, 10)
		adaptiveLimiter := NewAdaptiveRateLimiter(baseLimiter)

		// Make request
		req, err := http.NewRequest(http.MethodGet, server.URL, http.NoBody)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		defer resp.Body.Close()

		// Observe the response
		adaptiveLimiter.ObserveResponse(resp)

		// Check that limit was parsed
		assert.Equal(t, 1000, adaptiveLimiter.GetLastObservedLimit())
	})

	t.Run("handles X-RateLimit-Remaining zero with backoff", func(t *testing.T) {
		var requestCount atomic.Int32

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			count := requestCount.Add(1)

			if count == 1 {
				// Return success but with remaining=0
				resetTime := time.Now().Add(1 * time.Second).Unix()

				w.Header().Set(headerXRateLimitLimit, "100")
				w.Header().Set(headerXRateLimitRemaining, "0")
				w.Header().Set(headerXRateLimitReset, strconv.FormatInt(resetTime, 10))
				w.WriteHeader(http.StatusOK)

				return
			}

			w.Header().Set(headerXRateLimitRemaining, "99")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(
			WithTokenBucketRateLimit(100, 10, WithBehavior(RateLimitBehaviorWait), WithAdaptive()),
		)

		ctx := context.Background()

		// First request succeeds but exhausts limit
		resp, err := client.DoJSON(ctx, http.MethodGet, server.URL, nil)
		require.NoError(t, err)
		require.NotNil(t, resp)
		resp.Body.Close()

		// Second request should wait for reset time
		start := time.Now()
		resp, err = client.DoJSON(ctx, http.MethodGet, server.URL, nil)
		elapsed := time.Since(start)

		require.NoError(t, err)
		require.NotNil(t, resp)
		resp.Body.Close()

		assert.GreaterOrEqual(t, elapsed, 950*time.Millisecond)
		assert.Equal(t, int32(2), requestCount.Load())
	})

	t.Run("uses default 60s backoff when no headers present on 429", func(t *testing.T) {
		var requestCount atomic.Int32

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			count := requestCount.Add(1)

			if count == 1 {
				// Return 429 with no retry headers
				w.WriteHeader(http.StatusTooManyRequests)

				return
			}

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient(
			WithTokenBucketRateLimit(100, 10, WithBehavior(RateLimitBehaviorError), WithAdaptive()),
		)

		ctx := context.Background()

		// First request gets 429
		_, _ = client.DoJSON(ctx, http.MethodGet, server.URL, nil)

		// Immediately try again - should be blocked for ~60 seconds
		// We'll use error behavior to test without waiting
		_, err := client.DoJSON(ctx, http.MethodGet, server.URL, nil)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrRateLimitExceeded)

		// Only 1 request should have been made (second was blocked client-side)
		assert.Equal(t, int32(1), requestCount.Load())
	})

	t.Run("supports IETF draft RateLimit headers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			// IETF draft standard headers
			w.Header().Set(headerRateLimitLimit, "500")
			w.Header().Set(headerRateLimitRemaining, "499")
			w.Header().Set(headerRateLimitReset, strconv.FormatInt(time.Now().Add(60*time.Second).Unix(), 10))
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		baseLimiter := NewTokenBucketRateLimiter(100, 10)
		adaptiveLimiter := NewAdaptiveRateLimiter(baseLimiter)

		req, err := http.NewRequest(http.MethodGet, server.URL, http.NoBody)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		defer resp.Body.Close()

		adaptiveLimiter.ObserveResponse(resp)

		// Should parse IETF standard limit
		assert.Equal(t, 500, adaptiveLimiter.GetLastObservedLimit())
	})

	t.Run("concurrent requests with adaptive limiting", func(t *testing.T) {
		var requestCount atomic.Int32

		var blockedCount atomic.Int32

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			requestCount.Add(1)
			// Always return 429 with 5 second backoff
			w.Header().Set(headerRetryAfter, "5")
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer server.Close()

		client := NewClient(
			WithTokenBucketRateLimit(100, 50, WithBehavior(RateLimitBehaviorError), WithAdaptive()),
		)

		ctx := context.Background()

		// First request gets 429 and sets adaptive backoff
		_, _ = client.DoJSON(ctx, http.MethodGet, server.URL, nil)

		// Give it a moment to process the response
		time.Sleep(50 * time.Millisecond)

		// Launch multiple concurrent requests - should all be blocked by adaptive limiter
		var wg atomic.Int32
		for i := 0; i < 10; i++ {
			wg.Add(1)

			go func() {
				defer wg.Add(-1)

				_, err := client.DoJSON(ctx, http.MethodGet, server.URL, nil)
				if err != nil {
					// Check if it's rate limit exceeded using errors.Is for wrapped errors
					if errors.Is(err, ErrRateLimitExceeded) {
						blockedCount.Add(1)
					}
				}
			}()
		}

		// Wait for all goroutines
		for wg.Load() > 0 {
			time.Sleep(10 * time.Millisecond)
		}

		// All requests should have been blocked client-side
		assert.Equal(t, int32(10), blockedCount.Load(), "adaptive limiter should block all concurrent requests")
		assert.Equal(t, int32(1), requestCount.Load(), "should only make 1 server request (the first one)")
	})
}
