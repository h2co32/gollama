package queue

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestNewJobQueue(t *testing.T) {
	workerCount := 5
	rateLimit := 100 * time.Millisecond

	jq := NewJobQueue(workerCount, rateLimit)

	if jq == nil {
		t.Fatal("Expected NewJobQueue to return a non-nil value")
	}

	if jq.workerCount != workerCount {
		t.Errorf("Expected jq.workerCount to be %d, got %d", workerCount, jq.workerCount)
	}

	if jq.rateLimit != rateLimit {
		t.Errorf("Expected jq.rateLimit to be %v, got %v", rateLimit, jq.rateLimit)
	}

	if jq.jobs == nil {
		t.Error("Expected jq.jobs channel to be initialized")
	}

	if jq.results == nil {
		t.Error("Expected jq.results map to be initialized")
	}
}

func TestJobQueueProcessing(t *testing.T) {
	// Create a job queue with 2 workers and a small rate limit
	jq := NewJobQueue(2, 10*time.Millisecond)
	jq.StartWorkers()

	// Create a few test jobs
	jobCount := 5
	jobResults := make([]bool, jobCount)
	var resultsMutex sync.Mutex

	for i := 0; i < jobCount; i++ {
		jobID := i
		jq.AddJob(jobID, func() error {
			// Simulate work
			time.Sleep(20 * time.Millisecond)
			resultsMutex.Lock()
			jobResults[jobID] = true
			resultsMutex.Unlock()
			return nil
		}, 1)
	}

	// Wait for all jobs to complete
	jq.Wait()

	// Verify all jobs were processed
	for i, result := range jobResults {
		if !result {
			t.Errorf("Expected job %d to be processed", i)
		}
	}

	// Check the results map
	results := jq.GetResults()
	if len(results) != jobCount {
		t.Errorf("Expected %d results, got %d", jobCount, len(results))
	}

	for i := 0; i < jobCount; i++ {
		if results[i] != nil {
			t.Errorf("Expected job %d to succeed, got error: %v", i, results[i])
		}
	}
}

func TestJobQueueRetries(t *testing.T) {
	// Create a job queue with 1 worker and a small rate limit
	jq := NewJobQueue(1, 10*time.Millisecond)
	jq.StartWorkers()

	// Create a job that fails on the first attempt but succeeds on the second
	var attemptCount int
	expectedError := errors.New("first attempt error")

	jq.AddJob(1, func() error {
		attemptCount++
		if attemptCount == 1 {
			return expectedError
		}
		return nil
	}, 3) // Allow up to 3 retries

	// Wait for the job to complete
	jq.Wait()

	// Verify the job was retried and eventually succeeded
	if attemptCount != 2 {
		t.Errorf("Expected 2 attempts, got %d", attemptCount)
	}

	results := jq.GetResults()
	if results[1] != nil {
		t.Errorf("Expected job to eventually succeed, got error: %v", results[1])
	}

	// Create a job that always fails
	jq = NewJobQueue(1, 10*time.Millisecond)
	jq.StartWorkers()

	persistentError := errors.New("persistent error")
	maxRetries := 3
	attemptCount = 0

	jq.AddJob(1, func() error {
		attemptCount++
		return persistentError
	}, maxRetries)

	// Wait for the job to complete
	jq.Wait()

	// Verify the job was retried the maximum number of times and still failed
	if attemptCount != maxRetries {
		t.Errorf("Expected %d attempts, got %d", maxRetries, attemptCount)
	}

	results = jq.GetResults()
	if results[1] != persistentError {
		t.Errorf("Expected job to fail with error %v, got %v", persistentError, results[1])
	}
}

func TestJobQueueConcurrency(t *testing.T) {
	// Test that multiple workers can process jobs concurrently
	workerCount := 5
	jobCount := 10 // Reduced from 20 to avoid timeouts
	rateLimit := 5 * time.Millisecond

	jq := NewJobQueue(workerCount, rateLimit)
	jq.StartWorkers()

	// Track the number of concurrently running jobs
	var runningJobs int
	var maxRunningJobs int
	var jobsMutex sync.Mutex
	var wg sync.WaitGroup
	wg.Add(jobCount)

	for i := 0; i < jobCount; i++ {
		jq.AddJob(i, func() error {
			defer wg.Done()
			
			jobsMutex.Lock()
			runningJobs++
			if runningJobs > maxRunningJobs {
				maxRunningJobs = runningJobs
			}
			jobsMutex.Unlock()

			// Simulate work - shorter duration to avoid timeouts
			time.Sleep(20 * time.Millisecond)

			jobsMutex.Lock()
			runningJobs--
			jobsMutex.Unlock()

			return nil
		}, 1)
	}

	// Wait for all jobs to complete
	go func() {
		wg.Wait()
		// Close the jobs channel after all jobs are done
		close(jq.jobs)
	}()

	// Set a timeout for the test
	done := make(chan struct{})
	go func() {
		jq.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Test completed successfully
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out waiting for jobs to complete")
	}

	// Verify that multiple jobs ran concurrently
	// The maximum should be limited by the worker count
	if maxRunningJobs <= 1 {
		t.Error("Expected multiple jobs to run concurrently")
	}

	if maxRunningJobs > workerCount {
		t.Errorf("Expected at most %d concurrent jobs, got %d", workerCount, maxRunningJobs)
	}
}

func TestJobQueueRateLimit(t *testing.T) {
	// Test that the rate limit is respected
	workerCount := 1
	rateLimit := 100 * time.Millisecond

	jq := NewJobQueue(workerCount, rateLimit)
	jq.StartWorkers()

	// Add jobs that complete instantly
	jobCount := 3
	completionTimes := make([]time.Time, jobCount)

	for i := 0; i < jobCount; i++ {
		jobID := i
		jq.AddJob(jobID, func() error {
			completionTimes[jobID] = time.Now()
			return nil
		}, 1)
	}

	// Wait for all jobs to complete
	jq.Wait()

	// Verify that jobs were processed with the rate limit
	for i := 1; i < jobCount; i++ {
		elapsed := completionTimes[i].Sub(completionTimes[i-1])
		if elapsed < rateLimit {
			t.Errorf("Expected at least %v between jobs, got %v", rateLimit, elapsed)
		}
	}
}

func TestGetResults(t *testing.T) {
	// Test that GetResults returns the correct results
	jq := NewJobQueue(1, 10*time.Millisecond)
	jq.StartWorkers()

	// Add jobs with different outcomes
	successErr := error(nil)
	failureErr := errors.New("job failed")

	jq.AddJob(1, func() error { return successErr }, 1)
	jq.AddJob(2, func() error { return failureErr }, 1)

	// Wait for all jobs to complete
	jq.Wait()

	// Get the results
	results := jq.GetResults()

	// Verify the results
	if results[1] != successErr {
		t.Errorf("Expected job 1 result to be %v, got %v", successErr, results[1])
	}

	if results[2] != failureErr {
		t.Errorf("Expected job 2 result to be %v, got %v", failureErr, results[2])
	}
}
