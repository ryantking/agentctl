package hook

import (
	"os"

	"github.com/ryantking/agentctl/internal/hook"
	"github.com/spf13/cobra"
)

// NewHookPostEditCmd creates the hook post-edit command.
func NewHookPostEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "post-edit",
		Short: "PostToolUse hook for Edit tool",
		Long:  "Auto-commits changes if on a feature branch. Reads file path and session ID from stdin JSON.",
		RunE: func(_ *cobra.Command, _ []string) error {
			input, _ := hook.GetStdinData()
			filePath := hook.GetFilePath(input)
			_ = hook.PostEdit(filePath)
			os.Exit(0)
			return nil
		},
	}

	return cmd
}
