"""Hook commands for agentctl.

These commands are designed to be called directly from Claude Code hooks.
They handle stdin parsing, error handling, and exit codes appropriately
for use as hook commands (no redirects or || true needed).

All commands read session_id and tool_input from stdin JSON automatically.
"""

from __future__ import annotations

import json
import shutil
import subprocess
import sys
from datetime import datetime
from pathlib import Path

import typer
from git import InvalidGitRepositoryError, Repo
from git.exc import GitCommandError

from agentctl.cli.output import CLIError

app = typer.Typer(
    name="hook",
    help="Commands designed for use as Claude Code hooks.",
    no_args_is_help=True,
)


class HookError(CLIError):
    """Error raised by hook commands.

    Hook errors are output as JSON to stderr for Claude Code to parse.
    """

    def __init__(self, message: str, *, exit_code: int = 1, **data: str | int | None) -> None:
        super().__init__(message, exit_code=exit_code, data=dict(data))

    def output(self) -> None:
        """Output the error as JSON to stderr."""
        error_dict = {"error": self.message, **self.data}
        # Filter out None values
        error_dict = {k: v for k, v in error_dict.items() if v is not None}
        print(json.dumps(error_dict), file=sys.stderr)


def _get_stdin_data() -> dict | None:
    """Read stdin JSON data from hooks."""
    if sys.stdin.isatty():
        return None

    try:
        stdin_data = sys.stdin.read()
        if stdin_data:
            return json.loads(stdin_data)
    except (json.JSONDecodeError, OSError):
        pass

    return None


def _get_session_id(hook_input: dict | None) -> str | None:
    """Extract session_id from hook input."""
    if hook_input:
        return hook_input.get("session_id")
    return None


def _get_file_path(hook_input: dict | None) -> str | None:
    """Extract file_path from hook input (tool_input.file_path)."""
    if hook_input:
        tool_input = hook_input.get("tool_input", {})
        return tool_input.get("file_path")
    return None


def _get_transcript_path(hook_input: dict | None) -> str | None:
    """Extract transcript_path from hook input."""
    if hook_input:
        return hook_input.get("transcript_path")
    return None


def _is_subagent(transcript_path: str | None) -> bool:
    """Check if this is a subagent based on transcript path.

    Subagent transcripts are named 'agent-{id}.jsonl'.
    """
    if transcript_path:
        filename = Path(transcript_path).stem
        return filename.startswith("agent-")
    return False


def _normalize_path(path: str) -> str:
    """Normalize a file path for consistent locking."""
    return str(Path(path).resolve())


def _get_repo(file_path: str) -> Repo | None:
    """Get the git repo containing the file, or None if not in a repo."""
    try:
        return Repo(Path(file_path).parent, search_parent_directories=True)
    except InvalidGitRepositoryError:
        return None


def _is_main_branch(repo: Repo) -> bool:
    """Check if repo is on main/master branch."""
    try:
        branch = repo.active_branch.name
        return branch in ("main", "master")
    except (TypeError, AttributeError):
        # Detached HEAD or other edge case
        return False


def _git_add_and_commit(repo: Repo, file_path: str) -> bool:
    """Stage and commit a file with auto-generated message.

    Returns True if commit was made, False otherwise.
    """
    try:
        # Make path relative to repo root
        repo_root = Path(repo.working_dir)
        abs_path = Path(file_path).resolve()

        try:
            rel_path = abs_path.relative_to(repo_root)
        except ValueError:
            # File not in repo
            return False

        # Stage the file
        repo.index.add([str(rel_path)])

        # Check if there are staged changes for this file
        diff = repo.index.diff("HEAD", paths=[str(rel_path)])
        if not diff:
            # No changes to commit
            return False

        # Calculate lines changed
        changed_lines = 0
        for d in diff:
            if d.a_blob and d.b_blob:
                try:
                    a_lines = d.a_blob.data_stream.read().decode("utf-8", errors="replace").splitlines()
                    b_lines = d.b_blob.data_stream.read().decode("utf-8", errors="replace").splitlines()
                    changed_lines = abs(len(b_lines) - len(a_lines)) + sum(
                        1 for a, b in zip(a_lines, b_lines, strict=False) if a != b
                    )
                except Exception:
                    changed_lines = 10  # Default to moderate
            else:
                changed_lines = 10

        # Determine change size
        filename = abs_path.name
        if changed_lines < 10:
            size = "minor"
        elif changed_lines < 50:
            size = "moderate"
        else:
            size = "major"

        # Create commit
        msg = f"Update {filename}: {size} changes ({changed_lines} lines)"
        repo.index.commit(msg)

        return True

    except GitCommandError:
        return False
    except Exception:
        return False


