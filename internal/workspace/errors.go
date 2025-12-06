// Package workspace provides workspace management types and error definitions.
package workspace

import "fmt"

var (
	// ErrWorkspaceExists indicates a workspace already exists for the given branch.
	ErrWorkspaceExists = fmt.Errorf("workspace already exists")
	// ErrWorkspaceNotFound indicates the requested workspace was not found.
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
	// ErrBranchInUse indicates the branch is already checked out in another worktree.
	ErrBranchInUse = fmt.Errorf("branch is already checked out")
	// ErrNotInGitRepo indicates the operation was attempted outside a git repository.
	ErrNotInGitRepo = fmt.Errorf("not in a git repository")
)

// WorkspaceError represents a workspace operation error.
type WorkspaceError struct { //nolint:revive // Stuttering is acceptable for exported error types
	Workspace string
	Op        string
	Err       error
}

func (e *WorkspaceError) Error() string {
	return fmt.Sprintf("%s %s: %v", e.Op, e.Workspace, e.Err)
}

func (e *WorkspaceError) Unwrap() error {
	return e.Err
}
