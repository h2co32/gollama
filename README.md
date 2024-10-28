# **Gollama**

Gollama is a scalable, high-performance Golang application with built-in support for authentication, rate limiting, automatic scaling, tracing, and monitoring. Designed for production use, Gollama includes middleware for **JWT and HMAC authentication**, **OpenTelemetry tracing**, **Prometheus metrics**, **token bucket rate limiting**, **automatic worker scaling**, and **retry logic**.

## **Table of Contents**

- [Features](#features)
- [Requirements](#requirements)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Components](#components)
- [Endpoints](#endpoints)
- [Example Usage](#example-usage)

---

## **Features**

- **JWT and HMAC Authentication**: Middleware for secure access control.
- **Rate Limiting**: Token bucket rate limiter to control request rates.
- **Automatic Worker Scaling**: Autoscaler that adjusts worker counts based on system load.
- **Distributed Tracing**: Integrated with OpenTelemetry for end-to-end tracing.
- **Metrics and Monitoring**: Prometheus metrics collection for real-time monitoring.
- **Retry Logic**: Exponential backoff with jitter for handling transient failures.

## **Requirements**

- Go 1.17+
- [Prometheus](https://prometheus.io/) for metrics collection
- [OpenTelemetry](https://opentelemetry.io/) for tracing
- Redis (optional, if using distributed caching)
- TLS certificate and key files for secure connections (for production use)

## **Installation**

1. **Clone the Repository**:

   ```bash
   git clone https://github.com/yourusername/gollama.git
   cd gollama
   ```

2. **Install Dependencies**:

   Install required Go modules:

   ```bash
   go mod download
   ```

3. **Environment Setup**:

   Set up environment variables for authentication secrets, TLS paths, and other configurations (if needed).

---

## **Configuration**

Edit `constants.go` for default values, and customize the following parameters as needed:

- **Authentication**:
  - `JWTSecretKey`: The secret key used for JWT signing.
  - `HMACSecretKey`: The secret key used for HMAC signatures.

- **Rate Limiter**:
  - `DefaultRateLimitCapacity`: Max tokens allowed in the token bucket.
  - `DefaultRefillRate`: Rate at which tokens are added.

- **Autoscaler**:
  - `DefaultMinWorkers`: Minimum workers to keep active.
  - `DefaultMaxWorkers`: Maximum workers allowed.
  - `DefaultCPUThreshold`: CPU usage threshold to trigger scaling.

- **Retry**:
  - `DefaultMaxRetries`: Max retry attempts.
  - `DefaultInitialBackoff`: Initial backoff duration.
  - `DefaultMaxBackoff`: Maximum backoff duration.

---

## **Usage**

### **Starting the Server**

1. **Run the Server**:

   ```bash
   go run main.go
   ```

2. **Access API Endpoints**:

   The server runs on `http://localhost:8080` by default. Access endpoints like `/protected` and `/metrics` for testing.

---

## **Components**

### **1. Authentication Middleware**

- **File**: `auth_middleware.go`
- **Functionality**: Validates requests using **JWT** or **HMAC** authentication.
- **Usage**:
  - Set the `authType` as "jwt" or "hmac" when initializing `AuthMiddleware`.
  - Add the middleware to endpoints requiring authentication.

### **2. Rate Limiting**

- **File**: `rate_limiter.go`
- **Functionality**: Controls request rates using a token bucket algorithm.
- **Usage**:
  - Configure `DefaultRateLimitCapacity` and `DefaultRefillRate` in `constants.go`.
  - Call `Allow` or `Wait` on the `RateLimiter` instance to apply rate limiting.

### **3. Autoscaler**

- **File**: `autoscaler.go`
- **Functionality**: Automatically scales worker pools based on CPU usage.
- **Usage**:
  - Define min/max workers, CPU thresholds, and scaling intervals.
  - Use `Start()` and `Stop()` to manage the autoscaler.

### **4. Tracing and Observability**

- **Files**: `tracing.go`, `prometheus.go`
- **Functionality**: Provides **OpenTelemetry distributed tracing** and **Prometheus metrics**.
- **Usage**:
  - Configure an OTLP endpoint for tracing.
  - Access Prometheus metrics at `/metrics` (default port 2112).

### **5. Retry Logic**

- **File**: `retry.go`
- **Functionality**: Implements exponential backoff with optional jitter.
- **Usage**:
  - Pass an operation to `Retry()` with configuration for retries and backoff duration.

### **6. Utility Functions and Constants**

- **Files**: `constants.go`, `helpers.go`
- **Functionality**: Provides shared constants and helper functions for logging, JSON handling, and error management.
- **Usage**: Call helper functions for logging, JSON responses, etc., throughout the application.

---

## **Endpoints**

### **Protected Endpoints**

- **`/protected`** (requires JWT or HMAC)
  - **Method**: GET
  - **Description**: Protected route that requires JWT or HMAC authentication.

### **Monitoring and Metrics**

- **`/metrics`**
  - **Method**: GET
  - **Description**: Exposes Prometheus metrics for real-time monitoring.

- **`/health`**
  - **Method**: GET
  - **Description**: Health check endpoint to verify service status.

---

## **Example Usage**

### **JWT Authentication Example**

1. **Generate JWT**:

   ```go
   claims := jwt.MapClaims{"username": "user1", "exp": time.Now().Add(30 * time.Minute).Unix()}
   token, err := security.GenerateJWT("supersecretkey", claims)
   fmt.Println("JWT:", token)
   ```

2. **Access Protected Endpoint**:

   ```http
   GET /protected HTTP/1.1
   Authorization: Bearer <jwt_token>
   ```

### **Rate Limiting Example**

```go
limiter := rate_limiter.NewRateLimiter(5, time.Second, 1)
for i := 0; i < 10; i++ {
    if limiter.Allow() {
        fmt.Println("Request allowed")
    } else {
        fmt.Println("Rate limit exceeded, waiting for token")
        limiter.Wait(2 * time.Second)
    }
}
```

### **Autoscaling Example**

```go
scaler := autoscaler.NewAutoScaler(2, 10, 0.75, 2*time.Second, 2*time.Second)
scaler.Start()
time.Sleep(20 * time.Second)
scaler.Stop()
```

### **Tracing Example**

1. **Initialize Tracer**:

   ```go
   tracerProvider, _ := observability.InitTracer("gollama-service", "localhost:4318")
   defer tracerProvider.ShutdownTracer(context.Background())
   ```

2. **Start a Span**:

   ```go
   ctx, span := tracerProvider.StartSpan(context.Background(), "example-operation")
   defer tracerProvider.EndSpan(span, nil)
   ```

---

## **Contributing**

1. **Fork the repository** and create a new branch for your feature or bugfix.
2. **Make your changes**, write tests, and ensure everything passes.
3. **Submit a pull request** with a detailed description of your changes.

---

## **License**

This project is licensed under the MIT License.

---

## **Contact**

For questions, issues, or suggestions, please contact [your_email@example.com](mailto:your_email@example.com).

---

This README provides a thorough guide to using and extending Gollama, enabling seamless development, secure API access, and enhanced performance monitoring.

