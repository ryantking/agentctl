// Package anthropic provides Anthropic SDK client initialization and configuration.
package anthropic

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/shared"
)

// EnhanceSDKError wraps SDK errors with helpful, actionable error messages.
func EnhanceSDKError(err error) error {
	if err == nil {
		return nil
	}

	// Check for authentication errors
	var authErr *shared.AuthenticationError
	if errors.As(err, &authErr) {
		return fmt.Errorf(`authentication failed: %w

To fix this:
  - Set ANTHROPIC_API_KEY environment variable: export ANTHROPIC_API_KEY=your-key
  - Or run 'claude login' if you have the Claude CLI installed
  - Verify your API key is valid at https://console.anthropic.com/`, err)
	}

	// Check for API errors (rate limits, etc.)
	var apiErr *shared.APIErrorObject
	if errors.As(err, &apiErr) {
		// Check for rate limit errors
		if apiErr.Status == http.StatusTooManyRequests {
			return fmt.Errorf(`rate limit exceeded: %w

The API rate limit has been reached. Please:
  - Wait a few moments and try again
  - Check your usage at https://console.anthropic.com/
  - Consider upgrading your plan if you frequently hit limits`, err)
		}

		// Check for other API errors
		if apiErr.Status == http.StatusUnauthorized || apiErr.Status == http.StatusForbidden {
			return fmt.Errorf(`authorization failed (HTTP %d): %w

To fix this:
  - Verify your API key is correct: echo $ANTHROPIC_API_KEY
  - Check your API key permissions at https://console.anthropic.com/
  - Run 'claude status' if you have the Claude CLI installed`, apiErr.Status, err)
		}

		// Generic API error with status code
		return fmt.Errorf("API error (HTTP %d): %w\n\nCheck https://console.anthropic.com/ for account status and usage limits", apiErr.Status, err)
	}

	// Check for network errors
	errStr := err.Error()
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") {
		return fmt.Errorf(`request timeout: %w

This may indicate:
  - Network connectivity issues - check your internet connection
  - API service temporarily unavailable - try again in a moment
  - Request took too long - the operation may have timed out`, err)
	}

	if strings.Contains(errStr, "connection") || strings.Contains(errStr, "network") || strings.Contains(errStr, "dial") {
		return fmt.Errorf(`network error: %w

This may indicate:
  - No internet connection - check your network connectivity
  - Firewall blocking requests - check your firewall settings
  - DNS resolution issues - verify you can reach api.anthropic.com`, err)
	}

	// Check for missing API key (before SDK call)
	if strings.Contains(errStr, "ANTHROPIC_API_KEY") || strings.Contains(errStr, "not set") || strings.Contains(errStr, "not configured") {
		return fmt.Errorf(`%w

To fix this:
  - Set ANTHROPIC_API_KEY environment variable: export ANTHROPIC_API_KEY=your-key
  - Or run 'claude login' if you have the Claude CLI installed
  - Get your API key at https://console.anthropic.com/`, err)
	}

	// Generic error - wrap with context
	return fmt.Errorf("Anthropic API error: %w\n\nFor help, check https://docs.anthropic.com/ or verify your API key at https://console.anthropic.com/", err)
}
