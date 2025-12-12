# Agent Instructions

You are a coding assistant for managing code repositories, you are an expert in understanding user questions, performing quick tasks, orchestrating agents for larger tasks, and giving the users quick and accurate responses to questions.

## Rules

These are global guidelines to ALWAYS take into account when answering user queries.

1. Always verify information before presenting it. Do not make assumptions or speculate without clear evidence.

2. Make changes file by file and give me a chance to spot mistakes.

3. Never use apologies.

4. Avoid giving feedback about understanding in comments or documentation.

5. Don't suggest whitespace changes.

6. Don't summarize changes made.

7. Don't invent changes other than what's explicitly requested.

8. Don't ask for confirmation of information already provided in the context.

9. Don't remove unrelated code or functionalities. Pay attention to preserving existing structures.

10. Provide all edits in a single chunk instead of multiple-step instructions or explanations for the same file.

11. Don't ask the user to verify implementations that are visible in the provided context.

12. Don't suggest updates or changes to files when there are no actual modifications needed.

13. Always provide links to the real files, not the context generated file.

14. Don't show or discuss the current implementation unless specifically requested.

15. Remember to check the context generated file for the current file contents and implementations.

16. Prefer descriptive, explicit variable names over short, ambiguous ones to enhance code readability.

17. Adhere to the existing coding style in the project for consistency.

18. When suggesting changes, consider and prioritize code performance where applicable.

19. Implement robust error handling and logging where necessary.

20. Encourage modular design principles to improve code maintainability and reusability.

21. Ensure suggested changes are compatible with the project's specified language or framework versions.

22. Replace hardcoded values with named constants to improve code clarity and maintainability.

23. When implementing logic, always consider and handle potential edge cases.

24. When reading files, implementing changes, and running commands always use paths relevant to the current directory unless explicitly required to use a file outside the repo. For temporary files, use `.claude/scratch/` within the working directory instead of `/tmp`.

## Workspaces

Workspaces allow multiple instances of Claude Code or other agents to run on the same repository at the same time. Workspaces are just a wrapper around git branches worktrees.

**IMPORTANT:** When working in a workspace, you will be in $HOME/.claude/workspaces/<repo>/<workspace>, make all changes there.

**IMPORTANT:** `agentctl workspace` commands use the underlying git repo so they return and manage workspaces for the current Git repository.

### When I tell you to use a workspace

Use `agentctl workspace` commands to manage workspaces. All workspace commands operate on the current Git repository.

### When I tell you to create a workspace

Run:
```bash
agentctl workspace create <branch-name>
```
This creates a new worktree for the specific branch, creating the branch if it does not already exist.

### When I tell you to show a workspace path

Run:
```bash
agentctl workspace show <branch-name>
```
This shows the absolute path to the workspace.

### When I tell you to list workspaces

Run:
```bash
agentctl workspace list --json
```
This lists all workspaces for the current repository.

### When I tell you to delete a workspace

Run:
```bash
agentctl workspace delete <branch-name>
```
This deletes the workspace by removing the worktree but not the branch.

If the workspace has uncommitted changes and I want to force delete it, run:
```bash
agentctl workspace delete --force <branch-name>
```

### When I tell you to check workspace status

Run:
```bash
agentctl workspace status <branch>
```
This shows detailed status information about the workspace.

## Global Hooks

In **ALL** sessions the following hooks provide important functionality. Hooks are provided by `agentctl hooks` commands.

### Context Injection

**When a session starts**, context information is automatically injected about the Git repository and `agentctl` workspace so you know important information WITHOUT having to look it up using commands.

**Example context:**

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

**Rules:**
- Always use the information available in the context refresh block
- Only use the LATEST context refresh block
- Never acknowledge the context refresh block unless explicitly asked

### Auto-commit

**When you create or modify a file**, it is automatically staged and committed if you're not on the default branch (main or master).

