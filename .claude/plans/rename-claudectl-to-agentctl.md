# Repository Rename Plan: claudectl → agentctl

**Status**: Ready for implementation
**Created**: 2025-12-05
**GitHub Rename**: ✅ Complete (ryantking/agentctl)
**Remote URL**: ✅ Updated automatically

## Overview

This plan provides comprehensive guidance for renaming all source code references from `claudectl` to `agentctl` following the GitHub repository rename. The plan is organized by change type to enable systematic implementation.

## Scope Summary

- **30 files** contain references to "claudectl"
- **1 directory** needs renaming: `src/claudectl/` → `src/agentctl/`
- **5 main change categories**: Python package structure, documentation, configuration, URLs, tooling

## Implementation Strategy

Execute changes in this order to maintain working state throughout:

1. **Phase 1**: Rename source directory (structural change)
2. **Phase 2**: Update Python imports and package references
3. **Phase 3**: Update configuration files (pyproject.toml, settings.json)
4. **Phase 4**: Update documentation (README, CLAUDE.md, templates)
5. **Phase 5**: Update supporting files (workflows, research notes, plans)
6. **Phase 6**: Rebuild virtual environment and verify

## Phase 1: Directory Structure

### Source Directory Rename
**Action**: Rename the main package directory
```bash
git mv src/claudectl src/agentctl
```

**Impact**: This is the foundational change. All subsequent Python imports will need updating.

## Phase 2: Python Package References

### A. Package Name in pyproject.toml
**File**: `pyproject.toml`

**Line 2**: Package name declaration
```toml
# BEFORE
name = "claudectl"

# AFTER
name = "agentctl"
```

**Line 30**: CLI entry point
```toml
# BEFORE
claudectl = "claudectl.cli.main:main"

# AFTER
agentctl = "agentctl.cli.main:main"
```

**Lines 36-40**: Project URLs
```toml
# BEFORE
Homepage = "https://github.com/ryantking/claudectl"
Repository = "https://github.com/ryantking/claudectl"
Download = "https://github.com/ryantking/claudectl/releases/latest"
Issues = "https://github.com/ryantking/claudectl/issues"
Changelog = "https://github.com/ryantking/claudectl/releases"

# AFTER
Homepage = "https://github.com/ryantking/agentctl"
Repository = "https://github.com/ryantking/agentctl"
Download = "https://github.com/ryantking/agentctl/releases/latest"
Issues = "https://github.com/ryantking/agentctl/issues"
Changelog = "https://github.com/ryantking/agentctl/releases"
```

**Line 78**: Coverage target
```toml
# BEFORE
"--cov=claudectl",

# AFTER
"--cov=agentctl",
```

### B. Python Import Statements

All Python files with `from claudectl` imports need updating. Use global find-replace:

**Pattern**: `from claudectl` → `from agentctl`
**Pattern**: `import claudectl` → `import agentctl`

**Files affected** (11 files):
1. `src/agentctl/__init__.py` (line 6)
   ```python
   # BEFORE
   __version__ = version("claudectl")
   # AFTER
   __version__ = version("agentctl")
   ```

2. `src/agentctl/cli/main.py` (lines 13-16, 87, 92)
   ```python
   # BEFORE
   from claudectl.cli.commands.hook import app as hook_app
   from claudectl.cli.commands.init import app as init_app
   from claudectl.cli.commands.workspace import app as workspace_app
   from claudectl.cli.output import CLIError, handle_exception, is_json_output, set_json_output
   from claudectl import __version__
   typer.echo(f"claudectl {__version__}")

   # AFTER
   from agentctl.cli.commands.hook import app as hook_app
   from agentctl.cli.commands.init import app as init_app
   from agentctl.cli.commands.workspace import app as workspace_app
   from agentctl.cli.output import CLIError, handle_exception, is_json_output, set_json_output
   from agentctl import __version__
   typer.echo(f"agentctl {__version__}")
   ```

3. `src/agentctl/cli/main.py` (line 39)
   ```python
   # BEFORE
   name="claudectl",
   # AFTER
   name="agentctl",
   ```

