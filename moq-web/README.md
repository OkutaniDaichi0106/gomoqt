# @okudai/moq

[![JSR](https://jsr.io/badges/@okudai/moq)](https://jsr.io/@okudai/moq)

TypeScript/JavaScript implementation of Media over QUIC (MOQ Lite)
for Deno (and Node via JSR). It is the web/JS client counterpart to the Go
implementation in this repository.

## Overview

- Targets MOQ Lite over WebTransport
- Works with the Go server (`gomoqt`) for interop testing and browser clients
- Written for Deno first; Node/npm use is supported via JSR shim

## Installation

### Deno

Install from [JSR](https://jsr.io/@okudai/moq):

```bash
deno add jsr:@okudai/moq
```

Then import:

```typescript
import { Session } from "@okudai/moq";
```

### Node.js (npm via JSR)

```bash
npx jsr add @okudai/moq
```

## Usage (minimal)

```ts
import { Session } from "@okudai/moq";

const session = new Session({ url: "https://example.com/interop" });
// TODO: add track subscription/publish based on your app
await session.connect();
```

## Development

This project uses Deno as its primary environment (TypeScript-native with
built-in fmt/lint/test).

- Prerequisite: [Deno](https://deno.land/) v2.0 or later

Common tasks:

```bash
# Format code
deno task fmt      # or: deno fmt

# Lint code
deno task lint     # or: deno lint

# Run tests
deno task test     # or: deno test --allow-all

# Generate coverage
deno task coverage

# Generate HTML coverage report
deno task coverage:html
```

## Project Structure

```
moq-web/
├── deno.json         # Deno configuration
├── src/              # Source files
│   ├── mod.ts        # Main entry point
│   ├── **/*.ts       # Implementation
│   └── **/*_test.ts  # Tests (Deno convention)
```

## License

MIT
