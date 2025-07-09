package homesignal

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestScheduler_StartStop(t *testing.T) {
	testBothImplementations(t, func(t *testing.T, s Scheduler[struct{}], _ string) {
		t.Helper()

		require.NotNil(t, s)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			err := s.Start(ctx, func() struct{} { return struct{}{} })
			assert.NoError(t, err)
		}()

		sub := s.Subscribe()

		_, _ = sub.Next() // wait for the first tick
		err := s.Stop()
		require.NoError(t, err)

		assert.True(t, sub.IsClosed(), "subscription should be closed")
	})
}

func TestScheduler_Subscribe(t *testing.T) {
	testBothImplementations(t, func(t *testing.T, s Scheduler[struct{}], _ string) {
		t.Helper()

		sub := s.Subscribe()
		assert.NotNil(t, sub)
	})
}

func TestScheduler_Tick(t *testing.T) {
	testBothImplementations(t, func(t *testing.T, s Scheduler[struct{}], _ string) {
		t.Helper()

		sub := s.Subscribe()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			_ = s.Start(ctx, func() struct{} { return struct{}{} })
		}()

		_, ok := sub.Next()
		assert.True(t, ok, "expected a tick")
	})
}

func TestScheduler_StartTwice(t *testing.T) {
	testBothImplementations(t, func(t *testing.T, s Scheduler[struct{}], _ string) {
		t.Helper()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sub := s.Subscribe()

		go func() {
			err := s.Start(ctx, func() struct{} { return struct{}{} })
			assert.NoError(t, err)
		}()

		_, _ = sub.Next() // wait for the first tick

		err := s.Start(ctx, func() struct{} { return struct{}{} })
		require.Error(t, err)
		assert.Equal(t, ErrSchedulerAlreadyRunning, err)
	})
}

func TestScheduler_Unsubscribe(t *testing.T) {
	testBothImplementations(t, func(t *testing.T, s Scheduler[struct{}], _ string) {
		t.Helper()

		sub1 := s.Subscribe()
		sub2 := s.Subscribe()

		s.Unsubscribe(sub1)

		assert.True(t, sub1.IsClosed(), "sub1 should be closed after unsubscribe")
		assert.False(t, sub2.IsClosed(), "sub2 should not be closed")

		// test that unsubscribing a non-existent subscription doesn't panic
		s.Unsubscribe(sub1) // should be safe to call again

		require.NoError(t, s.Stop())
	})
}

// TestBrokerScheduler_ConcurrentSignalDelivery tests that the BrokerScheduler can handle concurrent signal delivery
func TestBrokerScheduler_ConcurrentSignalDelivery(t *testing.T) {
	cfg := NewConfig(
		WithPeriod(50*time.Millisecond),
		WithBufferSize(1), // small buffer to test blocking behavior
	)
	s := NewBrokerScheduler[struct{}](cfg)

	// create fast and slow subscribers
	fastSub := s.Subscribe()
	slowSub := s.Subscribe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = s.Start(ctx, func() struct{} { return struct{}{} })
	}()

	// fill the slow subscriber's buffer and don't read from it (making it "slow")
	slowSub.Next() // clear initial signal if any

	// collect fast subscriber signals
	fastSignals := make(chan struct{}, 10)
	done := make(chan struct{})

	// read from fast subscriber in a separate goroutine
	go func() {
		defer close(done)

		for {
			_, ok := fastSub.Next()
			if !ok {
				break
			}

			select {
			case fastSignals <- struct{}{}:
			default:
			}
		}
	}()

	// wait for multiple ticks and count signals received by fast subscriber
	time.Sleep(200 * time.Millisecond) // Allow 4 ticks

	_ = s.Stop()

	// wait for the goroutine to finish before closing the channel
	<-done

	close(fastSignals)

	fastCount := 0
	for range fastSignals {
		fastCount++
	}

	// fast subscriber should have received multiple signals despite slow subscriber
	assert.GreaterOrEqual(t, fastCount, 2, "fast subscriber should receive multiple signals even with slow subscriber")
}

