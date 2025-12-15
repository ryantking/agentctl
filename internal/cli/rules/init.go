package rules

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	anthclient "github.com/ryantking/agentctl/internal/anthropic"
	"github.com/ryantking/agentctl/internal/git"
	"github.com/ryantking/agentctl/internal/output"
	"github.com/ryantking/agentctl/internal/rules"
	"github.com/spf13/cobra"
)

// InitRules initializes the .agent directory with default rules.
// This function is exported so it can be called from other packages.
func InitRules(repoRoot string, force, noProject bool) error {
	// Determine .agent directory location (AGENTDIR env var or default)
	agentDir, err := getAgentDir(repoRoot)
	if err != nil {
		return fmt.Errorf("invalid AGENTDIR: %w", err)
	}

	fmt.Println("Initializing .agent directory...")

	// Create .agent/rules/ directory
	rulesDir := filepath.Join(agentDir, "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil { //nolint:gosec // Rules directory needs to be readable
		return fmt.Errorf("failed to create rules directory: %w", err)
	}

	// Copy default rules from embedded rules/ directory
	if err := copyDefaultRules(rulesDir, force); err != nil {
		return err
	}

	// Create .agent/research/ directory
	researchDir := filepath.Join(agentDir, "research")
	if err := os.MkdirAll(researchDir, 0755); err != nil { //nolint:gosec // Research directory needs to be readable
		return fmt.Errorf("failed to create research directory: %w", err)
	}
	relResearchPath, _ := filepath.Rel(repoRoot, researchDir)
	fmt.Printf("  • %s (created)\n", relResearchPath)

	// Generate project.md unless noProject flag
	if !noProject {
		if err := generateProjectMD(agentDir, repoRoot, force); err != nil {
			// Non-fatal: warn but continue
			fmt.Printf("  → Project.md generation skipped: %v\n", err)
		}
	}

	fmt.Println("\n✓ .agent directory initialized successfully")
	return nil
}

// NewRulesInitCmd creates the rules init command.
func NewRulesInitCmd() *cobra.Command {
	var force, noProject bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize .agent directory with default rules",
		Long: `Initialize .agent directory structure. Copies default rules from agentctl's rules/ directory
to .agent/rules/. Generates .agent/project.md using Anthropic SDK. Creates .agent/research/ directory.

Respects AGENTDIR environment variable (defaults to .agent).`,
		RunE: func(_ *cobra.Command, _ []string) error {
			var repoRoot string
			var err error

			repoRoot, err = git.GetRepoRoot()
			if err != nil {
				output.Errorf("%v\n\nRun from inside a git repository", err)
				return err
			}

			return InitRules(repoRoot, force, noProject)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing files")
	cmd.Flags().BoolVar(&noProject, "no-project", false, "Skip project.md generation")

	return cmd
}

// copyDefaultRules copies default rules from embedded rules/ directory to .agent/rules/.
func copyDefaultRules(targetDir string, force bool) error {
	// Read all .mdc files from embedded rules directory
	ruleFiles, err := rules.ReadRulesDir()
	if err != nil {
		return fmt.Errorf("failed to read embedded rules directory: %w", err)
	}

	copied := 0
	for _, ruleFile := range ruleFiles {
		if !strings.HasSuffix(ruleFile, ".mdc") {
			continue
		}

		destPath := filepath.Join(targetDir, ruleFile)

		// Check if file exists and skip if not forcing
		if _, err := os.Stat(destPath); err == nil && !force {
			relPath, _ := filepath.Rel(targetDir, destPath)
			fmt.Printf("  • %s (skipped)\n", relPath)
			continue
		}

		// Read rule from embedded filesystem
		data, err := rules.GetRule(ruleFile)
		if err != nil {
			return fmt.Errorf("failed to read embedded rule file %s: %w", ruleFile, err)
		}

		// Check if file existed before writing
		existed := false
		if _, err := os.Stat(destPath); err == nil {
			existed = true
		}

		// Write to destination
		if err := os.WriteFile(destPath, data, 0644); err != nil { //nolint:gosec // Rule files need to be readable
			return fmt.Errorf("failed to write rule file %s: %w", ruleFile, err)
		}

		// Determine status for output
		relPath, _ := filepath.Rel(targetDir, destPath)
		status := "created"
		if existed {
			status = "overwritten"
		}
		fmt.Printf("  • %s (%s)\n", relPath, status)
		copied++
	}

	if copied == 0 {
		fmt.Printf("  → No rules copied (all already exist, use --force to overwrite)\n")
	} else {
		fmt.Printf("  → Copied %d rule(s)\n", copied)
	}

	return nil
}

// generateProjectMD generates .agent/project.md using Anthropic SDK with tool use.
func generateProjectMD(agentDir, repoRoot string, force bool) error {
	projectMDPath := filepath.Join(agentDir, "project.md")

	// Check if file exists and skip if not forcing
	if _, err := os.Stat(projectMDPath); err == nil && !force {
		fmt.Printf("  • project.md (skipped)\n")
		return nil
	}

	// Check if API key is configured
	if !anthclient.IsConfigured() {
		return fmt.Errorf(`ANTHROPIC_API_KEY environment variable not set

To fix this:
  - Set ANTHROPIC_API_KEY environment variable: export ANTHROPIC_API_KEY=your-key
  - Or run 'claude login' if you have the Claude CLI installed
  - Get your API key at https://console.anthropic.com/`)
	}

	client, err := anthclient.NewClient()
	if err != nil {
		return anthclient.EnhanceSDKError(err)
	}

	// Create tool registry with repository exploration tools
	registry := NewToolRegistry()
	if err := RegisterRepoTools(registry, repoRoot); err != nil {
		return fmt.Errorf("failed to register repository tools: %w", err)
	}

	// Create conversation with tool use support
	conv := NewConversation(client, registry)

	prompt := `Analyze this repository and provide a concise overview:
- Main purpose and key technologies
- Directory structure (2-3 levels max)
- Entry points and main files
- Build/run commands (check for package.json scripts, Makefile targets, Justfile recipes, etc.)
- Available scripts and automation tools

Use the list_directory and read_file tools to explore the repository structure and key files.

Format as clean markdown starting at heading level 2 (##), keep it brief (under 500 words).`

	fmt.Print("  → Generating project.md with Anthropic SDK (using tool use)...")

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Add user message and send (tool use is handled automatically)
	conv.AddUserMessage(prompt)
	content, err := conv.Send(ctx, anthropic.ModelClaudeSonnet4_5, 2000)
	if err != nil {
		return anthclient.EnhanceSDKError(fmt.Errorf("failed to generate project.md: %w", err))
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return fmt.Errorf("empty output from Anthropic API")
	}

	// Write project.md
	if err := os.WriteFile(projectMDPath, []byte(content), 0644); err != nil { //nolint:gosec // Project file needs to be readable
		return fmt.Errorf("failed to write project.md: %w", err)
	}

	fmt.Println(" done")
	fmt.Printf("  • project.md (created)\n")

	return nil
}
