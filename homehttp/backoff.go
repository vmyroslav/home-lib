package homehttp

import (
	"net/http"
	"time"
)

type BackoffStrategy interface {
	Backoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration
}

type BackoffStrategyFunc func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration

func (f BackoffStrategyFunc) Backoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration { //nolint:gocritic
	return f(min, max, attemptNum, resp)
}

func ConstantBackoff(t time.Duration) BackoffStrategyFunc {
	return func(_, _ time.Duration, _ int, _ *http.Response) time.Duration {
		return t
	}
}

func LinearBackoff(t time.Duration) BackoffStrategyFunc {
	return func(min, _ time.Duration, attemptNum int, _ *http.Response) time.Duration {
		return min + time.Duration(attemptNum)*t
	}
}

func NoBackoff() BackoffStrategyFunc {
	return func(_, _ time.Duration, _ int, _ *http.Response) time.Duration {
		return 0
	}
}
