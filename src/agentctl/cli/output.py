"""Output formatting utilities for claudectl."""

from __future__ import annotations

import json
from dataclasses import dataclass, field
from typing import Any

import typer

# Global state for output format
_state: dict[str, Any] = {"json": False}


class CLIError(Exception):
    """Base exception for CLI errors with structured output support.

    Exceptions inheriting from this class will be caught by the CLI's
    top-level exception handler and formatted according to the output mode.

    Attributes:
        message: Human-readable error message
        exit_code: Exit code for the CLI (default 1)
        data: Optional dict with additional info for JSON output
    """

    def __init__(
        self,
        message: str,
        *,
        exit_code: int = 1,
        data: dict[str, Any] | None = None,
    ) -> None:
        super().__init__(message)
        self.message = message
        self.exit_code = exit_code
        self.data = data or {}

    def to_result(self) -> Result:
        """Convert to a Result for output formatting."""
        return Result(success=False, message=self.message, data=self.data)


def handle_exception(exc: Exception) -> int:
    """Handle an exception and return the appropriate exit code.

    This function formats the exception according to the current output mode
    (JSON or human-readable) and prints to stderr.

    Supports:
    - CLIError: Uses the exception's exit_code and data
    - Any exception with to_result() method: Uses that for formatting
    - Any exception with exit_code attribute: Uses that exit code
    - Generic exceptions: Wraps message with exit code 1

    Args:
        exc: The exception to handle

    Returns:
        The exit code to use
    """
    # Get exit code (default 1)
    exit_code = getattr(exc, "exit_code", 1)

    # Get result for formatting
    if isinstance(exc, CLIError):
        result = exc.to_result()
    elif hasattr(exc, "to_result"):
        # Duck typing for exceptions with to_result method
        result = exc.to_result()  # pyright: ignore[reportAttributeAccessIssue]
    else:
        # Generic exception - wrap it
        data = getattr(exc, "data", {})
        result = Result(success=False, message=str(exc), data=data)

    if is_json_output():
        typer.echo(json.dumps(result.to_dict()), err=True)
    else:
        typer.secho(f"Error: {result.message}", fg=typer.colors.RED, err=True)

    return exit_code


def set_json_output(enabled: bool) -> None:
    """Set the global JSON output mode."""
    _state["json"] = enabled


def is_json_output() -> bool:
    """Check if JSON output is enabled."""
    return _state["json"]


@dataclass
class Result:
    """Base result class for command output."""

    success: bool = True
    message: str = ""
    data: dict[str, Any] = field(default_factory=dict)

    def to_dict(self) -> dict[str, Any]:
        """Convert to dictionary for JSON output."""
        result: dict[str, Any] = {"success": self.success}
        if self.message:
            result["message"] = self.message
        if self.data:
            result.update(self.data)
        return result


def output(result: Result) -> None:
    """Output a result in the appropriate format."""
    if is_json_output():
        typer.echo(json.dumps(result.to_dict(), indent=2))
    else:
        _print_result(result)


def output_data(data: dict[str, Any]) -> None:
    """Output raw data in the appropriate format."""
    if is_json_output():
        typer.echo(json.dumps(data, indent=2))
    else:
        _print_dict(data)


def _print_result(result: Result) -> None:
    """Print result with colored status indicator."""
    if result.success:
        if result.message:
            typer.secho("✓ ", fg=typer.colors.GREEN, nl=False)
            typer.echo(result.message)
    else:
        if result.message:
            typer.secho("✗ ", fg=typer.colors.RED, nl=False, err=True)
            typer.echo(result.message, err=True)


def _print_dict(data: dict[str, Any], indent: int = 0) -> None:
    """Print a dictionary in a readable format."""
    prefix = "  " * indent
    for key, value in data.items():
        if isinstance(value, dict):
            typer.echo(f"{prefix}{key}:")
            _print_dict(value, indent + 1)
        elif isinstance(value, list):
            typer.echo(f"{prefix}{key}:")
            for item in value:
                if isinstance(item, dict):
                    _print_dict(item, indent + 1)
                    typer.echo()
                else:
                    typer.echo(f"{prefix}  - {item}")
        else:
            typer.echo(f"{prefix}{key}: {value}")


def print_table(
    headers: list[str],
    rows: list[list[str]],
    title: str | None = None,
) -> None:
    """Print a simple table."""
    if is_json_output():
        # Convert to list of dicts for JSON output
        data = [dict(zip(headers, row, strict=True)) for row in rows]
        typer.echo(json.dumps(data, indent=2))
        return

    if title:
        typer.echo(f"\n  {title}")
        typer.echo("  " + "-" * 50)

    if not rows:
        typer.echo("  (none)")
        return

    # Calculate column widths
    widths = [len(h) for h in headers]
    for row in rows:
        for i, cell in enumerate(row):
            widths[i] = max(widths[i], len(str(cell)))

    # Print header
    header_line = "  " + "  ".join(h.ljust(widths[i]) for i, h in enumerate(headers))
    typer.echo(header_line)
    typer.echo("  " + "  ".join("-" * w for w in widths))

    # Print rows
    for row in rows:
        row_line = "  " + "  ".join(str(cell).ljust(widths[i]) for i, cell in enumerate(row))
        typer.echo(row_line)

    typer.echo()


def print_status_section(title: str, items: dict[str, tuple[str, str | None]]) -> None:
    """Print a status section with colored values.

    Args:
        title: Section title
        items: Dict of label -> (value, color) where color is optional
    """
    if is_json_output():
        return  # JSON mode handles this differently

    typer.echo(f"\n  {title}")
    typer.echo("  " + "-" * 40)

    for label, (value, color) in items.items():
        typer.echo(f"  {label}: ", nl=False)
        if color:
            typer.secho(value, fg=getattr(typer.colors, color.upper(), None))
        else:
            typer.echo(value)
