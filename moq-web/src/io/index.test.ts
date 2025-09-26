import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import * as ioModule from './index';
import { Reader } from './reader';
import { Writer } from './writer';
import { StreamError } from './error';

describe('IO Module Index', () => {
  describe('exports', () => {
    it('should export Reader class', () => {
      expect(ioModule.Reader).toBeDefined();
      expect(ioModule.Reader).toBe(Reader);
      expect(typeof ioModule.Reader).toBe('function');
    });

    it('should export Writer class', () => {
      expect(ioModule.Writer).toBeDefined();
      expect(ioModule.Writer).toBe(Writer);
      expect(typeof ioModule.Writer).toBe('function');
    });

    it('should be able to instantiate Reader from index export', () => {
      const readableStream = new ReadableStream<Uint8Array>();
      const reader = new ioModule.Reader(readableStream);
      
      expect(reader).toBeInstanceOf(Reader);
      expect(reader).toBeInstanceOf(ioModule.Reader);
    });

    it('should be able to instantiate Writer from index export', () => {
      const writableStream = new WritableStream<Uint8Array>();
      const writer = new ioModule.Writer(writableStream);
      
      expect(writer).toBeInstanceOf(Writer);
      expect(writer).toBeInstanceOf(ioModule.Writer);
    });

    it('should export StreamError class', () => {
      expect(ioModule.StreamError).toBeDefined();
      expect(ioModule.StreamError).toBe(StreamError);
      expect(typeof ioModule.StreamError).toBe('function');
    });
  });

  describe('module structure', () => {
    it('should have the expected exported members', () => {
      const exportedKeys = Object.keys(ioModule);
      
      expect(exportedKeys).toContain('Reader');
      expect(exportedKeys).toContain('Writer');
    });

    it('should not have any unexpected exports', () => {
      const exportedKeys = Object.keys(ioModule);
      const expectedKeys = ['Reader', 'Writer', 'StreamError', 'EOF', 'MAX_VARINT1', 'MAX_VARINT2', 'MAX_VARINT4', 'MAX_VARINT8', 'varintLen', 'stringLen', 'bytesLen', 'BytesBuffer', 'BufferPool', 'MAX_BYTES_LENGTH', 'MAX_UINT', 'writeVarint', 'readVarint', 'writeBigVarint', 'readBigVarint', 'writeUint8Array', 'readUint8Array', 'writeString', 'readString', 'DefaultBytesPoolOptions', 'DefaultBufferPool'];
      
      // Check that all exported keys are expected
      exportedKeys.forEach(key => {
        expect(expectedKeys).toContain(key);
      });
      
      // Check that all expected keys are exported
      expectedKeys.forEach(key => {
        expect(exportedKeys).toContain(key);
      });
    });

    it('should export constructable classes', () => {
      expect(() => {
        new ioModule.Reader(new ReadableStream<Uint8Array>());
      }).not.toThrow();

      expect(() => {
        new ioModule.Writer(new WritableStream<Uint8Array>());
      }).not.toThrow();
    });
  });

  describe('re-export functionality', () => {
    it('should maintain all functionality of Reader through re-export', async () => {
      const readableStream = new ReadableStream<Uint8Array>({
        start(controller) {
          controller.enqueue(new Uint8Array([42]));
          controller.close();
        }
      });
      
      const reader = new ioModule.Reader(readableStream);
      const [result, error] = await reader.readUint8();
      
      expect(error).toBeUndefined();
      expect(result).toBe(42);
      
      await reader.cancel(new StreamError(0, 'Test cleanup'));
    });

    it('should maintain all functionality of Writer through re-export', async () => {
      const writtenData: Uint8Array[] = [];
      const writableStream = new WritableStream<Uint8Array>({
        write(chunk) {
          writtenData.push(chunk);
        }
      });
      
      const writer = new ioModule.Writer(writableStream);
      writer.writeBoolean(true);
      await writer.flush();
      
      expect(writtenData).toHaveLength(1);
      expect(writtenData[0]).toEqual(new Uint8Array([1]));
      
      await writer.close();
    });
  });

  describe('type compatibility', () => {
    it('should have compatible types between direct import and re-export', () => {
      // Create separate streams for each reader to avoid locking issues
      const readableStream1 = new ReadableStream<Uint8Array>();
      const readableStream2 = new ReadableStream<Uint8Array>();
      const writableStream1 = new WritableStream<Uint8Array>();
      const writableStream2 = new WritableStream<Uint8Array>();
      
      // These should be type-compatible
      const directReader: Reader = new ioModule.Reader(readableStream1);
      const reExportReader: ioModule.Reader = new Reader(readableStream2);
      
      const directWriter: Writer = new ioModule.Writer(writableStream1);
      const reExportWriter: ioModule.Writer = new Writer(writableStream2);
      
      expect(directReader).toBeInstanceOf(ioModule.Reader);
      expect(reExportReader).toBeInstanceOf(Reader);
      expect(directWriter).toBeInstanceOf(ioModule.Writer);
      expect(reExportWriter).toBeInstanceOf(Writer);
    });
  });
});
