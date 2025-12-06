package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
)

// Repo wraps a go-git repository and provides convenience methods.
type Repo struct {
	*git.Repository
	path string
}

// OpenRepo opens a git repository at the given path.
func OpenRepo(path string) (*Repo, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	return &Repo{Repository: repo, path: path}, nil
}

// OpenRepoWithDiscover opens a git repository, discovering the root from the given path.
func OpenRepoWithDiscover(path string) (*Repo, error) {
	repoPath, err := discoverRepoRoot(path)
	if err != nil {
		return nil, err
	}
	return OpenRepo(repoPath)
}

// discoverRepoRoot finds the git repository root by walking up the directory tree.
// Correctly handles worktrees by finding the actual repository root.
func discoverRepoRoot(startPath string) (string, error) {
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}

	current := absPath
	for {
		gitDir := filepath.Join(current, ".git")
		info, err := os.Stat(gitDir)
		if err == nil {
			if info.IsDir() {
				// Regular git repository
				return current, nil
			}
			// Check if it's a worktree (file containing path to actual git dir)
			if !info.IsDir() {
				data, err := os.ReadFile(gitDir) //nolint:gosec // Reading .git file is safe, path is controlled
				if err == nil {
					// Format: gitdir: <path>
					if len(data) > 8 && string(data[:8]) == "gitdir: " {
						gitDirPath := string(data[8:])
						gitDirPath = strings.TrimSpace(gitDirPath)
						// Resolve relative paths
						if !filepath.IsAbs(gitDirPath) {
							gitDirPath = filepath.Join(current, gitDirPath)
						}
						
						// For worktrees, the gitDirPath points to .git/worktrees/<name>
						// The actual repo root is the parent of the .git directory
						// So we need to go up from .git/worktrees/<name> to .git to get the repo root
						worktreeGitDir := gitDirPath
						
						// Check if this is a worktree git dir (contains HEAD, index, etc.)
						// The actual repo root is 3 levels up: worktrees/<name> -> worktrees -> .git -> repo root
						if strings.Contains(worktreeGitDir, "/.git/worktrees/") {
							// Extract the .git directory path
							parts := strings.Split(worktreeGitDir, "/.git/worktrees/")
							if len(parts) > 0 {
								repoRoot := parts[0]
								// Verify it's a valid git repo by checking for .git directory
								if _, err := os.Stat(filepath.Join(repoRoot, ".git")); err == nil {
									return repoRoot, nil
								}
							}
						}
						
						// Fallback: try to find repo root by going up from worktree git dir
						// The worktree git dir is at .git/worktrees/<name>, so repo root is 3 levels up
						repoRoot := filepath.Dir(filepath.Dir(filepath.Dir(worktreeGitDir)))
						if _, err := os.Stat(filepath.Join(repoRoot, ".git")); err == nil {
							return repoRoot, nil
						}
						
						// If we can't find it, return the worktree directory itself
						// (this is what git rev-parse --show-toplevel returns for worktrees)
						return current, nil
					}
				}
			}
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", ErrNotInGitRepo
		}
		current = parent
	}
}

// ErrNotInGitRepo is returned when a git repository cannot be found.
var ErrNotInGitRepo = fmt.Errorf("not in a git repository")
