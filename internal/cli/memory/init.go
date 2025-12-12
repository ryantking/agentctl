package memory

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ryantking/agentctl/internal/git"
	"github.com/ryantking/agentctl/internal/output"
	"github.com/ryantking/agentctl/internal/templates"
	"github.com/spf13/cobra"
)

// NewMemoryInitCmd creates the memory init command.
func NewMemoryInitCmd() *cobra.Command {
	var globalInstall, force, noIndex bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize agent memory files (AGENTS.md and CLAUDE.md)",
		Long: `Initialize agent memory files from templates. Installs AGENTS.md and CLAUDE.md with proper import syntax.
By default, skips existing files unless --force is specified.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			var target string
			var err error

			if globalInstall {
				home, err := os.UserHomeDir()
				if err != nil {
					output.Errorf("failed to get home directory: %v", err)
					return err
				}
				target = filepath.Join(home, ".claude")
			} else {
				target, err = git.GetRepoRoot()
				if err != nil {
					output.Errorf("%v\n\nRun from inside a git repository or use --global", err)
					return err
				}
			}

			fmt.Println("Installing memory files...")

			// Install AGENTS.md
			if err := InstallTemplate("AGENTS.md", target, force); err != nil {
				output.Error(err)
				return err
			}

			// Install CLAUDE.md
			if err := InstallTemplate("CLAUDE.md", target, force); err != nil {
				output.Error(err)
				return err
			}

			// Optionally run indexing (unless --no-index or --global)
			if !noIndex && !globalInstall {
				if err := IndexRepository(target); err != nil {
					// Non-fatal: warn but continue
					fmt.Printf("  → Repository indexing skipped: %v\n", err)
				}
			}

			fmt.Println("\n✓ Memory files initialized successfully")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&globalInstall, "global", "g", false, "Install to $HOME/.claude instead of current repository")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing files")
	cmd.Flags().BoolVar(&noIndex, "no-index", false, "Skip repository indexing step")

	return cmd
}

// InstallTemplate installs a template file to the target directory.
// This function is exported so it can be called from other packages.
func InstallTemplate(templateName, targetDir string, force bool) error {
	destPath := filepath.Join(targetDir, templateName)

	// Check if file exists and skip if not forcing
	if _, err := os.Stat(destPath); err == nil && !force {
		relPath, _ := filepath.Rel(targetDir, destPath)
		fmt.Printf("  • %s (skipped)\n", relPath)
		return nil
	}

	// Read template from embedded filesystem
	data, err := templates.GetTemplate(templateName)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", templateName, err)
	}

	// Ensure target directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil { //nolint:gosec // Template directories need to be readable
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if file existed before writing
	existed := false
	if _, err := os.Stat(destPath); err == nil {
		existed = true
	}

	// Write template file
	if err := os.WriteFile(destPath, data, 0644); err != nil { //nolint:gosec // Template files need to be readable
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Determine status for output
	relPath, _ := filepath.Rel(targetDir, destPath)
	status := "created"
	if existed {
		status = "overwritten"
	}
	fmt.Printf("  • %s (%s)\n", relPath, status)

	return nil
}

// IndexRepository generates repository overview and injects into AGENTS.md.
// This function is exported so it can be called from other packages.
func IndexRepository(targetDir string) error {
	return indexRepository(targetDir)
}

func indexRepository(targetDir string) error {
	if _, err := exec.LookPath("claude"); err != nil {
		return fmt.Errorf("claude CLI not found")
	}

	prompt := `Analyze this repository and provide a concise overview:
- Main purpose and key technologies
- Directory structure (2-3 levels max)
- Entry points and main files
- Build/run commands (check for package.json scripts, Makefile targets, Justfile recipes, etc.)
- Available scripts and automation tools

Format as clean markdown starting at heading level 3 (###), keep it brief (under 500 words).`

	fmt.Print("  → Indexing repository with Claude CLI...")

	cmdCtx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "claude", "--print", "--output-format", "text", prompt)
	cmd.Dir = targetDir
	cmd.Env = os.Environ()

	output, err := cmd.Output()
	if err != nil {
		return err
	}

	indexContent := strings.TrimSpace(string(output))
	if indexContent == "" {
		return fmt.Errorf("empty output from Claude CLI")
	}

	if err := insertRepositoryIndex(targetDir, indexContent); err != nil {
		return err
	}

	fmt.Println(" done")
	return nil
}

func insertRepositoryIndex(targetDir, indexContent string) error {
	agentsMDPath := filepath.Join(targetDir, "AGENTS.md")
	if _, err := os.Stat(agentsMDPath); os.IsNotExist(err) {
		return fmt.Errorf("AGENTS.md not found")
	}

	data, err := os.ReadFile(agentsMDPath) //nolint:gosec // Path is controlled, reading template files
	if err != nil {
		return err
	}

	content := string(data)
	startMarker := "<!-- REPOSITORY_INDEX_START -->"
	endMarker := "<!-- REPOSITORY_INDEX_END -->"

	startIdx := strings.Index(content, startMarker)
	endIdx := strings.Index(content, endMarker)

	if startIdx == -1 || endIdx == -1 {
		return fmt.Errorf("repository index markers not found in AGENTS.md")
	}

	updatedContent := content[:startIdx+len(startMarker)] + "\n" + indexContent + "\n" + content[endIdx:]

	return os.WriteFile(agentsMDPath, []byte(updatedContent), 0644) //nolint:gosec // Template files need to be readable
}
