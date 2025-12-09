import { assertEquals } from "@std/assert";
import { AnnouncePleaseMessage } from "./announce_please.ts";
import { Buffer } from "@okdaichi/golikejs/bytes";
import type { Writer } from "@okdaichi/golikejs/io";

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
			// Encode using Buffer
			const buffer = Buffer.make(100);
			const message = new AnnouncePleaseMessage(input);
			const encodeErr = await message.encode(buffer);
			assertEquals(encodeErr, undefined, `encode failed for ${caseName}`);

			// Decode from a new buffer with written data
			const readBuffer = Buffer.make(100);
			await readBuffer.write(buffer.bytes());
			const decodedMessage = new AnnouncePleaseMessage({});
			const decodeErr = await decodedMessage.decode(readBuffer);
			assertEquals(decodeErr, undefined, `decode failed for ${caseName}`);
			assertEquals(
				decodedMessage.prefix,
				input.prefix,
				`prefix mismatch for ${caseName}`,
			);
		});
	}

	await t.step("decode should return error when readUint16 fails", async () => {
		const buffer = Buffer.make(0); // Empty buffer
		const message = new AnnouncePleaseMessage({});
		const err = await message.decode(buffer);
		assertEquals(err !== undefined, true);
	});

	await t.step("decode should return error when readFull fails", async () => {
		const buffer = Buffer.make(10);
		// Write message length = 10 (uint16 big-endian), but no data follows
		await buffer.write(new Uint8Array([0x00, 0x0a]));
		const message = new AnnouncePleaseMessage({});
		const err = await message.decode(buffer);
		assertEquals(err !== undefined, true);
	});

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

			const message = new AnnouncePleaseMessage({ prefix: "test" });
			const err = await message.encode(mockWriter);
			assertEquals(err instanceof Error, true);
		},
	);

	await t.step(
		"encode should return error when writeString fails",
		async () => {
			let callCount = 0;
			const mockWriter: Writer = {
				async write(p: Uint8Array): Promise<[number, Error | undefined]> {
					callCount++;
					if (callCount > 1) {
						return [0, new Error("Write failed on string")];
					}
					return [p.length, undefined];
				},
			};

			const message = new AnnouncePleaseMessage({ prefix: "test" });
			const err = await message.encode(mockWriter);
			assertEquals(err instanceof Error, true);
		},
	);
});
