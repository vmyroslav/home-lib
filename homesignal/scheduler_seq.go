package homesignal

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// SequentialScheduler manages periodic signal distribution to multiple subscribers.
// This is a sequential implementation where signals are sent to subscribers
// one-by-one within the main SequentialScheduler loop.
type SequentialScheduler[T any] struct {
	logger        *zerolog.Logger
	runCancel     context.CancelFunc
	subscriptions []*JobSignal[T]
	period        time.Duration
	mu            sync.RWMutex
	buffer        uint16
	isRunning     bool
}

// NewSequentialScheduler creates a new SequentialScheduler with the given config.
func NewSequentialScheduler[T any](cfg Config) *SequentialScheduler[T] {
	return &SequentialScheduler[T]{
		period:        cfg.Period,
		buffer:        cfg.BufferSize,
		logger:        cfg.Logger,
		subscriptions: make([]*JobSignal[T], 0),
	}
}

// Subscribe creates a new subscription to the SequentialScheduler.
func (s *SequentialScheduler[T]) Subscribe() *JobSignal[T] {
	sub := NewJobSignal[T](uuid.New().String(), s.buffer)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.subscriptions = append(s.subscriptions, sub)

	return sub
}

// Unsubscribe removes a subscription from the SequentialScheduler and closes it.
func (s *SequentialScheduler[T]) Unsubscribe(sub *JobSignal[T]) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, subscription := range s.subscriptions {
		if subscription == sub {
			// remove the subscription from the slice
			s.subscriptions = append(s.subscriptions[:i], s.subscriptions[i+1:]...)

			if !subscription.IsClosed() {
				subscription.Close()
			}

			s.logger.Debug().Str("sub_id", sub.ID()).Msg("unsubscribed")

			break
		}
	}
}

// Start starts the SequentialScheduler's main loop. It will block until the provided
// context is canceled or until Stop() is called.
func (s *SequentialScheduler[T]) Start(ctx context.Context, signalFactory func() T) error {
	s.mu.Lock()

	if s.isRunning {
		s.mu.Unlock()
		return ErrSchedulerAlreadyRunning
	}

	s.isRunning = true
	runCtx, runCancel := context.WithCancel(ctx) // internal, cancellable context to allow Stop() to work
	s.runCancel = runCancel
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.isRunning = false
		s.mu.Unlock()
		s.logger.Debug().Msg("SequentialScheduler stopped")
	}()

	ticker := time.NewTicker(s.period)
	defer ticker.Stop()

	s.logger.Debug().Msg("SequentialScheduler started")

	for {
		select {
		case <-runCtx.Done():
			return nil
		case <-ticker.C:
			signalToSend := signalFactory()

			s.mu.RLock()
			// create a copy of the subscriptions to iterate over to prevent
			// holding a lock during the send loop
			subsCopy := make([]*JobSignal[T], len(s.subscriptions))
			copy(subsCopy, s.subscriptions)
			s.mu.RUnlock()

			for _, sub := range subsCopy {
				sub.Send(signalToSend)
			}

			s.logger.Debug().Int("sub_count", len(subsCopy)).Msg("tick sent")
		}
	}
}

// Stop gracefully shuts down the SequentialScheduler.
// It stops the main loop and closes all active subscriptions.
func (s *SequentialScheduler[T]) Stop() error {
	s.mu.Lock()

	if !s.isRunning {
		s.mu.Unlock()

		return nil
	}

	s.logger.Debug().Msg("stop signal received")

	if s.runCancel != nil {
		s.runCancel()
	}

	subsToClose := make([]*JobSignal[T], len(s.subscriptions))
	copy(subsToClose, s.subscriptions)

	// reset internal state for a potential restart
	s.subscriptions = make([]*JobSignal[T], 0)
	s.isRunning = false

	s.mu.Unlock()

	for _, sub := range subsToClose {
		if sub != nil && !sub.IsClosed() {
			sub.Close()
		}
	}

	s.logger.Debug().Msg("SequentialScheduler closed")

	return nil
}
