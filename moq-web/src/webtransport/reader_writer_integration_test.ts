import { assertEquals } from "../../deps.ts";
import { Writer } from './writer.ts';
import { Reader } from './reader.ts';

// Helper function to create isolated writer/reader pair
function createIsolatedStreams(): { writer: Writer; reader: Reader; cleanup: () => Promise<void> } {
  // Use a TransformStream to connect writer -> reader synchronously
  const ts = new TransformStream<Uint8Array, Uint8Array>();
  const writer = new Writer({stream: ts.writable, transfer: undefined, streamId: 0n});
  const reader = new Reader({stream: ts.readable, transfer: undefined, streamId: 0n});
  
  return {
    writer,
    reader,
    cleanup: async () => {
      try {
        await writer.close();
      } catch (error: any) {
        // Ignore errors during cleanup - stream might already be closed
        if ((error as any)?.code !== 'ERR_INVALID_STATE') {
          throw error;
        }
      }
      // Ensure writer/reader internal promises have settled
      try { await writer.closed(); } catch (_e) { /* ignore */ }
      try { await reader.closed(); } catch (_e) { /* ignore */ }
    }
  };
}

Deno.test('webtransport/reader-writer integration - varint round-trip tests', async (t) => {
  await t.step('single byte varint (< 64)', async () => {
    const testValue = 42n;
    const { writer, reader, cleanup } = createIsolatedStreams();
    try {
      writer.writeBigVarint(testValue);
      await writer.flush();
      await writer.close();
      const [readValue, err] = await reader.readBigVarint();
      assertEquals(err, undefined);
      assertEquals(readValue, testValue);
    } finally {
      await cleanup();
    }
  });

  await t.step('two byte varint (< 16384)', async () => {
    const testValue = 300n;
    const { writer, reader, cleanup } = createIsolatedStreams();
    try {
      writer.writeBigVarint(testValue);
      await writer.flush();
      await writer.close();
      const [readValue, err] = await reader.readBigVarint();
      assertEquals(err, undefined);
      assertEquals(readValue, testValue);
    } finally {
      await cleanup();
    }
  });

  await t.step('string array round-trip', async () => {
    const testValue = ['hello', 'world', 'test'];
    const { writer, reader, cleanup } = createIsolatedStreams();
    try {
      writer.writeStringArray(testValue);
      await writer.flush();
      await writer.close();
      const [readValue, err] = await reader.readStringArray();
      assertEquals(err, undefined);
      assertEquals(readValue, testValue);
    } finally {
      await cleanup();
    }
  });

  await t.step('empty string array round-trip', async () => {
    const testValue: string[] = [];
    const { writer, reader, cleanup } = createIsolatedStreams();
    try {
      writer.writeStringArray(testValue);
      await writer.flush();
      await writer.close();
      const [readValue, err] = await reader.readStringArray();
      assertEquals(err, undefined);
      assertEquals(readValue, testValue);
    } finally {
      await cleanup();
    }
  });

  await t.step('string round-trip', async () => {
    const testValue = 'hello world';
    const { writer, reader, cleanup } = createIsolatedStreams();
    try {
      writer.writeString(testValue);
      await writer.flush();
      await writer.close();
      const [readValue, err] = await reader.readString();
      assertEquals(err, undefined);
      assertEquals(readValue, testValue);
    } finally {
      await cleanup();
    }
  });

  await t.step('uint8 round-trip', async () => {
    const testValue = 123;
    const { writer, reader, cleanup } = createIsolatedStreams();
    try {
      writer.writeUint8(testValue);
      await writer.flush();
      await writer.close();
      const [readValue, err] = await reader.readUint8();
      assertEquals(err, undefined);
      assertEquals(readValue, testValue);
    } finally {
      await cleanup();
    }
  });

  await t.step('boolean round-trip', async () => {
    const testValue = true;
    const { writer, reader, cleanup } = createIsolatedStreams();
    try {
      writer.writeBoolean(testValue);
      await writer.flush();
      await writer.close();
      const [readValue, err] = await reader.readBoolean();
      assertEquals(err, undefined);
      assertEquals(readValue, testValue);
    } finally {
      await cleanup();
    }
  });

  await t.step('uint8 array round-trip', async () => {
    const testValue = new Uint8Array([1,2,3,4,5]);
    const { writer, reader, cleanup } = createIsolatedStreams();
    try {
      writer.writeUint8Array(testValue);
      await writer.flush();
      await writer.close();
      const [readValue, err] = await reader.readUint8Array();
      assertEquals(err, undefined);
      assertEquals(readValue, testValue);
    } finally {
      await cleanup();
    }
  });

  await t.step('multiple data types in sequence', async () => {
    const { writer, reader, cleanup } = createIsolatedStreams();
    try {
      writer.writeBoolean(true);
      writer.writeBigVarint(123n);
      writer.writeString('test');
      writer.writeUint8Array(new Uint8Array([1,2,3]));
      writer.writeStringArray(['a','b','c']);
      await writer.flush();
      await writer.close();

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
      assertEquals(bytes1, new Uint8Array([1,2,3]));

      const [strArray1, err5] = await reader.readStringArray();
      assertEquals(err5, undefined);
      assertEquals(strArray1, ['a','b','c']);
    } finally {
      await cleanup();
    }
  });
});
