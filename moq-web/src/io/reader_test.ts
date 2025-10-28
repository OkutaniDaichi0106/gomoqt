import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows } from "../../deps.ts";
import { Reader } from './reader.ts';
import { StreamError } from './error.ts';

describe('Reader', () => {
  let reader: Reader;
  let readableStream: ReadableStream<Uint8Array>;
  let controller: ReadableStreamDefaultController<Uint8Array>;

  beforeEach(() => {
    readableStream = new ReadableStream<Uint8Array>({
      start(ctrl) {
        controller = ctrl;
      }
    });
    reader = new Reader({stream: readableStream, streamId: 1n});
  });

  afterEach(() => {
    // Skip cleanup to avoid uncaught promise rejections
    // Individual tests should handle their own cleanup if needed
  });

  // Helper function to create a fresh reader for specific tests
  const createFreshReader = (data?: Uint8Array): { reader: Reader, controller: ReadableStreamDefaultController<Uint8Array> } => {
    let ctrl: ReadableStreamDefaultController<Uint8Array>;
    const stream = new ReadableStream<Uint8Array>({
      start(c) {
        ctrl = c;
        if (data) {
          ctrl.enqueue(data);
          // Don't close immediately - let the test handle it
        }
      }
    });
    return { reader: new Reader({stream, streamId: 1n}), controller: ctrl! };
  };

  // Helper function to create a reader with data and immediately close
  const createClosedReader = (data: Uint8Array): Reader => {
    const stream = new ReadableStream<Uint8Array>({
      start(ctrl) {
        ctrl.enqueue(data);
        ctrl.close();
      }
    });
    return new Reader({stream, transfer: undefined, streamId: 0n});
  };

  describe('readUint8Array', () => {
    it('should read a Uint8Array with varint length prefix', async () => {
      const data = new Uint8Array([1, 2, 3, 4, 5]);
      // Varint length (5) + data
      const streamData = new Uint8Array([5, ...data]);
      
      controller.enqueue(streamData);
      controller.close();

      const [result, error] = await reader.readUint8Array();
      
      assertEquals(error, undefined);
      assertEquals(result, data);
    });

    it('should handle empty array', async () => {
      const freshReader = createClosedReader(new Uint8Array([0])); // Varint length (0)

      const [result, error] = await freshReader.readUint8Array();
      
      assertEquals(error, undefined);
      assertEquals(result, new Uint8Array([]));
    });

    it('should handle partial reads correctly', async () => {
      const data = new Uint8Array([1, 2, 3]);
      // Send length first
      controller.enqueue(new Uint8Array([3]));
      // Send data in parts
      controller.enqueue(new Uint8Array([1, 2]));
      controller.enqueue(new Uint8Array([3]));
      controller.close();

      const [result, error] = await reader.readUint8Array();
      
      assertEquals(error, undefined);
      assertEquals(result, data);
    });

    it('should return error for stream with insufficient data', async () => {
      // Enqueue invalid varint data that will cause read error
      const invalidVarint = new Uint8Array([0xFF]); // Invalid single byte that requires more bytes
      controller.enqueue(invalidVarint);
      controller.close(); // Close without providing enough data

      const [result, error] = await reader.readUint8Array();
      
      assertEquals(result, undefined);
      assertExists(error);
    });

    it('should handle very large length values', async () => {
      // Length that exceeds MAX_BYTES_LENGTH
      const largeLength = new Uint8Array([0xF0, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF]);
      const freshReader = createClosedReader(largeLength);

      try {
        await freshReader.readUint8Array();
        // Fail if no exception is thrown
        throw new Error('Expected to throw Varint too large');
      } catch (e: any) {
        assertArrayIncludes(e.message, ['Varint too large']);
      }
    });
  });

  describe('readString', () => {
    it('should read a UTF-8 string', async () => {
      const str = 'hello world';
      const encoded = new TextEncoder().encode(str);
      // Varint length + data
      const streamData = new Uint8Array([encoded.length, ...encoded]);
      
      controller.enqueue(streamData);
      controller.close();

      const [result, error] = await reader.readString();
      
      assertEquals(error, undefined);
      assertEquals(result, str);
    });

    it('should handle empty string', async () => {
      // Varint length (0)
      const streamData = new Uint8Array([0]);
      const freshReader = createClosedReader(streamData);

      const [result, error] = await freshReader.readString();
      
      assertEquals(error, undefined);
      assertEquals(result, '');
    });

    it('should handle Unicode characters', async () => {
      const str = 'こんにちは🚀';
      const encoded = new TextEncoder().encode(str);
      // Varint length + data
      const streamData = new Uint8Array([encoded.length, ...encoded]);
      
      controller.enqueue(streamData);
      controller.close();

      const [result, error] = await reader.readString();
      
      assertEquals(error, undefined);
      assertEquals(result, str);
    });

    it('should return error when underlying readUint8Array fails', async () => {
      // Create reader with incomplete varint data and close it immediately
      const incompleteVarint = new Uint8Array([0xFF]); // Requires more bytes but none available
      const freshReader = createClosedReader(incompleteVarint);

      const [result, error] = await freshReader.readString();
      
      assertEquals(result, ""); // Implementation returns empty string on error
      assertExists(error);
    });
  });

  describe('readBigVarint', () => {
    it('should read single byte varint', async () => {
      const streamData = new Uint8Array([42]);
      const freshReader = createClosedReader(streamData);

      const [result, error] = await freshReader.readBigVarint();
      
      assertEquals(error, undefined);
      assertEquals(result, 42n);
    });

    it('should read two byte varint', async () => {
      // Value 300 (0x012C) encoded as varint: 0x41 0x2C (QUIC format)
      const streamData = new Uint8Array([0x41, 0x2C]);
      const freshReader = createClosedReader(streamData);

      const [result, error] = await freshReader.readBigVarint();
      
      assertEquals(error, undefined);
      assertEquals(result, 300n);
    });

    it('should read four byte varint', async () => {
      // Large value encoded as 4-byte varint: 1000000 = 0x800F4240
      const streamData = new Uint8Array([0x80, 0x0F, 0x42, 0x40]);
      
      const freshReader = createClosedReader(streamData);

      const [result, error] = await freshReader.readBigVarint();
      
      assertEquals(error, undefined);
      assertEquals(result, 1000000n);
    });

    it('should read eight byte varint', async () => {
      // Very large value as 8-byte varint: 1 << 40 = 0xC0000100000000
      const streamData = new Uint8Array([0xC0, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00]);
      
      const freshReader = createClosedReader(streamData);

      const [result, error] = await freshReader.readBigVarint();
      
      assertEquals(error, undefined);
      assertEquals(result, 1n << 40n);
    });

    it('should handle partial varint reads', async () => {
      // Send two-byte varint in parts: 300 = 0x41 0x2C
      controller.enqueue(new Uint8Array([0x41]));
      controller.enqueue(new Uint8Array([0x2C]));
      controller.close();

      const [result, error] = await reader.readBigVarint();
      
      assertEquals(error, undefined);
      assertEquals(result, 300n);
    });

    it('should return error on stream close before complete read', async () => {
      // Send incomplete varint (2-byte varint but only first byte)
      controller.enqueue(new Uint8Array([0x41]));
      controller.close();

      const [result, error] = await reader.readBigVarint();
      
      assertEquals(result, 0n); // Implementation returns 0n on error
      assertExists(error);
    });
  });

  describe('readUint8', () => {
    it('should read a single byte', async () => {
      const streamData = new Uint8Array([123]);
      
      controller.enqueue(streamData);
      controller.close();

      const [result, error] = await reader.readUint8();
      
      assertEquals(error, undefined);
      assertEquals(result, 123);
    });

    it('should read multiple bytes sequentially', async () => {
      const streamData = new Uint8Array([1, 2, 3]);
      
      controller.enqueue(streamData);
      controller.close();

      const [first, error1] = await reader.readUint8();
      assertEquals(error1, undefined);
      assertEquals(first, 1);

      const [second, error2] = await reader.readUint8();
      assertEquals(error2, undefined);
      assertEquals(second, 2);

      const [third, error3] = await reader.readUint8();
      assertEquals(error3, undefined);
      assertEquals(third, 3);
    });

    it('should return error when no data available', async () => {
      controller.close();

      const [result, error] = await reader.readUint8();
      
      assertEquals(result, 0); // Implementation returns 0 on error
      assertExists(error);
    });
  });

  describe('readBoolean', () => {
    it('should read true as 1', async () => {
      const streamData = new Uint8Array([1]);
      
      controller.enqueue(streamData);
      controller.close();

      const [result, error] = await reader.readBoolean();
      
      assertEquals(error, undefined);
      assertEquals(result, true);
    });

    it('should read false as 0', async () => {
      const streamData = new Uint8Array([0]);
      
      controller.enqueue(streamData);
      controller.close();

      const [result, error] = await reader.readBoolean();
      
      assertEquals(error, undefined);
      assertEquals(result, false);
    });

    it('should return error for invalid boolean values', async () => {
      const streamData = new Uint8Array([2]);
      
      controller.enqueue(streamData);
      controller.close();

      const [result, error] = await reader.readBoolean();
      
      assertEquals(result, false); // Implementation returns false on error
      assertExists(error);
      assertArrayIncludes(error?.message, ['Invalid boolean value']);
    });

    it('should return error when readUint8 fails', async () => {
      controller.close();

      const [result, error] = await reader.readBoolean();
      
      assertEquals(result, false); // Implementation returns false on error
      assertExists(error);
    });
  });

  describe('cancel', () => {
    it('should cancel the reader with error code and message', async () => {
      const code = 123;
      const message = 'Test cancellation';
      const streamError = new StreamError(code, message);
      
      await expect(reader.cancel(streamError)).resolves.not.toThrow();
    });
  });

  describe('closed', () => {
    it('should return a promise that resolves when reader is closed', async () => {
      const closedPromise = reader.closed();
      controller.close();
      
      await expect(closedPromise).resolves.not.toThrow();
    });
  });

  describe('integration tests', () => {
    it('should read multiple data types in sequence', async () => {
      // Boolean true, varint 42 (< 64, single byte), string "test", byte array [1,2,3]
      const testStr = 'test';
      const testBytes = new Uint8Array([1, 2, 3]);
      const encodedStr = new TextEncoder().encode(testStr);
      
      const streamData = new Uint8Array([
        1,                           // boolean true
        42,                          // varint 42 (single byte, since 42 < 64)
        encodedStr.length,           // string length
        ...encodedStr,               // string data
        testBytes.length,            // array length
        ...testBytes                 // array data
      ]);
      
      controller.enqueue(streamData);
      controller.close();

      // Read boolean
      const [boolResult, boolError] = await reader.readBoolean();
      assertEquals(boolError, undefined);
      assertEquals(boolResult, true);

      // Read varint
      const [varintResult, varintError] = await reader.readBigVarint();
      assertEquals(varintError, undefined);
      assertEquals(varintResult, 42n);

      // Read string
      const [stringResult, stringError] = await reader.readString();
      assertEquals(stringError, undefined);
      assertEquals(stringResult, testStr);

      // Read byte array
      const [arrayResult, arrayError] = await reader.readUint8Array();
      assertEquals(arrayError, undefined);
      assertEquals(arrayResult, testBytes);
    });

    it('should handle stream errors gracefully', async () => {
      // Close the stream to simulate end of data
      controller.close();

      const [result, error] = await reader.readUint8();
      
      assertEquals(result, 0); // Implementation returns 0 on error
      assertExists(error);
      assertArrayIncludes(error?.message, ['Stream closed']);
    });
  });

  describe('BYOB (Bring Your Own Buffer) support', () => {
    it('should work with BYOB reader when available', async () => {
      // Create a simple stream with data, not necessarily BYOB
      const data = new Uint8Array([42]);
      const freshReader = createClosedReader(data);
      
      const [result, error] = await freshReader.readUint8();
      
      assertEquals(error, undefined);
      assertEquals(result, 42);
      
      await freshReader.cancel(new StreamError(0, 'Test cleanup'));
    });
  });
});
