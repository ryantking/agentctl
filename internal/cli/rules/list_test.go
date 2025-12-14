package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListRules(t *testing.T) {
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, "rules")

	// Create rules directory
	if err := os.MkdirAll(rulesDir, 0750); err != nil { //nolint:gosec // Test directory
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
	if err := os.MkdirAll(rulesDir, 0750); err != nil { //nolint:gosec // Test directory
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

func TestValidateRuleMetadata(t *testing.T) {
	tests := []struct {
		name      string
		metadata  RuleMetadata
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid metadata",
			metadata: RuleMetadata{
				Name:        "Test Rule",
				Description: "A test rule",
				WhenToUse:   "When testing",
				Priority:    1,
				Version:     "1.0.0",
			},
			wantError: false,
		},
		{
			name: "missing name",
			metadata: RuleMetadata{
				Description: "A test rule",
				WhenToUse:   "When testing",
			},
			wantError: true,
			errorMsg:  "required field 'name' is empty",
		},
		{
			name: "missing description",
			metadata: RuleMetadata{
				Name:      "Test Rule",
				WhenToUse: "When testing",
			},
			wantError: true,
			errorMsg:  "required field 'description' is empty",
		},
		{
			name: "missing when-to-use",
			metadata: RuleMetadata{
				Name:        "Test Rule",
				Description: "A test rule",
			},
			wantError: true,
			errorMsg:  "required field 'when-to-use' is empty",
		},
		{
			name: "priority too high",
			metadata: RuleMetadata{
				Name:        "Test Rule",
				Description: "A test rule",
				WhenToUse:   "When testing",
				Priority:    5,
			},
			wantError: true,
			errorMsg:  "priority must be 0-4",
		},
		{
			name: "priority negative",
			metadata: RuleMetadata{
				Name:        "Test Rule",
				Description: "A test rule",
				WhenToUse:   "When testing",
				Priority:    -1,
			},
			wantError: true,
			errorMsg:  "priority must be 0-4",
		},
		{
			name: "invalid semver",
			metadata: RuleMetadata{
				Name:        "Test Rule",
				Description: "A test rule",
				WhenToUse:   "When testing",
				Version:     "invalid",
			},
			wantError: true,
			errorMsg:  "version should follow semver format",
		},
		{
			name: "valid semver with v prefix",
			metadata: RuleMetadata{
				Name:        "Test Rule",
				Description: "A test rule",
				WhenToUse:   "When testing",
				Version:     "v1.0.0",
			},
			wantError: false,
		},
		{
			name: "empty version is valid",
			metadata: RuleMetadata{
				Name:        "Test Rule",
				Description: "A test rule",
				WhenToUse:   "When testing",
				Version:     "",
			},
			wantError: false,
		},
		{
			name: "whitespace-only name",
			metadata: RuleMetadata{
				Name:        "   ",
				Description: "A test rule",
				WhenToUse:   "When testing",
			},
			wantError: true,
			errorMsg:  "required field 'name' is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRuleMetadata(tt.metadata, "test.mdc")
			if (err != nil) != tt.wantError {
				t.Errorf("validateRuleMetadata() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.wantError && tt.errorMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("validateRuleMetadata() error = %v, want error containing %q", err, tt.errorMsg)
				}
			}
		})
	}
}

