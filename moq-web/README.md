# @okudai/moq

[![JSR](https://jsr.io/badges/@okudai/moq)](https://jsr.io/@okudai/moq)

A TypeScript/JavaScript implementation of Media over QUIC Transport (MoQT) for
Deno environments.

This library enables clients to connect and communicate with the Go
implementation of the Media over QUIC protocol (gomoqt).

## Installation

### For Deno

Install from [JSR](https://jsr.io/@okudai/moq):

```bash
deno add jsr:@okudai/moq
```

Then import:

```typescript
import { Session } from "@okudai/moq";
```

### For Node.js (npm)

```bash
npx jsr add @okudai/moq
```

## Development

This project uses Deno as its primary development environment, providing
TypeScript-native development with built-in testing, formatting, and linting.

### Prerequisites

- [Deno](https://deno.land/) v2.0 or later

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
├── deno.json           # Deno configuration
├── src/                # Source files
│   ├── mod.ts         # Main entry point
│   ├── **/*.ts        # Implementation
│   └── **/*_test.ts   # Tests (Deno convention)
```

### Testing

Tests use Deno's standard testing library with BDD-style syntax:

```typescript
import { assertEquals } from "@std/assert";
import { describe, it } from "@std/testing/bdd";

describe("MyFeature", () => {
	it("should work correctly", () => {
		assertEquals(1 + 1, 2);
	});
});
```

## License

MIT
