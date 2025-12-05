# Python to Go Migration Plan: agentctl

## Overview

This plan outlines the migration of `agentctl` from Python (3,276 LOC) to Go. The application is a CLI tool for managing Claude Code workspaces using git worktrees.

---

## 1. Technology Stack

### Core Dependencies

| Component | Library | Version | Justification |
|-----------|---------|---------|---------------|
| CLI Framework | `github.com/spf13/cobra` | v1.8+ | Battle-tested (kubectl, helm), excellent subcommand support, shell completion |
| Terminal Styling | `github.com/charmbracelet/lipgloss` | v1.0+ | Modern CSS-like styling, Charmbracelet ecosystem |
| TUI Components | `github.com/charmbracelet/bubbles` | v0.20+ | Spinners, tables, progress bars |
| Interactive Forms | `github.com/charmbracelet/huh` | v0.6+ | Confirmations, selections |
| Git Operations | `os/exec` + git CLI | - | go-git lacks worktree support |
| Configuration | `github.com/knadh/koanf/v2` | v2.1+ | Lighter than Viper, preserves key case |
| JSON Handling | `encoding/json` | stdlib | Sufficient for needs |
| Testing | `github.com/stretchr/testify` | v1.9+ | Assertions and mocking |

### Development Tools

| Tool | Purpose | Configuration File |
|------|---------|-------------------|
| golangci-lint v2 | Linting & static analysis | `.golangci.yml` |
| gofumpt | Code formatting | Via gopls settings |
| Task (go-task) | Build automation | `Taskfile.yml` |
| GoReleaser | Release & distribution | `.goreleaser.yaml` |
| govulncheck | Vulnerability scanning | - |

---

## 2. Directory Layout

```
agentctl/
├── cmd/
│   └── agentctl/
│       └── main.go                 # Entry point (~30 lines)
├── internal/
│   ├── cli/                        # Cobra command definitions
│   │   ├── root.go                 # Root command, global flags, Execute()
│   │   ├── version.go              # version command
│   │   ├── status.go               # status command
│   │   ├── workspace/              # workspace subcommand group
│   │   │   ├── workspace.go        # Parent command
│   │   │   ├── create.go
│   │   │   ├── list.go
│   │   │   ├── show.go
│   │   │   ├── status.go
│   │   │   ├── diff.go
│   │   │   ├── delete.go
│   │   │   ├── clean.go
│   │   │   └── open.go
│   │   ├── hook/                   # hook subcommand group
│   │   │   ├── hook.go             # Parent command
│   │   │   ├── post_edit.go
│   │   │   ├── post_write.go
│   │   │   ├── context_info.go
│   │   │   └── notify.go           # notify-input, notify-stop, etc.
│   │   └── init/                   # init subcommand
│   │       └── init.go
│   ├── workspace/                  # Workspace business logic
│   │   ├── workspace.go            # Workspace struct, methods
│   │   ├── manager.go              # WorkspaceManager (create/delete/clean)
│   │   ├── errors.go               # Custom error types
│   │   └── workspace_test.go
│   ├── hook/                       # Hook implementations
│   │   ├── autocommit.go           # Auto-commit logic
│   │   ├── context.go              # Context injection
│   │   ├── notify.go               # Desktop notifications
│   │   └── hook_test.go
│   ├── git/                        # Git operations
│   │   ├── git.go                  # Repo root, branch queries
│   │   ├── worktree.go             # Worktree parsing and operations
│   │   ├── status.go               # Git status parsing
│   │   └── git_test.go
│   ├── config/                     # Configuration management
│   │   ├── config.go               # Config loading
│   │   └── config_test.go
│   ├── output/                     # Output formatting
│   │   ├── result.go               # Result struct
│   │   ├── format.go               # JSON/human output
│   │   └── styles.go               # lipgloss styles
│   └── templates/                  # Template handling
│       ├── embed.go                # go:embed templates
│       └── install.go              # Template installation logic
├── templates/                      # Bundled templates (embedded)
│   ├── CLAUDE.md
│   ├── settings.json
│   ├── agents/
│   └── skills/
├── .golangci.yml                   # Linter configuration
├── .goreleaser.yaml                # Release configuration
├── Taskfile.yml                    # Build automation
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

---

## 3. Type Mappings

### Python to Go Translation

| Python | Go |
|--------|-----|
| `@dataclass Workspace` | `type Workspace struct` |
| `Exception` hierarchy | Custom error types with `error` interface |
| `Result` dataclass | `type Result[T any] struct` |
| `typer.Typer` app | `*cobra.Command` |
| `subprocess.run()` | `exec.Command().Run()` |
| `pathlib.Path` | `string` + `path/filepath` |
| `json.loads/dumps` | `json.Unmarshal/Marshal` |
| `sys.stdin` | `os.Stdin` + `bufio.Scanner` |
| `os.environ` | `os.Environ()` |
| `shutil.copy2()` | `io.Copy()` + `os.Chmod()` |

### Error Types

```go
// internal/workspace/errors.go
package workspace

