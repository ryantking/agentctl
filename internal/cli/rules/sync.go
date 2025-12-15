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
	var cursor, claude, agents, claudeMD, force, dryRun bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync rules to different formats and locations",
		Long: `Sync rules to different formats and locations:
- Cursor: Copies .agent/rules/*.mdc to .cursor/rules/
- Claude: Converts rules to .claude/skills/<name>/SKILL.md
- AGENTS.md: Generates table of contents listing all rules
- CLAUDE.md: Generates simple CLAUDE.md with project overview

If no flags are specified, syncs to all formats.

Use --dry-run to preview changes without writing files.`,
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

			if dryRun {
				fmt.Println("DRY RUN MODE - No files will be written")
				fmt.Println()
			}

			var errors []string

			if cursor {
				if err := syncToCursor(repoRoot, rulesDir, dryRun); err != nil {
					errors = append(errors, fmt.Sprintf("Cursor sync: %v", err))
				}
			}

			if claude {
				if err := syncToClaudeSkills(repoRoot, rulesDir, force, dryRun); err != nil {
					errors = append(errors, fmt.Sprintf("Claude skills sync: %v", err))
				}
			}

			if agents {
				if err := syncToAGENTSMD(repoRoot, agentDir, rulesDir, dryRun); err != nil {
					errors = append(errors, fmt.Sprintf("AGENTS.md sync: %v", err))
				}
			}

			if claudeMD {
				if err := syncToCLAUDEMD(repoRoot, agentDir, rulesDir, dryRun); err != nil {
					errors = append(errors, fmt.Sprintf("CLAUDE.md sync: %v", err))
				}
			}

			if len(errors) > 0 {
				for _, errMsg := range errors {
					fmt.Printf("  ✗ %s\n", errMsg)
				}
				return fmt.Errorf("some sync operations failed")
			}

			if dryRun {
				fmt.Println("\n✓ Dry run completed - no files were written")
			} else {
				fmt.Println("\n✓ Sync completed successfully")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&cursor, "cursor", false, "Sync to .cursor/rules/")
	cmd.Flags().BoolVar(&claude, "claude", false, "Sync to .claude/skills/")
	cmd.Flags().BoolVar(&agents, "agents", false, "Generate AGENTS.md table of contents")
	cmd.Flags().BoolVar(&claudeMD, "claude-md", false, "Generate CLAUDE.md")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing Claude skills (skipped by default)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be changed without writing files")

	return cmd
}

// syncToCursor copies .agent/rules/*.mdc to .cursor/rules/.
func syncToCursor(repoRoot, rulesDir string, dryRun bool) error {
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

		relPath, _ := filepath.Rel(repoRoot, destPath)

		if dryRun {
			// Check if file exists and show diff
			if existingData, err := os.ReadFile(destPath); err == nil { //nolint:gosec // Reading for comparison
				if string(existingData) == string(data) {
					fmt.Printf("  • %s (unchanged)\n", relPath)
				} else {
					fmt.Printf("  • %s (would be updated)\n", relPath)
				}
			} else {
				fmt.Printf("  • %s (would be created)\n", relPath)
			}
		} else {
			// Write to destination
			if err := os.WriteFile(destPath, data, 0644); err != nil { //nolint:gosec // Rule files need to be readable
				return fmt.Errorf("failed to write rule file %s: %w", entry.Name(), err)
			}
			fmt.Printf("  • %s (synced)\n", relPath)
		}
		copied++
	}

	fmt.Printf("  → Synced %d rule(s) to Cursor\n", copied)
	return nil
}

// syncToClaudeSkills converts each .agent/rules/*.mdc to .claude/skills/<name>/SKILL.md.
func syncToClaudeSkills(repoRoot, rulesDir string, force, dryRun bool) error {
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
		relPath, _ := filepath.Rel(repoRoot, skillMDPath)

		// Check if skill already exists
		existingData, err := os.ReadFile(skillMDPath) //nolint:gosec // Path is validated to be within repository root
		exists := err == nil

		switch {
		case exists:
			// Skill exists
			if !force && !dryRun {
				fmt.Printf("  ⚠ %s (skipped, already exists, use --force to overwrite)\n", relPath)
				continue
			}
			if dryRun {
				if string(existingData) == skillContent {
					fmt.Printf("  • %s (unchanged)\n", relPath)
				} else {
					fmt.Printf("  • %s (would be overwritten)\n", relPath)
				}
				converted++
				continue
			}
			// Force overwrite
			fmt.Printf("  • %s (overwritten)\n", relPath)
		case os.IsNotExist(err):
			// Skill doesn't exist
			if dryRun {
				fmt.Printf("  • %s (would be created)\n", relPath)
				converted++
				continue
			}
			fmt.Printf("  • %s (synced)\n", relPath)
		default:
			// Other error checking file
			return fmt.Errorf("failed to check skill file %s: %w", skillName, err)
		}

		if !dryRun {
			if err := os.WriteFile(skillMDPath, []byte(skillContent), 0644); err != nil { //nolint:gosec // Skill file needs to be readable
				return fmt.Errorf("failed to write skill file %s: %w", skillName, err)
			}
		}
		converted++
	}

	fmt.Printf("  → Converted %d rule(s) to Claude skills\n", converted)
	return nil
}

