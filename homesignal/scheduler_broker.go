package homesignal

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/google/uuid"
)

// BrokerScheduler manages periodic signal distribution to multiple subscribers.
// It sends signals concurrently to all subscribers to prevent head-of-line blocking
// where slow subscribers would delay signal delivery to fast subscribers.
type BrokerScheduler[T any] struct {
	logger        *zerolog.Logger
	shutdownCh    chan struct{}
	subscriptions []*JobSignal[T]
	period        time.Duration
	mu            sync.RWMutex
	buffer        uint16
	isRunning     bool
}

// NewBrokerScheduler creates a new BrokerScheduler with the given config.
func NewBrokerScheduler[T any](cfg Config) *BrokerScheduler[T] {
	s := &BrokerScheduler[T]{
		period:        cfg.Period,
		buffer:        cfg.BufferSize,
		subscriptions: make([]*JobSignal[T], 0),
		isRunning:     false,
		logger:        cfg.Logger,
		shutdownCh:    make(chan struct{}),
	}

	return s
}

// Subscribe creates a new subscription to the BrokerScheduler.
// The subscription will receive a signal on each BrokerScheduler tick.
func (s *BrokerScheduler[T]) Subscribe() *JobSignal[T] {
	sub := NewJobSignal[T](uuid.New().String(), s.buffer)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.subscriptions = append(s.subscriptions, sub)

	return sub
}

// Unsubscribe removes a subscription from the BrokerScheduler and closes it.
// This helps prevent resource leaks by properly cleaning up unused subscriptions.
func (s *BrokerScheduler[T]) Unsubscribe(sub *JobSignal[T]) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, subscription := range s.subscriptions {
		if subscription == sub {
			// remove the subscription from the slice
			s.subscriptions = append(s.subscriptions[:i], s.subscriptions[i+1:]...)

			if !subscription.IsClosed() {
				subscription.Close()
			}

			break
		}
	}
}

// Start starts the BrokerScheduler.
// It will send signals to all subscriptions concurrently on each tick.
// Signals are sent in parallel to prevent slow subscribers from blocking fast ones.
func (s *BrokerScheduler[T]) Start(ctx context.Context, signalFactory func() T) error {
	s.mu.Lock()

	if s.isRunning {
		s.mu.Unlock()

		return ErrSchedulerAlreadyRunning
	}

	s.isRunning = true
	shutdownCh := s.shutdownCh
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.isRunning = false
		s.mu.Unlock()
		s.logger.Debug().Msg("BrokerScheduler stopped")
	}()

	ticker := time.NewTicker(s.period)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Debug().Msg("context shutdown, BrokerScheduler stopping")
			return nil
		case <-shutdownCh:
			s.logger.Debug().Msg("shutdown signal received, BrokerScheduler stopping")
			return nil
		case <-ticker.C:
			signalToSend := signalFactory()

			s.mu.RLock()
			subscriptions := make([]*JobSignal[T], len(s.subscriptions))
			copy(subscriptions, s.subscriptions)
			s.mu.RUnlock()

			for _, sub := range subscriptions {
				go func(subscription *JobSignal[T]) {
					subscription.Send(signalToSend)
				}(sub)
			}

			s.logger.Debug().Msg("BrokerScheduler tick")
		}
	}
}

func (s *BrokerScheduler[T]) Stop() error {
	s.mu.Lock()

	if !s.isRunning {
		s.mu.Unlock()

		return nil
	}

	if s.shutdownCh != nil {
		close(s.shutdownCh)
		s.shutdownCh = make(chan struct{}) // create a new channel for the next start
	}

	s.isRunning = false

	subsToClose := make([]*JobSignal[T], len(s.subscriptions))
	copy(subsToClose, s.subscriptions)
	s.subscriptions = s.subscriptions[:0] // clear the original slice

	s.mu.Unlock()

	for _, sub := range subsToClose {
		if sub != nil && !sub.IsClosed() {
			sub.Close()
		}
	}

	s.logger.Debug().Msg("BrokerScheduler closed")

	return nil
}
