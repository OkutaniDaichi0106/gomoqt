import { assertEquals, assertExists, assertInstanceOf } from "@std/assert";
import type { TrackConfig } from "./subscribe_stream.ts";
import { ReceiveSubscribeStream, SendSubscribeStream } from "./subscribe_stream.ts";
import { SubscribeMessage, SubscribeOkMessage } from "./internal/message/mod.ts";
import { background } from "@okudai/golikejs/context";
import type { Info } from "./info.ts";
import { SubscribeID } from "./alias.ts";
import { MockStream } from "./internal/webtransport/mock_stream_test.ts";

/**
 * Creates mock objects for ReceiveSubscribeStream testing.
 * Uses dependency injection pattern for clean mocking.
 */
function makeReceiveMocks() {
	const mockStream = new MockStream(42n);

	const mockSubscribe = new SubscribeMessage({
		subscribeId: 789,
		broadcastPath: "/receive/path",
		trackName: "receive-track",
		trackPriority: 3,
		minGroupSequence: 5n,
		maxGroupSequence: 150n,
	});

	const ctx = background();
	const receiveStream = new ReceiveSubscribeStream(ctx, mockStream as any, mockSubscribe);
	return {
		ctx,
		mockWriter: mockStream.writable,
		mockReader: mockStream.readable,
		mockSubscribe,
		receiveStream,
		mockStream,
	} as const;
}

Deno.test("SendSubscribeStream - Normal Cases", async (t) => {
	await t.step("should create instance with correct properties", () => {
		const mockStream = new MockStream(42n);
		const mockSubscribe = new SubscribeMessage({
			subscribeId: 123,
			broadcastPath: "/test/path",
			trackName: "test-track",
			trackPriority: 1,
			minGroupSequence: 0n,
			maxGroupSequence: 100n,
		});
		const mockSubscribeOk = new SubscribeOkMessage({
			groupPeriod: 100,
			messageLength: 0,
		});
		const ctx = background();
		const sendStream = new SendSubscribeStream(
			ctx,
			mockStream as any,
			mockSubscribe,
			mockSubscribeOk,
		);

		assertInstanceOf(sendStream, SendSubscribeStream);
		assertExists(sendStream.context);
		assertEquals(sendStream.subscribeId, 123);

		// Verify config matches subscribe message
		const config = sendStream.config;
		assertEquals(config.trackPriority, 1);
		assertEquals(config.minGroupSequence, 0n);
		assertEquals(config.maxGroupSequence, 100n);

		// Verify info matches subscribeOk message
		assertEquals(sendStream.info, mockSubscribeOk);
	});

	await t.step("should update config and call flush", async () => {
		const mockStream = new MockStream(42n);
		const mockSubscribe = new SubscribeMessage({
			subscribeId: 123,
			broadcastPath: "/test/path",
			trackName: "test-track",
			trackPriority: 1,
			minGroupSequence: 0n,
			maxGroupSequence: 100n,
		});
		const mockSubscribeOk = new SubscribeOkMessage({
			groupPeriod: 100,
			messageLength: 0,
		});
		const ctx = background();
		const sendStream = new SendSubscribeStream(
			ctx,
			mockStream as any,
			mockSubscribe,
			mockSubscribeOk,
		);
		const mockWriter = mockStream.writable;

		const newConfig: TrackConfig = {
			trackPriority: 2,
			minGroupSequence: 10n,
			maxGroupSequence: 200n,
		};

		const result = await sendStream.update(newConfig);

		assertEquals(result, undefined);
		assertEquals((mockWriter.flush as any).calls.length > 0, true);

		// Verify config was actually updated
		const cfg = sendStream.config;
		assertEquals(cfg.trackPriority, 2);
		assertEquals(cfg.minGroupSequence, 10n);
		assertEquals(cfg.maxGroupSequence, 200n);

		// Cleanup
		mockStream.reset();
	});
});

Deno.test("SendSubscribeStream - Error Cases", async (t) => {
	await t.step("should return error when flush fails", async () => {
		const mockStream = new MockStream(42n);
		mockStream.writable.flushError = new Error("Flush failed");

		const mockSubscribe = new SubscribeMessage({
			subscribeId: 123,
			broadcastPath: "/test/path",
			trackName: "test-track",
			trackPriority: 1,
			minGroupSequence: 0n,
			maxGroupSequence: 100n,
		});
		const mockSubscribeOk = new SubscribeOkMessage({
			groupPeriod: 100,
			messageLength: 0,
		});
		const ctx = background();
		const sendStream = new SendSubscribeStream(
			ctx,
			mockStream as any,
			mockSubscribe,
			mockSubscribeOk,
		);

		const result = await sendStream.update({
			trackPriority: 2,
			minGroupSequence: 10n,
			maxGroupSequence: 200n,
		});

		// Verify error is returned
		assertInstanceOf(result, Error);
		const msg = (result as Error).message ?? "";
		assertEquals(/Failed to (write|flush) subscribe update/.test(msg), true);

		// Cleanup
		mockStream.reset();
	});

	await t.step("should cancel writer when closed with error", async () => {
		const mockStream = new MockStream(42n);
		const mockSubscribe = new SubscribeMessage({
			subscribeId: 123,
			broadcastPath: "/test/path",
			trackName: "test-track",
			trackPriority: 1,
			minGroupSequence: 0n,
			maxGroupSequence: 100n,
		});
		const mockSubscribeOk = new SubscribeOkMessage({
			groupPeriod: 100,
			messageLength: 0,
		});
		const ctx = background();
		const sendStream = new SendSubscribeStream(
			ctx,
			mockStream as any,
			mockSubscribe,
			mockSubscribeOk,
		);
		const mockWriter = mockStream.writable;

		await sendStream.closeWithError(500, "Test error");

		// Verify cancel was called on writer
		assertEquals((mockWriter.cancel as any).calls.length > 0, true);

		// Cleanup
		mockStream.reset();
	});
});

