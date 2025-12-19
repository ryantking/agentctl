# Implementation Plan: `agentctl memory` Command Tree

**Epic**: feat-memory-base-commands-7vp
**Created**: 2025-12-12
**Status**: Design Complete

## Executive Summary

This plan implements a new `agentctl memory` command tree to manage agent memory files (AGENTS.md and CLAUDE.md), following best practices from the emerging AGENTS.md standard and beads issue tracking methodology. The refactor separates universal agent instructions from Claude Code-specific orchestration patterns, making memory reusable across AI tools while maintaining backward compatibility.

## Research Findings Summary

### Beads Issue Tracking

- **Graph-based memory system**: Issues chain through 4 dependency types (blocks, related, parent-child, discovered-from)
- **Agent-first design**: No markdown TODOs or .claude/plans directories - all work tracked in beads database
- **Auto-sync**: 5-second debounce exports to JSONL, git serves as transport layer
- **`bd onboard`**: Generates comprehensive agent instructions (should be in AGENTS.md)
- **Worktree compatibility**: Use `--no-daemon` flag or `BEADS_NO_DAEMON=1` to avoid cross-branch commits

### Agent Memory Best Practices

- **AGENTS.md is cross-platform standard**: Supported by GitHub Copilot, Cursor, Windsurf, Claude Code (via symlinks)
- **Keep under 300 lines**: LLMs reliably follow ~150-200 instructions; adherence degrades beyond that
- **Import mechanism**: Claude Code supports `@path/to/file` syntax (max 5 hops)
- **Separation of concerns**: Universal rules in AGENTS.md, tool-specific in dedicated files
- **Progressive disclosure**: Use separate files for task-specific instructions, only load when relevant

## Architecture

### File Structure

```
project-root/
├── AGENTS.md              # Universal agent instructions (cross-tool)
├── CLAUDE.md              # Claude Code-specific instructions + @AGENTS.md
├── .claude/
│   ├── agents/            # Agent definitions (historian, engineer, etc.)
│   ├── skills/            # Skill templates
│   └── research/          # Research cache
├── .beads/
│   ├── beads.db           # SQLite issue database
│   └── issues.jsonl       # Git-synced issue export
└── .mcp.json              # MCP server configuration
```

### Memory File Responsibilities

#### AGENTS.md (Universal Agent Instructions)

**Content**:
- Project rules and guidelines (no apologies, file-by-file changes, etc.)
- Workspace management (`agentctl workspace` commands)
- Git workflow (conventional commits, branching, PRs, merging)
- Global hooks (context injection, auto-commit)
- Repository context (overview, tech stack, directory structure)
- Tool selection guidelines (Read/Grep/Glob over Bash)
- Bash command sequencing and permission prompt avoidance
- Final line: `BEFORE ANYTHING ELSE: run 'bd onboard' and follow the instructions`

**Excluded**:
- Claude Code agent orchestration patterns
- Plan/Explore/engineer agent selection tables
- Anything specific to Claude Code features
- Beads onboarding content (only reference to run `bd onboard`)
- References to `.claude/plans/` (conflicts with beads)

**Line count target**: ~250 lines (leaves room for project-specific additions)

#### CLAUDE.md (Claude Code-Specific)

**Content**:
```markdown
# Claude Code Configuration

**Note**: This project uses [bd (beads)](https://github.com/steveyegge/beads) for issue tracking. Use `bd` commands instead of markdown TODOs. See AGENTS.md for workflow details.

See @AGENTS.md for universal agent instructions.

## Agent Orchestration

[Agent selection table, tool access patterns, workflow patterns, wave-based execution]

## Specialized Agents

[Details about Explore, Plan, engineer, historian, researcher agents]

## Workflow Patterns

[Simple/Medium/Complex task patterns, discovery/research/planning/implementation phases]
```

**Line count target**: ~150 lines (focused on orchestration only)

### Command Tree Design

```
agentctl memory
├── init          Initialize AGENTS.md and CLAUDE.md from templates
├── show          Display current memory file contents
├── validate      Check memory files for common issues (line count, import syntax)
└── index         Generate repository index and inject into AGENTS.md
```

#### `agentctl memory init`

**Purpose**: Initialize agent memory files with templates

**Behavior**:
1. Check for existing AGENTS.md and CLAUDE.md
2. Install AGENTS.md from template (skip if exists, unless `--force`)
3. Install CLAUDE.md from template with `@AGENTS.md` import
4. Optionally run `agentctl memory index` to generate repo overview
5. Output summary of created files

