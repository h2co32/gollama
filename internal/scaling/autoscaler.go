package autoscaler

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// WorkerFunc represents the function that each worker will execute
type WorkerFunc func() error

// AutoScaler manages a worker pool that scales based on system load
type AutoScaler struct {
	workerPool        chan struct{}
	minWorkers        int
	maxWorkers        int
	cpuThreshold      float64
	scaleUpInterval   time.Duration
	scaleDownInterval time.Duration
	wg                sync.WaitGroup
	stopChan          chan struct{}
}

// NewAutoScaler initializes a new AutoScaler with the specified parameters
func NewAutoScaler(minWorkers, maxWorkers int, cpuThreshold float64, scaleUpInterval, scaleDownInterval time.Duration) *AutoScaler {
	as := &AutoScaler{
		workerPool:        make(chan struct{}, maxWorkers),
		minWorkers:        minWorkers,
		maxWorkers:        maxWorkers,
		cpuThreshold:      cpuThreshold,
		scaleUpInterval:   scaleUpInterval,
		scaleDownInterval: scaleDownInterval,
		stopChan:          make(chan struct{}),
	}

	for i := 0; i < minWorkers; i++ {
		as.workerPool <- struct{}{}
	}

	return as
}

// Start begins monitoring system load and scaling workers accordingly
func (as *AutoScaler) Start() {
	go as.monitorLoad()
}

// monitorLoad periodically checks CPU usage and scales workers up or down
func (as *AutoScaler) monitorLoad() {
	for {
		select {
		case <-as.stopChan:
			return
		default:
			cpuUsage := getCPUUsage()
			currentWorkers := len(as.workerPool)

			if cpuUsage > as.cpuThreshold && currentWorkers < as.maxWorkers {
				as.scaleUp()
			} else if cpuUsage < as.cpuThreshold && currentWorkers > as.minWorkers {
				as.scaleDown()
			}

			time.Sleep(2 * time.Second)
		}
	}
}

// scaleUp adds workers up to the maximum limit
func (as *AutoScaler) scaleUp() {
	as.wg.Add(1)
	defer as.wg.Done()

	select {
	case as.workerPool <- struct{}{}:
		fmt.Println("Scaled up, current workers:", len(as.workerPool))
	case <-time.After(as.scaleUpInterval):
		fmt.Println("Scale-up timed out")
	}
}

// scaleDown removes a worker down to the minimum limit
func (as *AutoScaler) scaleDown() {
	as.wg.Add(1)
	defer as.wg.Done()

	select {
	case <-as.workerPool:
		fmt.Println("Scaled down, current workers:", len(as.workerPool))
	case <-time.After(as.scaleDownInterval):
		fmt.Println("Scale-down timed out")
	}
}

// Stop stops the autoscaler
func (as *AutoScaler) Stop() {
	close(as.stopChan)
	as.wg.Wait()
}

// getCPUUsage simulates CPU usage checking (customize for real usage)
func getCPUUsage() float64 {
	// Simulate CPU usage (in production, replace with real monitoring)
	cpuUsage := float64(runtime.NumGoroutine()) / float64(runtime.NumCPU())
	fmt.Printf("CPU Usage: %.2f\n", cpuUsage*100)
	return cpuUsage
}