import "errors"

var (
    ErrWorkspaceExists   = errors.New("workspace already exists")
    ErrWorkspaceNotFound = errors.New("workspace not found")
    ErrBranchInUse       = errors.New("branch is already checked out")
    ErrNotInGitRepo      = errors.New("not in a git repository")
)

type WorkspaceError struct {
    Workspace string
    Op        string
    Err       error
}

func (e *WorkspaceError) Error() string {
    return fmt.Sprintf("%s %s: %v", e.Op, e.Workspace, e.Err)
}

func (e *WorkspaceError) Unwrap() error {
    return e.Err
}
```

### Result Type

```go
// internal/output/result.go
package output

type Result struct {
    Success bool        `json:"success"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
}

func Success(data interface{}) Result {
    return Result{Success: true, Data: data}
}

func Error(msg string) Result {
    return Result{Success: false, Message: msg}
}
```

---

## 4. Command Structure

### Root Command

```go
// internal/cli/root.go
package cli

import (
    "github.com/spf13/cobra"
)

var (
    jsonOutput bool
)

func NewRootCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "agentctl",
        Short: "CLI for managing Claude Code workspaces",
    }

    cmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

    cmd.AddCommand(
        NewVersionCmd(),
        NewStatusCmd(),
        NewWorkspaceCmd(),
        NewHookCmd(),
        NewInitCmd(),
    )

    return cmd
}
```

### Workspace Subcommand Example

```go
// internal/cli/workspace/create.go
package workspace

import (
    "github.com/spf13/cobra"
    ws "github.com/ryantking/agentctl/internal/workspace"
)

func NewCreateCmd() *cobra.Command {
    var baseBranch string

    cmd := &cobra.Command{
        Use:   "create <branch>",
        Short: "Create a new workspace",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            manager, err := ws.NewManager()
            if err != nil {
                return err
            }
            return manager.CreateWorkspace(args[0], baseBranch)
        },
    }

    cmd.Flags().StringVar(&baseBranch, "base", "", "Base branch for new workspace")

    return cmd
}
```

---

## 5. Git Operations

### Worktree Management

```go
// internal/git/worktree.go
package git

import (
    "bufio"
    "os/exec"
    "strings"
)

type Worktree struct {
    Path   string
    Branch string
    Commit string
}

func ListWorktrees(repoRoot string) ([]Worktree, error) {
    cmd := exec.Command("git", "-C", repoRoot, "worktree", "list", "--porcelain")
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    return parseWorktreeList(string(output)), nil
}

func AddWorktree(repoRoot, path, branch string, createBranch bool) error {
    args := []string{"-C", repoRoot, "worktree", "add"}
    if createBranch {
        args = append(args, "-b", branch)
    }
    args = append(args, path, branch)

    cmd := exec.Command("git", args...)
    return cmd.Run()
}

func RemoveWorktree(repoRoot, path string, force bool) error {
    args := []string{"-C", repoRoot, "worktree", "remove"}
    if force {
        args = append(args, "--force")
    }
    args = append(args, path)

    cmd := exec.Command("git", args...)
    return cmd.Run()
}
```

---

## 6. Template Embedding

```go
// internal/templates/embed.go
package templates

import "embed"

//go:embed all:templates
var FS embed.FS

func GetTemplate(name string) ([]byte, error) {
    return FS.ReadFile("templates/" + name)
}
```

---

## 7. Configuration Files

### Taskfile.yml

```yaml
version: '3'

vars:
  VERSION:
    sh: git describe --tags --always --dirty 2>/dev/null || echo "dev"
  COMMIT:
    sh: git rev-parse --short HEAD 2>/dev/null || echo "none"
  LDFLAGS: -s -w -X main.version={{.VERSION}} -X main.commit={{.COMMIT}}
  BINARY: agentctl

tasks:
  default:
    deps: [build]

  build:
    desc: Build the binary
    cmds:
      - go build -ldflags "{{.LDFLAGS}}" -o {{.BINARY}} ./cmd/agentctl

  install:
    desc: Install globally
    cmds:
      - go install -ldflags "{{.LDFLAGS}}" ./cmd/agentctl

  test:
    desc: Run tests
    cmds:
      - go test -race -coverprofile=coverage.out ./...

  lint:
    desc: Run linters
    cmds:
      - golangci-lint run

  fmt:
    desc: Format code
    cmds:
      - gofumpt -w .
      - goimports -w .

  vuln:
    desc: Check for vulnerabilities
    cmds:
      - govulncheck ./...

  ci:
    desc: Run all CI checks
    deps: [lint, test, vuln]

  clean:
    desc: Clean build artifacts
    cmds:
      - rm -f {{.BINARY}}
      - rm -f coverage.out

  release:
    desc: Create a release (use with VERSION=x.y.z)
    cmds:
      - git tag -a v{{.VERSION}} -m "Release v{{.VERSION}}"
      - git push origin v{{.VERSION}}
      - goreleaser release --clean
```

### .golangci.yml

```yaml
version: "2"

formatters:
  enable:
    - goimports
    - gofumpt

linters:
  default: standard
  enable:
    - gosec
    - bodyclose
    - govet
    - staticcheck
    - errcheck
    - gocritic
    - gocyclo
    - err113
    - wrapcheck
    - modernize

  settings:
    gocyclo:
      min-complexity: 15

issues:
  exclude-rules:
    - path: "_test\\.go"
      linters: [errcheck, gosec]
```

### .goreleaser.yaml

```yaml
version: 2

project_name: agentctl

builds:
  - main: ./cmd/agentctl
    binary: agentctl
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}
    flags:
      - -trimpath

archives:
  - formats: [tar.gz]
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

brews:
  - name: agentctl
    repository:
      owner: ryantking
      name: homebrew-tap
    homepage: https://github.com/ryantking/agentctl
    description: CLI for managing Claude Code workspaces
    license: MIT
    install: |
      bin.install "agentctl"

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^chore:'
```

### go.mod

```go
module github.com/ryantking/agentctl

go 1.23

require (
    github.com/spf13/cobra v1.8.1
    github.com/charmbracelet/lipgloss v1.0.0
    github.com/charmbracelet/bubbles v0.20.0
    github.com/charmbracelet/huh v0.6.0
    github.com/knadh/koanf/v2 v2.1.2
    github.com/stretchr/testify v1.9.0
)
```

---

## 8. Implementation Phases

### Phase 1: Foundation (4 packages)

**Goal**: Core infrastructure and basic CLI skeleton

1. **internal/git/** - Git operations
   - `git.go`: GetRepoRoot, GetCurrentBranch, BranchExists
   - `worktree.go`: ListWorktrees, AddWorktree, RemoveWorktree, parseWorktreeList
   - `status.go`: IsWorktreeClean, GetStatusSummary

2. **internal/output/** - Output handling
   - `result.go`: Result struct, Success/Error constructors
   - `format.go`: FormatJSON, FormatHuman functions
   - `styles.go`: lipgloss styles for success/error/info

3. **internal/workspace/** - Workspace domain
   - `workspace.go`: Workspace struct, ToMap, IsManaged, IsClean
   - `errors.go`: Custom error types

4. **cmd/agentctl/main.go** + **internal/cli/root.go**
   - Entry point, root command, version command

**Deliverable**: `agentctl version` works

### Phase 2: Workspace Commands (2 packages)

**Goal**: Full workspace management

1. **internal/workspace/manager.go**
   - CreateWorkspace, DeleteWorkspace, CleanWorkspaces
   - GetWorkspaceStatus, GetWorkspaceDiff

2. **internal/cli/workspace/**
   - All workspace subcommands (create, list, show, status, diff, delete, clean, open)

**Deliverable**: `agentctl workspace create/list/delete` work

### Phase 3: Hook Commands (2 packages)

**Goal**: Claude Code hook integration

1. **internal/hook/**
   - `autocommit.go`: PostEdit, PostWrite logic
   - `context.go`: ContextInfo generation
   - `notify.go`: Desktop notifications

2. **internal/cli/hook/**
   - All hook subcommands

**Deliverable**: All hooks functional

### Phase 4: Init & Templates (2 packages)

**Goal**: Template installation

1. **internal/templates/**
   - `embed.go`: Embedded templates
   - `install.go`: Template installation logic

2. **internal/cli/init/**
   - init command with flags

**Deliverable**: `agentctl init` works

### Phase 5: Polish & Release

**Goal**: Production readiness

1. Shell completion (Cobra built-in)
2. Man page generation
3. Integration tests
4. CI/CD workflows
5. Homebrew formula
6. Documentation

---

## 9. Testing Strategy

### Unit Tests

```go
// internal/git/worktree_test.go
func TestParseWorktreeList(t *testing.T) {
    input := `worktree /path/to/main
HEAD abc123
branch refs/heads/main

worktree /path/to/feature
HEAD def456
branch refs/heads/feature
`
    worktrees := parseWorktreeList(input)

    assert.Len(t, worktrees, 2)
    assert.Equal(t, "/path/to/main", worktrees[0].Path)
    assert.Equal(t, "main", worktrees[0].Branch)
}
```

### Integration Tests

```go
// internal/workspace/manager_test.go
func TestWorkspaceManager_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Create temp git repo
    tmpDir := t.TempDir()
    exec.Command("git", "init", tmpDir).Run()
    exec.Command("git", "-C", tmpDir, "commit", "--allow-empty", "-m", "init").Run()

    manager, err := workspace.NewManagerAt(tmpDir)
    require.NoError(t, err)

    // Test create
    ws, err := manager.CreateWorkspace("test-branch", "")
    require.NoError(t, err)
    assert.Equal(t, "test-branch", ws.Branch)

    // Test delete
    err = manager.DeleteWorkspace("test-branch", false)
    require.NoError(t, err)
}
```

---

## 10. Migration Checklist

### Before Starting
- [ ] Set up Go 1.23+ environment
- [ ] Install Task: `brew install go-task/tap/go-task`
- [ ] Install golangci-lint: `brew install golangci-lint`
- [ ] Install gofumpt: `go install mvdan.cc/gofumpt@latest`
- [ ] Install GoReleaser: `brew install goreleaser`

### Phase 1 Checklist
- [ ] Initialize go.mod
- [ ] Create directory structure
- [ ] Implement internal/git package
- [ ] Implement internal/output package
- [ ] Implement internal/workspace (domain only)
- [ ] Create cmd/agentctl/main.go
- [ ] Create internal/cli/root.go
- [ ] Implement version command
- [ ] Write unit tests
- [ ] Configure golangci-lint
- [ ] Set up Taskfile.yml

### Phase 2 Checklist
- [ ] Implement WorkspaceManager
- [ ] Implement workspace create command
- [ ] Implement workspace list command
- [ ] Implement workspace show command
- [ ] Implement workspace status command
- [ ] Implement workspace diff command
- [ ] Implement workspace delete command
- [ ] Implement workspace clean command
- [ ] Implement workspace open command
- [ ] Write integration tests

### Phase 3 Checklist
- [ ] Implement hook autocommit logic
- [ ] Implement hook context injection
- [ ] Implement hook notifications
- [ ] Implement post-edit command
- [ ] Implement post-write command
- [ ] Implement context-info command
- [ ] Implement notify-* commands
- [ ] Test with Claude Code integration

### Phase 4 Checklist
- [ ] Embed templates with go:embed
- [ ] Implement template installation
- [ ] Implement settings merge logic
- [ ] Implement init command
- [ ] Test template installation

### Phase 5 Checklist
- [ ] Add shell completion
- [ ] Set up .goreleaser.yaml
- [ ] Create GitHub Actions CI workflow
- [ ] Create release workflow
- [ ] Set up Homebrew tap
- [ ] Write README
- [ ] Tag v1.0.0

---

## 11. Key Differences from Python

| Aspect | Python | Go |
|--------|--------|-----|
| Error handling | Exceptions + try/except | Return errors + if err != nil |
| Package visibility | `_private` convention | lowercase unexported |
| Entry point | `if __name__ == "__main__"` | `func main()` |
| Type hints | Optional annotations | Mandatory static types |
| Dependency resolution | pip/uv at runtime | Compiled into binary |
| CLI parsing | typer decorators | Cobra command tree |
| Rich output | rich.Console | lipgloss styles |
| Templates | importlib.resources | embed.FS |
| Testing | pytest + fixtures | testing package + testify |
| Build | uv build | go build |

---

## 12. Notes

### go-git Limitation
go-git does NOT support `git worktree` commands. Always use `os/exec` with the git CLI for:
- `git worktree add`
- `git worktree list`
- `git worktree remove`

### Shell Completion
Cobra provides excellent built-in completion. Generate with:
```go
cmd.AddCommand(&cobra.Command{
    Use:   "completion [bash|zsh|fish|powershell]",
    Short: "Generate completion script",
    RunE:  completionFunc,
})
```

### Version Injection
Use ldflags at build time:
```go
// main.go
var version = "dev"

func main() {
    fmt.Println("agentctl", version)
}
```
Build: `go build -ldflags "-X main.version=1.0.0"`
