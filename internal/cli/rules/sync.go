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

// NewRulesSyncCmd creates the rules sync command.
func NewRulesSyncCmd() *cobra.Command {
	var cursor, claude, agents, claudeMD bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync rules to different formats and locations",
		Long: `Sync rules to different formats and locations:
- Cursor: Copies .agent/rules/*.mdc to .cursor/rules/
- Claude: Converts rules to .claude/skills/<name>/SKILL.md
- AGENTS.md: Generates table of contents listing all rules
- CLAUDE.md: Generates simple CLAUDE.md with project overview

If no flags are specified, syncs to all formats.`,
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

			// Check if rules directory exists
			if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
				return fmt.Errorf("rules directory not found: %s\n\nRun 'agentctl rules init' to initialize", rulesDir)
			}

			// If no flags specified, sync to all formats
			if !cursor && !claude && !agents && !claudeMD {
				cursor = true
				claude = true
				agents = true
				claudeMD = true
			}

			var errors []string

			if cursor {
				if err := syncToCursor(repoRoot, rulesDir); err != nil {
					errors = append(errors, fmt.Sprintf("Cursor sync: %v", err))
				}
			}

			if claude {
				if err := syncToClaudeSkills(repoRoot, rulesDir); err != nil {
					errors = append(errors, fmt.Sprintf("Claude skills sync: %v", err))
				}
			}

			if agents {
				if err := syncToAGENTSMD(repoRoot, agentDir, rulesDir); err != nil {
					errors = append(errors, fmt.Sprintf("AGENTS.md sync: %v", err))
				}
			}

			if claudeMD {
				if err := syncToCLAUDEMD(repoRoot, agentDir, rulesDir); err != nil {
					errors = append(errors, fmt.Sprintf("CLAUDE.md sync: %v", err))
				}
			}

			if len(errors) > 0 {
				for _, errMsg := range errors {
					fmt.Printf("  ✗ %s\n", errMsg)
				}
				return fmt.Errorf("some sync operations failed")
			}

			fmt.Println("\n✓ Sync completed successfully")
			return nil
		},
	}

	cmd.Flags().BoolVar(&cursor, "cursor", false, "Sync to .cursor/rules/")
	cmd.Flags().BoolVar(&claude, "claude", false, "Sync to .claude/skills/")
	cmd.Flags().BoolVar(&agents, "agents", false, "Generate AGENTS.md table of contents")
	cmd.Flags().BoolVar(&claudeMD, "claude-md", false, "Generate CLAUDE.md")

	return cmd
}

// syncToCursor copies .agent/rules/*.mdc to .cursor/rules/.
func syncToCursor(repoRoot, rulesDir string) error {
	cursorRulesDir := filepath.Join(repoRoot, ".cursor", "rules")
	
	fmt.Println("Syncing to Cursor...")
	
	// Create .cursor/rules directory
	if err := os.MkdirAll(cursorRulesDir, 0755); err != nil { //nolint:gosec // Rules directory needs to be readable
		return fmt.Errorf("failed to create .cursor/rules directory: %w", err)
	}

	// Read all .mdc files from .agent/rules/
	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		return fmt.Errorf("failed to read rules directory: %w", err)
	}

	copied := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".mdc") {
			continue
		}

		sourcePath := filepath.Join(rulesDir, entry.Name())
		destPath := filepath.Join(cursorRulesDir, entry.Name())

		// Read source file
		data, err := os.ReadFile(sourcePath) //nolint:gosec // Reading from controlled source directory
		if err != nil {
			return fmt.Errorf("failed to read rule file %s: %w", entry.Name(), err)
		}

		// Write to destination
		if err := os.WriteFile(destPath, data, 0644); err != nil { //nolint:gosec // Rule files need to be readable
			return fmt.Errorf("failed to write rule file %s: %w", entry.Name(), err)
		}

		relPath, _ := filepath.Rel(repoRoot, destPath)
		fmt.Printf("  • %s (synced)\n", relPath)
		copied++
	}

	fmt.Printf("  → Synced %d rule(s) to Cursor\n", copied)
	return nil
}

// syncToClaudeSkills converts each .agent/rules/*.mdc to .claude/skills/<name>/SKILL.md.
func syncToClaudeSkills(repoRoot, rulesDir string) error {
	claudeSkillsDir := filepath.Join(repoRoot, ".claude", "skills")
	
	fmt.Println("Syncing to Claude skills...")
	
	// Read all .mdc files from .agent/rules/
	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		return fmt.Errorf("failed to read rules directory: %w", err)
	}

	converted := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".mdc") {
			continue
		}

		rulePath := filepath.Join(rulesDir, entry.Name())
		metadata, err := parseRuleMetadata(rulePath)
		if err != nil {
			// Skip files with invalid frontmatter
			continue
		}

		// Use rule name as skill name (sanitize for filesystem)
		skillName := sanitizeSkillName(metadata.Name)
		if skillName == "" {
			// Fallback to filename without extension
			skillName = strings.TrimSuffix(entry.Name(), ".mdc")
			skillName = sanitizeSkillName(skillName)
		}

		skillDir := filepath.Join(claudeSkillsDir, skillName)
		if err := os.MkdirAll(skillDir, 0755); err != nil { //nolint:gosec // Skill directory needs to be readable
			return fmt.Errorf("failed to create skill directory %s: %w", skillName, err)
		}

		// Read rule content
		ruleData, err := os.ReadFile(rulePath) //nolint:gosec // Reading from controlled source directory
		if err != nil {
			return fmt.Errorf("failed to read rule file %s: %w", entry.Name(), err)
		}

		// Extract body (content after frontmatter)
		body := extractBody(string(ruleData))

		// Create SKILL.md with Claude skill format
		skillContent := fmt.Sprintf(`---
name: %s
description: %s
---

%s`, metadata.Name, metadata.Description, body)

		skillMDPath := filepath.Join(skillDir, "SKILL.md")
		if err := os.WriteFile(skillMDPath, []byte(skillContent), 0644); err != nil { //nolint:gosec // Skill file needs to be readable
			return fmt.Errorf("failed to write skill file %s: %w", skillName, err)
		}

		relPath, _ := filepath.Rel(repoRoot, skillMDPath)
		fmt.Printf("  • %s (synced)\n", relPath)
		converted++
	}

	fmt.Printf("  → Converted %d rule(s) to Claude skills\n", converted)
	return nil
}

