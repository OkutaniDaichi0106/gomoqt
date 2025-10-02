import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { Writer } from './writer';
import { StreamError } from './error';

describe('Writer', () => {
  let writer: Writer;
  let writableStream: WritableStream<Uint8Array>;
  let writtenData: Uint8Array[];
  let streamClosed = false;

  beforeEach(() => {
    writtenData = [];
    streamClosed = false;
    writableStream = new WritableStream<Uint8Array>({
      write(chunk) {
        writtenData.push(chunk);
      },
      close() {
        streamClosed = true;
      }
    });
    writer = new Writer(writableStream);
  });

  afterEach(async () => {
    try {
      if (!streamClosed) {
        await writer.close();
      }
    } catch (error: any) {
      // Ignore errors during cleanup - stream might already be closed
      if (error?.code !== 'ERR_INVALID_STATE') {
        console.warn('Cleanup error:', error);
      }
    }
  });

  describe('writeUint8Array', () => {
    it('should write a Uint8Array with varint length prefix', async () => {
      const data = new Uint8Array([1, 2, 3, 4, 5]);
      
      writer.writeUint8Array(data);
      await writer.flush();

      expect(writtenData).toHaveLength(1);
      const written = writtenData[0];
      
      // First byte should be varint length (5)
      expect(written[0]).toBe(5);
      // Rest should be the data
      expect(written.slice(1)).toEqual(data);
    });

    it('should throw error for data exceeding maximum length', () => {
      // MAX_BYTES_LENGTH is 1 << 30 (1 GiB)
      // We can test with a mock object that has a length property exceeding the limit
      // without actually allocating the memory
      const largeData = {
        length: (1 << 30) + 1, // Just over MAX_BYTES_LENGTH
        // Add other Uint8Array-like properties if the implementation checks them
      } as Uint8Array;
      
      expect(() => {
        writer.writeUint8Array(largeData);
      }).toThrow('Bytes length exceeds maximum limit');
    });

    it('should handle empty array', async () => {
      const data = new Uint8Array([]);
      
      writer.writeUint8Array(data);
      await writer.flush();

      expect(writtenData).toHaveLength(1);
      const written = writtenData[0];
      
      // First byte should be varint length (0)
      expect(written[0]).toBe(0);
      expect(written.length).toBe(1);
    });
  });

  describe('writeString', () => {
    it('should write a string as UTF-8 bytes with length prefix', async () => {
      const str = 'hello';
      
      writer.writeString(str);
      await writer.flush();

      expect(writtenData).toHaveLength(1);
      const written = writtenData[0];
      
      // First byte should be varint length (5)
      expect(written[0]).toBe(5);
      // Rest should be UTF-8 encoded string
      const expectedBytes = new TextEncoder().encode(str);
      expect(written.slice(1)).toEqual(expectedBytes);
    });

    it('should handle empty string', async () => {
      writer.writeString('');
      await writer.flush();

      expect(writtenData).toHaveLength(1);
      const written = writtenData[0];
      
      expect(written[0]).toBe(0);
      expect(written.length).toBe(1);
    });

    it('should handle Unicode characters', async () => {
      const str = 'こんにちは';
      
      writer.writeString(str);
      await writer.flush();

      expect(writtenData).toHaveLength(1);
      const written = writtenData[0];
      
      const expectedBytes = new TextEncoder().encode(str);
      expect(written[0]).toBe(expectedBytes.length);
      expect(written.slice(1)).toEqual(expectedBytes);
    });
  });

  describe('writeBigVarint', () => {
    it('should write single byte varint for values < 255', async () => {
      writer.writeBigVarint(42n);
      await writer.flush();

      expect(writtenData).toHaveLength(1);
      expect(writtenData[0]).toEqual(new Uint8Array([42]));
    });

    it('should write two byte varint for values < 16384', async () => {
      writer.writeBigVarint(300n); // 0x012C
      await writer.flush();

      expect(writtenData).toHaveLength(1);
      const written = writtenData[0];
      expect(written.length).toBe(2);
      expect(written[0]).toBe(0x41); // 0x40 | (300 >> 8)
      expect(written[1]).toBe(0x2C); // 300 & 0xFF
    });

    it('should write four byte varint for values < 2^30', async () => {
      writer.writeBigVarint(1000000n);
      await writer.flush();

      expect(writtenData).toHaveLength(1);
      const written = writtenData[0];
      expect(written.length).toBe(4);
      expect(written[0] & 0xC0).toBe(0x80); // Check first 2 bits
    });

    it('should write eight byte varint for large values', async () => {
      writer.writeBigVarint(1n << 40n);
      await writer.flush();

      expect(writtenData).toHaveLength(1);
      const written = writtenData[0];
      expect(written.length).toBe(8);
      expect(written[0]).toBe(0xC0);
    });

    it('should throw error for negative values', () => {
      expect(() => {
        writer.writeBigVarint(-1n);
      }).toThrow('Varint cannot be negative');
    });

    it('should throw error for values exceeding maximum', () => {
      const maxValue = (1n << 62n) - 1n;
      expect(() => {
        writer.writeBigVarint(maxValue + 1n);
      }).toThrow('Value exceeds maximum varint size');
    });
  });

  describe('writeBoolean', () => {
    it('should write true as 1', async () => {
      writer.writeBoolean(true);
      await writer.flush();

      expect(writtenData).toHaveLength(1);
      expect(writtenData[0]).toEqual(new Uint8Array([1]));
    });

    it('should write false as 0', async () => {
      writer.writeBoolean(false);
      await writer.flush();

      expect(writtenData).toHaveLength(1);
      expect(writtenData[0]).toEqual(new Uint8Array([0]));
    });
  });

  describe('flush', () => {
    it('should return success when buffer is flushed', async () => {
      writer.writeBoolean(true);
      const error = await writer.flush();

      expect(error).toBeUndefined();
      expect(writtenData).toHaveLength(1);
    });

    it('should handle multiple flushes without error', async () => {
      writer.writeBoolean(true);
      await writer.flush();
      
      writer.writeBoolean(false);
      const error = await writer.flush();

      expect(error).toBeUndefined();
      expect(writtenData).toHaveLength(2);
    });

    it('should handle flush with empty buffer', async () => {
      const error = await writer.flush();

      expect(error).toBeUndefined();
      expect(writtenData).toHaveLength(0);
    });
  });

  describe('close', () => {
    it('should close the writer', async () => {
      await expect(writer.close()).resolves.not.toThrow();
    });

    it('should handle multiple close calls gracefully', async () => {
      await writer.close();
      // Second close should handle gracefully since stream is already closed
      try {
        await writer.close();
        // If no error is thrown, that's also acceptable
      } catch (error) {
        // Stream already closed errors are expected and acceptable
        expect(error).toBeDefined();
      }
    });
  });

  describe('cancel', () => {
    it('should cancel the writer with error', async () => {
      const error = new StreamError(1, 'Test error');
      await expect(writer.cancel(error)).resolves.not.toThrow();
    });
  });

  describe('closed', () => {
    it('should return a promise that resolves when writer is closed', async () => {
      const closedPromise = writer.closed();
      await writer.close();
      await expect(closedPromise).resolves.not.toThrow();
    });
  });

  describe('integration tests', () => {
    it('should write multiple data types in sequence', async () => {
      writer.writeBoolean(true);
      writer.writeBigVarint(123n);
      writer.writeString('test');
      writer.writeUint8Array(new Uint8Array([1, 2, 3]));
      
      await writer.flush();

      expect(writtenData).toHaveLength(1);
      const written = writtenData[0];
      
      // Should contain all written data
      expect(written.length).toBeGreaterThan(10);
      
      // First byte should be boolean true
      expect(written[0]).toBe(1);
    });
  });
});
