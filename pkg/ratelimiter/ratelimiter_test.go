package ratelimiter

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	rate := 10.0
	interval := 100 * time.Millisecond
	capacity := 20.0

	rl := New(rate, interval, capacity)

	if rl == nil {
		t.Fatal("Expected New to return a non-nil value")
	}

	if rl.rate != rate {
		t.Errorf("Expected rl.rate to be %f, got %f", rate, rl.rate)
	}

	if rl.interval != interval {
		t.Errorf("Expected rl.interval to be %v, got %v", interval, rl.interval)
	}

	if rl.capacity != capacity {
		t.Errorf("Expected rl.capacity to be %f, got %f", capacity, rl.capacity)
	}

	if rl.tokens != capacity {
		t.Errorf("Expected rl.tokens to be %f, got %f", capacity, rl.tokens)
	}

	// lastRefillTime should be close to now
	now := time.Now()
	if now.Sub(rl.lastRefillTime) > 100*time.Millisecond {
		t.Errorf("Expected rl.lastRefillTime to be close to now, got %v", rl.lastRefillTime)
	}

	// Test default capacity
	rl = New(rate, interval, 0)
	if rl.capacity != rate {
		t.Errorf("Expected default capacity to be equal to rate (%f), got %f", rate, rl.capacity)
	}
}

func TestAllow(t *testing.T) {
	// Create a rate limiter with 5 tokens per second
	rl := New(5, time.Second, 5)

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

	// Wait for a refill (should get 1 token after 200ms with rate=5/s)
	time.Sleep(210 * time.Millisecond)

	// Test that we can use the refilled token
	if !rl.Allow() {
		t.Error("Expected Allow() to return true after refill")
	}

	// Test that no more tokens are available
	if rl.Allow() {
		t.Error("Expected Allow() to return false after using refilled token")
	}
}

func TestAllowN(t *testing.T) {
	// Create a rate limiter with 10 tokens per second and capacity of 10
	rl := New(10, time.Second, 10)

	// Test that we can use 5 tokens at once
	if !rl.AllowN(5) {
		t.Error("Expected AllowN(5) to return true")
	}

	// Test that we can use 5 more tokens
	if !rl.AllowN(5) {
		t.Error("Expected AllowN(5) to return true after using 5 tokens")
	}

	// Test that we can't use any more tokens
	if rl.AllowN(1) {
		t.Error("Expected AllowN(1) to return false after using all tokens")
	}

	// Test that we can't use more tokens than capacity
	rl = New(10, time.Second, 10)
	if rl.AllowN(11) {
		t.Error("Expected AllowN(11) to return false when capacity is 10")
	}
}

func TestRefill(t *testing.T) {
	// Create a rate limiter with 10 tokens per second
	rl := New(10, time.Second, 10)

	// Use all tokens
	for i := 0; i < 10; i++ {
		rl.Allow()
	}

	// Manually set lastRefillTime to a time in the past
	rl.mu.Lock()
	rl.lastRefillTime = time.Now().Add(-500 * time.Millisecond) // 0.5 seconds ago
	rl.mu.Unlock()

	// Call Allow() which should trigger a refill
	result := rl.Allow()

	// Should have refilled 10 tokens/second * 0.5 seconds = 5 tokens
	// And then used 1 token for the Allow() call
	if !result {
		t.Error("Expected Allow() to return true after refill")
	}

	// Should have approximately 4 tokens left (5 refilled - 1 used)
	rl.mu.Lock()
	tokensLeft := rl.tokens
	rl.mu.Unlock()

	// Allow for some small timing differences
	if tokensLeft < 3.5 || tokensLeft > 4.5 {
		t.Errorf("Expected approximately 4 tokens left, got %f", tokensLeft)
	}

	// Test that refill doesn't exceed capacity
	rl.mu.Lock()
	rl.tokens = 5 // Set to 5 tokens
	rl.lastRefillTime = time.Now().Add(-2 * time.Second) // 2 seconds ago
	rl.mu.Unlock()

	rl.Allow() // This should trigger a refill and then use a token

	rl.mu.Lock()
	tokensLeft = rl.tokens
	rl.mu.Unlock()

	// Should have refilled to capacity (10) and then used 1 token
	if tokensLeft < 8.5 || tokensLeft > 9.5 {
		t.Errorf("Expected approximately 9 tokens left after refill to capacity, got %f", tokensLeft)
	}
}

