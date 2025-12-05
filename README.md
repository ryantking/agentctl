# agentctl

A CLI tool for managing Claude Code configurations, hooks, and isolated workspaces using git worktrees.

## Features

- **Workspace Management**: Create isolated git worktree-based workspaces for parallel Claude Code sessions
- **Hook Integration**: Seamless integration with Claude Code lifecycle hooks
- **Auto-commit**: Automatic git commits on feature branches for Edit/Write operations
- **Context Injection**: Live git/workspace status injected into every Claude prompt
- **Notification System**: macOS notifications for Claude Code events

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

# Create a new workspace
agentctl workspace create my-feature-branch

# List all workspaces
agentctl workspace list

# Show workspace details
agentctl workspace status my-feature-branch

# Delete a workspace
agentctl workspace delete my-feature-branch
```

## Commands

### Workspace Commands

- `agentctl workspace create <branch>` - Create new workspace with git worktree
- `agentctl workspace list [--json]` - List all managed workspaces
- `agentctl workspace show <branch>` - Print workspace path
- `agentctl workspace status <branch>` - Show detailed workspace status
- `agentctl workspace diff <branch> [--target <branch>]` - Show git diff from workspace to target branch
- `agentctl workspace delete <branch> [--force]` - Delete a workspace
- `agentctl workspace clean` - Remove all clean workspaces
- `agentctl workspace open <branch>` - Open Claude in a workspace directory

### Hook Commands

Hook commands are designed to be called from Claude Code hooks:

- `agentctl hook post-edit` - Auto-commit Edit tool changes
- `agentctl hook post-write` - Auto-commit Write tool changes (new files)
- `agentctl hook inject-context` - Inject git/workspace context into prompts
- `agentctl hook notify-input` - Send notification when Claude needs input
- `agentctl hook notify-stop` - Send notification when Claude completes a task
- `agentctl hook notify-error` - Send error notification
- `agentctl hook notify-test` - Send a test notification

### Init Command

- `agentctl init` - Initialize Claude Code configuration
  - `--global` - Install to $HOME/.claude instead of current repository
  - `--force` - Overwrite existing files
  - `--no-index` - Skip Claude CLI repository indexing

## Development

### Prerequisites

- Go 1.23+
- just
- golangci-lint
- gofumpt
- govulncheck
- macOS (for full feature support)

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

### Running Tests

```bash
just test
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
