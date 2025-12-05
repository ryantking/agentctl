"""agentctl - CLI for managing Claude Code configurations."""

from __future__ import annotations

import json
import shutil
import subprocess
from dataclasses import dataclass
from typing import Annotated

import typer

from agentctl.cli.commands.hook import app as hook_app
from agentctl.cli.commands.init import app as init_app
from agentctl.cli.commands.workspace import app as workspace_app
from agentctl.cli.output import CLIError, handle_exception, is_json_output, set_json_output


@dataclass
class StatusInfo:
    """System status information."""

    claude_installed: bool
    claude_version: str | None
    claude_path: str | None

    def to_dict(self) -> dict:
        """Convert to dictionary for JSON output."""
        return {
            "claude": {
                "installed": self.claude_installed,
                "version": self.claude_version,
                "path": self.claude_path,
            },
        }


app = typer.Typer(
    name="agentctl",
    help="CLI for managing Claude Code configurations and workflows.",
    no_args_is_help=True,
)

# Register subcommands
app.add_typer(hook_app, name="hook")
app.add_typer(init_app, name="init")
app.add_typer(workspace_app, name="workspace")


@app.callback()
def main_callback(
    json_output: Annotated[
        bool,
        typer.Option(
            "--json",
            "-j",
            help="Output result as JSON",
        ),
    ] = False,
) -> None:
    """CLI for managing Claude Code configurations and workflows."""
    set_json_output(json_output)


def get_claude_info() -> tuple[bool, str | None, str | None]:
    """Get Claude Code installation info."""
    claude_path = shutil.which("claude")
    if not claude_path:
        return False, None, None

    try:
        result = subprocess.run(
            ["claude", "--version"],
            capture_output=True,
            text=True,
            timeout=10,
        )
        version = result.stdout.strip() if result.returncode == 0 else None
        return True, version, claude_path
    except (subprocess.TimeoutExpired, FileNotFoundError):
        return True, None, claude_path


@app.command()
def version() -> None:
    """Show the current version."""
    from agentctl import __version__

    if is_json_output():
        typer.echo(json.dumps({"version": __version__}))
    else:
        typer.echo(f"agentctl {__version__}")


@app.command()
def status() -> None:
    """Show the status of Claude Code."""
    # Get Claude info
    claude_installed, claude_version, claude_path = get_claude_info()

    info = StatusInfo(
        claude_installed=claude_installed,
        claude_version=claude_version,
        claude_path=claude_path,
    )

    if is_json_output():
        typer.echo(json.dumps(info.to_dict(), indent=2))
    else:
        _print_status(info)


def _print_status(info: StatusInfo) -> None:
    """Print status in human-readable format."""
    typer.echo("\n  Claude Code")
    typer.echo("  " + "-" * 40)
    if info.claude_installed:
        typer.echo("  Status:   ", nl=False)
        typer.secho("installed", fg=typer.colors.GREEN)
        typer.echo(f"  Version:  {info.claude_version or 'unknown'}")
        typer.echo(f"  Path:     {info.claude_path}")
    else:
        typer.echo("  Status:   ", nl=False)
        typer.secho("not installed", fg=typer.colors.RED)
    typer.echo()


def main() -> None:
    """Entry point for the CLI."""
    try:
        app()
    except CLIError as e:
        # Handle our custom CLI errors
        exit_code = handle_exception(e)
        raise SystemExit(exit_code) from None
    except Exception as e:
        # Check if exception has CLI-compatible attributes
        if hasattr(e, "exit_code") or hasattr(e, "to_result"):
            exit_code = handle_exception(e)
            raise SystemExit(exit_code) from None
        # Re-raise unexpected exceptions
        raise


if __name__ == "__main__":
    main()
