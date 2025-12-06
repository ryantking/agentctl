// Package git provides Git repository utilities and worktree management.
package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)


// GetRepoRoot returns the root directory of the current git repository.
// Correctly handles worktrees by finding the actual repository root
// instead of the worktree directory.
func GetRepoRoot() (string, error) {
	wd, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	return discoverRepoRoot(wd)
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
// Correctly handles worktrees by detecting the branch from the current directory.
func GetCurrentBranch(path string) (string, error) {
	// Discover the actual repo root (handles worktrees)
	repoRoot, err := discoverRepoRoot(path)
	if err != nil {
		return "", err
	}

	// Check if we're in a worktree by looking for .git file
	gitDir := filepath.Join(path, ".git")
	if info, err := os.Stat(gitDir); err == nil && !info.IsDir() {
		// We're in a worktree - read HEAD from the worktree's git directory
		data, err := os.ReadFile(gitDir) //nolint:gosec // Reading .git file is safe, path is controlled
		if err == nil && len(data) > 8 && string(data[:8]) == "gitdir: " {
			worktreeGitDir := strings.TrimSpace(string(data[8:]))
			if !filepath.IsAbs(worktreeGitDir) {
				worktreeGitDir = filepath.Join(path, worktreeGitDir)
			}
			headFile := filepath.Join(worktreeGitDir, "HEAD")
			if headData, err := os.ReadFile(headFile); err == nil { //nolint:gosec // Reading HEAD file is safe, path is controlled
				headRef := strings.TrimSpace(string(headData))
				// Format: ref: refs/heads/branch-name or just commit hash
				if strings.HasPrefix(headRef, "ref: ") {
					headRef = strings.TrimPrefix(headRef, "ref: ")
					headRef = strings.TrimSpace(headRef)
				}
				if strings.HasPrefix(headRef, "refs/heads/") {
					return strings.TrimPrefix(headRef, "refs/heads/"), nil
				}
				// Detached HEAD
				return "", nil
			}
		}
	}

	// Regular repo or fallback - use go-git
	repo, err := OpenRepo(repoRoot)
	if err != nil {
		return "", err
	}

	head, err := repo.Head()
	if err != nil {
		return "", nil // Detached HEAD or no commits
	}

	if !head.Name().IsBranch() {
		return "", nil // Not on a branch
	}

	return head.Name().Short(), nil
}

// BranchExists checks if a branch exists locally or remotely.
func BranchExists(repoRoot, branchName string) (bool, error) {
	repo, err := OpenRepo(repoRoot)
	if err != nil {
		return false, err
	}

	// Check local branches
	branches, err := repo.Branches()
	if err != nil {
		return false, err
	}

	found := false
	err = branches.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().Short() == branchName {
			found = true
			return fmt.Errorf("found") // Break iteration
		}
		return nil
	})
	if err != nil && err.Error() == "found" {
		return true, nil
	}
	if found {
		return true, nil
	}

	// Check remote branches
	remotes, err := repo.Remotes()
	if err != nil {
		return false, err
	}

	for _, remote := range remotes {
		refs, err := remote.List(&git.ListOptions{})
		if err != nil {
			continue
		}
		for _, ref := range refs {
			if ref.Name().Short() == fmt.Sprintf("%s/%s", remote.Config().Name, branchName) {
				return true, nil
			}
		}
	}

	return false, nil
}
