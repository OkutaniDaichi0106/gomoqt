import { assertEquals } from "@std/assert";
import { SessionClientMessage } from "./session_client.ts";
import { Buffer } from "@okudai/golikejs/bytes";
import type { Writer } from "@okudai/golikejs/io";

Deno.test("SessionClientMessage", async (t) => {
	await t.step("should be defined", () => {
		assertEquals(SessionClientMessage !== undefined, true);
	});

	await t.step("should create instance with default values", () => {
		const message = new SessionClientMessage();

		assertEquals(message.versions.size, 0);
		assertEquals(message.extensions.size, 0);
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

	await t.step(
		"should calculate correct length with single version and no extensions",
		() => {
			const versions = new Set<number>([1]);
			const extensions = new Map<number, Uint8Array>();

			const message = new SessionClientMessage({ versions, extensions });
			const length = message.len;

			assertEquals(length > 0, true);
			assertEquals(typeof length, "number");
		},
	);

	await t.step("should calculate correct length with multiple versions", () => {
		const versions = new Set<number>([0xffffff00, 1, 2]);
		const extensions = new Map<number, Uint8Array>();

		const message = new SessionClientMessage({
			versions,
			extensions: extensions,
		});
		const length = message.len;

		assertEquals(length > 0, true);
		assertEquals(typeof length, "number");
	});

	await t.step("should calculate correct length with extensions", () => {
		const versions = new Set<number>([0xffffff00]);
		const extensions = new Map<number, Uint8Array>();
		extensions.set(1, new TextEncoder().encode("test"));
		extensions.set(2, new Uint8Array([1, 2, 3]));

		const message = new SessionClientMessage({ versions, extensions });
		const length = message.len;

		assertEquals(length > 0, true);
		assertEquals(typeof length, "number");
	});

	await t.step("should calculate correct length with no versions and no extensions", () => {
		const versions = new Set<number>();
		const extensions = new Map<number, Uint8Array>();

		const message = new SessionClientMessage({ versions, extensions });
		const length = message.len;

		// Length should be: varint(0) + varint(0) = 2 bytes
		assertEquals(length, 2);
	});

	await t.step(
		"should encode and decode with single version and no extensions",
		async () => {
			const versions = new Set<number>([1]);
			const extensions = new Map<number, Uint8Array>();

			const buffer = Buffer.make(100);
			const message = new SessionClientMessage({ versions, extensions });
			const encodeErr = await message.encode(buffer);
			assertEquals(encodeErr, undefined);

			const readBuffer = Buffer.make(100);
			await readBuffer.write(buffer.bytes());
			const decodedMessage = new SessionClientMessage({
				versions: new Set<number>(),
				extensions: new Map<number, Uint8Array>(),
			});
			const decodeErr = await decodedMessage.decode(readBuffer);
			assertEquals(decodeErr, undefined);
			assertEquals(decodedMessage instanceof SessionClientMessage, true);

			assertEquals(decodedMessage?.versions.size, 1);
			const decodedVersions = Array.from(decodedMessage?.versions || []);
			const originalVersions = Array.from(versions);
			assertEquals(decodedVersions, originalVersions);
			assertEquals(decodedMessage?.extensions.size, 0);
		},
	);

	await t.step("should encode and decode with multiple versions", async () => {
		const versions = new Set<number>([1, 2, 100]);
		const extensions = new Map<number, Uint8Array>();

		const buffer = Buffer.make(100);
		const message = new SessionClientMessage({ versions, extensions });
		const encodeErr = await message.encode(buffer);
		assertEquals(encodeErr, undefined);

		const readBuffer = Buffer.make(100);
		await readBuffer.write(buffer.bytes());
		const decodedMessage = new SessionClientMessage({
			versions: new Set<number>(),
			extensions: new Map<number, Uint8Array>(),
		});
		const decodeErr = await decodedMessage.decode(readBuffer);
		assertEquals(decodeErr, undefined);

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

		const buffer = Buffer.make(200);
		const message = new SessionClientMessage({ versions, extensions });
		const encodeErr = await message.encode(buffer);
		assertEquals(encodeErr, undefined);

		const readBuffer = Buffer.make(200);
		await readBuffer.write(buffer.bytes());
		const decodedMessage = new SessionClientMessage({
			versions: new Set<number>(),
			extensions: new Map<number, Uint8Array>(),
		});
		const decodeErr = await decodedMessage.decode(readBuffer);
		assertEquals(decodeErr, undefined);

		assertEquals(decodedMessage.versions.size, 1);
		assertEquals(decodedMessage.versions.has(1), true);
		assertEquals(decodedMessage.extensions.size, 3);
		assertEquals(
			new TextDecoder().decode(decodedMessage.extensions.get(1)!),
			"test-string",
		);
		assertEquals(
			decodedMessage.extensions.get(2),
			new Uint8Array([1, 2, 3, 4, 5]),
		);
		assertEquals(
			new TextDecoder().decode(decodedMessage.extensions.get(100)!),
			"another-extension",
		);
	});

	await t.step("should encode and decode with empty versions set", async () => {
		const versions = new Set<number>();
		const extensions = new Map<number, Uint8Array>();

		const buffer = Buffer.make(100);
		const message = new SessionClientMessage({ versions, extensions });
		const encodeErr = await message.encode(buffer);
		assertEquals(encodeErr, undefined);

		const readBuffer = Buffer.make(100);
		await readBuffer.write(buffer.bytes());
		const decodedMessage = new SessionClientMessage({
			versions: new Set<number>(),
			extensions: new Map<number, Uint8Array>(),
		});
		const decodeErr = await decodedMessage.decode(readBuffer);
		assertEquals(decodeErr, undefined);

		assertEquals(decodedMessage.versions.size, 0);
		assertEquals(decodedMessage.extensions.size, 0);
	});

	await t.step("should encode and decode with only extensions", async () => {
		const versions = new Set<number>();
		const extensions = new Map<number, Uint8Array>();
		extensions.set(1, new Uint8Array([1, 2, 3]));

		const buffer = Buffer.make(100);
		const message = new SessionClientMessage({ versions, extensions });
		const encodeErr = await message.encode(buffer);
		assertEquals(encodeErr, undefined);

		const readBuffer = Buffer.make(100);
		await readBuffer.write(buffer.bytes());
		const decodedMessage = new SessionClientMessage({
			versions: new Set<number>(),
			extensions: new Map<number, Uint8Array>(),
		});
		const decodeErr = await decodedMessage.decode(readBuffer);
		assertEquals(decodeErr, undefined);

		assertEquals(decodedMessage.versions.size, 0);
		assertEquals(decodedMessage.extensions.size, 1);
	});

	await t.step("should handle empty extension data", async () => {
		const versions = new Set<number>([1]);
		const extensions = new Map<number, Uint8Array>();
		extensions.set(1, new Uint8Array([]));
		extensions.set(2, new TextEncoder().encode(""));

		const buffer = Buffer.make(100);
		const message = new SessionClientMessage({ versions, extensions });
		const encodeErr = await message.encode(buffer);
		assertEquals(encodeErr, undefined);

		const readBuffer = Buffer.make(100);
		await readBuffer.write(buffer.bytes());
		const decodedMessage = new SessionClientMessage({
			versions: new Set<number>(),
			extensions: new Map<number, Uint8Array>(),
		});
		const decodeErr = await decodedMessage.decode(readBuffer);
		assertEquals(decodeErr, undefined);

		assertEquals(decodedMessage.extensions.size, 2);
		assertEquals(decodedMessage.extensions.get(1), new Uint8Array([]));
		assertEquals(
			new TextDecoder().decode(decodedMessage.extensions.get(2)!),
			"",
		);
	});

	await t.step(
		"decode should return error when readUint16 fails for message length",
		async () => {
			const buffer = Buffer.make(0); // Empty buffer
			const message = new SessionClientMessage({
				versions: new Set<number>(),
				extensions: new Map<number, Uint8Array>(),
			});
			const err = await message.decode(buffer);
			assertEquals(err !== undefined, true);
		},
	);

	await t.step(
		"decode should return error when readFull fails",
		async () => {
			const buffer = Buffer.make(10);
			// Write message length = 10, but no data follows
			await buffer.write(new Uint8Array([0x00, 0x0a])); // msgLen = 10 (big-endian)

			const message = new SessionClientMessage({
				versions: new Set<number>(),
				extensions: new Map<number, Uint8Array>(),
			});
			const err = await message.decode(buffer);
			assertEquals(err !== undefined, true);
		},
	);

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

			const message = new SessionClientMessage({
				versions: new Set<number>([1]),
				extensions: new Map<number, Uint8Array>(),
			});
			const err = await message.encode(mockWriter);
			assertEquals(err instanceof Error, true);
		},
	);

	await t.step(
		"encode should return error when writeVarint fails for versions size",
		async () => {
			let callCount = 0;
			const mockWriter: Writer = {
				async write(p: Uint8Array): Promise<[number, Error | undefined]> {
					callCount++;
					if (callCount > 1) {
						return [0, new Error("Write failed on versions size")];
					}
					return [p.length, undefined];
				},
			};

			const message = new SessionClientMessage({
				versions: new Set<number>([1]),
				extensions: new Map<number, Uint8Array>(),
			});
			const err = await message.encode(mockWriter);
			assertEquals(err instanceof Error, true);
		},
	);

	await t.step(
		"encode should return error when writeVarint fails for version value",
		async () => {
			let callCount = 0;
			const mockWriter: Writer = {
				async write(p: Uint8Array): Promise<[number, Error | undefined]> {
					callCount++;
					if (callCount > 2) {
						return [0, new Error("Write failed on version value")];
					}
					return [p.length, undefined];
				},
			};

			const message = new SessionClientMessage({
				versions: new Set<number>([1]),
				extensions: new Map<number, Uint8Array>(),
			});
			const err = await message.encode(mockWriter);
			assertEquals(err instanceof Error, true);
		},
	);

	await t.step(
		"encode should return error when writeVarint fails for second version",
		async () => {
			let callCount = 0;
			const mockWriter: Writer = {
				async write(p: Uint8Array): Promise<[number, Error | undefined]> {
					callCount++;
					if (callCount > 3) {
						return [0, new Error("Write failed on second version")];
					}
					return [p.length, undefined];
				},
			};

			const message = new SessionClientMessage({
				versions: new Set<number>([1, 2]),
				extensions: new Map<number, Uint8Array>(),
			});
			const err = await message.encode(mockWriter);
			assertEquals(err instanceof Error, true);
		},
	);

	await t.step(
		"encode should return error when writeVarint fails for extensions size",
		async () => {
			let callCount = 0;
			const mockWriter: Writer = {
				async write(p: Uint8Array): Promise<[number, Error | undefined]> {
					callCount++;
					if (callCount > 3) {
						return [0, new Error("Write failed on extensions size")];
					}
					return [p.length, undefined];
				},
			};

			const message = new SessionClientMessage({
				versions: new Set<number>([1]),
				extensions: new Map<number, Uint8Array>(),
			});
			const err = await message.encode(mockWriter);
			assertEquals(err instanceof Error, true);
		},
	);

	await t.step(
		"encode should return error when writeVarint fails for extension ID",
		async () => {
			let callCount = 0;
			const mockWriter: Writer = {
				async write(p: Uint8Array): Promise<[number, Error | undefined]> {
					callCount++;
					if (callCount > 4) {
						return [0, new Error("Write failed on extension ID")];
					}
					return [p.length, undefined];
				},
			};

			const message = new SessionClientMessage({
				versions: new Set<number>([1]),
				extensions: new Map<number, Uint8Array>([[1, new Uint8Array([1])]]),
			});
			const err = await message.encode(mockWriter);
			assertEquals(err instanceof Error, true);
		},
	);

	await t.step(
		"encode should return error when writeBytes fails for extension data",
		async () => {
			let callCount = 0;
			const mockWriter: Writer = {
				async write(p: Uint8Array): Promise<[number, Error | undefined]> {
					callCount++;
					if (callCount > 5) {
						return [0, new Error("Write failed on extension data")];
					}
					return [p.length, undefined];
				},
			};

			const message = new SessionClientMessage({
				versions: new Set<number>([1]),
				extensions: new Map<number, Uint8Array>([[1, new Uint8Array([1])]]),
			});
			const err = await message.encode(mockWriter);
			assertEquals(err instanceof Error, true);
		},
	);

	await t.step(
		"encode should return error when writeVarint fails for second extension ID",
		async () => {
			let callCount = 0;
			const mockWriter: Writer = {
				async write(p: Uint8Array): Promise<[number, Error | undefined]> {
					callCount++;
					if (callCount > 7) {
						return [0, new Error("Write failed on second extension ID")];
					}
					return [p.length, undefined];
				},
			};

			const message = new SessionClientMessage({
				versions: new Set<number>([1]),
				extensions: new Map<number, Uint8Array>([
					[1, new Uint8Array([1])],
					[2, new Uint8Array([2])],
				]),
			});
			const err = await message.encode(mockWriter);
			assertEquals(err instanceof Error, true);
		},
	);

	await t.step(
		"encode should return error when writeBytes fails for second extension data",
		async () => {
			let callCount = 0;
			const mockWriter: Writer = {
				async write(p: Uint8Array): Promise<[number, Error | undefined]> {
					callCount++;
					if (callCount > 8) {
						return [0, new Error("Write failed on second extension data")];
					}
					return [p.length, undefined];
				},
			};

			const message = new SessionClientMessage({
				versions: new Set<number>([1]),
				extensions: new Map<number, Uint8Array>([
					[1, new Uint8Array([1])],
					[2, new Uint8Array([2])],
				]),
			});
			const err = await message.encode(mockWriter);
			assertEquals(err instanceof Error, true);
		},
	);
});
