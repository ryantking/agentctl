package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatePrompt(t *testing.T) {
	tests := []struct {
		name    string
		prompt  string
		wantErr bool
	}{
		{
			name:    "valid prompt",
			prompt:  "Always use conventional commits",
			wantErr: false,
		},
		{
			name:    "too short",
			prompt:  "short",
			wantErr: true,
		},
		{
			name:    "exactly minimum length",
			prompt:  "1234567890", // 10 chars
			wantErr: false,
		},
		{
			name:    "one char less than minimum",
			prompt:  "123456789", // 9 chars
			wantErr: true,
		},
		{
			name:    "too long",
			prompt:  strings.Repeat("a", 1001),
			wantErr: true,
		},
		{
			name:    "exactly maximum length",
			prompt:  strings.Repeat("a", 1000),
			wantErr: false,
		},
		{
			name:    "whitespace only",
			prompt:  "          ",
			wantErr: true,
		},
		{
			name:    "whitespace trimmed",
			prompt:  "   valid prompt   ",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePrompt(tt.prompt)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePrompt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRuleName(t *testing.T) {
	tests := []struct {
		name    string
		ruleName string
		wantErr bool
	}{
		{
			name:    "valid name",
			ruleName: "git-workflow",
			wantErr: false,
		},
		{
			name:    "valid with numbers",
			ruleName: "rule-123",
			wantErr: false,
		},
		{
			name:    "empty name",
			ruleName: "",
			wantErr: true,
		},
		{
			name:    "too long",
			ruleName: strings.Repeat("a", 51),
			wantErr: true,
		},
		{
			name:    "exactly maximum length",
			ruleName: strings.Repeat("a", 50),
			wantErr: false,
		},
		{
			name:    "uppercase letters",
			ruleName: "Git-Workflow",
			wantErr: true,
		},
		{
			name:    "underscores",
			ruleName: "git_workflow",
			wantErr: true,
		},
		{
			name:    "spaces",
			ruleName: "git workflow",
			wantErr: true,
		},
		{
			name:    "special characters",
			ruleName: "git@workflow",
			wantErr: true,
		},
		{
			name:    "starts with hyphen",
			ruleName: "-git-workflow",
			wantErr: false, // Current validation allows this
		},
		{
			name:    "ends with hyphen",
			ruleName: "git-workflow-",
			wantErr: false, // Current validation allows this
		},
		{
			name:    "multiple hyphens",
			ruleName: "git--workflow",
			wantErr: false, // Allowed, though not ideal
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRuleName(tt.ruleName)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRuleName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDescription(t *testing.T) {
	tests := []struct {
		name        string
		description string
		wantErr     bool
	}{
		{
			name:        "valid description",
			description: "A valid description",
			wantErr:     false,
		},
		{
			name:        "empty",
			description: "",
			wantErr:     true,
		},
		{
			name:        "whitespace only",
			description: "   ",
			wantErr:     true,
		},
		{
			name:        "too long",
			description: strings.Repeat("a", 201),
			wantErr:     true,
		},
		{
			name:        "exactly maximum length",
			description: strings.Repeat("a", 200),
			wantErr:     false,
		},
		{
			name:        "whitespace trimmed",
			description: "   valid description   ",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDescription(tt.description)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDescription() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}


func TestValidateAppliesTo(t *testing.T) {
	tests := []struct {
		name    string
		tools   []string
		wantErr bool
	}{
		{
			name:    "valid known tool",
			tools:   []string{"claude"},
			wantErr: false,
		},
		{
			name:    "multiple known tools",
			tools:   []string{"claude", "cursor", "windsurf"},
			wantErr: false,
		},
		{
			name:    "unknown tool",
			tools:   []string{"unknown-tool"},
			wantErr: false, // Warning only, not error
		},
		{
			name:    "mixed known and unknown",
			tools:   []string{"claude", "unknown-tool"},
			wantErr: false, // Warning only
		},
		{
			name:    "empty tool",
			tools:   []string{""},
			wantErr: true,
		},
		{
			name:    "whitespace tool",
			tools:   []string{"  "},
			wantErr: true,
		},
		{
			name:    "case insensitive",
			tools:   []string{"CLAUDE", "Cursor"},
			wantErr: false,
		},
		{
			name:    "empty list",
			tools:   []string{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAppliesTo(tt.tools)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAppliesTo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckNameConflict(t *testing.T) {
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, ".agent", "rules")
	//nolint:gosec // Test directory permissions are acceptable for temporary test directories
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatalf("failed to create rules directory: %v", err)
	}

	// Create existing rule
	existingRule := filepath.Join(rulesDir, "existing-rule.mdc")
	//nolint:gosec // Test file permissions are acceptable for temporary test files
	if err := os.WriteFile(existingRule, []byte("---\nname: Existing Rule\n---\n"), 0644); err != nil {
		t.Fatalf("failed to create existing rule: %v", err)
	}

	tests := []struct {
		name    string
		ruleName string
		wantErr bool
	}{
		{
			name:    "no conflict",
			ruleName: "new-rule",
			wantErr: false,
		},
		{
			name:    "conflict with existing",
			ruleName: "existing-rule",
			wantErr: true,
		},
		{
			name:    "case insensitive conflict",
			ruleName: "EXISTING-RULE",
			wantErr: true,
		},
		{
			name:    "no conflict different name",
			ruleName: "different-rule",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkNameConflict(rulesDir, tt.ruleName)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkNameConflict() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckNameConflictNoDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, ".agent", "rules")

	// Directory doesn't exist - should not error
	err := checkNameConflict(rulesDir, "any-rule")
	if err != nil {
		t.Errorf("checkNameConflict() with non-existent directory should not error, got: %v", err)
	}
}
