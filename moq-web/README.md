# @okutanidaichi/moqt

A TypeScript/JavaScript implementation of Media over QUIC Transport (MoQT) for both Deno and Node.js environments.

This library enables clients to connect and communicate with the Go implementation of the Media over QUIC protocol (gomoqt).

## Installation

### For Deno

Import directly from the repository or JSR (when published):

```typescript
import { Session } from "https://deno.land/x/moqt/mod.ts";
// or
import { Session } from "jsr:@okutanidaichi/moqt";
```

### For Node.js (npm)

```bash
npm install @okutanidaichi/moqt
```

## Development

This project uses Deno as its primary development environment, providing TypeScript-native development with built-in testing, formatting, and linting.

### Prerequisites

- [Deno](https://deno.land/) v1.40 or later

### Getting Started

```bash
# Run tests
deno task test

# Run tests in watch mode
deno task test:watch

# Generate coverage report
deno task coverage

# Generate HTML coverage report
deno task coverage:html
```

### Project Structure

```
moq-web/
├── mod.ts              # Main entry point
├── deps.ts             # Centralized dependencies
├── deno.json           # Deno configuration
├── src/                # Source files
│   ├── **/*.ts        # Implementation
│   └── **/*_test.ts   # Tests (Deno convention)
└── DENO_MIGRATION.md   # Detailed migration documentation
```

### Testing

Tests use Deno's standard testing library with BDD-style syntax:

```typescript
import { assertEquals, describe, it } from "../deps.ts";

describe("MyFeature", () => {
	it("should work correctly", () => {
		assertEquals(1 + 1, 2);
	});
});
```

## Migration from Node.js

This project was recently migrated from Node.js + Vitest to pure Deno. For detailed migration notes and patterns, see [DENO_MIGRATION.md](./DENO_MIGRATION.md).

Key changes:

- Test files renamed from `*.test.ts` to `*_test.ts`
- Vitest replaced with Deno standard library testing utilities
- All imports now include explicit `.ts` extensions
- No build step required for development

## License

MIT