// TestBrokerScheduler_SlowSubscriberDoesNotBlockSchedulerPeriod tests that a slow subscriber does not block the scheduler's period
func TestBrokerScheduler_SlowSubscriberDoesNotBlockSchedulerPeriod(t *testing.T) {
	cfg := NewConfig(
		WithPeriod(20*time.Millisecond),
		WithBufferSize(1),
	)
	s := NewBrokerScheduler[struct{}](cfg)

	// create one normal subscriber and one slow subscriber
	normalSub := s.Subscribe()
	slowSub := s.Subscribe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = s.Start(ctx, func() struct{} { return struct{}{} })
	}()

	// let the scheduler start and fill the slow subscriber's buffer
	time.Sleep(100 * time.Millisecond)

	// read one signal from slowSub to fill it's buffer, then don't read anymore
	slowSub.Next()

	// track scheduler tick timing via normal subscriber
	tickTimes := make(chan time.Time, 10)
	done := make(chan struct{})

	// monitor normal subscriber to track scheduler ticks
	go func() {
		defer close(done)

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			_, ok := normalSub.Next()
			if !ok {
				return
			}

			select {
			case tickTimes <- time.Now():
			case <-ctx.Done():
				return
			default:
			}
		}
	}()

	time.Sleep(400 * time.Millisecond)

	_ = s.Stop()

	cancel()

	<-done

	close(tickTimes)

	times := make([]time.Time, 0, 10)
	for t := range tickTimes {
		times = append(times, t)
	}

	// should have received multiple ticks
	assert.GreaterOrEqual(t, len(times), 2, "should receive multiple ticks")

	if len(times) >= 2 {
		var maxInterval time.Duration

		for i := 1; i < len(times); i++ {
			interval := times[i].Sub(times[i-1])
			if interval > maxInterval {
				maxInterval = interval
			}
		}

		assert.LessOrEqual(t, maxInterval, 500*time.Millisecond,
			"max tick interval should not be excessively delayed by slow subscriber")
	}
}

// TestSequentialScheduler_OrderedDelivery tests that the scheduler delivers signals in a predictable order
func TestSequentialScheduler_OrderedDelivery(t *testing.T) {
	t.Parallel()

	cfg := NewConfig(
		WithPeriod(50*time.Millisecond),
		WithBufferSize(10),
	)
	s := NewSequentialScheduler[int](cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	counter := 0

	go func() {
		_ = s.Start(ctx, func() int {
			counter++
			return counter
		})
	}()

	// subscribe and verify we get sequential values
	sub := s.Subscribe()

	for i := 1; i <= 3; i++ {
		val, ok := sub.Next()
		assert.True(t, ok)
		assert.Equal(t, i, val, "should receive values in sequential order")
	}

	_ = s.Stop()
}

// TestBrokerScheduler_BroadcastsToAll tests that the BrokerScheduler broadcasts signals to all subscribers
func TestBrokerScheduler_BroadcastsToAll(t *testing.T) {
	t.Parallel()

	const (
		numSubscribers = 3
		testTimeout    = 200 * time.Millisecond
	)

	cfg := NewConfig(
		WithPeriod(100*time.Millisecond),
		WithBufferSize(5),
	)
	s := NewBrokerScheduler[int](cfg)

	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		cancel()

		_ = s.Stop()
	}()

	go func() { _ = s.Start(ctx, func() int { return 42 }) }()

	subs := make([]*JobSignal[int], numSubscribers)
	for i := 0; i < numSubscribers; i++ {
		subs[i] = s.Subscribe()
	}

	var wg sync.WaitGroup
	wg.Add(numSubscribers)

	for i := 0; i < numSubscribers; i++ {
		go func(subIndex int, sub *JobSignal[int]) {
			defer wg.Done()

			val, ok := sub.Next()

			assert.True(t, ok, "subscriber %d should have received a signal", subIndex+1)
			assert.Equal(t, 42, val, "subscriber %d received incorrect value", subIndex+1)
		}(i, subs[i])
	}

	waitChan := make(chan struct{})

	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case <-waitChan: // all subscribers received a signal
	case <-time.After(testTimeout):
		// at least one subscriber never received the signal
		t.Fatal("Test timed out: not all subscribers received the broadcast signal.")
	}
}

// testBothImplementations helper function to run tests on both implementations
func testBothImplementations(t *testing.T, testFunc func(t *testing.T, s Scheduler[struct{}], name string)) {
	t.Helper()

	cfg := NewConfig(
		WithPeriod(10*time.Millisecond),
		WithBufferSize(10),
	)

	t.Run("BrokerScheduler", func(t *testing.T) {
		t.Parallel()

		s := NewBrokerScheduler[struct{}](cfg)
		testFunc(t, s, "BrokerScheduler")
	})

	t.Run("SequentialScheduler", func(t *testing.T) {
		t.Parallel()

		s := NewSequentialScheduler[struct{}](cfg)
		testFunc(t, s, "SequentialScheduler")
	})
}
