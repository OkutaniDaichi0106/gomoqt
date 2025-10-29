import { assertEquals, assertExists } from "../../deps.ts";
import { Reader } from '../webtransport/reader.ts';
import { Writer } from '../webtransport/writer.ts';

// Reuse the same helper used in existing tests but in a Deno.test friendly file
export function createIsolatedStreams(): { writer: Writer; reader: Reader; cleanup: () => Promise<void> } {
  const chunks: Uint8Array[] = [];
  let writerClosed = false;

  const writableStream = new WritableStream<Uint8Array>({
    write(chunk) {
      const copy = chunk instanceof Uint8Array ? chunk.slice() : new Uint8Array(chunk);
      chunks.push(copy);
    },
    close() {
      writerClosed = true;
    }
  }, {
    highWaterMark: 16384,
    size(chunk) { return chunk.byteLength; }
  });

  const writer = new Writer({stream: writableStream, transfer: undefined, streamId: 0n});

  let chunkIndex = 0;
  const readableStream = new ReadableStream<Uint8Array>({
    pull(controller) {
      if (chunkIndex < chunks.length) {
        controller.enqueue(chunks[chunkIndex++]);
      } else if (writerClosed) {
        controller.close();
      }
    }
  }, {
    highWaterMark: 16384,
    size(chunk) { return chunk.byteLength; }
  });

  const reader = new Reader({stream: readableStream, transfer: undefined, streamId: 0n});

  return {
    writer,
    reader,
    cleanup: async () => {
      try {
        if (!writerClosed) await writer.close();
      } catch { /* ignore */ }
    }
  };
}

Deno.test("message/test-utils - basic read/write operations", async (t) => {
  await t.step("write/read Uint8Array", async () => {
    const { writer, reader, cleanup } = createIsolatedStreams();
    try {
      assertExists(writer);
      assertExists(reader);
      const data = new Uint8Array([1,2,3,4,5]);
      writer.writeUint8Array(data);
      await writer.flush();
      await writer.close();

      const [readData, err] = await reader.readUint8Array();
      assertEquals(err, undefined);
      assertEquals(readData, data);
    } finally {
      await cleanup();
    }
  });

  await t.step("write/read string", async () => {
    const { writer, reader, cleanup } = createIsolatedStreams();
    try {
      const s = "Hello, World!";
      writer.writeString(s);
      await writer.flush();
      await writer.close();

      const [rs, err] = await reader.readString();
      assertEquals(err, undefined);
      assertEquals(rs, s);
    } finally { await cleanup(); }
  });

  await t.step("write/read boolean", async () => {
    const { writer, reader, cleanup } = createIsolatedStreams();
    try {
      writer.writeBoolean(true);
      writer.writeBoolean(false);
      await writer.flush();
      await writer.close();

      const [v1, e1] = await reader.readBoolean();
      assertEquals(e1, undefined);
      assertEquals(v1, true);

      const [v2, e2] = await reader.readBoolean();
      assertEquals(e2, undefined);
      assertEquals(v2, false);
    } finally { await cleanup(); }
  });
});
