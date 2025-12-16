import { assert, assertEquals } from "@std/assert";
import { SessionUpdateMessage } from "./session_update.ts";
import { Buffer } from "@okdaichi/golikejs/bytes";
import type { Writer } from "@okdaichi/golikejs/io";

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
			// Encode using Buffer
			const buffer = Buffer.make(100);
			const message = new SessionUpdateMessage(input);
			const encodeErr = await message.encode(buffer);
			assertEquals(encodeErr, undefined, `encode failed for ${caseName}`);

			// Decode from a new buffer with written data
			const readBuffer = Buffer.make(100);
			await readBuffer.write(buffer.bytes());
			const decodedMessage = new SessionUpdateMessage({});
			const decodeErr = await decodedMessage.decode(readBuffer);
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
			const buffer = Buffer.make(0); // Empty buffer
			const message = new SessionUpdateMessage({});
			const err = await message.decode(buffer);
			assertEquals(err !== undefined, true);
		},
	);

	await t.step(
		"decode should return error when readFull fails",
		async () => {
			const buffer = Buffer.make(10);
			// Write message length = 5 (varint), but no data follows
			await buffer.write(new Uint8Array([0x05]));

			const message = new SessionUpdateMessage({});
			const err = await message.decode(buffer);
			assert(err !== undefined);
		},
	);

	// Encode error tests using mockWriter with callCount tracking
	await t.step(
		"encode should return error when writeUint16 fails",
		async () => {
			let callCount = 0;
			const mockWriter: Writer = {
				async write(_p: Uint8Array): Promise<[number, Error | undefined]> {
					callCount++;
					if (callCount > 0) {
						return [0, new Error("Write failed")];
					}
					return [_p.length, undefined];
				},
			};

			const message = new SessionUpdateMessage({ bitrate: 1000 });
			const err = await message.encode(mockWriter);
			assertEquals(err instanceof Error, true);
		},
	);

	await t.step(
		"encode should return error when writing bitrate fails",
		async () => {
			let callCount = 0;
			const mockWriter: Writer = {
				async write(p: Uint8Array): Promise<[number, Error | undefined]> {
					callCount++;
					if (callCount > 1) {
						return [0, new Error("Write failed")];
					}
					return [p.length, undefined];
				},
			};

			const message = new SessionUpdateMessage({ bitrate: 1000 });
			const err = await message.encode(mockWriter);
			assertEquals(err instanceof Error, true);
		},
	);
});