def _git_add_and_commit_new_file(repo: Repo, file_path: str) -> bool:
    """Stage and commit a new file.

    Returns True if commit was made, False otherwise.
    """
    try:
        # Make path relative to repo root
        repo_root = Path(repo.working_dir)
        abs_path = Path(file_path).resolve()

        try:
            rel_path = abs_path.relative_to(repo_root)
        except ValueError:
            # File not in repo
            return False

        # Stage the file
        repo.index.add([str(rel_path)])

        # Check if file is staged
        if not repo.index.diff("HEAD", paths=[str(rel_path)]):
            # Check if it's a new untracked file that's now staged
            staged = [item.a_path for item in repo.index.diff("HEAD")]
            if str(rel_path) not in staged:
                return False

        filename = abs_path.name
        msg = f"Add new file: {filename}"
        repo.index.commit(msg)

        return True

    except GitCommandError:
        return False
    except Exception:
        return False


@app.command("post-edit")
def post_edit() -> None:
    """PostToolUse hook for Edit tool.

    Auto-commits changes if on a feature branch.
    Reads file path and session ID from stdin JSON.

    Exit codes:
        0: Success
    """
    hook_input = _get_stdin_data()
    file_path = _get_file_path(hook_input)

    # Auto-commit logic (only on feature branches)
    if file_path:
        repo = _get_repo(file_path)
        if repo and not _is_main_branch(repo):
            _git_add_and_commit(repo, file_path)

    raise typer.Exit(0)


@app.command("post-write")
def post_write() -> None:
    """PostToolUse hook for Write tool (new files).

    Auto-commits new files if on a feature branch.
    Reads file path and session ID from stdin JSON.

    Exit codes:
        0: Success
    """
    hook_input = _get_stdin_data()
    file_path = _get_file_path(hook_input)

    # Auto-commit logic (only on feature branches)
    if file_path:
        repo = _get_repo(file_path)
        if repo and not _is_main_branch(repo):
            _git_add_and_commit_new_file(repo, file_path)

    raise typer.Exit(0)


# =============================================================================
# Notification Commands
# =============================================================================

APP_NAME = "Claude Code"
CLAUDE_SENDER = "com.anthropic.claudefordesktop"


def _get_project_name() -> str:
    """Get project name from current directory."""
    return Path.cwd().name


def _get_time() -> str:
    """Get current time in readable format."""
    return datetime.now().strftime("%-I:%M %p")


def _has_terminal_notifier() -> bool:
    """Check if terminal-notifier is available."""
    return shutil.which("terminal-notifier") is not None


def _extract_final_response(transcript_path: str, max_length: int = 200) -> str | None:
    """Extract the final assistant response from a transcript JSONL file.

    Returns the text content truncated to max_length, or None if not found.
    """
    try:
        path = Path(transcript_path).expanduser()
        if not path.exists():
            return None

        # Read the transcript and find the last assistant message
        last_response = None
        with open(path) as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                try:
                    entry = json.loads(line)
                    # Look for assistant messages with text content
                    if entry.get("type") == "assistant":
                        message = entry.get("message", {})
                        content = message.get("content", [])
                        # Extract text from content blocks
                        for block in content:
                            if isinstance(block, dict) and block.get("type") == "text":
                                last_response = block.get("text", "")
                            elif isinstance(block, str):
                                last_response = block
                except json.JSONDecodeError:
                    continue

        if last_response:
            # Truncate and clean up for notification
            text = last_response.strip()
            # Take first line or truncate
            first_line = text.split("\n")[0]
            # Strip markdown formatting (bold, italic, code)
            import re

            first_line = re.sub(r"\*\*(.+?)\*\*", r"\1", first_line)  # **bold**
            first_line = re.sub(r"\*(.+?)\*", r"\1", first_line)  # *italic*
            first_line = re.sub(r"`(.+?)`", r"\1", first_line)  # `code`
            first_line = re.sub(r"^#+\s*", "", first_line)  # # headers
            if len(first_line) > max_length:
                return first_line[: max_length - 3] + "..."
            return first_line

        return None
    except Exception:
        return None


