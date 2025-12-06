// Package hook provides hook command implementations for auto-commit and other lifecycle events.
package hook

import (
	"fmt"
	"path/filepath"

	gogit "github.com/go-git/go-git/v5"
	"github.com/ryantking/agentctl/internal/git"
)

// PostEdit auto-commits changes if on a feature branch.
// Reads file path from stdin JSON.
func PostEdit(filePath string) error {
	if filePath == "" {
		return nil
	}

	repoRoot, err := git.GetRepoRoot()
	if err != nil {
		return nil // Not in a repo, skip
	}

	branch, err := git.GetCurrentBranch(repoRoot)
	if err != nil || branch == "" {
		return nil
	}

	if isMainBranch(branch) {
		return nil // Skip on main/master
	}

	return gitAddAndCommit(repoRoot, filePath)
}

// PostWrite auto-commits new files if on a feature branch.
// Reads file path from stdin JSON.
func PostWrite(filePath string) error {
	if filePath == "" {
		return nil
	}

	repoRoot, err := git.GetRepoRoot()
	if err != nil {
		return nil // Not in a repo, skip
	}

	branch, err := git.GetCurrentBranch(repoRoot)
	if err != nil || branch == "" {
		return nil
	}

	if isMainBranch(branch) {
		return nil // Skip on main/master
	}

	return gitAddAndCommitNewFile(repoRoot, filePath)
}

func isMainBranch(branch string) bool {
	return branch == "main" || branch == "master"
}

func gitAddAndCommit(repoRoot, filePath string) error {
	repo, err := git.OpenRepo(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Make path relative to repo root
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}
	relPath, err := filepath.Rel(repoRoot, absPath)
	if err != nil {
		return err
	}

	// Stage the file
	if _, err := worktree.Add(relPath); err != nil {
		return fmt.Errorf("failed to stage file: %w", err)
	}

	// Check if there are staged changes
	status, err := worktree.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	fileStatus, ok := status[relPath]
	if !ok || fileStatus.Staging == gogit.Unmodified {
		// No changes to commit
		return nil
	}

	// Calculate commit message
	filename := filepath.Base(filePath)
	msg := fmt.Sprintf("Update %s: moderate changes", filename)

	// Create commit
	_, err = worktree.Commit(msg, &gogit.CommitOptions{})
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	return nil
}

func gitAddAndCommitNewFile(repoRoot, filePath string) error {
	repo, err := git.OpenRepo(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Make path relative to repo root
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}
	relPath, err := filepath.Rel(repoRoot, absPath)
	if err != nil {
		return err
	}

	// Stage the file
	if _, err := worktree.Add(relPath); err != nil {
		return fmt.Errorf("failed to stage file: %w", err)
	}

	// Check if file is staged
	status, err := worktree.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	fileStatus, ok := status[relPath]
	if !ok || fileStatus.Staging == gogit.Unmodified {
		// No changes to commit
		return nil
	}

	filename := filepath.Base(filePath)
	msg := fmt.Sprintf("Add new file: %s", filename)

	// Create commit
	_, err = worktree.Commit(msg, &gogit.CommitOptions{})
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	return nil
}
