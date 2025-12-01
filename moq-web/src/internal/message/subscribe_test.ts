import { assertEquals } from "@std/assert";
import { SubscribeMessage } from "./subscribe.ts";
import { ReceiveStream, SendStream } from "../webtransport/mod.ts";

Deno.test("SubscribeMessage - encode/decode roundtrip - multiple scenarios", async (t) => {
  const testCases = {
    "normal case": {
      subscribeId: 123,
      broadcastPath: "path",
      trackName: "track",
      trackPriority: 1,
      minGroupSequence: 2,
      maxGroupSequence: 3,
    },
    "large sequence numbers": {
      subscribeId: 1000000,
      broadcastPath: "long/path/to/resource",
      trackName: "long-track-name-with-hyphens",
      trackPriority: 255,
      minGroupSequence: 1000000,
      maxGroupSequence: 2000000,
    },
    "zero values": {
      subscribeId: 0,
      broadcastPath: "",
      trackName: "",
      trackPriority: 0,
      minGroupSequence: 0,
      maxGroupSequence: 0,
    },
    "single character paths": {
      subscribeId: 1,
      broadcastPath: "a",
      trackName: "b",
      trackPriority: 1,
      minGroupSequence: 1,
      maxGroupSequence: 2,
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

      const message = new SubscribeMessage(input);
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

      const decodedMessage = new SubscribeMessage({});
      const decodeErr = await decodedMessage.decode(reader);
      assertEquals(decodeErr, undefined, `decode failed for ${caseName}`);

      // Verify all fields match
      assertEquals(
        decodedMessage.subscribeId,
        input.subscribeId,
        `subscribeId mismatch for ${caseName}`,
      );
      assertEquals(
        decodedMessage.broadcastPath,
        input.broadcastPath,
        `broadcastPath mismatch for ${caseName}`,
      );
      assertEquals(
        decodedMessage.trackName,
        input.trackName,
        `trackName mismatch for ${caseName}`,
      );
      assertEquals(
        decodedMessage.trackPriority,
        input.trackPriority,
        `trackPriority mismatch for ${caseName}`,
      );
      assertEquals(
        decodedMessage.minGroupSequence,
        input.minGroupSequence,
        `minGroupSequence mismatch for ${caseName}`,
      );
      assertEquals(
        decodedMessage.maxGroupSequence,
        input.maxGroupSequence,
        `maxGroupSequence mismatch for ${caseName}`,
      );
    });
  }

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

      const message = new SubscribeMessage({});
      const err = await message.decode(reader);
      assertEquals(err !== undefined, true);
    },
  );

  await t.step(
    "decode should return error when reading subscribeId fails",
    async () => {
      const buffer = new Uint8Array([5]); // only message length
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

      const message = new SubscribeMessage({});
      const err = await message.decode(reader);
      assertEquals(err !== undefined, true);
    },
  );

  await t.step(
    "decode should return error when reading broadcastPath fails",
    async () => {
      const buffer = new Uint8Array([5, 1]); // message length, subscribeId, but no broadcastPath
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

      const message = new SubscribeMessage({});
      const err = await message.decode(reader);
      assertEquals(err !== undefined, true);
    },
  );

  await t.step(
    "decode should return error when reading trackName fails",
    async () => {
      const buffer = new Uint8Array([6, 1, 0]); // message length, subscribeId, empty broadcastPath, but no trackName
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

      const message = new SubscribeMessage({});
      const err = await message.decode(reader);
      assertEquals(err !== undefined, true);
    },
  );
});
