package homehttp

import (
	"net/http"
	"testing"
	"time"
)

func TestBackoffStrategies(t *testing.T) {
	testCases := []struct {
		name     string
		strategy BackoffStrategy
		min      time.Duration
		max      time.Duration
		attempt  int
		resp     *http.Response
		expected time.Duration
	}{
		{
			name:     "ConstantBackoff",
			strategy: ConstantBackoff(2 * time.Second),
			min:      0,
			max:      0,
			attempt:  0,
			resp:     nil,
			expected: 2 * time.Second,
		},
		{
			name:     "LinearBackoff",
			strategy: LinearBackoff(2 * time.Second),
			min:      1 * time.Second,
			max:      10 * time.Second,
			attempt:  2,
			resp:     nil,
			expected: 5 * time.Second,
		},
		{
			name:     "NoBackoff",
			strategy: NoBackoff(),
			min:      0,
			max:      0,
			attempt:  0,
			resp:     nil,
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.strategy.Backoff(tc.min, tc.max, tc.attempt, tc.resp)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}
