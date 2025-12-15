// Package agent provides CLI-based agent execution using the claude CLI.
package agent

import (
	"fmt"
	"strings"
)

// EnhanceSDKError wraps CLI errors with helpful, actionable error messages.
// Kept for backward compatibility - now handles CLI execution errors.
func EnhanceSDKError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Check for CLI not found errors
	if strings.Contains(errStr, "claude CLI not found") || strings.Contains(errStr, "executable file not found") {
		return fmt.Errorf(`%w

To fix this:
  - Install Claude Code: https://claude.ai/code
  - Or set ANTHROPIC_API_KEY environment variable: export ANTHROPIC_API_KEY=your-key
  - Get your API key at https://console.anthropic.com/`, err)
	}

	// Check for authentication errors from CLI
	if strings.Contains(errStr, "authentication") || strings.Contains(errStr, "not authenticated") {
		return fmt.Errorf(`%w

To fix this:
  - Run 'claude login' to authenticate with Claude Code
  - Or set ANTHROPIC_API_KEY environment variable: export ANTHROPIC_API_KEY=your-key
  - Get your API key at https://console.anthropic.com/`, err)
	}

	// Check for timeout errors
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") {
		return fmt.Errorf(`request timeout: %w

This may indicate:
  - Network connectivity issues - check your internet connection
  - Request took too long - try again or increase timeout`, err)
	}

	// Return error as-is (CLI provides good error messages)
	return err
}
