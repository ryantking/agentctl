// Package agent provides Anthropic SDK client initialization and configuration.
package agent

import (
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// NewClient creates a new Anthropic SDK client.
// Reads ANTHROPIC_API_KEY from environment variable.
// Returns error if API key is not configured.
func NewClient() (anthropic.Client, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return anthropic.Client{}, fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}

	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return client, nil
}

// NewClientOrNil creates a new Anthropic SDK client, or returns zero value if API key is not configured.
// This allows graceful fallback when SDK features are optional.
func NewClientOrNil() (anthropic.Client, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return anthropic.Client{}, nil // Graceful fallback - no error
	}

	return NewClient()
}

// IsConfigured checks if ANTHROPIC_API_KEY is set.
func IsConfigured() bool {
	return os.Getenv("ANTHROPIC_API_KEY") != ""
}
