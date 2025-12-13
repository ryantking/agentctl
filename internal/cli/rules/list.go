package rules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ryantking/agentctl/internal/git"
	"github.com/ryantking/agentctl/internal/output"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// RuleMetadata represents the frontmatter metadata from a rule file.
type RuleMetadata struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	WhenToUse   string   `yaml:"when-to-use"`
	AppliesTo   []string `yaml:"applies-to"`
	Priority    int      `yaml:"priority"`
	Tags        []string `yaml:"tags"`
	Version     string   `yaml:"version"`
}

// RuleInfo represents a rule with its metadata and filename.
type RuleInfo struct {
	Filename    string        `json:"filename"`
	Metadata    RuleMetadata  `json:"metadata"`
}

// NewRulesListCmd creates the rules list command.
func NewRulesListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all rules in .agent/rules/",
		Long: `List all rules in the .agent/rules/ directory. Shows rule name, description, and when-to-use
from frontmatter. Use --json for structured output.`,
		RunE: func(_ *cobra.Command, _ []string) error {
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

			rules, err := listRules(rulesDir)
			if err != nil {
				output.Error(err)
				return err
			}

			if jsonOutput {
				return outputJSON(rules)
			}

			return outputText(rules)
		},
	}

	cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output as JSON")

	return cmd
}

// getAgentDir returns the .agent directory path, respecting AGENTDIR env var.
func getAgentDir(repoRoot string) string {
	agentDir := os.Getenv("AGENTDIR")
	if agentDir == "" {
		agentDir = ".agent"
	}
	// If relative path, make it relative to repo root
	if !filepath.IsAbs(agentDir) {
		agentDir = filepath.Join(repoRoot, agentDir)
	}
	return agentDir
}

// listRules reads all .mdc files from the rules directory and extracts metadata.
func listRules(rulesDir string) ([]RuleInfo, error) {
	// Check if rules directory exists
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("rules directory not found: %s\n\nRun 'agentctl rules init' to initialize", rulesDir)
	}

	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules directory: %w", err)
	}

	var rules []RuleInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".mdc") {
			continue
		}

		rulePath := filepath.Join(rulesDir, entry.Name())
		metadata, err := parseRuleMetadata(rulePath)
		if err != nil {
			// Skip files with invalid frontmatter but continue
			continue
		}

		rules = append(rules, RuleInfo{
			Filename: entry.Name(),
			Metadata: metadata,
		})
	}

	return rules, nil
}

// parseRuleMetadata extracts YAML frontmatter from a rule file.
func parseRuleMetadata(rulePath string) (RuleMetadata, error) {
	data, err := os.ReadFile(rulePath) //nolint:gosec // Path is controlled, reading rule files
	if err != nil {
		return RuleMetadata{}, err
	}

	content := string(data)
	
	// Extract YAML frontmatter (between --- markers)
	frontmatter, err := extractFrontmatter(content)
	if err != nil {
		return RuleMetadata{}, err
	}

	var metadata RuleMetadata
	if err := yaml.Unmarshal([]byte(frontmatter), &metadata); err != nil {
		return RuleMetadata{}, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	return metadata, nil
}

// extractFrontmatter extracts YAML frontmatter from markdown content.
func extractFrontmatter(content string) (string, error) {
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
		return "", fmt.Errorf("no frontmatter found (missing ---)")
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
		return "", fmt.Errorf("unclosed frontmatter (missing closing ---)")
	}

	// Extract frontmatter content
	frontmatterLines := lines[startIdx+1 : endIdx]
	return strings.Join(frontmatterLines, "\n"), nil
}

// outputText outputs rules in human-readable format.
func outputText(rules []RuleInfo) error {
	if len(rules) == 0 {
		fmt.Println("No rules found.")
		return nil
	}

	for _, rule := range rules {
		fmt.Printf("Rule: %s\n", rule.Metadata.Name)
		if rule.Metadata.Description != "" {
			fmt.Printf("  Description: %s\n", rule.Metadata.Description)
		}
		if rule.Metadata.WhenToUse != "" {
			fmt.Printf("  When to use: %s\n", rule.Metadata.WhenToUse)
		}
		if len(rule.Metadata.AppliesTo) > 0 {
			fmt.Printf("  Applies to: %s\n", strings.Join(rule.Metadata.AppliesTo, ", "))
		}
		fmt.Printf("  File: %s\n", rule.Filename)
		fmt.Println()
	}

	return nil
}

// outputJSON outputs rules as JSON.
func outputJSON(rules []RuleInfo) error {
	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
