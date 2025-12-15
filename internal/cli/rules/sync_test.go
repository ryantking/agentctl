package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncToCursor(t *testing.T) {
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, ".agent", "rules")
	cursorRulesDir := filepath.Join(tmpDir, ".cursor", "rules")

	// Create rules directory with a test rule
	if err := os.MkdirAll(rulesDir, 0750); err != nil { //nolint:gosec // Test directory
		t.Fatalf("failed to create rules directory: %v", err)
	}

	testRule := `---
name: "Test Rule"
description: "A test rule"
when-to-use: "When testing"
---

## Content
Test content.`

	rulePath := filepath.Join(rulesDir, "test-rule.mdc")
	if err := os.WriteFile(rulePath, []byte(testRule), 0600); err != nil {
		t.Fatalf("failed to write test rule: %v", err)
	}

	// Test sync
	err := syncToCursor(tmpDir, rulesDir, false)
	if err != nil {
		t.Fatalf("syncToCursor() error = %v", err)
	}

	// Verify file was copied
	destPath := filepath.Join(cursorRulesDir, "test-rule.mdc")
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("Rule file should be copied to .cursor/rules/")
	}

	// Verify content matches
	destData, err := os.ReadFile(destPath) //nolint:gosec // Reading test file
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}

	if string(destData) != testRule {
		t.Error("Copied file content should match original")
	}
}

func TestSyncToClaudeSkills(t *testing.T) {
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, ".agent", "rules")
	claudeSkillsDir := filepath.Join(tmpDir, ".claude", "skills")

	// Create rules directory with a test rule
	if err := os.MkdirAll(rulesDir, 0750); err != nil { //nolint:gosec // Test directory
		t.Fatalf("failed to create rules directory: %v", err)
	}

	testRule := `---
name: "Test Rule"
description: "A test rule"
when-to-use: "When testing"
---

## Content
Test content.`

	rulePath := filepath.Join(rulesDir, "test-rule.mdc")
	if err := os.WriteFile(rulePath, []byte(testRule), 0600); err != nil {
		t.Fatalf("failed to write test rule: %v", err)
	}

	// Test sync
	err := syncToClaudeSkills(tmpDir, rulesDir, false, false)
	if err != nil {
		t.Fatalf("syncToClaudeSkills() error = %v", err)
	}

	// Verify skill directory was created
	skillDir := filepath.Join(claudeSkillsDir, "test-rule")
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		t.Error("Skill directory should be created")
	}

	// Verify SKILL.md exists
	skillMDPath := filepath.Join(skillDir, "SKILL.md")
	if _, err := os.Stat(skillMDPath); os.IsNotExist(err) {
		t.Error("SKILL.md should be created")
	}

	// Verify content
	skillData, err := os.ReadFile(skillMDPath) //nolint:gosec // Reading test file
	if err != nil {
		t.Fatalf("failed to read skill file: %v", err)
	}

	skillContent := string(skillData)
	if !strings.Contains(skillContent, "name: Test Rule") {
		t.Error("Skill file should contain rule name")
	}
	if !strings.Contains(skillContent, "description: A test rule") {
		t.Error("Skill file should contain rule description")
	}
	if !strings.Contains(skillContent, "## Content") {
		t.Error("Skill file should contain rule body content")
	}
}

