// Package hook provides hook command implementations.
package hook

import (
	"github.com/spf13/cobra"
)

// NewHookCmd creates the hook command group.
func NewHookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hook",
		Short: "Commands designed for use as Claude Code hooks",
		Long:  "Hook commands are designed to be called directly from Claude Code hooks. They handle stdin parsing, error handling, and exit codes appropriately for use as hook commands.",
	}

	cmd.AddCommand(
		NewHookPostEditCmd(),
		NewHookPostWriteCmd(),
		NewHookInjectContextCmd(),
		NewHookNotifyInputCmd(),
		NewHookNotifyStopCmd(),
		NewHookNotifyErrorCmd(),
	)

	return cmd
}