4. `src/agentctl/cli/commands/init.py` (lines 11-13, 69)
   ```python
   # BEFORE
   from claudectl.cli.output import Result, output
   from claudectl.core.git import NotInGitRepoError
   from claudectl.operations.init_ops import ImportDirNotFoundError, InitManager
   from claudectl.core.git import get_repo_root

   # AFTER
   from agentctl.cli.output import Result, output
   from agentctl.core.git import NotInGitRepoError
   from agentctl.operations.init_ops import ImportDirNotFoundError, InitManager
   from agentctl.core.git import get_repo_root
   ```

5. `src/agentctl/core/workspaces.py` (line 10)
   ```python
   # BEFORE
   from claudectl.core.git import get_repo_name, get_repo_root, is_worktree_clean
   # AFTER
   from agentctl.core.git import get_repo_name, get_repo_root, is_worktree_clean
   ```

6. `src/agentctl/operations/init_ops.py` (line 12)
   ```python
   # BEFORE
   from claudectl.operations.settings_merge import merge_settings_smart
   # AFTER
   from agentctl.operations.settings_merge import merge_settings_smart
   ```

7. `src/agentctl/operations/spawn.py` (line 9)
   ```python
   # BEFORE
   from claudectl.operations.mcp_config import generate_mcp_config
   # AFTER
   from agentctl.operations.mcp_config import generate_mcp_config
   ```

8. `src/agentctl/operations/context.py` (line 8)
   ```python
   # BEFORE
   from claudectl.core.git import get_repo_root
   # AFTER
   from agentctl.core.git import get_repo_root
   ```

### C. Module Docstrings

Update module-level docstrings that mention "claudectl":

1. `src/agentctl/__init__.py` (line 1)
   ```python
   # BEFORE
   """claudectl - CLI for managing Claude Code configurations."""
   # AFTER
   """agentctl - CLI for managing Claude Code configurations."""
   ```

2. `src/agentctl/cli/main.py` (line 1)
   ```python
   # BEFORE
   """claudectl - CLI for managing Claude Code configurations."""
   # AFTER
   """agentctl - CLI for managing Claude Code configurations."""
   ```

3. `src/agentctl/core/__init__.py` (line 1)
   ```python
   # BEFORE
   """Core utilities for claudectl."""
   # AFTER
   """Core utilities for agentctl."""
   ```

4. `src/agentctl/operations/spawn.py` (line 1)
   ```python
   # BEFORE
   """Process spawning utilities for claudectl."""
   # AFTER
   """Process spawning utilities for agentctl."""
   ```

5. `src/agentctl/operations/init_ops.py` (line 1)
   ```python
   # BEFORE
   """Core initialization operations for claudectl init command."""
   # AFTER
   """Core initialization operations for agentctl init command."""
   ```

6. `src/agentctl/cli/commands/init.py` (line 1)
   ```python
   # BEFORE
   """Init command for claudectl - initialize Claude Code configuration."""
   # AFTER
   """Init command for agentctl - initialize Claude Code configuration."""
   ```

### D. Class and Function Docstrings

1. `src/agentctl/core/workspaces.py` (lines 15, 18, 30)
   ```python
   # BEFORE
   """A git worktree managed by claudectl.
   ...
   by claudectl. Workspaces are discovered from git's authoritative
   ...
   """Check if this workspace is managed by claudectl.

   # AFTER
   """A git worktree managed by agentctl.
   ...
   by agentctl. Workspaces are discovered from git's authoritative
   ...
   """Check if this workspace is managed by agentctl.
   ```

## Phase 3: Configuration Files

### A. Claude Code Settings

**File**: `.claude/settings.json`

**Line 54**: Bash permission pattern
```json
// BEFORE
"Bash(claudectl:*)",

// AFTER
"Bash(agentctl:*)",
```

