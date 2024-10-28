package utils

import "time"

// HTTP Status Codes
const (
	StatusOK                  = 200
	StatusCreated             = 201
	StatusNoContent           = 204
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusConflict            = 409
	StatusInternalServerError = 500
)

// Error Messages
const (
	ErrUnauthorized       = "unauthorized access"
	ErrForbidden          = "forbidden access"
	ErrNotFound           = "resource not found"
	ErrConflict           = "resource conflict"
	ErrInternalServer     = "internal server error"
	ErrInvalidRequestBody = "invalid request body"
)

// API Endpoints
const (
	ModelDownloadEndpoint = "/api/models/download"
	ModelUploadEndpoint   = "/api/models/upload"
	HealthCheckEndpoint   = "/health"
	MetricsEndpoint       = "/metrics"
)

// JWT Constants
const (
	JWTIssuer     = "myapp"
	JWTSecretKey  = "supersecretkey"
	JWTExpiration = 24 * time.Hour
)

// Rate Limiter Constants
const (
	DefaultRateLimitCapacity = 100
	DefaultRefillRate        = time.Second
	DefaultRefillAmount      = 1
)

// Retry and Backoff Constants
const (
	DefaultMaxRetries     = 5
	DefaultInitialBackoff = 500 * time.Millisecond
	DefaultMaxBackoff     = 10 * time.Second
)

// AutoScaler Constants
const (
	DefaultMinWorkers        = 2
	DefaultMaxWorkers        = 10
	DefaultCPUThreshold      = 0.75
	DefaultScaleUpInterval   = 2 * time.Second
	DefaultScaleDownInterval = 5 * time.Second
)
