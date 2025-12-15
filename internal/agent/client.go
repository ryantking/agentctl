// Package agent provides CLI-based agent execution using the claude CLI.
package agent

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Agent represents a CLI-based agent executor.
type Agent struct {
	Type    string        // Agent type (e.g., "claude", "codex", "cursor")
	Binary  string        // Path to binary (defaults to Type)
	Timeout time.Duration // Execution timeout (default: 5 minutes)
}

// Option configures an Agent.
type Option func(*Agent)

// WithType sets the agent type and updates Binary to match if Binary is still default.
func WithType(agentType string) Option {
	return func(a *Agent) {
		oldType := a.Type
		a.Type = agentType
		// Auto-update Binary if it matches the old Type
		if a.Binary == "" || a.Binary == oldType {
			a.Binary = agentType
		}
	}
}

// WithBinary sets the CLI binary path.
func WithBinary(path string) Option {
	return func(a *Agent) {
		a.Binary = path
	}
}

// WithTimeout sets the execution timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(a *Agent) {
		a.Timeout = timeout
	}
}

// NewAgent creates a new Agent instance with defaults.
func NewAgent(opts ...Option) *Agent {
	agent := &Agent{
		Type:    "claude",
		Binary:  "claude",
		Timeout: 5 * time.Minute,
	}
	for _, opt := range opts {
		opt(agent)
	}
	return agent
}

// Validate checks if the agent binary exists and is executable.
func (a *Agent) Validate() error {
	if a.Binary == "" {
		return fmt.Errorf("agent binary path not set: %w", ErrValidationFailed)
	}

	if _, err := exec.LookPath(a.Binary); err != nil {
		return fmt.Errorf("agent binary %q not found in PATH: %w", a.Binary, ErrNotFound)
	}

	return nil
}

// Execute runs a prompt through the claude CLI and returns the response.
// Uses --print flag for non-interactive output.
func (a *Agent) Execute(ctx context.Context, prompt string) (string, error) {
	return a.ExecuteWithLogger(ctx, prompt, nil)
}

// ExecuteWithLogger runs a prompt through the claude CLI and returns the response.
// Uses --print flag for non-interactive output.
// If logger is nil, uses default logger.
func (a *Agent) ExecuteWithLogger(ctx context.Context, prompt string, logger *slog.Logger) (string, error) {
	if logger == nil {
		logger = slog.Default()
	}

	// Validate binary exists before execution
	if err := a.Validate(); err != nil {
		logger.Error("agent binary validation failed",
			slog.String("type", a.Type),
			slog.String("binary", a.Binary),
			slog.Any("error", err))
		return "", fmt.Errorf("agent binary validation failed: %w", err)
	}

	// Add default timeout if none in context
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, a.Timeout)
		defer cancel()
	}

	// Check if binary exists
	binPath, err := exec.LookPath(a.Binary)
	if err != nil {
		logger.Error("agent binary not found",
			slog.String("type", a.Type),
			slog.String("binary", a.Binary),
			slog.Any("error", err))
		return "", &AgentError{
			Type:     a.Type,
			Binary:   a.Binary,
			Args:     []string{"--print", prompt},
			ExitCode: -1,
			Stdout:   "",
			Stderr:   "",
			Err:      fmt.Errorf("agent binary %q not found in PATH: %w", a.Binary, ErrNotFound),
		}
	}

	// Build command: claude --print <prompt>
	cmd := exec.CommandContext(ctx, binPath, "--print", prompt) //nolint:gosec // claude CLI is a trusted local binary

	// Set working directory to current directory (CLI will use repo context)
	wd, _ := os.Getwd()
	cmd.Dir = wd

	// Capture stdout and stderr separately
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	logger.Info("executing agent command",
		slog.String("type", a.Type),
		slog.String("binary", binPath),
		slog.Any("args", []string{"--print", prompt}),
		slog.String("working_dir", wd))

	// Execute command
	if err := cmd.Run(); err != nil {
		// Get exit code
		exitCode := -1
		if cmd.ProcessState != nil {
			exitCode = cmd.ProcessState.ExitCode()
		}

		// Check context errors
		if ctx.Err() == context.Canceled {
			logger.Warn("agent execution cancelled",
				slog.String("type", a.Type),
				slog.String("binary", binPath),
				slog.Int("exit_code", exitCode),
				slog.String("stderr", stderr.String()))
			return "", &AgentError{
				Type:     a.Type,
				Binary:   binPath,
				Args:     []string{"--print", prompt},
				ExitCode: -1,
				Stdout:   stdout.String(),
				Stderr:   stderr.String(),
				Err:      ErrCanceled,
			}
		}
		if ctx.Err() == context.DeadlineExceeded {
			deadline, _ := ctx.Deadline()
			timeout := time.Until(deadline)
			logger.Error("agent execution timed out",
				slog.String("type", a.Type),
				slog.String("binary", binPath),
				slog.Duration("timeout", timeout),
				slog.String("stderr", stderr.String()))
			return "", &AgentError{
				Type:     a.Type,
				Binary:   binPath,
				Args:     []string{"--print", prompt},
				ExitCode: -1,
				Stdout:   stdout.String(),
				Stderr:   stderr.String(),
				Err:      fmt.Errorf("agent execution timed out after %s: %w", timeout, ErrTimeout),
			}
		}

		logger.Error("agent execution failed",
			slog.String("type", a.Type),
			slog.String("binary", binPath),
			slog.Int("exit_code", exitCode),
			slog.String("stderr", stderr.String()),
			slog.String("stdout", stdout.String()),
			slog.Any("error", err))

		return "", &AgentError{
			Type:     a.Type,
			Binary:   binPath,
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
		logger.Error("agent produced no output",
			slog.String("program", a.CLIPath),
			slog.String("bin_path", cliPath),
			slog.String("stderr", stderr.String()))
		return "", &AgentError{
			Program:  a.CLIPath,
			BinPath:  cliPath,
			Args:     []string{"--print", prompt},
			ExitCode: 0,
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			Err:      ErrEmptyOutput,
		}
	}

	logger.Info("agent execution succeeded",
		slog.String("program", a.CLIPath),
		slog.String("bin_path", cliPath),
		slog.Int("output_length", len(output)))

	return output, nil
}

// ExecuteWithSystem runs a prompt with a system message through the claude CLI.
func (a *Agent) ExecuteWithSystem(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return a.ExecuteWithSystemLogger(ctx, systemPrompt, userPrompt, nil)
}

// ExecuteWithSystemLogger runs a prompt with a system message through the claude CLI.
// If logger is nil, uses default logger.
func (a *Agent) ExecuteWithSystemLogger(ctx context.Context, systemPrompt, userPrompt string, logger *slog.Logger) (string, error) {
	if logger == nil {
		logger = slog.Default()
	}

	// Combine system and user prompts
	combinedPrompt := fmt.Sprintf("System: %s\n\nUser: %s", systemPrompt, userPrompt)
	return a.ExecuteWithLogger(ctx, combinedPrompt, logger)
}

// IsConfigured checks if the agent is configured (CLI available or API key set).
func IsConfigured() bool {
	// Check for API key first (fastest check)
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		return true
	}

	// Check if CLI is available
	agent := NewAgent()
	return agent.Validate() == nil
}

// NewClientOrNil returns a zero-value struct for backward compatibility.
//
// Deprecated: Use NewAgent() instead.
func NewClientOrNil() (struct{}, error) {
	return struct{}{}, nil
}
