// Package rules provides commands for managing agent rules in the .agent directory.
package rules

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

			// Ensure rules directory exists
			if err := os.MkdirAll(rulesDir, 0755); err != nil { //nolint:gosec // Rules directory needs to be readable
				return fmt.Errorf("failed to create rules directory: %w", err)
			}

			prompt := strings.Join(args, " ")

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
			if !strings.HasSuffix(filename, ".mdc") {
				filename += ".mdc"
			}

			rulePath := filepath.Join(rulesDir, filename)

			// Check if file already exists
			if _, err := os.Stat(rulePath); err == nil {
				return fmt.Errorf("rule file already exists: %s\n\nUse a different --name or remove the existing file", filename)
			}
			// File doesn't exist, which is what we want - continue

			// Generate rule content
			ruleContent, err := generateRuleContent(prompt, name, description, whenToUse, appliesTo)
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

// generateRuleContent generates rule content from a prompt using Claude CLI.
// If Claude CLI is not available, falls back to simple template-based generation.
func generateRuleContent(prompt, name, description, whenToUse string, appliesTo []string) (string, error) {
	// Try to use Claude CLI if available
	if _, err := exec.LookPath("claude"); err == nil {
		return generateRuleContentWithAgent(prompt, name, description, whenToUse, appliesTo)
	}

	// Fallback to template-based generation
	return generateRuleContentTemplate(prompt, name, description, whenToUse, appliesTo)
}

// generateRuleContentWithAgent spawns Claude CLI to generate rule content.
func generateRuleContentWithAgent(prompt, name, description, whenToUse string, appliesTo []string) (string, error) {
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

	fullPrompt := fmt.Sprintf("%s\n\n%s", systemPrompt, userPrompt)

	fmt.Print("  → Generating rule content with Claude CLI...")

	cmdCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "claude", "--print", "--output-format", "text", fullPrompt) //nolint:gosec // Claude CLI is user-controlled
	cmd.Env = os.Environ()

	output, err := cmd.Output()
	if err != nil {
		// If Claude CLI fails, fall back to template
		fmt.Println(" (failed, using template)")
		return generateRuleContentTemplate(prompt, name, description, whenToUse, appliesTo)
	}

	ruleContent := strings.TrimSpace(string(output))
	if ruleContent == "" {
		// Empty output, fall back to template
		fmt.Println(" (empty output, using template)")
		return generateRuleContentTemplate(prompt, name, description, whenToUse, appliesTo)
	}

	fmt.Println(" (done)")
	return ruleContent, nil
}

// generateRuleContentTemplate generates rule content using a simple template.
func generateRuleContentTemplate(prompt, name, description, whenToUse string, appliesTo []string) (string, error) {
	// If name not provided, extract from prompt
	if name == "" {
		words := strings.Fields(prompt)
		if len(words) > 0 {
			// Capitalize first letter of first word
			first := strings.ToLower(words[0])
			if len(first) > 0 {
				first = strings.ToUpper(first[:1]) + first[1:]
			}
			name = first
			if len(words) > 1 {
				// Capitalize first letter of second word
				second := strings.ToLower(words[1])
				if len(second) > 0 {
					second = strings.ToUpper(second[:1]) + second[1:]
				}
				name += " " + second
			}
		}
		if name == "" {
			name = "New Rule"
		}
	}

	// If description not provided, use prompt as description
	if description == "" {
		description = prompt
	}

	// If when-to-use not provided, generate a default
	if whenToUse == "" {
		whenToUse = fmt.Sprintf("When %s", strings.ToLower(prompt))
	}

	// Build frontmatter
	var frontmatter strings.Builder
	frontmatter.WriteString("---\n")
	frontmatter.WriteString(fmt.Sprintf("name: \"%s\"\n", name))
	frontmatter.WriteString(fmt.Sprintf("description: \"%s\"\n", description))
	frontmatter.WriteString(fmt.Sprintf("when-to-use: \"%s\"\n", whenToUse))
	
	if len(appliesTo) > 0 {
		appliesToStr := strings.Join(appliesTo, ", ")
		frontmatter.WriteString(fmt.Sprintf("applies-to: [%s]\n", appliesToStr))
	}
	
	frontmatter.WriteString("priority: 2\n")
	frontmatter.WriteString("tags: []\n")
	frontmatter.WriteString("version: \"1.0.0\"\n")
	frontmatter.WriteString("---\n\n")

	// Build body content
	var body strings.Builder
	body.WriteString(fmt.Sprintf("## %s\n\n", name))
	body.WriteString(fmt.Sprintf("%s\n\n", description))
	body.WriteString("## Guidelines\n\n")
	body.WriteString("When working with this rule:\n\n")
	body.WriteString(fmt.Sprintf("- %s\n", prompt))
	body.WriteString("\n## Examples\n\n")
	body.WriteString("(Add examples here)\n")

	return frontmatter.String() + body.String(), nil
}
