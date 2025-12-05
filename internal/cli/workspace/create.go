package workspace

import (
	"fmt"
	"os"

	"github.com/ryantking/agentctl/internal/context"
	"github.com/ryantking/agentctl/internal/output"
	"github.com/ryantking/agentctl/internal/workspace"
	"github.com/spf13/cobra"
)

// NewWorkspaceCreateCmd creates the workspace create command.
func NewWorkspaceCreateCmd() *cobra.Command {
	var baseBranch string

	cmd := &cobra.Command{
		Use:   "create <branch>",
		Short: "Create a new workspace with git worktree",
		Long: `Create a new workspace at ~/.claude/workspaces/<repo>/<branch>/
and copies necessary context files (CLAUDE.md, settings.local.json, .mcp.json).`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonMode, _ := cmd.Flags().GetBool("json")
			branch := args[0]

			manager, err := workspace.NewManager()
			if err != nil {
				if jsonMode {
					return output.ErrorJSON(err)
				}
				output.Error(err)
				return err
			}

			ws, err := manager.CreateWorkspace(branch, baseBranch)
			if err != nil {
				if jsonMode {
					return output.ErrorJSON(err)
				}
				output.Error(err)
				return err
			}

			// Copy Claude context files
			copiedFiles, err := context.CopyClaudeContext(ws.Path, ws.RepoRoot)
			if err != nil {
				// Non-fatal error, just log it
				if !jsonMode {
					fmt.Fprintf(os.Stderr, "Warning: failed to copy context files: %v\n", err)
				}
			}

			data := map[string]interface{}{
				"path":   ws.Path,
				"branch": ws.Branch,
				"commit": ws.Commit,
			}

			if jsonMode {
				return output.SuccessJSON(data)
			}

			fmt.Printf("Created workspace: %s\n", ws.Path)
			if len(copiedFiles) > 0 {
				fmt.Printf("Copied context: %v\n", copiedFiles)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&baseBranch, "base", "b", "", "Base branch to create from (defaults to current branch)")

	return cmd
}
