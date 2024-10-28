package rate_limiter

import (
	"sync"
	"time"
)

// RateLimiter controls the rate at which actions are allowed
type RateLimiter struct {
	capacity     int           // Maximum number of tokens
	tokens       int           // Current available tokens
	refillRate   time.Duration // Time interval to add one token
	refillAmount int           // Tokens added each interval
	lastRefill   time.Time     // Timestamp of the last refill
	lock         sync.Mutex    // Mutex for concurrency safety
}

// NewRateLimiter initializes a RateLimiter with specified capacity and refill rate
func NewRateLimiter(capacity int, refillRate time.Duration, refillAmount int) *RateLimiter {
	return &RateLimiter{
		capacity:     capacity,
		tokens:       capacity,
		refillRate:   refillRate,
		refillAmount: refillAmount,
		lastRefill:   time.Now(),
	}
}

// Allow checks if a token is available and, if so, decrements the token count
func (rl *RateLimiter) Allow() bool {
	rl.lock.Lock()
	defer rl.lock.Unlock()

	rl.refillTokens() // Refill tokens based on elapsed time

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// refillTokens refills tokens based on the time elapsed since the last refill
func (rl *RateLimiter) refillTokens() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)

	if elapsed >= rl.refillRate {
		tokensToAdd := int(elapsed/rl.refillRate) * rl.refillAmount
		rl.tokens = min(rl.capacity, rl.tokens+tokensToAdd)
		rl.lastRefill = now
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Wait blocks until a token is available or the timeout is reached
func (rl *RateLimiter) Wait(timeout time.Duration) bool {
	start := time.Now()
	for {
		if rl.Allow() {
			return true
		}
		if time.Since(start) >= timeout {
			return false
		}
		time.Sleep(10 * time.Millisecond) // Polling interval
	}
}
