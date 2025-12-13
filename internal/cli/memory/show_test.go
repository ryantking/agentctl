package memory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveImports(t *testing.T) {
	tmpDir := t.TempDir()

	// Create AGENTS.md
	agentsContent := `# Agents
Rule 1: Always verify
Rule 2: No apologies`
	agentsPath := filepath.Join(tmpDir, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte(agentsContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Create CLAUDE.md with import
	claudeContent := `# Claude Code
@AGENTS.md

## Agent Orchestration
[Claude-specific content]`
	claudePath := filepath.Join(tmpDir, "CLAUDE.md")
	if err := os.WriteFile(claudePath, []byte(claudeContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Test import resolution
	resolved := resolveImports(claudeContent, tmpDir)

	// Should contain content from AGENTS.md
	if !strings.Contains(resolved, "Rule 1: Always verify") {
		t.Error("Resolved content should include AGENTS.md content")
	}
	if !strings.Contains(resolved, "Rule 2: No apologies") {
		t.Error("Resolved content should include AGENTS.md content")
	}
	// Should contain original CLAUDE.md content
	if !strings.Contains(resolved, "Agent Orchestration") {
		t.Error("Resolved content should include CLAUDE.md content")
	}
}

func TestExtractImportsFromContent(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantImports []string
	}{
		{
			name: "single import",
			content: `# Test
@AGENTS.md
Content`,
			wantImports: []string{"AGENTS.md"},
		},
		{
			name: "multiple imports",
			content: `# Test
@AGENTS.md
@OTHER.md`,
			wantImports: []string{"AGENTS.md", "OTHER.md"},
		},
		{
			name: "no imports",
			content: `# Test
No imports`,
			wantImports: []string{},
		},
		{
			name: "import with whitespace",
			content: `# Test
  @AGENTS.md`,
			wantImports: []string{"AGENTS.md"},
		},
		{
			name: "non-md import ignored",
			content: `# Test
@package.json`,
			wantImports: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imports := extractImports(tt.content)
			if len(imports) != len(tt.wantImports) {
				t.Errorf("extractImports() = %v, want %v", imports, tt.wantImports)
				return
			}
			for i, imp := range imports {
				if imp != tt.wantImports[i] {
					t.Errorf("extractImports()[%d] = %v, want %v", i, imp, tt.wantImports[i])
				}
			}
		})
	}
}
