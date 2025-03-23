package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/h2co32/gollama/pkg/auth"
	"github.com/golang-jwt/jwt/v4"
)

func TestNewAuthMiddleware(t *testing.T) {
	options := AuthOptions{
		AuthType:   AuthTypeJWT,
		JWTSecret:  "jwt-secret",
		HMACSecret: "hmac-secret",
	}

	middleware := NewAuthMiddleware(options)

	if middleware == nil {
		t.Fatal("NewAuthMiddleware returned nil")
	}

	if middleware.options.AuthType != options.AuthType {
		t.Errorf("Expected AuthType to be %s, got %s", options.AuthType, middleware.options.AuthType)
	}

	if middleware.options.JWTSecret != options.JWTSecret {
		t.Errorf("Expected JWTSecret to be %s, got %s", options.JWTSecret, middleware.options.JWTSecret)
	}

	if middleware.options.HMACSecret != options.HMACSecret {
		t.Errorf("Expected HMACSecret to be %s, got %s", options.HMACSecret, middleware.options.HMACSecret)
	}
}

func TestJWTAuthMiddleware(t *testing.T) {
	// Create a JWT auth middleware
	options := AuthOptions{
		AuthType:   AuthTypeJWT,
		JWTSecret:  "jwt-secret",
		HMACSecret: "hmac-secret",
	}

	middleware := NewAuthMiddleware(options)

	// Create a test handler that returns the user claims from the context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetUserFromContext(r.Context())
		if !ok {
			http.Error(w, "No user claims in context", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(claims)
	})

	// Create a protected handler with the middleware
	protectedHandler := middleware.Middleware(testHandler)

	// Create a test request with a valid JWT token
	claims := map[string]interface{}{
		"user_id": 123,
		"role":    "admin",
	}
	token, err := auth.GenerateJWT(options.JWTSecret, claims)
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set(AuthHeaderKey, "Bearer "+token)
	recorder := httptest.NewRecorder()

	// Call the protected handler
	protectedHandler.ServeHTTP(recorder, req)

	// Check the response
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	// Parse the response body
	var responseClaims jwt.MapClaims
	if err := json.NewDecoder(recorder.Body).Decode(&responseClaims); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// Check that the claims were correctly passed through the context
	if responseClaims["user_id"] != float64(123) {
		t.Errorf("Expected user_id claim to be 123, got %v", responseClaims["user_id"])
	}

	if responseClaims["role"] != "admin" {
		t.Errorf("Expected role claim to be 'admin', got %v", responseClaims["role"])
	}

	// Test with invalid token
	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set(AuthHeaderKey, "Bearer invalid-token")
	recorder = httptest.NewRecorder()

	protectedHandler.ServeHTTP(recorder, req)

	// Should return unauthorized
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d for invalid token, got %d", http.StatusUnauthorized, recorder.Code)
	}

	// Test with missing token
	req = httptest.NewRequest("GET", "/protected", nil)
	recorder = httptest.NewRecorder()

	protectedHandler.ServeHTTP(recorder, req)

	// Should return unauthorized
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d for missing token, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestHMACAuthMiddleware(t *testing.T) {
	// Create an HMAC auth middleware
	options := AuthOptions{
		AuthType:   AuthTypeHMAC,
		JWTSecret:  "jwt-secret",
		HMACSecret: "hmac-secret",
	}

	middleware := NewAuthMiddleware(options)

	// Create a test handler that returns success
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	})

	// Create a protected handler with the middleware
	protectedHandler := middleware.Middleware(testHandler)

	// Create a test request with a valid HMAC signature
	body := `{"action":"test"}`
	signature := auth.GenerateHMAC(options.HMACSecret, body)

	req := httptest.NewRequest("POST", "/protected", nil)
	req.Header.Set(HMACHeaderKey, signature)
	req.Body = httptest.NewRequest("POST", "/protected", nil).Body // Empty body for this test
	recorder := httptest.NewRecorder()

	// Call the protected handler - should fail because the body is empty
	protectedHandler.ServeHTTP(recorder, req)

	// Should return unauthorized because the body doesn't match the signature
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d for body mismatch, got %d", http.StatusUnauthorized, recorder.Code)
	}

	// Test with missing signature
	req = httptest.NewRequest("POST", "/protected", nil)
	recorder = httptest.NewRecorder()

	protectedHandler.ServeHTTP(recorder, req)

	// Should return unauthorized
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d for missing signature, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestUnsupportedAuthType(t *testing.T) {
	// Create a middleware with an unsupported auth type
	options := AuthOptions{
		AuthType:   "unsupported",
		JWTSecret:  "jwt-secret",
		HMACSecret: "hmac-secret",
	}

	middleware := NewAuthMiddleware(options)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create a protected handler with the middleware
	protectedHandler := middleware.Middleware(testHandler)

	// Create a test request
	req := httptest.NewRequest("GET", "/protected", nil)
	recorder := httptest.NewRecorder()

	// Call the protected handler
	protectedHandler.ServeHTTP(recorder, req)

	// Should return unauthorized
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d for unsupported auth type, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestCustomErrorHandler(t *testing.T) {
	// Create a middleware with a custom error handler
	options := AuthOptions{
		AuthType:   AuthTypeJWT,
		JWTSecret:  "jwt-secret",
		HMACSecret: "hmac-secret",
		ErrorHandler: func(w http.ResponseWriter, err error) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Custom error: " + err.Error()))
		},
	}

	middleware := NewAuthMiddleware(options)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create a protected handler with the middleware
	protectedHandler := middleware.Middleware(testHandler)

	// Create a test request with an invalid token
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set(AuthHeaderKey, "Bearer invalid-token")
	recorder := httptest.NewRecorder()

	// Call the protected handler
	protectedHandler.ServeHTTP(recorder, req)

	// Should return forbidden (custom error handler)
	if recorder.Code != http.StatusForbidden {
		t.Errorf("Expected status code %d for custom error handler, got %d", http.StatusForbidden, recorder.Code)
	}

	// Should contain the custom error message
	if body := recorder.Body.String(); body[:13] != "Custom error:" {
		t.Errorf("Expected response to start with 'Custom error:', got '%s'", body)
	}
}

