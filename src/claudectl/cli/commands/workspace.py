"""Workspace management commands for claudectl.

Commands for managing Claude workspaces (git worktrees) with proper
separation from terminal multiplexing.
"""

from __future__ import annotations

import json
from typing import Annotated

import typer

from claudectl.cli.output import Result, is_json_output, output, print_table
from claudectl.domain.exceptions import (
    BranchInUseError,
    NotInGitRepoError,
    WorkspaceError,
    WorkspaceExistsError,
    WorkspaceNotFoundError,
)
from claudectl.operations.context import copy_claude_context
from claudectl.operations.spawn import spawn_claude_in_shell
from claudectl.operations.workspace_ops import WorkspaceManager

app = typer.Typer(
    name="workspace",
    help="Manage Claude workspaces (git worktrees).",
    no_args_is_help=True,
)


@app.command("create")
def create(
    branch: Annotated[
        str,
        typer.Argument(
            help="Branch name for the workspace (e.g., feat/add-auth). Creates the branch if it doesn't exist."
        ),
    ],
    base: Annotated[
        str | None,
        typer.Option(
            "--base",
            "-b",
            help="Base branch to create from (defaults to current branch)",
        ),
    ] = None,
) -> None:
    """Create a new workspace with git worktree.

    Creates a new workspace at ~/.claude/workspaces/<repo>/<branch>/
    and copies necessary context files (CLAUDE.md, settings.local.json).

    Examples:
        claudectl workspace create feat/add-auth
        claudectl workspace create fix/bug-123 --base main
    """
    try:
        manager = WorkspaceManager()
        workspace = manager.create_workspace(branch, base_branch=base)

        if not is_json_output():
            typer.echo(f"Created workspace: {workspace.path}")

        # Copy Claude context files
        copied_files = copy_claude_context(workspace.path, workspace.repo_root)
        if copied_files and not is_json_output():
            typer.echo(f"Copied context: {', '.join(copied_files)}")

        result = Result(
            success=True,
            message="Workspace created successfully",
            data={
                "path": str(workspace.path),
                "branch": workspace.branch,
                "commit": workspace.commit,
            },
        )
        output(result)

    except NotInGitRepoError as e:
        result = Result(success=False, message=str(e))
        output(result)
        raise typer.Exit(1) from None
    except (WorkspaceExistsError, BranchInUseError, WorkspaceError) as e:
        result = Result(success=False, message=str(e))
        output(result)
        raise typer.Exit(1) from None


@app.command("open")
def open_workspace(
    branch: Annotated[
        str,
        typer.Argument(help="Branch name of workspace to open"),
    ],
) -> None:
    """Open Claude in a workspace directory.

    Spawns Claude with full environment inheritance to ensure
    MCPs and PATH are available. Uses the same terminal window.

    Example:
        claudectl workspace open feat/add-auth
    """
    try:
        manager = WorkspaceManager()
        workspace = manager.get_workspace(branch)

        # Spawn Claude with full environment inheritance
        # This ensures MCPs (via HOME and PATH) are available
        process = spawn_claude_in_shell(workspace.path)

        # Wait for Claude to exit
        process.wait()

    except NotInGitRepoError as e:
        result = Result(success=False, message=str(e))
        output(result)
        raise typer.Exit(1) from None
    except WorkspaceNotFoundError as e:
        result = Result(success=False, message=str(e))
        output(result)
        raise typer.Exit(1) from None
    except (OSError, ValueError) as e:
        result = Result(success=False, message=f"Failed to spawn Claude: {e}")
        output(result)
        raise typer.Exit(1) from None


@app.command("show")
def show(
    branch: Annotated[
        str,
        typer.Argument(help="Branch name of workspace to show"),
    ],
) -> None:
    """Print workspace path (for shell integration).

    Outputs the absolute path to the workspace directory.
    Useful for shell functions and scripts that want to spawn
    Claude in a new terminal window.

    Example:
        cd $(claudectl workspace show feat/add-auth)
    """
    try:
        manager = WorkspaceManager()
        workspace = manager.get_workspace(branch)

        if is_json_output():
            result = Result(
                success=True,
                data={
                    "path": str(workspace.path),
                    "branch": workspace.branch,
                },
            )
            output(result)
        else:
            # Just print the path for easy shell integration
            typer.echo(str(workspace.path))

    except NotInGitRepoError as e:
        result = Result(success=False, message=str(e))
        output(result)
        raise typer.Exit(1) from None
    except WorkspaceNotFoundError as e:
        result = Result(success=False, message=str(e))
        output(result)
        raise typer.Exit(1) from None


