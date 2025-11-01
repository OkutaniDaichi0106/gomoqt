import { assertEquals, assertExists, assertInstanceOf, createMock } from "@std/assert";
import type { TrackConfig } from "./subscribe_stream.ts";
import { ReceiveSubscribeStream, SendSubscribeStream } from "./subscribe_stream.ts";
import type { SubscribeID } from "./subscribe_id.ts";
import type { SubscribeMessage, SubscribeOkMessage } from "./internal/message/mod.ts";
import type { Reader, Writer } from "./internal/webtransport/mod.ts";
import { background } from "@okudai/golikejs/context";
import type { Info } from "./info.ts";
import { StreamError } from "./internal//webtransport/mod.ts";

function makeSendMocks() {
	const mockWriter = {
		writeVarint: createMock(),
		writeBoolean: createMock(),
		writeBigVarint: createMock(),
		writeString: createMock(),
		writeUint8Array: createMock(),
		writeUint8: createMock(),
		flush: createMock<() => Promise<Error | undefined>>().mockResolvedValue(undefined),
		close: createMock().mockReturnValue(undefined),
		cancel: createMock().mockReturnValue(undefined),
		closed: createMock().mockReturnValue(Promise.resolve()),
	} as any as Writer;

	const mockReader = {
		readVarint: createMock(),
		readBoolean: createMock(),
		readBigVarint: createMock(),
		readString: createMock(),
		readStringArray: createMock(),
		readUint8Array: createMock(),
		readUint8: createMock(),
		copy: createMock(),
		fill: createMock(),
		cancel: createMock().mockReturnValue(undefined),
		closed: createMock().mockReturnValue(Promise.resolve()),
	} as any as Reader;

	const mockSubscribe: SubscribeMessage = {
		subscribeId: 123n,
		broadcastPath: "/test/path",
		trackName: "test-track",
		trackPriority: 1,
		minGroupSequence: 0n,
		maxGroupSequence: 100n,
	};

	const mockSubscribeOk: SubscribeOkMessage = {
		groupPeriod: 100,
		messageLength: 0,
		encode: createMock(),
		decode: createMock(),
	} as any;

	const ctx = background();
	const sendStream = new SendSubscribeStream(
		ctx,
		mockWriter,
		mockReader,
		mockSubscribe,
		mockSubscribeOk,
	);

	return { ctx, mockWriter, mockReader, mockSubscribe, mockSubscribeOk, sendStream } as const;
}

function makeReceiveMocks() {
	const mockWriter = {
		writeVarint: createMock(),
		writeBoolean: createMock(),
		writeBigVarint: createMock(),
		writeString: createMock(),
		writeUint8Array: createMock(),
		writeUint8: createMock(),
		flush: createMock<() => Promise<Error | undefined>>().mockResolvedValue(undefined),
		close: createMock().mockReturnValue(undefined),
		cancel: createMock().mockReturnValue(undefined),
		closed: createMock().mockReturnValue(Promise.resolve()),
	} as any as Writer;

	const mockReader = {
		readVarint: createMock().mockResolvedValue([0, new Error("EOF")]),
		readBoolean: createMock(),
		readBigVarint: createMock(),
		readString: createMock(),
		readStringArray: createMock(),
		readUint8Array: createMock(),
		readUint8: createMock(),
		copy: createMock(),
		fill: createMock(),
		cancel: createMock().mockReturnValue(undefined),
		closed: createMock().mockReturnValue(Promise.resolve()),
	} as any as Reader;

	const mockSubscribe: SubscribeMessage = {
		subscribeId: 789n,
		broadcastPath: "/receive/path",
		trackName: "receive-track",
		trackPriority: 3,
		minGroupSequence: 5n,
		maxGroupSequence: 150n,
	};

	const ctx = background();
	const receiveStream = new ReceiveSubscribeStream(ctx, mockWriter, mockReader, mockSubscribe);
	return { ctx, mockWriter, mockReader, mockSubscribe, receiveStream } as const;
}

