# Research: Go Error Handling Libraries and Patterns Beyond stdlib

Date: 2025-12-15
Focus: Error handling libraries, stack traces, structured logging, and production best practices
Agent: researcher

## Summary

The Go error handling ecosystem in 2025 recommends a "stdlib first" approach, using `errors.Is`, `errors.As`, and `%w` wrapping from the standard library for most cases. Third-party libraries like `cockroachdb/errors` provide compelling features for distributed systems (network portability, PII-safe reporting), while `pkg/errors` remains in maintenance mode. Stack traces should be used selectively due to performance costs, with alternatives like `errtrace` offering lighter-weight return traces.

## Key Findings

- **pkg/errors is in maintenance mode** - not accepting new functionality due to Go 1.13+ stdlib improvements [pkg.go.dev](https://pkg.go.dev/github.com/pkg/errors)
- **cockroachdb/errors** is the most feature-rich alternative with network portability and Sentry integration [GitHub](https://github.com/cockroachdb/errors)
- **errtrace** provides lightweight return traces (vs stack traces) with better performance [GitHub](https://github.com/bracesdev/errtrace)
- **Go 1.20** added `errors.Join` for wrapping multiple errors - no major errors package changes since [go.dev](https://go.dev/doc/go1.20)
- **slog** (Go 1.21+) is now the recommended structured logging solution for new projects [Better Stack](https://betterstack.com/community/guides/logging/logging-in-go/)

## Detailed Analysis

### 1. pkg/errors vs stdlib errors (Post Go 1.13)

**stdlib errors (recommended for most cases):**
- `errors.New()`, `fmt.Errorf()` with `%w` verb for wrapping
- `errors.Is()`, `errors.As()`, `errors.Unwrap()` for inspection
- `errors.Join()` (Go 1.20+) for multiple error wrapping
- No stack traces by default (intentional design choice)

**pkg/errors (maintenance mode):**
- Automatically captures stack traces with `errors.New()`, `errors.Wrap()`
- `%+v` formatting for stack trace output
- `errors.Cause()` for retrieving original error
- Compatible with Go 1.13+ via `Is()`, `As()`, `Unwrap()`
- Final v1.0 release planned, no new features

**When to use pkg/errors:**
- Legacy codebases already using it
- When stack traces are essential for debugging
- Performance is not critical

### 2. Stack Trace Approaches in Go

**Option A: pkg/errors (Legacy Standard)**
```go
import "github.com/pkg/errors"

err := errors.New("something failed") // includes stack trace
err = errors.Wrap(err, "context")     // adds context + stack
fmt.Printf("%+v", err)                // prints full stack
```

**Option B: cockroachdb/errors (Production-Grade)**
```go
import "github.com/cockroachdb/errors"

err := errors.New("failure")          // automatic stack trace
err = errors.Wrap(err, "context")     // additional context
fmt.Printf("%+v", err)                // verbose output with stack
```

**Option C: errtrace (Lightweight Alternative)**
```go
import "braces.dev/errtrace"

// Instead of full stack traces, captures return path
return errtrace.Wrap(err)             // wraps with return location
errtrace.FormatString(err)            // prints return trace
```

**Option D: go-errors/errors**
```go
import "github.com/go-errors/errors"

err := errors.New("failure")
err.ErrorStack()                      // returns stack as string
err.StackFrames()                     // returns []StackFrame
```

**Performance Comparison:**
| Library | Cost | Notes |
|---------|------|-------|
| stdlib errors.New | ~10 ns/op | No stack trace |
| pkg/errors | ~700 ns/op | Full stack capture |
| errtrace | Much faster | Incremental capture |

### 3. Structured Error Logging Patterns

**Recommended: slog (Go 1.21+)**
```go
import "log/slog"

logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
logger.Error("operation failed",
    "error", err,
    "user_id", userID,
    "request_id", reqID,
)
```

**High-Performance: zerolog**
```go
import "github.com/rs/zerolog"

log.Error().
    Err(err).
    Str("user_id", userID).
    Msg("operation failed")
```

**Enterprise: zap**
```go
import "go.uber.org/zap"

logger.Error("operation failed",
    zap.Error(err),
    zap.String("user_id", userID),
)
```

**Performance Benchmarks:**
| Library | ns/op | Allocations |
|---------|-------|-------------|
| zerolog | 380 | 1 |
| zap | 656 | 5 |
| slog | 2481 | 42 |

**Recommendation:** Use slog for new projects (stdlib, future-proof). Use zerolog for latency-sensitive applications. Use zap for enterprise features and ecosystem.

### 4. Error Handling in Production Systems

**Best Practices:**

1. **Categorize errors** - Distinguish transient (retry-worthy) from permanent errors
2. **Add context at boundaries** - Wrap errors when crossing package/service boundaries
3. **Use sentinel errors sparingly** - Prefer error types for rich information
4. **Consider API contracts** - Wrapping exposes internal errors to callers
5. **Distributed tracing** - Use trace_id in context for cross-service correlation

**Production Pattern:**
```go
// Define domain errors
var (
    ErrNotFound     = errors.New("resource not found")
    ErrUnauthorized = errors.New("unauthorized")
)

// Wrap with context at boundaries
func GetUser(id string) (*User, error) {
    user, err := db.FindUser(id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("get user %s: %w", id, err)
    }
    return user, nil
}

// Check errors at handling points
if errors.Is(err, ErrNotFound) {
    return http.StatusNotFound
}
```

### 5. When to Use Third-Party Packages vs stdlib

**Use stdlib when:**
- Simple error wrapping and inspection is sufficient
- Performance is critical (no stack trace overhead)
- Minimizing dependencies is a priority
- Building libraries (avoid forcing dependencies on users)

**Use cockroachdb/errors when:**
- Building distributed systems
- Need error portability across network
- Require Sentry integration with PII redaction
- Need rich error features (secondary causes, domains)

**Use pkg/errors when:**
- Legacy codebase already uses it
- Simple stack trace needs
- Gradual migration to stdlib planned

**Use errtrace when:**
- Want error traces without full stack overhead
- Performance-sensitive error handling
- Prefer return traces over stack traces

**Use go-errors/errors when:**
- Integration with bug tracking (Bugsnag)
- Need programmatic stack frame access

## Applicable Patterns for This Codebase

For a Go CLI tool like agentctl:

1. **Primary approach**: stdlib errors with `%w` wrapping
2. **Logging**: Consider slog for structured output (Go 1.21+ compatible)
3. **Stack traces**: Only if debugging complex async operations
4. **User errors vs internal errors**: Wrap internal errors, expose clean messages to users

Example pattern:
```go
// internal/errors.go
var (
    ErrWorkspaceNotFound = errors.New("workspace not found")
    ErrBranchExists      = errors.New("branch already exists")
)

// operations/workspace.go
func CreateWorkspace(name string) error {
    if err := git.CreateBranch(name); err != nil {
        if strings.Contains(err.Error(), "already exists") {
            return ErrBranchExists
        }
        return fmt.Errorf("create workspace %q: %w", name, err)
    }
    return nil
}
```

## Sources

### Primary Documentation
- [Go errors package](https://pkg.go.dev/errors)
- [pkg/errors package](https://pkg.go.dev/github.com/pkg/errors)
- [cockroachdb/errors](https://github.com/cockroachdb/errors)
- [errtrace](https://github.com/bracesdev/errtrace)
- [go-errors/errors](https://pkg.go.dev/github.com/go-errors/errors)

### Error Handling Best Practices
- [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- [JetBrains Go Error Handling Best Practices](https://www.jetbrains.com/guide/go/tutorials/handle_errors_in_go/best_practices/)
- [Better Stack: Fundamentals of Error Handling in Go](https://betterstack.com/community/guides/scaling-go/golang-errors/)
- [Effective Error Handling in Golang - Earthly Blog](https://earthly.dev/blog/golang-errors/)
- [DoltHub: Getting stack traces for errors in Go](https://www.dolthub.com/blog/2023-11-10-stack-traces-in-go/)

### Structured Logging
- [Golang Logging Libraries in 2025 - Uptrace](https://uptrace.dev/blog/golang-logging)
- [Better Stack: Logging in Go with Slog](https://betterstack.com/community/guides/logging/logging-in-go/)
- [Better Stack: Complete Guide to Zerolog](https://betterstack.com/community/guides/logging/zerolog/)
- [High-Performance Structured Logging with slog and zerolog - Leapcell](https://leapcell.io/blog/high-performance-structured-logging-in-go-with-slog-and-zerolog)

### Production Systems
- [Error Handling in Go Programs - GoTeleport](https://goteleport.com/blog/golang-error-handling/)
- [Error Handling and Logging in Go Applications 2025 - SecureGyan](https://www.securegyan.com/error-handling-and-logging-in-go-applications/)
- [Go Error Chains: Production-Grade Error Handling - BackendBytes](https://www.backendbytes.com/go/go-error-chains-production-grade-handling/)
- [The Go Ecosystem in 2025 - JetBrains Blog](https://blog.jetbrains.com/go/2025/11/10/go-language-trends-ecosystem-2025/)

### CockroachDB Errors Deep Dive
- [The Go standard error APIs - CockroachDB errors library, part 1](https://dr-knz.net/cockroachdb-errors-std-api.html)
- [Beyond fmt.Errorf() - CockroachDB errors library, part 4](https://dr-knz.net/cockroachdb-errors-everyday.html)

### Community Discussions
- [Docker/Moby: decide on future use of errors packages](https://github.com/moby/moby/discussions/46358)
- [Go proposal: add stack trace at error annotation](https://github.com/golang/go/issues/60873)

## Confidence Level

**High** - Based on official documentation, widely-referenced blog posts, and consistent recommendations across multiple authoritative sources. The "stdlib first" consensus is clear across Go community discussions and major projects.

## Related Questions

- How should errors be propagated in gRPC/HTTP services?
- What are patterns for error telemetry with OpenTelemetry?
- How to implement retry logic based on error types?
- What are best practices for error messages in CLIs?