def _send_notification(
    title: str,
    subtitle: str,
    message: str,
    sound: str | None = None,
    group: str | None = None,
) -> bool:
    """Send a macOS notification using terminal-notifier or osascript fallback."""
    if _has_terminal_notifier():
        cmd = [
            "terminal-notifier",
            "-title",
            title,
            "-subtitle",
            subtitle,
            "-message",
            message,
            "-sender",
            CLAUDE_SENDER,
        ]
        if sound:
            cmd.extend(["-sound", sound])
        if group:
            cmd.extend(["-group", group])

        try:
            subprocess.run(cmd, capture_output=True, check=False)
            return True
        except (subprocess.SubprocessError, FileNotFoundError):
            return False
    else:
        # Fallback to osascript
        sound_clause = f' sound name "{sound}"' if sound else ""
        script = f'display notification "{message}" with title "{title}" subtitle "{subtitle}"{sound_clause}'
        try:
            subprocess.run(["osascript", "-e", script], capture_output=True, check=False)
            return True
        except (subprocess.SubprocessError, FileNotFoundError):
            return False


@app.command("notify-input")
def notify_input() -> None:
    """Notification hook - sends notification when Claude needs input.

    Reads message and notification_type from stdin JSON to provide context.

    Exit codes:
        0: Always succeeds (notification is best-effort)
    """
    hook_input = _get_stdin_data()
    project_name = _get_project_name()

    # Extract message from hook input for richer notifications
    message = "Claude needs your input to continue"
    if hook_input:
        # Use the message field if available
        if hook_input.get("message"):
            message = hook_input["message"]
        # Add notification type context
        notification_type = hook_input.get("notification_type", "")
        if notification_type == "permission_prompt":
            message = hook_input.get("message", "Permission required")

    _send_notification(
        title=f"🔔 {APP_NAME}",
        subtitle=project_name,
        message=message,
        group=f"claude-code-{project_name}",
    )
    raise typer.Exit(0)


@app.command("notify-stop")
def notify_stop() -> None:
    """Stop hook - sends notification when Claude completes a task.

    Extracts Claude's final response from the transcript for the notification body.

    Exit codes:
        0: Always succeeds (notification is best-effort)
    """
    hook_input = _get_stdin_data()
    project_name = _get_project_name()
    time = _get_time()

    # Try to extract the final response from the transcript
    message = f"Completed at {time}"
    if hook_input and hook_input.get("transcript_path"):
        final_response = _extract_final_response(hook_input["transcript_path"])
        if final_response:
            message = final_response

    _send_notification(
        title=f"✅ {APP_NAME}",
        subtitle=project_name,
        message=message,
        group=f"claude-code-{project_name}",
    )
    raise typer.Exit(0)


@app.command("notify-error")
def notify_error() -> None:
    """Send error notification.

    Exit codes:
        0: Always succeeds (notification is best-effort)
    """
    hook_input = _get_stdin_data()
    project_name = _get_project_name()

    message = "An error occurred during task execution"
    if hook_input and hook_input.get("message"):
        message = hook_input["message"]

    _send_notification(
        title=f"❌ {APP_NAME}",
        subtitle=project_name,
        message=message,
        sound="Basso",
        group=f"claude-code-{project_name}",
    )
    raise typer.Exit(0)


@app.command("notify-test")
def notify_test() -> None:
    """Send a test notification to verify the system is working.

    Exit codes:
        0: Always succeeds
    """
    project_name = _get_project_name()
    has_notifier = _has_terminal_notifier()

    _send_notification(
        title=f"🧪 {APP_NAME}",
        subtitle=project_name,
        message="Notifications are working!",
        group=f"claude-code-{project_name}",
    )

    if has_notifier:
        typer.echo("✓ Test notification sent (using terminal-notifier)")
    else:
        typer.echo("✓ Test notification sent (using osascript fallback)")
        typer.echo("\n  Tip: Install terminal-notifier for more reliable notifications:")
        typer.echo("       brew install terminal-notifier")

    raise typer.Exit(0)


