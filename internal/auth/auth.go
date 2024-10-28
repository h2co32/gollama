package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// AuthType represents different authentication methods
type AuthType int

const (
	JWTAuth AuthType = iota
	HMACAuth
	OAuthAuth
)

// AuthConfig holds configuration details for each auth type
type AuthConfig struct {
	AuthType AuthType
	// JWT configuration
	JWTToken     string
	JWTExpiresAt time.Time
	// HMAC configuration
	HMACKey string
	// OAuth configuration
	OAuthTokenURL  string
	ClientID       string
	ClientSecret   string
	OAuthToken     string
	OAuthExpiresAt time.Time
}

// AuthMiddleware provides authentication functionality for HTTP requests
type AuthMiddleware struct {
	config AuthConfig
}

// NewAuthMiddleware initializes a new AuthMiddleware with the given configuration
func NewAuthMiddleware(config AuthConfig) *AuthMiddleware {
	return &AuthMiddleware{config: config}
}

// ProcessRequest adds the appropriate authentication header based on AuthType
func (a *AuthMiddleware) ProcessRequest(req *http.Request) (*http.Request, error) {
	switch a.config.AuthType {
	case JWTAuth:
		return a.addJWTAuth(req)
	case HMACAuth:
		return a.addHMACAuth(req)
	case OAuthAuth:
		return a.addOAuthAuth(req)
	default:
		return req, fmt.Errorf("unsupported auth type: %v", a.config.AuthType)
	}
}

// addJWTAuth adds a JWT token to the request
func (a *AuthMiddleware) addJWTAuth(req *http.Request) (*http.Request, error) {
	if time.Now().After(a.config.JWTExpiresAt) {
		return req, fmt.Errorf("JWT token expired")
	}
	req.Header.Set("Authorization", "Bearer "+a.config.JWTToken)
	return req, nil
}

// addHMACAuth adds an HMAC signature to the request for authentication
func (a *AuthMiddleware) addHMACAuth(req *http.Request) (*http.Request, error) {
	if a.config.HMACKey == "" {
		return req, fmt.Errorf("HMAC key is missing")
	}

	// Create HMAC hash of the request URL
	mac := hmac.New(sha256.New, []byte(a.config.HMACKey))
	mac.Write([]byte(req.URL.String()))
	signature := hex.EncodeToString(mac.Sum(nil))

	// Add signature to headers
	req.Header.Set("X-Signature", signature)
	return req, nil
}

// addOAuthAuth checks the OAuth token validity and adds it to the request
func (a *AuthMiddleware) addOAuthAuth(req *http.Request) (*http.Request, error) {
	// Refresh the token if expired
	if time.Now().After(a.config.OAuthExpiresAt) {
		if err := a.refreshOAuthToken(); err != nil {
			return req, fmt.Errorf("failed to refresh OAuth token: %w", err)
		}
	}

	req.Header.Set("Authorization", "Bearer "+a.config.OAuthToken)
	return req, nil
}

// refreshOAuthToken fetches a new OAuth token and updates the config
func (a *AuthMiddleware) refreshOAuthToken() error {
	// Prepare the request to fetch the OAuth token
	req, err := http.NewRequest("POST", a.config.OAuthTokenURL, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(a.config.ClientID, a.config.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Body = ioutil.NopCloser(strings.NewReader("grant_type=client_credentials"))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get OAuth token, status: %s", res.Status)
	}

	// Parse the response for the token and expiration
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(res.Body).Decode(&tokenResponse); err != nil {
		return err
	}

	// Update the config with the new token and expiration
	a.config.OAuthToken = tokenResponse.AccessToken
	a.config.OAuthExpiresAt = time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)

	return nil
}

// Utility functions for testing and error handling
func GenerateHMACSignature(data, key string) (string, error) {
	mac := hmac.New(sha256.New, []byte(key))
	if _, err := mac.Write([]byte(data)); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}
