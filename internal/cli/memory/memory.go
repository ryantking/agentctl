package memory

import (
	"github.com/spf13/cobra"
)

// NewMemoryCmd creates the memory command group.
func NewMemoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "memory",
		Short: "Manage agent memory files (AGENTS.md and CLAUDE.md)",
		Long:  "Commands for managing agent memory files following cross-platform standards.",
	}

	cmd.AddCommand(
		NewMemoryInitCmd(),
	)

	return cmd
}
