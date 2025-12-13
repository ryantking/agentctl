package memory

import (
	"github.com/spf13/cobra"
)

// NewMemoryCmd creates the memory command group.
// DEPRECATED: Use 'agentctl rules' commands instead. See docs/migration-memory-to-rules.md
func NewMemoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "[DEPRECATED] Manage agent memory files (AGENTS.md and CLAUDE.md)",
		Long: `[DEPRECATED] Commands for managing agent memory files following cross-platform standards.

The 'agentctl memory' commands are deprecated in favor of the new 'agentctl rules' system.
See 'agentctl rules --help' for the new commands, or docs/migration-memory-to-rules.md for migration guide.`,
		Deprecated: "Use 'agentctl rules' commands instead. See docs/migration-memory-to-rules.md",
	}

	cmd.AddCommand(
		NewMemoryInitCmd(),
		NewMemoryShowCmd(),
		NewMemoryValidateCmd(),
		NewMemoryIndexCmd(),
	)

	return cmd
}
