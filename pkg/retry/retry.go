// Package retry provides a flexible retry mechanism with exponential backoff and jitter.
//
// This package is designed to handle transient failures in network requests,
// database operations, or any other operation that might fail temporarily.
// It implements exponential backoff with optional jitter to avoid thundering herd problems.
//
// Example usage:
//
//	opts := retry.Options{
//		MaxAttempts:    5,
//		InitialBackoff: 100 * time.Millisecond,
//		MaxBackoff:     10 * time.Second,
//		Jitter:         true,
//	}
//
//	err := retry.Do(opts, func() error {
//		return makeNetworkRequest()
//	})
package retry

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// Version represents the current package version following semantic versioning.
const Version = "1.0.0"

// ErrMaxAttemptsReached is returned when the operation fails after all retry attempts.
var ErrMaxAttemptsReached = errors.New("maximum retry attempts reached")

// Options configures the retry mechanism.
type Options struct {
	// MaxAttempts is the maximum number of retry attempts.
	// Default: 3
	MaxAttempts int

	// InitialBackoff is the initial backoff duration.
	// Default: 100ms
	InitialBackoff time.Duration

	// MaxBackoff is the maximum backoff duration.
	// Default: 10s
	MaxBackoff time.Duration

	// Jitter determines whether to add randomness to backoff durations.
	// Adding jitter helps avoid retry storms when multiple clients are retrying.
	// Default: true
	Jitter bool

	// OnRetry is called before each retry attempt with the attempt number and error.
	// It can be used for logging or other side effects.
	// Optional.
	OnRetry func(attempt int, err error)
}

// DefaultOptions returns the default retry options.
func DefaultOptions() Options {
	return Options{
		MaxAttempts:    3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     10 * time.Second,
		Jitter:         true,
	}
}

// Do retries the provided operation based on the retry options.
// It returns nil if the operation succeeds, or an error if all retries fail.
func Do(opts Options, operation func() error) error {
	return DoWithContext(context.Background(), opts, func(ctx context.Context) error {
		return operation()
	})
}

// DoWithContext retries the provided operation with context support.
// The operation can be canceled via the context.
func DoWithContext(ctx context.Context, opts Options, operation func(ctx context.Context) error) error {
	backoff := opts.InitialBackoff
	if backoff <= 0 {
		backoff = DefaultOptions().InitialBackoff
	}

	maxBackoff := opts.MaxBackoff
	if maxBackoff <= 0 {
		maxBackoff = DefaultOptions().MaxBackoff
	}

	maxAttempts := opts.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = DefaultOptions().MaxAttempts
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("operation canceled: %w", ctx.Err())
		default:
			// Continue with retry
		}

		err := operation(ctx)
		if err == nil {
			return nil
		}

		lastErr = err

		if attempt == maxAttempts {
			return fmt.Errorf("%w: %v", ErrMaxAttemptsReached, lastErr)
		}

		if opts.OnRetry != nil {
			opts.OnRetry(attempt, err)
		}

		// Calculate backoff duration
		nextBackoff := calculateBackoff(backoff, maxBackoff, opts.Jitter)

		// Wait for backoff duration or until context is canceled
		timer := time.NewTimer(nextBackoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return fmt.Errorf("operation canceled during backoff: %w", ctx.Err())
		case <-timer.C:
			// Continue with next attempt
		}

		backoff = nextBackoff * 2
	}

	return fmt.Errorf("%w: %v", ErrMaxAttemptsReached, lastErr)
}

// calculateBackoff calculates the next backoff duration with optional jitter.
func calculateBackoff(currentBackoff, maxBackoff time.Duration, jitter bool) time.Duration {
	nextBackoff := currentBackoff
	if nextBackoff > maxBackoff {
		nextBackoff = maxBackoff
	}

	if jitter {
		nextBackoff = addJitter(nextBackoff)
	}

	return nextBackoff
}

// addJitter applies random jitter to the backoff duration.
// It returns a duration between 50% and 100% of the input duration.
func addJitter(duration time.Duration) time.Duration {
	jitter := time.Duration(rand.Int63n(int64(duration) / 2))
	return duration - jitter
}
