// Package ratelimiter provides a token bucket rate limiter for controlling request rates.
//
// This package helps prevent overloading services by limiting the rate at which
// operations can be performed. It implements a token bucket algorithm where tokens
// are added at a fixed rate up to a maximum capacity. Each operation consumes a token,
// and operations are only allowed when tokens are available.
//
// Example usage:
//
//	// Create a rate limiter with 10 tokens per second and a burst capacity of 20
//	limiter := ratelimiter.New(10, time.Second, 20)
//
//	// Check if an operation is allowed
//	if limiter.Allow() {
//		// Perform the operation
//	} else {
//		// Operation not allowed, handle accordingly
//	}
//
//	// Or wait until an operation is allowed (with timeout)
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	if err := limiter.Wait(ctx); err == nil {
//		// Perform the operation
//	} else {
//		// Timed out waiting for a token
//	}
package ratelimiter

import (
	"context"
	"sync"
	"time"
)

// Version represents the current package version following semantic versioning.
const Version = "1.0.0"

// RateLimiter implements a token bucket rate limiter.
type RateLimiter struct {
	rate           float64       // Tokens per second
	interval       time.Duration // Interval between token refills
	capacity       float64       // Maximum number of tokens
	tokens         float64       // Current number of tokens
	lastRefillTime time.Time     // Last time tokens were refilled
	mu             sync.Mutex    // Mutex for thread safety
}

// New creates a new RateLimiter with the specified rate and capacity.
//
// Parameters:
//   - rate: Number of operations allowed per interval
//   - interval: Time interval for rate calculation
//   - capacity: Maximum burst capacity (if not specified, defaults to rate)
//
// For example, New(10, time.Second, 20) creates a limiter that allows
// 10 operations per second with a burst capacity of 20.
func New(rate float64, interval time.Duration, capacity float64) *RateLimiter {
	if capacity <= 0 {
		capacity = rate
	}

	return &RateLimiter{
		rate:           rate,
		interval:       interval,
		capacity:       capacity,
		tokens:         capacity, // Start with full capacity
		lastRefillTime: time.Now(),
	}
}

// Allow checks if an operation is allowed and consumes a token if available.
// It returns true if the operation is allowed, false otherwise.
func (rl *RateLimiter) Allow() bool {
	return rl.AllowN(1)
}

// AllowN checks if n operations are allowed and consumes n tokens if available.
// It returns true if the operations are allowed, false otherwise.
func (rl *RateLimiter) AllowN(n float64) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refill()
	if rl.tokens >= n {
		rl.tokens -= n
		return true
	}
	return false
}

// Wait blocks until an operation is allowed or the context is canceled.
// It returns nil if a token was obtained, or an error if the context was canceled.
func (rl *RateLimiter) Wait(ctx context.Context) error {
	return rl.WaitN(ctx, 1)
}

// WaitN blocks until n operations are allowed or the context is canceled.
// It returns nil if n tokens were obtained, or an error if the context was canceled.
func (rl *RateLimiter) WaitN(ctx context.Context, n float64) error {
	// Fast path: check if we can get tokens immediately
	if rl.AllowN(n) {
		return nil
	}

	// Slow path: wait for tokens to become available
	ticker := time.NewTicker(rl.interval / 10) // Check more frequently than the refill interval
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if rl.AllowN(n) {
				return nil
			}
		}
	}
}

// refill adds tokens based on the time elapsed since the last refill.
// This method is not thread-safe and should be called with the mutex locked.
func (rl *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefillTime)
	rl.lastRefillTime = now

	// Calculate tokens to add based on elapsed time and rate
	tokensToAdd := float64(elapsed) / float64(rl.interval) * rl.rate
	if tokensToAdd > 0 {
		rl.tokens += tokensToAdd
		if rl.tokens > rl.capacity {
			rl.tokens = rl.capacity
		}
	}
}

// Available returns the current number of available tokens.
func (rl *RateLimiter) Available() float64 {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.refill()
	return rl.tokens
}

// Capacity returns the maximum number of tokens the limiter can hold.
func (rl *RateLimiter) Capacity() float64 {
	return rl.capacity
}

// Rate returns the rate at which tokens are added to the bucket.
func (rl *RateLimiter) Rate() float64 {
	return rl.rate
}
