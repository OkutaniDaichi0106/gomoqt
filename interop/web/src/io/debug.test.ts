import { describe, it, expect } from '@jest/globals';
import { Writer, Reader } from '../io';

describe('Simple Reader/Writer Debug Tests', () => {
  it('should write and read a single byte', async () => {
    console.log('Starting simple byte test...');
    
    // Create streams
    const transform = new TransformStream<Uint8Array, Uint8Array>();
    const writer = new Writer(transform.writable);
    const reader = new Reader(transform.readable);

    // Write a single byte
    console.log('Writing byte 42...');
    writer.writeUint8(42);
    
    console.log('Flushing...');
    const flushErr = await writer.flush();
    console.log('Flush result:', flushErr);
    
    console.log('Closing writer...');
    await writer.close();
    console.log('Writer closed');

    // Read the byte
    console.log('Reading byte...');
    try {
      const [value, err] = await reader.readUint8();
      console.log('Read result:', { value, err });
      
      expect(err).toBeUndefined();
      expect(value).toBe(42);
    } catch (error) {
      console.error('Read error:', error);
      throw error;
    }
  });

  it('should write and read using separate stream instances', async () => {
    console.log('Starting separate instances test...');
    
    // Use different approach - write to array and then read
    const chunks: Uint8Array[] = [];
    
    // Write phase
    const writableStream = new WritableStream<Uint8Array>({
      write(chunk) {
        console.log('Received chunk:', Array.from(chunk));
        chunks.push(chunk);
      }
    });
    
    const writer = new Writer(writableStream);
    writer.writeUint8(42);
    await writer.flush();
    await writer.close();
    
    console.log('All chunks:', chunks.map(c => Array.from(c)));
    
    // Read phase - create readable stream from collected chunks
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
    
    const reader = new Reader(readableStream);
    const [value, err] = await reader.readUint8();
    
    console.log('Read result:', { value, err });
    expect(err).toBeUndefined();
    expect(value).toBe(42);
  });
});
