package hook

import (
	"os"

	"github.com/ryantking/agentctl/internal/hook"
	"github.com/spf13/cobra"
)

// NewHookNotifyInputCmd creates the hook notify-input command.
func NewHookNotifyInputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notify-input",
		Short: "Notification hook - sends notification when Claude needs input",
		RunE: func(cmd *cobra.Command, args []string) error {
			input, _ := hook.GetStdinData()
			message := "Claude needs your input to continue"
			if input != nil && input.Message != "" {
				message = input.Message
			}
			_ = hook.NotifyInput(message)
			os.Exit(0)
			return nil
		},
	}

	return cmd
}

// NewHookNotifyStopCmd creates the hook notify-stop command.
func NewHookNotifyStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notify-stop",
		Short: "Stop hook - sends notification when Claude completes a task",
		RunE: func(cmd *cobra.Command, args []string) error {
			input, _ := hook.GetStdinData()
			transcriptPath := ""
			if input != nil {
				transcriptPath = hook.GetTranscriptPath(input)
			}
			_ = hook.NotifyStop(transcriptPath)
			os.Exit(0)
			return nil
		},
	}

	return cmd
}

// NewHookNotifyErrorCmd creates the hook notify-error command.
func NewHookNotifyErrorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notify-error",
		Short: "Send error notification",
		RunE: func(cmd *cobra.Command, args []string) error {
			input, _ := hook.GetStdinData()
			message := "An error occurred during task execution"
			if input != nil && input.Message != "" {
				message = input.Message
			}
			_ = hook.NotifyError(message)
			os.Exit(0)
			return nil
		},
	}

	return cmd
}