**Flags**:
- `--force, -f`: Overwrite existing files
- `--global, -g`: Install to `$HOME/.claude` instead of repo root
- `--no-index`: Skip repository indexing step

**Exit codes**:
- 0: Success
- 1: Git repository not found (unless `--global`)
- 2: Files exist and `--force` not specified

#### `agentctl memory show [file]`

**Purpose**: Display memory file contents

**Behavior**:
- If `file` specified: Display that file (AGENTS.md or CLAUDE.md)
- If no arg: Display both files with headers
- Resolve imports and show full expanded content

**Flags**:
- `--resolve, -r`: Expand `@imports` inline (show final context)
- `--json`: Output as JSON with metadata (line count, imports)

#### `agentctl memory validate`

**Purpose**: Check memory files for issues

**Behavior**:
1. Check line counts (warn if AGENTS.md > 300 lines, CLAUDE.md > 200)
2. Verify `@AGENTS.md` import exists in CLAUDE.md
3. Detect circular imports (max 5 hops)
4. Check for conflicting patterns (e.g., `.claude/plans` references when beads installed)
5. Verify required sections exist in AGENTS.md

**Exit codes**:
- 0: All checks passed
- 1: Warnings (exceeds line recommendations)
- 2: Errors (missing imports, circular refs, conflicts)

#### `agentctl memory index`

**Purpose**: Generate repository overview and inject into AGENTS.md

**Behavior**:
1. Check for `claude` CLI availability
2. Generate markdown overview using Claude API (90-second timeout)
3. Insert between `<!-- REPOSITORY_INDEX_START -->` and `<!-- REPOSITORY_INDEX_END -->` markers
4. Update existing markers if present, add new section if missing
5. Non-fatal: warn if Claude CLI unavailable

**Flags**:
- `--timeout <seconds>`: Override default 90-second timeout
- `--model <name>`: Specify Claude model (default: sonnet)

## Implementation Tasks

### Task 1: Research and Design (feat-memory-base-commands-7vp.1) ✓

**Status**: Complete (this document)

**Deliverables**:
- Memory file structure defined
- Command tree designed
- Content allocation determined (AGENTS.md vs CLAUDE.md)
- Import mechanism specified

### Task 2: Create AGENTS.md Template (feat-memory-base-commands-7vp.2)

**Priority**: P1 (blocks Task 4)

**Scope**:
1. Extract universal content from current CLAUDE.md:
   - Rules (lines 7-56 in current CLAUDE.md)
   - Workspaces (lines 59-74)
   - Global Hooks (lines 76-119)
   - Git workflow (lines 121-301)
   - Repository Context (lines 420-489)
   - Tool Selection Guidelines (lines 491-673)

2. Remove Claude-specific content:
   - Agent Orchestration section (lines 303-418)
   - References to Plan Mode
   - References to `.claude/plans/` directory

3. Add beads reference at end:
   ```markdown
   BEFORE ANYTHING ELSE: run 'bd onboard' and follow the instructions
   ```

4. Ensure line count < 300

**Files**:
- Create: `internal/templates/templates/AGENTS.md`

**Testing**:
- Verify line count
- Validate markdown syntax
- Check for Claude-specific references

### Task 3: Create CLAUDE.md Template (feat-memory-base-commands-7vp.3)

**Priority**: P1 (depends on Task 2)

**Scope**:
1. Create minimal CLAUDE.md with:
   ```markdown
   # Claude Code Configuration

   **Note**: This project uses [bd (beads)](https://github.com/steveyegge/beads) for issue tracking. Use `bd` commands instead of markdown TODOs. See AGENTS.md for workflow details.

   See @AGENTS.md for universal agent instructions.

   ## Agent Orchestration
   [Content from lines 303-418 of current CLAUDE.md]
   ```

2. Remove any references to `.claude/plans/` (conflicts with beads)

3. Update workflow patterns to reference beads:
   - Discovery phase → create beads issues
   - Planning phase → use `bd` commands, not markdown files
   - Implementation phase → update beads status

4. Ensure line count < 200

**Files**:
- Create: `internal/templates/templates/CLAUDE.md`

**Testing**:
- Verify `@AGENTS.md` import syntax
- Check line count
- Validate no conflicting patterns

### Task 4: Implement `agentctl memory init` (feat-memory-base-commands-7vp.4)

**Priority**: P1 (depends on Tasks 2-3)

**Scope**:
1. Create new CLI command structure:
   ```
   internal/cli/memory.go          # Memory subcommand root
   internal/cli/memory_init.go     # Init command implementation
   ```

