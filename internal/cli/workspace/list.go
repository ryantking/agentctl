package workspace

import (
	"fmt"
	"os"

	"github.com/ryantking/agentctl/internal/git"
	"github.com/ryantking/agentctl/internal/output"
	"github.com/ryantking/agentctl/internal/workspace"
	"github.com/spf13/cobra"
)

// NewWorkspaceListCmd creates the workspace list command.
func NewWorkspaceListCmd() *cobra.Command { //nolint:gocyclo // Complex command setup with multiple output formats
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all managed workspaces",
		Long:  "Shows workspaces in ~/.claude/workspaces/ with their status.",
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

			// Get all workspaces including main/master
			workspaces, err := manager.ListWorkspaces(false)
			if err != nil {
				if jsonMode {
					return output.ErrorJSON(err)
				}
				output.Error(err)
				return err
			}

			// Filter to only managed workspaces for JSON (backward compatibility)
			if jsonMode {
				var managed []workspace.Workspace
				for _, w := range workspaces {
					if w.IsManaged() && !w.IsMain {
						managed = append(managed, w)
					}
				}
				data := make([]map[string]interface{}, len(managed))
				for i, w := range managed {
					data[i] = workspaceToMapWithoutPath(w)
				}
				return output.WriteJSON(data)
			}

			// Get current branch to mark it with asterisk
			// Use current working directory to detect branch in worktrees
			var currentBranch string
			wd, err := os.Getwd()
			if err == nil {
				// Open repo from current directory to get correct branch in worktrees
				currentBranch, _ = git.GetCurrentBranch(wd)
			}

			// Simple text output - print immediately and exit
			if len(workspaces) == 0 {
				fmt.Print("\n  No workspaces found.\n\n  Create one with: agentctl workspace create <branch>\n\n")
				return nil
			}

			// Print workspaces with nice formatting
			for _, w := range workspaces {
				isClean, status := w.IsClean()
				statusIcon := "✓"
				if !isClean {
					statusIcon = "●"
				}

				branch := w.Branch
				if branch == "" {
					branch = "detached"
				}

				// Mark current workspace with asterisk
				marker := " "
				if branch == currentBranch {
					marker = "*"
				}

				// Format: * branch-name    ✓ clean    abc1234
				_, _ = fmt.Fprintf(os.Stdout, "%s %-30s %s %-20s %s\n",
					marker,
					branch,
					statusIcon,
					status,
					w.Commit,
				)
			}

			fmt.Println()
			return nil
		},
	}

	return cmd
}

// workspaceToMapWithoutPath converts workspace to map without path field.
func workspaceToMapWithoutPath(w workspace.Workspace) map[string]interface{} {
	isClean, status := w.IsClean()
	branch := w.Branch
	if branch == "" {
		branch = "detached"
	}
	return map[string]interface{}{
		"branch":     branch,
		"commit":     w.Commit,
		"is_main":    w.IsMain,
		"is_managed": w.IsManaged(),
		"is_clean":   isClean,
		"status":     status,
	}
}