**Lines 134, 144, 154, 165, 174, 184**: Hook commands
```json
// BEFORE
"command": "claudectl hook context-info"
"command": "claudectl hook notify-input"
"command": "claudectl hook notify-stop"
"command": "claudectl hook post-edit"
"command": "claudectl hook post-write"
"command": "claudectl hook context-info"

// AFTER
"command": "agentctl hook context-info"
"command": "agentctl hook notify-input"
"command": "agentctl hook notify-stop"
"command": "agentctl hook post-edit"
"command": "agentctl hook post-write"
"command": "agentctl hook context-info"
```

### B. Template Settings

**File**: `src/agentctl/templates/settings.json`

Same changes as `.claude/settings.json` above (lines 54, 134, 144, 154, 165, 174, 184).

### C. Justfile

**File**: `Justfile`

**Line 10**: Comment
```make
# BEFORE
# Install claudectl globally in editable mode

# AFTER
# Install agentctl globally in editable mode
```

### D. UV Lock File

**File**: `uv.lock`

**Line 93**: Package name entry
```toml
# BEFORE
name = "claudectl"

# AFTER
name = "agentctl"
```

**Note**: This file may regenerate automatically on next `uv sync`. Verify after rebuild.

## Phase 4: Documentation

### A. README.md

**File**: `README.md`

**Line 1**: Title
```markdown
# BEFORE
# claudectl

# AFTER
# agentctl
```

**Lines 19, 22, 29, 32**: Installation URLs
```markdown
# BEFORE
pip install https://github.com/carelesslisper/claudectl/releases/latest/download/claudectl.tar.gz
pip install https://github.com/carelesslisper/claudectl/releases/download/v0.1.0/claudectl-0.1.0.tar.gz
uv tool install https://github.com/carelesslisper/claudectl/releases/latest/download/claudectl.tar.gz
uv tool install https://github.com/carelesslisper/claudectl/releases/download/v0.1.0/claudectl-0.1.0.tar.gz

# AFTER
pip install https://github.com/ryantking/agentctl/releases/latest/download/agentctl.tar.gz
pip install https://github.com/ryantking/agentctl/releases/download/v0.1.0/agentctl-0.1.0.tar.gz
uv tool install https://github.com/ryantking/agentctl/releases/latest/download/agentctl.tar.gz
uv tool install https://github.com/ryantking/agentctl/releases/download/v0.1.0/agentctl-0.1.0.tar.gz
```

**Note**: Username changed from `carelesslisper` → `ryantking` (matches actual GitHub)

**Lines 38-39**: Clone instructions
```markdown
# BEFORE
git clone https://github.com/carelesslisper/claudectl.git
cd claudectl

# AFTER
git clone https://github.com/ryantking/agentctl.git
cd agentctl
```

**Lines 47-62**: CLI examples (all commands)
```markdown
# BEFORE
claudectl status
claudectl version
claudectl workspace create my-feature-branch
claudectl workspace list
claudectl workspace status my-feature-branch
claudectl workspace delete my-feature-branch

# AFTER
agentctl status
agentctl version
agentctl workspace create my-feature-branch
agentctl workspace list
agentctl workspace status my-feature-branch
agentctl workspace delete my-feature-branch
```

**Lines 69-83**: Command reference
```markdown
# BEFORE
- `claudectl workspace create <branch>` - Create new workspace with git worktree
- `claudectl workspace list [--json]` - List all managed workspaces
- `claudectl workspace show <branch>` - Print workspace path
- `claudectl workspace status <branch>` - Show detailed workspace status
- `claudectl workspace delete <branch>` - Delete a workspace
- `claudectl workspace clean` - Remove all clean workspaces
- `claudectl hook post-edit` - Auto-commit Edit tool changes
- `claudectl hook post-write` - Auto-commit Write tool changes (new files)
- `claudectl hook context-info` - Inject git/workspace context into prompts
- `claudectl hook notify-*` - Notification commands

# AFTER
- `agentctl workspace create <branch>` - Create new workspace with git worktree
- `agentctl workspace list [--json]` - List all managed workspaces
- `agentctl workspace show <branch>` - Print workspace path
- `agentctl workspace status <branch>` - Show detailed workspace status
- `agentctl workspace delete <branch>` - Delete a workspace
- `agentctl workspace clean` - Remove all clean workspaces
- `agentctl hook post-edit` - Auto-commit Edit tool changes
- `agentctl hook post-write` - Auto-commit Write tool changes (new files)
- `agentctl hook context-info` - Inject git/workspace context into prompts
- `agentctl hook notify-*` - Notification commands
```

