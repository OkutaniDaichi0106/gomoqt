import { assert, assertEquals } from "@std/assert";
import { SessionServerMessage } from "./session_server.ts";
import { Buffer } from "@okdaichi/golikejs/bytes";
import type { Writer } from "@okdaichi/golikejs/io";

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
			extensions: new Map<number, Uint8Array>([[
				100,
				new Uint8Array([0xff, 0xfe, 0xfd]),
			]]),
		},
	};

	for (const [caseName, input] of Object.entries(testCases)) {
		await t.step(caseName, async () => {
			// Encode using Buffer
			const buffer = Buffer.make(200);
			const message = new SessionServerMessage(input);
			const encodeErr = await message.encode(buffer);
			assertEquals(encodeErr, undefined, `encode failed for ${caseName}`);

			// Decode from a new buffer with written data
			const readBuffer = Buffer.make(200);
			await readBuffer.write(buffer.bytes());
			const decodedMessage = new SessionServerMessage({});
			const decodeErr = await decodedMessage.decode(readBuffer);
			assertEquals(decodeErr, undefined, `decode failed for ${caseName}`);
			assertEquals(
				decodedMessage.version,
				input.version,
				`version mismatch for ${caseName}`,
			);
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
			const buffer = Buffer.make(0); // Empty buffer
			const message = new SessionServerMessage({});
			const err = await message.decode(buffer);
			assertEquals(err !== undefined, true);
		},
	);

	await t.step(
		"decode should return error when readVarint fails for version",
		async () => {
			const buffer = Buffer.make(10);
			// message length = 1 (varint), but no version data
			await buffer.write(new Uint8Array([0x01]));
			const message = new SessionServerMessage({});
			const err = await message.decode(buffer);
			assert(err !== undefined);
		},
	);

	await t.step(
		"decode should return error when readVarint fails for extension count",
		async () => {
			const buffer = Buffer.make(10);
			// message length = 2 (varint), version = 1 (varint), but no extension count
			await buffer.write(new Uint8Array([0x02, 0x01]));
			const message = new SessionServerMessage({});
			const err = await message.decode(buffer);
			assert(err !== undefined);
		},
	);

	await t.step(
		"decode should return error when reading extension ID fails",
		async () => {
			const buffer = Buffer.make(10);
			// message length = 3 (varint), version = 1 (varint), extensionCount = 1 (varint), but no ID
			await buffer.write(new Uint8Array([0x03, 0x01, 0x01]));
			const message = new SessionServerMessage({});
			const err = await message.decode(buffer);
			assert(err !== undefined);
		},
	);

	await t.step(
		"decode should return error when reading extension data fails",
		async () => {
			const buffer = Buffer.make(10);
			// message length = 4 (varint), version = 1 (varint), extensionCount = 1 (varint), ID = 1 (varint), but no data
			await buffer.write(new Uint8Array([0x04, 0x01, 0x01, 0x01]));
			const message = new SessionServerMessage({});
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

			const message = new SessionServerMessage({
				version: 1,
				extensions: new Map(),
			});
			const err = await message.encode(mockWriter);
			assertEquals(err instanceof Error, true);
		},
	);

	await t.step(
		"encode should return error when writing version fails",
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

			const message = new SessionServerMessage({
				version: 1,
				extensions: new Map(),
			});
			const err = await message.encode(mockWriter);
			assertEquals(err instanceof Error, true);
		},
	);

	await t.step(
		"encode should return error when writing extensions size fails",
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

			const message = new SessionServerMessage({
				version: 1,
				extensions: new Map(),
			});
			const err = await message.encode(mockWriter);
			assertEquals(err instanceof Error, true);
		},
	);

	await t.step(
		"encode should return error when writing extension ID fails",
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

			const message = new SessionServerMessage({
				version: 1,
				extensions: new Map([[1, new Uint8Array([1, 2, 3])]]),
			});
			const err = await message.encode(mockWriter);
			assertEquals(err instanceof Error, true);
		},
	);

	await t.step(
		"encode should return error when writing extension data fails",
		async () => {
			let callCount = 0;
			const mockWriter: Writer = {
				async write(p: Uint8Array): Promise<[number, Error | undefined]> {
					callCount++;
					if (callCount > 4) {
						return [0, new Error("Write failed")];
					}
					return [p.length, undefined];
				},
			};

			const message = new SessionServerMessage({
				version: 1,
				extensions: new Map([[1, new Uint8Array([1, 2, 3])]]),
			});
			const err = await message.encode(mockWriter);
			assertEquals(err instanceof Error, true);
		},
	);
});
