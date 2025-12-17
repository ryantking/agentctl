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
	Globs       []string `yaml:"globs"`
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
		Long: `List all rules in the .agent/rules/ directory. Shows rule name, description, and globs
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
			agentDir, err := getAgentDir(repoRoot)
			if err != nil {
				output.Error(err)
				return err
			}
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
func getAgentDir(repoRoot string) (string, error) {
	agentDir := os.Getenv("AGENTDIR")
	if agentDir == "" {
		agentDir = ".agent"
	} else {
		// Validate AGENTDIR if set
		if err := validateAgentDir(agentDir); err != nil {
			return "", fmt.Errorf("invalid AGENTDIR environment variable: %w", err)
		}
	}
	// If relative path, make it relative to repo root
	if !filepath.IsAbs(agentDir) {
		agentDir = filepath.Join(repoRoot, agentDir)
	}
	return agentDir, nil
}

// validateAgentDir validates the AGENTDIR environment variable value.
func validateAgentDir(agentDir string) error {
	if agentDir == "" {
		return fmt.Errorf("AGENTDIR cannot be empty")
	}

	// Check for invalid characters (basic validation)
	invalidChars := []string{"..", "~", "$", "`"}
	for _, char := range invalidChars {
		if strings.Contains(agentDir, char) {
			var reason string
			switch char {
			case "..":
				reason = `".." could be used for path traversal attacks`
			case "~", "$":
				reason = fmt.Sprintf(`"%s" is a shell expansion character that could cause unexpected behavior`, char)
			case "`":
				reason = `"` + "`" + `" is a shell expansion character that could cause unexpected behavior`
			default:
				reason = fmt.Sprintf(`"%s" is not allowed`, char)
			}
			return fmt.Errorf(`AGENTDIR contains invalid character sequence: %s

This is not allowed because:
  - %s

To fix this:
  - Use a simple relative path like ".agent" or "custom-agent"
  - Or use an absolute path without these characters
  - Example: export AGENTDIR=/path/to/agent`, char, reason)
		}
	}

	// Check for absolute paths that don't exist (warn but allow)
	if filepath.IsAbs(agentDir) {
		if _, err := os.Stat(agentDir); os.IsNotExist(err) {
			// Allow non-existent absolute paths (will be created)
			// But validate parent directory exists
			parent := filepath.Dir(agentDir)
			if parent != agentDir && parent != "/" {
				if _, err := os.Stat(parent); os.IsNotExist(err) {
					return fmt.Errorf("AGENTDIR parent directory does not exist: %s: %w", parent, err)
				}
			}
		}
	}

	return nil
}

// listRules reads all .mdc files from the rules directory and extracts metadata.
func listRules(rulesDir string) ([]RuleInfo, error) {
	// Check if rules directory exists
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		return nil, fmt.Errorf(`rules directory not found: %s

To fix this:
  - Run 'agentctl rules init' to create the directory
  - Or check your AGENTDIR environment variable`, rulesDir)
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

	// Validate metadata schema
	if err := validateRuleMetadata(metadata, rulePath); err != nil {
		return RuleMetadata{}, err
	}

	return metadata, nil
}

// validateRuleMetadata validates rule frontmatter schema.
func validateRuleMetadata(metadata RuleMetadata, rulePath string) error {
	var errors []string

	// Required fields must be non-empty
	if strings.TrimSpace(metadata.Name) == "" {
		errors = append(errors, "required field 'name' is empty")
	}
	// Description and globs are optional

	// Priority must be 0-4
	if metadata.Priority < 0 || metadata.Priority > 4 {
		errors = append(errors, fmt.Sprintf("priority must be 0-4, got %d", metadata.Priority))
	}

	// Version should follow semver (basic validation)
	if metadata.Version != "" {
		if !isValidSemver(metadata.Version) {
			errors = append(errors, fmt.Sprintf("version should follow semver format (e.g., '1.0.0'), got '%s'", metadata.Version))
		}
	}

	if len(errors) > 0 {
		relPath := rulePath
		if cwd, err := os.Getwd(); err == nil {
			if rel, err := filepath.Rel(cwd, rulePath); err == nil {
				relPath = rel
			}
		}
		return fmt.Errorf(`validation errors in %s:
  - %s

Edit the file to fix these issues:
  vim %s`, relPath, strings.Join(errors, "\n  - "), relPath)
	}

	return nil
}

// isValidSemver performs basic semver validation (major.minor.patch).
func isValidSemver(version string) bool {
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")
	
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return false
	}

	// Check each part is numeric
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return false
		}
		// Allow numeric with optional pre-release/build metadata
		// Basic check: starts with digit
		if len(part) == 0 || (part[0] < '0' || part[0] > '9') {
			return false
		}
	}

	return true
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
		if len(rule.Metadata.Globs) > 0 {
			fmt.Printf("  Globs: %s\n", strings.Join(rule.Metadata.Globs, ", "))
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