func TestWait(t *testing.T) {
	// Create a rate limiter with 5 tokens per second
	rl := New(5, time.Second, 5)

	// Use all tokens
	for i := 0; i < 5; i++ {
		rl.Allow()
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	// Test that Wait returns nil when a token becomes available within the timeout
	start := time.Now()
	err := rl.Wait(ctx)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Expected Wait() to return nil, got %v", err)
	}

	// Should have waited approximately 200ms for a token (with rate=5/s)
	if elapsed < 180*time.Millisecond {
		t.Errorf("Expected Wait() to block for at least 180ms, only blocked for %v", elapsed)
	}

	// Use all tokens again
	for i := 0; i < 5; i++ {
		rl.Allow()
	}

	// Test that Wait returns context.DeadlineExceeded when no token becomes available within the timeout
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	start = time.Now()
	err = rl.Wait(ctx)
	elapsed = time.Since(start)

	if err != context.DeadlineExceeded {
		t.Errorf("Expected Wait() to return context.DeadlineExceeded, got %v", err)
	}

	if elapsed < 90*time.Millisecond {
		t.Errorf("Expected Wait() to block for at least 90ms, only blocked for %v", elapsed)
	}
}

func TestWaitN(t *testing.T) {
	// Create a rate limiter with 10 tokens per second
	rl := New(10, time.Second, 10)

	// Use all tokens
	for i := 0; i < 10; i++ {
		rl.Allow()
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Millisecond)
	defer cancel()

	// Test that WaitN returns nil when 5 tokens become available within the timeout
	start := time.Now()
	err := rl.WaitN(ctx, 5)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Expected WaitN() to return nil, got %v", err)
	}

	// Should have waited approximately 500ms for 5 tokens (with rate=10/s)
	if elapsed < 450*time.Millisecond {
		t.Errorf("Expected WaitN() to block for at least 450ms, only blocked for %v", elapsed)
	}

	// Use all tokens again
	for i := 0; i < 10; i++ {
		rl.Allow()
	}

	// Test that WaitN returns context.DeadlineExceeded when requesting more tokens than can be acquired within the timeout
	ctx, cancel = context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	start = time.Now()
	err = rl.WaitN(ctx, 5) // Would need 500ms to get 5 tokens
	elapsed = time.Since(start)

	if err != context.DeadlineExceeded {
		t.Errorf("Expected WaitN() to return context.DeadlineExceeded, got %v", err)
	}

	if elapsed < 290*time.Millisecond {
		t.Errorf("Expected WaitN() to block for at least 290ms, only blocked for %v", elapsed)
	}
}

func TestAvailable(t *testing.T) {
	// Create a rate limiter with 10 tokens per second
	rl := New(10, time.Second, 10)

	// Use 7 tokens
	for i := 0; i < 7; i++ {
		rl.Allow()
	}

	// Should have 3 tokens available
	available := rl.Available()
	if available < 2.9 || available > 3.1 {
		t.Errorf("Expected approximately 3 tokens available, got %f", available)
	}

	// Wait for some tokens to refill
	time.Sleep(200 * time.Millisecond)

	// Should have approximately 5 tokens available (3 + 10*0.2)
	available = rl.Available()
	if available < 4.8 || available > 5.2 {
		t.Errorf("Expected approximately 5 tokens available, got %f", available)
	}
}

func TestCapacity(t *testing.T) {
	capacity := 15.0
	rl := New(10, time.Second, capacity)

	if rl.Capacity() != capacity {
		t.Errorf("Expected Capacity() to return %f, got %f", capacity, rl.Capacity())
	}
}

func TestRate(t *testing.T) {
	rate := 10.0
	rl := New(rate, time.Second, 15)

	if rl.Rate() != rate {
		t.Errorf("Expected Rate() to return %f, got %f", rate, rl.Rate())
	}
}

func TestConcurrentAccess(t *testing.T) {
	// Create a rate limiter with 100 tokens per second
	rl := New(100, time.Second, 100)

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
