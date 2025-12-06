package hook

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ryantking/agentctl/internal/git"
	"github.com/ryantking/agentctl/internal/github"
	"github.com/ryantking/agentctl/internal/workspace"
)

// ContextInfo generates context information for injection into prompts.
func ContextInfo() (string, error) { //nolint:gocyclo // Complex context gathering logic
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	cwd, err = filepath.Abs(cwd)
	if err != nil {
		return "", err
	}

	var lines []string
	lines = append(lines, "<context-refresh>")
	lines = append(lines, fmt.Sprintf("Path: %s", cwd))

	// Current workspace (if in one)
	currentWS := getCurrentWorkspace(cwd)
	if currentWS != nil {
		wsLine := fmt.Sprintf("Current Workspace: %s", currentWS["branch"])
		if status, ok := currentWS["status"].(string); ok && status != "" {
			wsLine += fmt.Sprintf(" (%s)", status)
		}
		lines = append(lines, wsLine)
	}

	// Git info (current branch)
	branch := getGitBranch(cwd)
	if branch != "" {
		status := getGitStatusSummary(cwd)
		gitLine := fmt.Sprintf("Branch: %s", branch)
		if status != "" {
			gitLine += fmt.Sprintf(" (%s)", status)
		}
		lines = append(lines, gitLine)
	}

	// All git branches with cleanliness
	allBranches := getAllGitBranches(cwd)
	if len(allBranches) > 0 {
		lines = append(lines, "Git Branches:")
		for branchName, branchStatus := range allBranches {
			lines = append(lines, fmt.Sprintf("  %s: %s", branchName, branchStatus))
		}
	}

	// Workspace sessions
	workspaces := getWorkspaceSessions()
	if len(workspaces) > 0 {
		lines = append(lines, "Workspaces:")
		for _, ws := range workspaces {
			wsLine := fmt.Sprintf("  %v", ws["branch"])
			if status, ok := ws["status"].(string); ok && status != "" {
				wsLine += fmt.Sprintf(" (%s)", status)
			}
			lines = append(lines, wsLine)
		}
	}

	// PR status
	pr := getPRStatus(cwd)
	if pr != nil {
		prLine := fmt.Sprintf("PR #%v: %v", pr["number"], pr["title"])
		var details []string
		if review, ok := pr["review"].(string); ok && review != "" {
			details = append(details, strings.ToLower(strings.ReplaceAll(review, "_", " ")))
		}
		if checks, ok := pr["checks"].(string); ok && checks != "" {
			details = append(details, fmt.Sprintf("checks: %s", checks))
		}
		if len(details) > 0 {
			prLine += fmt.Sprintf(" (%s)", strings.Join(details, ", "))
		}
		lines = append(lines, prLine)
	}

	// Directory snapshot
	files := getDirectorySnapshot(cwd, 15)
	if len(files) > 0 {
		lines = append(lines, fmt.Sprintf("Directory: %s/", filepath.Base(cwd)))
		if len(files) <= 10 {
			lines = append(lines, fmt.Sprintf("  %s", strings.Join(files, ", ")))
		} else {
			lines = append(lines, fmt.Sprintf("  %s", strings.Join(files[:10], ", ")))
			lines = append(lines, fmt.Sprintf("  ... and %d more", len(files)-10))
		}
	}

	lines = append(lines, "</context-refresh>")
	return strings.Join(lines, "\n"), nil
}

func getGitBranch(repoRoot string) string {
	branch, err := git.GetCurrentBranch(repoRoot)
	if err != nil {
		return ""
	}
	return branch
}

func getGitStatusSummary(repoRoot string) string {
	status, err := git.GetStatusSummary(repoRoot)
	if err != nil {
		return ""
	}
	return status
}

func getAllGitBranches(repoRoot string) map[string]string {
	branches := make(map[string]string)
	lines, err := git.RunGitLines(repoRoot, "for-each-ref", "--format=%(refname:short)", "refs/heads/")
	if err != nil {
		return branches
	}

	for _, branchName := range lines {
		// Simplified: mark as unknown (checking cleanliness is expensive)
		branches[branchName] = "unknown"
	}

	return branches
}

func getWorkspaceSessions() []map[string]interface{} {
	// Get repo root to initialize workspace manager
	repoRoot, err := git.GetRepoRoot()
	if err != nil {
		return nil
	}

	manager, err := workspace.NewManagerAt(repoRoot)
	if err != nil {
		return nil
	}

	// List all managed workspaces
	workspaces, err := manager.ListWorkspaces(true)
	if err != nil {
		return nil
	}

	// Convert to map format for compatibility
	var result []map[string]interface{}
	for _, ws := range workspaces {
		isClean, status := ws.IsClean()
		result = append(result, map[string]interface{}{
			"path":   ws.Path,
			"branch": ws.Branch,
			"commit": ws.Commit,
			"status": status,
			"clean":  isClean,
		})
	}

	return result
}

func getCurrentWorkspace(cwd string) map[string]interface{} {
	workspaces := getWorkspaceSessions()
	if len(workspaces) == 0 {
		return nil
	}

	for _, ws := range workspaces {
		if path, ok := ws["path"].(string); ok {
			if strings.HasPrefix(cwd, path) {
				return ws
			}
		}
	}
	return nil
}

func getPRStatus(repoRoot string) map[string]interface{} {
	// Get current branch
	branch, err := git.GetCurrentBranch(repoRoot)
	if err != nil || branch == "" {
		return nil
	}

	// Skip main/master branches
	if branch == "main" || branch == "master" {
		return nil
	}

	// Create GitHub client (uses gh CLI authentication)
	ghClient, err := github.NewClient(repoRoot)
	if err != nil {
		return nil
	}

	ctx := context.Background()

	// Find PR for current branch
	pr, err := ghClient.GetPRForBranch(ctx, branch)
	if err != nil {
		return nil
	}

	result := map[string]interface{}{
		"number": pr.Number,
		"title":  pr.Title,
		"url":    pr.URL,
	}

	if pr.Review != "" {
		result["review"] = pr.Review
	}

	if pr.Checks != "" {
		result["checks"] = pr.Checks
	}

	return result
}

func getDirectorySnapshot(cwd string, maxFiles int) []string {
	var files []string
	priorityPatterns := []string{
		"*.go", "*.py", "*.ts", "*.tsx", "*.js", "*.jsx", "*.rs", "*.rb", "*.java",
		"Makefile", "justfile", "package.json", "pyproject.toml", "Cargo.toml", "go.mod", "Gemfile",
	}

	seen := make(map[string]bool)
	for _, pattern := range priorityPatterns {
		matches, err := filepath.Glob(filepath.Join(cwd, pattern))
		if err != nil {
			continue
		}
		for _, match := range matches {
			if info, err := os.Stat(match); err != nil || info.IsDir() {
				continue
			}
			name := filepath.Base(match)
			if !seen[name] {
				seen[name] = true
				files = append(files, name)
				if len(files) >= maxFiles {
					return files
				}
			}
		}
	}

	// Add some subdirectories if space remains
	if len(files) < maxFiles {
		entries, err := os.ReadDir(cwd)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
					dirName := entry.Name() + "/"
					if !seen[dirName] {
						seen[dirName] = true
						files = append(files, dirName)
						if len(files) >= maxFiles {
							break
						}
					}
				}
			}
		}
	}

	return files
}
