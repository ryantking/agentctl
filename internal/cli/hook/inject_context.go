package hook

import (
	"fmt"
	"os"

	"github.com/ryantking/agentctl/internal/hook"
	"github.com/spf13/cobra"
)

// NewHookInjectContextCmd creates the hook inject-context command.
func NewHookInjectContextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inject-context",
		Short: "UserPromptSubmit hook - injects live context into each user prompt",
		Long: `Outputs context information that gets automatically injected into
the conversation before Claude processes the user's message.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Consume stdin if present
			_, _ = hook.GetStdinData()

			context, err := hook.ContextInfo()
			if err != nil {
				return err
			}

			fmt.Println(context)
			os.Exit(0)
			return nil
		},
	}

	return cmd
}
