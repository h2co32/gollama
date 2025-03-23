package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/h2co32/gollama/internal/security"
	"github.com/h2co32/gollama/internal/utils"

	"github.com/golang-jwt/jwt/v4"
)

// Context keys for adding authentication data to the request context
type contextKey string

const (
	UserContextKey contextKey = "user"
	AuthHeaderKey  string     = "Authorization"
	HMACHeaderKey  string     = "X-Signature"
)

// AuthMiddleware manages JWT and HMAC authentication for protected routes
type AuthMiddleware struct {
	authType   string
	jwtSecret  string
	hmacSecret string
}

// NewAuthMiddleware initializes an AuthMiddleware with specified authentication type and keys
func NewAuthMiddleware(authType, jwtSecret, hmacSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		authType:   authType,
		jwtSecret:  jwtSecret,
		hmacSecret: hmacSecret,
	}
}

// Middleware intercepts HTTP requests and validates authentication headers
func (am *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error

		if am.authType == "jwt" {
			err = am.handleJWTAuth(w, r)
		} else if am.authType == "hmac" {
			err = am.handleHMACAuth(w, r)
		} else {
			utils.JSONResponse(w, http.StatusUnauthorized, map[string]string{"error": "unsupported authentication method"})
			return
		}

		if err != nil {
			utils.LogError("AuthMiddleware", err)
			utils.JSONResponse(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleJWTAuth verifies JWT tokens and adds user claims to the request context
func (am *AuthMiddleware) handleJWTAuth(w http.ResponseWriter, r *http.Request) error {
	tokenString := extractBearerToken(r.Header.Get(AuthHeaderKey))
	if tokenString == "" {
		return fmt.Errorf("missing or invalid authorization header")
	}

	claims, err := security.ValidateJWT(am.jwtSecret, tokenString)
	if err != nil {
		return fmt.Errorf("invalid JWT token: %w", err)
	}

	// Add JWT claims to the request context for downstream use
	ctx := context.WithValue(r.Context(), UserContextKey, claims)
	*r = *r.WithContext(ctx)
	return nil
}

// handleHMACAuth verifies HMAC signatures for request validation
func (am *AuthMiddleware) handleHMACAuth(w http.ResponseWriter, r *http.Request) error {
	signature := r.Header.Get(HMACHeaderKey)
	if signature == "" {
		return fmt.Errorf("missing HMAC signature")
	}

	bodyBytes, err := getRequestBody(r)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	if !security.ValidateHMAC(am.hmacSecret, string(bodyBytes), signature) {
		return fmt.Errorf("invalid HMAC signature")
	}
	return nil
}

// extractBearerToken extracts the JWT token from the Authorization header
func extractBearerToken(authHeader string) string {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(authHeader, "Bearer ")
}

// getRequestBody reads the request body for HMAC validation
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

// GetUserFromContext retrieves JWT claims from the request context
func GetUserFromContext(ctx context.Context) (jwt.MapClaims, bool) {
	claims, ok := ctx.Value(UserContextKey).(jwt.MapClaims)
	return claims, ok
}
