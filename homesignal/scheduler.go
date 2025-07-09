package homesignal

import (
	"context"
	"errors"
)

var ErrSchedulerAlreadyRunning = errors.New("scheduler is already running")

// Scheduler defines the contract for a periodic signal distributor.
type Scheduler[T any] interface {
	// Subscribe creates a new subscription to the scheduler's signals.
	Subscribe() *JobSignal[T]

	// Unsubscribe removes and closes a given subscription.
	Unsubscribe(sub *JobSignal[T])

	// Start begins the signal distribution. It blocks until the scheduler
	// is stopped via the parent context or a call to the Stop method.
	Start(ctx context.Context, signalFactory func() T) error

	// Stop gracefully shuts down the scheduler and all active subscriptions.
	Stop() error
}
