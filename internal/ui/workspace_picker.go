// Package ui provides user interface utilities for workspace selection.
package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ryantking/agentctl/internal/workspace"
)

// GetWorkspaceArg gets workspace name from args or prompts user to pick one using fzf.
func GetWorkspaceArg(args []string, workspaces []workspace.Workspace) (string, error) {
	if len(args) > 0 && args[0] != "" {
		return args[0], nil
	}

	// Check if we're in a non-interactive environment
	if !isTerminal(os.Stderr) {
		return "", fmt.Errorf("workspace name required when not in interactive terminal")
	}

	// Try to use fzf if available
	if fzfAvailable() {
		return pickWorkspaceWithFzf(workspaces)
	}

	// No fzf available, require branch name
	return "", fmt.Errorf("workspace name required (install fzf for interactive selection)")
}

// fzfAvailable checks if fzf is available in PATH.
func fzfAvailable() bool {
	_, err := exec.LookPath("fzf")
	return err == nil
}

// pickWorkspaceWithFzf uses fzf to let user select a workspace.
func pickWorkspaceWithFzf(workspaces []workspace.Workspace) (string, error) {
	if len(workspaces) == 0 {
		return "", fmt.Errorf("no workspaces available")
	}

	// Build input for fzf: format as "branch | status | commit"
	lines := make([]string, len(workspaces))
	for i, w := range workspaces {
		isClean, status := w.IsClean()
		statusIcon := "✓"
		if !isClean {
			statusIcon = "●"
		}
		branch := w.Branch
		if branch == "" {
			branch = "detached"
		}
		// Format: "branch | icon status | commit"
		lines[i] = fmt.Sprintf("%s | %s %s | %s", branch, statusIcon, status, w.Commit)
	}

	input := strings.Join(lines, "\n")

	// Run fzf with custom preview
	cmd := exec.Command("fzf",
		"--height", "40%",
		"--border",
		"--header", "Select workspace (use arrow keys, type to filter)",
		"--preview", "echo {}",
		"--preview-window", "down:3",
		"--delimiter", "|",
		"--with-nth", "1", // Only show branch name in main list
	)

	cmd.Stdin = strings.NewReader(input)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 130 {
			// Exit code 130 means user cancelled (Ctrl+C)
			return "", fmt.Errorf("selection cancelled")
		}
		return "", fmt.Errorf("fzf error: %w", err)
	}

	// Extract branch name from fzf output
	selected := strings.TrimSpace(string(output))
	if selected == "" {
		return "", fmt.Errorf("no workspace selected")
	}

	// Parse the selected line to get branch name
	parts := strings.Split(selected, "|")
	if len(parts) > 0 {
		branch := strings.TrimSpace(parts[0])
		// Verify it's a valid branch
		for _, w := range workspaces {
			if w.Branch == branch || (branch == "detached" && w.Branch == "") {
				return branch, nil
			}
		}
	}

	return "", fmt.Errorf("invalid selection")
}

func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}
