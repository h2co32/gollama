package loadbalancer

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// LoadBalancer manages a set of servers, routing requests to healthy ones
type LoadBalancer struct {
	servers          []string        // List of server URLs
	currentIndex     int             // Round-robin index
	healthChecks     map[string]bool // Server health status
	lock             sync.Mutex      // Mutex for concurrent access
	healthCheckFreq  time.Duration   // Frequency of health checks
	failureThreshold int             // Number of consecutive failures before marking a server as unhealthy
}

// NewLoadBalancer initializes a LoadBalancer with a list of servers and health check settings
func NewLoadBalancer(servers []string, healthCheckFreq time.Duration, failureThreshold int) *LoadBalancer {
	lb := &LoadBalancer{
		servers:          servers,
		currentIndex:     0,
		healthChecks:     make(map[string]bool),
		healthCheckFreq:  healthCheckFreq,
		failureThreshold: failureThreshold,
	}

	for _, server := range servers {
		lb.healthChecks[server] = true // Initialize all servers as healthy
	}

	go lb.startHealthChecks()

	return lb
}

// GetHealthyServer returns the next available healthy server in a round-robin fashion
func (lb *LoadBalancer) GetHealthyServer() (string, error) {
	lb.lock.Lock()
	defer lb.lock.Unlock()

	// Try each server in the list once, using round-robin
	for i := 0; i < len(lb.servers); i++ {
		server := lb.servers[lb.currentIndex]
		lb.currentIndex = (lb.currentIndex + 1) % len(lb.servers)

		if lb.healthChecks[server] {
			return server, nil
		}
	}

	return "", fmt.Errorf("no healthy servers available")
}

// startHealthChecks initiates periodic health checks on all servers
func (lb *LoadBalancer) startHealthChecks() {
	ticker := time.NewTicker(lb.healthCheckFreq)
	defer ticker.Stop()

	for range ticker.C {
		lb.HealthCheckServers()
	}
}

// HealthCheckServers performs concurrent health checks on all servers
func (lb *LoadBalancer) HealthCheckServers() {
	var wg sync.WaitGroup
	wg.Add(len(lb.servers))

	for _, server := range lb.servers {
		go func(server string) {
			defer wg.Done()
			isHealthy := lb.pingServerWithRetries(server, lb.failureThreshold)

			lb.lock.Lock()
			lb.healthChecks[server] = isHealthy
			lb.lock.Unlock()
		}(server)
	}

	wg.Wait()
}

// pingServerWithRetries checks server health with retries up to a failure threshold
func (lb *LoadBalancer) pingServerWithRetries(server string, maxRetries int) bool {
	for i := 0; i < maxRetries; i++ {
		if lb.pingServer(server) {
			return true
		}
		time.Sleep(100 * time.Millisecond) // Optional backoff between retries
	}
	return false
}

// pingServer checks if a server is reachable and returns true if healthy
func (lb *LoadBalancer) pingServer(server string) bool {
	client := http.Client{
		Timeout: 2 * time.Second, // Timeout for each ping attempt
	}

	res, err := client.Get(fmt.Sprintf("http://%s/health", server))
	if err != nil || res.StatusCode != http.StatusOK {
		return false
	}
	return true
}
