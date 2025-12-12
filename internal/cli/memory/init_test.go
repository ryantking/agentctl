package memory

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		template  string
		force     bool
		wantError bool
	}{
		{
			name:      "install AGENTS.md",
			template:  "AGENTS.md",
			force:     false,
			wantError: false,
		},
		{
			name:      "install CLAUDE.md",
			template:  "CLAUDE.md",
			force:     false,
			wantError: false,
		},
		{
			name:      "install with force",
			template:  "AGENTS.md",
			force:     true,
			wantError: false,
		},
		{
			name:      "skip existing file",
			template:  "AGENTS.md",
			force:     false,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InstallTemplate(tt.template, tmpDir, tt.force)
			if (err != nil) != tt.wantError {
				t.Errorf("InstallTemplate() error = %v, wantError %v", err, tt.wantError)
				return
			}

			// Verify file exists
			filePath := filepath.Join(tmpDir, tt.template)
			if _, err := os.Stat(filePath); os.IsNotExist(err) && !tt.wantError {
				t.Errorf("InstallTemplate() file %s was not created", tt.template)
			}
		})
	}
}

func TestExtractImports(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantImports []string
	}{
		{
			name: "single import on own line",
			content: `# Test
@AGENTS.md
Some content.`,
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
No imports here.`,
			wantImports: []string{},
		},
		{
			name: "import with leading whitespace",
			content: `# Test
  @AGENTS.md
Content.`,
			wantImports: []string{"AGENTS.md"},
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
