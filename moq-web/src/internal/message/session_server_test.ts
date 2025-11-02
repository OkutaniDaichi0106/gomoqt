import { assertEquals } from "@std/assert";
import { SessionServerMessage } from "./session_server.ts";
import { ReceiveStream, SendStream } from "../webtransport/mod.ts";

Deno.test("SessionServerMessage - encode/decode roundtrip - multiple scenarios", async (t) => {
	const testCases = {
		"with extensions": {
			version: 1,
			extensions: new Map<number, Uint8Array>([
				[1, new Uint8Array([1, 2, 3])],
				[2, new Uint8Array([4, 5, 6])],
			]),
		},
		"without extensions": {
			version: 1,
			extensions: new Map<number, Uint8Array>(),
		},
		"different version": {
			version: 2,
			extensions: new Map<number, Uint8Array>([[1, new Uint8Array([7, 8, 9])]]),
		},
		"single extension": {
			version: 1,
			extensions: new Map<number, Uint8Array>([[100, new Uint8Array([0xff, 0xfe, 0xfd])]]),
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

			const message = new SessionServerMessage(input);
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

			const decodedMessage = new SessionServerMessage({});
			const decodeErr = await decodedMessage.decode(reader);
			assertEquals(decodeErr, undefined, `decode failed for ${caseName}`);
			assertEquals(decodedMessage.version, input.version, `version mismatch for ${caseName}`);
			assertEquals(
				decodedMessage.extensions,
				input.extensions,
				`extensions mismatch for ${caseName}`,
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
				transfer: undefined,
				streamId: 0n,
			});

			const message = new SessionServerMessage({});
			const err = await message.decode(reader);
			assertEquals(err !== undefined, true);
		},
	);

	await t.step("decode should return error when readVarint fails for version", async () => {
		const buffer = new Uint8Array([1]); // only message length
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

		const message = new SessionServerMessage({});
		const err = await message.decode(reader);
		assertEquals(err !== undefined, true);
	});

	await t.step(
		"decode should return error when readVarint fails for extension count",
		async () => {
			const buffer = new Uint8Array([2, 1]); // message length, version, but no extension count
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

			const message = new SessionServerMessage({});
			const err = await message.decode(reader);
			assertEquals(err !== undefined, true);
		},
	);

	await t.step("decode should return error when reading extension ID fails", async () => {
		const buffer = new Uint8Array([3, 1, 1]); // message length, version, extensionCount=1, but no ID
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

		const message = new SessionServerMessage({});
		const err = await message.decode(reader);
		assertEquals(err !== undefined, true);
	});

	await t.step("decode should return error when reading extension data fails", async () => {
		const buffer = new Uint8Array([4, 1, 1, 1]); // message length, version, extensionCount=1, ID=1, but no data
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

		const message = new SessionServerMessage({});
		const err = await message.decode(reader);
		assertEquals(err !== undefined, true);
	});
});
