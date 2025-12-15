// Package agent provides CLI-based agent execution using the claude CLI.
package agent

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Agent executes prompts using the claude CLI.
type Agent struct {
	CLIPath string
}

// Option configures an Agent.
type Option func(*Agent)

// WithCLIPath sets a custom path to the claude binary.
func WithCLIPath(path string) Option {
	return func(a *Agent) {
		a.CLIPath = path
	}
}

// NewAgent creates a new agent that uses the claude CLI.
// The CLI handles authentication automatically (Claude Code session or API key).
func NewAgent(opts ...Option) *Agent {
	agent := &Agent{
		CLIPath: "claude",
	}

	// Apply options
	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

// Validate checks if the agent binary exists and is executable.
func (a *Agent) Validate() error {
	if a.CLIPath == "" {
		return fmt.Errorf("agent binary path not set")
	}

	if _, err := exec.LookPath(a.CLIPath); err != nil {
		return fmt.Errorf("agent binary %q not found in PATH: %w", a.CLIPath, err)
	}

	return nil
}

// Execute runs a prompt through the claude CLI and returns the response.
// Uses --print flag for non-interactive output.
func (a *Agent) Execute(ctx context.Context, prompt string) (string, error) {
	// Validate binary exists before execution
	if err := a.Validate(); err != nil {
		return "", fmt.Errorf("agent binary validation failed: %w", err)
	}

	// Add default timeout if none in context
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()
	}

	// Check if binary exists
	cliPath, err := exec.LookPath(a.CLIPath)
	if err != nil {
		return "", &AgentError{
			Program:  a.CLIPath,
			BinPath:  a.CLIPath,
			Args:     []string{"--print", prompt},
			ExitCode: -1,
			Stdout:   "",
			Stderr:   "",
			Err:      fmt.Errorf("agent binary %q not found in PATH: %w", a.CLIPath, err),
		}
	}

	// Build command: claude --print <prompt>
	cmd := exec.CommandContext(ctx, cliPath, "--print", prompt) //nolint:gosec // claude CLI is a trusted local binary

	// Set working directory to current directory (CLI will use repo context)
	wd, _ := os.Getwd()
	cmd.Dir = wd

	// Capture stdout and stderr separately
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	if err := cmd.Run(); err != nil {
		// Check context errors
		if ctx.Err() == context.Canceled {
			return "", &AgentError{
				Program:  a.CLIPath,
				BinPath:  cliPath,
				Args:     []string{"--print", prompt},
				ExitCode: -1,
				Stdout:   stdout.String(),
				Stderr:   stderr.String(),
				Err:      fmt.Errorf("agent execution cancelled by user"),
			}
		}
		if ctx.Err() == context.DeadlineExceeded {
			deadline, _ := ctx.Deadline()
			timeout := time.Until(deadline)
			return "", &AgentError{
				Program:  a.CLIPath,
				BinPath:  cliPath,
				Args:     []string{"--print", prompt},
				ExitCode: -1,
				Stdout:   stdout.String(),
				Stderr:   stderr.String(),
				Err:      fmt.Errorf("agent execution timed out after %s", timeout),
			}
		}

		// Get exit code
		exitCode := -1
		if cmd.ProcessState != nil {
			exitCode = cmd.ProcessState.ExitCode()
		}

		return "", &AgentError{
			Program:  a.CLIPath,
			BinPath:  cliPath,
			Args:     []string{"--print", prompt},
			ExitCode: exitCode,
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			Err:      err,
		}
	}

	// Validate output
	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return "", &AgentError{
			Program:  a.CLIPath,
			BinPath:  cliPath,
			Args:     []string{"--print", prompt},
			ExitCode: 0,
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			Err:      fmt.Errorf("agent produced no output"),
		}
	}

	return output, nil
}

// ExecuteWithSystem runs a prompt with a system message through the claude CLI.
// Combines system and user prompts into a single message.
func (a *Agent) ExecuteWithSystem(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	// Combine system and user prompts
	fullPrompt := fmt.Sprintf("%s\n\n%s", systemPrompt, userPrompt)
	return a.Execute(ctx, fullPrompt)
}

// IsConfigured checks if claude CLI is available or ANTHROPIC_API_KEY is set.
// The CLI handles auth automatically (Claude Code session or API key).
func IsConfigured() bool {
	// Check if CLI is available
	if _, err := exec.LookPath("claude"); err == nil {
		return true
	}
	// Fallback: check API key (for environments without CLI)
	return os.Getenv("ANTHROPIC_API_KEY") != ""
}

// NewClientOrNil is kept for backward compatibility with status.go.
// Returns a zero-value struct since we don't need SDK client anymore.
//nolint:revive // Function name kept for backward compatibility
func NewClientOrNil() (struct{}, error) {
	return struct{}{}, nil
}
