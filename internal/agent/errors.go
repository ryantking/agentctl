// Package agent provides CLI-based agent execution using the claude CLI.
package agent

import (
	"errors"
	"fmt"
)

// Sentinel errors for agent package.
var (
	// ErrNotFound indicates the agent binary was not found in PATH.
	ErrNotFound = errors.New("agent binary not found")

	// ErrTimeout indicates agent execution timed out.
	ErrTimeout = errors.New("agent execution timed out")

	// ErrCanceled indicates agent execution was cancelled.
	ErrCanceled = errors.New("agent execution cancelled")

	// ErrEmptyOutput indicates agent produced no output.
	ErrEmptyOutput = errors.New("agent produced no output")

	// ErrValidationFailed indicates agent binary validation failed.
	ErrValidationFailed = errors.New("agent binary validation failed")

	// ErrUnsupportedAgent indicates the agent type is not yet supported.
	ErrUnsupportedAgent = errors.New("agent type not supported")
)

// AgentError represents an error that occurred during agent execution.
//nolint:revive // Type name follows exec.ExitError pattern and is clear in context
type AgentError struct {
	Type     string   // Agent type (e.g., "claude", "codex", "cursor")
	Binary   string   // Full path to binary
	Args     []string // Command arguments
	ExitCode int      // Process exit code (-1 if not available)
	Stdout   string   // Standard output
	Stderr   string   // Standard error output
	Err      error    // Underlying error
}

// Error returns a formatted error message.
func (e *AgentError) Error() string {
	if e.ExitCode >= 0 {
		return fmt.Sprintf("agent %q failed (exit %d): %v", e.Type, e.ExitCode, e.Err)
	}
	return fmt.Sprintf("agent %q failed: %v", e.Type, e.Err)
}

// Unwrap returns the underlying error.
func (e *AgentError) Unwrap() error {
	return e.Err
}

// IsNotFound checks if an error is ErrNotFound.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsTimeout checks if an error is ErrTimeout.
func IsTimeout(err error) bool {
	return errors.Is(err, ErrTimeout)
}

// IsCanceled checks if an error is ErrCanceled.
func IsCanceled(err error) bool {
	return errors.Is(err, ErrCanceled)
}

// IsEmptyOutput checks if an error is ErrEmptyOutput.
func IsEmptyOutput(err error) bool {
	return errors.Is(err, ErrEmptyOutput)
}

// IsValidationFailed checks if an error is ErrValidationFailed.
func IsValidationFailed(err error) bool {
	return errors.Is(err, ErrValidationFailed)
}

// IsUnsupportedAgent checks if an error is ErrUnsupportedAgent.
func IsUnsupportedAgent(err error) bool {
	return errors.Is(err, ErrUnsupportedAgent)
}

// AsAgentError checks if an error is an AgentError and extracts it.
func AsAgentError(err error, target **AgentError) bool {
	var agentErr *AgentError
	if errors.As(err, &agentErr) {
		*target = agentErr
		return true
	}
	return false
}
