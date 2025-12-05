package workspace

import (
	"fmt"

	"github.com/ryantking/agentctl/internal/output"
	"github.com/ryantking/agentctl/internal/ui"
	"github.com/ryantking/agentctl/internal/workspace"
	"github.com/spf13/cobra"
)

// NewWorkspaceStatusCmd creates the workspace status command.
func NewWorkspaceStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "status [branch]",
		Short:             "Show detailed workspace status",
		Long:              "Displays status information including uncommitted changes, ahead/behind status relative to remote, and other details. If no branch is provided, opens an interactive picker.",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: completeWorkspaceNames,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonMode, _ := cmd.Flags().GetBool("json")

			manager, err := workspace.NewManager()
			if err != nil {
				if jsonMode {
					return output.ErrorJSON(err)
				}
				output.Error(err)
				return err
			}

			workspaces, err := manager.ListWorkspaces(true)
			if err != nil {
				if jsonMode {
					return output.ErrorJSON(err)
				}
				output.Error(err)
				return err
			}

			branch, err := ui.GetWorkspaceArg(args, workspaces)
			if err != nil {
				if jsonMode {
					return output.ErrorJSON(err)
				}
				output.Error(err)
				return err
			}

			ws, err := manager.GetWorkspace(branch)
			if err != nil {
				if jsonMode {
					return output.ErrorJSON(err)
				}
				output.Error(err)
				return err
			}

			statusInfo, err := manager.GetWorkspaceStatus(ws)
			if err != nil {
				if jsonMode {
					return output.ErrorJSON(err)
				}
				output.Error(err)
				return err
			}

			if jsonMode {
				return output.SuccessJSON(statusInfo)
			}

			// Human-readable output
			fmt.Printf("\nWorkspace: %v\n", statusInfo["branch"])
			fmt.Printf("Path:      %v\n", statusInfo["path"])
			fmt.Printf("Commit:    %v\n", statusInfo["commit"])
			fmt.Printf("Status:    %v\n", statusInfo["status"])

			if aheadBehind, ok := statusInfo["ahead_behind"].(map[string]int); ok {
				fmt.Printf("Sync:      %d ahead, %d behind origin\n", aheadBehind["ahead"], aheadBehind["behind"])
			}

			fmt.Println()

			return nil
		},
	}

	return cmd
}