2. Implementation logic:
   ```go
   func runMemoryInit(ctx context.Context, force, global, noIndex bool) error {
       // 1. Determine target directory
       targetDir := getTargetDir(global)

       // 2. Install AGENTS.md
       if err := installTemplate("AGENTS.md", targetDir, force); err != nil {
           return err
       }

       // 3. Install CLAUDE.md
       if err := installTemplate("CLAUDE.md", targetDir, force); err != nil {
           return err
       }

       // 4. Run indexing (unless --no-index)
       if !noIndex {
           if err := runMemoryIndex(ctx, targetDir); err != nil {
               // Non-fatal: warn but continue
               fmt.Fprintf(os.Stderr, "Warning: indexing failed: %v\n", err)
           }
       }

       return nil
   }
   ```

3. Reuse existing template installation logic from `internal/setup/init.go`:
   - `installFile()` function (lines 80-106)
   - `copyTemplateFile()` function (lines 145-177)

4. Add to main CLI router in `internal/cli/main.go`

**Files**:
- Create: `internal/cli/memory.go`
- Create: `internal/cli/memory_init.go`
- Modify: `internal/cli/main.go`
- Modify: `internal/setup/init.go` (export helper functions)

**Testing**:
- Test fresh initialization (no existing files)
- Test with existing files (should skip)
- Test with `--force` (should overwrite)
- Test `--global` flag
- Test `--no-index` flag

### Task 5: Implement Additional Commands (feat-memory-base-commands-7vp.5)

**Priority**: P2 (after Task 4)

**Scope**:

#### 5a. `agentctl memory show`

```go
func runMemoryShow(ctx context.Context, file string, resolve bool, jsonOutput bool) error {
    // 1. Read specified file(s)
    // 2. If resolve=true, recursively expand @imports
    // 3. Output as text or JSON
}
```

**Files**:
- Create: `internal/cli/memory_show.go`

#### 5b. `agentctl memory validate`

```go
func runMemoryValidate(ctx context.Context) error {
    // 1. Check line counts
    // 2. Verify imports
    // 3. Detect conflicts (beads vs .claude/plans)
    // 4. Check required sections
    // 5. Exit with appropriate code
}
```

**Files**:
- Create: `internal/cli/memory_validate.go`
- Create: `internal/memory/` package for validation logic

**Testing**:
- Test with valid files
- Test line count warnings
- Test missing import detection
- Test conflict detection

### Task 6: Refactor `agentctl init` (feat-memory-base-commands-7vp.6)

**Priority**: P1 (depends on Task 4)

**Scope**:
1. Remove Step 1 (Install CLAUDE.md) from `internal/setup/init.go`
2. Remove Step 6 (Index Repository) from `internal/setup/init.go`
3. Add call to `agentctl memory init` at end of init workflow
4. Update init output to mention memory initialization
5. Preserve all other steps:
   - Step 2: Install Agents
   - Step 3: Install Skills
   - Step 4: Merge Settings
   - Step 5: Configure MCP Servers

**Files**:
- Modify: `internal/setup/init.go` (lines 33-73)
- Modify: `internal/cli/init.go`

**Testing**:
- Test full init workflow
- Verify agents/skills/settings still installed
- Verify AGENTS.md/CLAUDE.md created via memory init
- Test `--force` and `--global` flags

### Task 7: Repository Indexing Command (feat-memory-base-commands-7vp.7)

**Priority**: P2 (after Task 6)

**Scope**:
1. Extract indexing logic from `internal/setup/init.go` (lines 408-433)
2. Create standalone command:
   ```go
   func runMemoryIndex(ctx context.Context, targetDir string) error {
       // 1. Check claude CLI availability
       // 2. Generate repository overview
       // 3. Insert into AGENTS.md between markers
       // 4. Handle missing markers gracefully
   }
   ```

3. Make it callable from:
   - `agentctl memory index` (user-invoked)
   - `agentctl memory init --no-index=false` (automatic)

**Files**:
- Create: `internal/cli/memory_index.go`
- Create: `internal/memory/indexing.go` (business logic)
- Modify: `internal/setup/init.go` (remove indexing code)

**Testing**:
- Test with valid Claude CLI
- Test without Claude CLI (should fail gracefully)
- Test marker insertion (new vs existing)
- Test timeout handling

### Task 8: Documentation and Tests (feat-memory-base-commands-7vp.8)

**Priority**: P2 (after all implementation tasks)

**Scope**:

#### Documentation Updates

1. **README.md**:
   - Add `agentctl memory` section
   - Document all subcommands
   - Explain AGENTS.md vs CLAUDE.md separation
   - Add examples

