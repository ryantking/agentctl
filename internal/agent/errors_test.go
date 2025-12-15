package agent

import (
	"errors"
	"strings"
	"testing"
)

func TestEnhanceSDKError_AuthenticationError(t *testing.T) {
	err := errors.New("not authenticated")

	enhanced := EnhanceSDKError(err)
	if enhanced == nil {
		t.Fatal("EnhanceSDKError should not return nil")
	}

	errStr := enhanced.Error()
	if !contains(errStr, "claude login") || !contains(errStr, "ANTHROPIC_API_KEY") {
		t.Errorf("Expected error to contain 'claude login' and 'ANTHROPIC_API_KEY', got: %s", errStr)
	}
}

func TestEnhanceSDKError_CLINotFound(t *testing.T) {
	err := errors.New("claude CLI not found: executable file not found")

	enhanced := EnhanceSDKError(err)
	if enhanced == nil {
		t.Fatal("EnhanceSDKError should not return nil")
	}

	errStr := enhanced.Error()
	if !contains(errStr, "claude CLI not found") || !contains(errStr, "Install Claude Code") {
		t.Errorf("Expected error to contain 'claude CLI not found' and 'Install Claude Code', got: %s", errStr)
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


func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
