package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// LogError logs an error with additional context
func LogError(context string, err error) {
	log.Printf("[ERROR] %s: %v\n", context, err)
}

// LogInfo logs general information messages
func LogInfo(context, message string) {
	log.Printf("[INFO] %s: %s\n", context, message)
}

// JSONResponse sends a JSON response with the specified status code and payload
func JSONResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		LogError("JSONResponse", err)
	}
}

// JSONDecode decodes JSON from an HTTP request body into a target structure
func JSONDecode(r *http.Request, target interface{}) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("%s: %w", ErrInvalidRequestBody, err)
	}
	return nil
}

// Retry executes a function with retries and exponential backoff
func Retry(operation func() error, maxRetries int, initialBackoff time.Duration) error {
	backoff := initialBackoff
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := operation(); err != nil {
			LogError("Retry operation failed", err)
			if attempt == maxRetries {
				return err
			}
			time.Sleep(backoff)
			backoff *= 2
		} else {
			return nil
		}
	}
	return fmt.Errorf("operation failed after %d retries", maxRetries)
}

// GenerateTimestamp generates a timestamp in a standard format
func GenerateTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// ValidateHTTPStatus checks if a status code represents a successful request
func ValidateHTTPStatus(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

// GenerateToken generates a simple, unique token (example only)
func GenerateToken() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// HandlePanic is a defer function to handle panics gracefully
func HandlePanic() {
	if r := recover(); r != nil {
		log.Printf("[PANIC RECOVERED] %v\n", r)
	}
}
