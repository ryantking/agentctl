package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Worktree represents a git worktree.
type Worktree struct {
	Path   string
	Branch string
	Commit string
}

// ListWorktrees returns all worktrees for a repository.
func ListWorktrees(repoRoot string) ([]Worktree, error) {
	var worktrees []Worktree

	// Add the main worktree
	mainWorktree, err := getMainWorktree(repoRoot)
	if err == nil {
		worktrees = append(worktrees, mainWorktree)
	}

	// List additional worktrees from .git/worktrees
	worktreesDir := filepath.Join(repoRoot, ".git", "worktrees")
	if info, err := os.Stat(worktreesDir); err == nil && info.IsDir() {
		entries, err := os.ReadDir(worktreesDir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				wtPath := filepath.Join(worktreesDir, entry.Name())
				wt, err := parseWorktreeDir(wtPath, repoRoot)
				if err == nil {
					worktrees = append(worktrees, wt)
				}
			}
		}
	}

	return worktrees, nil
}

// getMainWorktree gets information about the main worktree.
func getMainWorktree(repoRoot string) (Worktree, error) {
	wt := Worktree{Path: repoRoot}

	// Get branch name
	branch, err := GetCurrentBranch(repoRoot)
	if err == nil && branch != "" {
		wt.Branch = branch
	}

	// Get commit hash
	commit, err := RunGit(repoRoot, "rev-parse", "--short", "HEAD")
	if err == nil && commit != "" {
		wt.Commit = commit
	}

	return wt, nil
}

// parseWorktreeDir parses a worktree directory in .git/worktrees.
func parseWorktreeDir(worktreeDir, repoRoot string) (Worktree, error) {
	wt := Worktree{}

	// Read gitdir file to get the worktree .git path
	gitdirFile := filepath.Join(worktreeDir, "gitdir")
	data, err := os.ReadFile(gitdirFile) //nolint:gosec // Reading gitdir file is safe, path is controlled
	if err != nil {
		return wt, err
	}
	gitdirPath := strings.TrimSpace(string(data))
	
	// The gitdir file contains the path to the worktree's .git directory
	// The worktree path is the parent directory of that .git directory
	wt.Path = filepath.Dir(gitdirPath)

	// Read HEAD file to get commit/branch
	headFile := filepath.Join(worktreeDir, "HEAD")
	data, err = os.ReadFile(headFile) //nolint:gosec // Reading HEAD file is safe, path is controlled
	if err != nil {
		return wt, err
	}
	headRef := strings.TrimSpace(string(data))

	if strings.HasPrefix(headRef, "ref: refs/heads/") {
		wt.Branch = strings.TrimPrefix(headRef, "ref: refs/heads/")
		// Get commit from branch ref in main repo
		refPath := filepath.Join(repoRoot, ".git", headRef[5:]) // Skip "ref: "
		if data, err := os.ReadFile(refPath); err == nil { //nolint:gosec // Reading ref file is safe, path is controlled
			commit := strings.TrimSpace(string(data))
			if len(commit) > 8 {
				wt.Commit = commit[:8]
			} else {
				wt.Commit = commit
			}
		}
	} else {
		// Detached HEAD - commit hash is in the HEAD file
		commit := strings.TrimSpace(string(data))
		if len(commit) > 8 {
			wt.Commit = commit[:8]
		} else {
			wt.Commit = commit
		}
	}

	return wt, nil
}

// AddWorktree creates a new worktree.
// If createBranch is true, creates a new branch from baseBranch (or HEAD if baseBranch is empty).
// If createBranch is false, checks out the existing branch.
func AddWorktree(repoRoot, path, branch string, createBranch bool, baseBranch string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	if createBranch {
		// Create new branch from base
		if baseBranch == "" {
			// Use HEAD as base
			_, err = RunGit(repoRoot, "worktree", "add", "-b", branch, absPath)
		} else {
			// Use specified base branch
			_, err = RunGit(repoRoot, "worktree", "add", "-b", branch, absPath, baseBranch)
		}
		if err != nil {
			return fmt.Errorf("failed to create worktree with new branch: %w", err)
		}
	} else {
		// Checkout existing branch
		_, err = RunGit(repoRoot, "worktree", "add", absPath, branch)
		if err != nil {
			return fmt.Errorf("failed to create worktree: %w", err)
		}
	}

	return nil
}

// RemoveWorktree removes a worktree.
func RemoveWorktree(repoRoot, path string, force bool) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, absPath)

	_, err = RunGit(repoRoot, args...)
	if err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	return nil
}

// GetWorktreePath returns the absolute path of a worktree.
func GetWorktreePath(repoRoot, path string) (string, error) {
	// If path is relative, make it absolute relative to repo root
	if !filepath.IsAbs(path) {
		path = filepath.Join(repoRoot, path)
	}
	return filepath.Abs(path)
}
