# Research: Go CLI Error Handling Patterns for Subprocess Execution
Date: 2025-12-15
Focus: How popular Go CLI tools handle errors, subprocess execution, exit codes, and stderr presentation
Agent: researcher

## Summary

Popular Go CLI tools like kubectl, Docker CLI, and gh use a layered error handling approach with custom error types that encode exit codes, stderr capture patterns, and context-aware error wrapping. The dominant patterns include Cobra's RunE for error propagation, custom StatusError types for exit code preservation, and explicit stderr buffer management for subprocess execution.

## Key Findings

1. **Cobra's RunE over Run** - Always use RunE to return errors rather than handling them inline ([Cobra User Guide](https://github.com/spf13/cobra/blob/main/site/content/user_guide.md))
2. **Custom StatusError types** - Docker CLI and kubectl use dedicated error types that embed exit codes ([Docker CLI](https://github.com/docker/cli/blob/master/cmd/docker/docker.go))
3. **Explicit stderr capture** - Go's exec.ExitError only auto-populates Stderr from Output(), requiring manual buffer setup ([Go Issue #11381](https://github.com/golang/go/issues/11381))
4. **Semantic exit codes** - Square's exit package defines ranges: 80-99 for user errors, 100-119 for software errors ([square/exit](https://pkg.go.dev/github.com/square/exit))
5. **gh CLI error taxonomy** - FlagError, SilentError, CancelError, NoResultsError for different failure modes ([gh cmdutil](https://pkg.go.dev/github.com/cli/cli/v2/pkg/cmdutil))

## Detailed Analysis

### 1. Cobra Framework Error Handling

#### RunE vs Run
```go
// Preferred: RunE allows error propagation
var cmd = &cobra.Command{
    Use: "example",
    RunE: func(cmd *cobra.Command, args []string) error {
        if err := someOperation(); err != nil {
            return fmt.Errorf("operation failed: %w", err)
        }
        return nil
    },
}

// Avoid: Run swallows errors
var cmd = &cobra.Command{
    Run: func(cmd *cobra.Command, args []string) {
        // Must handle errors inline, less clean
    },
}
```

#### SilenceUsage and SilenceErrors
```go
// Prevent usage output on runtime errors (not flag errors)
cmd.SilenceUsage = true

// Prevent Cobra from printing errors (handle yourself)
cmd.SilenceErrors = true
```

**Best Practice**: Set `SilenceUsage = true` on the root command when using RunE, so runtime errors don't print usage information. Keep `SilenceErrors = false` to let Cobra handle error output unless you want custom formatting.

#### CheckErr Utility
```go
// cobra.CheckErr logs error and exits
func Execute() {
    err := rootCmd.Execute()
    cobra.CheckErr(err)
}
```

Sources: [Cobra User Guide](https://github.com/spf13/cobra/blob/main/site/content/user_guide.md), [JetBrains Error Handling Guide](https://www.jetbrains.com/guide/go/tutorials/cli-apps-go-cobra/error_handling/)

### 2. Subprocess Execution Patterns

#### Basic exec.Command Error Handling
```go
import (
    "bytes"
    "fmt"
    "os/exec"
)

func runCommand(name string, args ...string) error {
    var stdout, stderr bytes.Buffer
    cmd := exec.Command(name, args...)
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()
    if err != nil {
        // Check for exit error with code
        if exitErr, ok := err.(*exec.ExitError); ok {
            return fmt.Errorf("%s failed (exit %d): %s",
                name, exitErr.ExitCode(), stderr.String())
        }
        // Other errors (command not found, permission denied, etc.)
        return fmt.Errorf("%s failed: %w", name, err)
    }
    return nil
}
```

#### Extracting Exit Code
```go
import (
    "errors"
    "os/exec"
    "syscall"
)

func getExitCode(err error) int {
    var exitErr *exec.ExitError
    if errors.As(err, &exitErr) {
        // Go 1.12+: Use ExitCode() method
        return exitErr.ExitCode()

        // Alternative (older pattern):
        // if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok {
        //     return ws.ExitStatus()
        // }
    }
    return -1 // Unknown/not an exit error
}
```

#### Context-Aware Execution
```go
func runWithTimeout(ctx context.Context, name string, args ...string) error {
    cmd := exec.CommandContext(ctx, name, args...)
    err := cmd.Run()

    if err != nil {
        // Check if context caused the error
        if ctx.Err() == context.DeadlineExceeded {
            return fmt.Errorf("%s timed out: %w", name, err)
        }
        if ctx.Err() == context.Canceled {
            return fmt.Errorf("%s canceled: %w", name, err)
        }
        return err
    }
    return nil
}
```

Sources: [DoltHub os/exec Patterns](https://www.dolthub.com/blog/2022-11-28-go-os-exec-patterns/)

### 3. kubectl Error Handling Patterns

kubectl uses a centralized `CheckErr` function in `pkg/cmd/util/helpers.go`:

```go
const DefaultErrorExitCode = 1

func CheckErr(err error) {
    checkErr(err, fatalErrHandler)
}

func checkErr(err error, handleErr func(string, int)) {
    switch {
    case err == nil:
        return
    case err == ErrExit:
        handleErr("", DefaultErrorExitCode)
    case errors.As(err, &exec.ExitError{}):
        // Preserve original exit status from subprocess
        handleErr("", exitErr.ExitCode())
    case errors.As(err, &StatusError{}):
        handleErr(err.Error(), statusErr.Status().Code)
    default:
        handleErr(StandardErrorMessage(err), DefaultErrorExitCode)
    }
}
```

**Key patterns**:
- `ErrExit` sentinel for silent exits
- `StatusError` for Kubernetes API errors with codes
- `exec.ExitError` passthrough for subprocess exit codes
- `BehaviorOnFatal` allows test overrides

Sources: [kubectl helpers.go](https://github.com/kubernetes/kubectl/blob/master/pkg/cmd/util/helpers.go)

### 4. Docker CLI Error Handling

Docker CLI uses a `StatusError` type and sophisticated exit code mapping:

```go
// cli/command/cli.go
type StatusError struct {
    Status     string
    StatusCode int
}

func (e StatusError) Error() string {
    return e.Status
}

// cmd/docker/docker.go
func getExitCode(err error) int {
    if err == nil {
        return 0
    }

    // Check for StatusError with explicit code
    var statusErr cli.StatusError
    if errors.As(err, &statusErr) && statusErr.StatusCode != 0 {
        return statusErr.StatusCode
    }

    // Check for signal termination
    if sigErr, ok := err.(errCtxSignalTerminated); ok {
        // Unix convention: 128 + signal number
        return 128 + int(sigErr.signal)
    }

    // Check for subprocess exit code
    var exitErr *exec.ExitError
    if errors.As(err, &exitErr) {
        if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok {
            return ws.ExitStatus()
        }
    }

    return 1 // Default error exit
}
```

**Key patterns**:
- StatusError for API/container errors with codes
- Signal-terminated errors map to 128 + signal
- Subprocess exit codes preserved via exec.ExitError
- -1 indicates daemon connection lost

Sources: [Docker CLI docker.go](https://github.com/docker/cli/blob/master/cmd/docker/docker.go), [Docker CLI exec.go](https://github.com/docker/cli/blob/master/cli/command/container/exec.go)

### 5. GitHub CLI (gh) Error Taxonomy

gh CLI defines distinct error types for different failure modes:

```go
// pkg/cmdutil/errors.go

// FlagError triggers usage display
type FlagError struct {
    err error
}

func FlagErrorf(format string, args ...interface{}) error {
    return &FlagError{err: fmt.Errorf(format, args...)}
}

func (e *FlagError) Unwrap() error { return e.err }

// SilentError exits without message
var SilentError = errors.New("SilentError")

// CancelError indicates user cancellation
var CancelError = errors.New("CancelError")

// NoResultsError for empty queries
type NoResultsError struct {
    message string
}

func NewNoResultsError(msg string) NoResultsError {
    return NoResultsError{message: msg}
}

// Helper functions
func IsUserCancellation(err error) bool {
    return errors.Is(err, CancelError)
}
```

**Usage in commands**:
```go
func runCommand(opts *Options) error {
    // Flag validation errors show usage
    if opts.Interactive && opts.Batch {
        return cmdutil.FlagErrorf("cannot use both --interactive and --batch")
    }

    // User cancellation exits silently
    result, err := prompter.Confirm("Continue?")
    if err != nil {
        return cmdutil.CancelError
    }

    // Empty results are distinct from errors
    items, err := api.List()
    if err != nil {
        return err
    }
    if len(items) == 0 {
        return cmdutil.NewNoResultsError("no items found")
    }

    // Silent exit when error already displayed
    if processErr != nil {
        fmt.Fprintln(opts.ErrOut, processErr)
        return cmdutil.SilentError
    }

    return nil
}
```

Sources: [gh cmdutil](https://pkg.go.dev/github.com/cli/cli/v2/pkg/cmdutil), [go-gh](https://github.com/cli/go-gh/blob/trunk/gh.go)

### 6. go-gh Subprocess Execution

The go-gh library wraps gh CLI invocation:

```go
// gh.go
func Exec(args ...string) (stdout, stderr bytes.Buffer, err error) {
    return ExecContext(context.Background(), args...)
}

func ExecContext(ctx context.Context, args ...string) (stdout, stderr bytes.Buffer, err error) {
    ghExe, err := Path()
    if err != nil {
        return
    }
    return run(ctx, ghExe, nil, args...)
}

func run(ctx context.Context, ghExe string, env []string, args ...string) (stdout, stderr bytes.Buffer, err error) {
    cmd := exec.CommandContext(ctx, ghExe, args...)
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    if env != nil {
        cmd.Env = env
    }

    err = cmd.Run()
    if err != nil {
        err = fmt.Errorf("gh execution failed: %w", err)
    }
    return
}
```

**Key pattern**: Returns both stdout and stderr buffers alongside error, allowing caller to inspect output even on failure.

Sources: [go-gh gh.go](https://github.com/cli/go-gh/blob/trunk/gh.go)

### 7. terraform-exec Error Handling

Terraform-exec defines custom error types for specific failure conditions:

```go
// tfexec/errors.go

type ErrNoSuitableBinary struct {
    err error
}

func (e *ErrNoSuitableBinary) Unwrap() error { return e.err }

type ErrVersionMismatch struct {
    MinInclusive string
    MaxExclusive string
    Actual       string
}

type ErrManualEnvVar struct {
    Name string
}

// Internal error for context handling (golang/go#21880)
type cmdErr struct {
    err error
}

func (e cmdErr) Is(target error) bool {
    switch target {
    case context.Canceled, context.DeadlineExceeded:
        return errors.Is(e.err, target)
    }
    return false
}

func (e cmdErr) Unwrap() error { return e.err }
```

**Key patterns**:
- Structured error types with relevant fields (version ranges, env var names)
- `Is()` implementation for context error checking through wrapped errors
- Clear separation of user errors vs system errors

Sources: [terraform-exec Error Handling](https://deepwiki.com/hashicorp/terraform-exec/2.1-error-handling)

### 8. Semantic Exit Codes (Square)

The square/exit package defines meaningful exit code ranges:

```go
// exit.go
const (
    OK                = 0   // Success
    NotOK             = 1   // Generic failure

    // User errors: 80-99
    UsageError        = 80  // Wrong args/flags
    UnknownSubcommand = 81  // Bad subcommand
    RequirementNotMet = 82  // Missing prerequisite
    Forbidden         = 83  // Not authorized
    MovedPermanently  = 84  // Tool relocated

    // Software errors: 100-119
    InternalError     = 100 // Bug in program
    Unavailable       = 101 // Dependency down
)

type Error struct {
    Code  Code
    Cause error
}

func (e Error) ExitCode() int { return int(e.Code) }
func (e Error) Unwrap() error { return e.Cause }

// Helpers
func IsUserError(code Code) bool     { return code >= 80 && code < 100 }
func IsSoftwareError(code Code) bool { return code >= 100 && code < 120 }
func IsSignal(code Code) bool        { return code >= 128 || code == -1 }
func FromSignal(s syscall.Signal) Code { return Code(128 + int(s)) }
```

**Usage**:
```go
func main() {
    if err := run(); err != nil {
        var exitErr exit.Error
        if errors.As(err, &exitErr) {
            fmt.Fprintln(os.Stderr, err)
            os.Exit(exitErr.ExitCode())
        }
        fmt.Fprintln(os.Stderr, err)
        os.Exit(exit.NotOK)
    }
}
```

Sources: [square/exit](https://pkg.go.dev/github.com/square/exit)

## Applicable Patterns

For a Go CLI tool executing agent subprocesses:

### 1. Define Custom Error Types
```go
// errors.go
package agentctl

import (
    "errors"
    "fmt"
)

// AgentError wraps subprocess execution failures
type AgentError struct {
    Agent    string
    ExitCode int
    Stderr   string
    Err      error
}

func (e *AgentError) Error() string {
    if e.Stderr != "" {
        return fmt.Sprintf("agent %s failed (exit %d): %s", e.Agent, e.ExitCode, e.Stderr)
    }
    return fmt.Sprintf("agent %s failed (exit %d)", e.Agent, e.ExitCode)
}

func (e *AgentError) Unwrap() error { return e.Err }

// Silent exit (error already displayed)
var ErrSilent = errors.New("silent exit")

// User cancelled operation
var ErrCanceled = errors.New("operation canceled")
```

### 2. Subprocess Execution Helper
```go
// exec.go
func RunAgent(ctx context.Context, agent string, args ...string) error {
    var stdout, stderr bytes.Buffer
    cmd := exec.CommandContext(ctx, agent, args...)
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()
    if err != nil {
        exitCode := -1
        var exitErr *exec.ExitError
        if errors.As(err, &exitErr) {
            exitCode = exitErr.ExitCode()
        }

        return &AgentError{
            Agent:    agent,
            ExitCode: exitCode,
            Stderr:   strings.TrimSpace(stderr.String()),
            Err:      err,
        }
    }
    return nil
}
```

### 3. Main Error Handler
```go
// main.go
func main() {
    err := rootCmd.Execute()
    if err != nil {
        code := handleError(err)
        os.Exit(code)
    }
}

func handleError(err error) int {
    // Silent exit
    if errors.Is(err, ErrSilent) {
        return 1
    }

    // User cancellation
    if errors.Is(err, ErrCanceled) {
        return 130 // Standard for Ctrl+C
    }

    // Agent errors preserve exit code
    var agentErr *AgentError
    if errors.As(err, &agentErr) {
        if agentErr.ExitCode > 0 {
            return agentErr.ExitCode
        }
        return 1
    }

    // Default error
    fmt.Fprintln(os.Stderr, "Error:", err)
    return 1
}
```

### 4. Stderr Presentation
```go
// Present stderr to users cleanly
func formatAgentError(err *AgentError) string {
    var b strings.Builder
    b.WriteString(fmt.Sprintf("Agent '%s' failed", err.Agent))

    if err.ExitCode > 0 {
        b.WriteString(fmt.Sprintf(" (exit code %d)", err.ExitCode))
    }
    b.WriteString("\n")

    if err.Stderr != "" {
        b.WriteString("\nOutput:\n")
        // Indent stderr for clarity
        for _, line := range strings.Split(err.Stderr, "\n") {
            b.WriteString("  ")
            b.WriteString(line)
            b.WriteString("\n")
        }
    }

    return b.String()
}
```

## Best Practices Summary

1. **Use RunE in Cobra** - Return errors instead of handling inline
2. **Set SilenceUsage** - Prevent usage on runtime errors
3. **Define error taxonomy** - FlagError, SilentError, CancelError, etc.
4. **Capture stderr explicitly** - Don't rely on exec.ExitError.Stderr auto-population
5. **Preserve exit codes** - Pass through subprocess exit codes when meaningful
6. **Use errors.As/Is** - Modern Go error inspection over type assertions
7. **Wrap with context** - Include agent name, operation in error messages
8. **Signal handling** - Map SIGINT to exit 130 (128 + 2)
9. **Semantic codes** - Consider ranges for user vs system errors

## Sources

- [Cobra User Guide](https://github.com/spf13/cobra/blob/main/site/content/user_guide.md)
- [Docker CLI docker.go](https://github.com/docker/cli/blob/master/cmd/docker/docker.go)
- [Docker CLI exec.go](https://github.com/docker/cli/blob/master/cli/command/container/exec.go)
- [kubectl helpers.go](https://github.com/kubernetes/kubectl/blob/master/pkg/cmd/util/helpers.go)
- [gh cmdutil](https://pkg.go.dev/github.com/cli/cli/v2/pkg/cmdutil)
- [go-gh](https://github.com/cli/go-gh/blob/trunk/gh.go)
- [terraform-exec Error Handling](https://deepwiki.com/hashicorp/terraform-exec/2.1-error-handling)
- [square/exit](https://pkg.go.dev/github.com/square/exit)
- [DoltHub os/exec Patterns](https://www.dolthub.com/blog/2022-11-28-go-os-exec-patterns/)
- [Go Issue #11381 - stderr in ExitError](https://github.com/golang/go/issues/11381)
- [JetBrains Cobra Error Handling](https://www.jetbrains.com/guide/go/tutorials/cli-apps-go-cobra/error_handling/)

## Confidence Level

**High** - Patterns derived from production CLI tools (kubectl, Docker, gh) with millions of users. Code examples extracted directly from GitHub repositories and official documentation.

## Related Questions

- How do these patterns interact with structured logging (slog)?
- What patterns exist for retry logic on transient subprocess failures?
- How should timeout errors be distinguished from other failures?
- What accessibility considerations exist for CLI error presentation?
