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


class InitError(WorkspaceError):
    """Base exception for init operations."""

    pass


class ImportDirNotFoundError(InitError):
    """Raised when import directory cannot be found."""

    def __init__(self) -> None:
        super().__init__(
            "Import directory not found. "
            "This may indicate a corrupted installation."
        )
