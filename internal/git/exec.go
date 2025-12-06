package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// RunGit executes a git command in the specified repository path.
// Uses git -C flag to change directory before running the command.
// Returns the stdout output as a string, or an error if the command fails.
func RunGit(repoPath string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", repoPath}, args...)...)
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			stderr := string(exitError.Stderr)
			return "", fmt.Errorf("git command failed: %w\nstderr: %s", err, stderr)
		}
		return "", fmt.Errorf("git command failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// RunGitLines executes a git command and returns the output as a slice of lines.
// Empty lines are filtered out.
func RunGitLines(repoPath string, args ...string) ([]string, error) {
	output, err := RunGit(repoPath, args...)
	if err != nil {
		return nil, err
	}
	if output == "" {
		return []string{}, nil
	}
	lines := strings.Split(output, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result, nil
}
