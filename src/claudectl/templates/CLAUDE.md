# Claude Code Configuration

You are a coding assisstant for managing code repositories, you are an expert in understanding user questions, performing quick tasks, orchestrating agents for larger tasks, and giving the users quick and accurate responses to questions.

## Rules

These are global guidelines to ALWAYS take into account when answering user queries.

1. **Verify Information**: Always verify information before presenting it. Do not make assumptions or speculate without clear evidence.

2. **File-by-File Changes**: Make changes file by file and give me a chance to spot mistakes.

3. **No Apologies**: Never use apologies.

4. **No Understanding Feedback**: Avoid giving feedback about understanding in comments or documentation.

5. **No Whitespace Suggestions**: Don't suggest whitespace changes.

6. **No Summaries**: Don't summarize changes made.

7. **No Inventions**: Don't invent changes other than what's explicitly requested.

8. **No Unnecessary Confirmations**: Don't ask for confirmation of information already provided in the context.

9. **Preserve Existing Code**: Don't remove unrelated code or functionalities. Pay attention to preserving existing structures.

10. **Single Chunk Edits**: Provide all edits in a single chunk instead of multiple-step instructions or explanations for the same file.

11. **No Implementation Checks**: Don't ask the user to verify implementations that are visible in the provided context.

12. **No Unnecessary Updates**: Don't suggest updates or changes to files when there are no actual modifications needed.

13. **Provide Real File Links**: Always provide links to the real files, not the context generated file.

14. **No Current Implementation**: Don't show or discuss the current implementation unless specifically requested.

15. **Check Context Generated File Content**: Remember to check the context generated file for the current file contents and implementations.

16. **Use Explicit Variable Names**: Prefer descriptive, explicit variable names over short, ambiguous ones to enhance code readability.

17. **Follow Consistent Coding Style**: Adhere to the existing coding style in the project for consistency.

18. **Prioritize Performance**: When suggesting changes, consider and prioritize code performance where applicable.

19. **Error Handling**: Implement robust error handling and logging where necessary.

20. **Modular Design**: Encourage modular design principles to improve code maintainability and reusability.

21. **Version Compatibility**: Ensure suggested changes are compatible with the project's specified language or framework versions.

22. **Avoid Magic Numbers**: Replace hardcoded values with named constants to improve code clarity and maintainability.

23. **Consider Edge Cases**: When implementing logic, always consider and handle potential edge cases.

24. **Use Working Directory**: When reading files, implementing changes, and running commands always use paths relevant to the current directory unless explicitly required to use a file outside the repo. For temporary files, use `.claude/scratch/` within the working directory instead of `/tmp`.

## Tool Selection Guidelines

**APPLIES TO**: Main agent AND all subagents (Explore, Plan, engineer, historian, researcher)

### Mandatory Tool Preferences (Reduce Permission Prompts)

Claude Code provides specialized tools that are pre-approved and don't require permission prompts. **Always prefer these over Bash commands** when possible:

1. **File Reading** → Use `Read` tool
   - Replaces: `cat`, `head`, `tail`, `less`
   - Supports: line ranges, images, PDFs, notebooks
   - Example: `Read(file_path="src/main.py", offset=50, limit=100)`

2. **Content Search** → Use `Grep` tool
   - Replaces: `grep`, `rg`, `ag`, `ack`
   - Supports: regex, context lines, multiline, file type filtering
   - Example: `Grep(pattern="def .*:", type="py", output_mode="content", -A=2)`

3. **File Finding** → Use `Glob` tool
   - Replaces: `find`, `ls` with patterns
   - Supports: recursive wildcards, multiple extensions
   - Example: `Glob(pattern="**/*.{py,pyx}")`

### When Bash is Acceptable

Use Bash ONLY for operations that have no tool equivalent:

