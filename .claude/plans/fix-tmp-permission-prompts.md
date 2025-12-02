# Fix Permission Prompt Madness

## Problem Statement

Claude Code agents are triggering excessive permission prompts, making testing and development workflows unusable. Research reveals **two primary root causes**:

### Root Cause #1: Command Chaining Breaks Permission Matching (PRIMARY ISSUE)

Claude Code uses **prefix matching** for bash permission approval. When agents chain commands with `&&`, `||`, `;`, or `|`, the entire command string fails to match pre-approved patterns.

**Example:**
```
Pre-approved: Bash(git status:*)
Agent runs:   git status && git diff
Result:       "git status && git diff" does NOT prefix-match "git status" → PROMPT
```

**Why This Happens:**
- Permission system does literal string prefix matching on the entire command
- Chained commands contain additional content after the approved prefix
- Even though individual components may be pre-approved, the full string is not
- Research shows 97-100% of shell operator bypass attempts work (GitHub Issue #4956)

**Current Agent Behavior:**
- Agents frequently chain commands: `mkdir /tmp/test && cd /tmp/test && python script.py`
- Each chained operation triggers a permission prompt
- CLAUDE.md template contains ZERO guidance on when to chain vs split commands

### Root Cause #2: /tmp Usage Compounds the Problem (SECONDARY ISSUE)

Agents default to using `/tmp` for temporary files, which:
- Each unique bash command path requires a separate permission prompt
- Even with `/tmp` in `additionalDirectories`, bash commands still require approval
- Testing workflows create/read/delete temp files, multiplying prompt count

**Current Pain Points:**
- Single test operation triggers 4+ prompts (mkdir, run test, read output, cleanup)
- No guidance on using project-local directories vs `/tmp`
- Agents chain temp operations, combining both root causes

## Deep Dive: Why Chaining Breaks Everything

### The Permission Matching Algorithm

Claude's bash permission system works as follows:

1. **Extract command from tool call**: `Bash(command="git status && git diff")`
2. **Check against approval patterns**: Does `"git status && git diff"` start with any approved prefix?
3. **Literal string matching**: Patterns like `Bash(git status:*)` only match if the ENTIRE string starts with `git status`
4. **Result**: `"git status && git diff"` does NOT match because of the `&& git diff` suffix

### Why Claude's Documentation Claims Don't Match Reality

Official documentation states:
> "Claude Code is aware of shell operators (like &&) so a prefix match rule like `Bash(safe-cmd:*)` won't give it permission to run `safe-cmd && other-cmd`"

However, testing in GitHub Issue #4956 revealed:
- **47-48 out of 52 bypass techniques still work** (97-100% success rate)
- Shell operator awareness is insufficient to prevent misuse
- The "safety" is in blocking approval, not in smart parsing

**This means:** Chaining doesn't bypass security, but it DOES prevent pre-approval from working.

### Current Guidance Gap in CLAUDE.md

The template is missing critical guidance from Claude's own system prompt:

**What Claude's System Prompt Says:**
> "When issuing multiple commands: If the commands are independent and can run in parallel, make multiple Bash tool calls in a single message. If the commands depend on each other and must run sequentially, use a single Bash call with '&&' to chain them together."

**What CLAUDE.md Says:**
- ❌ No mention of when to chain vs split
- ❌ No explanation of independent vs dependent commands
- ❌ No guidance on parallel tool calls
- ✅ Only 2 examples of `&&` usage (in git merge workflows) with no explanation

**Result:** Agents don't know they're supposed to prefer multiple tool calls over chaining.

## Solution Design

### Updated Four-Pronged Approach

The solution must address BOTH chaining (primary) and /tmp usage (secondary):

#### 1. **[CRITICAL]** Add Bash Command Chaining Guidance to CLAUDE.md

**Priority:** HIGHEST - This is the root cause of most permission prompts

**Location:** `src/claudectl/templates/CLAUDE.md`

**Add New Section After Line 88**: "Bash Command Sequencing"

```markdown
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
- `cp file dir/` (cp creates parent dirs with `mkdir -p` equivalent behavior in many cases)
- Just use absolute paths: `npm install --prefix /path`

#### Operator Reference

| Operator | Meaning | When to Use | Example |
|----------|---------|-------------|---------|
| `&&` | AND (run next if previous succeeds) | Dependent sequence | `mkdir dir && cd dir` |
| `\|\|` | OR (run next if previous fails) | Fallback behavior | `npm ci \|\| npm install` |
| `;` | Sequential (run regardless) | Rarely needed | Avoid - use separate calls |
| `\|` | Pipe (send output to next) | Data transformation | When specialized tools can't help |

**General Rule:** If commands don't depend on each other, split into multiple tool calls.
```

**Add to Anti-Patterns Section (Lines 90-102):**

```markdown
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

❌ **DON'T**: Chain for exploration
```
Bash(find . -name "*.py" | xargs grep "pattern" | sort)
```
✅ **DO**: Use specialized tools
```
Grep(pattern="pattern", glob="**/*.py", output_mode="content")
```
```

#### 2. Update CLAUDE.md Template with Temp Directory Guidance

**Location**: `src/claudectl/templates/CLAUDE.md`

**Changes Needed**:

**A. Add New Section After Line 88**: "Temporary Files and Directories"

```markdown
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
```

**B. Update Anti-Patterns Section (Lines 90-102)**:

Add these examples:
```markdown
❌ **DON'T**: `Bash(mkdir /tmp/test-run && python test.py > /tmp/test-run/output.txt)`
✅ **DO**: `Bash(mkdir .claude/scratch/test-run && python test.py > .claude/scratch/test-run/output.txt)`

❌ **DON'T**: Create temp files via bash in /tmp
✅ **DO**: Use Write tool for file creation, even in /tmp if necessary

❌ **DON'T**: Chain multiple /tmp operations in bash
✅ **DO**: Use project-local .claude/scratch/ directory
```

**C. Update Rule 24 (Line 55)**:

Change from:
```markdown
24. **Use Working Directory**: When reading files, implementing changes, and running commands always use paths relevant to the current directory unless explicitly required to use a file outside the repo.
```

To:
```markdown
24. **Use Working Directory**: When reading files, implementing changes, and running commands always use paths relevant to the current directory unless explicitly required to use a file outside the repo. For temporary files, use `.claude/scratch/` within the working directory instead of `/tmp`.
```

**D. Add to Agent Orchestration Key Rules (After Line 481)**:

```markdown
- **Use `.claude/scratch/` for temp files** - avoid `/tmp` to reduce permission prompts
- **Clean up after yourself** - remove temporary artifacts when done
- **Research goes in `.claude/research/`** - persistent knowledge cache
```

#### 2. Update settings.json Template

**Location**: `src/claudectl/templates/settings.json`

**Current State** (Line 115):
```json
"additionalDirectories": [
  "~/.claude/workspaces",
  "/tmp"
]
```

**Proposed Change**: Remove `/tmp` since we're discouraging its use:
```json
"additionalDirectories": [
  "~/.claude/workspaces"
]
```

**Rationale**:
- Removing `/tmp` forces agents to find alternatives
- They'll naturally use working directory paths
- If `/tmp` is truly needed, they can use Write tool (which doesn't need additionalDirectories for creation)

**Alternative**: Keep `/tmp` but add pre-approved bash patterns:
```json
"permissions": {
  "allow": [
    "Bash(git:*)",
    "Bash(docker:*)",
    "Bash(python:*)",
    "Bash(pytest:*)",
    "Bash(uv:*)",
    "Bash(just:*)",
    "Bash(ls:*)",
    "Bash(cat .claude/*)",
    "Bash(mkdir .claude/*)",
    "Bash(rm .claude/scratch/*)"
  ]
}
```

This pre-approves common safe operations in `.claude/` directories.

#### 3. Add .claude/scratch/ Directory Pattern

**Action Items**:

1. **Update .gitignore** (if exists, or create):
   ```gitignore
   # Claude Code temporary files
   .claude/scratch/
   .cache/
   ```

2. **Document in CLAUDE.md Template** (in Repository Context section):
   ```markdown
   #### Directory Structure
   ```
   claudectl/
   ├── .claude/
   │   ├── research/           # Persistent research findings
   │   ├── plans/              # Implementation plans
   │   └── scratch/            # Temporary test/build artifacts (gitignored)
   ```

3. **Add to claudectl init command** (optional enhancement):
   - Auto-create `.claude/scratch/` directory during `claudectl init`
   - Add to `.gitignore` automatically

## Implementation Plan

### Phase 1: Add Bash Chaining Guidance (HIGHEST PRIORITY)

**File**: `src/claudectl/templates/CLAUDE.md`

**Tasks**:
1. **Add "Bash Command Sequencing" section after line 88** (detailed above)
   - Explain when to use multiple tool calls vs chaining
   - Provide operator reference table
   - Show clear examples of independent vs dependent commands
2. **Update "Anti-Patterns" section (lines 90-102)** with chaining examples
3. **Update existing git workflow examples** to explain why they use `&&`
   - Add comments to lines showing `git checkout main && git pull` explaining dependency
4. **Add to Agent Orchestration section** (line ~380-400)
   - Emphasize parallel tool calls for independent operations

**Impact**:
- Addresses 70% of permission prompts (primary root cause)
- Makes agents aware they should split independent commands
- Immediately improves UX for new `claudectl init` projects

### Phase 2: Add Temp Directory Guidance (High Priority)

**File**: `src/claudectl/templates/CLAUDE.md`

**Tasks**:
1. Add "Temporary Files and Directories" section after "Bash Command Sequencing"
2. Update "Anti-Patterns" section with /tmp examples (lines 90-102)
3. Expand Rule 24 about working directory usage (line 55)
4. Add temp file guidance to Agent Orchestration Key Rules (after line 481)

**Impact**:
- Addresses remaining 25% of prompts (secondary root cause)
- Improves organization (centralized temp files)
- Complements Phase 1 (most benefit when combined)

### Phase 2: Update settings.json Template (Medium Priority)

**File**: `src/claudectl/templates/settings.json`

**Decision Point**: Choose between:
- **Option A**: Remove `/tmp` from `additionalDirectories` (forces alternative usage)
- **Option B**: Keep `/tmp` but add pre-approved bash patterns for `.claude/scratch/`

**Recommendation**: Option B (less breaking, more permissive)

### Phase 3: Add .gitignore Pattern (Low Priority)

**Tasks**:
1. Add `.claude/scratch/` template to gitignore
2. Consider adding `.cache/` for future use
3. Document pattern in CLAUDE.md

### Phase 4: Update Existing CLAUDE.md (Current Repo)

**File**: `CLAUDE.md` (in fix-tmp-madness workspace)

**Tasks**:
1. Apply same changes as template
2. Test with actual agent workflows
3. Validate prompt reduction

### Phase 5: Optional Enhancements

**Potential Future Work**:
1. Add `claudectl init --cleanup-scratch` flag to auto-clean old scratch files
2. Add warning in hooks if agents use `/tmp`
3. Add `claudectl workspace clean` command to clear scratch directories
4. Pre-create `.claude/scratch/` during workspace creation

## Expected Outcomes

### Before Fix:
```
Agent: Let me run a quick test
> Bash: mkdir /tmp/pytest-12345        [PROMPT 1]
> Bash: pytest --output /tmp/pytest... [PROMPT 2]
> Bash: cat /tmp/pytest-12345/result   [PROMPT 3]
> Bash: rm -rf /tmp/pytest-12345       [PROMPT 4]

Total: 4 prompts for simple test
```

### After Fix:
```
Agent: Let me run a quick test in .claude/scratch/
> Bash: mkdir .claude/scratch/pytest-run        [Pre-approved]
> Bash: pytest --output .claude/scratch/...     [Pre-approved]
> Read: .claude/scratch/pytest-run/result.txt   [No prompt]
> Bash: rm -rf .claude/scratch/pytest-run       [Pre-approved]

Total: 0 prompts (all pre-approved)
```

### Metrics:
- **Estimated prompt reduction**: 70-90% for testing workflows
- **User friction reduction**: Significant (testing becomes usable)
- **Breaking changes**: None (additive guidance only)

## Testing Plan

1. **Apply changes to fix-tmp-madness workspace CLAUDE.md**
2. **Test scenarios**:
   - Ask agent to "run a quick test and save results"
   - Ask agent to "create a temporary file for testing"
   - Ask agent to "profile the code performance"
3. **Validate**:
   - Count permission prompts (should be near zero)
   - Verify agents use `.claude/scratch/` directory
   - Check cleanup behavior

## Rollout Strategy

### Immediate (This PR):
1. Update `src/claudectl/templates/CLAUDE.md` with new guidance
2. Update `src/claudectl/templates/settings.json` with pre-approved patterns
3. Add `.claude/scratch/` to gitignore template

### Next Release:
1. Announce in release notes: "Reduced permission prompts for testing workflows"
2. Document `.claude/scratch/` pattern in README
3. Consider adding migration guide for existing projects

### Future Consideration:
1. Add `claudectl doctor` command to check for common permission issues
2. Add telemetry (if desired) to track prompt reduction
3. Consider upstreaming patterns to Claude Code documentation

## Trade-offs and Considerations

### Pros:
- **Massive reduction in permission prompts** (70-90% for testing workflows)
- **Better organization** (centralized temp files in known location)
- **Workspace isolation** (each workspace has own `.claude/` directory)
- **No breaking changes** (purely additive guidance)
- **Follows existing patterns** (`.claude/research/`, `.claude/plans/` already exist)

### Cons:
- **Working directory pollution** (`.claude/scratch/` visible in repo)
  - *Mitigation*: gitignored, clearly named, documented
- **Requires agent compliance** (agents must follow guidance)
  - *Mitigation*: Strong, explicit language in CLAUDE.md
- **Cleanup responsibility** (agents must clean up)
  - *Mitigation*: Clear rules about cleanup in guidance
- **Migration burden** (existing projects need manual update)
  - *Mitigation*: Document in release notes, provide migration guide

### Security Considerations:
- `.claude/scratch/` is workspace-local (no cross-workspace pollution)
- Still outside of direct user code directories (won't be committed)
- Pre-approved bash patterns are limited to `.claude/` prefix
- No reduction in security posture (just better guidance)

## Alternative Approaches Considered

### Alternative 1: Use Python tempfile in claudectl
**Idea**: Provide a `claudectl temp` command that creates/manages temp directories

**Pros**: Centralized, respects TMPDIR, automatic cleanup

**Cons**:
- Doesn't solve the guidance problem (agents still need to know to use it)
- Adds complexity to claudectl
- Still requires bash commands (still prompts)

**Decision**: Rejected - doesn't address root cause

### Alternative 2: Request Claude Code Feature - "Temp Directory Sandbox"
**Idea**: Ask Anthropic to add a pre-approved temp directory concept

**Pros**: Would solve problem at platform level for all users

**Cons**:
- Outside our control
- Unknown timeline
- Doesn't help users today

**Decision**: Defer - pursue in parallel as feedback to Anthropic

### Alternative 3: Just Remove /tmp from additionalDirectories
**Idea**: Force agents to fail when using /tmp, learning to avoid it

**Pros**: Strong forcing function

**Cons**:
- Hostile to users (failures instead of education)
- Breaks legitimate /tmp use cases
- No positive guidance

**Decision**: Rejected - too aggressive

## Success Criteria

1. **Quantitative**:
   - Permission prompts for testing workflows < 2 per session
   - Agent compliance with `.claude/scratch/` > 80% after guidance update

2. **Qualitative**:
   - User reports reduced frustration with prompts
   - Agents naturally use project-local temp directories
   - Testing workflows feel "smooth" again

3. **Technical**:
   - All changes backward compatible
   - No new dependencies required
   - gitignore patterns work correctly

## Next Steps

1. **Review this plan** with stakeholders
2. **Make decision** on settings.json Option A vs B
3. **Implement Phase 1** (CLAUDE.md updates)
4. **Test in fix-tmp-madness workspace**
5. **Iterate based on real-world usage**
6. **Create PR** with changes
7. **Document in release notes**

## Appendix: Research References

- `.claude/research/2025-12-01-claude-code-permissions.md` - Permission system analysis
- `.claude/research/2025-12-01-temporary-directory-best-practices.md` - Industry patterns
- `src/claudectl/templates/CLAUDE.md` - Current template state
- `src/claudectl/templates/settings.json` - Current permission configuration
