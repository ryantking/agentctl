package anthropic

import (
	"os"
	"testing"
)

func TestIsConfigured(t *testing.T) {
	// Save original value
	originalKey := os.Getenv("ANTHROPIC_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("ANTHROPIC_API_KEY", originalKey)
		} else {
			os.Unsetenv("ANTHROPIC_API_KEY")
		}
	}()

	// Test unset
	_ = os.Unsetenv("ANTHROPIC_API_KEY") //nolint:errcheck // Test cleanup
	if IsConfigured() {
		t.Error("IsConfigured() should return false when API key is not set")
	}

	// Test set
	os.Setenv("ANTHROPIC_API_KEY", "test-key")
	if !IsConfigured() {
		t.Error("IsConfigured() should return true when API key is set")
	}
}

func TestNewClientOrNil(t *testing.T) {
	// Save original value
	originalKey := os.Getenv("ANTHROPIC_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("ANTHROPIC_API_KEY", originalKey)
		} else {
			_ = os.Unsetenv("ANTHROPIC_API_KEY") //nolint:errcheck // Test cleanup
		}
	}()

	// Test graceful fallback when not configured
	_ = os.Unsetenv("ANTHROPIC_API_KEY") //nolint:errcheck // Test cleanup
	client, err := NewClientOrNil()
	if err != nil {
		t.Errorf("NewClientOrNil() should not return error when API key not set, got %v", err)
	}
	// Check if client has no options (zero value)
	if len(client.Options) != 0 {
		t.Error("NewClientOrNil() should return client with no options when API key not set")
	}
}
