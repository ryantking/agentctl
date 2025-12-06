// Package workspace provides workspace management CLI commands.
package workspace

import (
	"fmt"

	"github.com/ryantking/agentctl/internal/output"
	"github.com/ryantking/agentctl/internal/workspace"
	"github.com/spf13/cobra"
)

// NewWorkspaceCleanCmd creates the workspace clean command.
func NewWorkspaceCleanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Remove all clean workspaces",
		Long:  "Removes all workspaces that have no uncommitted changes. Useful for cleanup after completing work.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			jsonMode, _ := cmd.Flags().GetBool("json")

			manager, err := workspace.NewManager()
			if err != nil {
				if jsonMode {
					return output.ErrorJSON(err)
				}
				output.Error(err)
				return err
			}

			removed, err := manager.CleanWorkspaces(true)
			if err != nil {
				if jsonMode {
					return output.ErrorJSON(err)
				}
				output.Error(err)
				return err
			}

			if len(removed) == 0 {
				if !jsonMode {
					fmt.Println("No clean workspaces to remove")
				}
				if jsonMode {
					return output.SuccessJSON(map[string]interface{}{
						"removed": []string{},
					})
				}
				return nil
			}

			data := map[string]interface{}{
				"removed": removed,
			}

			if jsonMode {
				return output.SuccessJSON(data)
			}

			fmt.Printf("Removed %d workspace(s)\n", len(removed))
			return nil
		},
	}

	return cmd
}
