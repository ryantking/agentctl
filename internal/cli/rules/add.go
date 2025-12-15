// Package rules provides commands for managing agent rules in the .agent directory.
package rules

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/shared/constant"
	anthclient "github.com/ryantking/agentctl/internal/claude"
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

// generateRuleContent generates rule content from a prompt using Anthropic SDK.
func generateRuleContent(prompt, name, description, whenToUse string, appliesTo []string) (string, error) {
	// Check if API key is configured
	if !anthclient.IsConfigured() {
		return "", anthclient.EnhanceSDKError(fmt.Errorf("ANTHROPIC_API_KEY environment variable not set"))
	}

	client, err := anthclient.NewClient()
	if err != nil {
		return "", anthclient.EnhanceSDKError(err)
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

	fmt.Print("  → Generating rule content with Anthropic SDK...")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create message request
	params := anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_5,
		MaxTokens: 4000,
		System: []anthropic.TextBlockParam{
			{
				Text: systemPrompt,
				Type: constant.Text("text"),
			},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.ContentBlockParamUnion{
				OfText: &anthropic.TextBlockParam{
					Text: userPrompt,
					Type: constant.Text("text"),
				},
			}),
		},
	}

	// Call Messages API
	msg, err := client.Messages.New(ctx, params)
	if err != nil {
		return "", anthclient.EnhanceSDKError(fmt.Errorf("failed to generate rule content: %w", err))
	}

	// Extract text content from response
	var ruleContent strings.Builder
	for _, block := range msg.Content {
		if block.Type == "text" && block.Text != "" {
			ruleContent.WriteString(block.Text)
		}
	}

	content := strings.TrimSpace(ruleContent.String())
	if content == "" {
		return "", fmt.Errorf("empty output from Anthropic API")
	}

	fmt.Println(" (done)")
	return content, nil
}

