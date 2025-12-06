// Package git provides Git repository utilities and worktree management.
package git

import (
	"fmt"
	"path/filepath"
	"strings"
)

// GetRepoRoot returns the root directory of the current git repository.
// Correctly handles worktrees by finding the actual repository root
// instead of the worktree directory.
func GetRepoRoot() (string, error) {
	wd, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	return GetRepoRootFromPath(wd)
}

// GetRepoRootFromPath returns the root directory of the git repository
// containing the given path. Uses git rev-parse --show-toplevel.
func GetRepoRootFromPath(path string) (string, error) {
	repoRoot, err := RunGit(path, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", ErrNotInGitRepo
	}
	return repoRoot, nil
}

// GetRepoName returns the name of the current git repository.
func GetRepoName() (string, error) {
	root, err := GetRepoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Base(root), nil
}

// GetCurrentBranch returns the name of the current branch.
// Returns empty string if in detached HEAD state.
// Correctly handles worktrees by using git rev-parse --abbrev-ref HEAD.
func GetCurrentBranch(path string) (string, error) {
	branch, err := RunGit(path, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	// If detached HEAD, git returns "HEAD", treat as empty
	if branch == "HEAD" {
		return "", nil
	}
	return branch, nil
}

// BranchExists checks if a branch exists locally or remotely.
func BranchExists(repoRoot, branchName string) (bool, error) {
	// Check local branch
	_, err := RunGit(repoRoot, "show-ref", "--verify", "--quiet", fmt.Sprintf("refs/heads/%s", branchName))
	if err == nil {
		return true, nil
	}

	// Check remote branch
	output, err := RunGit(repoRoot, "ls-remote", "--heads", "origin", branchName)
	if err != nil {
		return false, nil
	}
	// If output contains the branch name, it exists
	return strings.Contains(output, fmt.Sprintf("refs/heads/%s", branchName)), nil
}
