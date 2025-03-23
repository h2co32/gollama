package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func TestGenerateJWT(t *testing.T) {
	secretKey := "test-secret-key"
	claims := map[string]interface{}{
		"user_id": 123,
		"role":    "admin",
	}

	token, err := GenerateJWT(secretKey, claims)
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}

	if token == "" {
		t.Error("GenerateJWT returned empty token")
	}

	// Validate the token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		t.Fatalf("Failed to parse generated token: %v", err)
	}

	if !parsedToken.Valid {
		t.Error("Generated token is not valid")
	}

	// Check claims
	parsedClaims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("Failed to extract claims from token")
	}

	if parsedClaims["user_id"] != float64(123) {
		t.Errorf("Expected user_id claim to be 123, got %v", parsedClaims["user_id"])
	}

	if parsedClaims["role"] != "admin" {
		t.Errorf("Expected role claim to be 'admin', got %v", parsedClaims["role"])
	}
}

func TestGenerateJWTWithOptions(t *testing.T) {
	secretKey := "test-secret-key"
	claims := map[string]interface{}{
		"user_id": 123,
	}

	options := JWTOptions{
		ExpiresIn: 30 * time.Minute,
		Issuer:    "test-issuer",
		Audience:  "test-audience",
	}

	token, err := GenerateJWTWithOptions(secretKey, claims, options)
	if err != nil {
		t.Fatalf("GenerateJWTWithOptions failed: %v", err)
	}

	// Validate the token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		t.Fatalf("Failed to parse generated token: %v", err)
	}

	parsedClaims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("Failed to extract claims from token")
	}

	// Check standard claims
	if parsedClaims["iss"] != options.Issuer {
		t.Errorf("Expected issuer claim to be '%s', got %v", options.Issuer, parsedClaims["iss"])
	}

	if parsedClaims["aud"] != options.Audience {
		t.Errorf("Expected audience claim to be '%s', got %v", options.Audience, parsedClaims["aud"])
	}

	// Check expiration time
	now := time.Now().Unix()
	exp, ok := parsedClaims["exp"].(float64)
	if !ok {
		t.Fatal("Failed to extract exp claim from token")
	}

	// The expiration time should be approximately 30 minutes from now
	expectedExp := now + int64(30*time.Minute.Seconds())
	if int64(exp) < expectedExp-60 || int64(exp) > expectedExp+60 {
		t.Errorf("Expected exp claim to be around %d, got %v", expectedExp, exp)
	}

	// Check custom claims
	if parsedClaims["user_id"] != float64(123) {
		t.Errorf("Expected user_id claim to be 123, got %v", parsedClaims["user_id"])
	}
}

func TestValidateJWT(t *testing.T) {
	secretKey := "test-secret-key"
	claims := map[string]interface{}{
		"user_id": 123,
		"role":    "admin",
	}

	token, err := GenerateJWT(secretKey, claims)
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}

	// Validate the token
	validatedClaims, err := ValidateJWT(secretKey, token)
	if err != nil {
		t.Fatalf("ValidateJWT failed: %v", err)
	}

	// Check claims
	if validatedClaims["user_id"] != float64(123) {
		t.Errorf("Expected user_id claim to be 123, got %v", validatedClaims["user_id"])
	}

	if validatedClaims["role"] != "admin" {
		t.Errorf("Expected role claim to be 'admin', got %v", validatedClaims["role"])
	}

	// Test with invalid token
	_, err = ValidateJWT(secretKey, token+"invalid")
	if err == nil {
		t.Error("ValidateJWT should fail with invalid token")
	}

	// Test with wrong secret key
	_, err = ValidateJWT("wrong-secret", token)
	if err == nil {
		t.Error("ValidateJWT should fail with wrong secret key")
	}

	// Test with empty secret key
	_, err = ValidateJWT("", token)
	if err == nil {
		t.Error("ValidateJWT should fail with empty secret key")
	}

	// Test with empty token
	_, err = ValidateJWT(secretKey, "")
	if err == nil {
		t.Error("ValidateJWT should fail with empty token")
	}
}

func TestGenerateAndValidateHMAC(t *testing.T) {
	secretKey := "test-hmac-secret"
	data := "test-data"

	// Generate HMAC
	signature := GenerateHMAC(secretKey, data)
	if signature == "" {
		t.Error("GenerateHMAC returned empty signature")
	}

	// Validate HMAC
	valid := ValidateHMAC(secretKey, data, signature)
	if !valid {
		t.Error("ValidateHMAC failed for valid signature")
	}

	// Test with wrong data
	valid = ValidateHMAC(secretKey, "wrong-data", signature)
	if valid {
		t.Error("ValidateHMAC should fail with wrong data")
	}

	// Test with wrong secret key
	valid = ValidateHMAC("wrong-secret", data, signature)
	if valid {
		t.Error("ValidateHMAC should fail with wrong secret key")
	}

	// Test with wrong signature
	valid = ValidateHMAC(secretKey, data, signature+"invalid")
	if valid {
		t.Error("ValidateHMAC should fail with wrong signature")
	}
}

func TestExtractBearerToken(t *testing.T) {
	// Valid bearer token
	authHeader := "Bearer token123"
	token, err := ExtractBearerToken(authHeader)
	if err != nil {
		t.Errorf("ExtractBearerToken failed for valid header: %v", err)
	}
	if token != "token123" {
		t.Errorf("Expected token to be 'token123', got '%s'", token)
	}

	// Invalid format (missing "Bearer " prefix)
	authHeader = "token123"
	_, err = ExtractBearerToken(authHeader)
	if err == nil {
		t.Error("ExtractBearerToken should fail for header without Bearer prefix")
	}

	// Empty header
	authHeader = ""
	_, err = ExtractBearerToken(authHeader)
	if err == nil {
		t.Error("ExtractBearerToken should fail for empty header")
	}

	// Too short header
	authHeader = "Bear"
	_, err = ExtractBearerToken(authHeader)
	if err == nil {
		t.Error("ExtractBearerToken should fail for too short header")
	}
}

func TestDefaultJWTOptions(t *testing.T) {
	options := DefaultJWTOptions()

	if options.ExpiresIn != time.Hour {
		t.Errorf("Expected ExpiresIn to be 1 hour, got %v", options.ExpiresIn)
	}

	if options.Issuer != "" {
		t.Errorf("Expected Issuer to be empty, got '%s'", options.Issuer)
	}

	if options.Audience != "" {
		t.Errorf("Expected Audience to be empty, got '%s'", options.Audience)
	}
}
