// Package agent provides CLI-based agent execution using the claude CLI.
package agent

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
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

// Execute runs a prompt through the claude CLI and returns the response.
// Uses --print flag for non-interactive output.
func (a *Agent) Execute(ctx context.Context, prompt string) (string, error) {
	// Check if claude CLI is available
	cliPath, err := exec.LookPath(a.CLIPath)
	if err != nil {
		return "", fmt.Errorf("claude CLI not found: %w\n\nTo fix this:\n  - Install Claude Code: https://claude.ai/code\n  - Or set ANTHROPIC_API_KEY environment variable", err)
	}

	// Build command: claude --print <prompt>
	cmd := exec.CommandContext(ctx, cliPath, "--print", prompt) //nolint:gosec // claude CLI is a trusted local binary
	
	// Set working directory to current directory (CLI will use repo context)
	wd, _ := os.Getwd()
	cmd.Dir = wd

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("claude CLI execution failed: %w\n\nOutput: %s", err, string(output))
	}

	content := strings.TrimSpace(string(output))
	if content == "" {
		return "", fmt.Errorf("empty output from claude CLI")
	}

	return content, nil
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