Deno.test("ReceiveSubscribeStream - Normal Cases", async (t) => {
	await t.step("should create instance with correct properties", () => {
		const { receiveStream, mockStream } = makeReceiveMocks();

		assertInstanceOf(receiveStream, ReceiveSubscribeStream);
		assertExists(receiveStream.context);
		assertEquals(receiveStream.subscribeId, 789);

		// Verify track config matches subscribe message
		const cfg = receiveStream.trackConfig;
		assertEquals(cfg.trackPriority, 3);
		assertEquals(cfg.minGroupSequence, 5n);
		assertEquals(cfg.maxGroupSequence, 150n);

		// Cleanup
		mockStream.reset();
	});

	await t.step("should write info successfully (idempotent)", async () => {
		const { receiveStream, mockStream } = makeReceiveMocks();
		const info: Info = { groupPeriod: 100 };

		// First write
		const r1 = await receiveStream.writeInfo(info);
		assertEquals(r1, undefined);

		// Second write should also succeed (idempotent)
		const r2 = await receiveStream.writeInfo(info);
		assertEquals(r2, undefined);

		// Cleanup
		mockStream.reset();
	});

	await t.step("should close writer cleanly", async () => {
		const { receiveStream, mockWriter, mockStream } = makeReceiveMocks();

		await receiveStream.close();

		// Verify close was called
		assertEquals((mockWriter.close as any).calls.length > 0, true);
		// Context should not have error
		assertEquals(receiveStream.context.err(), undefined);

		// Cleanup
		mockStream.reset();
	});
});

Deno.test("ReceiveSubscribeStream - Error Cases", async (t) => {
	await t.step("should cancel both streams when closed with error", async () => {
		const { receiveStream, mockWriter, mockReader, mockStream } = makeReceiveMocks();

		await receiveStream.closeWithError(404, "Not found");

		// Verify both writer and reader were cancelled
		assertEquals((mockWriter.cancel as any).calls.length > 0, true);
		assertEquals((mockReader.cancel as any).calls.length > 0, true);

		// Cleanup
		mockStream.reset();
	});
});

Deno.test("Type Definitions - TrackConfig and SubscribeID", async (t) => {
	await t.step("should validate TrackConfig structure", () => {
		const config: TrackConfig = {
			trackPriority: 1,
			minGroupSequence: 0n,
			maxGroupSequence: 100n,
		};

		// Verify all properties have correct types
		assertEquals(typeof config.trackPriority, "number");
		assertEquals(typeof config.minGroupSequence, "bigint");
		assertEquals(typeof config.maxGroupSequence, "bigint");
	});

	await t.step("should validate SubscribeID is number type", () => {
		const id: SubscribeID = 123;
		assertEquals(typeof id, "number");
	});
});

Deno.test("ReceiveSubscribeStream - Additional Coverage", async (t) => {
	await t.step("should return error when context is cancelled", async () => {
		const { receiveStream, mockStream } = makeReceiveMocks();

		await receiveStream.closeWithError(500, "Context cancelled");

		const info: Info = { groupPeriod: 100 };
		const result = await receiveStream.writeInfo(info);

		assertExists(result);
		assertInstanceOf(result, Error);

		mockStream.reset();
	});

	await t.step("should return early when close is called with existing error", async () => {
		const { receiveStream, mockWriter, mockStream } = makeReceiveMocks();

		await receiveStream.closeWithError(500, "First error");

		await receiveStream.close();

		assertEquals((mockWriter.close as any).calls.length, 0);

		mockStream.reset();
	});

	await t.step("should return early when closeWithError is called twice", async () => {
		const { receiveStream, mockWriter, mockStream } = makeReceiveMocks();

		await receiveStream.closeWithError(500, "First error");

		const cancelCallsBefore = (mockWriter.cancel as any).calls.length;

		await receiveStream.closeWithError(404, "Second error");

		const cancelCallsAfter = (mockWriter.cancel as any).calls.length;

		assertEquals(cancelCallsBefore, cancelCallsAfter);

		mockStream.reset();
	});
});
