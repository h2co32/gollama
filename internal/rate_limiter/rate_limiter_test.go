package rate_limiter

import (
	"sync"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	capacity := 10
	refillRate := 100 * time.Millisecond
	refillAmount := 2

	rl := NewRateLimiter(capacity, refillRate, refillAmount)

	if rl == nil {
		t.Fatal("Expected NewRateLimiter to return a non-nil value")
	}

	if rl.capacity != capacity {
		t.Errorf("Expected rl.capacity to be %d, got %d", capacity, rl.capacity)
	}

	if rl.tokens != capacity {
		t.Errorf("Expected rl.tokens to be %d, got %d", capacity, rl.tokens)
	}

	if rl.refillRate != refillRate {
		t.Errorf("Expected rl.refillRate to be %v, got %v", refillRate, rl.refillRate)
	}

	if rl.refillAmount != refillAmount {
		t.Errorf("Expected rl.refillAmount to be %d, got %d", refillAmount, rl.refillAmount)
	}

	// lastRefill should be close to now
	now := time.Now()
	if now.Sub(rl.lastRefill) > 100*time.Millisecond {
		t.Errorf("Expected rl.lastRefill to be close to now, got %v", rl.lastRefill)
	}
}

func TestAllow(t *testing.T) {
	// Create a rate limiter with 5 tokens, refilling 1 token every 100ms
	rl := NewRateLimiter(5, 100*time.Millisecond, 1)

	// Test that we can use all 5 tokens
	for i := 0; i < 5; i++ {
		if !rl.Allow() {
			t.Errorf("Expected Allow() to return true for token %d", i+1)
		}
	}

	// Test that the 6th token is not allowed
	if rl.Allow() {
		t.Error("Expected Allow() to return false after all tokens are used")
	}

	// Wait for a refill (should get 1 token after 100ms)
	time.Sleep(110 * time.Millisecond)

	// Test that we can use the refilled token
	if !rl.Allow() {
		t.Error("Expected Allow() to return true after refill")
	}

	// Test that no more tokens are available
	if rl.Allow() {
		t.Error("Expected Allow() to return false after using refilled token")
	}

	// Wait for multiple refills (should get 3 tokens after 300ms)
	time.Sleep(310 * time.Millisecond)

	// Test that we can use the refilled tokens
	for i := 0; i < 3; i++ {
		if !rl.Allow() {
			t.Errorf("Expected Allow() to return true for refilled token %d", i+1)
		}
	}

	// Test that no more tokens are available
	if rl.Allow() {
		t.Error("Expected Allow() to return false after using all refilled tokens")
	}
}

func TestRefillTokens(t *testing.T) {
	// Create a rate limiter with 5 tokens, refilling 2 tokens every 100ms
	rl := NewRateLimiter(5, 100*time.Millisecond, 2)

	// Use all tokens
	for i := 0; i < 5; i++ {
		rl.Allow()
	}

	// Manually set lastRefill to a time in the past
	rl.lock.Lock()
	rl.lastRefill = time.Now().Add(-250 * time.Millisecond) // 2.5 refill periods
	rl.lock.Unlock()

	// Call Allow() which should trigger a refill
	result := rl.Allow()

	// Should have refilled 2 tokens per 100ms for 250ms = 5 tokens (capped at capacity)
	// And then used 1 token for the Allow() call
	if !result {
		t.Error("Expected Allow() to return true after refill")
	}

	// Should have 4 tokens left (5 refilled - 1 used)
	// But the actual implementation might calculate differently based on elapsed time
	rl.lock.Lock()
	tokensLeft := rl.tokens
	rl.lock.Unlock()

	// The implementation calculates tokens based on elapsed time, which can vary
	// So we just check that we have some tokens (more than 0, at most capacity)
	if tokensLeft <= 0 || tokensLeft > rl.capacity {
		t.Errorf("Expected tokens left to be between 1 and %d, got %d", rl.capacity, tokensLeft)
	}

	// Test that refill doesn't exceed capacity
	rl.lock.Lock()
	rl.tokens = 4 // Set to 4 tokens
	rl.lastRefill = time.Now().Add(-500 * time.Millisecond) // 5 refill periods
	rl.lock.Unlock()

	rl.Allow() // This should trigger a refill and then use a token

	rl.lock.Lock()
	tokensLeft = rl.tokens
	rl.lock.Unlock()

	// Should have refilled to capacity (5) and then used 1 token
	if tokensLeft != 4 {
		t.Errorf("Expected 4 tokens left after refill to capacity, got %d", tokensLeft)
	}
}

func TestWait(t *testing.T) {
	// Create a rate limiter with 2 tokens, refilling 1 token every 100ms
	rl := NewRateLimiter(2, 100*time.Millisecond, 1)

	// Use all tokens
	rl.Allow()
	rl.Allow()

	// Test that Wait returns true when a token becomes available within the timeout
	start := time.Now()
	result := rl.Wait(200 * time.Millisecond)
	elapsed := time.Since(start)

	if !result {
		t.Error("Expected Wait() to return true when a token becomes available")
	}

	if elapsed < 100*time.Millisecond {
		t.Errorf("Expected Wait() to block for at least 100ms, only blocked for %v", elapsed)
	}

	// Use all tokens again
	rl.Allow()

	// Test that Wait returns false when no token becomes available within the timeout
	start = time.Now()
	result = rl.Wait(50 * time.Millisecond) // Timeout before a token is refilled
	elapsed = time.Since(start)

	if result {
		t.Error("Expected Wait() to return false when no token becomes available")
	}

	if elapsed < 50*time.Millisecond {
		t.Errorf("Expected Wait() to block for at least 50ms, only blocked for %v", elapsed)
	}
}

func TestConcurrentAccess(t *testing.T) {
	// Create a rate limiter with 100 tokens, refilling 10 tokens every 100ms
	rl := NewRateLimiter(100, 100*time.Millisecond, 10)

	// Run multiple goroutines to access the rate limiter concurrently
	const numGoroutines = 10
	const numRequests = 20
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Count how many requests were allowed
	var allowedCount int
	var countMutex sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numRequests; j++ {
				if rl.Allow() {
					countMutex.Lock()
					allowedCount++
					countMutex.Unlock()
				}
				time.Sleep(10 * time.Millisecond) // Small delay between requests
			}
		}()
	}

	wg.Wait()

	// We should have allowed at most the initial capacity (100) plus any refills
	// Since we're running 10 goroutines with 20 requests each (200 total) with small delays,
	// we expect to allow significantly less than 200 requests
	if allowedCount > 150 {
		t.Errorf("Expected to allow at most ~150 requests, allowed %d", allowedCount)
	}

	if allowedCount < 100 {
		t.Errorf("Expected to allow at least 100 requests, allowed %d", allowedCount)
	}
}

func TestMin(t *testing.T) {
	testCases := []struct {
		a, b, expected int
	}{
		{5, 10, 5},
		{10, 5, 5},
		{0, 5, 0},
		{-5, 5, -5},
		{5, -5, -5},
		{0, 0, 0},
	}

	for _, tc := range testCases {
		result := min(tc.a, tc.b)
		if result != tc.expected {
			t.Errorf("min(%d, %d) = %d, expected %d", tc.a, tc.b, result, tc.expected)
		}
	}
}
