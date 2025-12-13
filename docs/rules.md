# Rules Command Documentation

Complete guide to the `agentctl rules` command tree for managing agent rules.

## Overview

The rules system provides a modular approach to managing agent instructions. Rules are stored in `.agent/rules/` as `.mdc` files (Markdown with frontmatter) and can be synced to different formats for different tools.

## Directory Structure

```
.agent/
├── rules/              # Rule files (.mdc format with frontmatter)
│   ├── git-workflow.mdc
│   ├── tool-selection.mdc
│   └── ...
├── research/           # Research artifacts (cached findings)
│   └── YYYY-MM-DD-topic.md
└── project.md          # High-level repository description
```

## Environment Variable

The `AGENTDIR` environment variable can override the default `.agent` location:

```bash
export AGENTDIR=/path/to/custom/agent/dir
agentctl rules init  # Creates /path/to/custom/agent/dir/
```

## Rule File Format (.mdc)

Rules are Markdown files with YAML frontmatter. The `.mdc` extension indicates "Markdown with frontmatter".

### Frontmatter Schema

#### Required Fields

- `name`: Human-readable rule name
- `description`: One-line description of what this rule covers
- `when-to-use`: Context for when this rule applies

#### Optional Fields

- `applies-to`: List of tools this rule applies to (default: all tools)
  - Examples: `["claude"]`, `["claude", "cursor", "windsurf"]`
- `priority`: 0-4, where 0 is highest priority (default: 2)
- `tags`: Array of tags for categorization
  - Examples: `["git", "workflow"]`, `["tools", "best-practices"]`
- `version`: Semantic version for rule evolution (default: "1.0.0")
- `depends-on`: Array of rule names this rule depends on
- `conflicts-with`: Array of rule names this rule conflicts with

### Example Rule File

```yaml
---
name: "Git Workflow"
description: "Conventional commits, branch management, and PR workflows"
when-to-use: "When committing changes, creating branches, or managing pull requests"
applies-to: ["claude", "cursor", "windsurf"]
priority: 1
tags: ["git", "workflow", "conventional-commits"]
version: "1.0.0"
---

## Conventional Commits

ALWAYS follow the Conventional Commit Messages specification...

[Rule body content in Markdown]
```

## Commands

### init

Initialize `.agent/` directory with default rules.

```bash
agentctl rules init [--force] [--no-project]
```

**Flags:**
- `--force` - Overwrite existing files
- `--no-project` - Skip project.md generation

**What it does:**
- Creates `.agent/rules/` directory
- Copies default rules from agentctl's embedded rules
- Creates `.agent/research/` directory
- Generates `.agent/project.md` using Claude CLI (unless `--no-project`)

### list

List all rules with metadata from frontmatter.

```bash
agentctl rules list [--json]
```

**Flags:**
- `--json` - Output as JSON for programmatic access

**Output:**
- Rule name, description, when-to-use, applies-to
- File name for each rule

### show

Display full rule content including frontmatter and body.

```bash
agentctl rules show <rule-name> [--raw]
```

**Arguments:**
- `rule-name` - Rule name (case-insensitive) or filename

**Flags:**
- `--raw` - Output raw mdc file without pretty-printing

**Examples:**
```bash
agentctl rules show git-workflow
agentctl rules show "Git Workflow"
agentctl rules show git-workflow.mdc
```

### add

Add a new rule from a prompt description.

```bash
agentctl rules add "<description>" [--name <filename>] [--description <text>] [--when-to-use <text>] [--applies-to <tools>]
```

**Arguments:**
- `description` - Prompt describing what the rule should do

**Flags:**
- `--name` - Filename for the rule (without .mdc extension)
- `--description` - Rule description (auto-generated from prompt if not provided)
- `--when-to-use` - When to use this rule (auto-generated if not provided)
- `--applies-to` - Comma-separated list of tools (default: claude)

**Examples:**
```bash
agentctl rules add "Always validate input before processing" --name input-validation
agentctl rules add "Use TypeScript strict mode" --applies-to "claude,cursor"
```

### remove

Remove rule files from `.agent/rules/`.

```bash
agentctl rules remove <rule-name...> [--force]
```

**Arguments:**
- `rule-name...` - One or more rule names or filenames

**Flags:**
- `--force` - Skip confirmation prompt

**Examples:**
```bash
agentctl rules remove git-workflow
agentctl rules remove tool-selection workspace-management --force
```

### sync

Sync rules to different formats and locations.

```bash
agentctl rules sync [--cursor] [--claude] [--agents] [--claude-md]
```

**Flags:**
- `--cursor` - Copy `.agent/rules/*.mdc` to `.cursor/rules/`
- `--claude` - Convert rules to `.claude/skills/<name>/SKILL.md`
- `--agents` - Generate `AGENTS.md` table of contents
- `--claude-md` - Generate `CLAUDE.md` overview

If no flags are specified, syncs to all formats.

**What each format does:**

- **Cursor**: Copies `.mdc` files directly - Cursor automatically loads them
- **Claude Skills**: Converts each rule to a skill directory with `SKILL.md` - Claude Code automatically loads skills
- **AGENTS.md**: Generates a markdown table listing all rules for non-Cursor/Claude agents
- **CLAUDE.md**: Generates a simple overview with project.md content and skills list

## Integration Points

### Cursor Integration

Rules sync to `.cursor/rules/*.mdc` - Cursor automatically loads these files.

### Claude Code Integration

Rules sync to `.claude/skills/<name>/SKILL.md` - Each rule becomes a Claude skill that agents can invoke.

### AGENTS.md Generation

For non-Cursor/Claude agents, `agentctl rules sync --agents` generates an AGENTS.md file with a table of contents listing all rules.

### CLAUDE.md Generation

For Claude Code, `agentctl rules sync --claude-md` generates a simple CLAUDE.md that references available skills.

## Best Practices

1. **Keep rules focused**: Each rule should cover a single topic or workflow
2. **Use descriptive names**: Rule names should clearly indicate what they cover
3. **Document when-to-use**: Help agents understand when to apply each rule
4. **Tag appropriately**: Use tags to categorize and filter rules
5. **Version rules**: Use version field to track rule evolution
6. **Sync regularly**: Run `agentctl rules sync` after adding or modifying rules

## Examples

### Creating a Custom Rule

```bash
# Add a new rule
agentctl rules add "Always run tests before committing" \
  --name pre-commit-tests \
  --applies-to "claude,cursor"

# Edit the rule file to refine content
vim .agent/rules/pre-commit-tests.mdc

# Sync to all formats
agentctl rules sync
```

### Listing and Viewing Rules

```bash
# List all rules
agentctl rules list

# List as JSON
agentctl rules list --json

# Show a specific rule
agentctl rules show git-workflow

# Show raw mdc file
agentctl rules show git-workflow --raw
```

### Syncing to Specific Formats

```bash
# Sync only to Cursor
agentctl rules sync --cursor

# Sync to Claude skills and generate AGENTS.md
agentctl rules sync --claude --agents

# Sync everything
agentctl rules sync
```

## Migration from Memory Commands

If you're migrating from the old `agentctl memory` commands:

1. Run `agentctl rules init` to create `.agent/` directory
2. Review default rules in `.agent/rules/`
3. Run `agentctl rules sync` to generate AGENTS.md and CLAUDE.md
4. Remove old AGENTS.md/CLAUDE.md if desired
5. Use `agentctl rules` commands going forward

See [Migration Guide](migration-memory-to-rules.md) for detailed steps.
