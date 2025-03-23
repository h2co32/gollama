// Package examples provides usage examples for the gollama library components.
package examples

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/h2co32/gollama/pkg/retry"
)

// RetryBasicExample demonstrates the basic usage of the retry package.
func RetryBasicExample() {
	// Create retry options
	opts := retry.DefaultOptions()
	opts.MaxAttempts = 5
	opts.InitialBackoff = 100 * time.Millisecond
	opts.OnRetry = func(attempt int, err error) {
		log.Printf("Attempt %d failed: %v. Retrying...", attempt, err)
	}

	// Function that will fail a few times before succeeding
	count := 0
	operation := func() error {
		count++
		if count < 3 {
			return errors.New("temporary error")
		}
		fmt.Println("Operation succeeded on attempt:", count)
		return nil
	}

	// Execute with retry
	err := retry.Do(opts, operation)
	if err != nil {
		log.Fatalf("Operation failed after retries: %v", err)
	}
}

// RetryHTTPExample demonstrates using retry for HTTP requests.
func RetryHTTPExample() {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Configure retry options
	opts := retry.Options{
		MaxAttempts:    5,
		InitialBackoff: 200 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
		Jitter:         true,
		OnRetry: func(attempt int, err error) {
			log.Printf("HTTP request attempt %d failed: %v", attempt, err)
		},
	}

	// Create HTTP client with shorter timeouts than the context
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Retry the HTTP request
	var resp *http.Response
	err := retry.DoWithContext(ctx, opts, func(ctx context.Context) error {
		req, err := http.NewRequestWithContext(ctx, "GET", "https://api.example.com/data", nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		var reqErr error
		resp, reqErr = client.Do(req)
		if reqErr != nil {
			return fmt.Errorf("request failed: %w", reqErr)
		}

		// Consider certain status codes as retriable errors
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			resp.Body.Close() // Avoid leaking resources
			return fmt.Errorf("server error: %d %s", resp.StatusCode, resp.Status)
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Failed to get data after retries: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Request succeeded with status: %s\n", resp.Status)
}

// RetryWithCustomBackoffExample demonstrates using retry with custom backoff logic.
func RetryWithCustomBackoffExample() {
	// Create retry options with custom backoff
	opts := retry.Options{
		MaxAttempts:    10,
		InitialBackoff: 50 * time.Millisecond,
		MaxBackoff:     2 * time.Second,
		Jitter:         true,
	}

	// Simulate a database operation that might fail
	operation := func() error {
		// Simulate random failures
		if time.Now().UnixNano()%3 == 0 {
			return errors.New("database connection error")
		}
		return nil
	}

	// Execute with retry
	err := retry.Do(opts, operation)
	if err != nil {
		log.Fatalf("Database operation failed: %v", err)
	}

	fmt.Println("Database operation completed successfully")
}