// syncToAGENTSMD generates AGENTS.md table of contents listing all rules.
func syncToAGENTSMD(repoRoot, agentDir, rulesDir string) error {
	fmt.Println("Generating AGENTS.md...")

	// Read project.md if it exists
	projectMDPath := filepath.Join(agentDir, "project.md")
	var projectContent string
	if data, err := os.ReadFile(projectMDPath); err == nil { //nolint:gosec // Reading project.md
		projectContent = string(data)
	}

	// List all rules
	ruleInfos, err := listRules(rulesDir)
	if err != nil {
		return fmt.Errorf("failed to list rules: %w", err)
	}

	// Generate AGENTS.md content
	var content strings.Builder
	
	content.WriteString("# Agent Instructions\n\n")
	
	if projectContent != "" {
		content.WriteString(projectContent)
		content.WriteString("\n\n")
	}

	content.WriteString("## Available Rules\n\n")
	content.WriteString("This repository uses a modular rules system. Rules are stored in `.agent/rules/` and can be synced to different formats.\n\n")

	if len(ruleInfos) == 0 {
		content.WriteString("No rules found. Run `agentctl rules init` to initialize.\n")
	} else {
		content.WriteString("| Rule | Description | When to Use |\n")
		content.WriteString("|------|-------------|-------------|\n")
		
		for _, rule := range ruleInfos {
			name := rule.Metadata.Name
			desc := rule.Metadata.Description
			whenToUse := rule.Metadata.WhenToUse
			
			// Escape pipe characters in markdown table
			name = strings.ReplaceAll(name, "|", "\\|")
			desc = strings.ReplaceAll(desc, "|", "\\|")
			whenToUse = strings.ReplaceAll(whenToUse, "|", "\\|")
			
			content.WriteString(fmt.Sprintf("| %s | %s | %s |\n", name, desc, whenToUse))
		}
		
		content.WriteString("\nFor more details, see `.agent/rules/` directory.\n")
	}

	// Write AGENTS.md
	agentsMDPath := filepath.Join(repoRoot, "AGENTS.md")
	if err := os.WriteFile(agentsMDPath, []byte(content.String()), 0644); err != nil { //nolint:gosec // AGENTS.md needs to be readable
		return fmt.Errorf("failed to write AGENTS.md: %w", err)
	}

	fmt.Printf("  • AGENTS.md (generated)\n")
	return nil
}

// syncToCLAUDEMD generates simple CLAUDE.md with project overview and skills list.
func syncToCLAUDEMD(repoRoot, agentDir, rulesDir string) error {
	fmt.Println("Generating CLAUDE.md...")

	// Read project.md if it exists
	projectMDPath := filepath.Join(agentDir, "project.md")
	var projectContent string
	if data, err := os.ReadFile(projectMDPath); err == nil { //nolint:gosec // Reading project.md
		projectContent = string(data)
	}

	// List all rules
	ruleInfos, err := listRules(rulesDir)
	if err != nil {
		return fmt.Errorf("failed to list rules: %w", err)
	}

	// Generate CLAUDE.md content
	var content strings.Builder
	
	content.WriteString("# Claude Code Configuration\n\n")
	
	if projectContent != "" {
		content.WriteString(projectContent)
		content.WriteString("\n\n")
	}

	content.WriteString("## Available Skills\n\n")
	content.WriteString("Claude Code automatically loads skills from `.claude/skills/`. Run `agentctl rules sync --claude` to sync rules to skills.\n\n")

	if len(ruleInfos) == 0 {
		content.WriteString("No rules found. Run `agentctl rules init` to initialize.\n")
	} else {
		for _, rule := range ruleInfos {
			content.WriteString(fmt.Sprintf("### %s\n\n", rule.Metadata.Name))
			content.WriteString(fmt.Sprintf("%s\n\n", rule.Metadata.Description))
			if rule.Metadata.WhenToUse != "" {
				content.WriteString(fmt.Sprintf("**When to use:** %s\n\n", rule.Metadata.WhenToUse))
			}
		}
	}

	// Write CLAUDE.md
	claudeMDPath := filepath.Join(repoRoot, "CLAUDE.md")
	if err := os.WriteFile(claudeMDPath, []byte(content.String()), 0644); err != nil { //nolint:gosec // CLAUDE.md needs to be readable
		return fmt.Errorf("failed to write CLAUDE.md: %w", err)
	}

	fmt.Printf("  • CLAUDE.md (generated)\n")
	return nil
}

// sanitizeSkillName converts a rule name to a valid skill directory name.
func sanitizeSkillName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)
	// Replace spaces and special chars with hyphens
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	// Remove invalid characters
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}