func TestGetUserFromContext(t *testing.T) {
	// Create a context with user claims
	claims := jwt.MapClaims{
		"user_id": float64(123),
		"role":    "admin",
	}
	ctx := context.WithValue(context.Background(), UserContextKey, claims)

	// Get the claims from the context
	retrievedClaims, ok := GetUserFromContext(ctx)
	if !ok {
		t.Fatal("GetUserFromContext returned false")
	}

	// Check that the claims match
	if retrievedClaims["user_id"] != float64(123) {
		t.Errorf("Expected user_id claim to be 123, got %v", retrievedClaims["user_id"])
	}

	if retrievedClaims["role"] != "admin" {
		t.Errorf("Expected role claim to be 'admin', got %v", retrievedClaims["role"])
	}

	// Test with a context that doesn't have user claims
	_, ok = GetUserFromContext(context.Background())
	if ok {
		t.Error("GetUserFromContext should return false for context without claims")
	}
}

func TestJSONResponse(t *testing.T) {
	// Test JSON response with data
	recorder := httptest.NewRecorder()
	data := map[string]string{"message": "test"}
	JSONResponse(recorder, http.StatusOK, data)

	// Check status code
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	// Check content type
	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type to be 'application/json', got '%s'", contentType)
	}

	// Check body
	var response map[string]string
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if response["message"] != "test" {
		t.Errorf("Expected message to be 'test', got '%s'", response["message"])
	}

	// Test JSON response with nil data
	recorder = httptest.NewRecorder()
	JSONResponse(recorder, http.StatusNoContent, nil)

	// Check status code
	if recorder.Code != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, recorder.Code)
	}

	// Check content type
	contentType = recorder.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type to be 'application/json', got '%s'", contentType)
	}

	// Check body (should be empty)
	if recorder.Body.Len() != 0 {
		t.Errorf("Expected empty body for nil data, got '%s'", recorder.Body.String())
	}
}
