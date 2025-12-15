# Research: Go 1.13+ Error Wrapping Best Practices
Date: 2025-12-15
Focus: Error wrapping with %w, errors.Is/As, layer patterns, and anti-patterns
Agent: researcher

## Summary

Go 1.13 introduced structured error wrapping via `fmt.Errorf` with the `%w` verb, along with `errors.Is()` and `errors.As()` for inspecting error chains. The key decision when wrapping errors is whether to expose the underlying error to callers (%w) or hide implementation details (%v). Wrapping makes errors part of your API contract, so deliberate decisions are required.

## Key Findings

- Use `%w` to wrap errors when callers need to inspect the underlying error; use `%v` to add context while hiding implementation details ([Go Blog](https://go.dev/blog/go1.13-errors))
- `errors.Is()` and `errors.As()` examine the entire error chain, replacing direct equality checks and type assertions ([errors package](https://pkg.go.dev/errors))
- Wrap errors at layer boundaries (repository->service->handler) to add context; avoid excessive wrapping within the same layer ([JetBrains Guide](https://www.jetbrains.com/guide/go/tutorials/handle_errors_in_go/best_practices/))
- Every error exposed via `%w` becomes part of your public API; clients may depend on it ([Go Wiki FAQ](https://go.dev/wiki/ErrorValueFAQ))

## Detailed Analysis

### 1. When to Use fmt.Errorf with %w vs %v

#### Use %w When:
- You intentionally want to expose the underlying error to callers
- Callers need to make decisions based on the specific error type
- The wrapped error is part of your documented API contract
- The error comes from external input (e.g., `io.Reader` provided by caller)

```go
// Good: Expose parse errors from user-provided reader
func Parse(r io.Reader) (*Config, error) {
    data, err := io.ReadAll(r)
    if err != nil {
        return nil, fmt.Errorf("reading config: %w", err)
    }
    // ...
}
```

#### Use %v When:
- The underlying error is an implementation detail
- You want to preserve abstraction layers
- You might change the underlying implementation later
- You're wrapping errors from internal dependencies

```go
// Good: Hide database implementation details
func (r *UserRepo) FindByID(id string) (*User, error) {
    row := r.db.QueryRow("SELECT * FROM users WHERE id = ?", id)
    var u User
    if err := row.Scan(&u.ID, &u.Name); err != nil {
        // Use %v to hide sql.ErrNoRows - caller shouldn't know we use SQL
        return nil, fmt.Errorf("finding user %s: %v", id, err)
    }
    return &u, nil
}
```

#### Key Decision Framework:

| Aspect | Use `%w` | Use `%v` |
|--------|----------|----------|
| Error chain inspection | Enabled | Disabled |
| API stability | Commits to error type | Flexible to change |
| Implementation details | Exposed | Hidden |
| Client dependency | Possible | Prevented |

### 2. How errors.Is() and errors.As() Work

#### errors.Is() - Sentinel Value Comparison

Examines the entire error chain looking for a match with a target error value.

```go
// Old pattern (doesn't check wrapped errors):
if err == io.ErrUnexpectedEOF {
    // handle
}

// New pattern (checks entire chain):
if errors.Is(err, io.ErrUnexpectedEOF) {
    // handle - works even if err wraps io.ErrUnexpectedEOF
}
```

**How it works:**
1. Compares `err` directly with target
2. If `err` implements `Unwrap() error`, recursively checks the unwrapped error
3. If `err` implements `Unwrap() []error`, checks each error in the slice
4. If `err` implements `Is(error) bool`, calls that method for custom matching

```go
// Custom Is method for flexible matching
type PathError struct {
    Path string
    User string
}

func (e *PathError) Is(target error) bool {
    t, ok := target.(*PathError)
    if !ok {
        return false
    }
    // Match if fields are equal or target field is empty (wildcard)
    return (e.Path == t.Path || t.Path == "") &&
           (e.User == t.User || t.User == "")
}

// Usage: matches any PathError with User="admin"
if errors.Is(err, &PathError{User: "admin"}) {
    // ...
}
```

#### errors.As() - Type Assertion

Finds the first error in the chain assignable to the target type.

```go
// Old pattern (only checks top-level error):
if e, ok := err.(*os.PathError); ok {
    fmt.Println(e.Path)
}

// New pattern (checks entire chain):
var pathErr *os.PathError
if errors.As(err, &pathErr) {
    fmt.Println(pathErr.Path) // Access custom fields
}
```

**How it works:**
1. The target must be a non-nil pointer
2. Traverses the error chain (pre-order, depth-first)
3. Assigns the first matching error to target
4. If error implements `As(any) bool`, calls that method

```go
// Custom error type with Unwrap
type QueryError struct {
    Query string
    Err   error
}

func (e *QueryError) Error() string {
    return fmt.Sprintf("query %q: %v", e.Query, e.Err)
}

func (e *QueryError) Unwrap() error {
    return e.Err
}

// Usage
var qErr *QueryError
if errors.As(err, &qErr) {
    log.Printf("Failed query: %s", qErr.Query)
}
```

### 3. Best Practices for Wrapping Errors at Different Layers

#### Repository Layer
- Translate storage-specific errors to domain errors
- Hide database implementation details
- Use sentinel errors for expected conditions

```go
var (
    ErrNotFound     = errors.New("not found")
    ErrDuplicate    = errors.New("duplicate entry")
)

func (r *UserRepo) FindByEmail(email string) (*User, error) {
    var user User
    err := r.db.Where("email = ?", email).First(&user).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            // Translate to domain error
            return nil, fmt.Errorf("user with email %s: %w", email, ErrNotFound)
        }
        // Hide GORM details for unexpected errors
        return nil, fmt.Errorf("querying user: %v", err)
    }
    return &user, nil
}
```

#### Service Layer
- Add business context to errors
- Wrap repository errors when propagating
- Log errors only when handling them

```go
func (s *UserService) Register(email, password string) (*User, error) {
    // Check for existing user
    existing, err := s.repo.FindByEmail(email)
    if err != nil && !errors.Is(err, ErrNotFound) {
        return nil, fmt.Errorf("checking existing user: %w", err)
    }
    if existing != nil {
        return nil, fmt.Errorf("registration failed: %w", ErrDuplicate)
    }

    // Create user...
    user, err := s.repo.Create(email, hashedPassword)
    if err != nil {
        return nil, fmt.Errorf("creating user: %w", err)
    }
    return user, nil
}
```

#### Handler/Transport Layer
- Map domain errors to HTTP status codes
- Log errors with full context
- Return sanitized error messages to clients

```go
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
    var req RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    user, err := h.service.Register(req.Email, req.Password)
    if err != nil {
        // Map domain errors to HTTP responses
        switch {
        case errors.Is(err, ErrDuplicate):
            http.Error(w, "email already registered", http.StatusConflict)
        case errors.Is(err, ErrNotFound):
            http.Error(w, "not found", http.StatusNotFound)
        default:
            // Log unexpected errors with full chain
            log.Printf("registration error: %v", err)
            http.Error(w, "internal error", http.StatusInternalServerError)
        }
        return
    }

    json.NewEncoder(w).Encode(user)
}
```

### 4. When to Expose vs Hide Underlying Errors

#### Expose (Use %w) When:
- Error is part of documented API contract
- Callers need to handle specific error conditions
- Error comes from caller-provided dependencies
- Building reusable libraries with stable error types

```go
// Document exposed errors
// FetchItem returns the named item.
// If no item exists, returns an error wrapping ErrNotFound.
func FetchItem(name string) (*Item, error) {
    item, err := db.Get(name)
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("item %q: %w", name, ErrNotFound)
    }
    // ...
}
```

#### Hide (Use %v) When:
- Error reveals implementation details
- You may change underlying implementation
- Error is from internal dependencies
- Security concerns about exposing internals

```go
// Hide implementation - might switch from Redis to Memcached
func (c *Cache) Get(key string) ([]byte, error) {
    val, err := c.redis.Get(ctx, key).Bytes()
    if err != nil {
        // Don't expose redis.Nil - use %v
        return nil, fmt.Errorf("cache miss for %s: %v", key, err)
    }
    return val, nil
}
```

### 5. Common Anti-Patterns to Avoid

#### Anti-Pattern 1: Ignoring Errors
```go
// BAD: Ignoring error with blank identifier
result, _ := riskyOperation()

// GOOD: Always handle or explicitly document why ignored
result, err := riskyOperation()
if err != nil {
    return fmt.Errorf("risky operation: %w", err)
}
```

#### Anti-Pattern 2: Bare Error Returns (No Context)
```go
// BAD: Loses context about where/why error occurred
func processFile(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return err  // Where did this come from?
    }
    return parse(data)
}

// GOOD: Add context at each layer
func processFile(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("reading file %s: %w", path, err)
    }
    if err := parse(data); err != nil {
        return fmt.Errorf("parsing file %s: %w", path, err)
    }
    return nil
}
```

#### Anti-Pattern 3: Over-Wrapping
```go
// BAD: Excessive wrapping creates verbose, redundant messages
func a() error {
    if err := b(); err != nil {
        return fmt.Errorf("in a: %w", err)
    }
    return nil
}

func b() error {
    if err := c(); err != nil {
        return fmt.Errorf("in b: %w", err)
    }
    return nil
}

// Results in: "in a: in b: in c: actual error"

// GOOD: Wrap at layer boundaries, add meaningful context
func a() error {
    if err := b(); err != nil {
        return fmt.Errorf("processing request: %w", err)
    }
    return nil
}
```

#### Anti-Pattern 4: Logging and Returning
```go
// BAD: Same error logged multiple times up the stack
func handler() error {
    err := service()
    if err != nil {
        log.Printf("service error: %v", err)  // Logged here
        return err
    }
    return nil
}

func service() error {
    err := repo()
    if err != nil {
        log.Printf("repo error: %v", err)  // And here
        return err
    }
    return nil
}

// GOOD: Log only at the top level
func handler() error {
    err := service()
    if err != nil {
        log.Printf("request failed: %v", err)  // Log once with full context
        return err
    }
    return nil
}

func service() error {
    err := repo()
    if err != nil {
        return fmt.Errorf("service operation: %w", err)  // Just wrap
    }
    return nil
}
```

#### Anti-Pattern 5: Using == Instead of errors.Is
```go
// BAD: Breaks with wrapped errors
if err == sql.ErrNoRows {
    // Won't match if err wraps sql.ErrNoRows
}

// GOOD: Works with error chains
if errors.Is(err, sql.ErrNoRows) {
    // Matches even if err wraps sql.ErrNoRows
}
```

#### Anti-Pattern 6: Type Assertion Instead of errors.As
```go
// BAD: Only checks top-level error
if pathErr, ok := err.(*os.PathError); ok {
    // Won't match if err wraps *os.PathError
}

// GOOD: Checks entire chain
var pathErr *os.PathError
if errors.As(err, &pathErr) {
    // Matches anywhere in chain
}
```

#### Anti-Pattern 7: Panic for Expected Errors
```go
// BAD: Using panic for expected conditions
func GetUser(id string) *User {
    user, err := db.Find(id)
    if err != nil {
        panic(err)  // Don't panic on expected errors
    }
    return user
}

// GOOD: Return errors for expected conditions
func GetUser(id string) (*User, error) {
    user, err := db.Find(id)
    if err != nil {
        return nil, fmt.Errorf("getting user %s: %w", id, err)
    }
    return user, nil
}
```

#### Anti-Pattern 8: Exposing Internal Errors in Libraries
```go
// BAD: Exposes internal dependency
func (c *Client) Fetch(url string) ([]byte, error) {
    resp, err := c.http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("fetch: %w", err)  // Exposes http implementation
    }
    // ...
}

// GOOD: Define library-specific errors
var ErrNetworkFailure = errors.New("network failure")

func (c *Client) Fetch(url string) ([]byte, error) {
    resp, err := c.http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("fetch %s: %w", url, ErrNetworkFailure)
    }
    // ...
}
```

## Applicable Patterns

For Go codebases using layered architecture:

1. **Define domain-specific sentinel errors** in a dedicated `errors.go` file
2. **Wrap at layer boundaries** with meaningful context
3. **Use `errors.Is()`** for sentinel comparisons, `errors.As()` for type inspection
4. **Document wrapped errors** in function comments when they're part of API
5. **Log errors once** at the top-most handler layer
6. **Use `%v` for internal implementation details**, `%w` for documented API errors

## Sources

- [Working with Errors in Go 1.13 - Official Go Blog](https://go.dev/blog/go1.13-errors)
- [errors package - Official Go Documentation](https://pkg.go.dev/errors)
- [Go Wiki: Error Values FAQ](https://go.dev/wiki/ErrorValueFAQ)
- [Best Practices for Error Handling in Go - JetBrains Guide](https://www.jetbrains.com/guide/go/tutorials/handle_errors_in_go/best_practices/)
- [Effective Error Handling in Golang - Earthly Blog](https://earthly.dev/blog/golang-errors/)
- [Wrap and Unwrap errors in Go - gosamples.dev](https://gosamples.dev/wrap-unwrap-errors/)
- [Handle errors in Go with errors.Is() and errors.As() - gosamples.dev](https://gosamples.dev/check-error-type/)
