# Deno Migration Guide

This document outlines the migration from Node.js + npm/pnpm + Vitest to pure Deno.

## Overview

The project has been migrated to use Deno's native testing and module system. This provides:
- TypeScript-native development without build tools
- Standard library testing utilities  
- Simplified dependency management
- Better performance and security

## Project Structure

```
moq-web/
├── mod.ts                 # Main entry point (replaces src/index.ts for Deno)
├── deps.ts                # Centralized dependency management
├── deno.json              # Deno configuration and tasks
├── src/                   # Source files
│   ├── **/*.ts           # Implementation files
│   └── **/*_test.ts      # Test files (renamed from *.test.ts)
```

## Configuration

### deno.json

The `deno.json` file configures the Deno environment and defines tasks:

- `deno task test` - Run all tests
- `deno task test:watch` - Run tests in watch mode
- `deno task coverage` - Generate coverage report
- `deno task coverage:html` - Generate HTML coverage report

### deps.ts

Centralized dependency file that re-exports:
- Deno standard library testing utilities (`@std/assert`, `@std/testing/bdd`)
- External dependencies (golikejs via npm:)

## Migration Changes

### 1. Test File Naming

- **Before**: `*.test.ts`
- **After**: `*_test.ts`

Example: `queue.test.ts` → `queue_test.ts`

### 2. Import Statements

**Before (Vitest)**:
```typescript
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { Queue } from './queue';
```

**After (Deno)**:
```typescript
import { describe, it, beforeEach, afterEach, assertEquals, assertExists } from "../deps.ts";
import { Queue } from "./queue.ts";
```

Key changes:
- Import from `deps.ts` instead of 'vitest'
- Add `.ts` extension to all relative imports
- Use specific assertion functions instead of `expect()`

### 3. Assertions

Vitest uses `expect()` style assertions, Deno uses function-based assertions:

| Vitest | Deno |
|--------|------|
| `expect(a).toBe(b)` | `assertEquals(a, b)` |
| `expect(a).toEqual(b)` | `assertEquals(a, b)` |
| `expect(a).toBeDefined()` | `assertExists(a)` |
| `expect(a).toBeUndefined()` | `assertEquals(a, undefined)` |
| `expect(a).toBeNull()` | `assertEquals(a, null)` |
| `expect(a).toBeTruthy()` | `assertEquals(!!a, true)` |
| `expect(a).toBeFalsy()` | `assertEquals(!!a, false)` |
| `expect(() => fn()).toThrow()` | `assertThrows(() => fn())` |
| `expect(a).toBeInstanceOf(B)` | `assertInstanceOf(a, B)` |

### 4. Mocking

Vitest mocking (`vi.mock()`, `vi.fn()`) needs to be replaced with Deno-compatible patterns.

For simple function mocks, use the custom `createMock` utility:

**Before (Vitest)**:
```typescript
import { vi } from 'vitest';

const mockFn = vi.fn().mockReturnValue(42);
const mockAsync = vi.fn().mockResolvedValue('result');
const mockImpl = vi.fn().mockImplementation((x) => x * 2);
```

**After (Deno)**:
```typescript
import { createMock } from "../deps.ts";

const mockFn = createMock<() => number>().mockReturnValue(42);
const mockAsync = createMock<() => Promise<string>>().mockResolvedValue("result");
const mockImpl = createMock<(x: number) => number>().mockImplementation((x) => x * 2);
```

For module mocking (`vi.mock()`), use manual mocks by creating object literals with mock functions:

**Before**:
```typescript
vi.mock('./module');
```

**After**:
```typescript
// Create manual mocks inline
const mockModule = {
  method1: createMock<() => void>(),
  method2: createMock<(x: number) => number>().mockReturnValue(42),
};
```

Note: Deno doesn't have built-in module mocking like Vitest. For complex mocking scenarios, consider:
- Dependency injection patterns
- Manual mocks
- Interface-based testing
- [Deno Mock library](https://deno.land/x/mock) for advanced cases


### 5. Source File Imports

All source files must use explicit `.ts` extensions:

**Before**:
```typescript
import { Reader } from './io';
import type { Context } from '../deps.ts';
```

**After**:
```typescript
import { Reader } from './io/index.ts';
import type { Context } from '../deps.ts';  // npm: prefix handled in deno.json
```

## Running Tests

### With Deno

```bash
# Run all tests
deno task test

# Run specific test file
deno test --allow-all src/queue_test.ts

# Run tests in watch mode
deno task test:watch

# Generate coverage
deno task coverage
```

### Development Workflow

1. Write test first (`*_test.ts` file)
2. Implement feature
3. Run `deno task test:watch` for immediate feedback
4. Refactor with confidence

## Removed Files

The following Node.js-specific files have been removed:
- `package.json`
- `pnpm-lock.yaml` / `package-lock.json`
- `node_modules/`
- `vitest.config.ts`
- `vitest.setup.ts`
- `tsconfig.json` (Deno uses deno.json)
- `tsconfig.test.json`
- `tsconfig.browser.json`
- `eslint.config.js` (use `deno lint` instead)

## Dependencies

### External Dependencies

The project uses `golikejs` for Go-like concurrency primitives:
- `../deps.ts` - Context management
- `golikejs/sync` - Mutex, RWMutex, WaitGroup

These are imported via npm: protocol in `deno.json`.

### Deno Standard Library

Testing utilities from Deno's standard library:
- `@std/assert` - Assertion functions
- `@std/testing/bdd` - BDD-style testing (describe/it)

## Migration Checklist

- [x] Create `deno.json` configuration
- [x] Create `mod.ts` entry point
- [x] Create `deps.ts` for centralized dependencies
- [ ] Rename all `*.test.ts` files to `*_test.ts`
- [ ] Convert all Vitest imports to Deno imports
- [ ] Replace all `expect()` assertions with Deno assertions
- [ ] Add `.ts` extensions to all relative imports in source files
- [ ] Remove mocking code and replace with Deno-compatible alternatives
- [ ] Run `deno task test` to verify all tests pass
- [ ] Remove Node.js-specific files
- [ ] Update CI/CD to use Deno commands

## CI/CD Updates

Update GitHub Actions workflow to use Deno:

```yaml
- name: Setup Deno
  uses: denoland/setup-deno@v1
  with:
    deno-version: v1.x

- name: Run tests
  run: deno task test

- name: Generate coverage
  run: deno task coverage
```

## Troubleshooting

### Import errors
Make sure all relative imports include `.ts` extension.

### Type errors
Deno is stricter about types. Use explicit type imports with `import type` where needed.

### Network errors during test
Ensure `--allow-net` flag is included if tests make network calls, or use `--allow-all` during development.

## Resources

- [Deno Manual](https://deno.land/manual)
- [Deno Standard Library](https://deno.land/std)
- [Testing in Deno](https://deno.land/manual/testing)
- [BDD Testing](https://deno.land/std/testing/bdd.ts)
