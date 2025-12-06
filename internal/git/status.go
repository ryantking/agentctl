package git

import (
	"fmt"
	"strconv"
	"strings"
)

// IsWorktreeClean checks if a worktree has uncommitted changes.
// Returns (isClean, statusMessage).
func IsWorktreeClean(worktreePath string) (bool, string) {
	lines, err := RunGitLines(worktreePath, "status", "--porcelain")
	if err != nil {
		return false, fmt.Sprintf("failed to check status: %v", err)
	}

	if len(lines) == 0 {
		return true, "clean"
	}

	// Parse porcelain output
	// Format: XY filename
	// X = status of index, Y = status of work tree
	// Common values: M = modified, A = added, D = deleted, ? = untracked, space = unmodified
	var staged, modified, untracked int
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		x := line[0]
		y := line[1]

		// Count staged changes (X != space)
		if x != ' ' && x != '?' {
			staged++
		}

		// Count modified/untracked in worktree (Y != space)
		if y != ' ' {
			if y == '?' {
				untracked++
			} else {
				modified++
			}
		}

		// Also count untracked files (X == '?' or Y == '?')
		if x == '?' {
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