Deno.test("SendSubscribeStream - basic behavior", async (t) => {
	await t.step("constructor and basic getters", () => {
		const { sendStream } = makeSendMocks();
		assertInstanceOf(sendStream, SendSubscribeStream);
		assertExists(sendStream.context);
		assertEquals(sendStream.subscribeId, 123n);
		const config = sendStream.config;
		assertEquals(config.trackPriority, 1);
		assertEquals(config.minGroupSequence, 0n);
		assertEquals(config.maxGroupSequence, 100n);
		assertEquals(sendStream.info, undefined as unknown); // info is provided by mockSubscribeOk in some tests; keep minimal check
	});

	await t.step("update - success path updates config and flush called", async () => {
		const { sendStream, mockWriter } = makeSendMocks();
		const newConfig = { trackPriority: 2, minGroupSequence: 10n, maxGroupSequence: 200n };
		const result = await sendStream.update(newConfig);
		assertEquals(result, undefined);
		// mockWriter.flush should have at least one call
		assertEquals(mockWriter.flush.calls.length > 0, true);
		const cfg = sendStream.config;
		assertEquals(cfg.trackPriority, 2);
		assertEquals(cfg.minGroupSequence, 10n);
		assertEquals(cfg.maxGroupSequence, 200n);
		// cleanup
		if (typeof sendStream.closeWithError === "function") {
			await sendStream.closeWithError(999, "test cleanup");
		}
	});

	await t.step("update - flush failure returns error", async () => {
		const { sendStream, mockWriter } = makeSendMocks();
		mockWriter.flush.mockResolvedValue(new Error("Flush failed"));
		const result = await sendStream.update({
			trackPriority: 2,
			minGroupSequence: 10n,
			maxGroupSequence: 200n,
		});
		// result should be an Error; check message for expected pattern
		assertInstanceOf(result, Error);
		const msg = (result as Error).message ?? "";
		assertEquals(/Failed to (write|flush) subscribe update/.test(msg), true);
		if (typeof sendStream.closeWithError === "function") {
			await sendStream.closeWithError(999, "test cleanup");
		}
	});

	await t.step("closeWithError cancels writer and sets context error", async () => {
		const { sendStream, mockWriter } = makeSendMocks();
		await sendStream.closeWithError(500, "Test error");
		// verify cancel was called with a StreamError-like object
		assertEquals(mockWriter.cancel.calls.length > 0, true);
		const firstArg = mockWriter.cancel.calls[0]?.[0];
		assertInstanceOf(firstArg, StreamError);
		assertInstanceOf(sendStream.context.err(), StreamError);
	});
});

Deno.test("ReceiveSubscribeStream - basic behavior", async (t) => {
	await t.step("constructor and basic getters", () => {
		const { receiveStream } = makeReceiveMocks();
		assertInstanceOf(receiveStream, ReceiveSubscribeStream);
		assertExists(receiveStream.context);
		assertEquals(receiveStream.subscribeId, 789n);
		const cfg = receiveStream.trackConfig;
		assertEquals(cfg.trackPriority, 3);
		assertEquals(cfg.minGroupSequence, 5n);
		assertEquals(cfg.maxGroupSequence, 150n);
	});

	await t.step("writeInfo writes successfully and is idempotent", async () => {
		const { receiveStream } = makeReceiveMocks();
		const info: Info = { groupPeriod: 100 };
		const r1 = await receiveStream.writeInfo(info);
		assertEquals(r1, undefined);
		const r2 = await receiveStream.writeInfo(info);
		assertEquals(r2, undefined);
	});

	await t.step("close cancels writer and leaves context cleared", async () => {
		const { receiveStream, mockWriter } = makeReceiveMocks();
		await receiveStream.close();
		assertEquals(mockWriter.close.calls.length > 0, true);
		assertEquals(receiveStream.context.err(), undefined);
	});

	await t.step("closeWithError cancels writer and reader and sets context error", async () => {
		const { receiveStream, mockWriter, mockReader } = makeReceiveMocks();
		await receiveStream.closeWithError(404, "Not found");
		assertEquals(mockWriter.cancel.calls.length > 0, true);
		assertEquals(mockReader.cancel.calls.length > 0, true);
		assertInstanceOf(receiveStream.context.err(), StreamError);
	});
});

Deno.test("Type definitions - TrackConfig and SubscribeID", async (t) => {
	await t.step("TrackConfig shape", () => {
		const config: TrackConfig = {
			trackPriority: 1,
			minGroupSequence: 0n,
			maxGroupSequence: 100n,
		};
		assertEquals(typeof config.trackPriority, "number");
		assertEquals(typeof config.minGroupSequence, "bigint");
		assertEquals(typeof config.maxGroupSequence, "bigint");
	});

	await t.step("SubscribeID is bigint", () => {
		const id: SubscribeID = 123n;
		assertEquals(typeof id, "bigint");
	});
});
