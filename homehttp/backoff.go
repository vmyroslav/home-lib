package homehttp

import (
	"net/http"
	"time"
)

type BackoffStrategy interface {
	Backoff(minT, maxT time.Duration, attemptNum int, resp *http.Response) time.Duration
}

type BackoffStrategyFunc func(minT, maxT time.Duration, attemptNum int, resp *http.Response) time.Duration

func (f BackoffStrategyFunc) Backoff(minT, maxT time.Duration, attemptNum int, resp *http.Response) time.Duration {
	return f(minT, maxT, attemptNum, resp)
}

func ConstantBackoff(t time.Duration) BackoffStrategyFunc {
	return func(_, _ time.Duration, _ int, _ *http.Response) time.Duration {
		return t
	}
}

func LinearBackoff(t time.Duration) BackoffStrategyFunc {
	return func(minT, _ time.Duration, attemptNum int, _ *http.Response) time.Duration {
		return minT + time.Duration(attemptNum)*t
	}
}

func NoBackoff() BackoffStrategyFunc {
	return func(_, _ time.Duration, _ int, _ *http.Response) time.Duration {
		return 0
	}
}
