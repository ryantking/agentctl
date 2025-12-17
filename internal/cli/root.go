package cli

import (
	"os"

	agentclient "github.com/ryantking/agentctl/internal/agent"
	"github.com/ryantking/agentctl/internal/cli/rules"
	"github.com/spf13/cobra"
)

var (
	agentType   string
	agentBinary string
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

	// Add global flags for agent configuration
	cmd.PersistentFlags().StringVar(&agentType, "agent-type",
		getEnvOrDefault("AGENTCTL_AGENT_TYPE", "claude"),
		"Agent type to use (claude, aider, etc.)")
	cmd.PersistentFlags().StringVar(&agentBinary, "agent-binary",
		os.Getenv("AGENTCTL_AGENT_BINARY"),
		"Path to agent binary (defaults to agent type)")

	// Keep legacy --agent-cli flag for backward compatibility
	cmd.PersistentFlags().StringVar(&agentBinary, "agent-cli", "",
		"(deprecated) Use --agent-binary instead")
	_ = cmd.PersistentFlags().MarkDeprecated("agent-cli", "use --agent-binary instead")

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

// getEnvOrDefault returns environment variable value or default.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// NewAgentFromFlags creates an Agent from global flags and environment variables.
func NewAgentFromFlags() *agentclient.Agent {
	opts := []agentclient.Option{
		agentclient.WithType(agentType),
	}

	// Only set custom binary if flag provided
	if agentBinary != "" {
		opts = append(opts, agentclient.WithBinary(agentBinary))
	}
	// Otherwise Binary defaults to Type value

	return agentclient.NewAgent(opts...)
}

// GetAgentCLIPath returns the agent CLI path from flag or environment variable.
// Deprecated: Use NewAgentFromFlags() instead.
func GetAgentCLIPath() string {
	// Check flag value first
	if agentBinary != "" {
		return agentBinary
	}

	// Check environment variable
	if envPath := os.Getenv("AGENTCTL_CLI_PATH"); envPath != "" {
		return envPath
	}

	// Default to agent type or "claude"
	if agentType != "" {
		return agentType
	}

	return "claude"
}
