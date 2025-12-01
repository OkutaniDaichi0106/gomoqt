import { assertEquals } from "@std/assert";
import { GroupMessage } from "./group.ts";
import { ReceiveStream, SendStream } from "../webtransport/mod.ts";

Deno.test("GroupMessage - encode/decode roundtrip - multiple scenarios", async (t) => {
  const testCases = {
    "normal case": {
      subscribeId: 123,
      sequence: 456,
    },
    "zero values": {
      subscribeId: 0,
      sequence: 0,
    },
    "large numbers": {
      subscribeId: 1000000,
      sequence: 2000000,
    },
    "single values": {
      subscribeId: 1,
      sequence: 1,
    },
  };

  for (const [caseName, input] of Object.entries(testCases)) {
    await t.step(caseName, async () => {
      // Create buffer for encoding
      const chunks: Uint8Array[] = [];
      const writableStream = new WritableStream({
        write(chunk) {
          chunks.push(chunk);
        },
      });
      const writer = new SendStream({
        stream: writableStream,
        streamId: 0n,
      });

      const msg = new GroupMessage(input);
      const encodeErr = await msg.encode(writer);
      assertEquals(encodeErr, undefined, `encode failed for ${caseName}`);

      // Combine chunks into single buffer
      const totalLength = chunks.reduce((sum, chunk) => sum + chunk.length, 0);
      const combinedBuffer = new Uint8Array(totalLength);
      let offset = 0;
      for (const chunk of chunks) {
        combinedBuffer.set(chunk, offset);
        offset += chunk.length;
      }

      // Create readable stream for decoding
      const readableStream = new ReadableStream({
        start(controller) {
          controller.enqueue(combinedBuffer);
          controller.close();
        },
      });
      const reader = new ReceiveStream({
        stream: readableStream,
        streamId: 0n,
      });

      const decodedMsg = new GroupMessage({});
      const decodeErr = await decodedMsg.decode(reader);
      assertEquals(decodeErr, undefined, `decode failed for ${caseName}`);
      assertEquals(
        decodedMsg.subscribeId,
        input.subscribeId,
        `subscribeId mismatch for ${caseName}`,
      );
      assertEquals(
        decodedMsg.sequence,
        input.sequence,
        `sequence mismatch for ${caseName}`,
      );
    });
  }
});

Deno.test("GroupMessage - error cases", async (t) => {
  await t.step(
    "decode should return error when readVarint fails for message length",
    async () => {
      const readableStream = new ReadableStream({
        start(controller) {
          controller.close(); // Close immediately to cause read error
        },
      });
      const reader = new ReceiveStream({
        stream: readableStream,
        streamId: 0n,
      });

      const msg = new GroupMessage({});
      const err = await msg.decode(reader);
      if (!(err !== undefined)) throw new Error("expected error from decode");
    },
  );

  await t.step(
    "decode should return error when reading subscribeId fails",
    async () => {
      // message length only, no subscribeId
      const buffer = new Uint8Array([1]);
      const readableStream = new ReadableStream({
        start(controller) {
          controller.enqueue(buffer);
          controller.close();
        },
      });
      const reader = new ReceiveStream({
        stream: readableStream,
        streamId: 0n,
      });

      const msg = new GroupMessage({});
      const err = await msg.decode(reader);
      if (!(err !== undefined)) {
        throw new Error(
          "expected error from decode subscribeId",
        );
      }
    },
  );

  await t.step(
    "decode should return error when reading sequence fails",
    async () => {
      // Provide length and subscribeId but no sequence
      const buffer = new Uint8Array([2, 1]);
      const readableStream = new ReadableStream({
        start(controller) {
          controller.enqueue(buffer);
          controller.close();
        },
      });
      const reader = new ReceiveStream({
        stream: readableStream,
        streamId: 0n,
      });

      const msg = new GroupMessage({});
      const err = await msg.decode(reader);
      if (!(err !== undefined)) {
        throw new Error(
          "expected error from decode sequence",
        );
      }
    },
  );
});
