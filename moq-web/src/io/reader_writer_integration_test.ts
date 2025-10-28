import { describe, it, assertEquals, assertExists } from "../../deps.ts";
import { Writer, Reader } from '../io';

// Helper function to create isolated writer/reader pair
function createIsolatedStreams(): { writer: Writer; reader: Reader; cleanup: () => Promise<void> } {
  const chunks: Uint8Array[] = [];
  
  const writableStream = new WritableStream<Uint8Array>({
    write(chunk) {
      chunks.push(chunk);
    }
  });
  
  const writer = new Writer({stream: writableStream, transfer: undefined, streamId: 0n});
  
  let chunkIndex = 0;
  const readableStream = new ReadableStream<Uint8Array>({
    pull(controller) {
      if (chunkIndex < chunks.length) {
        controller.enqueue(chunks[chunkIndex++]);
      } else {
        controller.close();
      }
    }
  });
  
  const reader = new Reader({stream: readableStream, transfer: undefined, streamId: 0n});
  
  return {
    writer,
    reader,
    cleanup: async () => {
      try {
        await writer.close();
      } catch (error) {
        // Ignore errors during cleanup - stream might already be closed
        if (error.code !== 'ERR_INVALID_STATE') {
          throw error;
        }
      }
    }
  };
}

describe('Reader/Writer Integration Tests', () => {
  describe('Varint round-trip tests', () => {
    it('should write and read single byte varint (< 64)', async () => {
      const testValue = 42n;
      const { writer, reader, cleanup } = createIsolatedStreams();
      
      try {
        // Write
        writer.writeBigVarint(testValue);
        await writer.flush();
        await writer.close();
        
        // Read
        const [readValue, err] = await reader.readBigVarint();
        
        assertEquals(err, undefined);
        assertEquals(readValue, testValue);
      } finally {
        await cleanup();
      }
    });

    it('should write and read two byte varint (< 16384)', async () => {
      const testValue = 300n;
      const { writer, reader, cleanup } = createIsolatedStreams();
      
      try {
        // Write
        writer.writeBigVarint(testValue);
        await writer.flush();
        await writer.close();
        
        // Read
        const [readValue, err] = await reader.readBigVarint();
        
        assertEquals(err, undefined);
        assertEquals(readValue, testValue);
      } finally {
        await cleanup();
      }
    });

    it('should write and read string array', async () => {
      const testValue = ['hello', 'world', 'test'];
      const { writer, reader, cleanup } = createIsolatedStreams();
      
      try {
        // Write
        writer.writeStringArray(testValue);
        await writer.flush();
        await writer.close();
        
        // Read
        const [readValue, err] = await reader.readStringArray();
        
        assertEquals(err, undefined);
        assertEquals(readValue, testValue);
      } finally {
        await cleanup();
      }
    });

    it('should write and read empty string array', async () => {
      const testValue: string[] = [];
      const { writer, reader, cleanup } = createIsolatedStreams();
      
      try {
        // Write
        writer.writeStringArray(testValue);
        await writer.flush();
        await writer.close();
        
        // Read
        const [readValue, err] = await reader.readStringArray();
        
        assertEquals(err, undefined);
        assertEquals(readValue, testValue);
      } finally {
        await cleanup();
      }
    });

    it('should write and read string', async () => {
      const testValue = 'hello world';
      const { writer, reader, cleanup } = createIsolatedStreams();
      
      try {
        // Write
        writer.writeString(testValue);
        await writer.flush();
        await writer.close();
        
        // Read
        const [readValue, err] = await reader.readString();
        
        assertEquals(err, undefined);
        assertEquals(readValue, testValue);
      } finally {
        await cleanup();
      }
    });

    it('should write and read uint8', async () => {
      const testValue = 123;
      const { writer, reader, cleanup } = createIsolatedStreams();
      
      try {
        // Write
        writer.writeUint8(testValue);
        await writer.flush();
        await writer.close();
        
        // Read
        const [readValue, err] = await reader.readUint8();
        
        assertEquals(err, undefined);
        assertEquals(readValue, testValue);
      } finally {
        await cleanup();
      }
    });

    it('should write and read boolean', async () => {
      const testValue = true;
      const { writer, reader, cleanup } = createIsolatedStreams();
      
      try {
        // Write
        writer.writeBoolean(testValue);
        await writer.flush();
        await writer.close();
        
        // Read
        const [readValue, err] = await reader.readBoolean();
        
        assertEquals(err, undefined);
        assertEquals(readValue, testValue);
      } finally {
        await cleanup();
      }
    });

    it('should write and read uint8 array', async () => {
      const testValue = new Uint8Array([1, 2, 3, 4, 5]);
      const { writer, reader, cleanup } = createIsolatedStreams();
      
      try {
        // Write
        writer.writeUint8Array(testValue);
        await writer.flush();
        await writer.close();
        
        // Read
        const [readValue, err] = await reader.readUint8Array();
        
        assertEquals(err, undefined);
        assertEquals(readValue, testValue);
      } finally {
        await cleanup();
      }
    });

    it('should write and read multiple data types in sequence', async () => {
      const { writer, reader, cleanup } = createIsolatedStreams();
      
      try {
        // Write multiple values
        writer.writeBoolean(true);
        writer.writeBigVarint(123n);
        writer.writeString('test');
        writer.writeUint8Array(new Uint8Array([1, 2, 3]));
        writer.writeStringArray(['a', 'b', 'c']);
        
        await writer.flush();
        await writer.close();

        // Read values in the same order
        const [bool1, err1] = await reader.readBoolean();
        assertEquals(err1, undefined);
        assertEquals(bool1, true);

        const [varint1, err2] = await reader.readBigVarint();
        assertEquals(err2, undefined);
        assertEquals(varint1, 123n);

        const [string1, err3] = await reader.readString();
        assertEquals(err3, undefined);
        assertEquals(string1, 'test');

        const [bytes1, err4] = await reader.readUint8Array();
        assertEquals(err4, undefined);
        assertEquals(bytes1, new Uint8Array([1, 2, 3]));

        const [strArray1, err5] = await reader.readStringArray();
        assertEquals(err5, undefined);
        assertEquals(strArray1, ['a', 'b', 'c']);
      } finally {
        await cleanup();
      }
    });
  });
});
