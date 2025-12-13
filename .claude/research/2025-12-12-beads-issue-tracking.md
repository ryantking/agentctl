# Research: Beads (bd) Issue Tracking System

Date: 2025-12-12
Focus: Core concepts, agent workflows, git integration, comparison to markdown TODOs
Agent: researcher

## Summary

Beads (bd) is a lightweight, git-backed issue tracker designed specifically for AI coding agents. It provides a distributed database architecture using SQLite locally with JSONL synchronization through git, enabling multiple agents or humans to work concurrently without conflicts. The system addresses a critical agent limitation: memory loss across sessions ("amnesia when dealing with complex nested plans").

## Key Findings

- **Distributed git-backed architecture**: No central server required, syncs via git push/pull [GitHub README](https://github.com/steveyegge/beads)
- **Four dependency types**: blocks, related, parent-child, discovered-from enable sophisticated work choreography [FAQ](https://github.com/steveyegge/beads/blob/main/docs/FAQ.md)
- **Hash-based IDs (v0.20.1+)**: Prevents ID collisions when multiple agents work on different branches concurrently [FAQ](https://github.com/steveyegge/beads/blob/main/docs/FAQ.md)
- **Auto-sync with 5-second debounce**: Automatic export to JSONL after CRUD operations [PLUGIN.md](https://github.com/steveyegge/beads/blob/main/docs/PLUGIN.md)
- **bd onboard command**: Generates standardized agent instructions, outputs markdown covering quick start, issue types, workflow integration [Issue #73](https://github.com/steveyegge/beads/issues/73)
- **bd prime**: Injects ~1-2k tokens of workflow context, auto-detects ephemeral branches [Releases](https://github.com/steveyegge/beads/releases)

## Detailed Analysis

### Core Concepts and Philosophy

Beads treats issue tracking as a **graph-based memory system** for coding agents. The name "beads" refers to how issues chain together through dependencies "like beads on a string," making them easy for agents to follow across complex task hierarchies.

The fundamental philosophy is:

1. **Agent-first design**: Humans don't use Beads directly; the coding agent files and manages issues on their behalf
2. **Offline-first**: All queries run against local SQLite database (~10ms), no network required
3. **Git as transport**: JSONL files are versioned in git, enabling distributed collaboration without a central server
4. **Focused scope**: "bd is focused - It tracks issues, dependencies, and ready work. That's it."

### Planning and Task Management

**Dependency Types:**

| Type | Purpose | Affects Ready Queue |
|------|---------|---------------------|
| `blocks` | Hard dependency, work cannot proceed | Yes |
| `related` | Soft cross-reference | No |
| `parent-child` | Epic-to-subtask hierarchy | No |
| `discovered-from` | Links newly found work to originating task | No |

**Ready Work Detection:**
The `bd ready` command automatically surfaces issues with no open blockers. When a blocking issue closes, downstream work automatically becomes available. This eliminates manual queue management.

**Hierarchical Organization:**
```bash
bd create "Auth System" -t epic -p 1  # Creates bd-a3f8e9
bd create "Design login UI" -p 1      # Auto-generates bd-a3f8e9.1
bd create "Backend validation" -p 1   # Auto-generates bd-a3f8e9.2
bd dep tree bd-a3f8e9                 # Visualize hierarchy
```

### Best Practices for Agent Workflows

**Session End Protocol (AGENTS.md recommendation):**
1. **File/update remaining work** - Create issues for discovered bugs and follow-ups
2. **Run quality gates** - Execute tests; file P0 issues if builds fail
3. **Sync carefully** - Reconcile local and remote issues methodically
4. **Verify clean state** - Ensure all changes are pushed
5. **Choose next work** - Provide formatted context for subsequent sessions

**Context Injection Methods:**
- `bd prime` - Injects ~1-2k tokens of workflow context (preferred, context-efficient)
- `bd hooks install` - Git hooks for automatic context injection
- MCP server - Full tool integration (10-50k tokens for schemas, higher latency)

**Multi-Agent Coordination:**
- Hash-based IDs prevent collisions on concurrent branches
- Agent Mail feature for 20-50x latency reduction vs git sync
- `--no-daemon` flag for git worktrees and CI/CD pipelines

### Comparison to Markdown TODO/Plan Files

| Aspect | Beads (bd) | Markdown TODOs |
|--------|------------|----------------|
| **Structure** | Database with schema, dependencies, priorities | Freeform text |
| **Discovery** | Automatic via `bd ready`, blocked work visible | Manual scanning |
| **Dependencies** | Four typed relationships, enforced | Implicit or manual tracking |
| **Concurrency** | Hash IDs prevent collisions | Merge conflicts common |
| **Session continuity** | Survives compaction, context window limits | Lost when agents restart |
| **Query capability** | `bd list --json`, filters, stats | Grep/search |
| **Sync** | Automatic 5-second debounce to git | Manual commits |

Key advantage: Beads prevents "lost work" when agents discover problems mid-task. Issues filed during work are automatically tracked rather than requiring explicit human intervention.

### The bd onboard Command

`bd onboard` generates comprehensive agent instructions covering:
- Quick start commands (`bd init`, `bd create`, `bd list`, `bd ready`)
- Issue types (bug, feature, task, epic, chore) and priorities (P0-P4)
- AI agent workflow integration patterns
- Auto-sync behavior explanation
- MCP server setup instructions
- Rules and best practices

Usage: `bd onboard >> AGENTS.md` or `bd onboard >> CLAUDE.md`

The command was added based on community feedback in [Issue #73](https://github.com/steveyegge/beads/issues/73), which requested standardized instructions for agents assembling development teams with specialized roles (FrontEnd, BackEnd, Security, CI, Reviewer).

### Git Integration

**Auto-Sync Mechanics:**
- **Export** (SQLite -> JSONL): 5-second debounce after any CRUD operation
- **Import** (JSONL -> SQLite): Triggered when JSONL file is newer (typically after `git pull`)
- **Storage location**: `.beads/issues.jsonl` in repository root

**Merge Conflict Resolution:**
- Custom git merge driver for intelligent JSONL conflict resolution
- Prevents duplicate IDs during concurrent branch work
- Deletions manifest with git history fallback for recovery

**Branch Workflows:**
- `--from-main` flag for branch-specific sync
- Ephemeral branch detection with auto-adjusted workflow
- Worktree support via `--no-daemon` flag

**Database Management:**
```bash
bd init                    # Initialize project
bd hooks install           # Install git hooks
bd compact --stats         # View database statistics
bd compact --analyze       # Identify archival candidates
bd doctor --check-health   # Run integrity checks
```

## Applicable Patterns

For this codebase (agentctl):

1. **Integration with workspace workflows**: bd's worktree support (`--no-daemon`) aligns with agentctl's workspace management
2. **Hook compatibility**: bd hooks can coexist with agentctl hooks (context-refresh, auto-commit)
3. **Agent context injection**: `bd prime` output could be incorporated into `agentctl hook context-info`
4. **Session memory**: Replace `.claude/plans/` markdown files with bd issues for persistent tracking

## Sources

- [GitHub Repository](https://github.com/steveyegge/beads)
- [FAQ Documentation](https://github.com/steveyegge/beads/blob/main/docs/FAQ.md)
- [Plugin Documentation](https://github.com/steveyegge/beads/blob/main/docs/PLUGIN.md)
- [Quickstart Guide](https://github.com/steveyegge/beads/blob/main/docs/QUICKSTART.md)
- [Installation Guide](https://github.com/steveyegge/beads/blob/main/docs/INSTALLING.md)
- [Extension Guide](https://github.com/steveyegge/beads/blob/main/docs/EXTENDING.md)
- [Issue #73: Agent Instructions](https://github.com/steveyegge/beads/issues/73)
- [Release Notes](https://github.com/steveyegge/beads/releases)

## Confidence Level

**High** - Multiple official documentation sources consulted, release notes verified feature availability, community discussion provided implementation context. The project is actively maintained (v0.29.0 as of late 2025) with comprehensive documentation.

## Related Questions

- How should agentctl's auto-commit hook interact with bd's auto-sync?
- Can bd's ready queue replace the TodoWrite tool workflow?
- Should workspace creation automatically run `bd init`?
- How to handle bd database across worktrees (shared vs isolated)?
