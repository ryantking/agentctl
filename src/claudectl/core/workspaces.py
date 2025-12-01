"""Workspace domain model and discovery."""

from __future__ import annotations

import re
import subprocess
from dataclasses import dataclass
from pathlib import Path

from claudectl.core.git import get_repo_name, get_repo_root, is_worktree_clean


@dataclass
class Workspace:
    """A git worktree managed by claudectl.

    Represents a single workspace (git worktree) that can be managed
    by claudectl. Workspaces are discovered from git's authoritative
    worktree list.
    """

    path: Path
    branch: str | None
    commit: str
    is_main: bool
    repo_root: Path

    @property
    def is_managed(self) -> bool:
        """Check if this workspace is managed by claudectl.

        Managed workspaces live under ~/.claude/workspaces/<repo>/
        """
        return ".claude/workspaces" in str(self.path)

    @property
    def is_clean(self) -> tuple[bool, str]:
        """Check if workspace has uncommitted changes.

        Returns:
            Tuple of (is_clean, status_message)
        """
        return is_worktree_clean(self.path)

    def to_dict(self) -> dict:
        """Convert to dictionary for JSON output."""
        is_clean, status = self.is_clean
        return {
            "path": str(self.path),
            "branch": self.branch,
            "commit": self.commit,
            "is_main": self.is_main,
            "is_managed": self.is_managed,
            "is_clean": is_clean,
            "status": status,
        }


def parse_worktree_list(output: str, repo_root: Path) -> list[Workspace]:
    """Parse git worktree list --porcelain output.

    Args:
        output: Output from git worktree list --porcelain
        repo_root: Root of the git repository

    Returns:
        List of Workspace objects
    """
    worktrees = []
    current: dict = {}

    for line in output.splitlines():
        if line.startswith("worktree "):
            current["path"] = Path(line.split(" ", 1)[1])
        elif line.startswith("HEAD "):
            current["commit"] = line.split(" ", 1)[1][:8]
        elif line.startswith("branch "):
            ref = line.split(" ", 1)[1]
            if ref.startswith("refs/heads/"):
                current["branch"] = ref[len("refs/heads/") :]
            else:
                current["branch"] = ref
        elif line == "" and current:
            worktrees.append(
                Workspace(
                    path=current.get("path", Path()),
                    branch=current.get("branch"),
                    commit=current.get("commit", ""),
                    is_main=len(worktrees) == 0,  # First entry is main
                    repo_root=repo_root,
                )
            )
            current = {}

    # Handle last entry if no trailing newline
    if current:
        worktrees.append(
            Workspace(
                path=current.get("path", Path()),
                branch=current.get("branch"),
                commit=current.get("commit", ""),
                is_main=len(worktrees) == 0,
                repo_root=repo_root,
            )
        )

    return worktrees


def discover_workspaces(repo_root: Path | None = None) -> list[Workspace]:
    """Discover all workspaces using git worktree list.

    Args:
        repo_root: Optional repository root. If None, uses current repo.

    Returns:
        List of Workspace objects for all worktrees
    """
    if repo_root is None:
        repo_root = get_repo_root()

    try:
        result = subprocess.run(
            ["git", "-C", str(repo_root), "worktree", "list", "--porcelain"],
            capture_output=True,
            text=True,
            check=True,
        )
        return parse_worktree_list(result.stdout, repo_root)
    except subprocess.CalledProcessError:
        return []


def find_workspace_by_branch(branch: str, repo_root: Path | None = None) -> Workspace | None:
    """Find a workspace by its branch name.

    Args:
        branch: Branch name to search for
        repo_root: Optional repository root

    Returns:
        Workspace if found, None otherwise
    """
    workspaces = discover_workspaces(repo_root)
    for workspace in workspaces:
        if workspace.branch == branch:
            return workspace
    return None


def sanitize_workspace_name(branch_name: str) -> str:
    """Generate a safe workspace directory name from branch name.

    Examples:
        feature/auth-api -> feature-auth-api
        feat/BS-123-new -> feat-BS-123-new
    """
    # Replace slashes and special chars with hyphens
    name = re.sub(r"[/\\]+", "-", branch_name)
    # Replace spaces and underscores with hyphens
    name = re.sub(r"[_\s]+", "-", name)
    # Remove any other unsafe characters
    name = re.sub(r"[^a-zA-Z0-9\-.]", "", name)
    # Remove leading/trailing hyphens
    name = name.strip("-")
    return name


def get_workspaces_base_path() -> Path:
    """Get the base path for workspaces: ~/.claude/workspaces/<repo-name>."""
    return Path.home() / ".claude" / "workspaces" / get_repo_name()


def get_workspace_path(branch_name: str, repo_root: Path | None = None) -> Path:
    """Calculate expected workspace path for a branch.

    Args:
        branch_name: Branch name
        repo_root: Optional repository root

    Returns:
        Path where workspace should be located
    """
    if repo_root is None:
        repo_root = get_repo_root()

    workspace_name = sanitize_workspace_name(branch_name)
    return Path.home() / ".claude" / "workspaces" / repo_root.name / workspace_name
