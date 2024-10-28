package queue

import (
	"fmt"
	"sync"
	"time"
)

// Job represents a unit of work to be processed by the job queue
type Job struct {
	ID      int
	Task    func() error
	Retries int
}

// JobQueue manages background job processing with a worker pool and rate limiting
type JobQueue struct {
	jobs         chan Job
	workerCount  int
	rateLimit    time.Duration
	wg           sync.WaitGroup
	results      map[int]error
	resultsMutex sync.Mutex
}

// NewJobQueue initializes a new JobQueue with the specified number of workers and rate limit
func NewJobQueue(workerCount int, rateLimit time.Duration) *JobQueue {
	return &JobQueue{
		jobs:        make(chan Job),
		workerCount: workerCount,
		rateLimit:   rateLimit,
		results:     make(map[int]error),
	}
}

// StartWorkers starts the worker pool to process jobs asynchronously
func (jq *JobQueue) StartWorkers() {
	for i := 0; i < jq.workerCount; i++ {
		go jq.worker(i)
	}
}

// worker is a function that processes jobs from the queue with rate limiting
func (jq *JobQueue) worker(workerID int) {
	for job := range jq.jobs {
		fmt.Printf("Worker %d processing job %d\n", workerID, job.ID)

		retryCount := job.Retries
		var err error
		for attempt := 1; attempt <= retryCount; attempt++ {
			err = job.Task()
			if err == nil {
				break
			}
			fmt.Printf("Job %d failed (attempt %d/%d): %v\n", job.ID, attempt, retryCount, err)
			time.Sleep(500 * time.Millisecond) // Backoff between retries
		}

		jq.resultsMutex.Lock()
		jq.results[job.ID] = err
		jq.resultsMutex.Unlock()

		time.Sleep(jq.rateLimit) // Rate limiting
		jq.wg.Done()
	}
}

// AddJob adds a job to the job queue for processing
func (jq *JobQueue) AddJob(id int, task func() error, retries int) {
	jq.wg.Add(1)
	jq.jobs <- Job{ID: id, Task: task, Retries: retries}
}

// Wait blocks until all jobs have been processed
func (jq *JobQueue) Wait() {
	jq.wg.Wait()
	close(jq.jobs)
}

// GetResults returns the job results after all jobs are processed
func (jq *JobQueue) GetResults() map[int]error {
	jq.resultsMutex.Lock()
	defer jq.resultsMutex.Unlock()
	return jq.results
}