### B. CLAUDE.md

**File**: `CLAUDE.md`

**All command references** (lines 63, 65-72, 76, 82, 196, 291, 423, 443-444, 461, 480, 483, 486, 606):

Replace all instances:
- `claudectl` → `agentctl`
- `src/claudectl/` → `src/agentctl/`
- `claudectl.cli.main:main` → `agentctl.cli.main:main`
- `.cache/claudectl/` → `.cache/agentctl/`

**Key sections to update**:

1. **Workspaces section** (lines 63-72): All command examples
2. **Global Hooks section** (line 76): "Hooks are provided by `agentctl hooks` commands"
3. **Context Injection** (line 82): Reference to agentctl workspace
4. **Workflow sections** (lines 196, 291): Delete workspace commands
5. **Repository Context** (lines 423-486): Entire repository overview section
6. **Tool Selection Guidelines** (line 606): Cache directory reference

### C. Template CLAUDE.md

**File**: `src/agentctl/templates/CLAUDE.md`

**Same changes as CLAUDE.md** above. This is the template used by `agentctl init`, so it must match.

**Lines to update**: 63, 65-72, 76, 82, 196, 291, 541

## Phase 5: Supporting Files

### A. GitHub Workflow

**File**: `.github/workflows/release.yml`

**Lines 43, 48, 53-54**: Install verification examples
```yaml
# BEFORE
pip install https://github.com/ryantking/claudectl/releases/download/${{ github.ref_name }}/claudectl-${{ github.ref_name }}.tar.gz
uv tool install https://github.com/ryantking/claudectl/releases/download/${{ github.ref_name }}/claudectl-${{ github.ref_name }}.tar.gz
git clone https://github.com/ryantking/claudectl.git
cd claudectl

# AFTER
pip install https://github.com/ryantking/agentctl/releases/download/${{ github.ref_name }}/agentctl-${{ github.ref_name }}.tar.gz
uv tool install https://github.com/ryantking/agentctl/releases/download/${{ github.ref_name }}/agentctl-${{ github.ref_name }}.tar.gz
git clone https://github.com/ryantking/agentctl.git
cd agentctl
```

### B. Research Notes

**Files** (5 files in `.claude/research/`):
- `2025-11-30-homebrew-github-packaging.md`
- `2025-11-30-python-uv-build-distribution.md`
- `2025-12-01-claude-code-bash-chaining-permissions.md`
- `2025-12-01-claude-code-parallel-tool-calls.md`
- `2025-12-01-claude-code-permissions.md`
- `2025-12-01-temporary-directory-best-practices.md`

**Action**: Use global find-replace across all research files:
- `claudectl` → `agentctl`
- `src/claudectl` → `src/agentctl`
- `.cache/claudectl` → `.cache/agentctl`

**Note**: These are historical documents, so updates are for consistency but not critical.

### C. Plan Documents

**File**: `.claude/plans/fix-tmp-permission-prompts.md`

**Lines 89, 169, 192**: Path references
```markdown
# BEFORE
**Location:** `src/claudectl/templates/CLAUDE.md`
**Location**: `src/claudectl/templates/CLAUDE.md`
3. **For Build/Runtime Caches** → Use `.cache/claudectl/` (gitignored)

# AFTER
**Location:** `src/agentctl/templates/CLAUDE.md`
**Location**: `src/agentctl/templates/CLAUDE.md`
3. **For Build/Runtime Caches** → Use `.cache/agentctl/` (gitignored)
```

### D. Skills

**File**: `src/agentctl/templates/skills/linear/SKILL.md`

**Lines 987, 1014, 1019**: Session management examples
```markdown
# BEFORE
- Check multiplexing: `claudectl session check`
5. Start parallel work: claudectl session create for each
claudectl session list

# AFTER
- Check multiplexing: `agentctl session check`
5. Start parallel work: agentctl session create for each
agentctl session list
```