@app.command("list")
def list_workspaces(
    json_output: Annotated[
        bool,
        typer.Option(
            "--json",
            "-j",
            help="Output as JSON list of workspaces",
        ),
    ] = False,
) -> None:
    """List all managed workspaces.

    Shows workspaces in ~/.claude/workspaces/ with their status.
    """
    try:
        manager = WorkspaceManager()
        workspaces = manager.list_workspaces(managed_only=True)

        if json_output:
            # Output just the list as JSON
            data = [w.to_dict() for w in workspaces]
            typer.echo(json.dumps(data, indent=2))
            return

        if is_json_output():
            result = Result(
                success=True,
                data={"workspaces": [w.to_dict() for w in workspaces]},
            )
            output(result)
            return

        if not workspaces:
            typer.echo("\n  No managed workspaces found.\n")
            typer.echo("  Create one with: claudectl workspace create <branch>\n")
            return

        rows = []
        for workspace in workspaces:
            is_clean, status = workspace.is_clean
            status_icon = "✓" if is_clean else "●"
            rows.append(
                [
                    workspace.branch or "detached",
                    str(workspace.path),
                    f"{status_icon} {status}",
                ]
            )

        print_table(
            headers=["Branch", "Path", "Status"],
            rows=rows,
            title="Managed Workspaces",
        )

    except NotInGitRepoError as e:
        result = Result(success=False, message=str(e))
        output(result)
        raise typer.Exit(1) from None


@app.command("delete")
def delete(
    branch: Annotated[
        str,
        typer.Argument(help="Branch name of workspace to delete"),
    ],
    force: Annotated[
        bool,
        typer.Option(
            "--force",
            "-f",
            help="Force deletion even if workspace has uncommitted changes",
        ),
    ] = False,
) -> None:
    """Delete a workspace.

    By default, only deletes workspaces with no uncommitted changes.
    Use --force to delete even with changes (WARNING: data loss).

    Examples:
        claudectl workspace delete feat/add-auth
        claudectl workspace delete feat/add-auth --force
    """
    try:
        manager = WorkspaceManager()
        manager.delete_workspace(branch, force=force)

        result = Result(
            success=True,
            message=f"Deleted workspace for branch: {branch}",
            data={"branch": branch},
        )
        output(result)

    except NotInGitRepoError as e:
        result = Result(success=False, message=str(e))
        output(result)
        raise typer.Exit(1) from None
    except (WorkspaceNotFoundError, WorkspaceError) as e:
        result = Result(success=False, message=str(e))
        output(result)
        raise typer.Exit(1) from None


@app.command("clean")
def clean() -> None:
    """Remove all clean workspaces.

    Removes all workspaces that have no uncommitted changes.
    Useful for cleanup after completing work.

    Example:
        claudectl workspace clean
    """
    try:
        manager = WorkspaceManager()
        removed = manager.clean_workspaces(check_merged=True)

        if not removed:
            result = Result(
                success=True,
                message="No clean workspaces to remove",
            )
        else:
            result = Result(
                success=True,
                message=f"Removed {len(removed)} workspace(s)",
                data={"removed": removed},
            )
        output(result)

    except NotInGitRepoError as e:
        result = Result(success=False, message=str(e))
        output(result)
        raise typer.Exit(1) from None


@app.command("status")
def status(
    branch: Annotated[
        str,
        typer.Argument(help="Branch name of workspace to check"),
    ],
) -> None:
    """Show detailed workspace status.

    Displays status information including uncommitted changes,
    ahead/behind status relative to remote, and other details.

    Example:
        claudectl workspace status feat/add-auth
    """
    try:
        manager = WorkspaceManager()
        workspace = manager.get_workspace(branch)
        status_info = manager.get_workspace_status(workspace)

        if is_json_output():
            result = Result(success=True, data=status_info)
            output(result)
            return

        # Human-readable output
        typer.echo(f"\nWorkspace: {status_info['branch']}")
        typer.echo(f"Path:      {status_info['path']}")
        typer.echo(f"Commit:    {status_info['commit']}")
        typer.echo(f"Status:    {status_info['status']}")

        if status_info.get("ahead_behind"):
            ab = status_info["ahead_behind"]
            typer.echo(f"Sync:      {ab['ahead']} ahead, {ab['behind']} behind origin")

        typer.echo()

    except NotInGitRepoError as e:
        result = Result(success=False, message=str(e))
        output(result)
        raise typer.Exit(1) from None
    except WorkspaceNotFoundError as e:
        result = Result(success=False, message=str(e))
        output(result)
        raise typer.Exit(1) from None


@app.command("diff")
def diff(
    branch: Annotated[
        str,
        typer.Argument(help="Branch name of workspace to diff"),
    ],
    target: Annotated[
        str,
        typer.Option(
            "--target",
            "-t",
            help="Target branch to compare against",
        ),
    ] = "main",
) -> None:
    """Show git diff from workspace to target branch.

    Displays the diff between the workspace branch and a target branch
    (defaults to main). Useful for reviewing changes before merging.

    Examples:
        claudectl workspace diff feat/add-auth
        claudectl workspace diff feat/add-auth --target develop
    """
    try:
        manager = WorkspaceManager()
        workspace = manager.get_workspace(branch)
        diff_output = manager.get_workspace_diff(workspace, target)

        if is_json_output():
            result = Result(
                success=True,
                data={
                    "branch": branch,
                    "target": target,
                    "diff": diff_output,
                },
            )
            output(result)
        else:
            # Output diff directly to stdout for paging
            typer.echo(diff_output)

    except NotInGitRepoError as e:
        result = Result(success=False, message=str(e))
        output(result)
        raise typer.Exit(1) from None
    except (WorkspaceNotFoundError, WorkspaceError) as e:
        result = Result(success=False, message=str(e))
        output(result)
        raise typer.Exit(1) from None
