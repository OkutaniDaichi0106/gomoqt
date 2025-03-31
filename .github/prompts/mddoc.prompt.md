# Markdown Documentation Generation Template for Go Components

This template helps you create prompts to generate comprehensive Markdown documentation for Go components.

## Basic Prompt Structure

```
# MD File Generation: [ComponentName].md

## Requirements

Please generate detailed Markdown documentation for the [ComponentName] component.

### Sections to Include

1. **Type Definitions**: Complete interface and struct definitions
2. **Package Dependencies**: Visualization of dependency structure
3. **Architecture**: Overview of design patterns and responsibility allocation
4. **Error Handling**: Error types and error handling patterns
5. **Code Examples**: Common usage patterns and edge cases
6. **Security**: Validation, authentication, and authorization methods
7. **Testing Strategy**: Test examples and mocking techniques
```

## 1. Section Creation Guide

### 1.1 Type Definitions
```go
// UserRepository interface
type UserRepository interface {
    FindByID(ctx context.Context, id string) (*User, error)
    // ...other methods
}

// User struct
type User struct {
    ID        string    `json:"id" db:"id"`
    Email     string    `json:"email" db:"email"`
    // ...other fields
}
```

### 1.2 Error Handling
```go
// Error definitions
var (
    ErrUserNotFound = errors.New("user not found")
    // ...other errors
)

// Error wrapping example
if err != nil {
    return nil, fmt.Errorf("failed to find user: %w", err)
}
```

### 1.3 Pattern Examples
```go
// Transaction handling
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return fmt.Errorf("failed to begin transaction: %w", err)
}

defer func() {
    if err != nil {
        tx.Rollback()
        return
    }
    err = tx.Commit()
}()

// Concurrency pattern
func FetchResources(ids []string) ([]*Resource, error) {
    var wg sync.WaitGroup
    results := make([]*Resource, len(ids))

    for i, id := range ids {
        wg.Add(1)
        go func(idx int, resourceID string) {
            defer wg.Done()
            results[idx], _ = fetchSingle(resourceID)
        }(i, id)
    }

    wg.Wait()
    return results, nil
}
```

## 2. Integration with Development Workflow

- **File Placement**: Save in the same directory as the related Go file (e.g., user_service.go and user_service.md)
- **Documentation Updates**: Update the MD file when code changes and include in review processes
- **AI Assistance**: Use as context information with AI development tools

## 3. Prompt Example

```
# MD File Generation: user_service.md

## Requirements

Please generate detailed Markdown documentation for the UserService component.

### Sections to Include

1. **Type Definitions**
   - UserService interface and methods
   - Related structs (User, UserInput, etc.)

2. **Dependencies**
   - Dependency on UserRepository
   - Dependencies on external authentication services

3. **Architecture**
   - Clean Architecture implementation
   - Repository pattern usage

4. **Error Handling**
   - Custom error type definitions
   - Error wrapping examples

5. **Code Examples**
   - Transaction handling
   - Authentication and authorization flows

6. **Security**
   - Input validation and sanitization
   - Authentication and authorization examples

7. **Testing Strategy**
   - Unit test examples
   - Integration test strategies
```

Use this template to efficiently create detailed documentation for your Go components.
