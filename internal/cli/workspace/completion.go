package workspace

import (
	"github.com/ryantking/agentctl/internal/workspace"
	"github.com/spf13/cobra"
)

// completeWorkspaceNames provides completion for workspace branch names.
func completeWorkspaceNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Don't complete if we already have an argument
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	manager, err := workspace.NewManager()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// Get all workspaces including main/master
	workspaces, err := manager.ListWorkspaces(false)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// Extract branch names
	branches := make([]string, 0, len(workspaces))
	for _, w := range workspaces {
		branch := w.Branch
		if branch == "" {
			branch = "detached"
		}
		// Filter by toComplete prefix if provided
		if toComplete == "" || len(branch) >= len(toComplete) && branch[:len(toComplete)] == toComplete {
			branches = append(branches, branch)
		}
	}

	return branches, cobra.ShellCompDirectiveNoFileComp
}
