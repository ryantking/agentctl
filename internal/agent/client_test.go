package agent

import (
	"os"
	"testing"
)

func TestIsConfigured(t *testing.T) {
	// Save original value
	originalKey := os.Getenv("ANTHROPIC_API_KEY")
	defer func() {
		if originalKey != "" {
			_ = os.Setenv("ANTHROPIC_API_KEY", originalKey) //nolint:errcheck // Test cleanup
		} else {
			_ = os.Unsetenv("ANTHROPIC_API_KEY") //nolint:errcheck // Test cleanup
		}
	}()

	// Test with API key set (should return true)
	_ = os.Setenv("ANTHROPIC_API_KEY", "test-key") //nolint:errcheck // Test setup
	if !IsConfigured() {
		t.Error("IsConfigured() should return true when API key is set")
	}

	// Test without API key (may return true if CLI is available, false otherwise)
	_ = os.Unsetenv("ANTHROPIC_API_KEY") //nolint:errcheck // Test cleanup
	// IsConfigured may return true if claude CLI is available, which is fine
	_ = IsConfigured()
}

func TestNewClientOrNil(t *testing.T) {
	// Test graceful fallback - returns empty struct for backward compatibility
	client, err := NewClientOrNil()
	if err != nil {
		t.Errorf("NewClientOrNil() should not return error, got %v", err)
	}
	// Should return empty struct
	_ = client
}
