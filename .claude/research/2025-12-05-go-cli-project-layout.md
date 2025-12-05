# Research: Go CLI Project Layout Best Practices 2024/2025
Date: 2025-12-05
Focus: Directory structure, module organization, build/release, and testing patterns for Go CLI applications
Agent: researcher

## Summary

Go CLI project layout best practices center on the golang-standards/project-layout conventions with `/cmd`, `/internal`, and optionally `/pkg` directories. For CLI applications, Cobra remains the dominant framework (35k+ stars), with emphasis on separating CLI layer from business logic. Build automation increasingly favors Task (Taskfile) over Makefile for Go projects, while GoReleaser handles cross-compilation and Homebrew tap distribution.

## Key Findings

- **Start simple, grow as needed**: Official Go docs recommend flat structure initially, adding `/cmd`, `/internal` only as complexity grows [Go Modules Layout](https://go.dev/doc/modules/layout)
- **`/internal` is compiler-enforced**: Go 1.4+ prevents external imports from `/internal` packages [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
- **`/pkg` is controversial**: Not an official standard; use only when explicitly exposing public APIs [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
- **Cobra dominates CLI space**: Used by kubectl, etcdctl, Hugo; supports nested commands and Viper integration [spf13/cobra](https://github.com/spf13/cobra)
- **Taskfile preferred over Makefile**: YAML syntax, cross-platform, checksum-based change detection [Applied Go](https://appliedgo.net/spotlight/just-make-a-task/)
- **GoReleaser + Homebrew**: One config file handles cross-compilation, releases, and tap formula generation [GoReleaser Docs](https://goreleaser.com/customization/homebrew_casks/)

## Detailed Analysis

### Directory Structure Conventions

The golang-standards/project-layout repository documents common patterns, though it explicitly states this is NOT an official Go standard. The key directories are:

**Core Go Directories:**

| Directory | Purpose | Compiler Enforced |
|-----------|---------|-------------------|
| `/cmd` | Main applications, one subdir per executable | No |
| `/internal` | Private code, not importable externally | Yes (Go 1.4+) |
| `/pkg` | Public library code (optional, use sparingly) | No |
| `/vendor` | Dependencies via `go mod vendor` | No |

**Application Support Directories:**

| Directory | Purpose |
|-----------|---------|
| `/api` | OpenAPI specs, protocol definitions |
| `/configs` | Configuration templates |
| `/scripts` | Build, install, analysis scripts |
| `/build` | CI/CD configs, packaging |
| `/test` | Integration tests, test data |
| `/docs` | Design documents, user docs |

**Anti-pattern**: Avoid `/src` - this is a Java/Python convention that creates unnecessary nesting in Go.

### CLI-Specific Structure (Cobra)

For a multi-command CLI like `agentctl`, the recommended structure separates:

1. **Entry point** (`cmd/agentctl/main.go`): Minimal, calls into internal
2. **Command definitions** (`internal/cli/` or `internal/cmd/`): Cobra command setup
3. **Business logic** (`internal/` subdirs): Actual implementation
4. **Shared utilities** (`internal/pkg/` or `pkg/`): Common code

```
agentctl/
├── cmd/
│   └── agentctl/
│       └── main.go           # Entry point, imports internal/cli
├── internal/
│   ├── cli/                  # CLI layer (Cobra commands)
│   │   ├── root.go          # Root command, global flags
│   │   ├── workspace.go     # workspace subcommand
│   │   ├── hook.go          # hook subcommand
│   │   └── init.go          # init subcommand
│   ├── workspace/           # Workspace business logic
│   ├── hook/                # Hook business logic
│   └── config/              # Configuration handling
├── go.mod
├── go.sum
└── Taskfile.yml             # Build automation
```

### Module Organization

**go.mod placement**: Always at repository root, module path should match repository:

```go
module github.com/owner/agentctl

go 1.22
```

**Installing**: Users can install via:
```bash
go install github.com/owner/agentctl/cmd/agentctl@latest
```

### Build Automation: Task vs Make vs Just

| Feature | Make | Task (Taskfile) | Just |
|---------|------|-----------------|------|
| Syntax | Tab-sensitive, cryptic | YAML, readable | Make-like, cleaner |
| Change detection | File timestamps | Checksums | None (command runner) |
| Cross-platform | Unix-centric | Native cross-platform | Cross-platform |
| Go ecosystem fit | Traditional | Modern Go projects | General purpose |
| Dependencies | File-based | Task-based | Recipe-based |

**Recommendation for Go CLI**: Use Taskfile.yml - YAML is familiar to Go developers, checksums are more reliable than timestamps, and descriptions improve discoverability.

Example Taskfile.yml:
```yaml
version: '3'

vars:
  BINARY: agentctl
  VERSION: '{{.GIT_TAG | default "dev"}}'

tasks:
  build:
    desc: Build the CLI
    cmds:
      - go build -ldflags "-X main.version={{.VERSION}}" -o bin/{{.BINARY}} ./cmd/agentctl
    sources:
      - ./**/*.go
    generates:
      - bin/{{.BINARY}}

  test:
    desc: Run tests
    cmds:
      - go test -v ./...

  lint:
    desc: Run linters
    cmds:
      - golangci-lint run

  install:
    desc: Install locally
    cmds:
      - go install ./cmd/agentctl
```

### Cross-Compilation

Go's built-in cross-compilation requires only environment variables:

```bash
# macOS ARM (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o bin/agentctl-darwin-arm64 ./cmd/agentctl

# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o bin/agentctl-darwin-amd64 ./cmd/agentctl

# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o bin/agentctl-linux-amd64 ./cmd/agentctl

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o bin/agentctl-linux-arm64 ./cmd/agentctl
```

For pure Go (no CGO), add `CGO_ENABLED=0` to avoid issues.

### GoReleaser Configuration

Example `.goreleaser.yaml`:
```yaml
version: 2
project_name: agentctl

builds:
  - id: agentctl
    main: ./cmd/agentctl
    binary: agentctl
    env:
      - CGO_ENABLED=0
    goos:
      - darwin
      - linux
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}}

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"

brews:
  - name: agentctl
    homepage: https://github.com/owner/agentctl
    description: CLI tool for managing Claude Code workspaces
    repository:
      owner: owner
      name: homebrew-tap
    commit_author:
      name: goreleaser
      email: bot@goreleaser.com
    install: |
      bin.install "agentctl"
    test: |
      system "#{bin}/agentctl", "--version"

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
```

### Testing Patterns

**Table-driven tests** are idiomatic Go:

```go
func TestWorkspaceCreate(t *testing.T) {
    tests := []struct {
        name    string
        args    []string
        wantErr bool
    }{
        {"valid branch", []string{"feat/new"}, false},
        {"empty name", []string{""}, true},
        {"invalid chars", []string{"feat@bad"}, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := createWorkspace(tt.args[0])
            if (err != nil) != tt.wantErr {
                t.Errorf("createWorkspace() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**CLI integration testing** with testscript (from Go team):
- Golden files for expected output
- Script-based test scenarios
- Used by Go toolchain itself

**Parallel tests**: Use `t.Parallel()` with local variable capture:
```go
for _, tt := range tests {
    tt := tt // capture range variable
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        // test code
    })
}
```

## Recommended Directory Structure for agentctl

Based on research, here's the recommended structure:

```
agentctl/
├── cmd/
│   └── agentctl/
│       └── main.go                 # Entry point (~20 lines)
├── internal/
│   ├── cli/                        # Cobra command layer
│   │   ├── root.go                 # Root cmd, global flags, Execute()
│   │   ├── workspace.go            # workspace create/list/show/delete
│   │   ├── hook.go                 # hook post-edit/post-write/etc
│   │   └── init.go                 # init command
│   ├── workspace/                  # Workspace business logic
│   │   ├── workspace.go            # Core workspace operations
│   │   ├── worktree.go            # Git worktree management
│   │   └── workspace_test.go
│   ├── hook/                       # Hook implementations
│   │   ├── context.go              # Context injection
│   │   ├── autocommit.go           # Auto-commit logic
│   │   └── hook_test.go
│   ├── git/                        # Git operations
│   │   ├── git.go                  # Git wrapper functions
│   │   └── git_test.go
│   └── config/                     # Configuration
│       ├── config.go
│       └── config_test.go
├── .goreleaser.yaml                # Release automation
├── Taskfile.yml                    # Build tasks
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

**Key design decisions:**

1. **Single binary**: One `cmd/agentctl/` directory
2. **CLI separation**: `internal/cli/` holds Cobra commands, thin layer calling into business logic
3. **Feature packages**: `internal/workspace/`, `internal/hook/` contain actual logic
4. **Shared git operations**: `internal/git/` centralizes git interactions
5. **No `/pkg`**: All code is internal since this is an end-user tool, not a library
6. **Taskfile over Makefile**: Modern, readable, checksum-based

## Sources

- [golang-standards/project-layout](https://github.com/golang-standards/project-layout)
- [Go Modules Layout - Official Docs](https://go.dev/doc/modules/layout)
- [Structuring Go Code for CLI Applications](https://www.bytesizego.com/blog/structure-go-cli-app)
- [spf13/cobra](https://github.com/spf13/cobra)
- [Just Make a Task - Applied Go](https://appliedgo.net/spotlight/just-make-a-task/)
- [GoReleaser Homebrew Casks](https://goreleaser.com/customization/homebrew_casks/)
- [Cross-compiling made easy with Golang](https://opensource.com/article/21/1/go-cross-compiling)
- [Writing Integration Tests for Go CLI](https://lucapette.me/writing/writing-integration-tests-for-a-go-cli-application/)
- [Dave Cheney: Prefer table driven tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [11 Tips for Structuring Go Projects](https://www.alexedwards.net/blog/11-tips-for-structuring-your-go-projects)