func TestSyncToClaudeSkillsSkipsExisting(t *testing.T) {
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, ".agent", "rules")
	claudeSkillsDir := filepath.Join(tmpDir, ".claude", "skills")

	// Create rules directory with a test rule
	if err := os.MkdirAll(rulesDir, 0750); err != nil { //nolint:gosec // Test directory
		t.Fatalf("failed to create rules directory: %v", err)
	}

	testRule := `---
name: "Test Rule"
description: "A test rule"
when-to-use: "When testing"
---

## Content
Test content.`

	rulePath := filepath.Join(rulesDir, "test-rule.mdc")
	if err := os.WriteFile(rulePath, []byte(testRule), 0600); err != nil {
		t.Fatalf("failed to write test rule: %v", err)
	}

	// Create existing skill with manual modifications
	skillDir := filepath.Join(claudeSkillsDir, "test-rule")
	if err := os.MkdirAll(skillDir, 0750); err != nil { //nolint:gosec // Test directory
		t.Fatalf("failed to create skill directory: %v", err)
	}

	existingSkill := `---
name: Manual Skill
description: Manually created skill
---

## Manual Content
This was manually created.`

	skillMDPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillMDPath, []byte(existingSkill), 0600); err != nil {
		t.Fatalf("failed to write existing skill: %v", err)
	}

	// Test sync without force (should skip)
	err := syncToClaudeSkills(tmpDir, rulesDir, false, false)
	if err != nil {
		t.Fatalf("syncToClaudeSkills() error = %v", err)
	}

	// Verify existing skill was not overwritten
	skillData, err := os.ReadFile(skillMDPath) //nolint:gosec // Reading test file
	if err != nil {
		t.Fatalf("failed to read skill file: %v", err)
	}

	if string(skillData) != existingSkill {
		t.Error("Existing skill should not be overwritten without --force")
	}

	// Test sync with force (should overwrite)
	err = syncToClaudeSkills(tmpDir, rulesDir, true, false)
	if err != nil {
		t.Fatalf("syncToClaudeSkills() with force error = %v", err)
	}

	// Verify skill was overwritten
	skillData, err = os.ReadFile(skillMDPath) //nolint:gosec // Reading test file
	if err != nil {
		t.Fatalf("failed to read skill file: %v", err)
	}

	if strings.Contains(string(skillData), "Manual Skill") {
		t.Error("Existing skill should be overwritten with --force")
	}
	if !strings.Contains(string(skillData), "Test Rule") {
		t.Error("Skill should contain new rule content after force overwrite")
	}
}

func TestSyncToAGENTSMD(t *testing.T) {
	tmpDir := t.TempDir()
	agentDir := filepath.Join(tmpDir, ".agent")
	rulesDir := filepath.Join(agentDir, "rules")

	// Create rules directory with a test rule
	if err := os.MkdirAll(rulesDir, 0750); err != nil { //nolint:gosec // Test directory
		t.Fatalf("failed to create rules directory: %v", err)
	}

	testRule := `---
name: "Test Rule"
description: "A test rule"
when-to-use: "When testing"
---

## Content
Test content.`

	rulePath := filepath.Join(rulesDir, "test-rule.mdc")
	if err := os.WriteFile(rulePath, []byte(testRule), 0600); err != nil {
		t.Fatalf("failed to write test rule: %v", err)
	}

	// Test sync
	err := syncToAGENTSMD(tmpDir, agentDir, rulesDir, false)
	if err != nil {
		t.Fatalf("syncToAGENTSMD() error = %v", err)
	}

	// Verify AGENTS.md was created
	agentsMDPath := filepath.Join(tmpDir, "AGENTS.md")
	if _, err := os.Stat(agentsMDPath); os.IsNotExist(err) {
		t.Error("AGENTS.md should be created")
	}

	// Verify content
	agentsData, err := os.ReadFile(agentsMDPath) //nolint:gosec // Reading test file
	if err != nil {
		t.Fatalf("failed to read AGENTS.md: %v", err)
	}

	agentsContent := string(agentsData)
	if !strings.Contains(agentsContent, "Test Rule") {
		t.Error("AGENTS.md should contain rule name")
	}
	if !strings.Contains(agentsContent, "A test rule") {
		t.Error("AGENTS.md should contain rule description")
	}
	if !strings.Contains(agentsContent, "When testing") {
		t.Error("AGENTS.md should contain when-to-use")
	}
}

