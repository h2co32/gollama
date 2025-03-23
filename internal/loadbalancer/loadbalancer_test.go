package loadbalancer

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewLoadBalancer(t *testing.T) {
	servers := []string{"server1:8080", "server2:8080", "server3:8080"}
	healthCheckFreq := 5 * time.Second
	failureThreshold := 3

	lb := NewLoadBalancer(servers, healthCheckFreq, failureThreshold)

	if lb == nil {
		t.Fatal("Expected NewLoadBalancer to return a non-nil value")
	}

	if len(lb.servers) != len(servers) {
		t.Errorf("Expected lb.servers to have length %d, got %d", len(servers), len(lb.servers))
	}

	for i, server := range lb.servers {
		if server != servers[i] {
			t.Errorf("Expected lb.servers[%d] to be '%s', got '%s'", i, servers[i], server)
		}
	}

	if lb.currentIndex != 0 {
		t.Errorf("Expected lb.currentIndex to be 0, got %d", lb.currentIndex)
	}

	if lb.healthCheckFreq != healthCheckFreq {
		t.Errorf("Expected lb.healthCheckFreq to be %v, got %v", healthCheckFreq, lb.healthCheckFreq)
	}

	if lb.failureThreshold != failureThreshold {
		t.Errorf("Expected lb.failureThreshold to be %d, got %d", failureThreshold, lb.failureThreshold)
	}

	// Check that all servers are initially marked as healthy
	for _, server := range servers {
		if !lb.healthChecks[server] {
			t.Errorf("Expected server '%s' to be marked as healthy", server)
		}
	}
}

func TestGetHealthyServer(t *testing.T) {
	servers := []string{"server1:8080", "server2:8080", "server3:8080"}
	lb := NewLoadBalancer(servers, 5*time.Second, 3)

	// Test round-robin behavior with all servers healthy
	for i := 0; i < len(servers)*2; i++ {
		server, err := lb.GetHealthyServer()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		expectedServer := servers[i%len(servers)]
		if server != expectedServer {
			t.Errorf("Expected server to be '%s', got '%s'", expectedServer, server)
		}
	}

	// Test with some servers unhealthy
	lb.healthChecks["server1:8080"] = false
	lb.healthChecks["server3:8080"] = false

	// Now only server2 is healthy, so it should always be returned
	for i := 0; i < 5; i++ {
		server, err := lb.GetHealthyServer()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		expectedServer := "server2:8080"
		if server != expectedServer {
			t.Errorf("Expected server to be '%s', got '%s'", expectedServer, server)
		}
	}

	// Test with all servers unhealthy
	lb.healthChecks["server2:8080"] = false

	_, err := lb.GetHealthyServer()
	if err == nil {
		t.Error("Expected error when all servers are unhealthy, got nil")
	}
}

func TestHealthCheckServers(t *testing.T) {
	// Create test servers
	healthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer healthyServer.Close()

	unhealthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer unhealthyServer.Close()

	// Extract host:port from the test server URLs
	healthyServerAddr := healthyServer.URL[7:] // Remove "http://"
	unhealthyServerAddr := unhealthyServer.URL[7:]

	// Create a load balancer with the test servers
	servers := []string{healthyServerAddr, unhealthyServerAddr}
	lb := NewLoadBalancer(servers, 5*time.Second, 1)

	// Run health checks
	lb.HealthCheckServers()

	// Allow some time for the health checks to complete
	time.Sleep(100 * time.Millisecond)

	// Verify the health status
	if !lb.healthChecks[healthyServerAddr] {
		t.Errorf("Expected healthy server '%s' to be marked as healthy", healthyServerAddr)
	}

	if lb.healthChecks[unhealthyServerAddr] {
		t.Errorf("Expected unhealthy server '%s' to be marked as unhealthy", unhealthyServerAddr)
	}
}

// TestPingServer tests the pingServer method
func TestPingServer(t *testing.T) {
	// Create test servers
	healthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer healthyServer.Close()

	unhealthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer unhealthyServer.Close()

	// Extract host:port from the test server URLs
	healthyServerAddr := healthyServer.URL[7:] // Remove "http://"
	unhealthyServerAddr := unhealthyServer.URL[7:]

	// Create a load balancer
	lb := NewLoadBalancer([]string{healthyServerAddr, unhealthyServerAddr}, 5*time.Second, 3)

	// Test pingServer with healthy server
	result := lb.pingServer(healthyServerAddr)
	if !result {
		t.Errorf("Expected pingServer to return true for healthy server '%s'", healthyServerAddr)
	}

	// Test pingServer with unhealthy server
	result = lb.pingServer(unhealthyServerAddr)
	if result {
		t.Errorf("Expected pingServer to return false for unhealthy server '%s'", unhealthyServerAddr)
	}

	// Test pingServer with non-existent server
	result = lb.pingServer("non-existent-server:8080")
	if result {
		t.Error("Expected pingServer to return false for non-existent server")
	}
}

// TestPingServerWithRetries tests the pingServerWithRetries method
func TestPingServerWithRetries(t *testing.T) {
	// Create a test server that fails the first two requests then succeeds
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			requestCount++
			if requestCount <= 2 {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Extract host:port from the test server URL
	serverAddr := server.URL[7:] // Remove "http://"

	// Create a load balancer
	lb := NewLoadBalancer([]string{serverAddr}, 5*time.Second, 3)

	// Test with max retries = 1 (should fail)
	requestCount = 0
	result := lb.pingServerWithRetries(serverAddr, 1)
	if result {
		t.Error("Expected pingServerWithRetries to return false with max retries = 1")
	}

	// Test with max retries = 3 (should succeed on the 3rd try)
	requestCount = 0
	result = lb.pingServerWithRetries(serverAddr, 3)
	if !result {
		t.Error("Expected pingServerWithRetries to return true with max retries = 3")
	}
	if requestCount != 3 {
		t.Errorf("Expected 3 requests, got %d", requestCount)
	}
}

// TestConcurrentAccess tests that the load balancer handles concurrent access correctly
func TestConcurrentAccess(t *testing.T) {
	servers := []string{"server1:8080", "server2:8080", "server3:8080"}
	lb := NewLoadBalancer(servers, 5*time.Second, 3)

	// Run multiple goroutines to access the load balancer concurrently
	const numGoroutines = 10
	const numRequests = 100
	done := make(chan bool)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numRequests; j++ {
				// Get a server
				_, _ = lb.GetHealthyServer()

				// Toggle a server's health (to test concurrent writes)
				server := servers[j%len(servers)]
				lb.lock.Lock()
				lb.healthChecks[server] = !lb.healthChecks[server]
				lb.lock.Unlock()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// If we got here without panicking, the test passes
}