**Note**: These may be future features. Update for consistency.

## Phase 6: Rebuild and Verification

### A. Clean and Rebuild Virtual Environment

```bash
# Remove old virtual environment
rm -rf .venv

# Sync dependencies (will regenerate uv.lock with new package name)
uv sync

# Verify package is installed
uv run python -c "import agentctl; print(agentctl.__version__)"

# Run all tests
just test
```

### B. Reinstall as Tool (if applicable)

```bash
# Uninstall old claudectl
uv tool uninstall claudectl

# Install new agentctl
just install

# Verify CLI works
agentctl --version
agentctl --help
```

### C. Verification Checklist

- [ ] Package imports as `agentctl`
- [ ] CLI binary is named `agentctl`
- [ ] `agentctl --version` shows correct version
- [ ] All tests pass
- [ ] `agentctl init` creates templates with correct references
- [ ] Hooks in `.claude/settings.json` reference `agentctl`
- [ ] README examples use `agentctl`
- [ ] GitHub URLs point to `ryantking/agentctl`

## Post-Implementation Tasks

### Update Next Release

1. **Version bump**: Consider bumping to 0.2.0 (minor) due to CLI rename
2. **Release notes**: Document breaking change (CLI binary name change)
3. **Migration guide**: Add note that users need to:
   - Uninstall `claudectl`: `uv tool uninstall claudectl`
   - Install `agentctl`: `uv tool install agentctl`
   - Update `.claude/settings.json` hooks to use `agentctl`

### GitHub Repository Settings

1. **Description**: Update repository description to reference "agentctl"
2. **Topics**: Add relevant tags (agent-management, claude-code, cli-tool)
3. **Redirects**: GitHub automatically redirects `claudectl` → `agentctl` URLs

## Files Not Requiring Changes

These files don't contain "claudectl" references:
- `src/agentctl/operations/settings_merge.py`
- `src/agentctl/operations/mcp_config.py`
- `src/agentctl/core/git.py`
- Test files (if they don't explicitly reference the package name)
- Build artifacts in `dist/` (will regenerate)

## Risk Assessment

**Low Risk Changes**:
- Documentation (README, CLAUDE.md, research notes)
- Comments and docstrings
- Configuration file comments

**Medium Risk Changes**:
- Python imports (automated testing will catch errors)
- CLI entry points (manual testing required)

**High Risk Changes**:
- Directory rename `src/claudectl/` → `src/agentctl/` (affects all imports)
- Package name in pyproject.toml (affects distribution)

**Mitigation**: Execute Phase 1 and Phase 2 together in a single commit to avoid broken import states.

## Implementation Commands Summary

```bash
# Phase 1: Directory rename
git mv src/claudectl src/agentctl

# Phase 2-5: Use Edit tool for all text changes
# (Implement file by file, testing incrementally)

# Phase 6: Rebuild
rm -rf .venv
uv sync
just test
just install
agentctl --version

# Commit everything
git add -A
git commit -m "feat: rename package from claudectl to agentctl

- Rename src/claudectl/ → src/agentctl/
- Update all Python imports to use agentctl
- Update CLI entry point from claudectl to agentctl
- Update all documentation and examples
- Update GitHub workflow references
- Update configuration files (.claude/settings.json)

BREAKING CHANGE: CLI binary renamed from claudectl to agentctl. Users must uninstall claudectl and reinstall as agentctl.

Co-Authored-By: Claude <noreply@anthropic.com>"

git push origin master
```

## Summary Statistics

- **Total files to modify**: 30
- **Python files**: 11
- **Config files**: 4
- **Documentation files**: 3
- **Research/plan files**: 7
- **Template files**: 3
- **Workflow files**: 1
- **Lock files**: 1

**Estimated complexity**: Medium (systematic but high volume of changes)
**Recommended approach**: Implement phases sequentially, commit after each phase
**Testing strategy**: Run test suite after Phases 1-2, verify CLI after Phase 6
