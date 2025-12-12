package memory

import (
	"fmt"
	"os"
	"path/filepath"

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
			if err := installTemplate("AGENTS.md", target, force); err != nil {
				output.Error(err)
				return err
			}

			// Install CLAUDE.md
			if err := installTemplate("CLAUDE.md", target, force); err != nil {
				output.Error(err)
				return err
			}

			// Optionally run indexing (unless --no-index or --global)
			if !noIndex && !globalInstall {
				if err := indexRepository(target); err != nil {
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

// installTemplate installs a template file to the target directory.
func installTemplate(templateName, targetDir string, force bool) error {
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

// indexRepository generates repository overview and injects into AGENTS.md.
func indexRepository(targetDir string) error {
	// This will be implemented in a later task (Task 7vp.7)
	// For now, just return nil to indicate it's not implemented yet
	return nil
}