# =============================================================================
# Context Info Hook - Injects live context into user prompts
# =============================================================================


def _get_git_branch() -> str | None:
    """Get the current git branch name."""
    try:
        repo = Repo(Path.cwd(), search_parent_directories=True)
        return repo.active_branch.name
    except (InvalidGitRepositoryError, TypeError):
        return None


def _get_git_status_summary() -> str | None:
    """Get a brief git status summary (dirty/clean, staged count)."""
    try:
        repo = Repo(Path.cwd(), search_parent_directories=True)
        # Check if dirty
        is_dirty = repo.is_dirty(untracked_files=True)
        if not is_dirty:
            return "clean"

        # Count changes
        staged = len(repo.index.diff("HEAD"))
        unstaged = len(repo.index.diff(None))
        untracked = len(repo.untracked_files)

        parts = []
        if staged:
            parts.append(f"{staged} staged")
        if unstaged:
            parts.append(f"{unstaged} modified")
        if untracked:
            parts.append(f"{untracked} untracked")

        return ", ".join(parts) if parts else "dirty"
    except (InvalidGitRepositoryError, GitCommandError):
        return None


def _get_pr_status() -> dict | None:
    """Get open PR status for current branch using gh CLI."""
    try:
        result = subprocess.run(
            ["gh", "pr", "view", "--json", "number,title,state,url,reviewDecision,statusCheckRollup"],
            capture_output=True,
            text=True,
            timeout=5,
            check=False,
        )
        if result.returncode != 0:
            return None

        pr_data = json.loads(result.stdout)
        if pr_data.get("state") != "OPEN":
            return None

        # Summarize check status
        checks = pr_data.get("statusCheckRollup", [])
        check_summary = None
        if checks:
            passed = sum(1 for c in checks if c.get("conclusion") == "SUCCESS")
            failed = sum(1 for c in checks if c.get("conclusion") == "FAILURE")
            pending = sum(1 for c in checks if c.get("status") == "IN_PROGRESS" or c.get("conclusion") is None)
            if failed:
                check_summary = f"{failed} failing"
            elif pending:
                check_summary = f"{pending} pending"
            elif passed:
                check_summary = f"{passed} passed"

        return {
            "number": pr_data.get("number"),
            "title": pr_data.get("title"),
            "review": pr_data.get("reviewDecision"),
            "checks": check_summary,
            "url": pr_data.get("url"),
        }
    except (subprocess.SubprocessError, json.JSONDecodeError, subprocess.TimeoutExpired):
        return None


def _get_directory_snapshot(max_files: int = 20) -> list[str]:
    """Get a snapshot of important files in the current directory."""
    cwd = Path.cwd()
    files = []

    # Priority patterns for important files
    priority_patterns = [
        "*.py",
        "*.ts",
        "*.tsx",
        "*.js",
        "*.jsx",
        "*.go",
        "*.rs",
        "*.rb",
        "*.java",
        "Makefile",
        "justfile",
        "package.json",
        "pyproject.toml",
        "Cargo.toml",
        "go.mod",
        "Gemfile",
    ]

    # Collect files by priority
    seen = set()
    for pattern in priority_patterns:
        for path in cwd.glob(pattern):
            if path.is_file() and path.name not in seen:
                seen.add(path.name)
                files.append(path.name)
                if len(files) >= max_files:
                    break
        if len(files) >= max_files:
            break

    # Add some subdirectories if space remains
    if len(files) < max_files:
        for path in sorted(cwd.iterdir()):
            if path.is_dir() and not path.name.startswith("."):
                dir_name = f"{path.name}/"
                if dir_name not in seen:
                    seen.add(dir_name)
                    files.append(dir_name)
                    if len(files) >= max_files:
                        break

    return sorted(files)


