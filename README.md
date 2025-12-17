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

# Initialize .agent directory with default rules
agentctl rules init

# List all rules
agentctl rules list

# Show a specific rule
agentctl rules show git-workflow

# Add a new rule
agentctl rules add "Always validate input" --name input-validation

# Sync rules to different formats
agentctl rules sync

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

Installs agents, skills, settings, MCP config, and initializes `.agent/` directory with default rules.

### Rules Commands

Manage agent rules in the `.agent/` directory. Rules are modular `.mdc` files with YAML frontmatter that can be synced to different formats.

- `agentctl rules init` - Initialize `.agent/` directory with default rules
  - `--force` - Overwrite existing files
  - `--no-project` - Skip project.md generation

- `agentctl rules list` - List all rules with metadata
  - `--json` - Output as JSON for programmatic access

- `agentctl rules show [rule-name]` - Display rule content
  - `--raw` - Output raw mdc file without pretty-printing

- `agentctl rules add [prompt]` - Add a new rule from a description
  - `--name <filename>` - Specify filename (without .mdc extension)
  - `--description <text>` - Rule description (auto-generated if not provided)
  - `--applies-to <tools>` - Comma-separated list of tools (default: claude)

- `agentctl rules remove [rule-name...]` - Remove rule files
  - `--force` - Skip confirmation prompt
  - Supports removing multiple rules at once

- `agentctl rules sync` - Sync rules to different formats
  - `--cursor` - Copy to `.cursor/rules/` (Cursor format)
  - `--claude` - Convert to `.claude/skills/<name>/SKILL.md` (Claude skills)
  - `--agents` - Generate `AGENTS.md` table of contents
  - `--claude-md` - Generate `CLAUDE.md` overview
  - If no flags specified, syncs to all formats

**Directory Structure:**
- `.agent/rules/` - Rule files (.mdc format with YAML frontmatter) - source of truth
- `.agent/research/` - Research artifacts (cached findings)
- `.agent/project.md` - High-level repository description

**Rule Schema:**
- `name` (required): Unique identifier for the rule
- `description` (optional): When/why this rule applies
- `globs` (optional): File patterns where rule is relevant (e.g., `["**/.beads/**"]`)
- `applies-to`, `priority`, `tags`, `version` (optional): Additional metadata

**Environment Variable:**
- `AGENTDIR` - Override default `.agent` location (defaults to `.agent`)

**Rule File Format (.mdc):**
Rules use YAML frontmatter with required field (`name`) and optional fields (`description`, `globs`, `applies-to`, `priority`, `tags`, `version`). See [docs/rules.md](docs/rules.md) for full schema documentation.

### Memory Commands (Deprecated)

**Note:** The `agentctl rules` system provides a modular approach to managing agent instructions. See [Rules Documentation](docs/rules.md) for details.

- `agentctl memory init` - Initialize memory files from templates (deprecated)
- `agentctl memory show [file]` - Display memory file contents (deprecated)
- `agentctl memory validate` - Validate memory files (deprecated)
- `agentctl memory index` - Generate repository overview (deprecated)

### Other Commands

- `agentctl version` - Show the current version
- `agentctl status` - Show authentication status and API connectivity
- `agentctl completion [bash|zsh|fish|powershell]` - Generate shell completion scripts

## Development

### Prerequisites

- Go 1.24+
- `just` (command runner)
- `golangci-lint` (for linting)
- `gofumpt` (for formatting)
- `govulncheck` (for vulnerability checking)
- macOS (for full feature support, including notifications)

### Authentication

Some features (like agent-based rule generation and project.md generation) require authentication with the Anthropic API. There are two authentication methods:

**Method 1: Claude Code Session (Automatic)**
If you're already logged into Claude Code, authentication is automatic. No setup needed!

**Method 2: API Key**
Set the `ANTHROPIC_API_KEY` environment variable:

```bash
export ANTHROPIC_API_KEY=your-api-key-here
```

Get your API key from [console.anthropic.com](https://console.anthropic.com/).

**Note:** If both are available, `ANTHROPIC_API_KEY` takes precedence over the Claude Code session.

If not configured, features that require authentication will show helpful error messages directing you to set up authentication.

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
