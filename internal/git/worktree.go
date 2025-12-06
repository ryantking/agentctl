package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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
	repo, err := OpenRepo(repoRoot)
	if err != nil {
		return nil, err
	}

	mainWorktree, err := getMainWorktree(repo, repoRoot)
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
func getMainWorktree(repo *Repo, repoRoot string) (Worktree, error) {
	wt := Worktree{Path: repoRoot}

	head, err := repo.Head()
	if err == nil {
		wt.Commit = head.Hash().String()[:8]
		if head.Name().IsBranch() {
			wt.Branch = head.Name().Short()
		}
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
func AddWorktree(repoRoot, path, branch string, createBranch bool, baseBranch string) error { //nolint:gocyclo // Complex worktree creation logic
	repo, err := OpenRepo(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Create the directory
	if err := os.MkdirAll(absPath, 0755); err != nil { //nolint:gosec // Worktree directories need to be readable
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var baseHash plumbing.Hash
	if createBranch {
		if baseBranch == "" {
			head, err := repo.Head()
			if err != nil {
				return fmt.Errorf("failed to get HEAD: %w", err)
			}
			baseHash = head.Hash()
		} else {
			ref, err := repo.Reference(plumbing.NewBranchReferenceName(baseBranch), true)
			if err != nil {
				return fmt.Errorf("base branch %s not found: %w", baseBranch, err)
			}
			baseHash = ref.Hash()
		}
	} else {
		ref, err := repo.Reference(plumbing.NewBranchReferenceName(branch), true)
		if err != nil {
			return fmt.Errorf("branch %s not found: %w", branch, err)
		}
		baseHash = ref.Hash()
	}

	// Create branch if needed
	if createBranch {
		refName := plumbing.NewBranchReferenceName(branch)
		ref := plumbing.NewHashReference(refName, baseHash)
		if err := repo.Storer.SetReference(ref); err != nil {
			return fmt.Errorf("failed to create branch: %w", err)
		}
	}

	// Create worktree directory structure
	worktreeID := generateWorktreeID()
	worktreeDir := filepath.Join(repoRoot, ".git", "worktrees", worktreeID)
	if err := os.MkdirAll(worktreeDir, 0755); err != nil { //nolint:gosec // Worktree directories need to be readable
		return fmt.Errorf("failed to create worktree directory: %w", err)
	}

	// Write gitdir file (points to worktree path)
	gitdirFile := filepath.Join(worktreeDir, "gitdir")
	if err := os.WriteFile(gitdirFile, []byte(absPath+"\n"), 0644); err != nil { //nolint:gosec // Git directory file needs to be readable
		return fmt.Errorf("failed to write gitdir: %w", err)
	}

	// Write HEAD file
	headFile := filepath.Join(worktreeDir, "HEAD")
	var headContent string
	if createBranch || !createBranch {
		headContent = fmt.Sprintf("ref: refs/heads/%s\n", branch)
	} else {
		headContent = baseHash.String() + "\n"
	}
	if err := os.WriteFile(headFile, []byte(headContent), 0644); err != nil { //nolint:gosec // Git HEAD file needs to be readable
		return fmt.Errorf("failed to write HEAD: %w", err)
	}

	// Create .git file in worktree pointing to worktree gitdir
	worktreeGitFile := filepath.Join(absPath, ".git")
	gitdirContent := fmt.Sprintf("gitdir: %s\n", worktreeDir)
	if err := os.WriteFile(worktreeGitFile, []byte(gitdirContent), 0644); err != nil { //nolint:gosec // Git .git file needs to be readable
		return fmt.Errorf("failed to write .git file: %w", err)
	}

	// Clone the repository to the worktree location using go-git
	worktreeRepo, err := git.PlainClone(absPath, false, &git.CloneOptions{
		URL: repoRoot,
	})
	if err != nil {
		// If clone fails, try opening existing
		worktreeRepo, err = git.PlainOpen(absPath)
		if err != nil {
			return fmt.Errorf("failed to setup worktree repository: %w", err)
		}
	}

	// Checkout the branch
	worktree, err := worktreeRepo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	checkoutOpts := git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
		Hash:   baseHash,
		Force:  true,
	}

	if err := worktree.Checkout(&checkoutOpts); err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	return nil
}

// RemoveWorktree removes a worktree.
func RemoveWorktree(repoRoot, path string, force bool) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Find the worktree entry in .git/worktrees
	worktreesDir := filepath.Join(repoRoot, ".git", "worktrees")
	entries, err := os.ReadDir(worktreesDir)
	if err != nil {
		return fmt.Errorf("failed to read worktrees directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		wtDir := filepath.Join(worktreesDir, entry.Name())
		gitdirFile := filepath.Join(wtDir, "gitdir")
		data, err := os.ReadFile(gitdirFile) //nolint:gosec // Reading gitdir file is safe, path is controlled
		if err != nil {
			continue
		}
		wtPath := strings.TrimSpace(string(data))
		if wtPath == absPath {
			// Found the worktree - remove it
			if force {
				// Remove the worktree directory
				if err := os.RemoveAll(absPath); err != nil {
					return fmt.Errorf("failed to remove worktree directory: %w", err)
				}
			} else {
				// Check if worktree is clean
				isClean, _ := IsWorktreeClean(absPath)
				if !isClean {
					return fmt.Errorf("worktree has uncommitted changes")
				}
				if err := os.RemoveAll(absPath); err != nil {
					return fmt.Errorf("failed to remove worktree directory: %w", err)
				}
			}
			// Remove the worktree entry
			if err := os.RemoveAll(wtDir); err != nil {
				return fmt.Errorf("failed to remove worktree entry: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("worktree not found at %s", path)
}

// generateWorktreeID generates a unique worktree ID.
func generateWorktreeID() string {
	// Simple implementation - in practice git uses a hash
	// For now, use a timestamp-based approach
	return fmt.Sprintf("worktree-%d", os.Getpid())
}

// GetWorktreePath returns the absolute path of a worktree.
func GetWorktreePath(repoRoot, path string) (string, error) {
	// If path is relative, make it absolute relative to repo root
	if !filepath.IsAbs(path) {
		path = filepath.Join(repoRoot, path)
	}
	return filepath.Abs(path)
}
