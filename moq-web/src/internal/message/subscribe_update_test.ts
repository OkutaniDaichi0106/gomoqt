import { assertEquals } from "@std/assert";
import { SubscribeUpdateMessage } from "./subscribe_update.ts";
import { Buffer } from "@okudai/golikejs/bytes";
import type { Writer } from "@okudai/golikejs/io";

Deno.test("SubscribeUpdateMessage - encode/decode roundtrip - multiple scenarios", async (t) => {
	const testCases = {
		"normal case": {
			trackPriority: 1,
			minGroupSequence: 2,
			maxGroupSequence: 3,
		},
		"zero values": {
			trackPriority: 0,
			minGroupSequence: 0,
			maxGroupSequence: 0,
		},
		"large sequence numbers": {
			trackPriority: 255,
			minGroupSequence: 1000000,
			maxGroupSequence: 2000000,
		},
		"same min and max sequence": {
			trackPriority: 10,
			minGroupSequence: 100,
			maxGroupSequence: 100,
		},
	};

	for (const [caseName, input] of Object.entries(testCases)) {
		await t.step(caseName, async () => {
			// Encode using Buffer
			const buffer = Buffer.make(100);
			const message = new SubscribeUpdateMessage(input);
			const encodeErr = await message.encode(buffer);
			assertEquals(encodeErr, undefined, `encode failed for ${caseName}`);

			// Decode from a new buffer with written data
			const readBuffer = Buffer.make(100);
			await readBuffer.write(buffer.bytes());
			const decodedMessage = new SubscribeUpdateMessage({});
			const decodeErr = await decodedMessage.decode(readBuffer);
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
			const buffer = Buffer.make(0); // Empty buffer
			const message = new SubscribeUpdateMessage({});
			const err = await message.decode(buffer);
			assertEquals(err !== undefined, true);
		},
	);

	await t.step(
		"decode should return error when reading subscribeId fails",
		async () => {
			const buffer = Buffer.make(10);
			await buffer.write(new Uint8Array([0x00, 0x05])); // message length = 5, but no data
			const message = new SubscribeUpdateMessage({});
			const err = await message.decode(buffer);
			assertEquals(err !== undefined, true);
		},
	);

	await t.step(
		"decode should return error when reading minGroupSequence fails",
		async () => {
			const buffer = Buffer.make(10);
			// Write message length = 6, subscribeId = 1, but no minGroupSequence
			await buffer.write(new Uint8Array([0x00, 0x06, 0x01]));
			const message = new SubscribeUpdateMessage({});
			const err = await message.decode(buffer);
			assertEquals(err !== undefined, true);
		},
	);

	await t.step("encode should return error when writeUint16 fails", async () => {
		const mockWriter: Writer = {
			async write(_p: Uint8Array): Promise<[number, Error | undefined]> {
				return [0, new Error("Write failed")];
			},
		};

		const message = new SubscribeUpdateMessage({
			trackPriority: 1,
			minGroupSequence: 0,
			maxGroupSequence: 100,
		});
		const err = await message.encode(mockWriter);
		assertEquals(err instanceof Error, true);
	});

	await t.step(
		"encode should return error when writeVarint fails for subscribeId",
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

			const message = new SubscribeUpdateMessage({
				trackPriority: 1,
				minGroupSequence: 0,
				maxGroupSequence: 100,
			});
			const err = await message.encode(mockWriter);
			assertEquals(err instanceof Error, true);
		},
	);

	await t.step(
		"encode should return error when writeVarint fails for trackPriority",
		async () => {
			let callCount = 0;
			const mockWriter: Writer = {
				async write(p: Uint8Array): Promise<[number, Error | undefined]> {
					callCount++;
					if (callCount > 2) {
						return [0, new Error("Write failed")];
					}
					return [p.length, undefined];
				},
			};

			const message = new SubscribeUpdateMessage({
				trackPriority: 1,
				minGroupSequence: 0,
				maxGroupSequence: 100,
			});
			const err = await message.encode(mockWriter);
			assertEquals(err instanceof Error, true);
		},
	);

	await t.step(
		"encode should return error when writeVarint fails for minGroupSequence",
		async () => {
			let callCount = 0;
			const mockWriter: Writer = {
				async write(p: Uint8Array): Promise<[number, Error | undefined]> {
					callCount++;
					if (callCount > 3) {
						return [0, new Error("Write failed")];
					}
					return [p.length, undefined];
				},
			};

			const message = new SubscribeUpdateMessage({
				trackPriority: 1,
				minGroupSequence: 0,
				maxGroupSequence: 100,
			});
			const err = await message.encode(mockWriter);
			assertEquals(err instanceof Error, true);
		},
	);

	await t.step(
		"encode should return error when writeVarint fails for maxGroupSequence",
		async () => {
			let callCount = 0;
			const mockWriter: Writer = {
				async write(p: Uint8Array): Promise<[number, Error | undefined]> {
					callCount++;
					if (callCount > 3) {
						return [0, new Error("Write failed")];
					}
					return [p.length, undefined];
				},
			};

			const message = new SubscribeUpdateMessage({
				trackPriority: 1,
				minGroupSequence: 0,
				maxGroupSequence: 100,
			});
			const err = await message.encode(mockWriter);
			assertEquals(err instanceof Error, true);
		},
	);
});
