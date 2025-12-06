package hook

import (
	"os"

	"github.com/ryantking/agentctl/internal/hook"
	"github.com/spf13/cobra"
)

// NewHookNotifyInputCmd creates the hook notify-input command.
func NewHookNotifyInputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notify-input [message]",
		Short: "Notification hook - sends notification when input is needed",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			input, _ := hook.GetStdinData()
			message := ""
			
			// Prefer message from stdin (hook input)
			if input != nil && input.Message != "" {
				message = input.Message
			} else if len(args) > 0 {
				// Fall back to command-line argument
				message = args[0]
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
		Short: "Stop hook - sends notification when a task completes",
		RunE: func(_ *cobra.Command, _ []string) error {
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
		Use:   "notify-error [message]",
		Short: "Send error notification",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			input, _ := hook.GetStdinData()
			message := ""
			
			// Prefer message from stdin (hook input)
			if input != nil && input.Message != "" {
				message = input.Message
			} else if len(args) > 0 {
				// Fall back to command-line argument
				message = args[0]
			}
			
			_ = hook.NotifyError(message)
			os.Exit(0)
			return nil
		},
	}

	return cmd
}

