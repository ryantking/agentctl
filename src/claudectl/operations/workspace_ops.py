"""High-level workspace operations."""

from __future__ import annotations

import subprocess
from pathlib import Path

from claudectl.domain.exceptions import (
    BranchInUseError,
    NotInGitRepoError,
    WorkspaceError,
    WorkspaceExistsError,
    WorkspaceNotFoundError,
)
from claudectl.domain.git import branch_exists, get_current_branch, get_repo_root
from claudectl.domain.workspace import (
    Workspace,
    discover_workspaces,
    find_workspace_by_branch,
    get_workspace_path,
)


class WorkspaceManager:
    """Manages workspace lifecycle operations."""

    def __init__(self, repo_root: Path | None = None):
        """Initialize workspace manager.

        Args:
            repo_root: Optional repository root. If None, uses current repo.

        Raises:
            NotInGitRepoError: If not in a git repository
        """
        try:
            self.repo_root = repo_root if repo_root else get_repo_root()
        except Exception as e:
            raise NotInGitRepoError("Not in a git repository") from e

    def list_workspaces(self, managed_only: bool = True) -> list[Workspace]:
        """List all workspaces.

        Args:
            managed_only: If True, only return managed workspaces

        Returns:
            List of Workspace objects
        """
        workspaces = discover_workspaces(self.repo_root)
        if managed_only:
            return [w for w in workspaces if w.is_managed and not w.is_main]
        return workspaces

    def get_workspace(self, branch: str) -> Workspace:
        """Find workspace by branch name.

        Args:
            branch: Branch name to search for

        Returns:
            Workspace object

        Raises:
            WorkspaceNotFoundError: If workspace not found
        """
        workspace = find_workspace_by_branch(branch, self.repo_root)
        if not workspace:
            raise WorkspaceNotFoundError(f"No workspace found for branch: {branch}")
        return workspace

    def create_workspace(
        self,
        branch: str,
        base_branch: str | None = None,
    ) -> Workspace:
        """Create a new workspace with worktree.

        Args:
            branch: Name of the branch (will be created if doesn't exist)
            base_branch: Base branch to create from (defaults to current branch)

        Returns:
            Created Workspace object

        Raises:
            WorkspaceExistsError: If workspace already exists
            BranchInUseError: If branch is checked out elsewhere
            WorkspaceError: For other git errors
        """
        workspace_path = get_workspace_path(branch, self.repo_root)

        # Check if workspace directory already exists
        if workspace_path.exists():
            raise WorkspaceExistsError(f"Workspace already exists: {workspace_path}")

        # Check if branch is already checked out in another worktree
        existing = find_workspace_by_branch(branch, self.repo_root)
        if existing:
            raise BranchInUseError(
                f"Branch '{branch}' is already checked out at: {existing.path}"
            )

        # Create parent directory
        workspace_path.parent.mkdir(parents=True, exist_ok=True)

        try:
            if branch_exists(branch):
                # Branch exists, just create worktree
                subprocess.run(
                    [
                        "git",
                        "-C",
                        str(self.repo_root),
                        "worktree",
                        "add",
                        str(workspace_path),
                        branch,
                    ],
                    check=True,
                    capture_output=True,
                    text=True,
                )
            else:
                # Create new branch from base
                if not base_branch:
                    base_branch = get_current_branch() or "HEAD"

                subprocess.run(
                    [
                        "git",
                        "-C",
                        str(self.repo_root),
                        "worktree",
                        "add",
                        "-b",
                        branch,
                        str(workspace_path),
                        base_branch,
                    ],
                    check=True,
                    capture_output=True,
                    text=True,
                )
        except subprocess.CalledProcessError as e:
            raise WorkspaceError(f"Failed to create worktree: {e.stderr}") from e

        # Return the newly created workspace
        workspace = find_workspace_by_branch(branch, self.repo_root)
        if not workspace:
            raise WorkspaceError("Workspace created but could not be found")
        return workspace

    def delete_workspace(self, branch: str, force: bool = False) -> bool:
        """Remove a workspace.

        Args:
            branch: Branch name of workspace to remove
            force: Force removal even if there are uncommitted changes

        Returns:
            True if removed, False if it didn't exist

        Raises:
            WorkspaceNotFoundError: If workspace not found
            WorkspaceError: If removal fails (e.g., uncommitted changes without force)
        """
        workspace = self.get_workspace(branch)

        # Check if clean
        if not force:
            is_clean, status = workspace.is_clean
            if not is_clean:
                raise WorkspaceError(
                    f"Workspace has uncommitted changes ({status}). Use --force to remove anyway."
                )

        cmd = [
            "git",
            "-C",
            str(self.repo_root),
            "worktree",
            "remove",
            str(workspace.path),
        ]
        if force:
            cmd.append("--force")

        try:
            subprocess.run(cmd, check=True, capture_output=True, text=True)
        except subprocess.CalledProcessError as e:
            raise WorkspaceError(f"Failed to remove worktree: {e.stderr}") from e

        # Clean up empty parent directories
        try:
            parent = workspace.path.parent
            if parent.exists() and not any(parent.iterdir()):
                parent.rmdir()
        except OSError:
            # Directory not empty or permission issues - safe to ignore
            pass

        return True

    def clean_workspaces(self, check_merged: bool = True) -> list[str]:
        """Remove clean/merged workspaces.

        Args:
            check_merged: If True, only remove workspaces with no uncommitted changes

        Returns:
            List of branch names that were removed
        """
        removed = []
        workspaces = self.list_workspaces(managed_only=True)

        for workspace in workspaces:
            if workspace.is_main:
                continue

            is_clean, _ = workspace.is_clean
            if not check_merged or is_clean:
                try:
                    if workspace.branch:
                        self.delete_workspace(workspace.branch, force=not check_merged)
                        removed.append(workspace.branch)
                except (WorkspaceError, WorkspaceNotFoundError):
                    # Skip workspaces that can't be deleted
                    pass

        return removed

    def get_workspace_status(self, workspace: Workspace) -> dict:
        """Get detailed workspace status.

        Args:
            workspace: Workspace to get status for

        Returns:
            Dictionary with status information
        """
        is_clean, status = workspace.is_clean

        # Get ahead/behind information
        ahead_behind = None
        if workspace.branch:
            try:
                result = subprocess.run(
                    [
                        "git",
                        "-C",
                        str(workspace.path),
                        "rev-list",
                        "--left-right",
                        "--count",
                        f"origin/{workspace.branch}...HEAD",
                    ],
                    capture_output=True,
                    text=True,
                    check=False,
                )
                if result.returncode == 0:
                    parts = result.stdout.strip().split()
                    if len(parts) == 2:
                        behind, ahead = int(parts[0]), int(parts[1])
                        ahead_behind = {"ahead": ahead, "behind": behind}
            except Exception:
                pass

        return {
            "path": str(workspace.path),
            "branch": workspace.branch,
            "commit": workspace.commit,
            "is_clean": is_clean,
            "status": status,
            "ahead_behind": ahead_behind,
        }

    def get_workspace_diff(
        self, workspace: Workspace, target_branch: str = "main"
    ) -> str:
        """Get git diff from workspace to target branch.

        Args:
            workspace: Workspace to diff
            target_branch: Target branch to compare against

        Returns:
            Git diff output
        """
        try:
            result = subprocess.run(
                [
                    "git",
                    "-C",
                    str(workspace.path),
                    "diff",
                    f"{target_branch}...HEAD",
                ],
                capture_output=True,
                text=True,
                check=True,
            )
            return result.stdout
        except subprocess.CalledProcessError as e:
            raise WorkspaceError(f"Failed to get diff: {e.stderr}") from e
