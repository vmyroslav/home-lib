package homehttp

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

var _ http.RoundTripper = (*roundTripperFunc)(nil)

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

type roundTripperMiddleware func(http.RoundTripper) http.RoundTripper

func chainRoundTrippers(base http.RoundTripper, middlewares ...roundTripperMiddleware) http.RoundTripper {
	rt := base

	if len(middlewares) > 0 {
		rt = middlewares[len(middlewares)-1](base)

		for i := len(middlewares) - 2; i >= 0; i-- {
			rt = middlewares[i](rt)
		}
	}

	return rt
}

// clientUserAgent adds a User-Agent header to the request.
func clientUserAgent(userAgent string) roundTripperMiddleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			req.Header.Set("User-Agent", userAgent)

			return next.RoundTrip(req)
		})
	}
}

// clientHeader adds a header to the request.
func clientHeader(key, value string) roundTripperMiddleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			req.Header.Set(key, value)

			return next.RoundTrip(req)
		})
	}
}

// clientAuthorizationToken adds an Authorization header to the request.
func clientAuthorizationToken(tp TokenProvider) roundTripperMiddleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			token, err := tp.GetToken(req.Context())
			if err != nil {
				return nil, errors.WithStack(err)
			}

			req.Header.Set("Authorization", fmt.Sprintf("%s %s", token.Type, token.AccessToken))

			return next.RoundTrip(req)
		})
	}
}

// RateLimitBehavior defines how the rate limiter behaves when the limit is reached.
type RateLimitBehavior int

const (
	// RateLimitBehaviorWait blocks until the rate limit allows the request.
	RateLimitBehaviorWait RateLimitBehavior = iota

	// RateLimitBehaviorError returns an error immediately if the rate limit is exceeded.
	RateLimitBehaviorError
)

// clientRateLimitStrategy adds rate limiting using a RateLimitStrategy.
func clientRateLimitStrategy(strategy RateLimitStrategy) roundTripperMiddleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			host := req.URL.Host

			// apply rate limiting before the request
			if err := strategy.Apply(req.Context(), host); err != nil {
				return nil, err
			}

			resp, err := next.RoundTrip(req)

			if resp != nil {
				strategy.Observe(resp)
			}

			return resp, err
		})
	}
}
