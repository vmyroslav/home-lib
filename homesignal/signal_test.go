package homesignal

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJobSignalClose(t *testing.T) {
	js := NewJobSignal[int]("test_close", 10)

	js.Close()
	assert.True(t, js.IsClosed())

	js.Send(100) // Send does not lead to panic in case of closed signal

	_, ok := js.Next()
	assert.False(t, ok)
}

func TestJobSignalConcurrentAccess(t *testing.T) {
	t.Parallel()

	js := NewJobSignal[string]("test_concurrent", 1000)

	var (
		wg                    sync.WaitGroup
		pg                    sync.WaitGroup
		concurrencyLevel      = 10
		signalsToSend         = 1000
		receivedSignalsAtomic int64
		sendSignals           int64
	)

	// start multiple goroutines to send signals
	pg.Add(concurrencyLevel)

	go func() {
		for i := 0; i < concurrencyLevel; i++ {
			go func() {
				defer pg.Done()

				for j := 0; j < signalsToSend/concurrencyLevel; j++ {
					js.Send(fmt.Sprintf("signal-%d-%d", i, j))
					atomic.AddInt64(&sendSignals, 1)
				}
			}()
		}
	}()

	wg.Add(concurrencyLevel)

	for i := 0; i < concurrencyLevel; i++ {
		go func() {
			defer wg.Done()

			for {
				_, ok := js.Next()
				if ok {
					atomic.AddInt64(&receivedSignalsAtomic, 1)
				} else {
					break
				}
			}
		}()
	}

	pg.Wait() // wait for all signals to be sent

	js.Close()

	wg.Wait() // wait for all signals to be received

	assert.Equal(t, sendSignals, receivedSignalsAtomic)
}

func TestJobSignalIsClosed(t *testing.T) {
	js := NewJobSignal[int]("test_is_closed", 10)
	assert.False(t, js.IsClosed())

	js.Close()
	assert.True(t, js.IsClosed())
}

func TestJobSignalID(t *testing.T) {
	id := "test_id"
	js := NewJobSignal[int](id, 10)
	assert.Equal(t, id, js.ID())
}

func TestJobSignalSendOnFullBufferDropsSignal(t *testing.T) {
	t.Parallel()

	js := NewJobSignal[int]("test_timeout", 1)

	// first send should succeed
	js.Send(1)

	// second send should be dropped immediately because buffer is full (non-blocking)
	start := time.Now()

	js.Send(2) // this should return immediately, dropping the signal

	elapsed := time.Since(start)

	assert.Less(t, elapsed, 10*time.Millisecond, "should return immediately (non-blocking)")

	// Verify only the first signal is in the buffer
	signal, ok := js.Next()
	assert.True(t, ok, "should receive first signal")
	assert.Equal(t, 1, signal, "should receive first signal value")

	// Buffer should be empty now (second signal was dropped)
	select {
	case <-js.signals:
		t.Fatal("buffer should be empty, second signal should have been dropped")
	default:
		// Expected: buffer is empty
	}
}

func TestJobSignalSendClosed(t *testing.T) {
	t.Parallel()

	js := NewJobSignal[int]("test_timeout_closed", 10)
	js.Close()

	// Send to closed signal should return immediately
	start := time.Now()

	js.Send(1)

	elapsed := time.Since(start)

	assert.Less(t, elapsed, 10*time.Millisecond, "should return immediately")
}

func TestJobSignalSendWithContext(t *testing.T) {
	t.Parallel()

	js := NewJobSignal[int]("test_context", 1)

	// First send should succeed
	ctx := context.Background()
	js.SendWithContext(ctx, 1)

	// Second send should be dropped immediately because buffer is full (non-blocking)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()

	js.SendWithContext(ctx, 2) // Should return immediately, dropping the signal

	elapsed := time.Since(start)

	// Should complete almost immediately since it's non-blocking
	assert.Less(t, elapsed, 10*time.Millisecond, "should return immediately (non-blocking)")

	// Verify only the first signal is in the buffer
	signal, ok := js.Next()
	assert.True(t, ok, "should receive first signal")
	assert.Equal(t, 1, signal, "should receive first signal value")

	// Buffer should be empty now (second signal was dropped)
	select {
	case <-js.signals:
		t.Fatal("buffer should be empty, second signal should have been dropped")
	default:
		// Expected: buffer is empty
	}
}

func TestJobSignalSendWithContextCancellation(t *testing.T) {
	t.Parallel()

	js := NewJobSignal[int]("test_context_cancel", 1)

	// Fill the buffer
	js.SendWithContext(context.Background(), 1)

	// Test with already canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	start := time.Now()

	js.SendWithContext(ctx, 2) // Should return immediately due to canceled context

	elapsed := time.Since(start)

	// Should complete almost immediately since context is already canceled
	assert.Less(t, elapsed, 10*time.Millisecond, "should return immediately with canceled context")

	// Verify only the first signal is in the buffer
	signal, ok := js.Next()
	assert.True(t, ok, "should receive first signal")
	assert.Equal(t, 1, signal, "should receive first signal value")

	// Buffer should be empty now (second signal was dropped due to canceled context)
	select {
	case <-js.signals:
		t.Fatal("buffer should be empty, second signal should have been dropped")
	default:
		// Expected: buffer is empty
	}
}

func TestJobSignalSendWithContextClosed(t *testing.T) {
	t.Parallel()

	js := NewJobSignal[int]("test_context_closed", 10)
	js.Close()

	// Send to closed signal should return immediately
	ctx := context.Background()
	start := time.Now()

	js.SendWithContext(ctx, 1)

	elapsed := time.Since(start)

	assert.Less(t, elapsed, 10*time.Millisecond, "should return immediately")
}

func TestJobSignalConcurrentSendWithContext(t *testing.T) {
	t.Parallel()

	js := NewJobSignal[int]("test_concurrent_context", 1)

	var wg sync.WaitGroup

	// Fill the buffer with one item
	js.Send(0)

	concurrency := 10
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func(val int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			js.SendWithContext(ctx, val)
		}(i)
	}

	wg.Wait()

	// Test passes if no panics or deadlocks occur
}
