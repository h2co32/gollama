package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDo_Success(t *testing.T) {
	// Test that a successful operation doesn't retry
	opts := Options{
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

	err := Do(opts, operation)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if attemptCount != 1 {
		t.Errorf("Expected 1 attempt, got %d", attemptCount)
	}
}

func TestDo_EventualSuccess(t *testing.T) {
	// Test that an operation that succeeds after a few attempts works correctly
	opts := Options{
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

	err := Do(opts, operation)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}
}

func TestDo_MaxAttemptsExceeded(t *testing.T) {
	// Test that an operation that always fails returns an error after max attempts
	opts := Options{
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

	err := Do(opts, operation)
	if err == nil {
		t.Error("Expected an error, got nil")
	}

	if attemptCount != opts.MaxAttempts {
		t.Errorf("Expected %d attempts, got %d", opts.MaxAttempts, attemptCount)
	}

	// Check that the error message contains the original error
	if err != nil && !errors.Is(err, ErrMaxAttemptsReached) {
		t.Errorf("Expected error to be ErrMaxAttemptsReached, got %v", err)
	}
}

func TestDoWithContext_Cancellation(t *testing.T) {
	// Test that context cancellation stops retries
	opts := Options{
		MaxAttempts:    10,
		InitialBackoff: 50 * time.Millisecond,
		MaxBackoff:     500 * time.Millisecond,
		Jitter:         false,
	}

	attemptCount := 0
	operation := func(ctx context.Context) error {
		attemptCount++
		return errors.New("error")
	}

	// Create a context that cancels after a short time
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := DoWithContext(ctx, opts, operation)
	if err == nil {
		t.Error("Expected an error, got nil")
	}

	// Should have made at least one attempt but not all attempts
	if attemptCount < 1 || attemptCount >= opts.MaxAttempts {
		t.Errorf("Expected between 1 and %d attempts, got %d", opts.MaxAttempts-1, attemptCount)
	}

	// Check that the error is related to context cancellation
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded error, got %v", err)
	}
}

func TestDo_WithOnRetryCallback(t *testing.T) {
	// Test that the OnRetry callback is called correctly
	opts := Options{
		MaxAttempts:    3,
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     100 * time.Millisecond,
		Jitter:         false,
	}

	callbackCount := 0
	expectedError := errors.New("test error")
	
	opts.OnRetry = func(attempt int, err error) {
		callbackCount++
		if err != expectedError {
			t.Errorf("Expected error %v in callback, got %v", expectedError, err)
		}
		if attempt < 1 || attempt >= opts.MaxAttempts {
			t.Errorf("Expected attempt between 1 and %d in callback, got %d", opts.MaxAttempts-1, attempt)
		}
	}

	attemptCount := 0
	operation := func() error {
		attemptCount++
		return expectedError
	}

	_ = Do(opts, operation)

	// OnRetry should be called MaxAttempts-1 times (not called after the last attempt)
	if callbackCount != opts.MaxAttempts-1 {
		t.Errorf("Expected OnRetry to be called %d times, got %d", opts.MaxAttempts-1, callbackCount)
	}
}

func TestDefaultOptions(t *testing.T) {
	// Test that default options are set correctly
	opts := DefaultOptions()
	
	if opts.MaxAttempts != 3 {
		t.Errorf("Expected MaxAttempts to be 3, got %d", opts.MaxAttempts)
	}
	
	if opts.InitialBackoff != 100*time.Millisecond {
		t.Errorf("Expected InitialBackoff to be 100ms, got %v", opts.InitialBackoff)
	}
	
	if opts.MaxBackoff != 10*time.Second {
		t.Errorf("Expected MaxBackoff to be 10s, got %v", opts.MaxBackoff)
	}
	
	if !opts.Jitter {
		t.Error("Expected Jitter to be true")
	}
}

func TestCalculateBackoff(t *testing.T) {
	// Test backoff calculation without jitter
	testCases := []struct {
		current  time.Duration
		max      time.Duration
		expected time.Duration
	}{
		{10 * time.Millisecond, 100 * time.Millisecond, 10 * time.Millisecond},
		{50 * time.Millisecond, 100 * time.Millisecond, 50 * time.Millisecond},
		{100 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
		{200 * time.Millisecond, 100 * time.Millisecond, 100 * time.Millisecond},
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

	// With jitter, the result should be between 50% and 100% of the current value (capped at max)
	minExpected := current / 2  // 50% of current value
	maxExpected := current      // 100% of current value

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