// syncToAGENTSMD generates AGENTS.md table of contents listing all rules.
func syncToAGENTSMD(repoRoot, agentDir, rulesDir string, dryRun bool) error {
	fmt.Println("Generating AGENTS.md...")

	// Read project.md if it exists
	projectMDPath := filepath.Join(agentDir, "project.md")
	var projectContent string
	if data, err := os.ReadFile(projectMDPath); err == nil { //nolint:gosec // Reading project.md
		projectContent = strings.TrimSpace(string(data))
	}
	// If project.md doesn't exist, projectContent remains empty and is skipped below

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
		content.WriteString("---\n\n")
	}

	content.WriteString("## Rules System\n\n")
	content.WriteString("This repository uses a modular rules system. Rules are stored in `.agent/rules/` as `.mdc` files (Markdown with frontmatter) and can be synced to different formats for different tools.\n\n")
	content.WriteString("### Managing Rules\n\n")
	content.WriteString("- **List rules**: `agentctl rules list`\n")
	content.WriteString("- **View rule**: `agentctl rules show <name>`\n")
	content.WriteString("- **Add rule**: `agentctl rules add \"<description>\"`\n")
	content.WriteString("- **Sync rules**: `agentctl rules sync` (generates this file)\n\n")

	if len(ruleInfos) == 0 {
		content.WriteString("## Available Rules\n\n")
		content.WriteString("No rules found. Run `agentctl rules init` to initialize with default rules.\n")
	} else {
		// Group rules by priority
		priorityGroups := make(map[int][]RuleInfo)
		for _, rule := range ruleInfos {
			priority := rule.Metadata.Priority
			priorityGroups[priority] = append(priorityGroups[priority], rule)
		}

		content.WriteString("## Available Rules\n\n")
		content.WriteString("Rules are organized by priority (0 = highest, 4 = lowest).\n\n")

		// Sort priorities and output groups
		priorities := []int{0, 1, 2, 3, 4}
		for _, priority := range priorities {
			rules, hasRules := priorityGroups[priority]
			if !hasRules || len(rules) == 0 {
				continue
			}

			var priorityLabel string
			switch priority {
			case 0:
				priorityLabel = "Priority 0 (Critical)"
			case 1:
				priorityLabel = "Priority 1 (High)"
			case 2:
				priorityLabel = "Priority 2 (Medium)"
			case 3:
				priorityLabel = "Priority 3 (Low)"
			case 4:
				priorityLabel = "Priority 4 (Backlog)"
			default:
				priorityLabel = fmt.Sprintf("Priority %d", priority)
			}

			content.WriteString(fmt.Sprintf("### %s\n\n", priorityLabel))
			content.WriteString("| Rule | Description | When to Use | Tags |\n")
			content.WriteString("|------|-------------|-------------|------|\n")

			for _, rule := range rules {
				name := rule.Metadata.Name
				desc := rule.Metadata.Description
				whenToUse := rule.Metadata.WhenToUse
				tags := strings.Join(rule.Metadata.Tags, ", ")
				if tags == "" {
					tags = "-"
				}

				// Escape pipe characters in markdown table
				name = strings.ReplaceAll(name, "|", "\\|")
				desc = strings.ReplaceAll(desc, "|", "\\|")
				whenToUse = strings.ReplaceAll(whenToUse, "|", "\\|")
				tags = strings.ReplaceAll(tags, "|", "\\|")

				content.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", name, desc, whenToUse, tags))
			}
			content.WriteString("\n")
		}

		content.WriteString("### Viewing Rules\n\n")
		content.WriteString("To view the full content of a rule:\n\n")
		content.WriteString("```bash\n")
		content.WriteString("agentctl rules show <rule-name>\n")
		content.WriteString("```\n\n")
		content.WriteString("For more details, see `.agent/rules/` directory.\n")
	}

	// Write AGENTS.md
	agentsMDPath := filepath.Join(repoRoot, "AGENTS.md")
	contentStr := content.String()

	if dryRun {
		// Check if file exists and show diff
		if existingData, err := os.ReadFile(agentsMDPath); err == nil { //nolint:gosec // Reading for comparison
			if string(existingData) == contentStr {
				fmt.Printf("  • AGENTS.md (unchanged)\n")
			} else {
				fmt.Printf("  • AGENTS.md (would be updated)\n")
			}
		} else {
			fmt.Printf("  • AGENTS.md (would be created)\n")
		}
		return nil
	}

	if err := os.WriteFile(agentsMDPath, []byte(contentStr), 0644); err != nil { //nolint:gosec // AGENTS.md needs to be readable
		return fmt.Errorf("failed to write AGENTS.md: %w", err)
	}

	fmt.Printf("  • AGENTS.md (generated)\n")
	return nil
}

