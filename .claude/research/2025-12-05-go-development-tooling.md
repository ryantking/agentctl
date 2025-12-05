# Research: Modern Go Development Tooling (2024/2025)
Date: 2025-12-05
Focus: Linters, formatters, build tools, dependency management, and CI/CD for Go CLI projects
Agent: researcher

## Summary

Modern Go development in 2024/2025 centers around golangci-lint v2 for linting (with 70+ linters), gofumpt as the de-facto formatter, Task or Just as Make alternatives, GoReleaser for releases with ldflags version injection, and govulncheck for security scanning. GitHub Actions provides robust CI/CD with matrix testing.

## Key Findings

- **golangci-lint v2** released in 2025 with new configuration format, `linters.default` replacing `enable-all/disable-all` [Source](https://golangci-lint.run/docs/configuration/)
- **gofumpt** is now the de-facto standard due to gopls integration, stricter than gofmt [Source](https://github.com/mvdan/gofumpt)
- **Task (go-task)** preferred over Make for Go projects due to YAML syntax and cross-platform support [Source](https://github.com/go-task/task)
- **govulncheck** analyzes actual code paths, not just dependency manifests [Source](https://go.dev/doc/tutorial/govulncheck)
- **GoReleaser** auto-injects `main.version`, `main.commit`, `main.date` via ldflags [Source](https://goreleaser.com/cookbooks/using-main.version/)

## Detailed Analysis

### 1. Linters and Static Analysis

#### golangci-lint v2 Configuration

The golangci-lint v2 release (2025) introduced significant changes:

**Key Changes from v1 to v2:**
- `enable-all` and `disable-all` replaced with `linters.default: all|standard|fast|none`
- New `golangci-lint fmt` command for formatting
- New `golangci-lint migrate` command to convert v1 configs
- Exclusion presets: `comments`, `std-error-handling`, `common-false-positives`, `legacy`

**Recommended .golangci.yml for CLI Projects:**

```yaml
version: "2"

formatters:
  enable:
    - goimports
    - gofumpt

linters:
  default: standard
  enable:
    # Security
    - gosec          # Security issues
    - bodyclose      # HTTP response body closure

    # Bugs
    - govet          # Suspicious constructs
    - staticcheck    # Comprehensive static analysis
    - errcheck       # Unchecked errors
    - ineffassign    # Useless assignments
    - typecheck      # Type checking

    # Style & Complexity
    - gocritic       # Opinionated linter
    - gocyclo        # Cyclomatic complexity
    - gocognit       # Cognitive complexity
    - funlen         # Function length
    - lll            # Line length

    # Code Quality
    - dupl           # Code duplication
    - unconvert      # Unnecessary conversions
    - unparam        # Unused parameters
    - nakedret       # Naked returns
    - prealloc       # Preallocate slices

    # Modernization
    - modernize      # Code modernization (new in v2)

    # Error Handling
    - err113         # Error wrapping (formerly goerr113)
    - wrapcheck      # Error wrapping from external packages

    # Imports
    - goimports      # Import formatting
    - depguard       # Dependency guards

  disable:
    - varnamelen     # Too noisy for CLI tools

  settings:
    gocyclo:
      min-complexity: 15
    gocognit:
      min-complexity: 20
    funlen:
      lines: 100
      statements: 50
    lll:
      line-length: 120
    gocritic:
      enabled-checks:
        - appendAssign
        - argOrder
        - badCall
        - badCond
        - captLocal
        - caseOrder
        - dupArg
        - dupBranchBody
        - dupCase
        - exitAfterDefer
        - sloppyLen

issues:
  exclude-dirs:
    - vendor
    - testdata
  exclude-files:
    - ".*_test\\.go$"
  exclude-rules:
    - path: "_test\\.go"
      linters:
        - errcheck
        - gosec
        - funlen
        - dupl

output:
  formats:
    - format: colored-line-number
  sort-results: true
```

**IDE Integration:**
- VSCode: Install Go extension, enable `go.lintTool: golangci-lint`
- GoLand: Built-in golangci-lint support via Settings > Tools > golangci-lint
- Neovim: Use nvim-lint with golangci-lint backend

### 2. Formatters

#### Comparison Table

| Tool | What It Does | When to Use |
|------|-------------|-------------|
| **gofmt** | Standard Go formatter | Legacy, basic formatting only |
| **goimports** | gofmt + import management | When you need auto-import fixes |
| **gofumpt** | Stricter gofmt superset | **Recommended default** - integrates with gopls |

#### gofumpt Rules
- No empty lines at function start/end
- No empty lines around lone statements
- No empty lines before simple error checks
- Consistent composite literal newlines

#### Configuration with gopls (VSCode settings.json):
```json
{
  "gopls": {
    "formatting.gofumpt": true
  }
}
```

### 3. Build Tools

#### Comparison: Task vs Make vs Just

| Feature | Make | Task | Just |
|---------|------|------|------|
| Syntax | Custom DSL | YAML | Makefile-like |
| Dependency checking | Timestamps | Checksums | None |
| Cross-platform | Limited | Excellent | Good |
| Pre-installed | Yes (Unix) | No | No |
| Best for | Build systems | Task automation | Command running |

#### Recommended: Task (Taskfile.yml)

```yaml
version: '3'

vars:
  BINARY_NAME: myapp
  VERSION:
    sh: git describe --tags --always --dirty
  COMMIT:
    sh: git rev-parse --short HEAD
  DATE:
    sh: date -u '+%Y-%m-%dT%H:%M:%SZ'
  LDFLAGS: >-
    -s -w
    -X main.version={{.VERSION}}
    -X main.commit={{.COMMIT}}
    -X main.date={{.DATE}}

tasks:
  default:
    deps: [build]

  deps:
    desc: Install dependencies
    cmds:
      - go mod download
      - go mod tidy

  build:
    desc: Build the binary
    cmds:
      - go build -ldflags "{{.LDFLAGS}}" -o {{.BINARY_NAME}} ./cmd/{{.BINARY_NAME}}

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
    cmds:
      - task: fmt
      - task: lint
      - task: test
      - task: vuln

  clean:
    desc: Clean build artifacts
    cmds:
      - rm -f {{.BINARY_NAME}}
      - rm -f coverage.out

  release:
    desc: Create a release with goreleaser
    cmds:
      - goreleaser release --clean
```

### 4. GoReleaser Configuration

**.goreleaser.yaml for CLI projects:**

```yaml
version: 2

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - id: myapp
    main: ./cmd/myapp
    binary: myapp
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    # Default ldflags inject main.version, main.commit, main.date
    # Custom ldflags if version is in another package:
    ldflags:
      - -s -w
      - -X github.com/org/myapp/internal/version.Version={{.Version}}
      - -X github.com/org/myapp/internal/version.Commit={{.Commit}}
      - -X github.com/org/myapp/internal/version.Date={{.Date}}
    mod_timestamp: "{{ .CommitTimestamp }}"
    flags:
      - -trimpath

archives:
  - id: default
    formats:
      - tar.gz
    format_overrides:
      - goos: windows
        formats:
          - zip
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^chore:'

release:
  github:
    owner: myorg
    name: myapp
  draft: false
  prerelease: auto
```

**Version package pattern (internal/version/version.go):**

```go
package version

var (
    Version = "dev"
    Commit  = "none"
    Date    = "unknown"
)

func Full() string {
    return Version + " (" + Commit + ") built " + Date
}
```

### 5. Dependency Management

#### Go Modules Best Practices

**go.mod management:**
```bash
# Initialize module
go mod init github.com/org/myapp

# Tidy dependencies
go mod tidy

# Verify checksums
go mod verify

# Download dependencies
go mod download

# View dependency graph
go mod graph
```

#### Private Module Handling

**Set GOPRIVATE:**
```bash
# Single repo
go env -w GOPRIVATE=github.com/myorg/private-repo

# Entire organization
go env -w GOPRIVATE=github.com/myorg/*

# Multiple patterns
go env -w GOPRIVATE=github.com/myorg/*,gitlab.com/mycompany/*
```

**Authentication methods:**

1. **Git URL rewriting (recommended for CI):**
```bash
git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"
```

2. **.netrc file (local development):**
```
machine github.com
  login USERNAME
  password PERSONAL_ACCESS_TOKEN
```

#### govulncheck Security Scanning

```bash
# Install
go install golang.org/x/vuln/cmd/govulncheck@latest

# Scan current module
govulncheck ./...

# Scan binary
govulncheck -mode=binary ./myapp

# JSON output for CI
govulncheck -json ./...
```

**Key advantage:** govulncheck only reports vulnerabilities in code paths you actually use, not just dependencies in go.mod.

### 6. CI/CD with GitHub Actions

**.github/workflows/ci.yml:**

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

env:
  GO_VERSION: '1.23'

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest

  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Run tests
        run: go test -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: coverage.out

  vuln:
    name: Vulnerability Check
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Install govulncheck
        run: go install golang.org/x/vuln/cmd/govulncheck@latest

      - name: Run govulncheck
        run: govulncheck ./...

  build:
    name: Build
    runs-on: ubuntu-latest
    needs: [lint, test, vuln]
    strategy:
      matrix:
        goos: [linux, darwin, windows]
        goarch: [amd64, arm64]
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Build
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: |
          go build -ldflags="-s -w" -o myapp-${{ matrix.goos }}-${{ matrix.goarch }} ./cmd/myapp
```

**.github/workflows/release.yml:**

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Tool Recommendations Summary

| Category | Recommended Tool | Alternative |
|----------|-----------------|-------------|
| **Linting** | golangci-lint v2 | staticcheck standalone |
| **Formatting** | gofumpt + goimports | gofmt (standard only) |
| **Task Runner** | Task (Taskfile) | Just, Make |
| **Releases** | GoReleaser | Manual scripts |
| **Security** | govulncheck | trivy, grype |
| **CI/CD** | GitHub Actions | GitLab CI, CircleCI |

## Sources

- [golangci-lint Documentation](https://golangci-lint.run/docs/configuration/)
- [golangci-lint Linters List](https://golangci-lint.run/docs/linters/)
- [Golden Config for golangci-lint](https://gist.github.com/maratori/47a4d00457a92aa426dbd48a18776322)
- [gofumpt GitHub](https://github.com/mvdan/gofumpt)
- [Task (go-task) GitHub](https://github.com/go-task/task)
- [Task vs Make Comparison](https://appliedgo.net/spotlight/just-make-a-task/)
- [GoReleaser Documentation](https://goreleaser.com/customization/builds/go/)
- [GoReleaser ldflags Cookbook](https://goreleaser.com/cookbooks/using-main.version/)
- [govulncheck Tutorial](https://go.dev/doc/tutorial/govulncheck)
- [Go Vulnerability Management](https://go.dev/blog/vuln)
- [GitHub Actions Go Documentation](https://docs.github.com/en/actions/use-cases-and-examples/building-and-testing/building-and-testing-go)
- [actions/setup-go](https://github.com/actions/setup-go)
- [golangci-lint-action](https://github.com/golangci/golangci-lint-action)
- [Go Modules Reference](https://go.dev/ref/mod)
- [Private Go Modules Guide](https://www.digitalocean.com/community/tutorials/how-to-use-a-private-go-module-in-your-own-project)
