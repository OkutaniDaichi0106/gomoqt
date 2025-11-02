import { assertEquals } from "@std/assert";
import { SubscribeUpdateMessage } from "./subscribe_update.ts";
import { ReceiveStream, SendStream } from "../webtransport/mod.ts";

Deno.test("SubscribeUpdateMessage - encode/decode roundtrip - multiple scenarios", async (t) => {
	const testCases = {
		"normal case": {
			trackPriority: 1,
			minGroupSequence: 2n,
			maxGroupSequence: 3n,
		},
		"zero values": {
			trackPriority: 0,
			minGroupSequence: 0n,
			maxGroupSequence: 0n,
		},
		"large sequence numbers": {
			trackPriority: 255,
			minGroupSequence: 1000000n,
			maxGroupSequence: 2000000n,
		},
		"same min and max sequence": {
			trackPriority: 10,
			minGroupSequence: 100n,
			maxGroupSequence: 100n,
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
				transfer: undefined,
				streamId: 0n,
			});

			const message = new SubscribeUpdateMessage(input);
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
				transfer: undefined,
				streamId: 0n,
			});

			const decodedMessage = new SubscribeUpdateMessage({});
			const decodeErr = await decodedMessage.decode(reader);
			assertEquals(decodeErr, undefined, `decode failed for ${caseName}`);
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
					controller.close();
				},
			});
			const reader = new ReceiveStream({
				stream: readableStream,
				transfer: undefined,
				streamId: 0n,
			});

			const message = new SubscribeUpdateMessage({});
			const err = await message.decode(reader);
			assertEquals(err !== undefined, true);
		},
	);

	await t.step("decode should return error when reading subscribeId fails", async () => {
		const buffer = new Uint8Array([5]); // only message length
		const readableStream = new ReadableStream({
			start(controller) {
				controller.enqueue(buffer);
				controller.close();
			},
		});
		const reader = new ReceiveStream({
			stream: readableStream,
			transfer: undefined,
			streamId: 0n,
		});

		const message = new SubscribeUpdateMessage({});
		const err = await message.decode(reader);
		assertEquals(err !== undefined, true);
	});

	await t.step("decode should return error when reading minGroupSequence fails", async () => {
		const buffer = new Uint8Array([6, 1]); // message length, subscribeId, but no minGroupSequence
		const readableStream = new ReadableStream({
			start(controller) {
				controller.enqueue(buffer);
				controller.close();
			},
		});
		const reader = new ReceiveStream({
			stream: readableStream,
			transfer: undefined,
			streamId: 0n,
		});

		const message = new SubscribeUpdateMessage({});
		const err = await message.decode(reader);
		assertEquals(err !== undefined, true);
	});
});
