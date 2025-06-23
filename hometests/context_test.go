package hometests

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestContextWithTimeout(t *testing.T) {
	t.Parallel()

	t.Run("creates context with timeout", func(t *testing.T) {
		t.Parallel()

		timeout := 100 * time.Millisecond

		ctx, cancel := ContextWithTimeout(t, timeout)
		defer cancel()

		// verify context has deadline
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Error("Expected context to have deadline")
		}

		// verify deadline is approximately correct
		expectedDeadline := time.Now().Add(timeout)
		if deadline.Sub(expectedDeadline) > 10*time.Millisecond {
			t.Errorf("Deadline %v is too far from expected %v", deadline, expectedDeadline)
		}
	})

	t.Run("context times out", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := ContextWithTimeout(t, 10*time.Millisecond)
		defer cancel()

		select {
		case <-ctx.Done():
			if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
				t.Errorf("Expected DeadlineExceeded, got %v", ctx.Err())
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Context should have timed out")
		}
	})
}

func TestWaitForContext(t *testing.T) {
	t.Parallel()

	t.Run("waits for context cancellation", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()

		// this should not fail
		WaitForContext(t, ctx, 100*time.Millisecond)
	})
}