def _get_all_git_branches() -> dict | None:
    """Get all local and remote git branches with cleanliness status."""
    try:
        repo = Repo(Path.cwd(), search_parent_directories=True)
        branches_info = {}

        # Get all local branches
        for ref in repo.heads:
            branch_name = ref.name
            # Check if this branch is dirty
            try:
                # Stash current changes, checkout branch, check status, restore
                original_branch = repo.active_branch.name if repo.active_branch else None
                repo.heads[branch_name].checkout()
                is_dirty = repo.is_dirty(untracked_files=True)
                if original_branch:
                    repo.heads[original_branch].checkout()
                branches_info[branch_name] = "dirty" if is_dirty else "clean"
            except (GitCommandError, TypeError):
                # Can't switch or check - mark as unknown
                branches_info[branch_name] = "unknown"

        return branches_info if branches_info else None
    except (InvalidGitRepositoryError, GitCommandError):
        return None


def _get_workspace_sessions() -> list[dict] | None:
    """Get workspace list using agentctl workspace list --json."""
    try:
        result = subprocess.run(
            ["agentctl", "workspace", "list", "--json"],
            capture_output=True,
            text=True,
            timeout=5,
            check=False,
        )
        if result.returncode != 0:
            return None

        workspaces = json.loads(result.stdout)
        if not workspaces:
            return None

        # Format: list of {branch, path, status, ...}
        return workspaces
    except (subprocess.SubprocessError, json.JSONDecodeError, subprocess.TimeoutExpired):
        return None


def _get_current_workspace(cwd: Path) -> dict | None:
    """Get the current workspace if cwd is inside an agentctl workspace."""
    workspaces = _get_workspace_sessions()
    if not workspaces:
        return None

    cwd_str = str(cwd.resolve())
    for ws in workspaces:
        ws_path = ws.get("path", "")
        if ws_path and cwd_str.startswith(ws_path):
            return ws
    return None


@app.command("context-info")
def context_info() -> None:
    """UserPromptSubmit hook - injects live context into each user prompt.

    Outputs context information that gets automatically injected into
    the conversation before Claude processes the user's message.

    Includes:
    - Current git branch and status (clean/dirty, staged/modified/untracked)
    - All git branches and their cleanliness status
    - All workspace sessions with status
    - Open PR status (number, title, review decision, check status)
    - Directory snapshot (important files and subdirectories)

    Exit codes:
        0: Success (context is output to stdout)
    """
    _get_stdin_data()  # Consume stdin if present
    cwd = Path.cwd().resolve()

    lines = []
    lines.append("<context-refresh>")

    # Absolute path
    lines.append(f"Path: {cwd}")

    # Current workspace (if in one)
    current_ws = _get_current_workspace(cwd)
    if current_ws:
        ws_line = f"Current Workspace: {current_ws.get('branch', 'unknown')}"
        if current_ws.get("status"):
            ws_line += f" ({current_ws['status']})"
        lines.append(ws_line)

    # Git info (current branch)
    branch = _get_git_branch()
    if branch:
        status = _get_git_status_summary()
        git_line = f"Branch: {branch}"
        if status:
            git_line += f" ({status})"
        lines.append(git_line)

    # All git branches with cleanliness
    all_branches = _get_all_git_branches()
    if all_branches:
        lines.append("Git Branches:")
        for branch_name, branch_status in sorted(all_branches.items()):
            lines.append(f"  {branch_name}: {branch_status}")

    # Workspace sessions
    workspaces = _get_workspace_sessions()
    if workspaces:
        lines.append("Workspaces:")
        for ws in workspaces:
            ws_line = f"  {ws.get('branch', 'unknown')}"
            if ws.get("status"):
                ws_line += f" ({ws['status']})"
            lines.append(ws_line)

    # PR status
    pr = _get_pr_status()
    if pr:
        pr_line = f"PR #{pr['number']}: {pr['title']}"
        details = []
        if pr.get("review"):
            details.append(pr["review"].lower().replace("_", " "))
        if pr.get("checks"):
            details.append(f"checks: {pr['checks']}")
        if details:
            pr_line += f" ({', '.join(details)})"
        lines.append(pr_line)

    # Directory snapshot
    files = _get_directory_snapshot(max_files=15)
    if files:
        lines.append(f"Directory: {cwd.name}/")
        lines.append(f"  {', '.join(files[:10])}")
        if len(files) > 10:
            lines.append(f"  ... and {len(files) - 10} more")

    lines.append("</context-refresh>")

    # Output as plain text - this gets injected as additionalContext
    print("\n".join(lines))
    raise typer.Exit(0)
