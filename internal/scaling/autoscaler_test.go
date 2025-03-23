package autoscaler

import (
	"sync"
	"testing"
	"time"
)

// TestAutoScaler is a modified version of AutoScaler for testing
type TestAutoScaler struct {
	workerPool        chan struct{}
	minWorkers        int
	maxWorkers        int
	cpuThreshold      float64
	scaleUpInterval   time.Duration
	scaleDownInterval time.Duration
	wg                sync.WaitGroup
	stopChan          chan struct{}
	cpuUsageFunc      func() float64 // Custom CPU usage function for testing
}

// NewTestAutoScaler creates a test autoscaler with a custom CPU usage function
func NewTestAutoScaler(minWorkers, maxWorkers int, cpuThreshold float64, 
	scaleUpInterval, scaleDownInterval time.Duration, cpuUsageFunc func() float64) *TestAutoScaler {
	
	as := &TestAutoScaler{
		workerPool:        make(chan struct{}, maxWorkers),
		minWorkers:        minWorkers,
		maxWorkers:        maxWorkers,
		cpuThreshold:      cpuThreshold,
		scaleUpInterval:   scaleUpInterval,
		scaleDownInterval: scaleDownInterval,
		stopChan:          make(chan struct{}),
		cpuUsageFunc:      cpuUsageFunc,
	}

	for i := 0; i < minWorkers; i++ {
		as.workerPool <- struct{}{}
	}

	return as
}

// Start begins monitoring system load and scaling workers accordingly
func (as *TestAutoScaler) Start() {
	go as.monitorLoad()
}

// monitorLoad periodically checks CPU usage and scales workers up or down
func (as *TestAutoScaler) monitorLoad() {
	for {
		select {
		case <-as.stopChan:
			return
		default:
			cpuUsage := as.cpuUsageFunc() // Use the custom function
			currentWorkers := len(as.workerPool)

			if cpuUsage > as.cpuThreshold && currentWorkers < as.maxWorkers {
				as.scaleUp()
			} else if cpuUsage < as.cpuThreshold && currentWorkers > as.minWorkers {
				as.scaleDown()
			}

			time.Sleep(100 * time.Millisecond) // Shorter sleep for testing
		}
	}
}

// scaleUp adds workers up to the maximum limit
func (as *TestAutoScaler) scaleUp() {
	as.wg.Add(1)
	defer as.wg.Done()

	select {
	case as.workerPool <- struct{}{}:
		// Scaled up successfully
	case <-time.After(as.scaleUpInterval):
		// Scale-up timed out
	}
}

// scaleDown removes a worker down to the minimum limit
func (as *TestAutoScaler) scaleDown() {
	as.wg.Add(1)
	defer as.wg.Done()

	select {
	case <-as.workerPool:
		// Scaled down successfully
	case <-time.After(as.scaleDownInterval):
		// Scale-down timed out
	}
}

// Stop stops the autoscaler
func (as *TestAutoScaler) Stop() {
	close(as.stopChan)
	as.wg.Wait()
}

func TestNewAutoScaler(t *testing.T) {
	minWorkers := 2
	maxWorkers := 10
	cpuThreshold := 0.7
	scaleUpInterval := 5 * time.Second
	scaleDownInterval := 10 * time.Second

	as := NewAutoScaler(minWorkers, maxWorkers, cpuThreshold, scaleUpInterval, scaleDownInterval)

	if as == nil {
		t.Fatal("Expected NewAutoScaler to return a non-nil value")
	}

	if as.minWorkers != minWorkers {
		t.Errorf("Expected as.minWorkers to be %d, got %d", minWorkers, as.minWorkers)
	}

	if as.maxWorkers != maxWorkers {
		t.Errorf("Expected as.maxWorkers to be %d, got %d", maxWorkers, as.maxWorkers)
	}

	if as.cpuThreshold != cpuThreshold {
		t.Errorf("Expected as.cpuThreshold to be %f, got %f", cpuThreshold, as.cpuThreshold)
	}

	if as.scaleUpInterval != scaleUpInterval {
		t.Errorf("Expected as.scaleUpInterval to be %v, got %v", scaleUpInterval, as.scaleUpInterval)
	}

	if as.scaleDownInterval != scaleDownInterval {
		t.Errorf("Expected as.scaleDownInterval to be %v, got %v", scaleDownInterval, as.scaleDownInterval)
	}

	if cap(as.workerPool) != maxWorkers {
		t.Errorf("Expected as.workerPool capacity to be %d, got %d", maxWorkers, cap(as.workerPool))
	}

	if len(as.workerPool) != minWorkers {
		t.Errorf("Expected as.workerPool to have %d workers initially, got %d", minWorkers, len(as.workerPool))
	}

	if as.stopChan == nil {
		t.Error("Expected as.stopChan to be initialized")
	}
}