func TestIsValidSemver(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		want     bool
	}{
		{
			name:    "valid semver",
			version: "1.0.0",
			want:    true,
		},
		{
			name:    "valid semver with v prefix",
			version: "v1.0.0",
			want:    true,
		},
		{
			name:    "valid semver with patch",
			version: "1.2.3",
			want:    true,
		},
		{
			name:    "invalid: not enough parts",
			version: "1.0",
			want:    false,
		},
		{
			name:    "invalid: too many parts",
			version: "1.0.0.0",
			want:    false,
		},
		{
			name:    "invalid: non-numeric",
			version: "a.b.c",
			want:    false,
		},
		{
			name:    "invalid: empty",
			version: "",
			want:    false,
		},
		{
			name:    "valid: with pre-release",
			version: "1.0.0-alpha",
			want:    true, // Basic validation allows this
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidSemver(tt.version)
			if got != tt.want {
				t.Errorf("isValidSemver(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestListRulesSkipsInvalidMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, "rules")

	// Create rules directory
	if err := os.MkdirAll(rulesDir, 0750); err != nil { //nolint:gosec // Test directory
		t.Fatalf("failed to create rules directory: %v", err)
	}

	// Create a valid rule
	validRule := `---
name: "Valid Rule"
description: "A valid rule"
when-to-use: "When valid"
---

Content.`

	validPath := filepath.Join(rulesDir, "valid.mdc")
	if err := os.WriteFile(validPath, []byte(validRule), 0600); err != nil {
		t.Fatalf("failed to write valid rule: %v", err)
	}

	// Create an invalid rule (missing required field)
	invalidRule := `---
name: ""
description: "Missing name"
when-to-use: "When invalid"
---

Content.`

	invalidPath := filepath.Join(rulesDir, "invalid.mdc")
	if err := os.WriteFile(invalidPath, []byte(invalidRule), 0600); err != nil {
		t.Fatalf("failed to write invalid rule: %v", err)
	}

	// Test listing rules - should skip invalid one
	rules, err := listRules(rulesDir)
	if err != nil {
		t.Fatalf("listRules() error = %v", err)
	}

	// Should only find the valid rule
	if len(rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(rules))
	}

	if len(rules) > 0 && rules[0].Filename != "valid.mdc" {
		t.Errorf("Expected 'valid.mdc', got '%s'", rules[0].Filename)
	}
}

func TestValidateAgentDir(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		agentDir  string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "empty string",
			agentDir:  "",
			wantError: true,
			errorMsg:  "cannot be empty",
		},
		{
			name:      "valid relative path",
			agentDir:  ".agent",
			wantError: false,
		},
		{
			name:      "valid custom name",
			agentDir:  ".custom-agent",
			wantError: false,
		},
		{
			name:      "contains ..",
			agentDir:  "../.agent",
			wantError: true,
			errorMsg:  "invalid character sequence",
		},
		{
			name:      "contains ~",
			agentDir:  "~/.agent",
			wantError: true,
			errorMsg:  "invalid character sequence",
		},
		{
			name:      "valid absolute path",
			agentDir:  tmpDir,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAgentDir(tt.agentDir)
			if (err != nil) != tt.wantError {
				t.Errorf("validateAgentDir(%q) error = %v, wantError %v", tt.agentDir, err, tt.wantError)
				return
			}
			if tt.wantError && tt.errorMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("validateAgentDir(%q) error = %v, want error containing %q", tt.agentDir, err, tt.errorMsg)
				}
			}
		})
	}
}

func TestGetAgentDir(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		envValue  string
		repoRoot  string
		wantError bool
	}{
		{
			name:      "default .agent",
			envValue:  "",
			repoRoot:  tmpDir,
			wantError: false,
		},
		{
			name:      "custom relative path",
			envValue:  ".custom",
			repoRoot:  tmpDir,
			wantError: false,
		},
		{
			name:      "invalid path with ..",
			envValue:  "../.agent",
			repoRoot:  tmpDir,
			wantError: true,
		},
		{
			name:      "valid absolute path",
			envValue:  tmpDir,
			repoRoot:  "/some/repo",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		if tt.envValue != "" {
			t.Setenv("AGENTDIR", tt.envValue)
		} else {
			_ = os.Unsetenv("AGENTDIR") //nolint:errcheck // Test cleanup
		}

			agentDir, err := getAgentDir(tt.repoRoot)
			if (err != nil) != tt.wantError {
				t.Errorf("getAgentDir() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if !tt.wantError && agentDir == "" {
				t.Error("getAgentDir() should return non-empty path")
			}
		})
	}
}
