import { describe, it, expect } from '@jest/globals';
import { Writer, Reader } from '../io';

/**
 * Helper function to create isolated writer/reader pair that avoids TransformStream deadlock
 * Uses a custom queue-based implementation for reliable testing - optimized for performance
 */
export function createIsolatedStreams(): { writer: Writer; reader: Reader; cleanup: () => Promise<void> } {
  const chunks: Uint8Array[] = [];
  let writerClosed = false;
  
  // Use a more efficient WritableStream implementation
  const writableStream = new WritableStream<Uint8Array>({
    write(chunk) {
      // Avoid copying if possible for performance
      chunks.push(chunk instanceof Uint8Array ? chunk : new Uint8Array(chunk));
    },
    close() {
      writerClosed = true;
    }
  }, {
    // Optimize chunk size for performance
    highWaterMark: 16384,
    size(chunk) { return chunk.byteLength; }
  });
  
  const writer = new Writer(writableStream);
  
  let chunkIndex = 0;
  // Use a more efficient ReadableStream implementation
  const readableStream = new ReadableStream<Uint8Array>({
    pull(controller) {
      if (chunkIndex < chunks.length) {
        controller.enqueue(chunks[chunkIndex++]);
      } else if (writerClosed) {
        controller.close();
      }
      // If not closed and no chunks available, just return (will be called again)
    }
  }, {
    // Optimize chunk size for performance
    highWaterMark: 16384,
    size(chunk) { return chunk.byteLength; }
  });
  
  const reader = new Reader(readableStream);
  
  return {
    writer,
    reader,
    cleanup: async () => {
      try {
        if (!writerClosed) {
          await writer.close();
        }
      } catch {
        // ignore cleanup errors for performance
      }
    }
  };
}

describe('Test Utilities', () => {
  it('should create isolated streams with working writer and reader', async () => {
    const { writer, reader, cleanup } = createIsolatedStreams();
    
    try {
      // Test that writer and reader are created
      expect(writer).toBeDefined();
      expect(reader).toBeDefined();
      
      // Test basic write and read functionality
      const testData = new Uint8Array([1, 2, 3, 4, 5]);
      writer.writeUint8Array(testData);
      await writer.flush();
      await writer.close();
      
      const [readData, error] = await reader.readUint8Array();
      expect(error).toBeUndefined();
      expect(readData).toEqual(testData);
    } finally {
      await cleanup();
    }
  });

  it('should handle string writing and reading', async () => {
    const { writer, reader, cleanup } = createIsolatedStreams();
    
    try {
      const testString = "Hello, World!";
      writer.writeString(testString);
      await writer.flush();
      await writer.close();
      
      const [readString, error] = await reader.readString();
      expect(error).toBeUndefined();
      expect(readString).toBe(testString);
    } finally {
      await cleanup();
    }
  });

  it('should handle string array writing and reading', async () => {
    const { writer, reader, cleanup } = createIsolatedStreams();
    
    try {
      const testArray = ["string1", "string2", "string3"];
      writer.writeStringArray(testArray);
      await writer.flush();
      await writer.close();
      
      const [readArray, error] = await reader.readStringArray();
      expect(error).toBeUndefined();
      expect(readArray).toEqual(testArray);
    } finally {
      await cleanup();
    }
  });

  it('should handle boolean writing and reading', async () => {
    const { writer, reader, cleanup } = createIsolatedStreams();
    
    try {
      writer.writeBoolean(true);
      writer.writeBoolean(false);
      await writer.flush();
      await writer.close();
      
      const [value1, error1] = await reader.readBoolean();
      expect(error1).toBeUndefined();
      expect(value1).toBe(true);
      
      const [value2, error2] = await reader.readBoolean();
      expect(error2).toBeUndefined();
      expect(value2).toBe(false);
    } finally {
      await cleanup();
    }
  });

  it('should handle varint writing and reading', async () => {
    const { writer, reader, cleanup } = createIsolatedStreams();
    
    try {
      const testValues = [0n, 1n, 255n, 256n, 65535n, 65536n];
      
      for (const value of testValues) {
        writer.writeBigVarint(value);
      }
      await writer.flush();
      await writer.close();
      
      for (const expectedValue of testValues) {
        const [readValue, error] = await reader.readBigVarint();
        expect(error).toBeUndefined();
        expect(readValue).toBe(expectedValue);
      }
    } finally {
      await cleanup();
    }
  });

  it('should cleanup resources properly', async () => {
    const { writer, reader, cleanup } = createIsolatedStreams();
    
    // Test that cleanup doesn't throw even if called multiple times
    await cleanup();
    await cleanup(); // Should not throw
    
    expect(true).toBe(true); // Test passes if no exception thrown
  });
});