func TestSyncToAGENTSMDWithoutProjectMD(t *testing.T) {
	tmpDir := t.TempDir()
	agentDir := filepath.Join(tmpDir, ".agent")
	rulesDir := filepath.Join(agentDir, "rules")

	// Create rules directory with a test rule (but no project.md)
	if err := os.MkdirAll(rulesDir, 0750); err != nil { //nolint:gosec // Test directory
		t.Fatalf("failed to create rules directory: %v", err)
	}

	testRule := `---
name: "Test Rule"
description: "A test rule"
when-to-use: "When testing"
---

## Content
Test content.`

	rulePath := filepath.Join(rulesDir, "test-rule.mdc")
	if err := os.WriteFile(rulePath, []byte(testRule), 0600); err != nil {
		t.Fatalf("failed to write test rule: %v", err)
	}

	// Test sync without project.md
	err := syncToAGENTSMD(tmpDir, agentDir, rulesDir)
	if err != nil {
		t.Fatalf("syncToAGENTSMD() error = %v", err)
	}

	// Verify AGENTS.md was created
	agentsMDPath := filepath.Join(tmpDir, "AGENTS.md")
	if _, err := os.Stat(agentsMDPath); os.IsNotExist(err) {
		t.Error("AGENTS.md should be created even without project.md")
	}

	// Verify content doesn't have empty sections
	agentsData, err := os.ReadFile(agentsMDPath) //nolint:gosec // Reading test file
	if err != nil {
		t.Fatalf("failed to read AGENTS.md: %v", err)
	}

	agentsContent := string(agentsData)
	if !strings.Contains(agentsContent, "Test Rule") {
		t.Error("AGENTS.md should contain rule name")
	}
	// Should not have empty project content sections
	if strings.Contains(agentsContent, "\n\n\n") {
		t.Error("AGENTS.md should not have empty sections")
	}
}

func TestSyncToCLAUDEMDWithoutProjectMD(t *testing.T) {
	tmpDir := t.TempDir()
	agentDir := filepath.Join(tmpDir, ".agent")
	rulesDir := filepath.Join(agentDir, "rules")

	// Create rules directory with a test rule (but no project.md)
	if err := os.MkdirAll(rulesDir, 0750); err != nil { //nolint:gosec // Test directory
		t.Fatalf("failed to create rules directory: %v", err)
	}

	testRule := `---
name: "Test Rule"
description: "A test rule"
when-to-use: "When testing"
---

## Content
Test content.`

	rulePath := filepath.Join(rulesDir, "test-rule.mdc")
	if err := os.WriteFile(rulePath, []byte(testRule), 0600); err != nil {
		t.Fatalf("failed to write test rule: %v", err)
	}

	// Test sync without project.md
	err := syncToCLAUDEMD(tmpDir, agentDir, rulesDir, false)
	if err != nil {
		t.Fatalf("syncToCLAUDEMD() error = %v", err)
	}

	// Verify CLAUDE.md was created
	claudeMDPath := filepath.Join(tmpDir, "CLAUDE.md")
	if _, err := os.Stat(claudeMDPath); os.IsNotExist(err) {
		t.Error("CLAUDE.md should be created even without project.md")
	}

	// Verify content doesn't have empty sections
	claudeData, err := os.ReadFile(claudeMDPath) //nolint:gosec // Reading test file
	if err != nil {
		t.Fatalf("failed to read CLAUDE.md: %v", err)
	}

	claudeContent := string(claudeData)
	if !strings.Contains(claudeContent, "Test Rule") {
		t.Error("CLAUDE.md should contain rule name")
	}
	// Should not have empty project content sections
	if strings.Contains(claudeContent, "\n\n\n") {
		t.Error("CLAUDE.md should not have empty sections")
	}
}

func TestSanitizeSkillName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     string
	}{
		{
			name:  "simple name",
			input: "Git Workflow",
			want:  "git-workflow",
		},
		{
			name:  "with special chars",
			input: "Tool Selection Guidelines!",
			want:  "tool-selection-guidelines",
		},
		{
			name:  "already lowercase",
			input: "test-rule",
			want:  "test-rule",
		},
		{
			name:  "with numbers",
			input: "Rule 123",
			want:  "rule-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeSkillName(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeSkillName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
