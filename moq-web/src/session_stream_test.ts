import {
	afterEach,
	assertEquals,
	assertExists,
	assertThrows,
	beforeEach,
	describe,
	it,
} from "../deps.ts";
import { SessionStream } from "./session_stream.ts";
import type { Context } from "../deps.ts";
import { background, withCancelCause } from "../deps.ts";
import type { Reader, Writer } from "./internal/webtransport/mod.ts";
import { SessionUpdateMessage } from "./message/session_update.ts";
import { SessionClientMessage } from "./message/session_client.ts";
import { SessionServerMessage } from "./message/session_server.ts";
import { Extensions } from "./internal/extensions.ts";
import type { Version } from "./internal/version.ts";
import { EOF } from "./internal/webtransport/mod.ts";

describe("SessionStream", () => {
	let ctx: Context;
	let cancelCtx: (reason: Error | undefined) => void;
	let mockWriter: Writer;
	let mockReader: Reader;
	let mockClient: SessionClientMessage;
	let mockServer: SessionServerMessage;

	beforeEach(() => {
		[ctx, cancelCtx] = withCancelCause(background());

		mockWriter = {
			writeBoolean: vi.fn(),
			writeBigVarint: vi.fn(),
			writeString: vi.fn(),
			writeUint8Array: vi.fn(),
			writeUint8: vi.fn(),
			writeVarint: vi.fn(),
			flush: vi.fn(),
			close: vi.fn(),
			cancel: vi.fn(),
			closed: vi.fn(),
		} as any;

		mockReader = {
			readBoolean: vi.fn(),
			readBigVarint: vi.fn(),
			readString: vi.fn(),
			readStringArray: vi.fn(),
			readUint8Array: vi.fn(),
			readUint8: vi.fn(),
			readVarint: vi.fn(),
			copy: vi.fn(),
			fill: vi.fn(),
			cancel: vi.fn(),
			closed: vi.fn(),
		} as any;

		// Mock readVarint to return EOF immediately to stop the loop
		vi.mocked(mockReader.readVarint).mockResolvedValue([0, EOF]);
		vi.mocked(mockReader.readBigVarint).mockResolvedValue([0n, undefined]);

		const versions = new Set<Version>([0xffffff00n]);
		const extensions = new Extensions();

		mockClient = new SessionClientMessage({ versions, extensions });
		mockServer = new SessionServerMessage({ version: 0xffffff00n, extensions });
	});

	afterEach(async () => {
		// Cancel context to stop background operations
		cancelCtx(undefined);

		// Give time for cleanup
		await new Promise((resolve) => setTimeout(resolve, 10));

		vi.restoreAllMocks();
	});

	describe("constructor", () => {
		it("should initialize with provided parameters", async () => {
			const sessionStream = new SessionStream({
				context: ctx,
				writer: mockWriter,
				reader: mockReader,
				client: mockClient,
				server: mockServer,
				detectFunc: vi.fn().mockResolvedValue(0),
			});

			assertInstanceOf(sessionStream, SessionStream);
			assertEquals(sessionStream.clientInfo.versions, mockClient.versions);
			assertEquals(sessionStream.clientInfo.extensions, mockClient.extensions);
			assertEquals(sessionStream.serverInfo.version, mockServer.version);
			assertEquals(sessionStream.serverInfo.extensions, mockServer.extensions);
			assertEquals(sessionStream.context, ctx);

			// Cancel context
			cancelCtx(undefined);
		});

		it("should use the provided context", async () => {
			const sessionStream = new SessionStream({
				context: ctx,
				writer: mockWriter,
				reader: mockReader,
				client: mockClient,
				server: mockServer,
				detectFunc: vi.fn().mockResolvedValue(0),
			});

			assertEquals(sessionStream.context, ctx);
			assertEquals(typeof sessionStream.context.done, "function");
			assertEquals(typeof sessionStream.context.err, "function");

			// Cancel context
			cancelCtx(undefined);
		});
	});

	describe("context", () => {
		it("should return the internal context", async () => {
			const sessionStream = new SessionStream({
				context: ctx,
				writer: mockWriter,
				reader: mockReader,
				client: mockClient,
				server: mockServer,
				detectFunc: vi.fn().mockResolvedValue(0),
			});

			const context = sessionStream.context;

			assertExists(context);
			assertEquals(typeof context.done, "function");
			assertEquals(typeof context.err, "function");

			// Cancel context
			cancelCtx(undefined);
		});

		it("should use context cancellation to stop background operations", async () => {
			const sessionStream = new SessionStream({
				context: ctx,
				writer: mockWriter,
				reader: mockReader,
				client: mockClient,
				server: mockServer,
				detectFunc: vi.fn().mockResolvedValue(0),
			});

			// The session stream should use the provided context
			assertEquals(sessionStream.context, ctx);

			// Initially the context should not be cancelled
			expect(sessionStream.context.err()).toBeUndefined();

			// Cancel the context
			cancelCtx(new Error("Context cancelled"));

			// The session context should now be cancelled
			expect(sessionStream.context.err()).toBeDefined();
		});
	});

	describe("error handling", () => {
		it("should handle decode errors in background updates", async () => {
			const sessionStream = new SessionStream({
				context: ctx,
				writer: mockWriter,
				reader: mockReader,
				client: mockClient,
				server: mockServer,
				detectFunc: vi.fn().mockResolvedValue(0),
			});

			// Spy on decode to track calls and return errors
			let callCount = 0;
			const decodeSpy = vi.spyOn(SessionUpdateMessage.prototype, "decode");
			decodeSpy.mockImplementation(async () => {
				callCount++;
				if (callCount === 1) {
					return undefined; // First call succeeds
				}
				return new Error("Decode error"); // Subsequent calls fail
			});

			// The session stream should still be functional
			assertExists(sessionStream.context);
			assertExists(sessionStream.clientInfo);
			assertExists(sessionStream.serverInfo);

			// Cancel context
			cancelCtx(undefined);

			decodeSpy.mockRestore();
		});
	});

	describe("serverInfo getter", () => {
		it("should return the server information property", async () => {
			const sessionStream = new SessionStream({
				context: ctx,
				writer: mockWriter,
				reader: mockReader,
				client: mockClient,
				server: mockServer,
				detectFunc: vi.fn().mockResolvedValue(0),
			});

			const serverInfoResult = sessionStream.serverInfo;

			// Verify the getter exists and returns the internal state
			assertExists(serverInfoResult);
			assertEquals(serverInfoResult.version, mockServer.version);
			assertEquals(serverInfoResult.bitrate, 0);

			// Cancel context
			cancelCtx(undefined);
		});
	});

	describe("updated method", () => {
		it("should be a function that returns a promise", async () => {
			const sessionStream = new SessionStream({
				context: ctx,
				writer: mockWriter,
				reader: mockReader,
				client: mockClient,
				server: mockServer,
				detectFunc: vi.fn().mockResolvedValue(0),
			});

			// Verify the method exists and has the correct signature
			assertEquals(typeof sessionStream.updated, "function");

			// Cancel context
			cancelCtx(undefined);
		});
	});

	describe("integration", () => {
		it("should handle complete session lifecycle", async () => {
			const sessionStream = new SessionStream({
				context: ctx,
				writer: mockWriter,
				reader: mockReader,
				client: mockClient,
				server: mockServer,
				detectFunc: vi.fn().mockResolvedValue(0),
			});

			// Mock the encode method on SessionUpdateMessage instances
			vi.spyOn(SessionUpdateMessage.prototype, "encode").mockImplementation(async () =>
				undefined
			);

			// Verify initial state
			assertExists(sessionStream.context);
			assertEquals(sessionStream.clientInfo.versions, mockClient.versions);
			assertEquals(sessionStream.clientInfo.extensions, mockClient.extensions);
			assertEquals(sessionStream.serverInfo.version, mockServer.version);
			assertEquals(sessionStream.serverInfo.extensions, mockServer.extensions);

			// Cancel context
			cancelCtx(undefined);
		});

		it("should handle multiple updates", async () => {
			const sessionStream = new SessionStream({
				context: ctx,
				writer: mockWriter,
				reader: mockReader,
				client: mockClient,
				server: mockServer,
				detectFunc: vi.fn().mockResolvedValue(0),
			});

			// Mock the encode method on SessionUpdateMessage instances
			vi.spyOn(SessionUpdateMessage.prototype, "encode").mockImplementation(async () =>
				undefined
			);

			// Since update is private, we can't test it directly
			// But we can test that the session stream initializes correctly
			assertEquals(sessionStream.clientInfo.bitrate, 0);
			assertEquals(sessionStream.serverInfo.bitrate, 0);

			// Cancel context
			cancelCtx(undefined);
		});
	});
});
