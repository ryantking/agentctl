package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindRuleFile(t *testing.T) {
	tmpDir := t.TempDir()
	rulesDir := filepath.Join(tmpDir, "rules")

	// Create rules directory
	if err := os.MkdirAll(rulesDir, 0755); err != nil { //nolint:gosec // Test directory
		t.Fatalf("failed to create rules directory: %v", err)
	}

	// Create a test rule file
	testRule := `---
name: "Test Rule"
description: "A test rule"
when-to-use: "When testing"
---

## Test Rule Content
This is a test rule.`

	rulePath := filepath.Join(rulesDir, "test-rule.mdc")
	if err := os.WriteFile(rulePath, []byte(testRule), 0600); err != nil {
		t.Fatalf("failed to write test rule: %v", err)
	}

	tests := []struct {
		name     string
		ruleName string
		wantErr  bool
	}{
		{
			name:     "find by filename",
			ruleName: "test-rule.mdc",
			wantErr:  false,
		},
		{
			name:     "find by filename without extension",
			ruleName: "test-rule",
			wantErr:  false,
		},
		{
			name:     "find by rule name",
			ruleName: "Test Rule",
			wantErr:  false,
		},
		{
			name:     "find by rule name case insensitive",
			ruleName: "test rule",
			wantErr:  false,
		},
		{
			name:     "not found",
			ruleName: "nonexistent",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundPath, err := findRuleFile(rulesDir, tt.ruleName)
			if (err != nil) != tt.wantErr {
				t.Errorf("findRuleFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && foundPath != rulePath {
				t.Errorf("findRuleFile() = %v, want %v", foundPath, rulePath)
			}
		})
	}
}

func TestExtractBody(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantBody string
	}{
		{
			name: "normal frontmatter",
			content: `---
name: Test
---

Body content here`,
			wantBody: "Body content here",
		},
		{
			name: "no frontmatter",
			content: "Just body content",
			wantBody: "Just body content",
		},
		{
			name: "empty body",
			content: `---
name: Test
---

`,
			wantBody: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := extractBody(tt.content)
			if strings.TrimSpace(body) != strings.TrimSpace(tt.wantBody) {
				t.Errorf("extractBody() = %v, want %v", body, tt.wantBody)
			}
		})
	}
}
