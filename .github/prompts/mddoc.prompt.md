# Go Implementation Documentation Generator Prompt v2

As a Go expert, your task is to create precise, standardized documentation for Go implementation files. Each documentation must strictly follow the template structure with no deviations.

## File Relationship

This template is designed to generate documentation for individual Go component files:
- Each Go component implementation file (named `component.go`) should have a corresponding documentation file (named `component.doc.md`)
- The documentation should cover exactly what is implemented in that specific Go file
- Focus on documenting the component's purpose, types, functions, and interactions with other components

## Usage Instructions

To use this template:
1. Examine the Go file for which documentation is needed (e.g., `component.go`)
2. Identify the component name, types, methods, and dependencies from the file
3. Generate a markdown document following this template's structure
4. Save the documentation with the same base name as the Go file but with `.doc.md` extension (e.g., `component.doc.md`)
5. Ensure all placeholders are replaced with concrete implementation details from the source file

## Documentation Structure Requirements

Every documentation MUST:
- Include ALL sections in the exact order specified
- Maintain consistent formatting throughout
- Use complete, executable code examples
- Provide specific implementation details, not generalizations
- Follow standard Go conventions for naming and documentation
- Reflect only what exists in the corresponding Go file

## Standardized Documentation Template

```markdown
# {ComponentName} Component Documentation

## Overview

{EXACTLY 3-5 full sentences describing the component's purpose, role in the architecture, and primary responsibilities. Must include how it interacts with other components.}

## Type Definitions

{FOR EACH type, follow this exact structure}

### {TypeName}

~~~go
// Complete type definition with comments following Go standards
type {TypeName} struct {
    {FieldName} {FieldType} // Single-line description of field purpose
    {FieldName} {FieldType} // Single-line description of field purpose
}
~~~

{EXACTLY 2-3 sentences describing the type's purpose and usage context}

- `{FieldName}`: {EXACTLY 1 sentence describing the field's purpose and constraints}
- `{FieldName}`: {EXACTLY 1 sentence describing the field's purpose and constraints}

## Package Dependencies

~~~go
import (
    // Standard library packages (alphabetically ordered)
    "{stdlib1}"
    "{stdlib2}"

    // External dependencies (alphabetically ordered)
    "{extdep1}"
    "{extdep2}"

    // Internal dependencies (alphabetically ordered)
    "{intdep1}"
    "{intdep2}"
)
~~~

~~~mermaid
graph TD
    {ComponentName}[{ComponentName}] --> {Dependency1}[{Dependency1}]
    {ComponentName} --> {Dependency2}[{Dependency2}]
    {ComponentName} --> {Dependency3}[{Dependency3}]
~~~

## References

This section documents types that are used in the component but not defined within this file. It provides brief descriptions to help resolve type references during code generation.

{FOR EACH type used but not defined in this file, follow this exact structure}

### {TypeName}
{EXACTLY 1-2 sentences describing the type and its role}

~~~go
package {PackageName}

type {TypeName} struct {
    {FieldName} {FieldType}
    {FieldName} {FieldType}
}

func ({TypeName}) {MethodName}({ArgumentName} {ArgumentType}) ({ReturnType})

func ({TypeName}) {MethodName}({ArgumentName} {ArgumentType}) ({ReturnType})

func {FunctionName}({ArgumentName} {ArgumentType}) ({ReturnType})
~~~

### {InterfaceName}
{EXACTLY 1-2 sentences describing the interface and its role}

~~~go
package {PackageName}

type {InterfaceName} interface {
    {MethodName}({ArgumentName} {ArgumentType}) ({ReturnType})
    {MethodName}({ArgumentName} {ArgumentType}) ({ReturnType})
}

// Implementations
// {ImplementingType} implements {InterfaceName}
~~~

## Architecture

{EXACTLY 3-5 sentences on the architectural approach and design patterns}

1. {Design pattern name}: {EXACTLY 1-2 sentences explaining how it's implemented}
2. {Design pattern name}: {EXACTLY 1-2 sentences explaining how it's implemented}
3. {Key architectural decision}: {EXACTLY 1-2 sentences explaining rationale}

## API Documentation

{FOR EACH exported function/method, follow this exact format}

### Method: {MethodName}

Signature:
~~~go
func ({ReceiverName} *{ReceiverType}) {MethodName}({ParamName} {ParamType}, {ParamName} {ParamType}) ({ReturnType}, error)
~~~

Description: {EXACTLY 1-2 sentences describing the method's purpose and behavior}

Parameters:
- `{ParamName}`: {EXACTLY 1 sentence describing the parameter purpose and constraints}
- `{ParamName}`: {EXACTLY 1 sentence describing the parameter purpose and constraints}

Return Values:
- `{ReturnType}`: {EXACTLY 1 sentence describing what is returned and when}
- `error`: Returns an error in the following conditions:
  - {Specific error condition 1}
  - {Specific error condition 2}
  - {Specific error condition 3 (if applicable)}

Usage Example:
~~~go
func Example{MethodName}() {
    // Complete runnable example with imports and variable declarations
    {receiverVar} := New{ReceiverType}({RequiredParams})
    result, err := {receiverVar}.{MethodName}({ParamValues})
    if err != nil {
        // Error handling with specific conditions
        log.Fatalf("Failed to {action}: %v", err)
        return
    }

    // Example using the result
    fmt.Printf("Result: %v\n", result)
}
~~~

Edge Cases:
- {Edge case 1}: {EXACTLY 1 sentence on how it's handled}
- {Edge case 2}: {EXACTLY 1 sentence on how it's handled}
- {Edge case 3 (if applicable)}: {EXACTLY 1 sentence on how it's handled}

## Testing Strategy

### Test Coverage Goals
- Unit test coverage: {EXACTLY a specific percentage e.g., 85%} of all functions and methods
- Integration test coverage: {EXACTLY a specific percentage e.g., 70%} of component interactions
- Edge case coverage: {EXACTLY a specific percentage e.g., 95%} of all identified edge cases

### Test Cases
1. {TestCase1}: {EXACTLY 1 sentence describing what is being tested}
2. {TestCase2}: {EXACTLY 1 sentence describing what is being tested}
3. {TestCase3}: {EXACTLY 1 sentence describing what is being tested}
4. {TestCase4}: {EXACTLY 1 sentence describing what is being tested}
5. {TestCase5}: {EXACTLY 1 sentence describing what is being tested}
6. {EdgeCase1}: {EXACTLY 1 sentence describing the edge case test}
7. {EdgeCase2}: {EXACTLY 1 sentence describing the edge case test}

### Test Approach
1. {Methodology1}: {EXACTLY 1-2 sentences describing this testing approach}
2. {Methodology2}: {EXACTLY 1-2 sentences describing this testing approach}
3. {Methodology3}: {EXACTLY 1-2 sentences describing this testing approach}

### Mock Strategies
1. {MockStrategy1}: {EXACTLY 1-2 sentences describing this mocking approach}
2. {MockStrategy2}: {EXACTLY 1-2 sentences describing this mocking approach}

## Performance Considerations

- Optimizations:
  1. {Optimization1}: {EXACTLY 1-2 sentences describing the optimization technique}
  2. {Optimization2}: {EXACTLY 1-2 sentences describing the optimization technique}
  3. {Optimization3}: {EXACTLY 1-2 sentences describing the optimization technique}

- Anti-patterns:
  1. {AntiPattern1}: {EXACTLY 1-2 sentences describing the issue and why to avoid it}
  2. {AntiPattern2}: {EXACTLY 1-2 sentences describing the issue and why to avoid it}
  3. {AntiPattern3}: {EXACTLY 1-2 sentences describing the issue and why to avoid it}

## Concurrency Considerations

- Thread Safety:
  - Component thread-safety: {EXACTLY "This component is thread-safe" OR "This component is not thread-safe"}
  - {SafetyMechanism1}: {EXACTLY 1 sentence explaining this mechanism}
  - {SafetyMechanism2}: {EXACTLY 1 sentence explaining this mechanism}

- Synchronization:
  - {SyncApproach1}: {EXACTLY 1-2 sentences describing this synchronization approach}
  - {SyncApproach2}: {EXACTLY 1-2 sentences describing this synchronization approach}

- Goroutine Management:
  - {Pattern1}: {EXACTLY 1-2 sentences describing goroutine usage pattern}
  - {Pattern2}: {EXACTLY 1-2 sentences describing goroutine usage pattern}

## Security Considerations

1. Input Validation: {EXACTLY 2-3 sentences describing the input validation strategy}

2. Resource Management: {EXACTLY 2-3 sentences describing the resource management strategy}

3. Data Protection: {EXACTLY 2-3 sentences describing the data protection strategy}

## Documentation Quality Checklist

Before submitting, verify your documentation meets these requirements:

✅ All sections are present in the exact order specified
✅ Every type, method, and function is fully documented
✅ All code examples are complete and executable
✅ Error handling follows Go best practices with specific error conditions
✅ Constraints and edge cases are explicitly documented
✅ Performance and concurrency patterns are specifically addressed
✅ Security considerations include concrete implementation examples