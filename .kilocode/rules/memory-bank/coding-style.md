# Go Coding Style Guide

## Code Organization & Structure

**Logical Segmentation**
- One concept per file (e.g., handlers, models, services).
- Keep functions under 20-30 lines when possible.
- Use meaningful package names that reflect functionality.
- Group related types and functions together.

**Interface Design**
- Prefer small, focused interfaces (1-3 methods).
- Define interfaces where they're used, not where they're implemented.
- Use composition over inheritance.

## Conciseness Rules

**Variable & Function Naming**
- Use short names in small scopes (`i`, `err`, `ctx`).
- Longer names for broader scopes (`userRepository`, `configManager`).
- Omit obvious type information (`users []User` not `userList []User`).

**Error Handling**
```go
// Prefer early returns
if err != nil {
    return err
}
// over nested if-else blocks
```

**Type Declarations**
- Use type inference: `users := []User{}` not `var users []User = []User{}`.
- Combine related variable declarations: `var (name string; age int)`.

## Go Idioms for Conciseness

**Zero Values**
- Leverage Go's zero values instead of explicit initialization.
- Use `var buf bytes.Buffer` instead of `buf := bytes.Buffer{}`.

**Struct Initialization**
```go
// Prefer struct literals with field names
user := User{Name: name, Email: email}
// over multiple assignments
```

**Method Receivers**
- Use pointer receivers for modification or large structs.
- Use value receivers for small, immutable data.

**Channel Operations**
- Use `select` with `default` for non-blocking operations.
- Prefer `range` over explicit channel reads.

## Context Reduction Strategies

**Function Signatures**
- Group related parameters into structs.
- Use functional options pattern for complex configurations.
- Return early and often to reduce nesting.

**Constants and Enums**
```go
const (
    StatusPending = iota
    StatusApproved
    StatusRejected
)
```

**Embed Common Patterns**
- Use `sync.Once` for lazy initialization.
- Embed `sync.Mutex` in structs needing synchronization.
- Use `context.Context` as first parameter in functions.

## File Organization Rules

**Package Structure**
```
/cmd          - main applications
/internal     - private code
/pkg          - public libraries
/api          - API definitions
```

**Import Grouping**
1. Standard library
2. Third-party packages  
3. Local packages
(Separate groups with blank lines)

**Testing**
- Place tests in same package with `_test.go` suffix.
- Use table-driven tests for multiple scenarios.
- Keep test functions focused and named clearly.