# agentctl

A CLI tool for managing Claude Code configurations, hooks, and isolated workspaces using git worktrees.

## Features

- **Workspace Management**: Create isolated git worktree-based workspaces for parallel Claude Code sessions
- **Hook Integration**: Seamless integration with Claude Code lifecycle hooks
- **Auto-commit**: Automatic git commits on feature branches for Edit/Write operations
- **Context Injection**: Live git/workspace status injected into every Claude prompt
- **Notification System**: macOS notifications for Claude Code events

## Installation

### Via pip from GitHub Release

```bash
# Latest release
pip install https://github.com/ryantking/agentctl/releases/latest/download/agentctl.tar.gz

# Specific version
pip install https://github.com/ryantking/agentctl/releases/download/v0.1.0/agentctl-0.1.0.tar.gz
```

### Via uv from GitHub Release

```bash
# Latest release
uv tool install https://github.com/ryantking/agentctl/releases/latest/download/agentctl.tar.gz

# Specific version
uv tool install https://github.com/ryantking/agentctl/releases/download/v0.1.0/agentctl-0.1.0.tar.gz
```

### From source

```bash
git clone https://github.com/ryantking/agentctl.git
cd agentctl
just install
```

## Quick Start

```bash
# Check installation
agentctl status

# Show version
agentctl version

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
- `agentctl workspace delete <branch>` - Delete a workspace
- `agentctl workspace clean` - Remove all clean workspaces

### Hook Commands

Hook commands are designed to be called from Claude Code hooks:

- `agentctl hook post-edit` - Auto-commit Edit tool changes
- `agentctl hook post-write` - Auto-commit Write tool changes (new files)
- `agentctl hook context-info` - Inject git/workspace context into prompts
- `agentctl hook notify-*` - Notification commands

## Development

### Prerequisites

- Python 3.13+
- uv
- just
- macOS (for full feature support)

### Setup

```bash
# Install dependencies
just deps

# Install in editable mode
just install

# Run tests
just test

# Run linter
just lint

# Format code
just format
```

### Running Tests

```bash
just test
```

### Building

```bash
just build
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
   - Builds the package
   - Creates GitHub Release

## License

MIT - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions welcome! Please open an issue or pull request.
