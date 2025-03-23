// Package middleware provides HTTP middleware components for authentication and other cross-cutting concerns.
//
// This package contains middleware that can be used with standard Go HTTP servers
// to add authentication, logging, rate limiting, and other functionality to HTTP handlers.
//
// Example usage:
//
//	// Create a new auth middleware for JWT authentication
//	authMiddleware := middleware.NewAuthMiddleware(middleware.AuthOptions{
//		AuthType:   middleware.AuthTypeJWT,
//		JWTSecret:  "your-jwt-secret",
//		HMACSecret: "your-hmac-secret",
//	})
//
//	// Use the middleware with an HTTP handler
//	http.Handle("/protected", authMiddleware.Middleware(http.HandlerFunc(protectedHandler)))
package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/h2co32/gollama/pkg/auth"

	"github.com/golang-jwt/jwt/v4"
)

// Version represents the current package version following semantic versioning.
const Version = "1.0.0"

// Context keys for adding authentication data to the request context
type contextKey string

const (
	// UserContextKey is the context key for storing user information
	UserContextKey contextKey = "user"
	
	// AuthHeaderKey is the HTTP header key for the Authorization header
	AuthHeaderKey string = "Authorization"
	
	// HMACHeaderKey is the HTTP header key for the HMAC signature
	HMACHeaderKey string = "X-Signature"
)

// Authentication types
const (
	// AuthTypeJWT specifies JWT token authentication
	AuthTypeJWT = "jwt"
	
	// AuthTypeHMAC specifies HMAC signature authentication
	AuthTypeHMAC = "hmac"
)

// AuthOptions configures the AuthMiddleware.
type AuthOptions struct {
	// AuthType specifies the authentication type (jwt or hmac)
	AuthType string
	
	// JWTSecret is the secret key for JWT token validation
	JWTSecret string
	
	// HMACSecret is the secret key for HMAC signature validation
	HMACSecret string
	
	// ErrorHandler is an optional custom error handler
	ErrorHandler func(w http.ResponseWriter, err error)
}

// AuthMiddleware manages JWT and HMAC authentication for protected routes.
type AuthMiddleware struct {
	options AuthOptions
}

// NewAuthMiddleware initializes an AuthMiddleware with specified options.
func NewAuthMiddleware(options AuthOptions) *AuthMiddleware {
	return &AuthMiddleware{
		options: options,
	}
}

// Middleware intercepts HTTP requests and validates authentication headers.
func (am *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error

		switch am.options.AuthType {
		case AuthTypeJWT:
			err = am.handleJWTAuth(w, r)
		case AuthTypeHMAC:
			err = am.handleHMACAuth(w, r)
		default:
			err = fmt.Errorf("unsupported authentication method: %s", am.options.AuthType)
		}

		if err != nil {
			am.handleError(w, err)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleError processes authentication errors.
func (am *AuthMiddleware) handleError(w http.ResponseWriter, err error) {
	if am.options.ErrorHandler != nil {
		am.options.ErrorHandler(w, err)
		return
	}

	// Default error handling
	log.Printf("Authentication error: %v", err)
	JSONResponse(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
}

// handleJWTAuth verifies JWT tokens and adds user claims to the request context.
func (am *AuthMiddleware) handleJWTAuth(w http.ResponseWriter, r *http.Request) error {
	tokenString, err := auth.ExtractBearerToken(r.Header.Get(AuthHeaderKey))
	if err != nil {
		return fmt.Errorf("missing or invalid authorization header: %w", err)
	}

	claims, err := auth.ValidateJWT(am.options.JWTSecret, tokenString)
	if err != nil {
		return fmt.Errorf("invalid JWT token: %w", err)
	}

	// Add JWT claims to the request context for downstream use
	ctx := context.WithValue(r.Context(), UserContextKey, claims)
	*r = *r.WithContext(ctx)
	return nil
}

// handleHMACAuth verifies HMAC signatures for request validation.
func (am *AuthMiddleware) handleHMACAuth(w http.ResponseWriter, r *http.Request) error {
	signature := r.Header.Get(HMACHeaderKey)
	if signature == "" {
		return fmt.Errorf("missing HMAC signature")
	}

	bodyBytes, err := getRequestBody(r)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	if !auth.ValidateHMAC(am.options.HMACSecret, string(bodyBytes), signature) {
		return fmt.Errorf("invalid HMAC signature")
	}
	return nil
}


// getRequestBody reads the request body for HMAC validation.
func getRequestBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, fmt.Errorf("request body is empty")
	}

	// Read the body and reset it so it can be read by other handlers
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return bodyBytes, nil
}

// GetUserFromContext retrieves JWT claims from the request context.
func GetUserFromContext(ctx context.Context) (jwt.MapClaims, bool) {
	claims, ok := ctx.Value(UserContextKey).(jwt.MapClaims)
	return claims, ok
}

// JSONResponse sends a JSON response with the specified status code and data.
func JSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
		}
	}
}
