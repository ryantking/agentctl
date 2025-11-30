"""Context and environment setup operations for workspaces."""

from __future__ import annotations

import shutil
from pathlib import Path

from claudectl.domain.git import get_repo_root


def copy_claude_context(workspace_path: Path, source_root: Path | None = None) -> list[str]:
    """Copy Claude local settings and CLAUDE.md to a workspace.

    Copies files that aren't tracked by git but are needed for Claude
    to have proper context and permissions in the new workspace.

    Files copied (if they exist in source but not in workspace):
    - .claude/settings.local.json
    - CLAUDE.md

    Args:
        workspace_path: Path to the workspace to copy files to
        source_root: Source repository root. If None, uses current repo.

    Returns:
        List of files that were copied
    """
    if source_root is None:
        source_root = get_repo_root()

    copied: list[str] = []

    # Files to copy (relative to repo root)
    files_to_copy = [
        Path(".claude") / "settings.local.json",
        Path("CLAUDE.md"),
    ]

    for rel_path in files_to_copy:
        source_file = source_root / rel_path
        dest_file = workspace_path / rel_path

        # Only copy if source exists and dest doesn't
        if source_file.exists() and not dest_file.exists():
            # Create parent directory if needed
            dest_file.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(source_file, dest_file)
            copied.append(str(rel_path))

    return copied
