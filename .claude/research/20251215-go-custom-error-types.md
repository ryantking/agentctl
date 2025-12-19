# Research: Go Custom Error Types and Sentinel Errors
Date: 2025-12-15
Focus: When and how to implement custom error types, sentinel errors, and error wrapping in Go
Agent: researcher

## Summary

Go provides three primary strategies for error handling: sentinel errors (predefined values), custom error types (structs implementing the error interface), and opaque errors (treating errors as black boxes). Since Go 1.13, the `errors.Is()`, `errors.As()`, and `Unwrap()` methods enable robust error chain inspection. Custom error types with rich context (exit codes, stderr, operation details) follow patterns established by `fs.PathError`, `exec.ExitError`, and Kubernetes `StatusError`.

## Key Findings

- Sentinel errors are simple but create package coupling and reduce flexibility [Dave Cheney](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully)
- Custom error types should implement `Unwrap()` to participate in error chains [Go Blog](https://go.dev/blog/go1.13-errors)
- `errors.Is()` is preferred over `==` comparison for wrapped errors [pkg.go.dev/errors](https://pkg.go.dev/errors)
- `errors.As()` is preferred over type assertions for error inspection [pkg.go.dev/errors](https://pkg.go.dev/errors)
- Standard library patterns like `fs.PathError` and `exec.ExitError` provide templates for rich error types [Go Source](https://go.dev/src/io/fs/fs.go)
- Performance note: `errors.Is()` with sentinel errors can be 5x slower than direct comparison [DoltHub](https://www.dolthub.com/blog/2024-05-31-benchmarking-go-error-handling/)

## Detailed Analysis

### 1. When to Create Custom Error Types vs Standard Errors

#### Use Standard `errors.New()` or `fmt.Errorf()` When:
- Error message is sufficient context
- Callers don't need to programmatically distinguish error types
- Error is internal implementation detail

```go
// Simple, sufficient for most cases
return errors.New("invalid input")
return fmt.Errorf("failed to process %s: %w", name, err)
```

#### Use Sentinel Errors When:
- Error represents a well-known, stable condition
- Callers need to check for specific error conditions
- Error is part of package's public API
- You have full control of consuming code (internal packages, tests)

```go
var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
)

// Usage
if errors.Is(err, ErrNotFound) {
    // handle not found
}
```

#### Use Custom Error Types When:
- Error needs to carry structured data (exit code, path, operation, etc.)
- Callers need to extract specific error fields
- Error requires custom matching logic (`Is()` method)
- Building public APIs where error details matter

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}
```

### 2. Implementing Unwrap(), Is(), and As() Methods

#### The `Unwrap()` Method

Implement `Unwrap()` when your error wraps another error:

```go
type QueryError struct {
    Query string
    Err   error
}

func (e *QueryError) Error() string {
    return fmt.Sprintf("query %q failed: %v", e.Query, e.Err)
}

// Unwrap returns the underlying error
func (e *QueryError) Unwrap() error {
    return e.Err
}

// Usage:
qErr := &QueryError{Query: "SELECT *", Err: sql.ErrNoRows}
wrapped := fmt.Errorf("operation failed: %w", qErr)

// All these work:
errors.Is(wrapped, sql.ErrNoRows)  // true - traverses chain
errors.Unwrap(wrapped)             // returns qErr
```

For Go 1.20+ with multiple wrapped errors:

```go
type MultiError struct {
    Errs []error
}

// Unwrap returns multiple errors (Go 1.20+)
func (e *MultiError) Unwrap() []error {
    return e.Errs
}
```

#### The `Is()` Method

Implement custom `Is()` for semantic equality matching:

```go
type Error struct {
    Path string
    User string
}

func (e *Error) Error() string {
    return fmt.Sprintf("user %s: path %s", e.User, e.Path)
}

// Custom Is() for partial matching
func (e *Error) Is(target error) bool {
    t, ok := target.(*Error)
    if !ok {
        return false
    }
    // Match if fields are equal or target field is empty (wildcard)
    return (e.Path == t.Path || t.Path == "") &&
           (e.User == t.User || t.User == "")
}

// Usage:
err := &Error{Path: "/etc/passwd", User: "root"}
errors.Is(err, &Error{User: "root"})      // true - partial match
errors.Is(err, &Error{Path: "/etc/passwd"}) // true - partial match
errors.Is(err, &Error{User: "admin"})     // false
```

#### The `As()` Method

Implement custom `As()` for type conversion logic:

```go
type HTTPError struct {
    StatusCode int
    Body       []byte
}

func (e *HTTPError) Error() string {
    return fmt.Sprintf("HTTP %d", e.StatusCode)
}

// Custom As() to handle interface targets
func (e *HTTPError) As(target any) bool {
    switch t := target.(type) {
    case **HTTPError:
        *t = e
        return true
    case *int:
        *t = e.StatusCode
        return true
    }
    return false
}

// Usage:
err := &HTTPError{StatusCode: 404, Body: []byte("not found")}

var httpErr *HTTPError
errors.As(err, &httpErr)  // true, httpErr is set

var code int
errors.As(err, &code)     // true, code = 404
```

### 3. Sentinel Errors: Pros, Cons, and When to Use

#### Pros
- Simple and well-understood
- Clear API documentation
- Compatible with `errors.Is()` and error wrapping
- Efficient for equality comparison (when not wrapped)

#### Cons
- Creates coupling between packages
- Exported variables can be modified (mutability risk)
- Lack flexibility - can't carry additional context
- Performance impact with `errors.Is()` chain traversal (~5x slower)
- Breaking changes if sentinel is removed or renamed

#### Best Practices for Sentinel Errors

```go
// Declare at package level, unexported for internal use
var errInternal = errors.New("internal error")

// Exported for public API, document the contract
var (
    // ErrNotFound is returned when the requested resource does not exist.
    ErrNotFound = errors.New("not found")

    // ErrPermission is returned when access is denied.
    ErrPermission = errors.New("permission denied")
)

// Always wrap with context
func FetchItem(id string) (*Item, error) {
    item := lookup(id)
    if item == nil {
        return nil, fmt.Errorf("item %q: %w", id, ErrNotFound)
    }
    return item, nil
}

// Callers use errors.Is()
if errors.Is(err, ErrNotFound) {
    // Handle not found case
}
```

#### When to Use Sentinel Errors

| Scenario | Use Sentinel? | Rationale |
|----------|---------------|-----------|
| Package-internal errors | Yes | Full control, easy testing |
| Well-known conditions (EOF, NotExist) | Yes | Standard library pattern |
| Public API stable errors | Yes | With documentation |
| Errors needing rich context | No | Use custom types |
| Errors with dynamic data | No | Use custom types |
| High-performance hot paths | Consider | Direct comparison faster |

### 4. Structuring Error Types for Rich Context

#### Pattern 1: fs.PathError (Standard Library)

Ideal for operation + resource + underlying error pattern:

```go
// From Go standard library io/fs
type PathError struct {
    Op   string  // Operation that failed ("open", "read", etc.)
    Path string  // File path involved
    Err  error   // Underlying error
}

func (e *PathError) Error() string {
    return e.Op + " " + e.Path + ": " + e.Err.Error()
}

func (e *PathError) Unwrap() error {
    return e.Err
}

// Timeout() checks if underlying error is a timeout
func (e *PathError) Timeout() bool {
    t, ok := e.Err.(interface{ Timeout() bool })
    return ok && t.Timeout()
}

// Usage:
err := &fs.PathError{
    Op:   "open",
    Path: "/etc/shadow",
    Err:  syscall.EACCES,
}
// Error(): "open /etc/shadow: permission denied"
```

#### Pattern 2: exec.ExitError (Exit Codes + Stderr)

Ideal for command execution with exit status:

```go
// From Go standard library os/exec
type ExitError struct {
    *os.ProcessState        // Embedded, provides ExitCode()
    Stderr []byte           // Captured stderr output
}

func (e *ExitError) Error() string {
    return e.ProcessState.String()
}

// Usage when running commands:
out, err := exec.Command("ls", "/nonexistent").Output()
if err != nil {
    var exitErr *exec.ExitError
    if errors.As(err, &exitErr) {
        fmt.Printf("Exit code: %d\n", exitErr.ExitCode())
        fmt.Printf("Stderr: %s\n", exitErr.Stderr)
    }
}
```

#### Pattern 3: Kubernetes StatusError (API Errors)

Ideal for REST APIs with structured error responses:

```go
// Simplified from k8s.io/apimachinery/pkg/api/errors
type StatusError struct {
    ErrStatus Status
}

type Status struct {
    Message string
    Reason  StatusReason
    Code    int32
    Details *StatusDetails
}

type StatusDetails struct {
    Name   string
    Group  string
    Kind   string
    Causes []StatusCause
}

func (e *StatusError) Error() string {
    return e.ErrStatus.Message
}

func (e *StatusError) Status() Status {
    return e.ErrStatus
}

// Constructor functions for specific error types
func NewNotFound(resource, name string) *StatusError {
    return &StatusError{
        ErrStatus: Status{
            Message: fmt.Sprintf("%s %q not found", resource, name),
            Reason:  StatusReasonNotFound,
            Code:    404,
            Details: &StatusDetails{Name: name, Kind: resource},
        },
    }
}

// Classification helpers
func IsNotFound(err error) bool {
    var statusErr *StatusError
    if errors.As(err, &statusErr) {
        return statusErr.ErrStatus.Reason == StatusReasonNotFound
    }
    return false
}
```

#### Pattern 4: gRPC Status Errors

Ideal for RPC/API with rich error details:

```go
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/genproto/googleapis/rpc/errdetails"
)

// Creating errors
func validateUsername(name string) error {
    if !isAlphanumeric(name) {
        st := status.New(codes.InvalidArgument, "invalid username")
        v := &errdetails.BadRequest_FieldViolation{
            Field:       "username",
            Description: "must contain only alphanumeric characters",
        }
        br := &errdetails.BadRequest{}
        br.FieldViolations = append(br.FieldViolations, v)
        st, _ = st.WithDetails(br)
        return st.Err()
    }
    return nil
}

// Handling errors
func handleError(err error) {
    st, ok := status.FromError(err)
    if !ok {
        // Not a gRPC status error
        return
    }

    log.Printf("Code: %s, Message: %s", st.Code(), st.Message())

    for _, detail := range st.Details() {
        switch t := detail.(type) {
        case *errdetails.BadRequest:
            for _, v := range t.FieldViolations {
                log.Printf("Field %s: %s", v.Field, v.Description)
            }
        }
    }
}
```

#### Pattern 5: Custom Rich Error Type (Combining Patterns)

For CLI applications needing exit codes, stderr, and wrapped errors:

```go
type CommandError struct {
    Command  string    // Command that failed
    Args     []string  // Command arguments
    ExitCode int       // Process exit code
    Stderr   string    // Captured stderr
    Err      error     // Underlying error
}

func (e *CommandError) Error() string {
    if e.Stderr != "" {
        return fmt.Sprintf("command %s failed (exit %d): %s",
            e.Command, e.ExitCode, e.Stderr)
    }
    return fmt.Sprintf("command %s failed (exit %d): %v",
        e.Command, e.ExitCode, e.Err)
}

func (e *CommandError) Unwrap() error {
    return e.Err
}

// ExitCoder interface for CLI frameworks
func (e *CommandError) ExitCode() int {
    return e.ExitCode
}

// Timeout check delegation
func (e *CommandError) Timeout() bool {
    t, ok := e.Err.(interface{ Timeout() bool })
    return ok && t.Timeout()
}

// Constructor
func NewCommandError(cmd string, args []string, exitCode int, stderr string, err error) *CommandError {
    return &CommandError{
        Command:  cmd,
        Args:     args,
        ExitCode: exitCode,
        Stderr:   strings.TrimSpace(stderr),
        Err:      err,
    }
}

// Usage:
err := NewCommandError("git", []string{"push"}, 1,
    "error: failed to push some refs",
    errors.New("remote rejected"))

// Check with errors.Is/As
var cmdErr *CommandError
if errors.As(err, &cmdErr) {
    os.Exit(cmdErr.ExitCode())
}
```

### 5. Standard Library and Popular Package Examples

#### Go Standard Library

| Type | Package | Use Case |
|------|---------|----------|
| `fs.PathError` | io/fs | File operations with path context |
| `exec.ExitError` | os/exec | Process execution failures |
| `net.OpError` | net | Network operations |
| `url.Error` | net/url | URL parsing/operations |
| `json.SyntaxError` | encoding/json | JSON parsing with offset |
| `strconv.NumError` | strconv | Number conversion errors |

#### Popular Packages

| Package | Error Pattern | Example |
|---------|--------------|---------|
| k8s.io/apimachinery | StatusError with constructor functions | `errors.NewNotFound()` |
| google.golang.org/grpc | Status with codes and details | `status.Error(codes.NotFound, "msg")` |
| github.com/pkg/errors | Wrap with stack traces | `errors.Wrap(err, "context")` |
| github.com/hashicorp/go-multierror | Multiple error aggregation | `multierror.Append(err1, err2)` |

## Applicable Patterns

For CLI applications (like agentctl):

1. **Use sentinel errors** for well-known conditions:
   ```go
   var (
       ErrWorkspaceNotFound = errors.New("workspace not found")
       ErrBranchExists      = errors.New("branch already exists")
   )
   ```

2. **Use custom error types** for command execution:
   ```go
   type GitError struct {
       Operation string
       ExitCode  int
       Stderr    string
       Err       error
   }
   ```

3. **Implement Unwrap()** for error chain traversal
4. **Provide Is() methods** for flexible matching
5. **Use constructor functions** for consistent error creation

## Sources

- [Go Blog: Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- [Go errors package documentation](https://pkg.go.dev/errors)
- [Go Source: io/fs/fs.go](https://go.dev/src/io/fs/fs.go)
- [Go Source: os/exec package](https://pkg.go.dev/os/exec#ExitError)
- [Dave Cheney: Don't just check errors, handle them gracefully](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully)
- [Kubernetes API errors package](https://pkg.go.dev/k8s.io/apimachinery/pkg/api/errors)
- [gRPC Errors in Go](https://jbrandhorst.com/post/grpc-errors/)
- [Go Error Handling: Sentinel vs Custom Types](https://alesr.github.io/posts/go-errors/)
- [DoltHub: Benchmarking Go Error Handling](https://www.dolthub.com/blog/2024-05-31-benchmarking-go-error-handling/)
- [gosamples.dev: Wrap and Unwrap errors](https://gosamples.dev/wrap-unwrap-errors/)

## Confidence Level

**High** - Based on official Go documentation, standard library source code, Go blog posts, and patterns from production systems (Kubernetes, gRPC). Consensus exists on best practices, with the main debate being around sentinel error usage in public APIs.

## Related Questions

- How to handle multiple errors (Go 1.20+ `errors.Join`)?
- What are stack trace patterns with github.com/pkg/errors vs standard library?
- How to implement retryable error interfaces?
- What logging strategies work well with custom error types?
