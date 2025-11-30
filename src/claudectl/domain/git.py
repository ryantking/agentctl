"""Low-level git operations for workspace management."""

from __future__ import annotations

from pathlib import Path

from git import InvalidGitRepositoryError, Repo
from git.exc import GitCommandError

from claudectl.domain.exceptions import NotInGitRepoError


def get_repo_root() -> Path:
    """Get the root directory of the current git repository.

    Correctly handles worktrees by finding the actual repository root
    instead of the worktree directory.

    Returns:
        Path to repository root

    Raises:
        NotInGitRepoError: If not in a git repository
    """
    try:
        repo = Repo(search_parent_directories=True)

        # For worktrees, repo.common_dir returns the shared .git directory
        # (which may contain relative path components like ..).
        # Resolve first to normalize, then get parent to get repo root.
        common_dir = Path(repo.common_dir).resolve()
        return common_dir.parent
    except InvalidGitRepositoryError as exc:
        raise NotInGitRepoError("Not in a git repository") from exc


def get_repo_name() -> str:
    """Get the name of the current git repository.

    Returns:
        Repository directory name
    """
    return get_repo_root().name


def get_current_branch() -> str | None:
    """Get the name of the current branch.

    Returns:
        Branch name, or None if in detached HEAD state
    """
    try:
        repo = Repo(search_parent_directories=True)
        if repo.head.is_detached:
            return None
        return repo.active_branch.name
    except (InvalidGitRepositoryError, TypeError):
        # TypeError can occur if head is detached
        return None


def branch_exists(branch_name: str) -> bool:
    """Check if a branch exists locally or remotely.

    Args:
        branch_name: Name of the branch to check

    Returns:
        True if branch exists, False otherwise

    Raises:
        NotInGitRepoError: If not in a git repository
    """
    try:
        repo = Repo(search_parent_directories=True)

        # Check local branches
        if branch_name in [head.name for head in repo.heads]:
            return True

        # Check remote branches (if origin exists)
        try:
            origin = repo.remotes.origin
            if branch_name in [ref.remote_head for ref in origin.refs]:
                return True
        except (AttributeError, IndexError):
            # No origin remote configured
            pass

        return False
    except InvalidGitRepositoryError as exc:
        raise NotInGitRepoError("Not in a git repository") from exc


def is_worktree_clean(worktree_path: Path) -> tuple[bool, str]:
    """Check if a worktree has uncommitted changes.

    Args:
        worktree_path: Path to the worktree to check

    Returns:
        Tuple of (is_clean, status_message)
    """
    try:
        repo = Repo(worktree_path)

        # Check for staged and unstaged changes
        if repo.is_dirty(untracked_files=False):
            # Count modified files
            modified_files = len(repo.index.diff(None)) + len(repo.index.diff("HEAD"))
            untracked_files = len(repo.untracked_files)

            parts = []
            if modified_files:
                parts.append(f"{modified_files} modified")
            if untracked_files:
                parts.append(f"{untracked_files} untracked")

            return False, ", ".join(parts)

        # Check for untracked files only
        if repo.untracked_files:
            return False, f"{len(repo.untracked_files)} untracked"

        return True, "Clean"

    except (InvalidGitRepositoryError, GitCommandError) as e:
        return False, f"Failed to check status: {e}"
