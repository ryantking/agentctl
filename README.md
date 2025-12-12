# agentctl

A CLI tool for managing Claude Code configurations, hooks, and isolated workspaces using git worktrees.

## Features

- **Workspace Management**: Create isolated git worktree-based workspaces for parallel Claude Code sessions
- **Hook Integration**: Seamless integration with Claude Code lifecycle hooks
- **Auto-commit**: Automatic git commits on feature branches for Edit/Write operations
- **Context Injection**: Live git/workspace status injected into every Claude prompt
- **Notification System**: macOS notifications with automatic agent detection (Claude Code, Cursor, Cursor Agent)
- **Tab Completion**: Intelligent tab completion for workspace commands

## Installation

### Via Go Install

```bash
go install github.com/ryantking/agentctl/cmd/agentctl@latest
```

### Download Binary

Download the appropriate binary for your platform from the [latest release](https://github.com/ryantking/agentctl/releases/latest).

Extract and place the binary in your PATH (e.g., `/usr/local/bin` or `~/bin`).

## Quick Start

```bash
# Check installation
agentctl status

# Show version
agentctl version

# Initialize Claude Code configuration
agentctl init

# Initialize or update memory files
agentctl memory init

# Validate memory files
agentctl memory validate

# Show memory file contents
agentctl memory show AGENTS.md

# Create a new workspace
agentctl workspace create my-feature-branch

# List all workspaces (with tab completion!)
agentctl workspace list

# Show workspace details
agentctl workspace status my-feature-branch

# Delete a workspace
agentctl workspace delete my-feature-branch
```

## Commands

### Workspace Commands

Manage git worktree-based workspaces for parallel development sessions.

- `agentctl workspace create <branch> [--base <branch>]` - Create new workspace with git worktree
- `agentctl workspace list [--json]` - List all workspaces (includes main/master, shows current with `*`)
- `agentctl workspace show [branch]` - Print workspace path (for shell integration)
- `agentctl workspace status [branch]` - Show detailed workspace status
- `agentctl workspace delete [branch] [--force]` - Delete a workspace
- `agentctl workspace clean` - Remove all clean workspaces

**Tab Completion**: Workspace commands (`show`, `status`, `delete`) support tab completion for branch names.

**JSON Output**: Use `--json` flag on any workspace command for programmatic access:

```bash
agentctl workspace list --json
agentctl workspace show refactor/golang --json
```

### Hook Commands

Hook commands are designed to be called from Claude Code hooks. They handle stdin parsing, error handling, and exit codes appropriately.

- `agentctl hook inject-context` - Inject git/workspace context into prompts
- `agentctl hook notify-input [message]` - Send notification when input is needed
- `agentctl hook notify-stop` - Send notification when a task completes
- `agentctl hook notify-error [message]` - Send error notification
- `agentctl hook post-edit` - Auto-commit Edit tool changes
- `agentctl hook post-write` - Auto-commit Write tool changes (new files)

**Notification Agent Detection**: Notifications automatically detect the agent environment and use the appropriate icon:
- **Cursor Agent** (TUI): Detected via `CURSOR_AGENT=1` and `CURSOR_CLI_COMPAT=1`
- **Cursor IDE**: Detected via `CURSOR_AGENT=1` (without `CURSOR_CLI_COMPAT`)
- **Claude Code**: Detected via `CLAUDECODE=1`

You can override the sender with `AGENT_NOTIFICATION_SENDER` environment variable.

### Init Command

Initialize Claude Code configuration in a repository or globally.

- `agentctl init` - Initialize Claude Code configuration
  - `--global` - Install to `$HOME/.claude` instead of current repository
  - `--force` - Overwrite existing files
  - `--no-index` - Skip Claude CLI repository indexing

Installs agents, skills, settings, MCP config, and memory files (AGENTS.md and CLAUDE.md).

### Memory Commands

Manage agent memory files (AGENTS.md and CLAUDE.md) following cross-platform standards.

- `agentctl memory init` - Initialize memory files from templates
  - `--global` - Install to `$HOME/.claude` instead of current repository
  - `--force` - Overwrite existing files
  - `--no-index` - Skip repository indexing step

- `agentctl memory show [file]` - Display memory file contents
  - `--resolve` - Expand @imports inline
  - `--json` - Output as JSON with metadata

- `agentctl memory validate` - Validate memory files for common issues
  - Checks line counts, imports, circular imports, required sections, and conflicts

- `agentctl memory index` - Generate repository overview and inject into AGENTS.md
  - `--timeout <seconds>` - Override default 90-second timeout

**Memory File Structure:**
- **AGENTS.md**: Universal agent instructions (cross-platform standard)
- **CLAUDE.md**: Claude Code-specific orchestration patterns (imports @AGENTS.md)

### Other Commands

- `agentctl version` - Show the current version
- `agentctl status` - Show the status of Claude Code installation
- `agentctl completion [bash|zsh|fish|powershell]` - Generate shell completion scripts

## Development

### Prerequisites

- Go 1.24+
- `just` (command runner)
- `golangci-lint` (for linting)
- `gofumpt` (for formatting)
- `govulncheck` (for vulnerability checking)
- macOS (for full feature support, including notifications)

### Setup

```bash
# Install system dependencies
just deps

# Install globally
just install

# Run tests
just test

# Run linter
just lint

# Format code
just format

# Run all CI checks
just ci
```

### Building

```bash
# Build binary
just build

# Clean build artifacts
just clean
```

## Release Process

1. Bump version:
   ```bash
   just release patch  # or minor, major
   ```

2. Push changes and tags:
   ```bash
   git push && git push --tags
   ```

3. GitHub Actions automatically:
   - Builds binaries for multiple platforms
   - Creates GitHub Release with assets

## License

MIT - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions welcome! Please open an issue or pull request.
