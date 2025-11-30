"""Exception hierarchy for workspace operations."""

from __future__ import annotations


class WorkspaceError(Exception):
    """Base exception for workspace-related errors."""

    pass


class NotInGitRepoError(WorkspaceError):
    """Raised when not in a git repository."""

    pass


class WorkspaceExistsError(WorkspaceError):
    """Raised when workspace already exists."""

    pass


class BranchInUseError(WorkspaceError):
    """Raised when branch is already checked out in another workspace."""

    pass


class WorkspaceNotFoundError(WorkspaceError):
    """Raised when workspace cannot be found."""

    pass
