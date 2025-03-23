package retry

import (
	"errors"
	"testing"
	"time"
)

func TestRetry_Success(t *testing.T) {
	// Test that a successful operation doesn't retry
	opts := RetryOptions{
		MaxAttempts:    3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Jitter:         false,
	}

	attemptCount := 0
	operation := func() error {
		attemptCount++
		return nil // Success on first attempt
	}

	err := Retry(opts, operation)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if attemptCount != 1 {
		t.Errorf("Expected 1 attempt, got %d", attemptCount)
	}
}

func TestRetry_EventualSuccess(t *testing.T) {
	// Test that an operation that succeeds after a few attempts works correctly
	opts := RetryOptions{
		MaxAttempts:    5,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Jitter:         false,
	}

	attemptCount := 0
	operation := func() error {
		attemptCount++
		if attemptCount < 3 {
			return errors.New("temporary error")
		}
		return nil // Success on third attempt
	}

	err := Retry(opts, operation)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}
}

func TestRetry_MaxAttemptsExceeded(t *testing.T) {
	// Test that an operation that always fails returns an error after max attempts
	opts := RetryOptions{
		MaxAttempts:    3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Jitter:         false,
	}

	attemptCount := 0
	expectedError := errors.New("persistent error")
	operation := func() error {
		attemptCount++
		return expectedError
	}

	err := Retry(opts, operation)
	if err == nil {
		t.Error("Expected an error, got nil")
	}

	if attemptCount != opts.MaxAttempts {
		t.Errorf("Expected %d attempts, got %d", opts.MaxAttempts, attemptCount)
	}

	// Check that the error message contains the original error
	if err != nil && !errors.Is(err, expectedError) {
		t.Errorf("Expected error to wrap %v, got %v", expectedError, err)
	}
}

func TestRetry_Backoff(t *testing.T) {
	// Test that backoff increases correctly
	opts := RetryOptions{
		MaxAttempts:    4,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Jitter:         false,
	}

	attemptCount := 0
	startTimes := make([]time.Time, 0, opts.MaxAttempts)
	operation := func() error {
		startTimes = append(startTimes, time.Now())
		attemptCount++
		return errors.New("error")
	}

	_ = Retry(opts, operation)

	if len(startTimes) != opts.MaxAttempts {
		t.Fatalf("Expected %d start times, got %d", opts.MaxAttempts, len(startTimes))
	}

	// Check that the time between attempts increases (with some tolerance for scheduling)
	for i := 1; i < len(startTimes); i++ {
		elapsed := startTimes[i].Sub(startTimes[i-1])
		expectedBackoff := opts.InitialBackoff * time.Duration(1<<uint(i-1))
		if expectedBackoff > opts.MaxBackoff {
			expectedBackoff = opts.MaxBackoff
		}

		// Allow for some scheduling variance
		minExpected := expectedBackoff - 5*time.Millisecond
		maxExpected := expectedBackoff + 15*time.Millisecond

		if elapsed < minExpected || elapsed > maxExpected {
			t.Errorf("Expected backoff between %v and %v for attempt %d, got %v", 
				minExpected, maxExpected, i, elapsed)
		}
	}
}

func TestCalculateBackoff_NoJitter(t *testing.T) {
	// Test backoff calculation without jitter
	testCases := []struct {
		current  time.Duration
		max      time.Duration
		expected time.Duration
	}{
		{10 * time.Millisecond, 100 * time.Millisecond, 20 * time.Millisecond},
		{50 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
		{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
	}

	for _, tc := range testCases {
		result := calculateBackoff(tc.current, tc.max, false)
		if result != tc.expected {
			t.Errorf("calculateBackoff(%v, %v, false) = %v, expected %v", 
				tc.current, tc.max, result, tc.expected)
		}
	}
}

func TestCalculateBackoff_WithJitter(t *testing.T) {
	// Test backoff calculation with jitter
	current := 100 * time.Millisecond
	max := 200 * time.Millisecond

	// With jitter, the result should be between 50% and 100% of the doubled value (capped at max)
	expected := 200 * time.Millisecond // doubled value
	minExpected := expected / 2        // 50% of doubled value
	maxExpected := expected            // 100% of doubled value

	// Run multiple times to account for randomness
	for i := 0; i < 100; i++ {
		result := calculateBackoff(current, max, true)
		if result < minExpected || result > maxExpected {
			t.Errorf("calculateBackoff(%v, %v, true) = %v, expected between %v and %v", 
				current, max, result, minExpected, maxExpected)
		}
	}
}

func TestAddJitter(t *testing.T) {
	// Test jitter calculation
	duration := 100 * time.Millisecond
	minExpected := duration / 2 // 50% of original
	maxExpected := duration     // 100% of original

	// Run multiple times to account for randomness
	for i := 0; i < 100; i++ {
		result := addJitter(duration)
		if result < minExpected || result > maxExpected {
			t.Errorf("addJitter(%v) = %v, expected between %v and %v", 
				duration, result, minExpected, maxExpected)
		}
	}
}
