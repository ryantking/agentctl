# Claude Code Configuration

**Note**: This project uses [bd (beads)](https://github.com/steveyegge/beads) for issue tracking. Use `bd` commands instead of markdown TODOs. See AGENTS.md for workflow details.

See @AGENTS.md for universal agent instructions.

## Agent Orchestration

**IMPORTANT:** You (the main agent) handle simple operations such as git management and small changes. Most work should be done by specialized agents that can be parallelized with results passed between them.

### When to use each agent type

| When I ask you to... | Use this agent | How many | Primary tools |
|---------------------|----------------|----------|---------------|
| Find files or code patterns ("Where is X defined?", "Show structure of Y") | Explore | Single | Read, Grep, Glob |
| Understand git history ("Why was this changed?", "How did this evolve?") | historian | Single | Bash (git), Read |
| Research external docs, APIs, best practices | researcher | Parallel (3-5) | WebSearch, WebFetch |
| Create an implementation plan for complex tasks | Plan | Single/Multiple | Read, Grep, Glob (spawn via Task tool) |
| Implement code changes | engineer | Single/Parallel | Edit, Write, Read, Bash |

### Tool Access Patterns

**Explore Agent** (Read-only specialist):
- **Primary tools**: Read, Grep, Glob (always prefer these)
- **Bash fallback**: Git commands, `ls`, pipelines when absolutely necessary
- **Prohibited**: Any write operations, network calls, destructive commands

**Plan Agent** (Strategy specialist):
- **Primary tools**: Read, Grep, Glob for code analysis
- **Bash fallback**: Git history, dependency trees, build tool queries
- **Prohibited**: Implementation commands (that's engineer's job)

**Engineer Agent** (Implementation specialist):
- **All tools**: Read, Edit, Write, Grep, Glob, Bash (full access)
- **Bash usage**: Build commands, tests, git operations, file modifications
- **Best practice**: Still prefer Read/Grep/Glob for exploration phase

### Tool Selection Decision Tree

```
Need to explore codebase?
├─ Finding files by pattern? → Use Glob
├─ Searching file contents? → Use Grep
├─ Reading specific files? → Use Read
└─ Git history/metadata? → Use Bash (git/ls)

Need to implement changes?
├─ Creating new file? → Use Write
├─ Modifying existing? → Use Edit (after Read)
└─ Running builds/tests? → Use Bash
```

### When I ask you to do a task

**Simple task** (single file, isolated change like "Fix typo in config.py"):
- Handle it directly OR spawn Explore → engineer

**Medium task** (multiple files, clear approach like "Add new API endpoint"):
- Spawn Explore → Plan → engineer

**Complex task** (multiple systems, uncertain approach like "Implement authentication system"):
- Spawn Explore (parallel 1-3 agents) + historian + researcher (parallel 3-5 agents) → Plan → engineer

**Note:** "Plan" refers to spawning Plan agent(s) explicitly via Task tool with `subagent_type="Plan"`, NOT entering Plan Mode with Shift+Tab.

### How to execute each workflow phase

**Wave 1: Discovery Phase** (always do this for non-trivial tasks)
- Spawn Explore agents (1-3 in parallel) to understand existing files and codebase structure
- Spawn historian to understand past decisions and designs from git history
- Use minimum agents needed (usually just 1 Explore agent)
- **Tool Usage**: Explore agents must use Read/Grep/Glob for 95% of operations
- Only fall back to Bash for git history or when piping is unavoidable
- Example good pattern: `Glob(**/*.py)` → `Grep(pattern="class ", glob="**/*.py")` → `Read(file_path="src/main.py")`
- Example bad pattern: `Bash(find . -name "*.py" | xargs grep "class")`

**Wave 2: Research Phase** (do this if external research is needed)
- Spawn researcher agents (3-5 in parallel) for external web searches, API docs, best practices
- Write findings to `.claude/research/<date>-<topic>.md` (relative path in working directory)
- Can run parallel to historian

**Wave 3: Planning Phase** (do this for complex tasks)
- Spawn Plan agent(s) explicitly via Task tool: `subagent_type="Plan"`
- Provide context from Discovery and Research phases to each Plan agent
- Plan agent(s) conduct read-only analysis and return recommendations as text
- Synthesize findings from multiple Plan agents (if used)
- **Use beads for task tracking**: Create issues with `bd create` and link related work with dependencies
- You may write consolidated plan to `.claude/research/<filename>.md` for reference (not required)
- Engineer agents will read from research findings and beads issues during implementation
- Can spawn multiple Plan agents for different perspectives on complex problems

**Wave 4: Implementation Phase** (do this after planning)
- Spawn engineer agent to implement code changes from approved plan
- Update beads issues as work progresses: `bd update <id> --status in_progress`
- Can spawn multiple engineer agents for parallel work on independent components
- Make minimal, focused changes following existing patterns
- Close beads issues when complete: `bd close <id> --reason "Done"`

### Rules for agent orchestration
- **Never exceed 10 concurrent agents** across all waves
- **Always pass full context** between agents (agents are stateless)
- **Have agents read from** `.claude/research/` (relative path, local to working directory) for cached knowledge
- **Spawn Plan agents via Task tool** with `subagent_type="Plan"` for analysis
- **Use beads for task tracking** - create and manage issues with `bd` commands instead of markdown files
- **Write research artifacts** to `.claude/research/` for reference
- **Plan Mode is optional** - toggle with Shift+Tab for manual read-only exploration if desired
- **Use relative paths** for files in working directory (known via `<context-refresh>`)
- **Use absolute paths** only when accessing files outside working directory
- **Use `.claude/scratch/` for temp files** - avoid `/tmp` to reduce permission prompts
- **Prefer parallel tool calls over chaining** - split independent bash commands to avoid permission prompts
- **Clean up temporary artifacts** when done
- **Never skip Wave 1** for non-trivial tasks (need codebase context)
- **Wave 2 is conditional** (skip if no research/history needed)
- **Always plan before Wave 4** for complex tasks
- **Never spawn agents from agents** - you (main agent) orchestrate only
