package git

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
)

// IsWorktreeClean checks if a worktree has uncommitted changes.
// Returns (isClean, statusMessage).
func IsWorktreeClean(worktreePath string) (bool, string) {
	repo, err := OpenRepo(worktreePath)
	if err != nil {
		return false, fmt.Sprintf("failed to open repository: %v", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return false, fmt.Sprintf("failed to get worktree: %v", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return false, fmt.Sprintf("failed to check status: %v", err)
	}

	if status.IsClean() {
		return true, "clean"
	}

	// Count changes
	var staged, modified, untracked int
	for _, fileStatus := range status {
		if fileStatus.Staging != git.Unmodified {
			staged++
		}
		if fileStatus.Worktree != git.Unmodified {
			modified++
		}
		if fileStatus.Staging == git.Untracked || fileStatus.Worktree == git.Untracked {
			untracked++
		}
	}

	var parts []string
	if staged > 0 {
		parts = append(parts, strconv.Itoa(staged)+" staged")
	}
	if modified > 0 {
		parts = append(parts, strconv.Itoa(modified)+" modified")
	}
	if untracked > 0 {
		parts = append(parts, strconv.Itoa(untracked)+" untracked")
	}

	if len(parts) == 0 {
		return true, "clean"
	}

	return false, strings.Join(parts, ", ")
}

// GetStatusSummary returns a brief git status summary.
func GetStatusSummary(repoRoot string) (string, error) {
	isClean, status := IsWorktreeClean(repoRoot)
	if isClean {
		return "clean", nil
	}
	return status, nil
}
