import { assertEquals } from "@std/assert";
import { AnnouncePleaseMessage } from "./announce_please.ts";
import { ReceiveStream, SendStream } from "../webtransport/mod.ts";

Deno.test("AnnouncePleaseMessage - encode/decode roundtrip - multiple scenarios", async (t) => {
	const testCases = {
		"normal case": {
			prefix: "test",
		},
		"empty prefix": {
			prefix: "",
		},
		"long prefix": {
			prefix: "very/long/path/to/namespace/prefix",
		},
		"prefix with special characters": {
			prefix: "test-prefix_123",
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

			const message = new AnnouncePleaseMessage(input);
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

			const decodedMessage = new AnnouncePleaseMessage({});
			const decodeErr = await decodedMessage.decode(reader);
			assertEquals(decodeErr, undefined, `decode failed for ${caseName}`);
			assertEquals(
				decodedMessage.prefix,
				input.prefix,
				`prefix mismatch for ${caseName}`,
			);
		});
	}

	await t.step("decode should return error when readVarint fails", async () => {
		const readableStream = new ReadableStream({
			start(controller) {
				controller.close(); // Close immediately to cause read error
			},
		});
		const reader = new ReceiveStream({
			stream: readableStream,
			streamId: 0n,
		});

		const message = new AnnouncePleaseMessage({});
		const err = await message.decode(reader);
		assertEquals(err !== undefined, true);
	});

	await t.step("decode should return error when readString fails", async () => {
		const buffer = new Uint8Array([5]); // message length = 5, but no string data
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

		const message = new AnnouncePleaseMessage({});
		const err = await message.decode(reader);
		assertEquals(err !== undefined, true);
	});
});
