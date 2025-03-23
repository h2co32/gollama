# Versioning Strategy for Gollama

This document outlines the versioning strategy for the Gollama library. We follow [Semantic Versioning 2.0.0](https://semver.org/) (SemVer) to ensure clear communication about changes and compatibility.

## Semantic Versioning

Gollama uses a version number format of `MAJOR.MINOR.PATCH` where:

- **MAJOR** version increases when incompatible API changes are made
- **MINOR** version increases when functionality is added in a backward-compatible manner
- **PATCH** version increases when backward-compatible bug fixes are implemented

## Package Versioning

Each package in the `pkg/` directory maintains its own version constant that follows the same semantic versioning principles. This allows individual packages to evolve at their own pace while maintaining compatibility guarantees.

```go
// Version represents the current package version following semantic versioning.
const Version = "1.0.0"
```

## API Stability Guarantees

### Stable APIs (v1.0.0 and above)

Packages with version 1.0.0 or higher provide the following stability guarantees:

- **Backward Compatibility**: APIs will remain backward compatible within the same major version.
- **Deprecation Policy**: APIs will be marked as deprecated for at least one minor release before removal in a major version.
- **Error Handling**: Error types and behavior will remain consistent within the same major version.

### Pre-release APIs (v0.x.x)

Packages with versions below 1.0.0 are considered pre-release and may undergo significant changes:

- **No Stability Guarantee**: APIs may change without notice.
- **Experimental Features**: These packages may contain experimental features that could be removed or significantly altered.
- **Limited Support**: While we strive for quality, these packages may have incomplete documentation or test coverage.

## Release Cycle

- **Regular Releases**: We aim to release new versions on a regular schedule, typically every 4-8 weeks.
- **Security Patches**: Critical security fixes will be released as soon as possible.
- **Long-term Support (LTS)**: Major versions will be supported with security patches for at least 12 months after the next major version is released.

## Upgrading Guide

When upgrading between versions, please refer to the following guidelines:

- **Patch Upgrades** (e.g., 1.0.0 to 1.0.1): Safe to upgrade with no code changes required.
- **Minor Upgrades** (e.g., 1.0.0 to 1.1.0): Generally safe to upgrade. New features may be available, but existing code should continue to work.
- **Major Upgrades** (e.g., 1.0.0 to 2.0.0): May require code changes. Refer to the migration guide in the release notes.

## Deprecation Process

1. **Marking as Deprecated**: Functions, methods, or types that are planned for removal will be marked with a deprecation notice in the documentation and code.
2. **Deprecation Period**: Deprecated APIs will remain functional for at least one minor release cycle.
3. **Removal**: Deprecated APIs will only be removed in a major version release.

## Version Management in Go Modules

Gollama follows Go module versioning best practices:

- Tagged releases follow the `v{MAJOR}.{MINOR}.{PATCH}` format (e.g., `v1.2.3`).
- Major version changes (v2 and beyond) will use the `/v{MAJOR}` import path suffix as recommended by Go modules.

## Checking Version Information

You can check the version of each package programmatically:

```go
import "github.com/h2co32/gollama/pkg/retry"

func main() {
    fmt.Println("Retry package version:", retry.Version)
}
```

## Changelog

All notable changes to this project will be documented in the [CHANGELOG.md](CHANGELOG.md) file, organized by version and categorized as:

- **Added**: New features
- **Changed**: Changes in existing functionality
- **Deprecated**: Features that will be removed in upcoming releases
- **Removed**: Features removed in this release
- **Fixed**: Bug fixes
- **Security**: Security improvements or vulnerability fixes
