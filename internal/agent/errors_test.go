package agent

import (
	"errors"
	"strings"
	"testing"
)

func TestEnhanceSDKError_AuthenticationError(t *testing.T) {
	err := errors.New("authentication_error: invalid api key")

	enhanced := EnhanceSDKError(err)
	if enhanced == nil {
		t.Fatal("EnhanceSDKError should not return nil")
	}

	errStr := enhanced.Error()
	if !contains(errStr, "authentication failed") || !contains(errStr, "ANTHROPIC_API_KEY") {
		t.Errorf("Expected error to contain 'authentication failed' and 'ANTHROPIC_API_KEY', got: %s", errStr)
	}
}

func TestEnhanceSDKError_RateLimit(t *testing.T) {
	err := errors.New("rate limit exceeded: HTTP 429")

	enhanced := EnhanceSDKError(err)
	if enhanced == nil {
		t.Fatal("EnhanceSDKError should not return nil")
	}

	errStr := enhanced.Error()
	if !contains(errStr, "rate limit") {
		t.Errorf("Expected error to contain 'rate limit', got: %s", errStr)
	}
}

func TestEnhanceSDKError_Unauthorized(t *testing.T) {
	err := errors.New("unauthorized: HTTP 401")

	enhanced := EnhanceSDKError(err)
	if enhanced == nil {
		t.Fatal("EnhanceSDKError should not return nil")
	}

	errStr := enhanced.Error()
	if !contains(errStr, "authorization failed") {
		t.Errorf("Expected error to contain 'authorization failed', got: %s", errStr)
	}
}

func TestEnhanceSDKError_NetworkError(t *testing.T) {
	err := errors.New("connection refused")

	enhanced := EnhanceSDKError(err)
	if enhanced == nil {
		t.Fatal("EnhanceSDKError should not return nil")
	}

	errStr := enhanced.Error()
	if !contains(errStr, "network error") {
		t.Errorf("Expected error to contain 'network error', got: %s", errStr)
	}
}

func TestEnhanceSDKError_Timeout(t *testing.T) {
	err := errors.New("context deadline exceeded")

	enhanced := EnhanceSDKError(err)
	if enhanced == nil {
		t.Fatal("EnhanceSDKError should not return nil")
	}

	errStr := enhanced.Error()
	if !contains(errStr, "timeout") {
		t.Errorf("Expected error to contain 'timeout', got: %s", errStr)
	}
}

func TestEnhanceSDKError_MissingAPIKey(t *testing.T) {
	err := errors.New("ANTHROPIC_API_KEY environment variable not set")

	enhanced := EnhanceSDKError(err)
	if enhanced == nil {
		t.Fatal("EnhanceSDKError should not return nil")
	}

	errStr := enhanced.Error()
	if !contains(errStr, "ANTHROPIC_API_KEY") {
		t.Errorf("Expected error to mention ANTHROPIC_API_KEY, got: %s", errStr)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
