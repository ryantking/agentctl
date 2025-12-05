package workspace

import (
	"fmt"

	"github.com/ryantking/agentctl/internal/output"
	"github.com/ryantking/agentctl/internal/ui"
	"github.com/ryantking/agentctl/internal/workspace"
	"github.com/spf13/cobra"
)

// NewWorkspaceDeleteCmd creates the workspace delete command.
func NewWorkspaceDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:               "delete [branch]",
		Short:             "Delete a workspace",
		Long:              `By default, only deletes workspaces with no uncommitted changes.
Use --force to delete even with changes (WARNING: data loss). If no branch is provided, opens an interactive picker.`,
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

			if err := manager.DeleteWorkspace(branch, force); err != nil {
				if jsonMode {
					return output.ErrorJSON(err)
				}
				output.Error(err)
				return err
			}

			data := map[string]interface{}{
				"branch": branch,
			}

			if jsonMode {
				return output.SuccessJSON(data)
			}

			fmt.Printf("Deleted workspace for branch: %s\n", branch)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force deletion even if workspace has uncommitted changes")

	return cmd
}
