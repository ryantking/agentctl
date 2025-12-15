// Package agent provides CLI-based agent execution using the claude CLI.
package agent

import (
	"fmt"
	"strings"
)

// AgentError represents an error that occurred during agent execution.
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
	var parts []string

	// Main error message
	if e.ExitCode >= 0 {
		parts = append(parts, fmt.Sprintf("agent %q failed (exit %d): %v", e.Program, e.ExitCode, e.Err))
	} else {
		parts = append(parts, fmt.Sprintf("agent %q failed: %v", e.Program, e.Err))
	}

	// Command details
	parts = append(parts, fmt.Sprintf("\nCommand: %s %s", e.BinPath, strings.Join(e.Args, " ")))

	// Stderr if available
	if e.Stderr != "" {
		parts = append(parts, fmt.Sprintf("\nStderr: %s", e.Stderr))
	}

	// Stdout if available (for debugging)
	if e.Stdout != "" && len(e.Stdout) < 200 {
		parts = append(parts, fmt.Sprintf("\nStdout: %s", e.Stdout))
	}

	return strings.Join(parts, "")
}

// Unwrap returns the underlying error.
func (e *AgentError) Unwrap() error {
	return e.Err
}
