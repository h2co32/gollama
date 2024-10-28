package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// TLSConfig creates a TLS configuration for secure connections
func TLSConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificates: %w", err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// GenerateJWT creates a signed JWT token with custom claims
func GenerateJWT(secretKey string, claims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}
	return tokenString, nil
}

// ValidateJWT validates a JWT token and returns its claims
func ValidateJWT(secretKey, tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token claims")
}

// GenerateHMAC generates an HMAC signature for the data using a secret key
func GenerateHMAC(secretKey, data string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// ValidateHMAC validates an HMAC signature for the given data and secret key
func ValidateHMAC(secretKey, data, signature string) bool {
	expectedMAC := GenerateHMAC(secretKey, data)
	return hmac.Equal([]byte(expectedMAC), []byte(signature))
}

// SecureRequest adds a JWT or HMAC header for securing API requests
func SecureRequest(req *http.Request, authType, key, data string) error {
	if authType == "jwt" {
		claims := jwt.MapClaims{
			"sub":   "user",
			"exp":   time.Now().Add(1 * time.Hour).Unix(),
			"iat":   time.Now().Unix(),
			"scope": "read",
		}
		token, err := GenerateJWT(key, claims)
		if err != nil {
			return fmt.Errorf("failed to generate JWT: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	} else if authType == "hmac" {
		signature := GenerateHMAC(key, data)
		req.Header.Set("X-Signature", signature)
	} else {
		return fmt.Errorf("unsupported auth type: %s", authType)
	}
	return nil
}
