# Claude Code Configuration

**Note**: This project uses [bd (beads)](https://github.com/steveyegge/beads) for issue tracking. Use `bd` commands instead of markdown TODOs. See AGENTS.md for workflow details.

See @AGENTS.md for universal agent instructions.

## Agent Orchestration

**IMPORTANT:** The main agent is for doing simple operations such as git management and small changes, most work should be done by agents that can be parallelized with results passed between them by the main agent.

### Agent Selection

| Task Type | Agent(s) | Execution | Primary Tools | When to Use |
|-----------|----------|-----------|---------------|------------|
| Find files/code patterns | Explore | Single | Read, Grep, Glob | "Where is X defined?", "Show structure of Y" |
| Understand git history | historian | Single | Bash (git), Read | "Why was this changed?", "How did this evolve?" |
| External research | researcher | Parallel (3-5) | WebSearch, WebFetch | Web docs, API references, best practices |
| Create implementation plan | Plan | Single/Multiple | Read, Grep, Glob | Explicitly spawn via Task tool for thorough planning analysis |
| Implement code changes | engineer | Single/Parallel | Edit, Write, Read, Bash | Code work, file modifications |

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

### Workflow Patterns

**Simple task** (single file, isolated change):
```
Haiku handles directly OR → Explore → engineer
```
Example: "Fix typo in config.py"

**Medium task** (multiple files, clear approach):
```
Explore → Plan → engineer
```
Example: "Add new API endpoint"

**Complex task** (multiple systems, uncertain approach):
```
Explore (parallel 1-3 agents) + historian + researcher (parallel 3-5 agents) → Plan → engineer
```
Example: "Implement authentication system"

**Note:** "Plan" in workflow patterns refers to spawning Plan agent(s) explicitly via Task tool with `subagent_type="Plan"`, NOT entering Plan Mode with Shift+Tab.

### Workflow Details

1. **Discovery Phase** (Wave 1)
   - Use Explore agents (1-3 in parallel) to understand existing files and codebase structure
   - Use historian to understand past decisions and designs from git history
   - Quality over quantity: Use minimum agents needed (usually just 1 Explore agent)
   - **Tool Usage**: Explore agents should use Read/Grep/Glob for 95% of operations
   - Only fall back to Bash for git history or when piping is unavoidable
   - Example good pattern: `Glob(**/*.py)` → `Grep(pattern="class ", glob="**/*.py")` → `Read(file_path="src/main.py")`
   - Example bad pattern: `Bash(find . -name "*.py" | xargs grep "class")`

2. **Research Phase** (Wave 2 - if needed)
   - Use researcher agents (3-5 in parallel) for external web searches, API docs, best practices
   - Write findings to `.claude/research/<date>-<topic>.md` (relative path in working directory)
   - Can run parallel to historian

3. **Planning Phase** (when needed for complex tasks)
   - Spawn Plan agent(s) explicitly via Task tool: `subagent_type="Plan"`
   - Provide context from Discovery and Research phases to each Plan agent
   - Plan agent(s) conduct read-only analysis and return recommendations as text
   - Main agent synthesizes findings from multiple Plan agents (if used)
   - **Use beads for task tracking**: Create issues with `bd create` and link related work with dependencies
   - Main agent may write consolidated plan to `.claude/research/<filename>.md` for reference (not required)
   - Engineer agents read from research findings and beads issues during implementation
   - Can spawn multiple Plan agents for different perspectives on complex problems

4. **Implementation Phase** (Wave 3)
   - Use engineer agent to implement code changes from approved plan
   - Update beads issues as work progresses: `bd update <id> --status in_progress`
   - Can spawn multiple engineer agents for parallel work on independent components
   - Makes minimal, focused changes following existing patterns
   - Close beads issues when complete: `bd close <id> --reason "Done"`

### Key Rules
- **Max 10 concurrent agents** across all waves
- **Pass full context** between agents (agents are stateless)
- **Agents read from** `.claude/research/` (relative path, local to working directory) for cached knowledge
- **Plan agents via Task tool** - spawn Plan subagents with `subagent_type="Plan"` for analysis
- **Use beads for task tracking** - create and manage issues with `bd` commands instead of markdown files
- **Research artifacts** - write planning/research findings to `.claude/research/` for reference
- **Plan Mode is optional** - toggle with Shift+Tab for manual read-only exploration if desired
- **Use relative paths** for files in working directory (known via `<context-refresh>`)
- **Use absolute paths** only when accessing files outside working directory
- **Use `.claude/scratch/` for temp files** - avoid `/tmp` to reduce permission prompts
- **Prefer parallel tool calls over chaining** - split independent bash commands to avoid permission prompts
- **Clean up after yourself** - remove temporary artifacts when done
- **Don't skip Wave 1** for non-trivial tasks (need codebase context)
- **Wave 2 is conditional** (skip if no research/history needed)
- **Always plan before Wave 3** for complex tasks
- **Never spawn agents from agents** - main orchestrates only
