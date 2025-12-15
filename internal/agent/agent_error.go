// Package agent provides CLI-based agent execution using the claude CLI.
package agent

//nolint:revive // AgentError is a custom error type following exec.ExitError pattern

import (
	"fmt"
	"strings"
)

// AgentError represents an error that occurred during agent execution.
//nolint:revive // Type name follows exec.ExitError pattern and is clear in context
type AgentError struct {
	Program  string   // Program name (e.g., "claude")
	BinPath  string   // Full path to binary
	Args     []string // Command arguments
	ExitCode int      // Process exit code (-1 if not available)
	Stdout   string   // Standard output
	Stderr   string   // Standard error output
	Err      error    // Underlying error
}

// Error returns a formatted error message.
func (e *AgentError) Error() string {
	if e.ExitCode >= 0 {
		return fmt.Sprintf("agent %q failed (exit %d): %v", e.Program, e.ExitCode, e.Err)
	}
	return fmt.Sprintf("agent %q failed: %v", e.Program, e.Err)
}

// Unwrap returns the underlying error.
func (e *AgentError) Unwrap() error {
	return e.Err
}
