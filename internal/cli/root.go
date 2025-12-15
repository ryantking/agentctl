package cli

import (
	"os"

	"github.com/ryantking/agentctl/internal/cli/rules"
	"github.com/spf13/cobra"
)

var agentCLIPath string

// Execute runs the CLI application.
func Execute() error {
	return NewRootCmd().Execute()
}

// NewRootCmd creates the root command.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agentctl",
		Short: "A CLI tool for managing Claude Code configurations, hooks, and isolated workspaces using git worktrees",
		Long:  "A CLI tool for managing Claude Code configurations, hooks, and isolated workspaces using git worktrees.",
	}

	// Add global flag for agent CLI path
	cmd.PersistentFlags().StringVar(&agentCLIPath, "agent-cli", "claude", "Path to AI agent CLI binary (default: claude)")

	cmd.AddCommand(
		NewVersionCmd(),
		NewStatusCmd(),
		NewWorkspaceCmd(),
		NewHookCmd(),
		NewInitCmd(),
		rules.NewRulesCmd(),
	)

	return cmd
}

// GetAgentCLIPath returns the agent CLI path from flag or environment variable.
// Checks --agent-cli flag first, then AGENTCTL_CLI_PATH env var, defaults to "claude".
func GetAgentCLIPath() string {
	// Use flag value if set (non-default)
	if agentCLIPath != "" && agentCLIPath != "claude" {
		return agentCLIPath
	}

	// Check environment variable
	if envPath := os.Getenv("AGENTCTL_CLI_PATH"); envPath != "" {
		return envPath
	}

	// Default to "claude"
	return "claude"
}
