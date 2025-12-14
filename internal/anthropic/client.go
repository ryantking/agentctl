// Package anthropic provides Anthropic SDK client initialization and configuration.
package anthropic

import (
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
)

// NewClient creates a new Anthropic SDK client.
// Reads ANTHROPIC_API_KEY from environment variable.
// Returns error if API key is not configured.
func NewClient() (*anthropic.Client, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}

	client, err := anthropic.NewClient(anthropic.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Anthropic client: %w", err)
	}

	return client, nil
}

// NewClientOrNil creates a new Anthropic SDK client, or returns nil if API key is not configured.
// This allows graceful fallback when SDK features are optional.
func NewClientOrNil() (*anthropic.Client, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, nil // Graceful fallback - no error
	}

	return NewClient()
}

// IsConfigured checks if ANTHROPIC_API_KEY is set.
func IsConfigured() bool {
	return os.Getenv("ANTHROPIC_API_KEY") != ""
}
