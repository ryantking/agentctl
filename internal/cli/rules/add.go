// Package rules provides commands for managing agent rules in the .agent directory.
package rules

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	agentclient "github.com/ryantking/agentctl/internal/agent"
	"github.com/ryantking/agentctl/internal/git"
	"github.com/ryantking/agentctl/internal/output"
	"github.com/spf13/cobra"
)

// NewRulesAddCmd creates the rules add command.
func NewRulesAddCmd() *cobra.Command {
	var name, description, whenToUse string
	var appliesTo []string

	cmd := &cobra.Command{
		Use:   "add [prompt]",
		Short: "Add a new rule from a prompt",
		Long: `Add a new rule by describing what it should do. The command will generate an mdc file
with proper frontmatter based on your description. Use --name to specify the filename.

Example:
  agentctl rules add "Always use conventional commits" --name git-commits`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			// Ensure rules directory exists
			if err := os.MkdirAll(rulesDir, 0755); err != nil { //nolint:gosec // Rules directory needs to be readable
				return fmt.Errorf("failed to create rules directory: %w", err)
			}

			prompt := strings.Join(args, " ")

			// Validate prompt
			if err := validatePrompt(prompt); err != nil {
				return err
			}

			// Validate description if provided
			if description != "" {
				if err := validateDescription(description); err != nil {
					return err
				}
			}

			// Validate when-to-use if provided
			if whenToUse != "" {
				if err := validateWhenToUse(whenToUse); err != nil {
					return err
				}
			}

			// Validate applies-to
			if err := validateAppliesTo(appliesTo); err != nil {
				return err
			}

			// Determine filename
			filename := name
			if filename == "" {
				// Generate filename from prompt (first few words, sanitized)
				words := strings.Fields(prompt)
				if len(words) > 0 {
					maxWords := 3
					if len(words) < maxWords {
						maxWords = len(words)
					}
					filename = sanitizeSkillName(strings.Join(words[:maxWords], "-"))
				}
				if filename == "" {
					filename = "new-rule"
				}
			}

			// Remove .mdc extension if present for validation
			filenameBase := strings.TrimSuffix(filename, ".mdc")
			if err := validateRuleName(filenameBase); err != nil {
				return fmt.Errorf("invalid rule name: %w", err)
			}

			// Check for name conflicts with existing rules
			if err := checkNameConflict(rulesDir, filenameBase); err != nil {
				return err
			}

			if !strings.HasSuffix(filename, ".mdc") {
				filename += ".mdc"
			}

			rulePath := filepath.Join(rulesDir, filename)

			// Check if file already exists
			if _, err := os.Stat(rulePath); err == nil {
				relPath, _ := filepath.Rel(repoRoot, rulePath)
				return fmt.Errorf(`rule file already exists: %s

To fix this:
  - Use a different --name: agentctl rules add "%s" --name different-name
  - Or remove the existing file: agentctl rules remove %s
  - Or view the existing rule: agentctl rules show %s`, relPath, prompt, filename, strings.TrimSuffix(filename, ".mdc"))
			}
			// File doesn't exist, which is what we want - continue

			// Create agent from global flags
			agent := newAgentFromCmdFlags(cmd)

			// Generate rule content
			ruleContent, err := generateRuleContent(agent, prompt, name, description, whenToUse, appliesTo)
			if err != nil {
				return fmt.Errorf("failed to generate rule content: %w", err)
			}

			// Write rule file
			if err := os.WriteFile(rulePath, []byte(ruleContent), 0644); err != nil { //nolint:gosec // Rule file needs to be readable
				return fmt.Errorf("failed to write rule file: %w", err)
			}

			relPath, _ := filepath.Rel(repoRoot, rulePath)
			fmt.Printf("  • %s (created)\n", relPath)
			fmt.Println("\n✓ Rule created successfully")
			fmt.Println("\nTip: Edit the rule file to refine the content, then run 'agentctl rules sync' to sync to other formats.")

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Filename for the rule (without .mdc extension)")
	cmd.Flags().StringVar(&description, "description", "", "Description for the rule (auto-generated if not provided)")
	cmd.Flags().StringVar(&whenToUse, "when-to-use", "", "When to use this rule (auto-generated if not provided)")
	cmd.Flags().StringSliceVar(&appliesTo, "applies-to", []string{"claude"}, "Tools this rule applies to (comma-separated)")

	return cmd
}

// newAgentFromCmdFlags creates an Agent from command global flags.
func newAgentFromCmdFlags(cmd *cobra.Command) *agentclient.Agent {
	opts := []agentclient.Option{}

	// Get agent type from flag or environment
	if agentType, err := cmd.Flags().GetString("agent-type"); err == nil && agentType != "" {
		opts = append(opts, agentclient.WithType(agentType))
	}

	// Get agent binary from flag or environment (check both --agent-binary and deprecated --agent-cli)
	if agentBinary, err := cmd.Flags().GetString("agent-binary"); err == nil && agentBinary != "" {
		opts = append(opts, agentclient.WithBinary(agentBinary))
	} else if agentCLI, err := cmd.Flags().GetString("agent-cli"); err == nil && agentCLI != "" {
		opts = append(opts, agentclient.WithBinary(agentCLI))
	}

	return agentclient.NewAgent(opts...)
}

// generateRuleContent generates rule content from a prompt using agent CLI.
func generateRuleContent(agent *agentclient.Agent, prompt, name, description, whenToUse string, appliesTo []string) (string, error) {
	// Check if agent CLI is configured
	if !agentclient.IsConfigured() {
		return "", fmt.Errorf("agent CLI not found or ANTHROPIC_API_KEY not set\n\nTo fix this:\n  - Install Claude Code: https://claude.ai/code\n  - Or set ANTHROPIC_API_KEY environment variable: export ANTHROPIC_API_KEY=your-key\n  - Get your API key at https://console.anthropic.com/")
	}

	systemPrompt := `You are creating a rule file (.mdc format) for an agent rules system.

The rule file format is:
- YAML frontmatter with metadata
- Markdown body with guidelines and examples

Required frontmatter fields:
- name: Human-readable rule name
- description: One-line description
- when-to-use: Context for when this rule applies
- applies-to: List of tools (default: ["claude"])
- priority: 0-4 (default: 2)
- tags: Array of tags
- version: Semantic version (default: "1.0.0")

The body should include:
- Clear guidelines on how to apply the rule
- Examples when helpful
- Best practices

Generate a complete .mdc rule file based on the user's prompt.`

	userPrompt := fmt.Sprintf(`Create a rule file for: %s`, prompt)

	// Add any provided metadata to the prompt
	if name != "" {
		userPrompt += fmt.Sprintf("\n\nRule name: %s", name)
	}
	if description != "" {
		userPrompt += fmt.Sprintf("\n\nDescription: %s", description)
	}
	if whenToUse != "" {
		userPrompt += fmt.Sprintf("\n\nWhen to use: %s", whenToUse)
	}
	if len(appliesTo) > 0 {
		userPrompt += fmt.Sprintf("\n\nApplies to: %s", strings.Join(appliesTo, ", "))
	}

	fmt.Print("  → Generating rule content with agent CLI...")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	content, err := agent.ExecuteWithSystemLogger(ctx, systemPrompt, userPrompt, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate rule content: %w", err)
	}

	fmt.Println(" (done)")
	return content, nil
}

// validatePrompt validates the prompt argument.
func validatePrompt(prompt string) error {
	prompt = strings.TrimSpace(prompt)
	if len(prompt) < 10 {
		return fmt.Errorf("prompt too short (minimum 10 characters)")
	}
	if len(prompt) > 1000 {
		return fmt.Errorf("prompt too long (maximum 1000 characters)")
	}
	return nil
}

// validateRuleName validates a rule name.
func validateRuleName(name string) error {
	if name == "" {
		return fmt.Errorf("rule name cannot be empty")
	}
	if len(name) > 50 {
		return fmt.Errorf("rule name too long (max 50 characters)")
	}
	matched, err := regexp.MatchString("^[a-z0-9-]+$", name)
	if err != nil {
		return fmt.Errorf("failed to validate rule name: %w", err)
	}
	if !matched {
		return fmt.Errorf("rule name can only contain lowercase letters, numbers, and hyphens")
	}
	return nil
}

// validateDescription validates the description field.
func validateDescription(description string) error {
	description = strings.TrimSpace(description)
	if description == "" {
		return fmt.Errorf("description cannot be empty or whitespace-only")
	}
	if len(description) > 200 {
		return fmt.Errorf("description too long (max 200 characters)")
	}
	return nil
}

// validateWhenToUse validates the when-to-use field.
func validateWhenToUse(whenToUse string) error {
	whenToUse = strings.TrimSpace(whenToUse)
	if whenToUse == "" {
		return fmt.Errorf("when-to-use cannot be empty or whitespace-only")
	}
	if len(whenToUse) > 300 {
		return fmt.Errorf("when-to-use too long (max 300 characters)")
	}
	return nil
}

// validateAppliesTo validates the applies-to tools list.
var knownTools = []string{"claude", "cursor", "windsurf", "aider"}

func validateAppliesTo(tools []string) error {
	for _, tool := range tools {
		tool = strings.TrimSpace(strings.ToLower(tool))
		if tool == "" {
			return fmt.Errorf("applies-to cannot contain empty values")
		}
		found := false
		for _, known := range knownTools {
			if tool == known {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("⚠ Warning: unknown tool '%s' (known tools: %s)\n", tool, strings.Join(knownTools, ", "))
		}
	}
	return nil
}

// checkNameConflict checks if a rule name conflicts with existing rules.
func checkNameConflict(rulesDir, name string) error {
	// Check if rules directory exists
	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		return nil // No conflicts if directory doesn't exist
	}

	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		return nil // Skip conflict check if we can't read directory
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".mdc") {
			continue
		}

		filenameBase := strings.TrimSuffix(entry.Name(), ".mdc")
		if strings.EqualFold(filenameBase, name) {
			return fmt.Errorf(`rule name conflicts with existing rule: %s

To fix this:
  - Use a different --name: agentctl rules add "<prompt>" --name different-name
  - Or remove the existing rule: agentctl rules remove %s`, entry.Name(), name)
		}
	}

	return nil
}

