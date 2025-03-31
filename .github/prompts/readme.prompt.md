# Package README Generation Prompt

Please generate a detailed README.md file based on the following package information.

## Package Information
Package Name: {{package_name}}
Package Path: {{package_path}}
Parent Package: {{parent_package_name (if applicable)}}

## Prompt

Please generate a README.md with the following format and content:

```markdown
# {{package_name}} Package

## Overview
Please provide a concise explanation of the main purpose and functionality of this package. Clarify why this package exists and its role within the overall system.

## Responsibilities
List the main responsibilities of this package in bullet points.

- Responsibility 1
- Responsibility 2
- Responsibility 3

## Key Interfaces and Components
Explain the main interfaces, structs, functions, etc. provided by this package.

### Interface Name/Struct Name/Function Name
Description and usage.

### Interface Name/Struct Name/Function Name
Description and usage.

## Interaction with Other Packages
Explain how this package interacts with other packages. Clarify dependencies and call relationships.

### Dependencies (packages this package depends on)
- Package name: Reason for dependency and usage

### Dependents (packages that depend on this package)
- Package name: How this package is used by the dependent

## Implementation Notes
Explain important considerations, design principles, patterns, etc. when implementing this package.

## Testing Strategy
Explain the recommended approach for testing this package.

## Future Extensibility
Explain how this package might be extended in the future.

## References
Provide links to reference materials or documentation related to the implementation of this package.
```

Please generate a README.md with specific information for the given package following the template above. Clearly define the package's purpose and responsibilities, and explain in detail its relationship with other packages. The content should help implementers understand the overall picture of the package and implement it appropriately.
