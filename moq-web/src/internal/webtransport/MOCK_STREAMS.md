# Mock Streams for Testing

This module provides mock implementations of WebTransport streams for testing purposes.

## Classes

### MockSendStream

Mock implementation of `SendStream` (write operations).

**Features:**
- Spy on all write methods (writeVarint, writeString, etc.)
- Track flush, close, and cancel calls
- Simulate flush errors

**Example:**
```typescript
import { MockSendStream } from "./mock_stream_test.ts";

const mock = new MockSendStream();

// Use in your code
mock.writeString("hello");
await mock.flush();

// Verify in tests
assertEquals(mock.flushCalls, 1);
```

### MockReceiveStream

Mock implementation of `ReceiveStream` (read operations).

**Features:**
- Spy on all read methods
- Provide mock data via `data` array
- Customize behavior with `*Impl` properties (e.g., `readVarintImpl`)
- Simulate read errors

**Example:**
```typescript
import { MockReceiveStream } from "./mock_stream_test.ts";

const mock = new MockReceiveStream();

// Setup mock data
mock.data = [
  new Uint8Array([5]),          // First read
  new Uint8Array([1, 2, 3])     // Second read
];

// Use in your code
const [value] = await mock.readVarint();
assertEquals(value, 5);
```

**Custom Implementations:**
```typescript
// Override specific read methods
mock.readVarintImpl = async () => {
  // Custom logic
  return [42, undefined];
};

const [value] = await mock.readVarint();
assertEquals(value, 42);
```

### MockStream

Mock implementation of bidirectional `Stream` (combines send and receive).

**Features:**
- Wraps both `MockSendStream` and `MockReceiveStream`
- Provides unified stream ID
- Reset both sides at once

**Example:**
```typescript
import { MockStream } from "./mock_stream_test.ts";

// Create bidirectional stream with ID
const mock = new MockStream(123n);

// Access writable side
await mock.writable.flush();
assertEquals(mock.writable.flushCalls, 1);

// Access readable side
mock.readable.data = [new Uint8Array([1, 2, 3])];
const [data] = await mock.readable.readUint8Array();

// Reset both sides
mock.reset();
```

## Usage in Tests

### Basic Pattern

```typescript
import { MockStream } from "./internal/webtransport/mock_stream_test.ts";

Deno.test("my test", async () => {
  const mockStream = new MockStream(42n);
  
  // Configure mock behavior
  mockStream.writable.flushError = new Error("Flush failed");
  
  // Use in your code
  const result = await myFunction(mockStream);
  
  // Verify behavior
  assertEquals(mockStream.writable.flushCalls > 0, true);
});
```

### Error Simulation

```typescript
// Simulate flush error
mockStream.writable.flushError = new Error("Network error");

// Simulate read error
mockStream.readable.readError = new Error("Parse error");
```

### Call Tracking

```typescript
// Check number of calls
assertEquals(mock.writable.flushCalls, 3);
assertEquals(mock.writable.closeCalls, 1);

// Check cancel calls with error details
assertEquals(mock.writable.cancelCalls.length, 1);
assertInstanceOf(mock.writable.cancelCalls[0], StreamError);
```

## Direct Construction

```typescript
import {
  MockStream,
  MockSendStream,
  MockReceiveStream,
} from "./mock_stream_test.ts";

// Create individual streams
const sendStream = new MockSendStream();
const receiveStream = new MockReceiveStream();

// Create bidirectional stream
const stream = new MockStream(123n);
```

## Testing Best Practices

1. **Reset between tests**: Call `mock.reset()` to clear call counts
2. **Configure before use**: Set up mock data and errors before passing to code under test
3. **Verify calls**: Check that expected methods were called with `*Calls` properties
4. **Use type assertions**: Cast to `any` when passing mocks to functions expecting real streams

```typescript
// Example test structure
Deno.test("feature test", async () => {
  // Setup
  const mock = new MockStream(1n);
  mock.readable.data = [...];
  
  // Execute
  const result = await featureUnderTest(mock as any);
  
  // Verify
  assertEquals(result, expectedValue);
  assertEquals(mock.writable.flushCalls, 1);
  
  // Cleanup (if needed)
  mock.reset();
});
```

## Related Files

- `mock_stream_test.ts` - Mock implementations
- `mock_stream.test.ts` - Unit tests for mocks
- `subscribe_stream_test.ts` - Example usage in real tests
- `group_stream_test.ts` - Example usage in real tests
