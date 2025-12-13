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
					filename = sanitizeSkillName(strings.Join(words[:min(3, len(words))], "-"))
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

// generateRuleContent generates rule content from a prompt.
// For now, this is a simple template-based approach. In the future, this could spawn an agent.
func generateRuleContent(prompt, name, description, whenToUse string, appliesTo []string) (string, error) {
	// If name not provided, extract from prompt
	if name == "" {
		words := strings.Fields(prompt)
		if len(words) > 0 {
			name = strings.Title(strings.ToLower(words[0]))
			if len(words) > 1 {
				name += " " + strings.Title(strings.ToLower(words[1]))
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
	body.WriteString(fmt.Sprintf("When working with this rule:\n\n"))
	body.WriteString(fmt.Sprintf("- %s\n", prompt))
	body.WriteString("\n## Examples\n\n")
	body.WriteString("(Add examples here)\n")

	return frontmatter.String() + body.String(), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