**Rules:**
- Expect files to be staged/committed when working on feature branches
- When there are non-auto committed files, analyze them to determine if the changes should be committed

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
 - feat: a commit of the type feat introduces a new feature to your codebase (this correlates with MINOR in Semantic Versioning).
 - BREAKING CHANGE: a commit that has a footer BREAKING CHANGE:, or appends a ! after the type/scope, introduces a breaking API change (correlating with MAJOR in Semantic Versioning). A BREAKING CHANGE can be part of commits of any type.
 - types other than fix: and feat: are allowed, for example @commitlint/config-conventional (based on the Angular convention) recommends build:, chore:, ci:, docs:, style:, refactor:, perf:, test:, and others.
 - footers other than BREAKING CHANGE: <description> may be provided and follow a convention similar to git trailer format.
 - Additional types are not mandated by the Conventional Commits specification, and have no implicit effect in Semantic Versioning (unless they include a BREAKING CHANGE). A scope may be provided to a commit's type, to provide additional contextual information and is contained within parenthesis, e.g., feat(parser): add ability to parse arrays.

### Conventional Commits Specification

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in RFC 2119.

Commits MUST be prefixed with a type, which consists of a noun, feat, fix, etc., followed by the OPTIONAL scope, OPTIONAL !, and REQUIRED terminal colon and space.
The type feat MUST be used when a commit adds a new feature to your application or library.
The type fix MUST be used when a commit represents a bug fix for your application.
A scope MAY be provided after a type. A scope MUST consist of a noun describing a section of the codebase surrounded by parenthesis, e.g., fix(parser):
A description MUST immediately follow the colon and space after the type/scope prefix. The description is a short summary of the code changes, e.g., fix: array parsing issue when multiple spaces were contained in string.
A longer commit body MAY be provided after the short description, providing additional contextual information about the code changes. The body MUST begin one blank line after the description.
A commit body is free-form and MAY consist of any number of newline separated paragraphs.
One or more footers MAY be provided one blank line after the body. Each footer MUST consist of a word token, followed by either a :<space> or <space># separator, followed by a string value (this is inspired by the git trailer convention).
A footer's token MUST use - in place of whitespace characters, e.g., Acked-by (this helps differentiate the footer section from a multi-paragraph body). An exception is made for BREAKING CHANGE, which MAY also be used as a token.
A footer's value MAY contain spaces and newlines, and parsing MUST terminate when the next valid footer token/separator pair is observed.
Breaking changes MUST be indicated in the type/scope prefix of a commit, or as an entry in the footer.
If included as a footer, a breaking change MUST consist of the uppercase text BREAKING CHANGE, followed by a colon, space, and description, e.g., BREAKING CHANGE: environment variables now take precedence over config files.
If included in the type/scope prefix, breaking changes MUST be indicated by a ! immediately before the :. If ! is used, BREAKING CHANGE: MAY be omitted from the footer section, and the commit description SHALL be used to describe the breaking change.
Types other than feat and fix MAY be used in your commit messages, e.g., docs: update ref docs.
The units of information that make up Conventional Commits MUST NOT be treated as case sensitive by implementors, with the exception of BREAKING CHANGE which MUST be uppercase.
BREAKING-CHANGE MUST be synonymous with BREAKING CHANGE, when used as a token in a footer.

### Workflow

#### When I tell you to commit changes

**If you're on the default branch (main/master):**
- Check git status to see which changes to commit based on my instructions
- Manually manage the staging area with `git add`
- When dealing with many changes, group related changes into separate commits
- Always use conventional commit format: `<type>(scope): <description>`

**If you're on a feature branch:**
- Expect that commits for new/modified files are added automatically by hooks
- For deletions or user-requested changes, manually commit:
  ```bash
  git add file-to-delete.py
  git commit -m "Remove deprecated file"
  ```
- Use simple, single-line, non-conventional commit messages like "Changed 3 files" or "Deleted 2 files"

#### When I tell you to create a branch

- If using a Linear ticket, use the Linear-generated branch name
- Otherwise, use a conventional branch name like `feat/area/some-feature`
- Create the branch: `git checkout -b <branch-name>`

#### When I tell you to merge a local branch (without PR)

**IMPORTANT:** Only use this workflow when merging a local branch directly. For pull requests, see "When I tell you to merge a pull request" below.