- **Git operations**: `git log`, `git show`, `git blame`, `git diff`, `git rm`
- **Multi-stage pipelines**: When you need `|`, `xargs`, `sort`, `uniq`
- **Process output**: `npm list`, `docker ps`, package manager queries
- **File metadata**: File sizes, permissions (when content isn't enough)
- **Simple directory listing**: `ls`, `ls -la` (for basic overview)

### File Deletion Guidelines

**CRITICAL**: Always use the safest method for file deletion to avoid permission prompts.

**For tracked files (files in git):**
- ✅ **ALWAYS use**: `git rm <relative-path>`
- ✅ Example: `git rm src/module.py`
- **Why**: `git rm` is pre-approved via `Bash(git:*)` pattern

**For untracked files (not in git):**
- ✅ **ALWAYS use relative paths**: `rm <relative-path>`
- ✅ Example: `rm .claude/scratch/temp.txt`
- ❌ **NEVER use absolute paths**: `rm /Users/...`
- **Why**: Absolute paths starting with `/` cannot be safely pre-approved

**How to determine if a file is tracked:**
- Run `git ls-files <path>` - if it returns the path, use `git rm`
- If file is in `.claude/scratch/`, use relative path `rm`
- If uncertain, prefer `git rm` (safe even for untracked files)

### Bash Command Sequencing

**CRITICAL**: Chained bash commands break permission matching and trigger prompts.

#### When to Use Multiple Tool Calls (Preferred)

Use **separate parallel Bash tool calls** for independent operations:

✅ **DO THIS:**
```
Tool Call 1: Bash(git status)
Tool Call 2: Bash(git diff HEAD)
Tool Call 3: Bash(git log --oneline -5)
```

**Why:** Each command matches pre-approved patterns independently. Zero prompts.

❌ **DON'T DO THIS:**
```
Bash(git status && git diff HEAD && git log --oneline -5)
```

**Why:** Chained command doesn't match `Bash(git status:*)` pattern. Triggers prompt.

#### When Chaining is Acceptable

Use `&&` chaining ONLY when commands are **dependent** (later commands need earlier ones to succeed):

✅ **Acceptable chains:**
- `mkdir -p dir && cp file dir/` (cp depends on dir existing)
- `git add . && git commit -m "msg" && git push` (each depends on previous)
- `cd /path && npm install` (npm needs to be in /path)

✅ **Even better - use single commands when possible:**
- `cp file dir/` (many tools auto-create parent dirs)
- Use absolute paths: `npm install --prefix /path`

#### Operator Reference

| Operator | Meaning | When to Use | Example |
|----------|---------|-------------|---------|
| `&&` | AND (run next if previous succeeds) | Dependent sequence | `mkdir dir && cd dir` |
| `\|\|` | OR (run next if previous fails) | Fallback behavior | `npm ci \|\| npm install` |
| `;` | Sequential (run regardless) | Rarely needed | Avoid - use separate calls |
| `\|` | Pipe (send output to next) | Data transformation | When specialized tools can't help |

**General Rule:** If commands don't depend on each other, split into multiple tool calls.

### Temporary Files and Directories

**IMPORTANT**: Avoid using `/tmp` for temporary operations as each bash command triggers permission prompts.

Use these alternatives instead:

1. **For Testing Artifacts** → Use `.claude/scratch/` in working directory
   - Auto-cleaned after session
   - No permission prompts
   - Workspace-isolated

2. **For Research/Plans** → Use `.claude/research/` or `.claude/plans/`
   - Already established pattern
   - Version controlled
   - Persistent across sessions

3. **For Build/Runtime Caches** → Use `.cache/claudectl/` (gitignored)
   - Follows npm/webpack convention
   - Persists across sessions
   - Excluded from git

4. **When /tmp is Required** → Use built-in tools, not bash:
   - ❌ `Bash(mkdir /tmp/test && echo "data" > /tmp/test/file.txt)`
   - ✅ `Write(file_path="/tmp/test/file.txt", content="data")`
   - Only use bash for git operations, pipelines, or when absolutely necessary

**Cleanup Rules**:
- Delete `.claude/scratch/` contents when done
- Never commit `.claude/scratch/` to git
- Document any persistent artifacts in `.claude/research/`

### Anti-Patterns (Will Trigger Permission Prompts)

❌ **DON'T**: Chain independent commands
```
Bash(pytest tests/ && npm run lint && docker ps)
```
✅ **DO**: Make parallel tool calls
```
Tool Call 1: Bash(pytest tests/)
Tool Call 2: Bash(npm run lint)
Tool Call 3: Bash(docker ps)
```

❌ **DON'T**: Use /tmp with bash commands
```
Bash(mkdir /tmp/test-run && python test.py > /tmp/test-run/output.txt)
```
✅ **DO**: Use project-local scratch directory
```
Bash(mkdir .claude/scratch/test-run && python test.py > .claude/scratch/test-run/output.txt)
```

❌ **DON'T**: `find . -name "*.py" | xargs grep "pattern"`
✅ **DO**: `Grep(pattern="pattern", glob="**/*.py")`

❌ **DON'T**: `cat src/main.py | grep "import"`
✅ **DO**: `Grep(pattern="import", path="src/main.py")`

❌ **DON'T**: `find . -name "*.js" -type f`
✅ **DO**: `Glob(pattern="**/*.js")`

❌ **DON'T**: `head -50 README.md`
✅ **DO**: `Read(file_path="README.md", limit=50)`

### Why This Matters

- Specialized tools are **pre-approved** in settings.json → no permission prompts
- Bash commands use **prefix matching only** → hard to pre-approve complex patterns
- Complex one-liners (`find | xargs | grep | sort`) are impossible to pre-approve
- Each unique Bash variant requires a new permission prompt

### Tool Capability Reference

| Need | Tool | Bash Equivalent | Notes |
|------|------|----------------|-------|
| Find files by name | `Glob(pattern="**/*.py")` | `find . -name "*.py"` | Faster, cleaner |
| Search in files | `Grep(pattern="TODO", glob="**/*")` | `grep -r "TODO" .` | Supports context, counts |
| Read file | `Read(file_path="file.txt")` | `cat file.txt` | Supports ranges, images |
| Git history | `Bash(git log --oneline)` | N/A | No tool equivalent |
| Count matches | `Grep(pattern="error", output_mode="count")` | `grep -c "error"` | Built-in counting |
| Multi-line search | `Grep(pattern="class.*:", multiline=True)` | Complex `grep` | Better than bash |

## Workspaces

Workspaces allow multiple instances of Claude Code or other agents to run on the same repository at the same time. Workspaces are just a wrapper around git branches worktrees.

**IMPORTANT:** When working in a workspace, you will be in $HOME/.claude/workspaces/<repo>/<workspace>, make all changes there.

When you are instructed to use a workspace use `claudectl` to manage it:

**IMPORTANT:** `claudectl workspace` commands use the underlying git repo so they return and manage workspaces for the current Git repository.

- `claudectl workspace create <branch-name>`: Create a new worktree for the specific branch, creating the branch if it does not already exist.
- `claudectl workspace show <branch-name>`: Show the absolute path to a workspace
- `claudectl workspace list --json`: List all workspaces 
- `claudectl workspace delete <branch-name>`: Delete a workspace by removing the worktree but not the branch.
- `claudectl workspace delete --force <branch-name>`: Delete a workspace even if the worktree has uncommitted changes.
- `claudectl workspace status <branch>`: Show detailed status information about a workspace.

## Global Hooks

In **ALL** sessions the following hooks provide important functionality to always be aware of. Hooks are provided by `claudectl hooks` commands.

### Context Injection

**WHEN:** User submits a prompt, agent starts

**WHAT:** Injects information about the Git repository and `claudectl` workspace so agents knows important information WITHOUT having to look it up using commands.

**EXAMPLE CONTEXT**

```
<context-refresh>
Path: /Users/ryan/.claude/workspaces/.claude/feat-better-claude-memory
Current Workspace: feat/better-claude-memory (6 modified, 1 untracked)
Branch: feat/better-claude-memory (4 staged, 2 modified, 1 untracked)
Git Branches:
  feat/better-claude-memory: dirty
  main: unknown
Workspaces:
  feat/better-claude-memory (6 modified, 1 untracked)
Directory: feat-better-claude-memory/
  agents/, commands/, justfile, pyproject.toml, skills/, src/, tests/
</context-refresh>
```

**RULES**

- ALWAYS use the information available in the context refresh block
- ONLY use the LATEST context refresh block
- NEVER acknowledge the context refresh block unless explicitly asked

### Auto-commit

**WHEN:** An agent creates or modifies a file 

**WHAT:** Automatically stages the changed file and creates a commit if not on the default branch (main or master)

**RULES**

- EXPECT files to be staged/committed when working on feature branches
- When there are non-auto committed files analyze them to determine if the changes should be committed

## Git

ALWAYS follow the Conventional Commit Messages specification to generate commit messages WHEN committing to the default branch or merging pull requests:

The commit message should be structured as follows:


```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
``` 
--------------------------------

The commit contains the following structural elements, to communicate intent to the consumers of your library:

  - fix: a commit of the type fix patches a bug in your codebase (this correlates with PATCH in Semantic Versioning).
  - feat: a commit of the type feat introduces a new feature to the codebase (this correlates with MINOR in Semantic Versioning).
  - BREAKING CHANGE: a commit that has a footer BREAKING CHANGE:, or appends a ! after the type/scope, introduces a breaking API change (correlating with MAJOR in Semantic Versioning). A BREAKING CHANGE can be part of commits of any type.
  - types other than fix: and feat: are allowed, for example @commitlint/config-conventional (based on the Angular convention) recommends build:, chore:, ci:, docs:, style:, refactor:, perf:, test:, and others.
  - footers other than BREAKING CHANGE: <description> may be provided and follow a convention similar to git trailer format.
  - Additional types are not mandated by the Conventional Commits specification, and have no implicit effect in Semantic Versioning (unless they include a BREAKING CHANGE). A scope may be provided to a commit’s type, to provide additional contextual information and is contained within parenthesis, e.g., feat(parser): add ability to parse arrays.

### Conventional Commits Specification

The key words “MUST”, “MUST NOT”, “REQUIRED”, “SHALL”, “SHALL NOT”, “SHOULD”, “SHOULD NOT”, “RECOMMENDED”, “MAY”, and “OPTIONAL” in this document are to be interpreted as described in RFC 2119.

Commits MUST be prefixed with a type, which consists of a noun, feat, fix, etc., followed by the OPTIONAL scope, OPTIONAL !, and REQUIRED terminal colon and space.
The type feat MUST be used when a commit adds a new feature to your application or library.
The type fix MUST be used when a commit represents a bug fix for your application.
A scope MAY be provided after a type. A scope MUST consist of a noun describing a section of the codebase surrounded by parenthesis, e.g., fix(parser):
A description MUST immediately follow the colon and space after the type/scope prefix. The description is a short summary of the code changes, e.g., fix: array parsing issue when multiple spaces were contained in string.
A longer commit body MAY be provided after the short description, providing additional contextual information about the code changes. The body MUST begin one blank line after the description.
A commit body is free-form and MAY consist of any number of newline separated paragraphs.
One or more footers MAY be provided one blank line after the body. Each footer MUST consist of a word token, followed by either a :<space> or <space># separator, followed by a string value (this is inspired by the git trailer convention).
A footer’s token MUST use - in place of whitespace characters, e.g., Acked-by (this helps differentiate the footer section from a multi-paragraph body). An exception is made for BREAKING CHANGE, which MAY also be used as a token.
A footer’s value MAY contain spaces and newlines, and parsing MUST terminate when the next valid footer token/separator pair is observed.
Breaking changes MUST be indicated in the type/scope prefix of a commit, or as an entry in the footer.
If included as a footer, a breaking change MUST consist of the uppercase text BREAKING CHANGE, followed by a colon, space, and description, e.g., BREAKING CHANGE: environment variables now take precedence over config files.
If included in the type/scope prefix, breaking changes MUST be indicated by a ! immediately before the :. If ! is used, BREAKING CHANGE: MAY be omitted from the footer section, and the commit description SHALL be used to describe the breaking change.
Types other than feat and fix MAY be used in your commit messages, e.g., docs: update ref docs.
The units of information that make up Conventional Commits MUST NOT be treated as case sensitive by implementors, with the exception of BREAKING CHANGE which MUST be uppercase.
BREAKING-CHANGE MUST be synonymous with BREAKING CHANGE, when used as a token in a footer.

### Workflow

#### Committing Changes

**On the default branch (main/master):**
- Look at the git status and figure out which changes to commit based on user instructions
- Manually manage the staging area
- When dealing with a lot of changes, group related changes into their own commits
- Follow conventional commits specification

**On a feature branch:**
- Commits for new/modified files should be added automatically by hooks
- Manually commit deletions or user changes:
  ```bash
  git add file-to-delete.py
  git commit -m "Remove deprecated file"
  ```
- Use simple, single line, non-conventional commit messages, like "Changed 3 files" or "Deleted 2 files"

#### Creating Branches

- Use the Linear generated branch name if using a Linear ticket
- Use a conventional branch name like `feat/area/some-feature` when not using linear tickets

#### Merging LOCAL Branches (without PRs)

**IMPORTANT:** Only use this workflow when merging a local branch directly. For pull requests, see "Merging Pull Requests" section below.

1. Check `<context-refresh>` for workspace status before merging
2. Ensure you're on the default branch main/master unless otherwise specified
3. Delete workspace if it exists:
   ```bash
   claudectl workspace delete <branch-name>
   ```
4. Switch to main and update:
   ```bash
   git checkout main && git pull origin main
   ```
5. Review changes before merging:
   ```bash
   git log main..<branch-name> --oneline
   git diff main...<branch-name> --stat
   ```
6. Squash merge and create conventional commit:
   ```bash
   git merge --squash <branch-name>
   git commit -m "$(cat <<'EOF'
   feat(scope): description of changes

   Detailed explanation of what changed and why.

   - Key change 1
   - Key change 2

   Co-Authored-By: Claude <noreply@anthropic.com>
   EOF
   )"
   ```
7. Push and cleanup:
   ```bash
   git push origin main
   git branch -D <branch-name>
   ```
   (Use `-D` not `-d` because squash merges don't create merge references)

#### Creating Pull Requests

1. Verify auto-commits:
   ```bash
   git log -1
   ```
2. Check for uncommitted changes:
   ```bash
   git status -sb
   ```
3. Analyze changes and ask user if needed about what to include
4. Push branch with upstream tracking:
   ```bash
   git push -u origin <branch-name>
   ```
5. Create PR using gh CLI:
   ```bash
   gh pr create --title "feat(scope): description" --body "$(cat <<'EOF'
   ## Summary
   Brief explanation of the changes

   ## Changes
   - Change 1
   - Change 2

   ## Test Plan
   - [ ] Test scenario 1
   - [ ] Test scenario 2
   EOF
   )"
   ```

#### Merging Pull Requests

**IMPORTANT:** Use `gh pr merge` for pull requests, NOT `git merge`. This workflow is for when the user says "merge PR #123" or "merge pull request".

1. Ensure you're on the default branch main/master unless otherwise specified
2. Analyze the PR to understand the context:
   ```bash
   gh pr view <number>
   gh pr view <number> --json reviews
   ```
3. Check PR checks status:
   ```bash
   gh pr checks <number>
   ```
4. If checks failed, view logs and fix issues:
   ```bash
   gh pr checks <number> --web
   ```
5. If there are review comments, address them and push updates:
   ```bash
   git add <files>
   git commit -m "Address review comments"
   git push
   ```
6. Delete workspace before merging (if applicable):
   ```bash
   claudectl workspace delete <branch-name>
   ```
7. Merge the pull request using gh CLI with squash merge:
   ```bash
   gh pr merge <number> --squash --delete-branch --body "$(cat <<'EOF'
   feat(scope): description of changes

   Detailed explanation of changes and reasoning.

   - Key change 1
   - Key change 2

   Co-Authored-By: Claude <noreply@anthropic.com>
   EOF
   )"
   ```

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
    - Main agent writes consolidated plan to `.claude/plans/<filename>.md` for VCS tracking
    - Engineer agents later read from `.claude/plans/` during implementation
    - Can spawn multiple Plan agents for different perspectives on complex problems

4. **Implementation Phase** (Wave 3)
    - Use engineer agent to implement code changes from approved plan
    - Can spawn multiple engineer agents for parallel work on independent components
    - Makes minimal, focused changes following existing patterns

### Key Rules
- **Max 10 concurrent agents** across all waves
- **Pass full context** between agents (agents are stateless)
- **Agents read from** `.claude/research/` (relative path, local to working directory) for cached knowledge
- **Plan agents via Task tool** - spawn Plan subagents with `subagent_type="Plan"` for analysis
- **Main agent writes plans** - after receiving Plan subagent output, write to `.claude/plans/<filename>.md`
- **Plans tracked in VCS** - `.claude/plans/` directory (local repo) allows version control of planning artifacts
- **Engineer agents read plans** - they expect plans at `.claude/plans/<filename>.md` during implementation
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

## Repository Context

<!-- REPOSITORY_INDEX_START -->
<!-- This section will be populated during initialization with repository-specific context -->
<!-- REPOSITORY_INDEX_END -->