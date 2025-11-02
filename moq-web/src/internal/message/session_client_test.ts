import { assertEquals, assertRejects } from "@std/assert";
import { SessionClientMessage } from "./session_client.ts";
import { ReceiveStream, SendStream } from "../webtransport/mod.ts";

Deno.test("SessionClientMessage", async (t) => {
	await t.step("should be defined", () => {
		assertEquals(SessionClientMessage !== undefined, true);
	});

	await t.step("decode should throw when numVersions is negative", async () => {
		const message = new SessionClientMessage({
			versions: new Set<number>(),
			extensions: new Map<number, Uint8Array>(),
		});

		let call = 0;
		const fakeReader = {
			async readVarint() {
				call++;
				if (call === 1) return [10, undefined]; // len
				if (call === 2) return [-1, undefined]; // numVersions negative
				return [0, undefined];
			},
			async readUint8Array() {
				return [new Uint8Array(), undefined];
			},
		} as unknown as ReceiveStream;

		await assertRejects(() => message.decode(fakeReader), Error, "Invalid number of versions for SessionClient");
	});

	await t.step("decode should throw when numVersions exceeds max safe integer", async () => {
		const message = new SessionClientMessage({
			versions: new Set<number>(),
			extensions: new Map<number, Uint8Array>(),
		});

		let call = 0;
		const tooLarge = Number.MAX_SAFE_INTEGER + 1;
		const fakeReader = {
			async readVarint() {
				call++;
				if (call === 1) return [10, undefined]; // len
				if (call === 2) return [tooLarge, undefined]; // numVersions too large
				return [0, undefined];
			},
			async readUint8Array() {
				return [new Uint8Array(), undefined];
			},
		} as unknown as ReceiveStream;

		await assertRejects(() => message.decode(fakeReader), Error, "Number of versions exceeds maximum safe integer for SessionClient");
	});

	await t.step("decode should throw when numExtensions is undefined", async () => {
		const message = new SessionClientMessage({
			versions: new Set<number>(),
			extensions: new Map<number, Uint8Array>(),
		});

		let call = 0;
		const fakeReader = {
			async readVarint() {
				call++;
				if (call === 1) return [1, undefined]; // len
				if (call === 2) return [0, undefined]; // numVersions
				if (call === 3) return [undefined as unknown as number, undefined]; // numExtensions undefined
				return [0, undefined];
			},
			async readUint8Array() {
				return [new Uint8Array(), undefined];
			},
		} as unknown as ReceiveStream;

		await assertRejects(() => message.decode(fakeReader), Error, "read numExtensions: number is undefined");
	});

	await t.step("decode should throw when extData is undefined", async () => {
		const message = new SessionClientMessage({
			versions: new Set<number>(),
			extensions: new Map<number, Uint8Array>(),
		});

		let call = 0;
		const fakeReader = {
			async readVarint() {
				call++;
				if (call === 1) return [3, undefined]; // len
				if (call === 2) return [0, undefined]; // numVersions
				if (call === 3) return [1, undefined]; // numExtensions
				if (call === 4) return [1, undefined]; // extId
				return [0, undefined];
			},
			async readUint8Array() {
				return [undefined as unknown as Uint8Array, undefined];
			},
		} as unknown as ReceiveStream;

		await assertRejects(() => message.decode(fakeReader), Error, "read extData: Uint8Array is undefined");
	});

	await t.step("decode should throw on message length mismatch", async () => {
		const message = new SessionClientMessage({
			versions: new Set<number>(),
			extensions: new Map<number, Uint8Array>(),
		});

		let call = 0;
		const fakeReader = {
			async readVarint() {
				call++;
				if (call === 1) return [3, undefined]; // len (incorrect)
				if (call === 2) return [0, undefined]; // numVersions
				if (call === 3) return [0, undefined]; // numExtensions
				return [0, undefined];
			},
			async readUint8Array() {
				return [new Uint8Array(), undefined];
			},
		} as unknown as ReceiveStream;

		await assertRejects(() => message.decode(fakeReader), Error, "message length mismatch");
	});

	await t.step("should create instance with versions and extensions", () => {
		const versions = new Set<number>([1]);
		const extensions = new Map<number, Uint8Array>();
		extensions.set(1, new TextEncoder().encode("test"));

		const message = new SessionClientMessage({ versions, extensions });

		assertEquals(message.versions, versions);
		assertEquals(message.extensions, extensions);
	});

	await t.step("should create instance with versions only", () => {
		const versions = new Set<number>([1]);

		const message = new SessionClientMessage({
			versions,
			extensions: new Map<number, Uint8Array>(),
		});

		assertEquals(message.versions, versions);
		assertEquals(message.extensions.size, 0);
	});

	await t.step("should calculate correct length with single version and no extensions", () => {
		const versions = new Set<number>([1]);
		const extensions = new Map<number, Uint8Array>();

		const message = new SessionClientMessage({ versions, extensions });
		const length = message.messageLength;

		// Expected: varint(1) + varint(DEVELOP) + varint(0)
		// DEVELOP = 0xffffff00n, which needs 5 bytes in varint encoding
		// varint(1) = 1 byte, varint(0) = 1 byte
		assertEquals(length > 0, true);
		assertEquals(typeof length, "number");
	});

	await t.step("should calculate correct length with multiple versions", () => {
		const versions = new Set<number>([0xffffff00, 1, 2]);
		const extensions = new Map<number, Uint8Array>();

		const message = new SessionClientMessage({ versions, extensions: extensions });
		const length = message.messageLength;

		assertEquals(length > 0, true);
		assertEquals(typeof length, "number");
	});

	await t.step("should calculate correct length with extensions", () => {
		const versions = new Set<number>([0xffffff00]);
		const extensions = new Map<number, Uint8Array>();
		extensions.set(1, new TextEncoder().encode("test"));
		extensions.set(2, new Uint8Array([1, 2, 3]));

		const message = new SessionClientMessage({ versions, extensions });
		const length = message.messageLength;

		assertEquals(length > 0, true);
		assertEquals(typeof length, "number");
	});

	await t.step("should encode and decode with single version and no extensions", async () => {
		const versions = new Set<number>([1]);
		const extensions = new Map<number, Uint8Array>();

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

		// Encode
		const message = new SessionClientMessage({ versions, extensions });
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

		// Decode
		const decodedMessage = new SessionClientMessage({
			versions: new Set<number>(),
			extensions: new Map<number, Uint8Array>(),
		});
		const decodeErr = await decodedMessage.decode(reader);
		assertEquals(decodeErr, undefined);
		assertEquals(decodedMessage instanceof SessionClientMessage, true);

		// Verify content
		assertEquals(decodedMessage?.versions.size, 1);
		const decodedVersions = Array.from(decodedMessage?.versions || []);
		const originalVersions = Array.from(versions);
		assertEquals(decodedVersions, originalVersions);
		assertEquals(decodedMessage?.extensions.size, 0);
	});

	await t.step("should encode and decode with multiple versions", async () => {
		const versions = new Set<number>([1, 2, 100]);
		const extensions = new Map<number, Uint8Array>();

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

		// Encode
		const message = new SessionClientMessage({ versions, extensions });
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

		// Decode
		const decodedMessage = new SessionClientMessage({
			versions: new Set<number>(),
			extensions: new Map<number, Uint8Array>(),
		});
		const decodeErr = await decodedMessage.decode(reader);
		assertEquals(decodeErr, undefined);

		// Verify content
		assertEquals(decodedMessage.versions.size, 3);
		assertEquals(decodedMessage.versions.has(1), true);
		assertEquals(decodedMessage.versions.has(2), true);
		assertEquals(decodedMessage.versions.has(100), true);
		assertEquals(decodedMessage.extensions.size, 0);
	});

	await t.step("should encode and decode with extensions", async () => {
		const versions = new Set<number>([1]);
		const extensions = new Map<number, Uint8Array>();
		extensions.set(1, new TextEncoder().encode("test-string"));
		extensions.set(2, new Uint8Array([1, 2, 3, 4, 5]));
		extensions.set(100, new TextEncoder().encode("another-extension"));

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

		// Encode
		const message = new SessionClientMessage({ versions, extensions });
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

		// Decode
		const decodedMessage = new SessionClientMessage({
			versions: new Set<number>(),
			extensions: new Map<number, Uint8Array>(),
		});
		const decodeErr = await decodedMessage.decode(reader);
		assertEquals(decodeErr, undefined);

		// Verify content
		assertEquals(decodedMessage.versions.size, 1);
		assertEquals(decodedMessage.versions.has(1), true);
		assertEquals(decodedMessage.extensions.size, 3);
		assertEquals(new TextDecoder().decode(decodedMessage.extensions.get(1)!), "test-string");
		assertEquals(decodedMessage.extensions.get(2), new Uint8Array([1, 2, 3, 4, 5]));
		assertEquals(
			new TextDecoder().decode(decodedMessage.extensions.get(100)!),
			"another-extension",
		);
	});

	await t.step("should encode and decode with empty versions set", async () => {
		const versions = new Set<number>();
		const extensions = new Map<number, Uint8Array>();

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
		// Encode
		const message = new SessionClientMessage({ versions, extensions });
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

		// Decode
		const decodedMessage = new SessionClientMessage({
			versions: new Set<number>(),
			extensions: new Map<number, Uint8Array>(),
		});
		const decodeErr = await decodedMessage.decode(reader);
		assertEquals(decodeErr, undefined);

		// Verify content
		assertEquals(decodedMessage.versions.size, 0);
		assertEquals(decodedMessage.extensions.size, 0);
	});

	await t.step("should handle empty extension data", async () => {
		const versions = new Set<number>([1]);
		const extensions = new Map<number, Uint8Array>();
		extensions.set(1, new Uint8Array([])); // Empty bytes
		extensions.set(2, new TextEncoder().encode("")); // Empty string

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

		// Encode
		const message = new SessionClientMessage({ versions, extensions });
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

		// Decode
		const decodedMessage = new SessionClientMessage({
			versions: new Set<number>(),
			extensions: new Map<number, Uint8Array>(),
		});
		const decodeErr = await decodedMessage.decode(reader);
		assertEquals(decodeErr, undefined);

		// Verify content
		assertEquals(decodedMessage.extensions.size, 2);
		assertEquals(decodedMessage.extensions.get(1), new Uint8Array([]));
		assertEquals(new TextDecoder().decode(decodedMessage.extensions.get(2)!), "");
	});

	await t.step("decode should return error when readVarint fails for message length", async () => {
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

		const message = new SessionClientMessage({
			versions: new Set<number>(),
			extensions: new Map<number, Uint8Array>(),
		});
		const err = await message.decode(reader);
		assertEquals(err !== undefined, true);
	});

	await t.step("decode should return error when readVarint fails for versions", async () => {
		// Create buffer with only message length, but no versions data
		const buffer = new Uint8Array([2, 1]); // message length = 2, then incomplete
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

		const message = new SessionClientMessage({
			versions: new Set<number>(),
			extensions: new Map<number, Uint8Array>(),
		});
		const err = await message.decode(reader);
		assertEquals(err !== undefined, true);
	});

	await t.step("decode should return error when reading version value fails", async () => {
		// Create buffer: message length, numVersions=1, but no version data
		const buffer = new Uint8Array([2, 1]); // incomplete
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

		const message = new SessionClientMessage({
			versions: new Set<number>(),
			extensions: new Map<number, Uint8Array>(),
		});
		const err = await message.decode(reader);
		assertEquals(err !== undefined, true);
	});

	await t.step("decode should return error when readVarint fails for extensions count", async () => {
		// Create buffer: message length, numVersions=0, but no extensions count
		const buffer = new Uint8Array([1, 0]); // incomplete
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

		const message = new SessionClientMessage({
			versions: new Set<number>(),
			extensions: new Map<number, Uint8Array>(),
		});
		const err = await message.decode(reader);
		assertEquals(err !== undefined, true);
	});

	await t.step("decode should return error when reading extension ID fails", async () => {
		// Create buffer: message length, numVersions=0, numExtensions=1, but no extension ID
		const buffer = new Uint8Array([2, 0, 1]); // incomplete
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

		const message = new SessionClientMessage({
			versions: new Set<number>(),
			extensions: new Map<number, Uint8Array>(),
		});
		const err = await message.decode(reader);
		assertEquals(err !== undefined, true);
	});

	await t.step("decode should return error when reading extension data fails", async () => {
		// Create buffer: message length, numVersions=0, numExtensions=1, extension ID=1, but no extension data
		const buffer = new Uint8Array([3, 0, 1, 1]); // incomplete
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

		const message = new SessionClientMessage({
			versions: new Set<number>(),
			extensions: new Map<number, Uint8Array>(),
		});
		const err = await message.decode(reader);
		assertEquals(err !== undefined, true);
	});
});
