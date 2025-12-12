package memory

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateMemoryFiles(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		setup       func() error
		wantError   bool
		wantWarnings bool
	}{
		{
			name: "valid files",
			setup: func() error {
				// Create valid AGENTS.md
				agentsContent := `# Agents
## Rules
Rules here
## Workspaces
Workspaces here
## Git
Git here
## Tool Selection Guidelines
Tools here
<!-- REPOSITORY_INDEX_START -->
<!-- REPOSITORY_INDEX_END -->`
				agentsPath := filepath.Join(tmpDir, "AGENTS.md")
				if err := os.WriteFile(agentsPath, []byte(agentsContent), 0644); err != nil {
					return err
				}

				// Create valid CLAUDE.md with import
				claudeContent := `# Claude Code
@AGENTS.md
## Agent Orchestration
Content`
				claudePath := filepath.Join(tmpDir, "CLAUDE.md")
				return os.WriteFile(claudePath, []byte(claudeContent), 0644)
			},
			wantError:   false,
			wantWarnings: false,
		},
		{
			name: "missing AGENTS.md",
			setup: func() error {
				claudePath := filepath.Join(tmpDir, "CLAUDE.md")
				return os.WriteFile(claudePath, []byte("# Test"), 0644)
			},
			wantError: true,
		},
		{
			name: "missing CLAUDE.md",
			setup: func() error {
				agentsPath := filepath.Join(tmpDir, "AGENTS.md")
				return os.WriteFile(agentsPath, []byte("# Test"), 0644)
			},
			wantError: true,
		},
		{
			name: "missing import in CLAUDE.md",
			setup: func() error {
				agentsPath := filepath.Join(tmpDir, "AGENTS.md")
				if err := os.WriteFile(agentsPath, []byte("# Test"), 0644); err != nil {
					return err
				}
				claudePath := filepath.Join(tmpDir, "CLAUDE.md")
				return os.WriteFile(claudePath, []byte("# Test\nNo import"), 0644)
			},
			wantError: true,
		},
		{
			name: "missing required section",
			setup: func() error {
				agentsPath := filepath.Join(tmpDir, "AGENTS.md")
				if err := os.WriteFile(agentsPath, []byte("# Test\n## Rules\n"), 0644); err != nil {
					return err
				}
				claudePath := filepath.Join(tmpDir, "CLAUDE.md")
				return os.WriteFile(claudePath, []byte("# Test\n@AGENTS.md"), 0644)
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Cleanup
			os.Remove(filepath.Join(tmpDir, "AGENTS.md"))
			os.Remove(filepath.Join(tmpDir, "CLAUDE.md"))

			if err := tt.setup(); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			err := validateMemoryFiles(tmpDir)
			hasError := err != nil

			if hasError != tt.wantError {
				t.Errorf("validateMemoryFiles() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestHasCircularImport(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file1.md that imports file2.md
	file1Path := filepath.Join(tmpDir, "file1.md")
	file1Content := `# File 1
@file2.md`
	os.WriteFile(file1Path, []byte(file1Content), 0644)

	// Create file2.md that imports file1.md (circular)
	file2Path := filepath.Join(tmpDir, "file2.md")
	file2Content := `# File 2
@file1.md`
	os.WriteFile(file2Path, []byte(file2Content), 0644)

	// Test circular import detection
	if !hasCircularImport(file1Content, tmpDir, 0, 5) {
		t.Error("Should detect circular import")
	}

	// Test non-circular import (file3 -> file4, no cycle)
	file4Path := filepath.Join(tmpDir, "file4.md")
	file4Content := `# File 4
No imports`
	os.WriteFile(file4Path, []byte(file4Content), 0644)

	file3Path := filepath.Join(tmpDir, "file3.md")
	file3Content := `# File 3
@file4.md`
	os.WriteFile(file3Path, []byte(file3Content), 0644)

	if hasCircularImport(file3Content, tmpDir, 0, 5) {
		t.Error("Should not detect circular import for non-circular case")
	}
}
