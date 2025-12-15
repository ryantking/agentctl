package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestFullRulesWorkflow tests the complete rules workflow from init to sync.
func TestFullRulesWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	setupTestRepo(t, tmpDir)

	rulesDir := filepath.Join(tmpDir, ".agent", "rules")
	agentDir := filepath.Join(tmpDir, ".agent")

	// Step 1: Initialize rules
	initRules(t, tmpDir)

	// Step 2: List and verify rules
	rules := listAndVerifyRules(t, rulesDir)

	// Step 3: Show a rule
	showRule(t, rulesDir, rules)

	// Step 4: Add a new rule
	newRulePath := addTestRule(t, rulesDir)

	// Step 5: Verify new rule appears
	verifyRuleInList(t, rulesDir, "integration-test.mdc", "Integration Test Rule")

	// Step 6: Sync to AGENTS.md
	verifyAGENTSMD(t, tmpDir, agentDir, rulesDir)

	// Step 7: Sync to CLAUDE.md
	verifyCLAUDEMD(t, tmpDir, agentDir, rulesDir)

	// Step 8: Sync to Claude skills
	verifyClaudeSkills(t, tmpDir, rulesDir)

	// Step 9: Sync to Cursor
	verifyCursorSync(t, tmpDir, rulesDir)

	// Step 10: Remove rule
	removeRule(t, newRulePath, rulesDir)
}

func setupTestRepo(t *testing.T, tmpDir string) {
	t.Helper()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0750); err != nil { //nolint:gosec // Test directory
		t.Fatalf("failed to create .git directory: %v", err)
	}
}

func initRules(t *testing.T, tmpDir string) {
	t.Helper()
	cmd := &cobra.Command{}
	cmd.Flags().String("agent-cli", "claude", "")
	err := InitRules(cmd, tmpDir, false, true, false) // force=false, noProject=true, verbose=false
	if err != nil {
		t.Fatalf("InitRules() error = %v", err)
	}

	rulesDir := filepath.Join(tmpDir, ".agent", "rules")
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		t.Fatal(".agent/rules directory should be created")
	}
}

func listAndVerifyRules(t *testing.T, rulesDir string) []RuleInfo {
	t.Helper()
	rules, err := listRules(rulesDir)
	if err != nil {
		t.Fatalf("listRules() error = %v", err)
	}

	if len(rules) == 0 {
		t.Fatal("Should have default rules after init")
	}

	return rules
}

func showRule(t *testing.T, rulesDir string, rules []RuleInfo) {
	t.Helper()
	if len(rules) == 0 {
		return
	}

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

func addTestRule(t *testing.T, rulesDir string) string {
	t.Helper()
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

	return newRulePath
}

func verifyRuleInList(t *testing.T, rulesDir, filename, expectedName string) {
	t.Helper()
	rules, err := listRules(rulesDir)
	if err != nil {
		t.Fatalf("listRules() error = %v", err)
	}

	found := false
	for _, rule := range rules {
		if rule.Filename == filename {
			found = true
			if rule.Metadata.Name != expectedName {
				t.Errorf("Expected name %q, got %q", expectedName, rule.Metadata.Name)
			}
		}
	}

	if !found {
		t.Errorf("Rule %s should appear in list", filename)
	}
}

func verifyAGENTSMD(t *testing.T, tmpDir, agentDir, rulesDir string) {
	t.Helper()
	err := syncToAGENTSMD(tmpDir, agentDir, rulesDir, false)
	if err != nil {
		t.Fatalf("syncToAGENTSMD() error = %v", err)
	}

	agentsMDPath := filepath.Join(tmpDir, "AGENTS.md")
	if _, err := os.Stat(agentsMDPath); os.IsNotExist(err) {
		t.Fatal("AGENTS.md should be created")
	}

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
}

func verifyCLAUDEMD(t *testing.T, tmpDir, agentDir, rulesDir string) {
	t.Helper()
	err := syncToCLAUDEMD(tmpDir, agentDir, rulesDir, false)
	if err != nil {
		t.Fatalf("syncToCLAUDEMD() error = %v", err)
	}

	claudeMDPath := filepath.Join(tmpDir, "CLAUDE.md")
	if _, err := os.Stat(claudeMDPath); os.IsNotExist(err) {
		t.Fatal("CLAUDE.md should be created")
	}

	claudeData, err := os.ReadFile(claudeMDPath) //nolint:gosec // Reading test file
	if err != nil {
		t.Fatalf("failed to read CLAUDE.md: %v", err)
	}

	claudeContent := string(claudeData)
	if !strings.Contains(claudeContent, "Integration Test Rule") {
		t.Error("CLAUDE.md should contain new rule as skill")
	}
}

func verifyClaudeSkills(t *testing.T, tmpDir, rulesDir string) {
	t.Helper()
	err := syncToClaudeSkills(tmpDir, rulesDir, false, false)
	if err != nil {
		t.Fatalf("syncToClaudeSkills() error = %v", err)
	}

	skillDir := filepath.Join(tmpDir, ".claude", "skills", "integration-test-rule")
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		t.Error("Skill directory should be created")
	}

	skillMDPath := filepath.Join(skillDir, "SKILL.md")
	if _, err := os.Stat(skillMDPath); os.IsNotExist(err) {
		t.Error("SKILL.md should be created")
	}
}

func verifyCursorSync(t *testing.T, tmpDir, rulesDir string) {
	t.Helper()
	err := syncToCursor(tmpDir, rulesDir, false)
	if err != nil {
		t.Fatalf("syncToCursor() error = %v", err)
	}

	cursorRulePath := filepath.Join(tmpDir, ".cursor", "rules", "integration-test.mdc")
	if _, err := os.Stat(cursorRulePath); os.IsNotExist(err) {
		t.Error("Rule should be copied to .cursor/rules/")
	}
}

func removeRule(t *testing.T, rulePath, rulesDir string) {
	t.Helper()
	if err := os.Remove(rulePath); err != nil {
		t.Fatalf("failed to remove rule: %v", err)
	}

	rules, err := listRules(rulesDir)
	if err != nil {
		t.Fatalf("listRules() after remove error = %v", err)
	}

	for _, rule := range rules {
		if rule.Filename == "integration-test.mdc" {
			t.Error("Removed rule should not appear in list")
		}
	}
}
