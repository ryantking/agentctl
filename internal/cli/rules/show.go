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

// NewRulesShowCmd creates the rules show command.
func NewRulesShowCmd() *cobra.Command {
	var raw bool

	cmd := &cobra.Command{
		Use:   "show [rule-name]",
		Short: "Display rule content",
		Long: `Display full rule content including frontmatter and body. Takes rule name or filename as argument.
Use --raw to output raw mdc without pretty-printing frontmatter.`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			var repoRoot string
			var err error

			repoRoot, err = git.GetRepoRoot()
			if err != nil {
				output.Errorf("%v\n\nRun from inside a git repository", err)
				return err
			}

			// Determine .agent directory location (AGENTDIR env var or default)
			agentDir, err := getAgentDir(repoRoot)
			if err != nil {
				output.Error(err)
				return err
			}
			rulesDir := filepath.Join(agentDir, "rules")

			ruleName := args[0]
			
			// Find rule file (by name or filename)
			rulePath, err := findRuleFile(rulesDir, ruleName)
			if err != nil {
				output.Error(err)
				return err
			}

			if raw {
				return showRaw(rulePath)
			}

			return showPretty(rulePath)
		},
	}

	cmd.Flags().BoolVarP(&raw, "raw", "r", false, "Output raw mdc file")

	return cmd
}

// findRuleFile finds a rule file by name or filename.
func findRuleFile(rulesDir, ruleName string) (string, error) {
	// Check if rules directory exists
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		return "", fmt.Errorf(`rules directory not found: %s

To fix this:
  - Run 'agentctl rules init' to create the directory
  - Or check your AGENTDIR environment variable`, rulesDir)
	}

	// Try exact filename match first
	exactPath := filepath.Join(rulesDir, ruleName)
	if _, err := os.Stat(exactPath); err == nil {
		return exactPath, nil
	}

	// Try with .mdc extension
	if !strings.HasSuffix(ruleName, ".mdc") {
		exactPath = filepath.Join(rulesDir, ruleName+".mdc")
		if _, err := os.Stat(exactPath); err == nil {
			return exactPath, nil
		}
	}

	// Search by rule name in frontmatter
	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		return "", fmt.Errorf("failed to read rules directory: %w", err)
	}

	var availableRules []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".mdc") {
			continue
		}

		rulePath := filepath.Join(rulesDir, entry.Name())
		metadata, err := parseRuleMetadata(rulePath)
		if err != nil {
			// Still include filename even if metadata parsing fails
			filenameBase := strings.TrimSuffix(entry.Name(), ".mdc")
			availableRules = append(availableRules, filenameBase)
			continue
		}

		// Match by name (case-insensitive)
		if strings.EqualFold(metadata.Name, ruleName) {
			return rulePath, nil
		}

		// Match by filename without extension
		filenameBase := strings.TrimSuffix(entry.Name(), ".mdc")
		if strings.EqualFold(filenameBase, ruleName) {
			return rulePath, nil
		}

		// Collect available rule names for error message
		availableRules = append(availableRules, filenameBase)
	}

	// Build error message with available rules
	if len(availableRules) > 0 {
		var ruleList strings.Builder
		ruleList.WriteString("rule not found: ")
		ruleList.WriteString(ruleName)
		ruleList.WriteString("\n\nAvailable rules:")
		for _, rule := range availableRules {
			ruleList.WriteString("\n  - ")
			ruleList.WriteString(rule)
		}
		ruleList.WriteString("\n\nRun 'agentctl rules list' to see all rules.")
		return "", fmt.Errorf("%s", ruleList.String()) //nolint:revive // Error message is built dynamically
	}

	return "", fmt.Errorf("rule not found: %s. No rules found. Run 'agentctl rules init' to initialize", ruleName)
}

// showRaw outputs the raw rule file content.
func showRaw(rulePath string) error {
	data, err := os.ReadFile(rulePath) //nolint:gosec // Path is controlled, reading rule files
	if err != nil {
		return fmt.Errorf("failed to read rule file: %w", err)
	}

	fmt.Print(string(data))
	return nil
}

// showPretty outputs the rule with pretty-printed frontmatter.
func showPretty(rulePath string) error {
	data, err := os.ReadFile(rulePath) //nolint:gosec // Path is controlled, reading rule files
	if err != nil {
		return fmt.Errorf("failed to read rule file: %w", err)
	}

	content := string(data)
	metadata, err := parseRuleMetadata(rulePath)
	if err != nil {
		// If we can't parse frontmatter, just show raw
		fmt.Print(content)
		return nil
	}

	// Extract body (content after frontmatter)
	body := extractBody(content)

	// Pretty print frontmatter
	fmt.Println("---")
	fmt.Printf("name: %s\n", metadata.Name)
	if metadata.Description != "" {
		fmt.Printf("description: %s\n", metadata.Description)
	}
	if metadata.WhenToUse != "" {
		fmt.Printf("when-to-use: %s\n", metadata.WhenToUse)
	}
	if len(metadata.AppliesTo) > 0 {
		fmt.Printf("applies-to: [%s]\n", strings.Join(metadata.AppliesTo, ", "))
	}
	if metadata.Priority != 0 {
		fmt.Printf("priority: %d\n", metadata.Priority)
	}
	if len(metadata.Tags) > 0 {
		fmt.Printf("tags: [%s]\n", strings.Join(metadata.Tags, ", "))
	}
	if metadata.Version != "" {
		fmt.Printf("version: %s\n", metadata.Version)
	}
	fmt.Println("---")
	fmt.Print(body)

	return nil
}

// extractBody extracts the markdown body after frontmatter.
func extractBody(content string) string {
	lines := strings.Split(content, "\n")
	
	// Find first ---
	startIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			startIdx = i
			break
		}
	}

	if startIdx == -1 {
		return content
	}

	// Find second ---
	endIdx := -1
	for i := startIdx + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return content
	}

	// Extract body content (after second ---)
	bodyLines := lines[endIdx+1:]
	return strings.Join(bodyLines, "\n")
}