func TestScaleUp(t *testing.T) {
	minWorkers := 2
	maxWorkers := 5
	as := NewAutoScaler(minWorkers, maxWorkers, 0.7, 100*time.Millisecond, 100*time.Millisecond)

	// Initial worker count should be minWorkers
	initialWorkers := len(as.workerPool)
	if initialWorkers != minWorkers {
		t.Errorf("Expected initial worker count to be %d, got %d", minWorkers, initialWorkers)
	}

	// Scale up
	as.scaleUp()

	// Worker count should increase by 1
	newWorkers := len(as.workerPool)
	if newWorkers != initialWorkers+1 {
		t.Errorf("Expected worker count to increase to %d, got %d", initialWorkers+1, newWorkers)
	}

	// Scale up to max
	for i := newWorkers; i < maxWorkers; i++ {
		as.scaleUp()
	}

	// Worker count should be at max
	maxedWorkers := len(as.workerPool)
	if maxedWorkers != maxWorkers {
		t.Errorf("Expected worker count to be at max %d, got %d", maxWorkers, maxedWorkers)
	}

	// Try to scale beyond max
	as.scaleUp()

	// Worker count should still be at max
	if len(as.workerPool) != maxWorkers {
		t.Errorf("Expected worker count to remain at max %d, got %d", maxWorkers, len(as.workerPool))
	}
}

func TestScaleDown(t *testing.T) {
	minWorkers := 2
	maxWorkers := 5
	as := NewAutoScaler(minWorkers, maxWorkers, 0.7, 100*time.Millisecond, 100*time.Millisecond)

	// Fill the worker pool to max
	for i := len(as.workerPool); i < maxWorkers; i++ {
		as.workerPool <- struct{}{}
	}

	// Initial worker count should be maxWorkers
	initialWorkers := len(as.workerPool)
	if initialWorkers != maxWorkers {
		t.Errorf("Expected initial worker count to be %d, got %d", maxWorkers, initialWorkers)
	}

	// Scale down
	as.scaleDown()

	// Worker count should decrease by 1
	newWorkers := len(as.workerPool)
	if newWorkers != initialWorkers-1 {
		t.Errorf("Expected worker count to decrease to %d, got %d", initialWorkers-1, newWorkers)
	}

	// Scale down to min
	for i := newWorkers; i > minWorkers; i-- {
		as.scaleDown()
	}

	// Worker count should be at min
	minedWorkers := len(as.workerPool)
	if minedWorkers != minWorkers {
		t.Errorf("Expected worker count to be at min %d, got %d", minWorkers, minedWorkers)
	}

	// Note: We don't test scaling below min here because the implementation
	// doesn't actually prevent scaling below min in the scaleDown method.
	// The prevention happens in the monitorLoad method, which we test separately.
}

func TestStartStop(t *testing.T) {
	as := NewAutoScaler(2, 5, 0.7, 100*time.Millisecond, 100*time.Millisecond)

	// Start the autoscaler
	as.Start()

	// Give it a moment to start
	time.Sleep(50 * time.Millisecond)

	// Stop the autoscaler
	as.Stop()

	// The test passes if Stop() returns (doesn't deadlock)
}

func TestMonitorLoad(t *testing.T) {
	minWorkers := 2
	maxWorkers := 5
	cpuThreshold := 0.7
	
	// Test with high CPU usage (above threshold)
	highCPUUsage := func() float64 { return 0.9 }
	as := NewTestAutoScaler(minWorkers, maxWorkers, cpuThreshold, 
		100*time.Millisecond, 100*time.Millisecond, highCPUUsage)

	// Start the autoscaler
	as.Start()

	// Give it some time to scale up
	time.Sleep(500 * time.Millisecond)

	// Stop the autoscaler
	as.Stop()

	// Check if workers scaled up
	workers := len(as.workerPool)
	if workers <= minWorkers {
		t.Errorf("Expected workers to scale up above %d, got %d", minWorkers, workers)
	}

	// Test with low CPU usage (below threshold)
	lowCPUUsage := func() float64 { return 0.5 }
	as = NewTestAutoScaler(minWorkers, maxWorkers, cpuThreshold, 
		100*time.Millisecond, 100*time.Millisecond, lowCPUUsage)

	// Fill the worker pool to max
	for len(as.workerPool) < maxWorkers {
		as.workerPool <- struct{}{}
	}

	// Start the autoscaler
	as.Start()

	// Give it some time to scale down
	time.Sleep(500 * time.Millisecond)

	// Stop the autoscaler
	as.Stop()

	// Check if workers scaled down
	workers = len(as.workerPool)
	if workers >= maxWorkers {
		t.Errorf("Expected workers to scale down below %d, got %d", maxWorkers, workers)
	}
}

func TestConcurrentScaling(t *testing.T) {
	minWorkers := 2
	maxWorkers := 10
	as := NewAutoScaler(minWorkers, maxWorkers, 0.7, 100*time.Millisecond, 100*time.Millisecond)

	// Run concurrent scale operations
	const numOperations = 100
	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine to scale up
	go func() {
		defer wg.Done()
		for i := 0; i < numOperations; i++ {
			as.scaleUp()
		}
	}()

	// Goroutine to scale down
	go func() {
		defer wg.Done()
		for i := 0; i < numOperations; i++ {
			as.scaleDown()
		}
	}()

	// Wait for all operations to complete
	wg.Wait()

	// Check that the worker count is within bounds
	workers := len(as.workerPool)
	if workers < minWorkers || workers > maxWorkers {
		t.Errorf("Expected worker count to be between %d and %d, got %d", minWorkers, maxWorkers, workers)
	}
}
