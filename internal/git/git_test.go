package git

import (
	"os"
	"testing"
)

func TestGetRepoRoot(t *testing.T) {
	// This test requires a git repository
	repoRoot, err := GetRepoRoot()
	if err != nil {
		t.Skip("Not in a git repository, skipping test")
	}
	if repoRoot == "" {
		t.Error("GetRepoRoot returned empty string")
	}
	if _, err := os.Stat(repoRoot); os.IsNotExist(err) {
		t.Errorf("GetRepoRoot returned non-existent path: %s", repoRoot)
	}
}

func TestGetRepoName(t *testing.T) {
	repoName, err := GetRepoName()
	if err != nil {
		t.Skip("Not in a git repository, skipping test")
	}
	if repoName == "" {
		t.Error("GetRepoName returned empty string")
	}
}

func TestBranchExists(t *testing.T) {
	repoRoot, err := GetRepoRoot()
	if err != nil {
		t.Skip("Not in a git repository, skipping test")
	}

	// Test with a branch that should exist
	exists, err := BranchExists(repoRoot, "main")
	if err != nil {
		t.Fatalf("BranchExists failed: %v", err)
	}
	// main branch should exist in most repos
	if !exists {
		// Try master as fallback
		exists, err = BranchExists(repoRoot, "master")
		if err != nil {
			t.Fatalf("BranchExists failed: %v", err)
		}
		if !exists {
			t.Log("Neither main nor master branch found, this might be a new repo")
		}
	}

	// Test with a branch that shouldn't exist
	exists, err = BranchExists(repoRoot, "nonexistent-branch-12345")
	if err != nil {
		t.Fatalf("BranchExists failed: %v", err)
	}
	if exists {
		t.Error("BranchExists returned true for nonexistent branch")
	}
}

// TestParseWorktreeList removed - worktree parsing now done via filesystem
// reading .git/worktrees directory structure instead of parsing CLI output

func TestIsWorktreeClean(t *testing.T) {
	repoRoot, err := GetRepoRoot()
	if err != nil {
		t.Skip("Not in a git repository, skipping test")
	}

	isClean, status := IsWorktreeClean(repoRoot)
	_ = isClean // We can't predict if it's clean or not
	if status == "" {
		t.Error("IsWorktreeClean returned empty status")
	}
}
