# **Gollama**

Gollama is a comprehensive Go library that provides essential components for building scalable, high-performance applications. It offers modular packages for authentication, rate limiting, observability, and more, designed to be used independently or together.

## **Table of Contents**

- [Overview](#overview)
- [Features](#features)
- [Installation](#installation)
- [Library Packages](#library-packages)
- [Examples](#examples)
- [Versioning](#versioning)
- [Releases](#releases)
- [Contributing](#contributing)
- [License](#license)

---

## **Overview**

Gollama started as an application but has been refactored into a modular library to make its powerful components available for use in any Go project. Each package follows Go best practices, is well-documented, and includes comprehensive examples.

---

## **Features**

- **Modular Design**: Use only the components you need
- **Well-Documented API**: Comprehensive documentation and examples
- **Production-Ready**: Designed for reliability and performance
- **Semantic Versioning**: Clear compatibility guarantees
- **Minimal Dependencies**: Each package has only the dependencies it needs

### **Core Functionality**

- **Authentication**: JWT and HMAC authentication utilities
- **Rate Limiting**: Token bucket rate limiter to control request rates
- **Observability**: OpenTelemetry integration for distributed tracing
- **Middleware**: HTTP middleware for authentication and other cross-cutting concerns
- **Retry Logic**: Exponential backoff with jitter for handling transient failures

---

## **Installation**

To use Gollama in your project, install it using Go modules:

```bash
go get github.com/h2co32/gollama
```

Or add it to your `go.mod` file:

```
require github.com/h2co32/gollama v1.0.0
```

Then import only the packages you need:

```go
import (
    "github.com/h2co32/gollama/pkg/auth"
    "github.com/h2co32/gollama/pkg/retry"
    // Import other packages as needed
)
```

---

## **Library Packages**

### **pkg/auth**

Authentication utilities for JWT and HMAC authentication.

```go
// Generate a JWT token
claims := map[string]interface{}{"user_id": 123, "role": "admin"}
token, err := auth.GenerateJWT("your-secret-key", claims)

// Validate a JWT token
claims, err := auth.ValidateJWT("your-secret-key", token)

// Generate an HMAC signature
signature := auth.GenerateHMAC("your-hmac-key", "data-to-sign")
```

### **pkg/middleware**

HTTP middleware components for authentication and other cross-cutting concerns.

```go
// Create a new auth middleware for JWT authentication
authMiddleware := middleware.NewAuthMiddleware(middleware.AuthOptions{
    AuthType:   middleware.AuthTypeJWT,
    JWTSecret:  "your-jwt-secret",
    HMACSecret: "your-hmac-secret",
})

// Use the middleware with an HTTP handler
http.Handle("/protected", authMiddleware.Middleware(http.HandlerFunc(protectedHandler)))
```

### **pkg/ratelimiter**

Token bucket rate limiter for controlling request rates.

```go
// Create a rate limiter with 10 tokens per second and a burst capacity of 20
limiter := ratelimiter.New(10, time.Second, 20)

// Check if an operation is allowed
if limiter.Allow() {
    // Perform the operation
} else {
    // Operation not allowed, handle accordingly
}

// Or wait until an operation is allowed (with timeout)
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
if err := limiter.Wait(ctx); err == nil {
    // Perform the operation
}
```

### **pkg/retry**

Flexible retry mechanism with exponential backoff and jitter.

```go
// Configure retry options
opts := retry.Options{
    MaxAttempts:    5,
    InitialBackoff: 100 * time.Millisecond,
    MaxBackoff:     10 * time.Second,
    Jitter:         true,
}

// Execute an operation with retry
err := retry.Do(opts, func() error {
    return makeNetworkRequest()
})
```

### **pkg/observability**

Tools for distributed tracing with OpenTelemetry.

```go
// Initialize a tracer provider
tp, err := observability.NewTracerProvider("my-service", "http://localhost:4318")
if err != nil {
    log.Fatalf("Failed to initialize tracer: %v", err)
}
defer tp.Shutdown(context.Background())

// Create a span
ctx, span := tp.StartSpan(context.Background(), "my-operation")
defer span.End()

// Add attributes to the span
observability.AddSpanAttributes(ctx, attribute.String("key", "value"))
```

---

## **Examples**

Each package includes comprehensive examples in the `pkg/examples` directory:

- **Authentication**: JWT and HMAC authentication examples
- **Middleware**: HTTP middleware usage examples
- **Rate Limiting**: Rate limiter usage in various scenarios
- **Retry Logic**: Retry patterns for different use cases
- **Observability**: Tracing examples for HTTP requests and error handling

To run an example:

```go
import "github.com/h2co32/gollama/pkg/examples"

func main() {
    examples.RetryBasicExample()
    examples.AuthJWTExample()
    // Run other examples as needed
}
```

---

## **Versioning**

Gollama follows [Semantic Versioning](https://semver.org/). See [VERSIONING.md](VERSIONING.md) for details on our versioning strategy and API stability guarantees.

For a list of changes in each release, see the [CHANGELOG.md](CHANGELOG.md).

## **Releases**

Gollama uses GitHub Actions for continuous integration and automated releases. Pre-built binaries for multiple platforms are available on the [Releases](https://github.com/h2co32/gollama/releases) page.

### **Installing from Releases**

1. Download the appropriate binary for your platform from the [Releases](https://github.com/h2co32/gollama/releases) page
2. Extract the archive (if applicable)
3. Move the binary to a location in your PATH

```bash
# Example for Linux/macOS
wget https://github.com/h2co32/gollama/releases/download/v1.0.0/gollama_Linux_x86_64.tar.gz
tar -xzf gollama_Linux_x86_64.tar.gz
sudo mv gollama /usr/local/bin/
```

### **Creating a Release**

For maintainers who want to create a new release:

1. Update the CHANGELOG.md with the changes in the new version
2. Tag the commit with the new version: `git tag v1.0.0`
3. Push the tag to GitHub: `git push origin v1.0.0`
4. The GitHub Actions workflow will automatically build and publish the release

---

## **Contributing**

Contributions are welcome! Please feel free to submit a Pull Request.

1. **Fork the repository** and create a new branch for your feature or bugfix.
2. **Make your changes**, write tests, and ensure everything passes.
3. **Submit a pull request** with a detailed description of your changes.

---

## **License**

This project is licensed under the MIT License.

---

This README provides an overview of the Gollama library. For detailed documentation on each package, refer to the package-level documentation and examples.
