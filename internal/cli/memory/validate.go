package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ryantking/agentctl/internal/git"
	"github.com/ryantking/agentctl/internal/output"
	"github.com/spf13/cobra"
)

// NewMemoryValidateCmd creates the memory validate command.
func NewMemoryValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate memory files for common issues",
		Long: `Validate memory files for common issues:
- Line count warnings (AGENTS.md > 300 lines, CLAUDE.md > 200)
- Missing @AGENTS.md import in CLAUDE.md
- Circular import detection (max 5 hops)
- Conflicting patterns (e.g., .claude/plans references when beads installed)
- Required sections in AGENTS.md`,
		RunE: func(_ *cobra.Command, _ []string) error {
			var target string
			var err error

			target, err = git.GetRepoRoot()
			if err != nil {
				// Try current directory if not in git repo
				target, err = os.Getwd()
				if err != nil {
					output.Errorf("failed to determine target directory: %v", err)
					return err
				}
			}

			return validateMemoryFiles(target)
		},
	}

	return cmd
}

func validateMemoryFiles(target string) error {
	agentsPath := filepath.Join(target, "AGENTS.md")
	claudePath := filepath.Join(target, "CLAUDE.md")

	var warnings []string
	var errors []string

	// Check if files exist
	agentsExists := fileExists(agentsPath)
	claudeExists := fileExists(claudePath)

	if !agentsExists {
		errors = append(errors, "AGENTS.md does not exist")
	}
	if !claudeExists {
		errors = append(errors, "CLAUDE.md does not exist")
	}

	if !agentsExists || !claudeExists {
		for _, err := range errors {
			output.Errorf("%s", err)
		}
		return fmt.Errorf("validation failed: missing required files")
	}

	// Validate AGENTS.md
	agentsData, err := os.ReadFile(agentsPath) //nolint:gosec // Path is controlled, reading template files
	if err != nil {
		return fmt.Errorf("failed to read AGENTS.md: %w", err)
	}
	agentsContent := string(agentsData)
	agentsLines := len(strings.Split(agentsContent, "\n"))

	if agentsLines > 300 {
		warnings = append(warnings, fmt.Sprintf("AGENTS.md has %d lines (recommended: < 300)", agentsLines))
	}

	// Check required sections in AGENTS.md
	requiredSections := []string{"Rules", "Workspaces", "Git", "Tool Selection Guidelines"}
	for _, section := range requiredSections {
		if !strings.Contains(agentsContent, "## "+section) {
			errors = append(errors, fmt.Sprintf("AGENTS.md missing required section: %s", section))
		}
	}

	// Validate CLAUDE.md
	claudeData, err := os.ReadFile(claudePath) //nolint:gosec // Path is controlled, reading template files
	if err != nil {
		return fmt.Errorf("failed to read CLAUDE.md: %w", err)
	}
	claudeContent := string(claudeData)
	claudeLines := len(strings.Split(claudeContent, "\n"))

	if claudeLines > 200 {
		warnings = append(warnings, fmt.Sprintf("CLAUDE.md has %d lines (recommended: < 200)", claudeLines))
	}

	// Check for @AGENTS.md import in CLAUDE.md
	if !strings.Contains(claudeContent, "@AGENTS.md") {
		errors = append(errors, "CLAUDE.md missing required @AGENTS.md import")
	}

	// Check for circular imports (max 5 hops)
	if hasCircularImport(claudeContent, target, 0, 5) {
		errors = append(errors, "Circular import detected (max 5 hops exceeded)")
	}

	// Check for conflicting patterns (.claude/plans references when beads might be used)
	if strings.Contains(agentsContent, ".claude/plans") || strings.Contains(claudeContent, ".claude/plans") {
		warnings = append(warnings, "Found .claude/plans references (conflicts with beads methodology)")
	}

	// Output results
	if len(warnings) > 0 {
		fmt.Println("Warnings:")
		for _, warning := range warnings {
			fmt.Printf("  ⚠ %s\n", warning)
		}
		fmt.Println()
	}

	if len(errors) > 0 {
		fmt.Println("Errors:")
		for _, err := range errors {
			fmt.Printf("  ✗ %s\n", err)
		}
		return fmt.Errorf("validation failed: %d error(s)", len(errors))
	}

	if len(warnings) == 0 {
		fmt.Println("✓ All validation checks passed")
	} else {
		fmt.Printf("✓ Validation passed with %d warning(s)\n", len(warnings))
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func hasCircularImport(content, baseDir string, depth, maxDepth int) bool {
	if depth > maxDepth {
		return true
	}

	imports := extractImports(content)
	for _, imp := range imports {
		importPath := filepath.Join(baseDir, imp)
		importData, err := os.ReadFile(importPath) //nolint:gosec // Path is controlled, checking for circular imports
		if err != nil {
			continue
		}
		if hasCircularImport(string(importData), baseDir, depth+1, maxDepth) {
			return true
		}
	}

	return false
}
