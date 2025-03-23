// Package auth provides authentication utilities for JWT and HMAC authentication.
//
// This package offers functions for generating and validating JWT tokens and HMAC signatures,
// which can be used to secure API endpoints and verify request authenticity.
//
// Example usage:
//
//	// Generate a JWT token
//	claims := map[string]interface{}{"user_id": 123, "role": "admin"}
//	token, err := auth.GenerateJWT("your-secret-key", claims)
//
//	// Validate a JWT token
//	claims, err := auth.ValidateJWT("your-secret-key", token)
//
//	// Generate an HMAC signature
//	signature := auth.GenerateHMAC("your-hmac-key", "data-to-sign")
//
//	// Validate an HMAC signature
//	isValid := auth.ValidateHMAC("your-hmac-key", "data-to-sign", signature)
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Version represents the current package version following semantic versioning.
const Version = "1.0.0"

// JWTOptions configures JWT token generation.
type JWTOptions struct {
	// ExpiresIn is the token expiration duration.
	// Default: 1 hour
	ExpiresIn time.Duration

	// Issuer is the token issuer claim.
	// Optional.
	Issuer string

	// Audience is the token audience claim.
	// Optional.
	Audience string
}

// DefaultJWTOptions returns the default JWT options.
func DefaultJWTOptions() JWTOptions {
	return JWTOptions{
		ExpiresIn: time.Hour,
	}
}

// GenerateJWT creates a new JWT token with the provided claims and secret key.
func GenerateJWT(secretKey string, claims map[string]interface{}) (string, error) {
	return GenerateJWTWithOptions(secretKey, claims, DefaultJWTOptions())
}

// GenerateJWTWithOptions creates a new JWT token with the provided claims, secret key, and options.
func GenerateJWTWithOptions(secretKey string, claims map[string]interface{}, options JWTOptions) (string, error) {
	if secretKey == "" {
		return "", fmt.Errorf("secret key cannot be empty")
	}

	// Create a new token
	tokenClaims := jwt.MapClaims{}

	// Add standard claims
	now := time.Now()
	tokenClaims["iat"] = now.Unix()
	tokenClaims["exp"] = now.Add(options.ExpiresIn).Unix()

	if options.Issuer != "" {
		tokenClaims["iss"] = options.Issuer
	}

	if options.Audience != "" {
		tokenClaims["aud"] = options.Audience
	}

	// Add custom claims
	for key, value := range claims {
		tokenClaims[key] = value
	}

	// Create the token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateJWT validates a JWT token and returns its claims if valid.
func ValidateJWT(secretKey string, tokenString string) (jwt.MapClaims, error) {
	if secretKey == "" {
		return nil, fmt.Errorf("secret key cannot be empty")
	}

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to extract claims")
	}

	return claims, nil
}

// GenerateHMAC generates an HMAC signature for the provided data using the secret key.
func GenerateHMAC(secretKey string, data string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// ValidateHMAC validates an HMAC signature against the provided data and secret key.
func ValidateHMAC(secretKey string, data string, signature string) bool {
	expectedSignature := GenerateHMAC(secretKey, data)
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

// ExtractBearerToken extracts the token from an Authorization header value.
func ExtractBearerToken(authHeader string) (string, error) {
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return "", fmt.Errorf("invalid authorization header format")
	}
	return authHeader[7:], nil
}
