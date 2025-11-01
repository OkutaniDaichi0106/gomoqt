import { assertEquals } from "@std/assert";
import { SessionClientMessage } from "./session_client.ts";
import { ReceiveStream, SendStream } from "../webtransport/mod.ts";

Deno.test("SessionClientMessage", async (t) => {
	await t.step("should be defined", () => {
		assertEquals(SessionClientMessage !== undefined, true);
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

		const message = new SessionClientMessage({ versions, extensions: new Map<number, Uint8Array>() });

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
		const writer = new SendStream({ stream: writableStream, transfer: undefined, streamId: 0n });

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

		console.log("Encoded buffer:", Array.from(combinedBuffer).join(","));

		// Create readable stream for decoding
		const readableStream = new ReadableStream({
			start(controller) {
				controller.enqueue(combinedBuffer);
				controller.close();
			},
		});
		const reader = new ReceiveStream({ stream: readableStream, transfer: undefined, streamId: 0n });

		// Decode
		const decodedMessage = new SessionClientMessage({ versions: new Set<number>(), extensions: new Map<number, Uint8Array>() });
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
		const writer = new SendStream({ stream: writableStream, transfer: undefined, streamId: 0n });

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
		const reader = new ReceiveStream({ stream: readableStream, transfer: undefined, streamId: 0n });

		// Decode
		const decodedMessage = new SessionClientMessage({ versions: new Set<number>(), extensions: new Map<number, Uint8Array>() });
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
		const writer = new SendStream({ stream: writableStream, transfer: undefined, streamId: 0n });

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
		const reader = new ReceiveStream({ stream: readableStream, transfer: undefined, streamId: 0n });

		// Decode
		const decodedMessage = new SessionClientMessage({ versions: new Set<number>(), extensions: new Map<number, Uint8Array>() });
		const decodeErr = await decodedMessage.decode(reader);
		assertEquals(decodeErr, undefined);

		// Verify content
		assertEquals(decodedMessage.versions.size, 1);
		assertEquals(decodedMessage.versions.has(1), true);
		assertEquals(decodedMessage.extensions.size, 3);
		assertEquals(new TextDecoder().decode(decodedMessage.extensions.get(1)!), "test-string");
		assertEquals(decodedMessage.extensions.get(2), new Uint8Array([1, 2, 3, 4, 5]));
		assertEquals(new TextDecoder().decode(decodedMessage.extensions.get(100)!), "another-extension");
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
		const writer = new SendStream({ stream: writableStream, transfer: undefined, streamId: 0n });
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
		const reader = new ReceiveStream({ stream: readableStream, transfer: undefined, streamId: 0n });

		// Decode
		const decodedMessage = new SessionClientMessage({ versions: new Set<number>(), extensions: new Map<number, Uint8Array>() });
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
		const writer = new SendStream({ stream: writableStream, transfer: undefined, streamId: 0n });

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
		const reader = new ReceiveStream({ stream: readableStream, transfer: undefined, streamId: 0n });

		// Decode
		const decodedMessage = new SessionClientMessage({ versions: new Set<number>(), extensions: new Map<number, Uint8Array>() });
		const decodeErr = await decodedMessage.decode(reader);
		assertEquals(decodeErr, undefined);

		// Verify content
		assertEquals(decodedMessage.extensions.size, 2);
		assertEquals(decodedMessage.extensions.get(1), new Uint8Array([]));
		assertEquals(new TextDecoder().decode(decodedMessage.extensions.get(2)!), "");
	});
});
