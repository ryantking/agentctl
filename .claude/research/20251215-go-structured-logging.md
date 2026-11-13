# Research: Go Structured Logging Libraries 2025
Date: 2025-12-15
Focus: Best Go structured logging libraries comparison - slog, zerolog, zap, logrus
Agent: researcher

## Summary

For new Go projects in 2025, **slog** (Go 1.21+ stdlib) is the recommended default choice due to standard library stability and ecosystem unification. For performance-critical applications, **zerolog** offers the best raw performance with zero allocations. **Zap** remains excellent for enterprise applications requiring deep customization. **Logrus is in maintenance mode** and should only be used for existing projects - new projects should avoid it.

## Key Findings

- **slog is now the idiomatic choice** for new Go projects since Go 1.21, providing structured logging in the standard library with pluggable backends [Go Blog](https://go.dev/blog/slog)
- **zerolog is the fastest** (~30 ns/op, 0 allocs) followed by zap (~71 ns/op), then slog (~174 ns/op) [Better Stack Benchmarks](https://betterstack.com/community/guides/logging/best-golang-logging-libraries/)
- **logrus is in maintenance mode** - maintainers explicitly recommend zerolog, zap, or apex for new projects [Logrus GitHub](https://github.com/sirupsen/logrus)
- **12-factor apps should log to stdout** unbuffered, letting the execution environment handle routing [12factor.net](https://12factor.net/logs)
- **Major projects migrating to slog**: containerd, Cilium, Teleport, k6, GitLab are all investigating or implementing logrus-to-slog migrations

## Detailed Analysis

### Library Comparison Matrix

| Library | Performance | Allocations | Status | Best For |
|---------|-------------|-------------|--------|----------|
| **slog** | 174 ns/op | 0 allocs/op | Active (stdlib) | New projects, ecosystem unity |
| **zerolog** | 30 ns/op | 0 allocs/op | Active | High-throughput systems |
| **zap** | 71 ns/op | 0 allocs/op | Active | Enterprise, customization |
| **logrus** | 2,231 ns/op | 23 allocs/op | Maintenance | Legacy only |

### slog (Go 1.21+ Standard Library)

**Recommendation: Default choice for new projects in 2025**

The Go team added slog to provide a common framework for structured logging, addressing ecosystem fragmentation where popular packages like logrus were used in 100,000+ projects.

**Design Principles:**
- Simplicity-first API that doesn't sacrifice performance
- Non-replacement: not intended to replace existing packages
- Interoperability: Frontend Logger/backend Handler architecture allows packages to share interfaces
- 95% of logging calls pass 5 or fewer attributes (optimized for this)

**Code Example - Basic Usage:**
```go
import "log/slog"

// Simple logging
slog.Info("hello, world")
slog.Info("hello, world", "user", os.Getenv("USER"))

// With JSON output
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
logger.Info("request processed", "method", "GET", "status", 200)

// High-performance with typed attributes
slog.LogAttrs(context.Background(), slog.LevelInfo, "hello",
    slog.String("user", os.Getenv("USER")))
```

**Error Context Logging:**
```go
logger.Error("database connection failed",
    slog.Any("error", err),
    slog.String("host", "db.example.com"),
    slog.Int("retry_count", 3),
)
```

**Pros:**
- Standard library stability and maintenance
- Pluggable Handler backends
- Context-aware logging support
- Zero external dependencies

**Cons:**
- Slower than zerolog/zap
- More verbose API for typed fields

### zerolog

**Recommendation: Performance-critical applications**

Currently the fastest structured logging framework for Go with zero-allocation JSON logging.

**Code Example:**
```go
logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

// Structured logging with error context
logger.Error().
    Err(err).
    Str("host", "db.example.com").
    Int("retry_count", 3).
    Msg("database connection failed")

// CLI error context
logger.Error().
    Str("command", "deploy").
    Int("exit_code", 1).
    Str("stderr", errorOutput).
    Strs("args", args).
    Msg("command execution failed")
```

**Pros:**
- Exceptional performance (30 ns/op)
- Zero allocations under normal use
- Clean fluent API
- Context-aware via `context.Context`

**Cons:**
- Limited format options (JSON/CBOR only)
- Less customizable than zap

### zap (Uber)

**Recommendation: Enterprise applications requiring deep customization**

Pioneered minimal-allocation logging at Uber. Provides both typed `Logger` (fast) and `SugaredLogger` (flexible) APIs.

**Code Example:**
```go
// Strongly-typed (faster)
logger.Info("user signed in",
    zap.Int("userid", 123456),
    zap.String("provider", "facebook"),
)

// Sugared (more flexible)
sugar.Infow("user signed in", "userid", 123456, "provider", "facebook")

// Error context
logger.Error("command failed",
    zap.Error(err),
    zap.String("command", cmd),
    zap.Int("exit_code", exitCode),
    zap.String("stderr", stderr),
)
```

**Pros:**
- Exceptional customization via zapcore
- Dual API approaches (typed/sugared)
- Production-proven at Uber scale
- slog integration available (zapslog)

**Cons:**
- Setup complexity
- Less intuitive initial configuration

### logrus

**Recommendation: DO NOT use for new projects**

The maintainer explicitly states: "Logrus is in maintenance-mode. We will not be introducing new features." Recommended alternatives are zerolog, zap, and apex.

**When still acceptable:**
- Existing projects already using it effectively
- Projects prioritizing backward compatibility over features

**Migration Path:**
- Use [sloghandler](https://github.com/niondir/sloghandler) to bridge logrus to slog during migration
- Major projects (containerd, Cilium, k6) are actively migrating away

## CLI Application Patterns

### stdout vs stderr Convention

**12-Factor Recommendation:** Applications should write logs to stdout unbuffered, letting the execution environment handle routing.

**Unix Philosophy:** Reserve stdout for program output, stderr for diagnostic information.

**Go Standard Library Default:** `log` package writes to stderr by default.

**Practical Pattern for CLIs:**
```go
// Create separate loggers for different outputs
infoLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))
errorLogger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
    Level: slog.LevelError,
}))

// Or use a custom handler that routes by level
type LevelRouter struct {
    stdout slog.Handler
    stderr slog.Handler
}

func (r *LevelRouter) Handle(ctx context.Context, rec slog.Record) error {
    if rec.Level >= slog.LevelError {
        return r.stderr.Handle(ctx, rec)
    }
    return r.stdout.Handle(ctx, rec)
}
```

### Logging Error Context for CLI Tools

**Essential fields for command execution errors:**
```go
logger.LogAttrs(ctx, slog.LevelError, "command execution failed",
    slog.String("command", cmdName),
    slog.Int("exit_code", exitCode),
    slog.String("stderr", stderrOutput),
    slog.Any("args", cmdArgs),
    slog.Duration("duration", elapsed),
    slog.String("working_dir", cwd),
)
```

**For wrapped errors with stack traces:**
```go
// Use xerrors or similar for stack traces
func fmtErr(err error) slog.Value {
    var groupValues []slog.Attr
    groupValues = append(groupValues, slog.String("msg", err.Error()))

    frames := marshalStack(err) // Extract stack frames
    if frames != nil {
        groupValues = append(groupValues, slog.Any("trace", frames))
    }
    return slog.GroupValue(groupValues...)
}

logger.Error("operation failed", slog.Any("error", fmtErr(err)))
```

## Applicable Patterns for agentctl

Given agentctl is a CLI tool for managing Claude Code configurations:

1. **Use slog** as the logging library - it's the idiomatic choice for 2025 Go projects
2. **Route by level**: Info/Debug to stdout, Warn/Error to stderr
3. **Structure error context**:
   ```go
   logger.Error("workspace creation failed",
       slog.String("workspace", name),
       slog.String("branch", branchName),
       slog.Any("error", err),
       slog.String("git_stderr", gitStderr),
   )
   ```
4. **Use TextHandler for development**, JSONHandler if logs will be aggregated
5. **Consider zerolog** if profiling shows logging as a bottleneck

## Sources

- [Structured Logging with slog - Go Blog (Official)](https://go.dev/blog/slog)
- [log/slog Package Documentation](https://pkg.go.dev/log/slog)
- [Logging in Go: A Comparison of Top Libraries - Better Stack](https://betterstack.com/community/guides/logging/best-golang-logging-libraries/)
- [Go Logging Benchmarks Repository](https://github.com/betterstack-community/go-logging-benchmarks)
- [Logging in Go with Slog: Ultimate Guide - Better Stack](https://betterstack.com/community/guides/logging/logging-in-go/)
- [Logrus GitHub - Maintenance Mode Announcement](https://github.com/sirupsen/logrus)
- [12-Factor App: Logs](https://12factor.net/logs)
- [Best Go Logging Tools in 2025 - Dash0](https://www.dash0.com/faq/best-go-logging-tools-in-2025-a-comprehensive-guide)
- [containerd: Consider Replacing Logrus](https://github.com/containerd/containerd/issues/8280)
- [Cilium: Migration to slog](https://github.com/cilium/cilium/pull/38409)

## Confidence Level

**High** - Based on official Go team documentation, maintained benchmark repositories with weekly updates, explicit maintainer statements (logrus), and consistent recommendations across multiple authoritative sources. The slog recommendation is particularly well-supported given its inclusion in the standard library and the Go team's explicit design goals.

## Related Questions

- How to implement custom slog handlers for specific output formats?
- What's the best way to integrate slog with OpenTelemetry for distributed tracing?
- How do popular CLIs (kubectl, docker, gh) handle their logging?
- What are the performance implications of logging in hot paths vs cold paths?
