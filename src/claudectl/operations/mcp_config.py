"""MCP configuration utilities."""

from __future__ import annotations

from pathlib import Path


def generate_mcp_config() -> Path | None:
    """Generate or locate MCP configuration for Claude.

    Returns:
        Path to MCP config file if available, None otherwise.
    """
    # Check for user's Claude config file
    claude_config = Path.home() / ".claude.json"
    if claude_config.exists():
        return claude_config

    return None
