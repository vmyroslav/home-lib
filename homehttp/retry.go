package homehttp

import (
	"context"
	"net/http"
)

// RetryStrategy classifies the response and error into retry decision.
type RetryStrategy interface {
	Classify(ctx context.Context, resp *http.Response, err error) bool
}

// RetryStrategyFunc is a function that implements RetryStrategy.
type RetryStrategyFunc func(ctx context.Context, resp *http.Response, err error) bool

func (f RetryStrategyFunc) Classify(ctx context.Context, resp *http.Response, err error) bool {
	return f(ctx, resp, err)
}

// MultiRetryStrategies is a classifier that combines multiple classifiers.
type MultiRetryStrategies []RetryStrategy

// Classify implements retrier.Classifier.
func (s MultiRetryStrategies) Classify(ctx context.Context, resp *http.Response, err error) bool {
	for _, strategy := range s {
		if decision := strategy.Classify(ctx, resp, err); decision {
			return decision
		}
	}

	return false
}

var (
	// NoRetry is a classifier that never retries.
	NoRetry = RetryStrategyFunc(func(context.Context, *http.Response, error) bool {
		return false
	})

	// RetryOn500x returns a classifier that retries on HTTP errors (5xx).
	RetryOn500x = RetryStrategyFunc(func(ctx context.Context, resp *http.Response, err error) bool {
		return resp != nil && resp.StatusCode >= http.StatusInternalServerError
	})
)

type NoRetryStrategy struct{}

func (s *NoRetryStrategy) Classify(_ context.Context, _ *http.Response, _ error) bool {
	return false
}
