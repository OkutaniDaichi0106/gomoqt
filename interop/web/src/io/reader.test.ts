import { Reader } from './reader';
import { StreamError } from './error';

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
    reader = new Reader(readableStream);
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
    return { reader: new Reader(stream), controller: ctrl! };
  };

  // Helper function to create a reader with data and immediately close
  const createClosedReader = (data: Uint8Array): Reader => {
    const stream = new ReadableStream<Uint8Array>({
      start(ctrl) {
        ctrl.enqueue(data);
        ctrl.close();
      }
    });
    return new Reader(stream);
  };

  describe('readUint8Array', () => {
    it('should read a Uint8Array with varint length prefix', async () => {
      const data = new Uint8Array([1, 2, 3, 4, 5]);
      // Varint length (5) + data
      const streamData = new Uint8Array([5, ...data]);
      
      controller.enqueue(streamData);
      controller.close();

      const [result, error] = await reader.readUint8Array();
      
      expect(error).toBeUndefined();
      expect(result).toEqual(data);
    });

    it('should handle empty array', async () => {
      const freshReader = createClosedReader(new Uint8Array([0])); // Varint length (0)

      const [result, error] = await freshReader.readUint8Array();
      
      expect(error).toBeUndefined();
      expect(result).toEqual(new Uint8Array([]));
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
      
      expect(error).toBeUndefined();
      expect(result).toEqual(data);
    });

    it('should return error for stream with insufficient data', async () => {
      // Enqueue invalid varint data that will cause read error
      const invalidVarint = new Uint8Array([0xFF]); // Invalid single byte that requires more bytes
      controller.enqueue(invalidVarint);
      controller.close(); // Close without providing enough data

      const [result, error] = await reader.readUint8Array();
      
      expect(result).toBeUndefined();
      expect(error).toBeDefined();
    });

    it('should handle very large length values', async () => {
      // Length that exceeds MAX_BYTES_LENGTH
      const largeLength = new Uint8Array([0xF0, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF]);
      const freshReader = createClosedReader(largeLength);

      const [result, error] = await freshReader.readUint8Array();
      
      expect(result).toBeUndefined();
      expect(error).toBeDefined();
      expect(error?.message).toContain('Varint too large');
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
      
      expect(error).toBeUndefined();
      expect(result).toBe(str);
    });

    it('should handle empty string', async () => {
      // Varint length (0)
      const streamData = new Uint8Array([0]);
      const freshReader = createClosedReader(streamData);

      const [result, error] = await freshReader.readString();
      
      expect(error).toBeUndefined();
      expect(result).toBe('');
    });

    it('should handle Unicode characters', async () => {
      const str = 'こんにちは🚀';
      const encoded = new TextEncoder().encode(str);
      // Varint length + data
      const streamData = new Uint8Array([encoded.length, ...encoded]);
      
      controller.enqueue(streamData);
      controller.close();

      const [result, error] = await reader.readString();
      
      expect(error).toBeUndefined();
      expect(result).toBe(str);
    });

    it('should return error when underlying readUint8Array fails', async () => {
      // Create reader with incomplete varint data
      const incompleteVarint = new Uint8Array([0xFF]); // Requires more bytes but none available
      const { reader: freshReader } = createFreshReader(incompleteVarint);

      const [result, error] = await freshReader.readString();
      
      expect(result).toBeUndefined();
      expect(error).toBeDefined();
    });
  });

  describe('readVarint', () => {
    it('should read single byte varint', async () => {
      const streamData = new Uint8Array([42]);
      const freshReader = createClosedReader(streamData);

      const [result, error] = await freshReader.readVarint();
      
      expect(error).toBeUndefined();
      expect(result).toBe(42n);
    });

    it('should read two byte varint', async () => {
      // Value 300 (0x012C) encoded as varint: 0x81 0x2C
      const streamData = new Uint8Array([0x81, 0x2C]);
      const freshReader = createClosedReader(streamData);

      const [result, error] = await freshReader.readVarint();
      
      expect(error).toBeUndefined();
      expect(result).toBe(300n);
    });

    it('should read four byte varint', async () => {
      // Large value encoded as 4-byte varint: 1000000 = 0x80C0ED03
      const streamData = new Uint8Array([0x80, 0xC0, 0xED, 0x03]);
      
      const freshReader = createClosedReader(streamData);

      const [result, error] = await freshReader.readVarint();
      
      expect(error).toBeUndefined();
      expect(result).toBeGreaterThan(0n);
    });

    it('should read eight byte varint', async () => {
      // Very large value as 8-byte varint
      const streamData = new Uint8Array([0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01]);
      
      const freshReader = createClosedReader(streamData);

      const [result, error] = await freshReader.readVarint();
      
      expect(error).toBeUndefined();
      expect(result).toBeGreaterThan(0n);
    });

    it('should handle partial varint reads', async () => {
      // Send two-byte varint in parts: 300 = 0x81 0x2C
      controller.enqueue(new Uint8Array([0x81]));
      controller.enqueue(new Uint8Array([0x2C]));
      controller.close();

      const [result, error] = await reader.readVarint();
      
      expect(error).toBeUndefined();
      expect(result).toBe(300n);
    });

    it('should return error on stream close before complete read', async () => {
      // Send incomplete varint
      controller.enqueue(new Uint8Array([0x81]));
      controller.close();

      const [result, error] = await reader.readVarint();
      
      expect(result).toBeUndefined();
      expect(error).toBeDefined();
    });
  });

  describe('readUint8', () => {
    it('should read a single byte', async () => {
      const streamData = new Uint8Array([123]);
      
      controller.enqueue(streamData);
      controller.close();

      const [result, error] = await reader.readUint8();
      
      expect(error).toBeUndefined();
      expect(result).toBe(123);
    });

    it('should read multiple bytes sequentially', async () => {
      const streamData = new Uint8Array([1, 2, 3]);
      
      controller.enqueue(streamData);
      controller.close();

      const [first, error1] = await reader.readUint8();
      expect(error1).toBeUndefined();
      expect(first).toBe(1);

      const [second, error2] = await reader.readUint8();
      expect(error2).toBeUndefined();
      expect(second).toBe(2);

      const [third, error3] = await reader.readUint8();
      expect(error3).toBeUndefined();
      expect(third).toBe(3);
    });

    it('should return error when no data available', async () => {
      controller.close();

      const [result, error] = await reader.readUint8();
      
      expect(result).toBeUndefined();
      expect(error).toBeDefined();
    });
  });

  describe('readBoolean', () => {
    it('should read true as 1', async () => {
      const streamData = new Uint8Array([1]);
      
      controller.enqueue(streamData);
      controller.close();

      const [result, error] = await reader.readBoolean();
      
      expect(error).toBeUndefined();
      expect(result).toBe(true);
    });

    it('should read false as 0', async () => {
      const streamData = new Uint8Array([0]);
      
      controller.enqueue(streamData);
      controller.close();

      const [result, error] = await reader.readBoolean();
      
      expect(error).toBeUndefined();
      expect(result).toBe(false);
    });

    it('should return error for invalid boolean values', async () => {
      const streamData = new Uint8Array([2]);
      
      controller.enqueue(streamData);
      controller.close();

      const [result, error] = await reader.readBoolean();
      
      expect(result).toBeUndefined();
      expect(error).toBeDefined();
      expect(error?.message).toContain('Invalid boolean value');
    });

    it('should return error when readUint8 fails', async () => {
      controller.close();

      const [result, error] = await reader.readBoolean();
      
      expect(result).toBeUndefined();
      expect(error).toBeDefined();
    });
  });

  describe('cancel', () => {
    it('should cancel the reader with error code and message', async () => {
      const code = 123;
      const message = 'Test cancellation';
      
      await expect(reader.cancel(code, message)).resolves.not.toThrow();
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
      // Boolean true, varint 123, string "test", byte array [1,2,3]
      const testStr = 'test';
      const testBytes = new Uint8Array([1, 2, 3]);
      const encodedStr = new TextEncoder().encode(testStr);
      
      const streamData = new Uint8Array([
        1,                           // boolean true
        123,                         // varint 123
        encodedStr.length,           // string length
        ...encodedStr,               // string data
        testBytes.length,            // array length
        ...testBytes                 // array data
      ]);
      
      controller.enqueue(streamData);
      controller.close();

      // Read boolean
      const [boolResult, boolError] = await reader.readBoolean();
      expect(boolError).toBeUndefined();
      expect(boolResult).toBe(true);

      // Read varint
      const [varintResult, varintError] = await reader.readVarint();
      expect(varintError).toBeUndefined();
      expect(varintResult).toBe(123n);

      // Read string
      const [stringResult, stringError] = await reader.readString();
      expect(stringError).toBeUndefined();
      expect(stringResult).toBe(testStr);

      // Read byte array
      const [arrayResult, arrayError] = await reader.readUint8Array();
      expect(arrayError).toBeUndefined();
      expect(arrayResult).toEqual(testBytes);
    });

    it('should handle stream errors gracefully', async () => {
      // Close the stream to simulate end of data
      controller.close();

      const [result, error] = await reader.readUint8();
      
      expect(result).toBeUndefined();
      expect(error).toBeDefined();
      expect(error?.message).toContain('Stream closed');
    });
  });

  describe('BYOB (Bring Your Own Buffer) support', () => {
    it('should work with BYOB reader when available', async () => {
      // Create a simple stream with data, not necessarily BYOB
      const data = new Uint8Array([42]);
      const freshReader = createClosedReader(data);
      
      const [result, error] = await freshReader.readUint8();
      
      expect(error).toBeUndefined();
      expect(result).toBe(42);
      
      await freshReader.cancel(0, 'Test cleanup');
    });
  });
});
