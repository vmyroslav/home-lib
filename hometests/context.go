package hometests

import (
	"context"
	"testing"
	"time"
)

// ContextWithTimeout creates a context with timeout for testing.
func ContextWithTimeout(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	t.Cleanup(cancel)

	return ctx, cancel
}

// WaitForContext waits for a context to be done or times out.
// Useful for testing context cancellation behavior.
func WaitForContext(t *testing.T, ctx context.Context, timeout time.Duration) { //nolint:revive
	t.Helper()

	select {
	case <-ctx.Done():
		// Context was canceled as expected
		return
	case <-time.After(timeout):
		t.Fatalf("context was not canceled within %v", timeout)
	}
}
