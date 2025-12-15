# Research: Go Error Message Formatting and CLI UX Best Practices
Date: 2025-12-15
Focus: Error message conventions, formatting, actionability, and exit codes for Go CLI applications
Agent: researcher

## Summary

Go has established conventions for error message formatting: lowercase without trailing punctuation, add context without duplication using `fmt.Errorf` with `%w`, and make errors actionable by explaining what went wrong, why, and how to fix it. CLI applications should use exit codes from the sysexits.h convention (64-78 range) and direct errors to stderr.

## Key Findings

- Error strings should NOT be capitalized (unless starting with proper nouns) and should NOT end with punctuation [Go Wiki](https://go.dev/wiki/CodeReviewComments)
- Use `%w` verb to wrap errors when callers need programmatic access; use `%v` when you only want the message [Google Go Style](https://google.github.io/styleguide/go/best-practices.html)
- Good error messages communicate: what went wrong, why, and what to do about it [jayconrod.com](https://jayconrod.com/posts/116/error-handling-guidelines-for-go)
- Exit codes 64-78 are defined by sysexits.h and provide semantic meaning [Linux man pages](https://man7.org/linux/man-pages/man3/sysexits.h.3head.html)
- Handle errors once: don't log AND return the same error [Uber Go Guide](https://github.com/uber-go/guide/blob/master/style.md)

## Detailed Analysis

### 1. Error Message Formatting Conventions

#### Capitalization and Punctuation

From the [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments):

```go
// CORRECT - lowercase, no punctuation
fmt.Errorf("something bad")
fmt.Errorf("opening config file: %w", err)

// INCORRECT - capitalized, has punctuation
fmt.Errorf("Something bad.")
fmt.Errorf("Failed to open file: %w", err)
```

**Rationale**: Error messages are usually printed following other context. For example:
```go
log.Printf("Reading %s: %v", filename, err)
// Output: "Reading config.yaml: something bad"
// vs incorrect: "Reading config.yaml: Something bad."
```

The lowercase convention prevents awkward mid-sentence capitals.

**Exception**: Proper nouns and acronyms may be capitalized:
```go
fmt.Errorf("OAuth token expired")
fmt.Errorf("HTTP request failed: %w", err)
```

### 2. Adding Context Without Duplication

From [Google's Go Style Guide](https://google.github.io/styleguide/go/best-practices.html):

#### Good Context Addition

```go
// GOOD - adds meaningful context the caller doesn't have
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("loading config: %w", err)
        // os.Open already includes the path in its error
    }
    // ...
}
```

#### Redundant Context (Avoid)

```go
// BAD - duplicates information already in the error
func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("could not open %s: %w", path, err)
        // Path is ALREADY in os.Open's error message
    }
}
```

#### When NOT to Add Context

From Google's guide: Don't add annotations if "its sole purpose is to indicate a failure without adding new information."

```go
// GOOD - return as-is when no new context to add
func getUser(id string) (*User, error) {
    user, err := db.FindUser(id)
    if err != nil {
        return nil, err  // db.FindUser's error is sufficient
    }
    return user, nil
}
```

### 3. Choosing Between %w and %v

| Use `%w` when... | Use `%v` when... |
|------------------|------------------|
| Callers need `errors.Is()` / `errors.As()` | Hiding implementation details |
| Error is part of your API contract | Crossing system boundaries (RPC, storage) |
| Wrapping sentinel errors | Transforming errors for logging |

```go
// Use %w - caller may need to check for specific error
if err != nil {
    return fmt.Errorf("database query: %w", err)
}

// Use %v - hiding implementation, caller shouldn't depend on wrapped error
if err != nil {
    return fmt.Errorf("user lookup failed: %v", err)
}
```

**Placement**: Place `%w` at the end of the format string so error text mirrors error chain structure:
```go
fmt.Errorf("loading user %s: %w", userID, err)
// Produces: "loading user alice: database connection refused"
```

### 4. Writing Actionable Error Messages

From [jayconrod.com](https://jayconrod.com/posts/116/error-handling-guidelines-for-go), good error messages communicate three things:

1. **What went wrong** - the core problem
2. **Why it went wrong** - the underlying cause
3. **What can be done to fix it** - actionable steps (when applicable)

#### Good vs Bad Examples

```go
// BAD - generic, not actionable
errors.New("invalid input")
errors.New("operation failed")
errors.New("error occurred")

// GOOD - specific and actionable
errors.New("port 8080 already in use: try a different port with --port flag")
errors.New("config file not found at ~/.myapp/config.yaml: run 'myapp init' to create one")
fmt.Errorf("invalid date format %q: expected YYYY-MM-DD", input)
```

#### Include Relevant Values

```go
// BAD - no context
errors.New("file not found")

// GOOD - includes the problematic value
fmt.Errorf("file not found: %s", filename)

// GOOD - includes multiple relevant values
fmt.Errorf("user %s not in group %s", userID, groupName)
```

#### Avoid Implementation Details

```go
// BAD - exposes internal function names
errors.New("parseConfig: unmarshalYAML: field 'port' error")

// GOOD - user-focused message
errors.New("invalid config: 'port' must be a number between 1-65535")
```

### 5. Multi-Line Error Messages for Complex Failures

For CLI applications with complex failures, structure multi-line errors clearly:

```go
// Simple single-line for common cases
fmt.Errorf("connection refused: %s:%d", host, port)

// Multi-line for complex failures
msg := `validation failed:
  - field 'name': required but missing
  - field 'port': must be 1-65535, got %d
  - field 'timeout': must be positive duration`
fmt.Errorf(msg, port)
```

#### Popular CLI Patterns

**kubectl style** - structured output with clear categories:
```
error: unable to connect to server
  - server: https://kubernetes.local:6443
  - reason: connection refused
  - suggestion: check if the cluster is running
```

**Docker style** - clear cause and action:
```
Error response from daemon: pull access denied for myimage
You may need to 'docker login' first
```

### 6. Exit Codes and Error Categories

#### Reserved Exit Codes (Avoid)

| Code | Meaning | Avoid Because |
|------|---------|---------------|
| 0 | Success | Standard |
| 1 | General error | Too generic, but acceptable |
| 2 | Shell builtin misuse | Reserved by shells |
| 126 | Cannot execute | Reserved by shells |
| 127 | Command not found | Reserved by shells |
| 128+ | Fatal signals | 128+N = signal N |
| 255 | Out of range | Reserved |

#### sysexits.h Convention (Recommended for CLIs)

From [sysexits.h](https://man7.org/linux/man-pages/man3/sysexits.h.3head.html):

| Code | Name | Use Case |
|------|------|----------|
| 0 | EX_OK | Successful termination |
| 64 | EX_USAGE | Command line usage error (bad flags, wrong args) |
| 65 | EX_DATAERR | Input data format error |
| 66 | EX_NOINPUT | Cannot open input file |
| 67 | EX_NOUSER | Addressee/user unknown |
| 68 | EX_NOHOST | Host name unknown |
| 69 | EX_UNAVAILABLE | Service unavailable |
| 70 | EX_SOFTWARE | Internal software error (bug) |
| 71 | EX_OSERR | System error (can't fork, etc.) |
| 72 | EX_OSFILE | Critical OS file missing |
| 73 | EX_CANTCREAT | Can't create output file |
| 74 | EX_IOERR | Input/output error |
| 75 | EX_TEMPFAIL | Temporary failure (retry may help) |
| 76 | EX_PROTOCOL | Remote protocol error |
| 77 | EX_NOPERM | Permission denied |
| 78 | EX_CONFIG | Configuration error |

#### Go Implementation Pattern

```go
const (
    ExitOK          = 0
    ExitUsage       = 64
    ExitDataErr     = 65
    ExitNoInput     = 66
    ExitUnavailable = 69
    ExitSoftware    = 70
    ExitIOErr       = 74
    ExitTempFail    = 75
    ExitNoPerm      = 77
    ExitConfig      = 78
)

func main() {
    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)

        var exitErr *ExitError
        if errors.As(err, &exitErr) {
            os.Exit(exitErr.Code)
        }
        os.Exit(1)
    }
}
```

### 7. CLI Error Handling Patterns

#### Cobra: Use RunE Instead of Run

From [JetBrains Guide](https://www.jetbrains.com/guide/go/tutorials/cli-apps-go-cobra/error_handling/):

```go
var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "My application",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Return errors instead of calling os.Exit
        if len(args) == 0 {
            return errors.New("at least one argument required")
        }
        return nil
    },
}
```

#### Centralized Error Handling

From [dev.to](https://dev.to/eminetto/error-handling-of-cli-applications-in-golang-4c7l):

```go
func main() {
    if err := rootCmd.Execute(); err != nil {
        // Single point for error output
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

**Benefits**:
- Consistent error formatting
- `defer` statements always run
- Single point for logging/metrics
- Testable error paths

#### Control Error Display in Cobra

```go
rootCmd.SilenceUsage = true   // Don't show usage on runtime errors
rootCmd.SilenceErrors = true  // Handle error display yourself
```

### 8. Good vs Bad Error Messages Summary

| Aspect | Bad | Good |
|--------|-----|------|
| Capitalization | "Something went wrong" | "something went wrong" |
| Punctuation | "file not found." | "file not found" |
| Specificity | "invalid input" | "invalid port: must be 1-65535" |
| Blame | "you entered invalid data" | "invalid date format" |
| Jargon | "EOF in YAML unmarshal" | "unexpected end of config file" |
| Actionability | "permission denied" | "permission denied: try running with sudo" |
| Context | "error occurred" | "connecting to database: connection refused" |
| Duplication | "opening /etc/config: open /etc/config: no such file" | "loading config: no such file" |

## Applicable Patterns

For the agentctl codebase:

1. **Lowercase errors** - All error strings should start lowercase
2. **No trailing punctuation** - Remove periods from error messages
3. **Context wrapping** - Use `fmt.Errorf("action: %w", err)` pattern
4. **Exit code constants** - Define sysexits.h constants for semantic exits
5. **Centralized handling** - Return errors to main(), handle output there
6. **Actionable messages** - Include suggestions when users can fix the issue

## Sources

- [Go Wiki: Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Google Go Style Best Practices](https://google.github.io/styleguide/go/best-practices.html)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Error Handling Guidelines for Go - jayconrod.com](https://jayconrod.com/posts/116/error-handling-guidelines-for-go)
- [Error Handling and Go - Go Blog](https://go.dev/blog/error-handling-and-go)
- [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- [Error Handling in Cobra - JetBrains Guide](https://www.jetbrains.com/guide/go/tutorials/cli-apps-go-cobra/error_handling/)
- [Error Handling of CLI Applications in Golang - dev.to](https://dev.to/eminetto/error-handling-of-cli-applications-in-golang-4c7l)
- [sysexits.h - Linux man pages](https://man7.org/linux/man-pages/man3/sysexits.h.3head.html)
- [Exit Codes With Special Meanings - TLDP](https://tldp.org/LDP/abs/html/exitcodes.html)
- [Standard Exit Status Codes - Baeldung](https://www.baeldung.com/linux/status-codes)
- [Error Message Guidelines - NN/g](https://www.nngroup.com/articles/error-message-guidelines/)
- [When Life Gives You Lemons - Wix UX](https://wix-ux.com/when-life-gives-you-lemons-write-better-error-messages-46c5223e1a2f)
