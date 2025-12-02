# Research: Temporary Directory Best Practices for CLI Tools

Date: 2025-12-01
Focus: Where should claudectl guide agents to create temporary files?
Agent: researcher

## Summary

Development tools should use system temporary directories (`TMPDIR`/`/tmp`) by default rather than project-local directories. The Python `tempfile` module provides secure, cross-platform abstractions. For workspace-aware tools like claudectl, a hybrid approach using project-local `.cache/` directories for build artifacts and system temp for ephemeral files offers the best balance.

## Key Findings

### 1. System Temporary Directories Are Preferred Default

- Use `/tmp` or `TMPDIR` for short-lived, ephemeral files [Stack Exchange](https://softwareengineering.stackexchange.com/questions/314796/should-temporary-files-be-saved-to-tmp-or-the-current-working-directory)
- **Always respect `TMPDIR` environment variable** - this is the canonical way to override temp location [systemd.io](https://systemd.io/TEMPORARY_DIRECTORIES/)
- OS ensures unique filenames, backup software skips these directories, may be RAM-backed for speed
- `/tmp` clears on reboot (10-day aging), `/var/tmp` persists (30-day aging)

### 2. When to Use Project-Local Directories

- Build artifacts that should be cached between runs → project-local `.cache/`
- Files that should be committed or tracked → project directories
- Large files that shouldn't consume RAM (if /tmp is tmpfs)
- Files needed across worktrees → must be in shared location

### 3. How Major Tools Handle Temp Files

| Tool | Location | Notes |
|------|----------|-------|
| **pytest** | System temp (`pytest-NUM` subdirs) | Uses `tmp_path` fixture, auto-cleanup, keeps last 3 runs [pytest docs](https://docs.pytest.org/en/stable/how-to/tmp_path.html) |
| **npm** | `~/.npm` (cache), `node_modules/.cache` (build) | Global cache for packages, project-local for build tools [npm docs](https://docs.npmjs.com/cli/v7/configuring-npm/folders/) |
| **cargo** | `target/` (project), system temp (install) | Build artifacts in project, temp installs use TMPDIR [Cargo issue](https://github.com/rust-lang/cargo/issues/6227) |
| **webpack/jest** | `node_modules/.cache/<tool>` | Convention for build tool caches [Medium](https://jongleberry.medium.com/speed-up-your-ci-and-dx-with-node-modules-cache-ac8df82b7bb0) |

### 4. XDG Base Directory Specification

- `XDG_RUNTIME_DIR`: Session-bound, small files, sockets, deleted on logout [Arch Wiki](https://wiki.archlinux.org/title/XDG_Base_Directory)
- `XDG_CACHE_HOME` (`~/.cache`): Persistent cache, survives sessions, regenerable
- For CLI tools, `XDG_CACHE_HOME` is appropriate for persistent caches

### 5. Python tempfile Best Practices

```python
# Preferred: Context manager with automatic cleanup
from tempfile import TemporaryDirectory

with TemporaryDirectory() as tmpdir:
    # Files automatically cleaned up on exit
    work_file = Path(tmpdir) / "scratch.txt"
```

Key features [Python docs](https://docs.python.org/3/library/tempfile.html):
- `gettempdir()`: Returns platform-appropriate temp dir (respects TMPDIR)
- `TemporaryDirectory()`: Secure creation, automatic cleanup
- `NamedTemporaryFile()`: For files needing a visible path
- Security: Uses `O_EXCL` flag, unpredictable names, owner-only permissions

### 6. Security Considerations

**DO:**
- Use `mkstemp()`, `mkdtemp()`, or Python `tempfile` module
- Create files with restrictive permissions (0600/0700)
- Clean up files even on exception (use context managers)
- Generate random, unpredictable filenames

**DON'T:**
- Use predictable names (enables DoS attacks)
- Leave temp files without cleanup handlers
- Create temp files in shared directories without proper security
- Rely solely on systemd `PrivateTmp=` for security

### 7. Git Worktree Considerations

For claudectl's workspace model (git worktrees):
- Worktrees share `.git/objects` but have separate working directories
- Temp files in project directory are worktree-specific (good for isolation)
- System temp is shared across worktrees (may cause conflicts with same filenames)
- Solution: Include worktree identifier in temp file paths

## Detailed Analysis

### The Hybrid Approach for claudectl

Given claudectl manages isolated workspaces via git worktrees, the recommended approach is:

1. **For ephemeral scratch files** (agent working memory, intermediate outputs):
   - Use Python `tempfile.TemporaryDirectory()`
   - Automatically respects `TMPDIR`
   - Automatic cleanup on completion
   - Include workspace identifier in prefix: `tempfile.mkdtemp(prefix=f"claude-{workspace_name}-")`

2. **For cached research/plans** (cross-session persistence):
   - Use project-local `.claude/` directory (already exists)
   - `.claude/research/` for research findings
   - `.claude/plans/` for implementation plans
   - Version-controlled, survives across sessions

3. **For build artifacts** (if claudectl builds anything):
   - Use `.cache/` in project root
   - Add to `.gitignore`
   - Follows npm/webpack convention

### Avoiding Temp File Conflicts Across Parallel Agents

When multiple Claude agents run in parallel:
- Each agent should get a unique temp directory
- Use `tempfile.mkdtemp()` with descriptive prefix
- Never share temp paths between agents without explicit coordination
- Consider using `XDG_RUNTIME_DIR` for session-bound IPC files

### Cleanup Strategy

1. **Short-lived files**: Context managers with `TemporaryDirectory`
2. **Session-bound files**: Register cleanup with `atexit` module
3. **Persistent caches**: Implement age-based pruning (like pytest's 3-run limit)

## Applicable Patterns for claudectl

### Pattern 1: Agent Scratch Space

```python
from tempfile import TemporaryDirectory
from pathlib import Path

def create_agent_workspace(agent_id: str, workspace_name: str) -> Path:
    """Create isolated temp directory for agent work."""
    return TemporaryDirectory(
        prefix=f"claude-{workspace_name}-{agent_id}-",
        # Respects TMPDIR automatically
    )
```

### Pattern 2: Persistent Research Cache

```python
RESEARCH_DIR = Path(".claude/research")

def cache_research(topic: str, content: str) -> Path:
    """Cache research to project-local directory."""
    RESEARCH_DIR.mkdir(parents=True, exist_ok=True)
    date = datetime.now().strftime("%Y-%m-%d")
    path = RESEARCH_DIR / f"{date}-{slugify(topic)}.md"
    path.write_text(content)
    return path
```

### Pattern 3: Worktree-Safe Temp Files

```python
import os
from pathlib import Path
from tempfile import gettempdir

def get_workspace_temp_dir(workspace_path: Path) -> Path:
    """Get temp directory that includes workspace context."""
    workspace_id = workspace_path.name
    base = Path(os.environ.get("TMPDIR", gettempdir()))
    temp_dir = base / f"claudectl-{workspace_id}"
    temp_dir.mkdir(exist_ok=True)
    return temp_dir
```

## Recommendations for claudectl

### Immediate Recommendations

1. **Document temp file guidance in CLAUDE.md**:
   - Agents should use `tempfile.TemporaryDirectory()` for scratch work
   - Research goes to `.claude/research/` (project-local)
   - Plans go to `.claude/plans/` (project-local)

2. **Avoid /tmp path hardcoding**:
   - Always use `tempfile.gettempdir()` or `TemporaryDirectory()`
   - Respect `TMPDIR` environment variable

3. **Include workspace context in temp paths**:
   - Prevents conflicts when multiple workspaces are active
   - Use workspace name as prefix: `claude-{workspace}-{purpose}-`

4. **Add cleanup hooks**:
   - Register `atexit` handlers for cleanup
   - Consider maximum age policy for cached files

### Directory Structure Recommendation

```
/tmp/claude-{workspace}-{agent-id}/     # Ephemeral agent work
    scratch/
    intermediate/

{project}/.claude/                       # Persistent, VCS-tracked
    research/                            # Cached research (dated)
    plans/                               # Implementation plans

{project}/.cache/                        # Persistent, gitignored
    claudectl/                           # Build/runtime caches
```

## Sources

- [Stack Exchange: /tmp vs current directory](https://softwareengineering.stackexchange.com/questions/314796/should-temporary-files-be-saved-to-tmp-or-the-current-working-directory)
- [systemd.io: Temporary Directories](https://systemd.io/TEMPORARY_DIRECTORIES/)
- [pytest: tmp_path documentation](https://docs.pytest.org/en/stable/how-to/tmp_path.html)
- [Python: tempfile module](https://docs.python.org/3/library/tempfile.html)
- [npm: folders documentation](https://docs.npmjs.com/cli/v7/configuring-npm/folders/)
- [Cargo issue #6227](https://github.com/rust-lang/cargo/issues/6227)
- [Arch Wiki: XDG Base Directory](https://wiki.archlinux.org/title/XDG_Base_Directory)
- [OpenStack: Using Temporary Files Securely](https://security.openstack.org/guidelines/dg_using-temporary-files-securely.html)
- [node_modules/.cache convention](https://jongleberry.medium.com/speed-up-your-ci-and-dx-with-node-modules-cache-ac8df82b7bb0)

## Confidence Level

**High** - Findings are consistent across multiple authoritative sources (Python docs, systemd docs, XDG specification, major tool implementations). The hybrid approach aligns with established conventions in the ecosystem.

## Related Questions

- Should claudectl provide a CLI flag to override temp directory location?
- How should temp files be handled when a workspace is deleted?
- Should there be automatic pruning of old research/plan files?
