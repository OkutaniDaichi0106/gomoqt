import { assertEquals } from "@std/assert";
import { SubscribeOkMessage } from "./subscribe_ok.ts";
import { ReceiveStream, SendStream } from "../webtransport/mod.ts";

Deno.test("SubscribeOkMessage - encode/decode roundtrip", async (t) => {
	await t.step("should encode and decode empty message", async () => {
		// Create buffer for encoding
		const chunks: Uint8Array[] = [];
		const writableStream = new WritableStream({
			write(chunk) {
				chunks.push(chunk);
			},
		});
		const writer = new SendStream({
			stream: writableStream,
			transfer: undefined,
			streamId: 0n,
		});

		const message = new SubscribeOkMessage({});
		const encodeErr = await message.encode(writer);
		assertEquals(encodeErr, undefined);

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
			transfer: undefined,
			streamId: 0n,
		});

		const decodedMessage = new SubscribeOkMessage({});
		const decodeErr = await decodedMessage.decode(reader);
		assertEquals(decodeErr, undefined);
	});
});
