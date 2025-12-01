import { assertEquals } from "@std/assert";
import { SessionUpdateMessage } from "./session_update.ts";
import { ReceiveStream, SendStream } from "../webtransport/mod.ts";

Deno.test("SessionUpdateMessage - encode/decode roundtrip - multiple scenarios", async (t) => {
  const testCases = {
    "normal bitrate": {
      bitrate: 1000,
    },
    "zero bitrate": {
      bitrate: 0,
    },
    "high bitrate": {
      bitrate: 10000000,
    },
    "low bitrate": {
      bitrate: 1,
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

      const message = new SessionUpdateMessage(input);
      const encodeErr = await message.encode(writer);
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

      const decodedMessage = new SessionUpdateMessage({});
      const decodeErr = await decodedMessage.decode(reader);
      assertEquals(decodeErr, undefined, `decode failed for ${caseName}`);
      assertEquals(
        decodedMessage.bitrate,
        input.bitrate,
        `bitrate mismatch for ${caseName}`,
      );
    });
  }

  await t.step(
    "decode should return error when readVarint fails for message length",
    async () => {
      const readableStream = new ReadableStream({
        start(controller) {
          controller.close();
        },
      });
      const reader = new ReceiveStream({
        stream: readableStream,
        streamId: 0n,
      });

      const message = new SessionUpdateMessage({});
      const err = await message.decode(reader);
      assertEquals(err !== undefined, true);
    },
  );

  await t.step(
    "decode should return error when reading bitrate fails",
    async () => {
      const buffer = new Uint8Array([2]); // only message length
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

      const message = new SessionUpdateMessage({});
      const err = await message.decode(reader);
      assertEquals(err !== undefined, true);
    },
  );
});
