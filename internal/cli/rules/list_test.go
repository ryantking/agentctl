package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListRules(t *testing.T) {
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, "rules")

	// Create rules directory
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules directory: %v", err)
	}

	// Create a test rule file
	testRule := `---
name: "Test Rule"
description: "A test rule"
when-to-use: "When testing"
applies-to: ["claude"]
priority: 1
tags: ["test"]
version: "1.0.0"
---

## Test Rule Content
This is a test rule.`

	rulePath := filepath.Join(rulesDir, "test-rule.mdc")
	if err := os.WriteFile(rulePath, []byte(testRule), 0600); err != nil {
		t.Fatalf("failed to write test rule: %v", err)
	}

	// Test listing rules
	rules, err := listRules(rulesDir)
	if err != nil {
		t.Fatalf("listRules() error = %v", err)
	}

	if len(rules) == 0 {
		t.Error("listRules() should return at least one rule")
	}

	found := false
	for _, rule := range rules {
		if rule.Filename == "test-rule.mdc" {
			found = true
			if rule.Metadata.Name != "Test Rule" {
				t.Errorf("Expected name 'Test Rule', got '%s'", rule.Metadata.Name)
			}
			if rule.Metadata.Description != "A test rule" {
				t.Errorf("Expected description 'A test rule', got '%s'", rule.Metadata.Description)
			}
		}
	}

	if !found {
		t.Error("Expected to find test-rule.mdc in list")
	}
}

func TestListRulesEmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, "rules")

	// Create empty rules directory
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules directory: %v", err)
	}

	// Test listing empty directory
	rules, err := listRules(rulesDir)
	if err != nil {
		t.Fatalf("listRules() error = %v", err)
	}

	if len(rules) != 0 {
		t.Errorf("Expected 0 rules, got %d", len(rules))
	}
}

func TestListRulesMissingDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, "nonexistent", "rules")

	// Test listing non-existent directory
	_, err := listRules(rulesDir)
	if err == nil {
		t.Error("listRules() should return error for non-existent directory")
	}
}
