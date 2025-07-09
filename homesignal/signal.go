package homesignal

import (
	"context"
	"sync"
)

// JobSignal is a thread-safe wrapper around a channel that provides context support.
// All send operations are non-blocking and will silently drop signals that cannot be delivered
// (due to a full buffer or a canceled context).
type JobSignal[T any] struct {
	signals chan T
	id      string
	mu      sync.RWMutex
	closed  bool
}

// NewJobSignal creates a new JobSignal with the specified ID and buffer capacity.
// The buffer capacity determines how many signals can be queued before sends block or timeout.
func NewJobSignal[T any](id string, bufCap uint16) *JobSignal[T] {
	return &JobSignal[T]{
		id:      id,
		signals: make(chan T, bufCap),
		closed:  false,
	}
}

// ID returns the unique identifier of this JobSignal.
func (s *JobSignal[T]) ID() string {
	return s.id
}

// Send attempts to send a signal to the channel.
// This is a non-blocking operation. If the channel's buffer is full the signal is silently dropped.
func (s *JobSignal[T]) Send(signal T) {
	s.mu.RLock()

	if s.closed {
		s.mu.RUnlock()

		return
	}

	select {
	case s.signals <- signal: // signal sent successfully
	default: // signal dropped
	}

	s.mu.RUnlock()
}

// SendWithContext first checks if the provided context is already canceled.
// If not, it attempts to send a signal. This is a non-blocking operation;
// if the channel's buffer is full or the signal is closed, the signal is silently dropped.
func (s *JobSignal[T]) SendWithContext(ctx context.Context, signal T) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return
	}

	select {
	case <-ctx.Done():
		return
	case s.signals <- signal: // signal sent successfully
	default: // channel buffer is full, signal dropped
	}
}

// Next returns the next signal from the signal channel and identifies if the signal channel is closed.
func (s *JobSignal[T]) Next() (T, bool) {
	var zero T

	signal, ok := <-s.signals
	if !ok {
		return zero, false
	}

	return signal, true
}

// IsClosed returns true if the signal channel is closed.
func (s *JobSignal[T]) IsClosed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.closed
}

// Close closes the signal channel.
// Ideally, this method should be called only by the owner of the signal channel.
// This method is safe to call multiple times.
func (s *JobSignal[T]) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.closed {
		s.closed = true

		close(s.signals)
	}
}