// syncToCLAUDEMD generates simple CLAUDE.md with project overview and skills list.
func syncToCLAUDEMD(repoRoot, agentDir, rulesDir string, dryRun bool) error {
	fmt.Println("Generating CLAUDE.md...")

	// Read project.md if it exists
	projectMDPath := filepath.Join(agentDir, "project.md")
	var projectContent string
	if data, err := os.ReadFile(projectMDPath); err == nil { //nolint:gosec // Reading project.md
		projectContent = strings.TrimSpace(string(data))
	}
	// If project.md doesn't exist, projectContent remains empty and is skipped below

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
		content.WriteString("---\n\n")
	}

	content.WriteString("## Skills\n\n")
	content.WriteString("Claude Code automatically loads skills from `.claude/skills/`. Each skill provides specialized knowledge or workflows for this repository.\n\n")
	content.WriteString("### Syncing Skills\n\n")
	content.WriteString("To sync rules to Claude skills:\n\n")
	content.WriteString("```bash\n")
	content.WriteString("agentctl rules sync --claude\n")
	content.WriteString("```\n\n")
	content.WriteString("This converts `.agent/rules/*.mdc` files into `.claude/skills/<name>/SKILL.md` format.\n\n")

	if len(ruleInfos) == 0 {
		content.WriteString("### Available Skills\n\n")
		content.WriteString("No skills found. Run `agentctl rules init` to initialize, then `agentctl rules sync --claude` to sync rules to skills.\n")
	} else {
		content.WriteString("### Available Skills\n\n")
		content.WriteString("The following skills are available in this repository:\n\n")

		for i, rule := range ruleInfos {
			content.WriteString(fmt.Sprintf("#### %s\n\n", rule.Metadata.Name))
			content.WriteString(fmt.Sprintf("%s\n\n", rule.Metadata.Description))
			
			if rule.Metadata.WhenToUse != "" {
				content.WriteString(fmt.Sprintf("**When to use:** %s\n\n", rule.Metadata.WhenToUse))
			}

			if len(rule.Metadata.AppliesTo) > 0 {
				content.WriteString(fmt.Sprintf("**Applies to:** %s\n\n", strings.Join(rule.Metadata.AppliesTo, ", ")))
			}

			if len(rule.Metadata.Tags) > 0 {
				content.WriteString(fmt.Sprintf("**Tags:** %s\n\n", strings.Join(rule.Metadata.Tags, ", ")))
			}

			// Add separator between skills (except last one)
			if i < len(ruleInfos)-1 {
				content.WriteString("---\n\n")
			}
		}

		content.WriteString("### Managing Skills\n\n")
		content.WriteString("- **View skill**: `agentctl rules show <skill-name>`\n")
		content.WriteString("- **List all skills**: `agentctl rules list`\n")
		content.WriteString("- **Add new skill**: `agentctl rules add \"<description>\"`\n")
		content.WriteString("- **Sync all skills**: `agentctl rules sync --claude`\n\n")
		content.WriteString("Skills are automatically loaded by Claude Code when present in `.claude/skills/`.\n")
	}

	// Write CLAUDE.md
	claudeMDPath := filepath.Join(repoRoot, "CLAUDE.md")
	contentStr := content.String()

	if dryRun {
		// Check if file exists and show diff
		if existingData, err := os.ReadFile(claudeMDPath); err == nil { //nolint:gosec // Reading for comparison
			if string(existingData) == contentStr {
				fmt.Printf("  • CLAUDE.md (unchanged)\n")
			} else {
				fmt.Printf("  • CLAUDE.md (would be updated)\n")
			}
		} else {
			fmt.Printf("  • CLAUDE.md (would be created)\n")
		}
		return nil
	}

	if err := os.WriteFile(claudeMDPath, []byte(contentStr), 0644); err != nil { //nolint:gosec // CLAUDE.md needs to be readable
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