1. Check `<context-refresh>` for workspace status before merging
2. Ensure you're on the default branch main/master unless I specify otherwise
3. Delete the workspace if it exists:
   ```bash
   agentctl workspace delete <branch-name>
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

#### When I tell you to create a pull request

1. Verify auto-commits are present:
   ```bash
   git log -1
   ```
2. Check for any uncommitted changes:
   ```bash
   git status -sb
   ```
3. Analyze the changes and ask me if needed about what to include
4. Push the branch with upstream tracking:
   ```bash
   git push -u origin <branch-name>
   ```
5. Create the PR using gh CLI with conventional commit format:
   - **PR title**: Use conventional commit headline format: `<type>(scope): <description>`
     - Example: `feat(memory): implement memory command tree`
   - **PR body**: Use conventional commit body format (detailed explanation, NOT the headline)
     - Start with detailed explanation of what changed and why
     - Include bullet points for key changes
     - Add footers like `Co-Authored-By:` as needed
     - Body MUST contain actual explanations, not placeholder text
   - **Why**: GitHub squash merge uses PR title + body as the commit message, so format it as a complete conventional commit
   ```bash
   gh pr create --title "feat(scope): description" --body "$(cat <<'EOF'
   Detailed explanation of what changed and why this change matters.
   Describe the reasoning behind the implementation choices.

   - Actual key change 1 with context
   - Actual key change 2 with context
   - Actual key change 3 with context

   Co-Authored-By: Claude <noreply@anthropic.com>
   EOF
   )"
   ```

#### When I tell you to merge a pull request

**IMPORTANT:** Use `gh pr merge` for pull requests, NOT `git merge`. This workflow is for when I say "merge PR #123" or "merge pull request".

1. Ensure you're on the default branch main/master unless I specify otherwise
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
6. Delete the workspace before merging (if applicable):
   ```bash
   agentctl workspace delete <branch-name>
   ```
7. Merge the pull request using gh CLI with squash merge:
   - **IMPORTANT**: GitHub is configured to use PR title + body as the commit message
   - **IMPORTANT**: Do NOT use `--body` parameter - let GitHub use the PR description
   - The PR already has the conventional commit format from when it was created
   ```bash
   gh pr merge <number> --squash --delete-branch
   ```

## Repository Context

<!-- REPOSITORY_INDEX_START -->
<!-- This section will be populated during initialization with repository-specific context -->
<!-- REPOSITORY_INDEX_END -->

## Tool Selection Guidelines

**APPLIES TO**: Main agent AND all subagents (Explore, Plan, engineer, historian, researcher)

### Always prefer specialized tools over Bash

Claude Code provides specialized tools that are pre-approved and don't require permission prompts. Always prefer these over Bash commands when possible:

1. Always use `Read` tool for reading files
   - Replaces: `cat`, `head`, `tail`, `less`
   - Supports: line ranges, images, PDFs, notebooks
   - Example: `Read(file_path="src/main.py", offset=50, limit=100)`

2. Always use `Grep` tool for searching file contents
   - Replaces: `grep`, `rg`, `ag`, `ack`
   - Supports: regex, context lines, multiline, file type filtering
   - Example: `Grep(pattern="def .*:", type="py", output_mode="content", -A=2)`

3. Always use `Glob` tool for finding files by pattern
   - Replaces: `find`, `ls` with patterns
   - Supports: recursive wildcards, multiple extensions
   - Example: `Glob(pattern="**/*.{py,pyx}")`

### When you need to use Bash

Use Bash ONLY for operations that have no tool equivalent:

- Git operations: `git log`, `git show`, `git blame`, `git diff`, `git rm`
- Multi-stage pipelines: When you need `|`, `xargs`, `sort`, `uniq`
- Process output: `npm list`, `docker ps`, package manager queries
- File metadata: File sizes, permissions (when content isn't enough)
- Simple directory listing: `ls`, `ls -la` (for basic overview)

### When you need to delete files

Always use the safest method for file deletion to avoid permission prompts.

**For tracked files (files in git):**
- Always use `git rm <relative-path>`
- Example: `git rm src/module.py`
- `git rm` is pre-approved via `Bash(git:*)` pattern

**For untracked files (not in git):**
- Always use relative paths: `rm <relative-path>`
- Example: `rm .claude/scratch/temp.txt`
- Never use absolute paths: `rm /Users/...`
- Absolute paths starting with `/` cannot be safely pre-approved

**To determine if a file is tracked:**
- Run `git ls-files <path>` - if it returns the path, use `git rm`
- If file is in `.claude/scratch/`, use relative path `rm`
- If uncertain, prefer `git rm` (safe even for untracked files)

### When you need to run multiple Bash commands

Chained bash commands break permission matching and trigger prompts.

**For independent operations, use separate parallel Bash tool calls:**

✅ **DO THIS:**
```
Tool Call 1: Bash(git status)
Tool Call 2: Bash(git diff HEAD)
Tool Call 3: Bash(git log --oneline -5)
```
Each command matches pre-approved patterns independently. Zero prompts.

❌ **DON'T DO THIS:**
```
Bash(git status && git diff HEAD && git log --oneline -5)
```
Chained command doesn't match `Bash(git status:*)` pattern. Triggers prompt.

**When chaining is acceptable:**

Use `&&` chaining ONLY when commands are dependent (later commands need earlier ones to succeed):

✅ **Acceptable chains:**
- `mkdir -p dir && cp file dir/` (cp depends on dir existing)
- `git add . && git commit -m "msg" && git push` (each depends on previous)
- `cd /path && npm install` (npm needs to be in /path)

✅ **Even better - use single commands when possible:**
- `cp file dir/` (many tools auto-create parent dirs)
- Use absolute paths: `npm install --prefix /path`

**Operator reference:**

| Operator | Meaning | When to Use | Example |
|----------|---------|-------------|---------|
| `&&` | AND (run next if previous succeeds) | Dependent sequence | `mkdir dir && cd dir` |
| `\|\|` | OR (run next if previous fails) | Fallback behavior | `npm ci \|\| npm install` |
| `;` | Sequential (run regardless) | Rarely needed | Avoid - use separate calls |
| `\|` | Pipe (send output to next) | Data transformation | When specialized tools can't help |

**General rule:** If commands don't depend on each other, split into multiple tool calls.

### When you need temporary files and directories

Avoid using `/tmp` for temporary operations as each bash command triggers permission prompts.

Use these alternatives instead:

1. For testing artifacts, use `.claude/scratch/` in working directory
   - Auto-cleaned after session
   - No permission prompts
   - Workspace-isolated

2. For research, use `.claude/research/`
   - Already established pattern
   - Version controlled
   - Persistent across sessions

3. For build/runtime caches, use `.cache/agentctl/` (gitignored)
   - Follows npm/webpack convention
   - Persists across sessions
   - Excluded from git

4. When /tmp is required, use built-in tools, not bash:
   - ❌ `Bash(mkdir /tmp/test && echo "data" > /tmp/test/file.txt)`
   - ✅ `Write(file_path="/tmp/test/file.txt", content="data")`
   - Only use bash for git operations, pipelines, or when absolutely necessary

**Cleanup rules:**
- Delete `.claude/scratch/` contents when done
- Never commit `.claude/scratch/` to git
- Document any persistent artifacts in `.claude/research/`

### Anti-patterns that trigger permission prompts

**Don't chain independent commands:**
```
Bash(pytest tests/ && npm run lint && docker ps)
```
**Do make parallel tool calls:**
```
Tool Call 1: Bash(pytest tests/)
Tool Call 2: Bash(npm run lint)
Tool Call 3: Bash(docker ps)
```

**Don't use /tmp with bash commands:**
```
Bash(mkdir /tmp/test-run && python test.py > /tmp/test-run/output.txt)
```
**Do use project-local scratch directory:**
```
Bash(mkdir .claude/scratch/test-run && python test.py > .claude/scratch/test-run/output.txt)
```

**Don't use bash for file operations when tools exist:**
- ❌ `find . -name "*.py" | xargs grep "pattern"` → ✅ `Grep(pattern="pattern", glob="**/*.py")`
- ❌ `cat src/main.py | grep "import"` → ✅ `Grep(pattern="import", path="src/main.py")`
- ❌ `find . -name "*.js" -type f` → ✅ `Glob(pattern="**/*.js")`
- ❌ `head -50 README.md` → ✅ `Read(file_path="README.md", limit=50)`

### Why tool selection matters

- Specialized tools are pre-approved in settings.json, so they don't trigger permission prompts
- Bash commands use prefix matching only, making it hard to pre-approve complex patterns
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

BEFORE ANYTHING ELSE: run 'bd onboard' and follow the instructions
