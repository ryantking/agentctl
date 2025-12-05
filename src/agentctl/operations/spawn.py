"""Process spawning utilities for agentctl."""

from __future__ import annotations

import os
import subprocess
from pathlib import Path

from agentctl.operations.mcp_config import generate_mcp_config

# Type alias for process result (Python 3.12+ style)
type SpawnedProcess = subprocess.Popen[bytes]


def spawn_claude_in_shell(working_dir: str | Path | None = None) -> SpawnedProcess:
    """Spawn Claude with full environment inheritance.

    Spawns Claude directly via subprocess.Popen with the complete parent
    environment. This ensures Claude has access to:
    - HOME (to locate ~/.claude.json with MCP configurations)
    - PATH (to execute MCP server commands)
    - All other environment variables from the parent process

    Args:
        working_dir: Directory to spawn Claude in. If None, uses current directory.

    Returns:
        The spawned subprocess.Popen object.

    Raises:
        OSError: If spawning fails (e.g., claude command not found).
        ValueError: If working_dir doesn't exist or is not a directory.

    Example:
        >>> process = spawn_claude_in_shell("/path/to/workspace")
        >>> process.wait()  # Wait for Claude to exit
    """
    # Validate and resolve working directory
    if working_dir is not None:
        work_path = Path(working_dir).resolve()
        if not work_path.exists():
            raise ValueError(f"Working directory does not exist: {working_dir}")
        if not work_path.is_dir():
            raise ValueError(f"Working directory is not a directory: {working_dir}")
        cwd = work_path
    else:
        cwd = None  # Use current working directory

    # Generate MCP configuration
    mcp_config_path = generate_mcp_config()

    # Build claude command arguments
    cmd = ["claude"]
    if mcp_config_path:
        cmd.extend(["--mcp-config", str(mcp_config_path)])

    # Spawn Claude directly with inherited environment
    # stdin/stdout/stderr remain connected to terminal (None means inherit)
    try:
        process = subprocess.Popen(
            cmd,
            cwd=cwd,
            env=os.environ.copy(),  # CRITICAL: Pass full parent environment
            stdin=None,  # Inherit from parent (terminal)
            stdout=None,  # Inherit from parent (terminal)
            stderr=None,  # Inherit from parent (terminal)
        )
        return process
    except FileNotFoundError as e:
        raise OSError(
            "Failed to spawn Claude: command not found. Ensure 'claude' is installed and in your PATH."
        ) from e
    except OSError as e:
        raise OSError(f"Failed to spawn Claude: {e}") from e
