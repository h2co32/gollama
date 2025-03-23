package observability

import (
	"testing"
)

// TestDefaultTracerOptions tests the DefaultTracerOptions function
func TestDefaultTracerOptions(t *testing.T) {
	options := DefaultTracerOptions()

	if options.SamplingRatio != 1.0 {
		t.Errorf("Expected SamplingRatio to be 1.0, got %f", options.SamplingRatio)
	}

	if options.ServiceVersion != "unknown" {
		t.Errorf("Expected ServiceVersion to be 'unknown', got '%s'", options.ServiceVersion)
	}

	if options.ServiceNamespace != "" {
		t.Errorf("Expected ServiceNamespace to be empty, got '%s'", options.ServiceNamespace)
	}

	if len(options.AdditionalAttributes) != 0 {
		t.Errorf("Expected AdditionalAttributes to be empty, got %d attributes", len(options.AdditionalAttributes))
	}
}

// TestVersion tests that the Version constant is set
func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Expected Version to be set, got empty string")
	}

	// Version should follow semantic versioning (x.y.z)
	if len(Version) < 5 { // At minimum "1.0.0" is 5 characters
		t.Errorf("Expected Version to follow semantic versioning (x.y.z), got %s", Version)
	}
}
