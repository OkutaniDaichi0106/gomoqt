# Go Coding Style Guide for gomoqt

This style guide outlines the coding conventions and best practices to follow when contributing to the gomoqt project. Adhering to these guidelines ensures consistency across the codebase and makes the code more maintainable.

## Package Structure

- **Main Package**: `moqt` - Core implementation of the MOQ Transfork protocol
- **Internal Packages**: Place implementation details under `moqt/internal/`
- **Examples**: Place example applications under `examples/` directory
- **Support Libraries**: Additional packages like `lomc` (Low Overhead Media Container) and `catalog` (MOQ Catalog)

## Naming Conventions

- Use **CamelCase** for exported functions, types, and variables (e.g., `Session`, `TrackPriority`)
- Use **camelCase** for unexported functions, types, and variables (e.g., `init()`)
- Use **ALL_CAPS** for constants when appropriate
- Type names should be descriptive nouns
- Function names should be descriptive verbs
- Test functions should be named `Test<FunctionName>_<Scenario>` (e.g., `TestFrameMessage_EncodeDecode`)
- Interface names typically do not end with '-er' (e.g., use `Session` not `Sessioner`)

## Code Organization

- Group related functions and types together
- Place type definitions before their methods
- Organize imports in the standard Go way:
  - Standard library first
  - Third-party packages next
  - Project packages last
- Use blank lines to separate logical sections within functions
- When implementing test files, follow the `_test` package approach:
  ```go
  package message_test  // Not just "package message"
  ```

## Error Handling

- Return errors rather than using panics
- Use meaningful error messages that assist in troubleshooting
- For logging, use the `slog` package
- Initialize loggers in a consistent way:
```go
if c.Logger == nil {
    c.Logger = slog.Default()
}
```
- Wrap errors with context information when appropriate
- Log errors with context before returning them:
```go
if err != nil {
    c.Logger.Error("failed to establish connection", "error", err)
    return nil, fmt.Errorf("failed to establish connection: %w", err)
}
```

## Documentation

- All exported functions, types, and variables should have documentation comments
- Documentation should follow Go standard format:
  - Start with the name of the thing being documented
  - Use complete sentences with proper punctuation
- Package-level documentation should provide an overview of the package's purpose
- Include example usage for complex functions or types
- Document implementation status in README files
- For implementation status tracking, use these emojis in markdown:
  - `:white_check_mark:` - Implemented and tested
  - `:construction:` - Partially implemented
  - `:x:` - Not implemented

## Testing

- Write tests for all exported functionality
- Place tests in a separate package with suffix `_test` (e.g., `package message_test`)
- Use the `testing` package and the `testify` library for assertions:
```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)
```
- Use `assert` for non-fatal assertions and `require` for fatal assertions
- Test both success and failure paths
- Aim for high test coverage, particularly for complex logic
- Use a consistent pattern for encoding/decoding tests:
```go
func TestSomeMessage_EncodeDecode(t *testing.T) {
    // Setup test data

    // Test encoding
    var buf bytes.Buffer
    err := message.Encode(&buf)
    require.NoError(t, err)

    // Test decoding
    var decoded message.Type
    err = decoded.Decode(&buf)
    require.NoError(t, err)

    // Verify results
    assert.Equal(t, original, decoded)
}
```

## Formatting and Style

- Use `gofmt` or `goimports` to format code
- Use the project's linter configuration:
```bash
golangci-lint run --config=.github/.golangci.yml
```
- Line length should be reasonable (the project uses lll linter)
- Avoid unnecessary comments that just repeat what the code does
- Use comments to explain "why", not "what"
- Use the `just` command runner for common development tasks:
```bash
just fmt      # Format code
just lint     # Run linter
just test     # Run tests
just check    # Overall quality checks
```

## Concurrency

- Use proper synchronization with mutexes or channels
- Document concurrency guarantees for public APIs
- Avoid sharing mutable state without proper synchronization
- Use context for cancellation where appropriate
- Place context as the first parameter for functions that accept it:
```go
func DoSomething(ctx context.Context, param1 Type, param2 Type) Result
```

## Go Version

- The project requires Go 1.22 or later
- Don't use features from experimental or unreleased Go versions

## Dependencies

- Minimize external dependencies
- For QUIC and WebTransport, use the standard libraries:
  - `github.com/quic-go/quic-go`
  - `github.com/quic-go/webtransport-go`
- For testing, use `github.com/stretchr/testify`
- Keep dependencies updated with Dependabot

## Error Messages

- Error messages should be clear and actionable
- Begin error messages with lowercase letters
- Do not include punctuation at the end of error messages
- Include relevant context in errors
- Use `fmt.Errorf` with `%w` for wrapping errors

## Examples

### Good Function Declaration and Documentation

```go
// NewSession creates a new MOQT session with the provided options.
// It returns a Session interface and any error encountered during initialization.
func NewSession(opts SessionOptions) (Session, error) {
    // Implementation
}
```

### Good Error Handling

```go
func (c *Client) Dial(url string, ctx context.Context) (Session, *Info, error) {
    c.init()
    
    if url == "" {
        return nil, nil, errors.New("url cannot be empty")
    }
    
    // Implementation
    
    if err != nil {
        c.Logger.Error("failed to establish connection", "error", err)
        return nil, nil, fmt.Errorf("failed to establish connection: %w", err)
    }
    
    return session, info, nil
}
```

### Good Test Pattern

```go
func TestFrameMessage_EncodeDecode(t *testing.T) {
    // Setup
    original := message.FrameMessage{
        // Initialize with test data
    }
    
    // Test encoding
    var buf bytes.Buffer
    err := original.Encode(&buf)
    require.NoError(t, err)
    
    // Test decoding
    var decoded message.FrameMessage
    err = decoded.Decode(&buf)
    require.NoError(t, err)
    
    // Verify results
    assert.Equal(t, original, decoded)
}
```

## Best Practices

1. **Follow Go's Idioms**: Write idiomatic Go code
2. **Keep Functions Focused**: Each function should do one thing well
3. **Avoid Global State**: Prefer passing dependencies explicitly
4. **Embrace Interfaces**: Use interfaces for flexibility and testability
5. **Use Context Appropriately**: For cancellation, deadlines, and request-scoped values
6. **Optimize Judiciously**: Write clear code first, optimize only when necessary
7. **Consider Security**: Validate inputs, handle errors, and avoid common vulnerabilities
8. **Internationalization**: Support both English (README.md) and Japanese (README.ja.md) documentation

When implementing new features or modifying existing code, please follow these guidelines to maintain the quality and consistency of the gomoqt project.
