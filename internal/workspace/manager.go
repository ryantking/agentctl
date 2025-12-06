package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ryantking/agentctl/internal/git"
)

// WorkspaceManager manages workspace lifecycle operations.
type WorkspaceManager struct { //nolint:revive // Stuttering is acceptable for exported manager types
	repoRoot string
}

// NewManager creates a new WorkspaceManager.
func NewManager() (*WorkspaceManager, error) {
	repoRoot, err := git.GetRepoRoot()
	if err != nil {
		return nil, ErrNotInGitRepo
	}
	return &WorkspaceManager{repoRoot: repoRoot}, nil
}

// NewManagerAt creates a new WorkspaceManager at a specific repository root.
func NewManagerAt(repoRoot string) (*WorkspaceManager, error) {
	return &WorkspaceManager{repoRoot: repoRoot}, nil
}

// ListWorkspaces lists all workspaces.
func (m *WorkspaceManager) ListWorkspaces(managedOnly bool) ([]Workspace, error) {
	workspaces, err := DiscoverWorkspaces(m.repoRoot)
	if err != nil {
		return nil, err
	}
	if managedOnly {
		var managed []Workspace
		for _, w := range workspaces {
			if w.IsManaged() && !w.IsMain {
				managed = append(managed, w)
			}
		}
		return managed, nil
	}
	return workspaces, nil
}

// GetWorkspace finds workspace by branch name.
func (m *WorkspaceManager) GetWorkspace(branch string) (*Workspace, error) {
	workspace, err := FindWorkspaceByBranch(branch, m.repoRoot)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, fmt.Errorf("%w: %s", ErrWorkspaceNotFound, branch)
	}
	return workspace, nil
}

// CreateWorkspace creates a new workspace with worktree.
func (m *WorkspaceManager) CreateWorkspace(branch string, baseBranch string) (*Workspace, error) {
	workspacePath, err := GetWorkspacePath(branch, m.repoRoot)
	if err != nil {
		return nil, err
	}

	// Check if workspace directory already exists
	if _, err := os.Stat(workspacePath); err == nil {
		return nil, fmt.Errorf("%w: %s", ErrWorkspaceExists, workspacePath)
	}

	// Check if branch is already checked out in another worktree
	existing, err := FindWorkspaceByBranch(branch, m.repoRoot)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("%w: %s", ErrBranchInUse, existing.Path)
	}

	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(workspacePath), 0755); err != nil { //nolint:gosec // Workspace directories need to be readable
		return nil, fmt.Errorf("failed to create workspace directory: %w", err)
	}

	// Check if branch exists
	branchExists, err := git.BranchExists(m.repoRoot, branch)
	if err != nil {
		return nil, err
	}

	if branchExists {
		// Branch exists, just create worktree
		if err := git.AddWorktree(m.repoRoot, workspacePath, branch, false, ""); err != nil {
			return nil, fmt.Errorf("failed to create worktree: %w", err)
		}
	} else {
		// Create new branch from base
		if baseBranch == "" {
			baseBranch, err = git.GetCurrentBranch(m.repoRoot)
			if err != nil || baseBranch == "" {
				baseBranch = "HEAD"
			}
		}
		if err := git.AddWorktree(m.repoRoot, workspacePath, branch, true, baseBranch); err != nil {
			return nil, fmt.Errorf("failed to create worktree: %w", err)
		}
	}

	// Return the newly created workspace
	workspace, err := FindWorkspaceByBranch(branch, m.repoRoot)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, fmt.Errorf("workspace created but could not be found")
	}
	return workspace, nil
}

// DeleteWorkspace removes a workspace.
func (m *WorkspaceManager) DeleteWorkspace(branch string, force bool) error {
	workspace, err := m.GetWorkspace(branch)
	if err != nil {
		return err
	}

	// Check if clean
	if !force {
		isClean, status := workspace.IsClean()
		if !isClean {
			return fmt.Errorf("workspace has uncommitted changes (%s). Use --force to remove anyway", status)
		}
	}

	if err := git.RemoveWorktree(m.repoRoot, workspace.Path, force); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	// Clean up empty parent directories
	parent := filepath.Dir(workspace.Path)
	for parent != filepath.Dir(parent) {
		dir, err := os.ReadDir(parent)
		if err != nil {
			break
		}
		if len(dir) > 0 {
			break // Directory not empty
		}
		if err := os.Remove(parent); err != nil {
			break
		}
		parent = filepath.Dir(parent)
	}

	return nil
}

// CleanWorkspaces removes clean/merged workspaces.
func (m *WorkspaceManager) CleanWorkspaces(checkMerged bool) ([]string, error) {
	var removed []string
	workspaces, err := m.ListWorkspaces(true)
	if err != nil {
		return nil, err
	}

	for _, workspace := range workspaces {
		if workspace.IsMain {
			continue
		}

		isClean, _ := workspace.IsClean()
		if !checkMerged || isClean {
			if workspace.Branch != "" {
				if err := m.DeleteWorkspace(workspace.Branch, !checkMerged); err != nil {
					// Skip workspaces that can't be deleted
					continue
				}
				removed = append(removed, workspace.Branch)
			}
		}
	}

	return removed, nil
}

// GetWorkspaceStatus gets detailed status information for a workspace.
func (m *WorkspaceManager) GetWorkspaceStatus(workspace *Workspace) (map[string]interface{}, error) {
	isClean, status := git.IsWorktreeClean(workspace.Path)

	result := map[string]interface{}{
		"path":     workspace.Path,
		"branch":   workspace.Branch,
		"commit":   workspace.Commit,
		"is_clean": isClean,
		"status":   status,
	}

	// Get ahead/behind information
	if workspace.Branch != "" {
		// Get local branch commit
		localCommit, err := git.RunGit(workspace.Path, "rev-parse", "HEAD")
		if err == nil {
			// Get remote branch commit
			remoteCommit, err := git.RunGit(workspace.Path, "rev-parse", fmt.Sprintf("origin/%s", workspace.Branch))
			if err == nil {
				// Calculate ahead/behind
				ahead, behind, err := calculateAheadBehind(workspace.Path, localCommit, remoteCommit)
				if err == nil {
					result["ahead_behind"] = map[string]int{
						"ahead":  ahead,
						"behind": behind,
					}
				}
			}
		}
	}

	return result, nil
}

// calculateAheadBehind calculates how many commits ahead and behind local is compared to remote.
func calculateAheadBehind(repoPath, localCommit, remoteCommit string) (int, int, error) {
	// Count commits in local but not in remote (ahead)
	aheadStr, err := git.RunGit(repoPath, "rev-list", "--count", fmt.Sprintf("%s..%s", remoteCommit, localCommit))
	if err != nil {
		return 0, 0, err
	}
	ahead, err := strconv.Atoi(aheadStr)
	if err != nil {
		return 0, 0, err
	}

	// Count commits in remote but not in local (behind)
	behindStr, err := git.RunGit(repoPath, "rev-list", "--count", fmt.Sprintf("%s..%s", localCommit, remoteCommit))
	if err != nil {
		return 0, 0, err
	}
	behind, err := strconv.Atoi(behindStr)
	if err != nil {
		return 0, 0, err
	}

	return ahead, behind, nil
}

// GetWorkspaceDiff gets git diff from workspace to target branch.
func (m *WorkspaceManager) GetWorkspaceDiff(workspace *Workspace, targetBranch string) (string, error) {
	// Get diff between target branch and HEAD
	diff, err := git.RunGit(workspace.Path, "diff", fmt.Sprintf("%s..HEAD", targetBranch))
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %w", err)
	}

	return diff, nil
}
