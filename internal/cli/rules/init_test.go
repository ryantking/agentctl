package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitRules(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a fake git repo
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0750); err != nil { //nolint:gosec // Test directory
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Test initialization
	err := InitRules(tmpDir, false, true, false) // force=false, noProject=true, verbose=false
	if err != nil {
		t.Fatalf("InitRules() error = %v", err)
	}

	// Verify .agent/rules/ directory exists
	rulesDir := filepath.Join(tmpDir, ".agent", "rules")
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		t.Error(".agent/rules directory should be created")
	}

	// Verify .agent/research/ directory exists
	researchDir := filepath.Join(tmpDir, ".agent", "research")
	if _, err := os.Stat(researchDir); os.IsNotExist(err) {
		t.Error(".agent/research directory should be created")
	}

	// Verify rules were copied
	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		t.Fatalf("failed to read rules directory: %v", err)
	}

	if len(entries) == 0 {
		t.Error("Rules directory should contain rule files")
	}
}

func TestInitRulesWithForce(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a fake git repo
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0750); err != nil { //nolint:gosec // Test directory
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// First initialization
	err := InitRules(tmpDir, false, true, false)
	if err != nil {
		t.Fatalf("InitRules() error = %v", err)
	}

	// Second initialization with force
	err = InitRules(tmpDir, true, true, false)
	if err != nil {
		t.Fatalf("InitRules() with force error = %v", err)
	}

	// Verify rules directory still exists
	rulesDir := filepath.Join(tmpDir, ".agent", "rules")
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		t.Error("Rules directory should still exist after force init")
	}
}
