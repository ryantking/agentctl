package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ryantking/agentctl/internal/git"
	"github.com/ryantking/agentctl/internal/output"
	"github.com/spf13/cobra"
)

// NewRulesRemoveCmd creates the rules remove command.
func NewRulesRemoveCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "remove [rule-name...]",
		Short: "Remove rule files from .agent/rules/",
		Long: `Remove rule files from .agent/rules/. Takes rule name or filename as argument.
Supports removing multiple rules at once. Prompts for confirmation unless --force is used.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			var repoRoot string
			var err error

			repoRoot, err = git.GetRepoRoot()
			if err != nil {
				output.Errorf("%v\n\nRun from inside a git repository", err)
				return err
			}

			// Determine .agent directory location (AGENTDIR env var or default)
			agentDir := getAgentDir(repoRoot)
			rulesDir := filepath.Join(agentDir, "rules")

			// Check if rules directory exists
			if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
				return fmt.Errorf("rules directory not found: %s\n\nRun 'agentctl rules init' to initialize", rulesDir)
			}

			var removed []string
			var errors []string

			for _, ruleName := range args {
				rulePath, err := findRuleFile(rulesDir, ruleName)
				if err != nil {
					errors = append(errors, fmt.Sprintf("  ✗ %s: %v", ruleName, err))
					continue
				}

				// Prompt for confirmation unless --force
				if !force {
					relPath, _ := filepath.Rel(repoRoot, rulePath)
					fmt.Printf("Remove rule: %s? [y/N]: ", relPath)
					var response string
					_, _ = fmt.Scanln(&response) // Ignore error - user input
					if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
						fmt.Printf("  → Skipped %s\n", ruleName)
						continue
					}
				}

				// Remove the file
				if err := os.Remove(rulePath); err != nil {
					errors = append(errors, fmt.Sprintf("  ✗ %s: failed to remove: %v", ruleName, err))
					continue
				}

				relPath, _ := filepath.Rel(repoRoot, rulePath)
				fmt.Printf("  • %s (removed)\n", relPath)
				removed = append(removed, ruleName)
			}

			// Print errors if any
			if len(errors) > 0 {
				for _, errMsg := range errors {
					fmt.Println(errMsg)
				}
			}

			if len(removed) > 0 {
				fmt.Printf("\n✓ Removed %d rule(s)\n", len(removed))
			}

			if len(errors) > 0 {
				return fmt.Errorf("failed to remove %d rule(s)", len(errors))
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}
