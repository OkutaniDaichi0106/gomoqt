# Mock Utilities Implementation

## Overview

This document describes the new mock utilities implementation for the moq-web project, which transitions from using `test-utils/mock.ts` to using Deno's standard library `std/testing/mock`.

## Files Created

### 1. `src/mock/index.ts`
Location: `c:\Users\daich\OneDrive\gomoqt\moq-web\src\mock\index.ts`

This file serves as the central mock utility module for the project. It:
- Re-exports all core mocking functions from `@std/testing/mock` (JSR)
- Provides compatibility wrappers (`createMock`, `createSpy`) for easier transition
- Maintains API compatibility with the old `test-utils/mock.ts`

**Key Exports:**
- Core std/testing/mock functions: `spy`, `stub`, `assertSpyCall`, `assertSpyCalls`, etc.
- Compatibility functions: `createMock`, `createSpy`
- Helper functions: `returnsNext`, `returnsThis`, `returnsArg`, `returnsArgs`, `resolvesNext`

### 2. `mock_test.ts`
Location: `c:\Users\daich\OneDrive\gomoqt\moq-web\mock_test.ts`

Comprehensive test file demonstrating all mock capabilities:

**Test Suites:**
1. **createMock** - Simple mock function creation and manipulation
   - Track function calls
   - Set return values
   - Mock async functions
   - Override implementations
   - Reset mock state

2. **createSpy** - Function spying with compatibility wrapper
   - Spy on existing functions
   - Preserve original behavior

3. **std/testing/mock - spy** - Native spy functionality
   - Create spies on functions
   - Create method spies on objects

4. **std/testing/mock - stub** - Replace method behavior
   - Stub methods with custom implementations
   - Use helper functions like `returnsNext`

5. **Complex mocking example** - Real-world usage patterns
   - Mock logger in service classes
   - Test method calls with arguments

6. **Async mocking** - Async function mocking
   - Mock async functions
   - Handle multiple async calls

## Usage

### Simple Mock Functions
```typescript
import { createMock } from "./src/mock/index.ts";

const mockFn = createMock<(x: number) => number>();
mockFn.mockReturnValue(42);
const result = mockFn(10);

// Check calls
console.log(mockFn.calls); // [[10]]
console.log(mockFn.results); // [42]
```

### Spying on Object Methods (std/testing/mock)
```typescript
import { spy } from "./src/mock/index.ts";

const obj = {
  getValue: () => 42,
};

using methodSpy = spy(obj, "getValue");
obj.getValue();

// Assertions
import { assertSpyCalls } from "./src/mock/index.ts";
assertSpyCalls(methodSpy, 1);
```

### Stubbing Methods (std/testing/mock)
```typescript
import { stub, returnsNext } from "./src/mock/index.ts";

const obj = {
  calculate: (a: number, b: number) => a + b,
};

using calculateStub = stub(obj, "calculate", () => 100);
obj.calculate(5, 5); // Returns 100
```

## Migration Notes

- **Backward Compatibility:** The `test-utils/mock.ts` is not being replaced yet, so existing tests continue to work
- **New Code:** Start using `src/mock/index.ts` for new tests
- **Gradual Migration:** As files are updated, gradually replace old mock imports with new ones
- **Advantages of std/testing/mock:**
  - Native Deno standard library support
  - Better TypeScript support
  - More powerful features (stubs, constructor spies, etc.)
  - Active maintenance by Deno core team

## Test Results

All tests pass successfully:
```
ok | 6 passed (15 steps) | 0 failed (37ms)
```

## Running Tests

```bash
deno test mock_test.ts --unstable-kv --unstable-ffi --reload
```

## Future Work

- Replace usage of `test-utils/mock.ts` with `src/mock/index.ts` in other test files
- Add more specialized mock builders if needed
- Consider adding mock fixtures for common patterns
