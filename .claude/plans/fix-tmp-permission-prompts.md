# Fix /tmp Permission Prompt Madness

## Problem Statement

Claude Code agents are triggering excessive permission prompts when performing testing and temporary file operations. The root cause is that agents default to using `/tmp` for temporary files, and each unique bash command path requires a separate permission prompt due to Claude's prefix-based approval system.

**Current Pain Points:**
- Agents create temp directories in `/tmp` via bash commands
- Each unique bash command (mkdir, cat, rm) triggers a permission prompt
- Even with `/tmp` in `additionalDirectories`, bash commands still require approval
- Testing workflows become unusable due to prompt fatigue

## Root Cause Analysis

### Why /tmp Triggers So Many Prompts

1. **Bash vs Built-in Tools**: The permission system treats these differently:
   - Built-in tools (Read, Grep, Glob): Permission-free within `additionalDirectories`
   - Bash commands: Always require approval, even for `/tmp` operations

2. **Prefix Matching Limitations**: Claude's bash approval uses exact prefix matching:
   ```
   mkdir /tmp/test-foo       # Prompt 1
   echo "data" > /tmp/test-foo/file.txt  # Prompt 2
   cat /tmp/test-foo/file.txt            # Prompt 3
   rm -rf /tmp/test-foo                  # Prompt 4
   ```
   Each unique command is treated as a new permission request.

3. **No Guidance in CLAUDE.md**: The template provides zero guidance about:
   - Where to create temporary files
   - When to use `/tmp` vs project-local directories
   - How to avoid permission prompt cascades
   - Best practices for temporary file management

### Why This Happens

From the research:
- claudectl itself doesn't use `/tmp` in its source code
- The CLAUDE.md template doesn't instruct agents on temp file practices
- Agents default to standard Unix behavior (using `/tmp`)
- The permission system can't distinguish "temporary testing operations" from "potentially dangerous operations"

## Solution Design

### Three-Pronged Approach

#### 1. Update CLAUDE.md Template with Explicit Guidance

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

### Phase 1: Update CLAUDE.md Template (High Priority)

**File**: `src/claudectl/templates/CLAUDE.md`

**Tasks**:
1. Add "Temporary Files and Directories" section after line 88
2. Update "Anti-Patterns" section with /tmp examples (lines 90-102)
3. Expand Rule 24 about working directory usage (line 55)
4. Add temp file guidance to Agent Orchestration Key Rules (after line 481)

**Impact**: Immediately reduces prompt fatigue for new projects initialized with `claudectl init`

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