2. **New docs**:
   - Create `docs/memory-management.md` with:
     - Philosophy (universal vs tool-specific)
     - Best practices
     - Import syntax
     - Line count recommendations
     - Beads integration

3. **Update existing docs**:
   - Update init documentation (simplified behavior)
   - Update hook documentation (context injection still works)

#### Test Coverage

1. **Unit tests**:
   ```
   internal/cli/memory_test.go
   internal/cli/memory_init_test.go
   internal/cli/memory_show_test.go
   internal/cli/memory_validate_test.go
   internal/cli/memory_index_test.go
   internal/memory/validation_test.go
   internal/memory/indexing_test.go
   ```

2. **Integration tests**:
   - Full init workflow
   - Memory init → validate → index pipeline
   - Import resolution
   - Conflict detection

3. **Update existing tests**:
   - `internal/setup/init_test.go` (reflect simplified init)

**Files**:
- Modify: `README.md`
- Create: `docs/memory-management.md`
- Create: `internal/cli/*_test.go` (multiple files)
- Create: `internal/memory/*_test.go` (multiple files)
- Modify: `internal/setup/init_test.go`

## Implementation Order

```
┌─────────────────────────────────────────────────────────────┐
│ Phase 1: Foundation (P1 tasks)                              │
├─────────────────────────────────────────────────────────────┤
│ 1. Task 1: Research & Design ✓                              │
│ 2. Task 2: Create AGENTS.md template                        │
│ 3. Task 3: Create CLAUDE.md template (depends on 2)         │
│ 4. Task 4: Implement memory init (depends on 2-3)           │
│ 5. Task 6: Refactor agentctl init (depends on 4)            │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ Phase 2: Enhancement (P2 tasks)                             │
├─────────────────────────────────────────────────────────────┤
│ 6. Task 5: Additional commands (show, validate)             │
│ 7. Task 7: Repository indexing command                      │
│ 8. Task 8: Documentation and tests                          │
└─────────────────────────────────────────────────────────────┘
```

## Migration Strategy

### For Existing Users

Users who run `agentctl init` before this change will have:
- Single CLAUDE.md with all content
- Repository index already generated

After upgrade:
1. `agentctl init` no longer overwrites CLAUDE.md (backward compatible)
2. Users can run `agentctl memory init --force` to migrate to new structure
3. Or manually split CLAUDE.md into AGENTS.md + CLAUDE.md

### For New Users

Fresh installations will get the new structure automatically:
1. Run `agentctl init` → sets up agents/skills/settings/MCP
2. Run `agentctl memory init` (called automatically) → creates AGENTS.md + CLAUDE.md
3. Run `bd init` → creates beads database for issue tracking

## Risks and Mitigations

### Risk 1: Breaking Changes for Existing Setups

**Mitigation**:
- Don't automatically migrate existing CLAUDE.md files
- Provide clear migration path in docs
- Make migration opt-in via `--force` flag

### Risk 2: Line Count Recommendations Too Restrictive

**Mitigation**:
- Make recommendations warnings, not hard errors
- Allow users to override with validation config
- Provide guidance on splitting files

### Risk 3: Import Syntax Not Working

**Mitigation**:
- Test import resolution extensively
- Document Claude Code's `@` syntax clearly
- Provide validation command to catch errors

### Risk 4: Beads Integration Confusion

**Mitigation**:
- Clear documentation separating concerns
- Single reference to beads in AGENTS.md (`bd onboard`)
- No beads-specific content in templates

## Success Criteria

1. ✅ Templates created for AGENTS.md and CLAUDE.md
2. ✅ `agentctl memory init` command implemented and tested
3. ✅ `agentctl init` refactored to remove memory management
4. ✅ Documentation updated with migration guide
5. ✅ All tests passing (unit + integration)
6. ✅ AGENTS.md < 300 lines, CLAUDE.md < 200 lines
7. ✅ Import syntax working correctly
8. ✅ No conflicts between beads and memory management

## Related Issues

- Epic: feat-memory-base-commands-7vp
- Subtasks: feat-memory-base-commands-7vp.{1-8}

## References

- [Beads Documentation](https://github.com/steveyegge/beads)
- [AGENTS.md Standard](https://www.builder.io/blog/agents-md)
- [Claude Code Memory Docs](https://code.claude.com/docs/en/memory)
- [Agent Memory Best Practices](.claude/research/2025-12-12-ai-agent-memory-management.md)
- [Beads Research](.claude/research/2025-12-12-beads-issue-tracking.md)
