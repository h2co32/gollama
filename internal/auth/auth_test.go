package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewAuthMiddleware(t *testing.T) {
	config := AuthConfig{
		AuthType: JWTAuth,
		JWTToken: "test-token",
	}
	
	middleware := NewAuthMiddleware(config)
	
	if middleware == nil {
		t.Fatal("Expected NewAuthMiddleware to return a non-nil value")
	}
	
	if middleware.config.AuthType != JWTAuth {
		t.Errorf("Expected middleware.config.AuthType to be %v, got %v", JWTAuth, middleware.config.AuthType)
	}
	
	if middleware.config.JWTToken != "test-token" {
		t.Errorf("Expected middleware.config.JWTToken to be 'test-token', got '%s'", middleware.config.JWTToken)
	}
}

func TestJWTAuth(t *testing.T) {
	// Test with valid JWT token
	validConfig := AuthConfig{
		AuthType:     JWTAuth,
		JWTToken:     "valid-jwt-token",
		JWTExpiresAt: time.Now().Add(1 * time.Hour), // Valid for 1 hour
	}
	
	middleware := NewAuthMiddleware(validConfig)
	req, _ := http.NewRequest("GET", "https://example.com/api", nil)
	
	processedReq, err := middleware.ProcessRequest(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	authHeader := processedReq.Header.Get("Authorization")
	expectedHeader := "Bearer valid-jwt-token"
	if authHeader != expectedHeader {
		t.Errorf("Expected Authorization header to be '%s', got '%s'", expectedHeader, authHeader)
	}
	
	// Test with expired JWT token
	expiredConfig := AuthConfig{
		AuthType:     JWTAuth,
		JWTToken:     "expired-jwt-token",
		JWTExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}
	
	expiredMiddleware := NewAuthMiddleware(expiredConfig)
	req, _ = http.NewRequest("GET", "https://example.com/api", nil)
	
	_, err = expiredMiddleware.ProcessRequest(req)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
	
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("Expected error message to contain 'expired', got '%s'", err.Error())
	}
}

func TestHMACAuth(t *testing.T) {
	// Test with valid HMAC key
	validConfig := AuthConfig{
		AuthType: HMACAuth,
		HMACKey:  "secret-hmac-key",
	}
	
	middleware := NewAuthMiddleware(validConfig)
	req, _ := http.NewRequest("GET", "https://example.com/api", nil)
	
	processedReq, err := middleware.ProcessRequest(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Verify the signature
	signature := processedReq.Header.Get("X-Signature")
	if signature == "" {
		t.Error("Expected X-Signature header to be set")
	}
	
	// Manually calculate the expected signature
	mac := hmac.New(sha256.New, []byte("secret-hmac-key"))
	mac.Write([]byte(req.URL.String()))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	
	if signature != expectedSignature {
		t.Errorf("Expected signature to be '%s', got '%s'", expectedSignature, signature)
	}
	
	// Test with missing HMAC key
	invalidConfig := AuthConfig{
		AuthType: HMACAuth,
		HMACKey:  "", // Empty key
	}
	
	invalidMiddleware := NewAuthMiddleware(invalidConfig)
	req, _ = http.NewRequest("GET", "https://example.com/api", nil)
	
	_, err = invalidMiddleware.ProcessRequest(req)
	if err == nil {
		t.Error("Expected error for missing HMAC key, got nil")
	}
	
	if !strings.Contains(err.Error(), "missing") {
		t.Errorf("Expected error message to contain 'missing', got '%s'", err.Error())
	}
}

func TestUnsupportedAuthType(t *testing.T) {
	// Test with unsupported auth type
	invalidConfig := AuthConfig{
		AuthType: AuthType(999), // Invalid auth type
	}
	
	middleware := NewAuthMiddleware(invalidConfig)
	req, _ := http.NewRequest("GET", "https://example.com/api", nil)
	
	_, err := middleware.ProcessRequest(req)
	if err == nil {
		t.Error("Expected error for unsupported auth type, got nil")
	}
	
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("Expected error message to contain 'unsupported', got '%s'", err.Error())
	}
}

func TestGenerateHMACSignature(t *testing.T) {
	// Test HMAC signature generation
	data := "test-data"
	key := "test-key"
	
	signature, err := GenerateHMACSignature(data, key)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	// Manually calculate the expected signature
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	expectedSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	
	if signature != expectedSignature {
		t.Errorf("Expected signature to be '%s', got '%s'", expectedSignature, signature)
	}
}

// Mock HTTP server for OAuth tests
type mockOAuthServer struct {
	server *httptest.Server
}

func newMockOAuthServer() *mockOAuthServer {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		
		if r.URL.Path != "/oauth/token" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		
		// Check basic auth
		username, password, ok := r.BasicAuth()
		if !ok || username != "test-client-id" || password != "test-client-secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		
		// Return a mock token response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token":"new-oauth-token","expires_in":3600}`))
	})
	
	server := httptest.NewServer(handler)
	return &mockOAuthServer{server: server}
}

func (m *mockOAuthServer) Close() {
	m.server.Close()
}

func TestOAuthAuth(t *testing.T) {
	// Start mock OAuth server
	mockServer := newMockOAuthServer()
	defer mockServer.Close()
	
	// Test with valid OAuth token
	validConfig := AuthConfig{
		AuthType:       OAuthAuth,
		OAuthToken:     "valid-oauth-token",
		OAuthExpiresAt: time.Now().Add(1 * time.Hour), // Valid for 1 hour
	}
	
	middleware := NewAuthMiddleware(validConfig)
	req, _ := http.NewRequest("GET", "https://example.com/api", nil)
	
	processedReq, err := middleware.ProcessRequest(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	authHeader := processedReq.Header.Get("Authorization")
	expectedHeader := "Bearer valid-oauth-token"
	if authHeader != expectedHeader {
		t.Errorf("Expected Authorization header to be '%s', got '%s'", expectedHeader, authHeader)
	}
	
	// Test with expired OAuth token and valid refresh config
	expiredConfig := AuthConfig{
		AuthType:       OAuthAuth,
		OAuthToken:     "expired-oauth-token",
		OAuthExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		OAuthTokenURL:  mockServer.server.URL + "/oauth/token",
		ClientID:       "test-client-id",
		ClientSecret:   "test-client-secret",
	}
	
	expiredMiddleware := NewAuthMiddleware(expiredConfig)
	req, _ = http.NewRequest("GET", "https://example.com/api", nil)
	
	// This should trigger a token refresh
	processedReq, err = expiredMiddleware.ProcessRequest(req)
	if err != nil {
		t.Errorf("Expected no error after token refresh, got %v", err)
	}
	
	authHeader = processedReq.Header.Get("Authorization")
	expectedHeader = "Bearer new-oauth-token"
	if authHeader != expectedHeader {
		t.Errorf("Expected Authorization header to be '%s', got '%s'", expectedHeader, authHeader)
	}
	
	// Verify the token was updated in the config
	if expiredMiddleware.config.OAuthToken != "new-oauth-token" {
		t.Errorf("Expected config.OAuthToken to be updated to 'new-oauth-token', got '%s'", expiredMiddleware.config.OAuthToken)
	}
	
	// Verify the expiration was updated
	if time.Until(expiredMiddleware.config.OAuthExpiresAt) < 3500*time.Second {
		t.Errorf("Expected config.OAuthExpiresAt to be updated to ~1 hour in the future")
	}
}
