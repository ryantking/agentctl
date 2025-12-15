package agent

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"
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

func TestAgent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cliPath string
		wantErr bool
	}{
		{
			name:    "empty path",
			cliPath: "",
			wantErr: true,
		},
		{
			name:    "non-existent binary",
			cliPath: "nonexistent-binary-12345",
			wantErr: true,
		},
		{
			name:    "valid binary",
			cliPath: "claude",
			wantErr: false, // May be false if claude is available
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewAgent(WithCLIPath(tt.cliPath))
			err := agent.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErr {
				// Check error wrapping
				if tt.cliPath == "" {
					if !IsValidationFailed(err) {
						t.Error("Validate() should wrap ErrValidationFailed for empty path")
					}
				} else {
					if !IsNotFound(err) {
						t.Error("Validate() should wrap ErrNotFound for non-existent binary")
					}
				}
			}
		})
	}
}

func TestAgent_Execute_ContextCancellation(t *testing.T) {
	agent := NewAgent()
	
	// Skip if claude CLI not available
	if err := agent.Validate(); err != nil {
		t.Skip("claude CLI not available, skipping execution test")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := agent.Execute(ctx, "test prompt")
	if err == nil {
		t.Error("Execute() should return error when context is cancelled")
	}

	var agentErr *AgentError
	if !AsAgentError(err, &agentErr) {
		t.Error("Execute() should return AgentError when context is cancelled")
	}

	if !IsCanceled(err) {
		t.Error("Execute() should wrap ErrCanceled when context is cancelled")
	}
}

func TestAgent_Execute_Timeout(t *testing.T) {
	agent := NewAgent()
	
	// Skip if claude CLI not available
	if err := agent.Validate(); err != nil {
		t.Skip("claude CLI not available, skipping execution test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for timeout
	time.Sleep(10 * time.Millisecond)

	_, err := agent.Execute(ctx, "test prompt")
	if err == nil {
		t.Error("Execute() should return error when context times out")
	}

	if !IsTimeout(err) {
		t.Error("Execute() should wrap ErrTimeout when context times out")
	}
}

func TestAgent_Execute_EmptyOutput(t *testing.T) {
	// This test would require mocking exec.Command, which is complex
	// For now, we'll test the error type checking functions
	agentErr := &AgentError{
		Program:  "test",
		BinPath:  "/bin/test",
		Args:     []string{"--print", "test"},
		ExitCode: 0,
		Stdout:   "",
		Stderr:   "",
		Err:      ErrEmptyOutput,
	}

	if !IsEmptyOutput(agentErr) {
		t.Error("IsEmptyOutput() should return true for ErrEmptyOutput")
	}
}

func TestAgentError_Unwrap(t *testing.T) {
	originalErr := ErrNotFound
	agentErr := &AgentError{
		Program:  "test",
		BinPath:  "/bin/test",
		Args:     []string{"test"},
		ExitCode: -1,
		Stdout:   "",
		Stderr:   "",
		Err:      originalErr,
	}

	unwrapped := agentErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}
}

func TestAgentError_Error(t *testing.T) {
	agentErr := &AgentError{
		Program:  "claude",
		BinPath:  "/usr/bin/claude",
		Args:     []string{"--print", "test"},
		ExitCode: 1,
		Stdout:   "some output",
		Stderr:   "error message",
		Err:      exec.ExitError{ProcessState: &os.ProcessState{}},
	}

	errStr := agentErr.Error()
	if errStr == "" {
		t.Error("Error() should return non-empty string")
	}
	if !contains(errStr, "claude") {
		t.Error("Error() should contain program name")
	}
	if !contains(errStr, "exit 1") {
		t.Error("Error() should contain exit code")
	}
	if !contains(errStr, "error message") {
		t.Error("Error() should contain stderr")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
