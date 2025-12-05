package cli

import (
	"github.com/spf13/cobra"
)

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

	cmd.AddCommand(
		NewVersionCmd(),
		NewStatusCmd(),
		NewWorkspaceCmd(),
		NewHookCmd(),
		NewInitCmd(),
	)

	return cmd
}
