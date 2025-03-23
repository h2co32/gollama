package examples

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/h2co32/gollama/pkg/ratelimiter"
)

// RateLimiterBasicExample demonstrates basic usage of the rate limiter.
func RateLimiterBasicExample() {
	// Create a rate limiter with 5 tokens per second and a burst capacity of 10
	limiter := ratelimiter.New(5, time.Second, 10)

	fmt.Println("Starting rate limiter example...")
	fmt.Printf("Initial tokens available: %.2f\n", limiter.Available())

	// Try to perform 15 operations in quick succession
	for i := 1; i <= 15; i++ {
		if limiter.Allow() {
			fmt.Printf("Operation %d allowed\n", i)
		} else {
			fmt.Printf("Operation %d denied (rate limit exceeded)\n", i)
		}
	}

	// Wait for tokens to refill
	fmt.Println("\nWaiting for 1 second to allow tokens to refill...")
	time.Sleep(time.Second)
	fmt.Printf("Tokens available after waiting: %.2f\n", limiter.Available())
}

// RateLimiterWaitExample demonstrates using Wait to block until operations are allowed.
func RateLimiterWaitExample() {
	// Create a rate limiter with 2 tokens per second
	limiter := ratelimiter.New(2, time.Second, 2)

	fmt.Println("Starting rate limiter wait example...")

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Perform 5 operations, waiting for each one
	for i := 1; i <= 5; i++ {
		start := time.Now()
		err := limiter.Wait(ctx)
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("Operation %d failed: %v\n", i, err)
			break
		}

		fmt.Printf("Operation %d allowed after waiting %.2f seconds\n", i, elapsed.Seconds())
	}
}

// RateLimiterConcurrentExample demonstrates using a rate limiter with concurrent operations.
func RateLimiterConcurrentExample() {
	// Create a rate limiter with 3 tokens per second
	limiter := ratelimiter.New(3, time.Second, 3)

	fmt.Println("Starting concurrent rate limiter example...")

	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup
	
	// Launch 10 concurrent operations
	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// Create a context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			
			start := time.Now()
			err := limiter.Wait(ctx)
			elapsed := time.Since(start)
			
			if err != nil {
				fmt.Printf("Worker %d timed out: %v\n", id, err)
				return
			}
			
			fmt.Printf("Worker %d completed after waiting %.2f seconds\n", id, elapsed.Seconds())
			
			// Simulate work
			time.Sleep(100 * time.Millisecond)
		}(i)
	}
	
	// Wait for all goroutines to finish
	wg.Wait()
	fmt.Println("All workers completed")
}

// RateLimiterAPIExample demonstrates using a rate limiter for API requests.
func RateLimiterAPIExample() {
	// Create a rate limiter for an API with a limit of 2 requests per second
	apiLimiter := ratelimiter.New(2, time.Second, 2)
	
	fmt.Println("Starting API rate limiter example...")
	
	// Simulate making API requests
	for i := 1; i <= 5; i++ {
		// Create a context with timeout for each request
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		
		start := time.Now()
		fmt.Printf("Attempting API request %d...\n", i)
		
		// Wait until we're allowed to make the request
		err := apiLimiter.Wait(ctx)
		if err != nil {
			fmt.Printf("API request %d timed out: %v\n", i, err)
			cancel()
			continue
		}
		
		// Simulate API request
		fmt.Printf("Making API request %d after waiting %.2f seconds\n", i, time.Since(start).Seconds())
		
		// Simulate response time
		time.Sleep(200 * time.Millisecond)
		fmt.Printf("API request %d completed\n", i)
		
		cancel()
	}
}
