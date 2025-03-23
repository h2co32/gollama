package examples

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/h2co32/gollama/pkg/auth"
	"github.com/h2co32/gollama/pkg/middleware"
)

// AuthJWTExample demonstrates generating and validating JWT tokens.
func AuthJWTExample() {
	// Create claims for the token
	claims := map[string]interface{}{
		"user_id": 123,
		"role":    "admin",
		"name":    "John Doe",
	}

	// Generate a JWT token with default options (1 hour expiration)
	secretKey := "your-jwt-secret-key"
	token, err := auth.GenerateJWT(secretKey, claims)
	if err != nil {
		log.Fatalf("Failed to generate JWT: %v", err)
	}

	fmt.Printf("Generated JWT token: %s\n\n", token)

	// Validate the token
	validatedClaims, err := auth.ValidateJWT(secretKey, token)
	if err != nil {
		log.Fatalf("Failed to validate JWT: %v", err)
	}

	fmt.Println("Validated JWT claims:")
	for key, value := range validatedClaims {
		fmt.Printf("  %s: %v\n", key, value)
	}
}

// AuthJWTWithOptionsExample demonstrates JWT token generation with custom options.
func AuthJWTWithOptionsExample() {
	// Create claims for the token
	claims := map[string]interface{}{
		"user_id": 456,
		"role":    "user",
	}

	// Create custom JWT options
	options := auth.JWTOptions{
		ExpiresIn: 15 * time.Minute,
		Issuer:    "gollama-auth-example",
		Audience:  "api.example.com",
	}

	// Generate a JWT token with custom options
	secretKey := "your-jwt-secret-key"
	token, err := auth.GenerateJWTWithOptions(secretKey, claims, options)
	if err != nil {
		log.Fatalf("Failed to generate JWT with options: %v", err)
	}

	fmt.Printf("Generated JWT token with custom options: %s\n", token)
}

// AuthHMACExample demonstrates generating and validating HMAC signatures.
func AuthHMACExample() {
	// Data to sign
	data := `{"user_id": 123, "action": "update_profile"}`
	secretKey := "your-hmac-secret-key"

	// Generate HMAC signature
	signature := auth.GenerateHMAC(secretKey, data)
	fmt.Printf("Generated HMAC signature: %s\n", signature)

	// Validate HMAC signature
	isValid := auth.ValidateHMAC(secretKey, data, signature)
	fmt.Printf("Signature is valid: %t\n", isValid)

	// Try with tampered data
	tamperedData := `{"user_id": 123, "action": "delete_account"}`
	isValid = auth.ValidateHMAC(secretKey, tamperedData, signature)
	fmt.Printf("Tampered data signature is valid: %t\n", isValid)
}

// AuthMiddlewareExample demonstrates using the auth middleware with an HTTP server.
func AuthMiddlewareExample() {
	// Create a JWT secret key
	jwtSecret := "your-jwt-secret-key"
	hmacSecret := "your-hmac-secret-key"

	// Create an auth middleware for JWT authentication
	jwtAuthMiddleware := middleware.NewAuthMiddleware(middleware.AuthOptions{
		AuthType:   middleware.AuthTypeJWT,
		JWTSecret:  jwtSecret,
		HMACSecret: hmacSecret,
	})

	// Create an auth middleware for HMAC authentication
	hmacAuthMiddleware := middleware.NewAuthMiddleware(middleware.AuthOptions{
		AuthType:   middleware.AuthTypeHMAC,
		JWTSecret:  jwtSecret,
		HMACSecret: hmacSecret,
	})

	// Create a protected handler that requires authentication
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user claims from context (for JWT auth)
		claims, ok := middleware.GetUserFromContext(r.Context())
		if ok {
			middleware.JSONResponse(w, http.StatusOK, map[string]interface{}{
				"message": "Protected resource accessed successfully",
				"user":    claims,
			})
		} else {
			middleware.JSONResponse(w, http.StatusOK, map[string]string{
				"message": "Protected resource accessed successfully (no user claims)",
			})
		}
	})

	// Register routes with different authentication methods
	http.Handle("/protected/jwt", jwtAuthMiddleware.Middleware(protectedHandler))
	http.Handle("/protected/hmac", hmacAuthMiddleware.Middleware(protectedHandler))

	// Create a public handler that doesn't require authentication
	http.HandleFunc("/public", func(w http.ResponseWriter, r *http.Request) {
		middleware.JSONResponse(w, http.StatusOK, map[string]string{
			"message": "Public resource accessed",
		})
	})

	// Start the server
	fmt.Println("Starting server on http://localhost:8080")
	fmt.Println("Routes:")
	fmt.Println("  - GET /public (no auth required)")
	fmt.Println("  - GET /protected/jwt (JWT auth required)")
	fmt.Println("  - POST /protected/hmac (HMAC auth required)")
	fmt.Println("\nPress Ctrl+C to stop the server")

	// In a real application, you would start the server here:
	// log.Fatal(http.ListenAndServe(":8080", nil))
	
	// For this example, we'll just show how to access the endpoints
	fmt.Println("\nExample curl commands:")
	
	// Generate a sample JWT token for the example
	token, _ := auth.GenerateJWT(jwtSecret, map[string]interface{}{"user_id": 123})
	fmt.Printf("curl -H 'Authorization: Bearer %s' http://localhost:8080/protected/jwt\n", token)
	
	// Generate a sample HMAC signature for the example
	data := `{"action":"test"}`
	signature := auth.GenerateHMAC(hmacSecret, data)
	fmt.Printf("curl -X POST -d '%s' -H 'X-Signature: %s' http://localhost:8080/protected/hmac\n", data, signature)
}
