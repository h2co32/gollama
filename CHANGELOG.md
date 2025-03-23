# Changelog

All notable changes to the Gollama project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial library structure with modular packages
- Comprehensive documentation for all public APIs
- Examples for each package showing common use cases

## [1.0.0] - 2025-03-23

### Added
- Public API in `pkg/` directory with the following packages:
  - `pkg/retry`: Flexible retry mechanism with exponential backoff and jitter
  - `pkg/ratelimiter`: Token bucket rate limiter for controlling request rates
  - `pkg/auth`: Authentication utilities for JWT and HMAC authentication
  - `pkg/middleware`: HTTP middleware components for authentication
  - `pkg/observability`: Tools for distributed tracing with OpenTelemetry

### Changed
- Restructured project to follow Go library best practices
- Improved documentation with examples for all packages
- Reduced coupling between packages for better modularity

### Removed
- Removed direct dependencies between public packages

## [0.9.0] - 2025-03-01

### Added
- Initial implementation of Gollama as an application
- JWT and HMAC authentication
- Rate limiting with token bucket algorithm
- Automatic worker scaling
- Distributed tracing with OpenTelemetry
- Prometheus metrics collection
- Retry logic with exponential backoff
