package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFullRulesWorkflow tests the complete rules workflow from init to sync.
func TestFullRulesWorkflow(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a fake git repo
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0750); err != nil { //nolint:gosec // Test directory
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Step 1: Initialize rules
	err := InitRules(tmpDir, false, true) // force=false, noProject=true
	if err != nil {
		t.Fatalf("InitRules() error = %v", err)
	}

	// Verify .agent/rules/ directory exists
	rulesDir := filepath.Join(tmpDir, ".agent", "rules")
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		t.Fatal(".agent/rules directory should be created")
	}

	// Step 2: List rules
	rules, err := listRules(rulesDir)
	if err != nil {
		t.Fatalf("listRules() error = %v", err)
	}

	if len(rules) == 0 {
		t.Fatal("Should have default rules after init")
	}

	// Step 3: Show a rule
	if len(rules) > 0 {
		rulePath, err := findRuleFile(rulesDir, rules[0].Metadata.Name)
		if err != nil {
			t.Fatalf("findRuleFile() error = %v", err)
		}

		data, err := os.ReadFile(rulePath) //nolint:gosec // Reading test file
		if err != nil {
			t.Fatalf("failed to read rule file: %v", err)
		}

		if len(data) == 0 {
			t.Error("Rule file should have content")
		}
	}

	// Step 4: Add a new rule
	testRule := `---
name: "Integration Test Rule"
description: "A rule created during integration test"
when-to-use: "When testing integration"
applies-to: ["claude"]
priority: 2
tags: ["test", "integration"]
version: "1.0.0"
---

## Integration Test Rule

This rule was created during integration testing.
`

	newRulePath := filepath.Join(rulesDir, "integration-test.mdc")
	if err := os.WriteFile(newRulePath, []byte(testRule), 0600); err != nil {
		t.Fatalf("failed to write new rule: %v", err)
	}

	// Step 5: Verify new rule appears in list
	rules, err = listRules(rulesDir)
	if err != nil {
		t.Fatalf("listRules() after add error = %v", err)
	}

	found := false
	for _, rule := range rules {
		if rule.Filename == "integration-test.mdc" {
			found = true
			if rule.Metadata.Name != "Integration Test Rule" {
				t.Errorf("Expected name 'Integration Test Rule', got '%s'", rule.Metadata.Name)
			}
		}
	}

	if !found {
		t.Error("New rule should appear in list")
	}

	// Step 6: Sync to AGENTS.md
	agentDir := filepath.Join(tmpDir, ".agent")
	err = syncToAGENTSMD(tmpDir, agentDir, rulesDir)
	if err != nil {
		t.Fatalf("syncToAGENTSMD() error = %v", err)
	}

	// Verify AGENTS.md was created
	agentsMDPath := filepath.Join(tmpDir, "AGENTS.md")
	if _, err := os.Stat(agentsMDPath); os.IsNotExist(err) {
		t.Fatal("AGENTS.md should be created")
	}

	// Verify content includes rules
	agentsData, err := os.ReadFile(agentsMDPath) //nolint:gosec // Reading test file
	if err != nil {
		t.Fatalf("failed to read AGENTS.md: %v", err)
	}

	agentsContent := string(agentsData)
	if !strings.Contains(agentsContent, "Integration Test Rule") {
		t.Error("AGENTS.md should contain new rule")
	}
	if !strings.Contains(agentsContent, "Priority 2 (Medium)") {
		t.Error("AGENTS.md should contain priority grouping")
	}

	// Step 7: Sync to CLAUDE.md
	err = syncToCLAUDEMD(tmpDir, agentDir, rulesDir)
	if err != nil {
		t.Fatalf("syncToCLAUDEMD() error = %v", err)
	}

	// Verify CLAUDE.md was created
	claudeMDPath := filepath.Join(tmpDir, "CLAUDE.md")
	if _, err := os.Stat(claudeMDPath); os.IsNotExist(err) {
		t.Fatal("CLAUDE.md should be created")
	}

	// Verify content includes skills
	claudeData, err := os.ReadFile(claudeMDPath) //nolint:gosec // Reading test file
	if err != nil {
		t.Fatalf("failed to read CLAUDE.md: %v", err)
	}

	claudeContent := string(claudeData)
	if !strings.Contains(claudeContent, "Integration Test Rule") {
		t.Error("CLAUDE.md should contain new rule as skill")
	}

	// Step 8: Sync to Claude skills
	err = syncToClaudeSkills(tmpDir, rulesDir, false)
	if err != nil {
		t.Fatalf("syncToClaudeSkills() error = %v", err)
	}

	// Verify skill was created
	skillDir := filepath.Join(tmpDir, ".claude", "skills", "integration-test-rule")
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		t.Error("Skill directory should be created")
	}

	skillMDPath := filepath.Join(skillDir, "SKILL.md")
	if _, err := os.Stat(skillMDPath); os.IsNotExist(err) {
		t.Error("SKILL.md should be created")
	}

	// Step 9: Sync to Cursor
	err = syncToCursor(tmpDir, rulesDir)
	if err != nil {
		t.Fatalf("syncToCursor() error = %v", err)
	}

	// Verify rule was copied to Cursor
	cursorRulePath := filepath.Join(tmpDir, ".cursor", "rules", "integration-test.mdc")
	if _, err := os.Stat(cursorRulePath); os.IsNotExist(err) {
		t.Error("Rule should be copied to .cursor/rules/")
	}

	// Step 10: Remove rule
	if err := os.Remove(newRulePath); err != nil {
		t.Fatalf("failed to remove rule: %v", err)
	}

	// Verify rule no longer appears in list
	rules, err = listRules(rulesDir)
	if err != nil {
		t.Fatalf("listRules() after remove error = %v", err)
	}

	for _, rule := range rules {
		if rule.Filename == "integration-test.mdc" {
			t.Error("Removed rule should not appear in list")
		}
	}
}
