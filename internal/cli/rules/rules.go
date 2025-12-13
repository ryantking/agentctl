package rules

import (
	"github.com/spf13/cobra"
)

// NewRulesCmd creates the rules command group.
func NewRulesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rules",
		Short: "Manage agent rules (.agent directory)",
		Long:  "Commands for managing agent rules in the .agent directory.",
	}

	cmd.AddCommand(
		NewRulesInitCmd(),
		NewRulesListCmd(),
		NewRulesShowCmd(),
		NewRulesRemoveCmd(),
	)

	return cmd
}
