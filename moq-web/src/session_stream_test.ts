import { assertEquals, assertExists } from "@std/assert";
import { spy } from "@std/testing/mock";
import { SessionStream } from "./session_stream.ts";
import { background, withCancelCause } from "@okudai/golikejs/context";
import {
	SessionClientMessage,
	SessionServerMessage,
	SessionUpdateMessage,
} from "./internal/message/mod.ts";
import { MockReceiveStream, MockSendStream, MockStream } from "./mock_stream_test.ts";
import { Buffer } from "@okudai/golikejs/bytes";
import { DEFAULT_VERSION } from "./version.ts";
import { EOFError } from "@okudai/golikejs/io";

Deno.test("SessionStream", async (t) => {
	await t.step("constructor initializes clientInfo and serverInfo correctly", async () => {
		const [ctx] = withCancelCause(background());
		const mockStream = new MockStream({ id: 1n });
		const client = new SessionClientMessage({
			versions: new Set([DEFAULT_VERSION]),
			extensions: new Map(),
		});
		const server = new SessionServerMessage({
			version: DEFAULT_VERSION,
			extensions: new Map(),
		});

		const ss = new SessionStream({
			context: ctx,
			stream: mockStream,
			client,
			server,
			detectFunc: async () => 0,
		});

		assertEquals(ss.clientInfo.versions.has(DEFAULT_VERSION), true);
		assertEquals(ss.serverInfo.version, DEFAULT_VERSION);
		assertEquals(ss.clientInfo.bitrate, 0);
		assertEquals(ss.serverInfo.bitrate, 0);
		assertExists(ss.context);

		await ss.waitForBackgroundTasks();
	});

	await t.step("handleUpdates updates serverInfo.bitrate on receiving update", async () => {
		const [ctx] = withCancelCause(background());

		// Encode a SessionUpdateMessage
		const updateMsg = new SessionUpdateMessage({ bitrate: 5000 });
		const encodeBuf = Buffer.make(128);
		await updateMsg.encode(encodeBuf);
		const updateData = encodeBuf.bytes();

		let readOffset = 0;
		const mockReadable = new MockReceiveStream({
			id: 2n,
			read: spy(async (p: Uint8Array) => {
				if (readOffset >= updateData.length) {
					// Wait indefinitely after data is consumed (simulating no more updates)
					return await new Promise<[number, Error | undefined]>(() => {});
				}
				const n = Math.min(p.length, updateData.length - readOffset);
				p.set(updateData.subarray(readOffset, readOffset + n));
				readOffset += n;
				return [n, undefined] as [number, Error | undefined];
			}),
		});
		const mockStream = new MockStream({ id: 2n, readable: mockReadable });

		const client = new SessionClientMessage({
			versions: new Set([DEFAULT_VERSION]),
			extensions: new Map(),
		});
		const server = new SessionServerMessage({
			version: DEFAULT_VERSION,
			extensions: new Map(),
		});

		const ss = new SessionStream({
			context: ctx,
			stream: mockStream,
			client,
			server,
			detectFunc: async () => 0,
		});

		// Wait for the update to be processed
		await new Promise((r) => setTimeout(r, 50));

		assertEquals(ss.serverInfo.bitrate, 5000);
	});

	await t.step("context cancellation cancels streams", async () => {
		const [ctx, cancel] = withCancelCause(background());

		const writableCancelCalls: number[] = [];
		const readableCancelCalls: number[] = [];
		const mockWritable = new MockSendStream({
			id: 4n,
			cancel: spy(async (code: number) => {
				writableCancelCalls.push(code);
			}),
		});
		const mockReadable = new MockReceiveStream({
			id: 4n,
			cancel: spy(async (code: number) => {
				readableCancelCalls.push(code);
			}),
		});
		const mockStream = new MockStream({
			id: 4n,
			writable: mockWritable,
			readable: mockReadable,
		});

		const client = new SessionClientMessage({
			versions: new Set([DEFAULT_VERSION]),
			extensions: new Map(),
		});
		const server = new SessionServerMessage({
			version: DEFAULT_VERSION,
			extensions: new Map(),
		});

		new SessionStream({
			context: ctx,
			stream: mockStream,
			client,
			server,
			detectFunc: async () => 0,
		});

		// Cancel the context
		cancel(new Error("test cancel"));
		await new Promise((r) => setTimeout(r, 10));

		assertEquals(writableCancelCalls.length, 1);
		assertEquals(readableCancelCalls.length, 1);
	});

	await t.step("handleUpdates cancels context on decode error", async () => {
		const [ctx] = withCancelCause(background());

		// Provide invalid data that will cause decode error
		const invalidData = new Uint8Array([0x80, 0x80, 0x80, 0x80, 0x80]); // Invalid varint
		let readOffset = 0;
		const mockReadable = new MockReceiveStream({
			id: 5n,
			read: spy(async (p: Uint8Array) => {
				if (readOffset >= invalidData.length) {
					return [0, new EOFError()] as [number, Error | undefined];
				}
				const n = Math.min(p.length, invalidData.length - readOffset);
				p.set(invalidData.subarray(readOffset, readOffset + n));
				readOffset += n;
				return [n, undefined] as [number, Error | undefined];
			}),
		});
		const mockStream = new MockStream({ id: 5n, readable: mockReadable });

		const client = new SessionClientMessage({
			versions: new Set([DEFAULT_VERSION]),
			extensions: new Map(),
		});
		const server = new SessionServerMessage({
			version: DEFAULT_VERSION,
			extensions: new Map(),
		});

		const ss = new SessionStream({
			context: ctx,
			stream: mockStream,
			client,
			server,
			detectFunc: async () => 0,
		});

		// Wait for error to be processed
		await new Promise((r) => setTimeout(r, 50));

		// Context should have error
		assertExists(ss.context.err());

		await ss.waitForBackgroundTasks();
	});

	await t.step("waitForBackgroundTasks waits for all tasks to complete", async () => {
		const [ctx] = withCancelCause(background());

		// EOF immediately to end handleUpdates quickly
		const mockReadable = new MockReceiveStream({
			id: 6n,
			read: spy(async () => [0, new EOFError()] as [number, Error | undefined]),
		});
		const mockStream = new MockStream({ id: 6n, readable: mockReadable });

		const client = new SessionClientMessage({
			versions: new Set([DEFAULT_VERSION]),
			extensions: new Map(),
		});
		const server = new SessionServerMessage({
			version: DEFAULT_VERSION,
			extensions: new Map(),
		});

		const ss = new SessionStream({
			context: ctx,
			stream: mockStream,
			client,
			server,
			detectFunc: async () => 0,
		});

		// This should complete without hanging
		await ss.waitForBackgroundTasks();

		// Context should have error due to EOF
		assertExists(ss.context.err());
	});

	await t.step("clientInfo and serverInfo extensions are accessible", async () => {
		const [ctx] = withCancelCause(background());
		const mockStream = new MockStream({ id: 7n });

		const clientExtensions = new Map<number, Uint8Array>([
			[1, new Uint8Array([1, 2, 3])],
		]);
		const serverExtensions = new Map<number, Uint8Array>([
			[2, new Uint8Array([4, 5, 6])],
		]);

		const client = new SessionClientMessage({
			versions: new Set([DEFAULT_VERSION]),
			extensions: clientExtensions,
		});
		const server = new SessionServerMessage({
			version: DEFAULT_VERSION,
			extensions: serverExtensions,
		});

		const ss = new SessionStream({
			context: ctx,
			stream: mockStream,
			client,
			server,
			detectFunc: async () => 0,
		});

		assertExists(ss.clientInfo.extensions);
		assertExists(ss.serverInfo.extensions);

		await ss.waitForBackgroundTasks();
	});
});
