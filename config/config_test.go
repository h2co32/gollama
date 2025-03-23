package config

import (
	"testing"
	"time"
)

func TestDefaultProfile(t *testing.T) {
	// Test that DefaultProfile has expected values
	if DefaultProfile.MaxRetries != 3 {
		t.Errorf("Expected DefaultProfile.MaxRetries to be 3, got %d", DefaultProfile.MaxRetries)
	}
	
	expectedTimeout := 5 * time.Second
	if DefaultProfile.Timeout != expectedTimeout {
		t.Errorf("Expected DefaultProfile.Timeout to be %v, got %v", expectedTimeout, DefaultProfile.Timeout)
	}
	
	if DefaultProfile.RateLimit != 10 {
		t.Errorf("Expected DefaultProfile.RateLimit to be 10, got %d", DefaultProfile.RateLimit)
	}
	
	// Test model settings
	expectedTemp := 0.7
	if temp, ok := DefaultProfile.ModelSettings["temperature"].(float64); !ok || temp != expectedTemp {
		t.Errorf("Expected DefaultProfile.ModelSettings[\"temperature\"] to be %v, got %v", expectedTemp, DefaultProfile.ModelSettings["temperature"])
	}
	
	expectedMaxTokens := 1024
	if maxTokens, ok := DefaultProfile.ModelSettings["max_tokens"].(int); !ok || maxTokens != expectedMaxTokens {
		t.Errorf("Expected DefaultProfile.ModelSettings[\"max_tokens\"] to be %v, got %v", expectedMaxTokens, DefaultProfile.ModelSettings["max_tokens"])
	}
}

func TestProductionProfile(t *testing.T) {
	// Test that ProductionProfile has expected values
	if ProductionProfile.MaxRetries != 5 {
		t.Errorf("Expected ProductionProfile.MaxRetries to be 5, got %d", ProductionProfile.MaxRetries)
	}
	
	expectedTimeout := 10 * time.Second
	if ProductionProfile.Timeout != expectedTimeout {
		t.Errorf("Expected ProductionProfile.Timeout to be %v, got %v", expectedTimeout, ProductionProfile.Timeout)
	}
	
	if ProductionProfile.RateLimit != 100 {
		t.Errorf("Expected ProductionProfile.RateLimit to be 100, got %d", ProductionProfile.RateLimit)
	}
	
	// Test model settings
	expectedTemp := 0.5
	if temp, ok := ProductionProfile.ModelSettings["temperature"].(float64); !ok || temp != expectedTemp {
		t.Errorf("Expected ProductionProfile.ModelSettings[\"temperature\"] to be %v, got %v", expectedTemp, ProductionProfile.ModelSettings["temperature"])
	}
	
	expectedMaxTokens := 2048
	if maxTokens, ok := ProductionProfile.ModelSettings["max_tokens"].(int); !ok || maxTokens != expectedMaxTokens {
		t.Errorf("Expected ProductionProfile.ModelSettings[\"max_tokens\"] to be %v, got %v", expectedMaxTokens, ProductionProfile.ModelSettings["max_tokens"])
	}
}

func TestConfigProfileCustomization(t *testing.T) {
	// Test creating a custom profile
	customProfile := ConfigProfile{
		MaxRetries: 10,
		Timeout:    30 * time.Second,
		RateLimit:  50,
		ModelSettings: map[string]interface{}{
			"temperature": 0.8,
			"max_tokens":  4096,
			"top_p":       0.95,
		},
	}
	
	// Verify custom values
	if customProfile.MaxRetries != 10 {
		t.Errorf("Expected customProfile.MaxRetries to be 10, got %d", customProfile.MaxRetries)
	}
	
	expectedTimeout := 30 * time.Second
	if customProfile.Timeout != expectedTimeout {
		t.Errorf("Expected customProfile.Timeout to be %v, got %v", expectedTimeout, customProfile.Timeout)
	}
	
	if customProfile.RateLimit != 50 {
		t.Errorf("Expected customProfile.RateLimit to be 50, got %d", customProfile.RateLimit)
	}
	
	// Test custom model settings
	expectedTemp := 0.8
	if temp, ok := customProfile.ModelSettings["temperature"].(float64); !ok || temp != expectedTemp {
		t.Errorf("Expected customProfile.ModelSettings[\"temperature\"] to be %v, got %v", expectedTemp, customProfile.ModelSettings["temperature"])
	}
	
	expectedMaxTokens := 4096
	if maxTokens, ok := customProfile.ModelSettings["max_tokens"].(int); !ok || maxTokens != expectedMaxTokens {
		t.Errorf("Expected customProfile.ModelSettings[\"max_tokens\"] to be %v, got %v", expectedMaxTokens, customProfile.ModelSettings["max_tokens"])
	}
	
	expectedTopP := 0.95
	if topP, ok := customProfile.ModelSettings["top_p"].(float64); !ok || topP != expectedTopP {
		t.Errorf("Expected customProfile.ModelSettings[\"top_p\"] to be %v, got %v", expectedTopP, customProfile.ModelSettings["top_p"])
	}
}
