package retry

import (
	"fmt"
	"math/rand"
	"time"
)

// RetryOptions configures the retry mechanism
type RetryOptions struct {
	MaxAttempts    int           // Maximum number of retry attempts
	InitialBackoff time.Duration // Initial backoff duration
	MaxBackoff     time.Duration // Maximum backoff duration
	Jitter         bool          // Add jitter to avoid collision
}

// Retry retries the provided operation based on the retry options
func Retry(opts RetryOptions, operation func() error) error {
	backoff := opts.InitialBackoff

	for attempt := 1; attempt <= opts.MaxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		if attempt == opts.MaxAttempts {
			return fmt.Errorf("operation failed after %d attempts: %w", attempt, err)
		}

		fmt.Printf("Attempt %d failed: %v. Retrying in %v...\n", attempt, err, backoff)

		time.Sleep(backoff)
		backoff = calculateBackoff(backoff, opts.MaxBackoff, opts.Jitter)
	}

	return nil
}

// calculateBackoff calculates the next backoff duration with optional jitter
func calculateBackoff(currentBackoff, maxBackoff time.Duration, jitter bool) time.Duration {
	nextBackoff := currentBackoff * 2
	if nextBackoff > maxBackoff {
		nextBackoff = maxBackoff
	}

	if jitter {
		nextBackoff = addJitter(nextBackoff)
	}

	return nextBackoff
}

// addJitter applies random jitter to the backoff duration
func addJitter(duration time.Duration) time.Duration {
	jitter := time.Duration(rand.Int63n(int64(duration / 2)))
	return duration - jitter
}
