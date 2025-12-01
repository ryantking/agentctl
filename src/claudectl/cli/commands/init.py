"""Init command for claudectl - initialize Claude Code configuration."""

from __future__ import annotations

from pathlib import Path
from typing import Annotated

import typer

from claudectl.cli.output import Result, output
from claudectl.core.git import NotInGitRepoError
from claudectl.operations.init_ops import ImportDirNotFoundError, InitManager

app = typer.Typer(
    name="init",
    help="Initialize Claude Code configuration.",
    no_args_is_help=False,
)


@app.callback(invoke_without_command=True)
def init(
    ctx: typer.Context,
    global_install: Annotated[
        bool,
        typer.Option(
            "--global",
            "-g",
            help="Install to $HOME/.claude instead of current repository",
        ),
    ] = False,
    force: Annotated[
        bool,
        typer.Option(
            "--force",
            "-f",
            help="Overwrite existing files",
        ),
    ] = False,
    no_index: Annotated[
        bool,
        typer.Option(
            "--no-index",
            help="Skip Claude CLI repository indexing",
        ),
    ] = False,
    verbose: Annotated[
        bool,
        typer.Option(
            "--verbose",
            "-v",
            help="Show detailed file operations",
        ),
    ] = False,
) -> None:
    """Initialize Claude Code configuration.

    Installs CLAUDE.md, agents, skills, and settings from the bundled
    import/ directory. By default, skips existing files.

    Examples:
        claudectl init
        claudectl init --global
        claudectl init --force
        claudectl init --no-index --verbose
    """
    # Skip if subcommand invoked
    if ctx.invoked_subcommand is not None:
        return

    try:
        # Determine target directory
        if global_install:
            target = Path.home() / ".claude"
        else:
            from claudectl.core.git import get_repo_root

            target = get_repo_root()

        # Run initialization
        manager = InitManager(target)
        result = manager.install(
            force=force,
            skip_index=no_index or global_install,
            verbose=verbose,
        )

        # Output result
        output(result)

    except NotInGitRepoError as e:
        # Suggest --global if not in repo
        msg = f"{e}\n\nRun from inside a git repository or use --global"
        result = Result(success=False, message=msg)
        output(result)
        raise typer.Exit(1) from None
    except ImportDirNotFoundError as e:
        result = Result(success=False, message=str(e))
        output(result)
        raise typer.Exit(1) from None
