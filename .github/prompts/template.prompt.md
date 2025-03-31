# Go File Generation Prompt Template

This template helps you create prompts to effectively generate Go files (XXX.go) and their test files (XXX_test.go). You can also generate comprehensive documentation. Replace the placeholders with your specific requirements.

## Basic Prompt Structure

```
# Go File Generation: [Filename]

## Functional Requirements

Please generate a Go file and its test file with the following functionality:
[Description of purpose and features]

### Key Features
- [Feature_1]: [Description]
- [Feature_2]: [Description]
- ...

### Technical Constraints
- Go Version: [GO_Version]
- Performance Requirements: [Performance Requirements]
- Error Handling Strategy: [Error Handling Strategy]

Please implement with appropriate tests and detailed documentation.
```

---

## 1. File Specifications

### 1.1 Purpose and Scope
- **Clearly Define**: What problem does this Go file solve?
- **Example**: "Create a file that efficiently validates and processes JSON configuration files with strict type checking"

### 1.2 Input/Output Expectations
- **Input Type**: What kind of data does this code receive?
- **Output Type**: What kind of data does this code generate?
- **Example**: "Receives a file path string and returns a structured Config object or appropriate error"

### 1.3 Technical Requirements
- Go version compatibility
- Allowed/prohibited external dependencies
- Performance expectations
- Memory constraints

---

## 2. Architecture Guidelines

### 2.1 Design Patterns
- Suggest recommended design patterns when applicable
- **Example**: "Use the repository pattern for data access to ensure clear separation of concerns"

### 2.2 Code Structure
- Package structure
- File naming conventions
- Interface definitions
- **Example**: "Create a clean API with a facade interface that hides implementation details"

---

## 3. Function Specifications

For each key function, provide the following:

```
### Function: [Function Name]

Description: [Detailed Description]

Signature:
```go
func FunctionName(param1 Type, param2 Type) (ReturnType, error)
```

Parameters:
- param1: [Description of parameter]
- param2: [Description of parameter]

Return Values:
- [Description of return value]
- error: [Error conditions]

Usage Example:
```go
result, err := FunctionName(value1, value2)
if err != nil {
    // Error handling
}
// Use the result
```

Edge Cases to Handle:
- [Edge Case_1]
- [Edge Case_2]
```

---

## 4. Test Requirements

### 4.1 Test Coverage Goals
- Minimum test coverage percentage
- Important paths to test
- **Example**: "Focus on error handling paths and ensure at least 80% code coverage"

### 4.2 Test Cases
- List of test scenarios to implement
- Edge cases to validate
- **Example**:
  ```
  - Test successful parsing of valid input
  - Test appropriate errors for invalid input formats
  - Test handling of empty/nil values
  - Test performance with large inputs
  ```

### 4.3 Test Approach
- Unit test structure
- Mock requirements
- Benchmark expectations

---

## 5. Documentation Requirements

### 5.1 Code Documentation
- Comment style guidelines
- Documentation format (godoc conventions)
- **Example**: "All exported functions should have documentation comments following Go standard practices"

### 5.2 Usage Examples
- Request specific usage examples
- **Example**: "Include examples of common use cases, including error handling"

### 5.3 README Content
- Installation instructions
- Quick start guide
- API overview

---

## 6. Implementation Constraints

### 6.1 Performance Considerations
- Time complexity requirements
- Memory usage limits
- Concurrency expectations
- **Example**: "Key functions should have O(n) time complexity and be safe for concurrent use"

### 6.2 Error Handling Strategy
- Types of errors to use
- Error wrapping approach
- Logging expectations
- **Example**: "Use custom error types for specific error conditions. Wrap errors with context information"

### 6.3 Security Requirements
- Input validation approach
- Security best practices
- **Example**: "Implement strict input validation. Do not use user input directly in file paths"

---

## 7. Prompt Example

```
# Go File Generation: configparser.go

## Functional Requirements

Please generate a Go file and its test file with the following functionality:
[Description of purpose and features]

### Key Features
- LoadConfig: Load and parse the config file from the specified path
- ValidateConfig: Validate the config against a schema
- GetConfigValue: Type-safe accessor for config values

### Technical Constraints
- Go Version: 1.18 or higher
- Performance Requirements: Efficiently handle configs up to 10MB
- Error Handling Strategy: Custom error types with detailed context

Please implement with appropriate tests and detailed documentation.
```

---

## 8. Expected Output

Clearly state the expected deliverables:

1. **Main Code File (XXX.go)**: Fully implemented Go file with appropriate error handling and documentation
2. **Test File (XXX_test.go)**: Comprehensive tests covering functionality and edge cases
3. **Documentation**: API description and usage examples
4. **Usage Examples**: Simple examples demonstrating the code's functionality

---

This template can be used to create detailed prompts for generating Go files and their test files. Adjust the sections based on specific requirements. A structured prompt reduces the need for revisions and clarifications, leading to higher quality code generation.