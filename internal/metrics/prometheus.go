package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsProvider holds Prometheus metrics collectors for tracking request metrics
type MetricsProvider struct {
	requestCount   *prometheus.CounterVec
	requestLatency *prometheus.HistogramVec
	errorCount     *prometheus.CounterVec
}

// NewMetricsProvider initializes and registers Prometheus metrics
func NewMetricsProvider() *MetricsProvider {
	mp := &MetricsProvider{
		requestCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "requests_total",
				Help: "Total number of requests processed, labeled by endpoint and status.",
			},
			[]string{"endpoint", "status"},
		),
		requestLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "request_latency_seconds",
				Help:    "Request latency in seconds, labeled by endpoint.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"endpoint"},
		),
		errorCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "errors_total",
				Help: "Total number of errors encountered, labeled by endpoint and error type.",
			},
			[]string{"endpoint", "error_type"},
		),
	}

	// Register metrics with Prometheus
	prometheus.MustRegister(mp.requestCount)
	prometheus.MustRegister(mp.requestLatency)
	prometheus.MustRegister(mp.errorCount)

	return mp
}

// TrackRequest increments the request counter and records latency
func (mp *MetricsProvider) TrackRequest(endpoint, status string, duration time.Duration) {
	mp.requestCount.WithLabelValues(endpoint, status).Inc()
	mp.requestLatency.WithLabelValues(endpoint).Observe(duration.Seconds())
}

// TrackError increments the error counter for the specified error type
func (mp *MetricsProvider) TrackError(endpoint, errorType string) {
	mp.errorCount.WithLabelValues(endpoint, errorType).Inc()
}

// ServeMetrics provides an HTTP endpoint for Prometheus to scrape metrics
func (mp *MetricsProvider) ServeMetrics(port int) {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		addr := fmt.Sprintf(":%d", port)
		fmt.Printf("Serving Prometheus metrics on %s/metrics\n", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			fmt.Printf("Error starting Prometheus metrics server: %v\n", err)
		}
	}()
}
