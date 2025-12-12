# Research: AI Agent Memory Management Best Practices 2024-2025
Date: 2025-12-12
Focus: Separating agent-specific vs tool-specific instructions, memory organization, import mechanisms, cross-tool reusability
Agent: researcher

## Summary

The AI coding assistant ecosystem is converging on **AGENTS.md** as a universal standard for agent instructions, while tool-specific files (CLAUDE.md, .cursorrules, .windsurfrules) handle platform-specific features. Best practices emphasize hierarchical memory structures, progressive disclosure of context, and import mechanisms to avoid duplication. The key insight is separating "what any agent needs to know" (AGENTS.md) from "what this specific tool needs" (tool-specific files).

## Key Findings

1. **AGENTS.md is emerging as a cross-platform standard** - Supported by GitHub Copilot, Google Gemini CLI, OpenAI Codex, Cursor, Windsurf, and others (Claude Code notably absent but workaround exists via symlinks) [Source](https://www.builder.io/blog/agents-md)

2. **Claude Code uses a four-tier hierarchical memory system** - Enterprise policy, Project memory (CLAUDE.md), User memory (~/.claude/CLAUDE.md), and Project local memory (CLAUDE.local.md) [Source](https://code.claude.com/docs/en/memory)

3. **Import syntax enables modularity** - Claude Code supports `@path/to/import` syntax with recursive imports (max 5 hops) for composable memory files [Source](https://code.claude.com/docs/en/memory)

4. **Keep core memory under 300 lines** - Research shows LLMs can follow ~150-200 instructions reliably; beyond that, adherence degrades [Source](https://www.humanlayer.dev/blog/writing-a-good-claude-md)

5. **Progressive disclosure prevents context bloat** - Use separate files (e.g., `agent_docs/`) for task-specific instructions, only loading when relevant [Source](https://www.humanlayer.dev/blog/writing-a-good-claude-md)

6. **Tool-specific file locations are standardizing** - Cursor uses `.cursor/rules/*.mdc`, Windsurf uses `.windsurf/rules/`, VS Code uses `.github/copilot-instructions.md` [Source](https://dev.to/idavidov13/one-file-to-rule-them-all-cursor-windsurf-and-vs-code-hh2)

## Detailed Analysis

### 1. Separating Agent-Specific vs Tool-Specific Instructions

The emerging pattern separates instructions into two categories:

**Agent-Specific (AGENTS.md or shared CLAUDE.md):**
- Technology stack and versions
- Project structure and architecture
- Code conventions and patterns
- Testing approaches
- Build/run commands
- Domain-specific rules

**Tool-Specific:**
- Custom slash commands (Claude Code: `.claude/commands/`)
- MCP server configurations (Claude Code: `.mcp.json`)
- Tool-specific syntax or features
- Platform-specific behaviors
- Tool workflow preferences

The migration path is:
1. Put universal agent guidance in AGENTS.md
2. Create symlinks for backward compatibility: `ln -s AGENTS.md CLAUDE.md`
3. Keep tool-specific features in dedicated files

### 2. Memory File Organization and Structure

**Claude Code Hierarchy (highest to lowest precedence):**

| Level | Location | Purpose | Shared With |
|-------|----------|---------|-------------|
| Enterprise | `/Library/Application Support/ClaudeCode/CLAUDE.md` | Org-wide policies | All users |
| Project | `./CLAUDE.md` or `./.claude/CLAUDE.md` | Team instructions | Via git |
| Project Rules | `./.claude/rules/*.md` | Modular topic files | Via git |
| User | `~/.claude/CLAUDE.md` | Personal preferences | Just you |
| Project Local | `./CLAUDE.local.md` | Private project prefs | Just you |

**Recommended Structure:**
```
project/
├── AGENTS.md              # Universal agent instructions (symlinked to CLAUDE.md)
├── CLAUDE.md              # Symlink to AGENTS.md OR Claude-specific additions
├── CLAUDE.local.md        # Personal preferences (gitignored)
├── .claude/
│   ├── CLAUDE.md          # Alternative location for project memory
│   ├── commands/          # Custom slash commands
│   ├── rules/             # Modular rule files
│   │   ├── code-style.md
│   │   ├── testing.md
│   │   ├── api-design.md
│   │   └── frontend/      # Subdirectories allowed
│   └── research/          # Cached research findings
├── .cursor/
│   └── rules/
│       └── project.mdc    # Cursor-specific rules
├── .windsurf/
│   └── rules/
│       └── rules.md       # Windsurf-specific rules
└── .github/
    └── copilot-instructions.md  # GitHub Copilot instructions
```

### 3. Import/Include Mechanisms

**Claude Code Import Syntax:**
```markdown
# Main CLAUDE.md
See @README for project overview.
See @docs/architecture.md for system design.

# Include personal overrides
@~/.claude/my-project-instructions.md

# Reference external files
See @package.json for available commands.
```

**Rules:**
- Imports use `@path/to/file` syntax
- Both relative and absolute paths supported
- NOT evaluated inside code blocks
- Recursive imports allowed (max depth: 5)
- View loaded memory with `/memory` command

**Path-Specific Rules (YAML frontmatter):**
```markdown
---
paths: src/api/**/*.ts
---

# API Development Rules
- All endpoints must include input validation
- Use standard error response format
```

### 4. Cross-Tool Reusability Patterns

**Strategy 1: AGENTS.md as Single Source of Truth**
- Keep all universal instructions in AGENTS.md
- Symlink tool-specific files: `ln -s AGENTS.md CLAUDE.md`
- Add .gitignore entry for symlinked files
- Use setup script for team onboarding

**Strategy 2: Layered Configuration**
```
AGENTS.md (base layer - all tools read this)
   └── CLAUDE.md (Claude-specific additions)
   └── .cursor/rules/project.mdc (Cursor-specific)
   └── .windsurf/rules/rules.md (Windsurf-specific)
```

**Strategy 3: Import-Based Composition**
```markdown
# CLAUDE.md
@AGENTS.md

# Claude-specific additions below
## Claude Code Features
- Use /memory to edit memory files
- Use # prefix for quick memory additions
```

### 5. What Belongs in Shared vs Tool-Specific Memory

**Shared Memory (AGENTS.md or common CLAUDE.md):**

| Category | Examples |
|----------|----------|
| Tech Stack | "React 18, TypeScript 5.3, Tailwind CSS" |
| Architecture | "Monorepo using Turborepo, shared packages in /packages" |
| Conventions | "Use 2-space indentation, prefer named exports" |
| Commands | "npm run test -- --watch for development" |
| Patterns | "Use custom hooks for data fetching, avoid useEffect for side effects" |
| Gotchas | "Database migrations require manual rollback scripts" |

**Tool-Specific Memory:**

| Tool | Content Type |
|------|--------------|
| Claude Code | Slash commands, MCP configurations, `/memory` usage, agent orchestration patterns |
| Cursor | `.mdc` rule files, Composer workflows, `@` symbol usage |
| Windsurf | Cascade memory hints, flow-specific instructions |
| Copilot | GitHub-specific conventions, PR review preferences |

### 6. Memory Initialization and Maintenance Patterns

**Initialization Approaches:**

1. **Bootstrap with /init (Claude Code)**
   - Generates basic CLAUDE.md from codebase analysis
   - Review and refine output (auto-generated content often suboptimal)

2. **Progressive Building**
   - Start with minimal instructions
   - Add rules as patterns emerge ("trial and error approach")
   - Use `#` shortcut to add memories during sessions

3. **Team Template**
   - Maintain a starter AGENTS.md in team wiki
   - Clone and customize per project
   - Include organization-wide standards

**Maintenance Patterns:**

1. **Regular Pruning**
   - Review monthly for outdated instructions
   - Remove rules that are now enforced by tooling (linters, formatters)
   - Keep under 300 lines total

2. **Effectiveness Testing**
   - Treat as prompt engineering - iterate on instructions
   - Add emphasis keywords ("IMPORTANT", "MUST") for critical rules
   - Test with real prompts, observe agent behavior

3. **Version Control**
   - Commit AGENTS.md and project CLAUDE.md
   - Keep CLAUDE.local.md in .gitignore
   - Document changes in commit messages

4. **Hierarchical Updates**
   - User-level rules for personal preferences
   - Project-level rules for team standards
   - Use path-specific rules sparingly (only when truly needed)

## Best Practices Summary

### DO:
- Keep memory files lean (<300 lines for main file)
- Be specific: "Use 2-space indentation" not "Format code properly"
- Use structure: bullet points grouped under markdown headings
- Import additional files rather than duplicating content
- Separate universal agent rules from tool-specific features
- Reference file:line locations instead of copying code snippets
- Include the WHY not just the WHAT
- Test and iterate like prompt engineering

### DON'T:
- Include generic instructions ("write clean code")
- Put style rules in memory (use linters instead)
- Mix task-specific with universal instructions
- Exceed LLM instruction limits (~150-200 reliable)
- Auto-generate without review
- Duplicate content across files
- Include secrets or credentials

## Applicable Patterns for This Codebase

Based on the agentctl repository context:

1. **Consider AGENTS.md adoption** - Create AGENTS.md for universal instructions, symlink CLAUDE.md for Claude Code compatibility

2. **Modularize rules** - Move topic-specific content to `.claude/rules/` (e.g., `workspaces.md`, `hooks.md`, `git-workflow.md`)

3. **Use import mechanism** - Main CLAUDE.md can import modular rules and reference docs

4. **Path-specific rules** - Use YAML frontmatter for language-specific rules (e.g., Python patterns only for `**/*.py`)

5. **Research caching** - Continue using `.claude/research/` for persistent cross-session findings

## Sources

- [Claude Code Memory Documentation](https://code.claude.com/docs/en/memory)
- [Claude Code Best Practices (Anthropic)](https://www.anthropic.com/engineering/claude-code-best-practices)
- [AGENTS.md Guide (Builder.io)](https://www.builder.io/blog/agents-md)
- [CLAUDE.md to AGENTS.md Migration](https://solmaz.io/log/2025/09/08/claude-md-agents-md-migration-guide/)
- [Writing a Good CLAUDE.md (HumanLayer)](https://www.humanlayer.dev/blog/writing-a-good-claude-md)
- [One File to Rule Them All (DEV.to)](https://dev.to/idavidov13/one-file-to-rule-them-all-cursor-windsurf-and-vs-code-hh2)
- [JetBrains Research: Context Management](https://blog.jetbrains.com/research/2025/12/efficient-context-management/)
- [AGENTS.md Standard (GitHub Copilot Changelog)](https://github.blog/changelog/2025-08-28-copilot-coding-agent-now-supports-agents-md-custom-instructions/)
- [AGENTS.md Medium Article](https://addozhang.medium.com/agents-md-a-new-standard-for-unified-coding-agent-instructions-0635fc5cb759)
- [AI Agent Memory Guide (Redis)](https://redis.io/blog/build-smarter-ai-agents-manage-short-term-and-long-term-memory-with-redis/)
- [Mem0 Universal Memory Layer](https://github.com/mem0ai/mem0)

## Confidence Level

**High** - Multiple authoritative sources (official documentation, engineering blogs from Anthropic, JetBrains, GitHub) converge on similar patterns. The AGENTS.md standard is explicitly supported by major vendors. The hierarchical memory and import mechanisms are well-documented in official Claude Code docs.

## Related Questions

- How should memory initialization differ between greenfield projects vs established codebases?
- What patterns work best for team memory vs individual developer memory?
- How can memory files be tested for effectiveness before deployment?
- What metrics indicate memory file bloat or ineffectiveness?
- How do MCP servers interact with memory files for extended capabilities?
