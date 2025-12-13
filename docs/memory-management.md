# Memory Management Guide

This guide explains how to manage agent memory files (AGENTS.md and CLAUDE.md) using `agentctl memory` commands.

## Philosophy

### Universal vs Tool-Specific Instructions

Agent memory is separated into two categories:

**AGENTS.md (Universal Instructions):**
- Cross-platform standard supported by GitHub Copilot, Cursor, Windsurf, and Claude Code
- Contains rules, workflows, and guidelines that apply to any AI coding assistant
- Examples: Git workflow, coding standards, project structure, build commands

**CLAUDE.md (Claude Code-Specific):**
- Contains Claude Code-specific orchestration patterns
- Imports AGENTS.md using `@AGENTS.md` syntax
- Examples: Agent selection, workflow phases, tool access patterns

### Why Separate?

- **Reusability**: AGENTS.md can be used across different AI tools
- **Maintainability**: Smaller, focused files are easier to maintain
- **Clarity**: Clear separation between universal and tool-specific patterns
- **Standards**: Follows emerging AGENTS.md standard for cross-platform compatibility

## File Structure

```
project-root/
├── AGENTS.md              # Universal agent instructions (cross-tool)
├── CLAUDE.md              # Claude Code-specific instructions + @AGENTS.md
├── .claude/
│   ├── agents/            # Agent definitions (historian, engineer, etc.)
│   ├── skills/            # Skill templates
│   └── research/          # Research cache
└── .beads/
    ├── beads.db           # SQLite issue database
    └── issues.jsonl       # Git-synced issue export
```

## Import Syntax

Claude Code supports importing other files using `@path/to/file` syntax:

```markdown
# CLAUDE.md
See @AGENTS.md for universal agent instructions.

## Claude Code-Specific Content
[Agent orchestration patterns here]
```

**Rules:**
- Imports use `@path/to/file` syntax
- Both relative and absolute paths supported
- Recursive imports allowed (max depth: 5 hops)
- Imports are resolved when files are loaded

## Line Count Recommendations

**AGENTS.md:**
- Target: ~250 lines (leaves room for project-specific additions)
- Maximum: 300 lines (adherence degrades beyond this)
- If exceeding: Split into modular files in `.claude/rules/`

**CLAUDE.md:**
- Target: ~150 lines (focused on orchestration only)
- Maximum: 200 lines
- Should be minimal since it imports AGENTS.md

**Why these limits?**
- LLMs reliably follow ~150-200 instructions
- Beyond 300 lines, adherence degrades significantly
- Smaller files are easier to maintain and update

## Best Practices

### DO:
- Keep memory files lean (<300 lines for AGENTS.md)
- Be specific: "Use 2-space indentation" not "Format code properly"
- Use structure: bullet points grouped under markdown headings
- Import additional files rather than duplicating content
- Separate universal agent rules from tool-specific features
- Include the WHY not just the WHAT
- Test and iterate like prompt engineering

### DON'T:
- Include generic instructions ("write clean code")
- Put style rules in memory (use linters instead)
- Mix task-specific with universal instructions
- Exceed LLM instruction limits (~150-200 reliable)
- Duplicate content across files
- Include secrets or credentials

## Beads Integration

This project uses [bd (beads)](https://github.com/steveyegge/beads) for issue tracking.

**Key Points:**
- Use `bd` commands instead of markdown TODOs
- All task tracking happens in beads database
- Run `bd onboard` to get started
- AGENTS.md references beads at the end: "BEFORE ANYTHING ELSE: run 'bd onboard'"

**Why beads?**
- Dependency-aware: Track blockers and relationships
- Git-friendly: Auto-syncs to JSONL for version control
- Agent-optimized: JSON output, ready work detection
- Prevents duplicate tracking systems

## Commands

### Initialize Memory Files

```bash
# Initialize AGENTS.md and CLAUDE.md from templates
agentctl memory init

# With options
agentctl memory init --force          # Overwrite existing files
agentctl memory init --global         # Install to $HOME/.claude
agentctl memory init --no-index       # Skip repository indexing
```

### Show Memory Files

```bash
# Show both files
agentctl memory show

# Show specific file
agentctl memory show AGENTS.md

# Resolve imports inline
agentctl memory show --resolve

# JSON output with metadata
agentctl memory show --json
```

### Validate Memory Files

```bash
# Check for common issues
agentctl memory validate
```

Validates:
- Line counts (warns if AGENTS.md > 300, CLAUDE.md > 200)
- Missing @AGENTS.md import in CLAUDE.md
- Circular imports (max 5 hops)
- Conflicting patterns (e.g., .claude/plans references)
- Required sections in AGENTS.md

### Index Repository

```bash
# Generate repository overview and inject into AGENTS.md
agentctl memory index

# With custom timeout
agentctl memory index --timeout 120
```

Requires Claude CLI to be installed. Generates markdown overview and inserts it between `<!-- REPOSITORY_INDEX_START -->` and `<!-- REPOSITORY_INDEX_END -->` markers in AGENTS.md.

## Migration from Single CLAUDE.md

If you have an existing CLAUDE.md with all content:

1. **Run memory init** to create new structure:
   ```bash
   agentctl memory init --force
   ```

2. **Review the templates** - they separate universal from Claude-specific content

3. **Migrate custom content** - Add project-specific rules to AGENTS.md

4. **Verify** - Run `agentctl memory validate` to check for issues

## Maintenance

### Regular Review
- Review monthly for outdated instructions
- Remove rules enforced by tooling (linters, formatters)
- Keep under line count recommendations

### Effectiveness Testing
- Treat as prompt engineering - iterate on instructions
- Add emphasis keywords ("IMPORTANT", "MUST") for critical rules
- Test with real prompts, observe agent behavior

### Version Control
- Commit AGENTS.md and CLAUDE.md
- Keep CLAUDE.local.md in .gitignore (if using)
- Document changes in commit messages

## Related Resources

- [AGENTS.md Standard](https://www.builder.io/blog/agents-md)
- [Claude Code Memory Docs](https://code.claude.com/docs/en/memory)
- [Beads Documentation](https://github.com/steveyegge/beads)
- Research: `.claude/research/2025-12-12-ai-agent-memory-management.md`
