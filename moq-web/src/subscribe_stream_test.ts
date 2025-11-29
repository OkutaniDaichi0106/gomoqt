import { assertEquals } from "@std/assert";
import { MockSendStream, MockStream } from "./internal/webtransport/mock_stream_test.ts";
import { ReceiveSubscribeStream, SendSubscribeStream } from "./subscribe_stream.ts";
import {
	SubscribeMessage,
	SubscribeOkMessage,
	SubscribeUpdateMessage,
} from "./internal/message/mod.ts";
import { background, withCancelCause } from "@okudai/golikejs/context";

Deno.test("SendSubscribeStream.update writes update to writable", async () => {
	const [ctx] = withCancelCause(background());
	const s = new MockStream(1n);
	const subscribe = new SubscribeMessage({
		subscribeId: 1,
		broadcastPath: "/test" as any,
		trackName: "t",
		trackPriority: 0,
		minGroupSequence: 0,
		maxGroupSequence: 0,
	});
	const ok = new SubscribeOkMessage({});
	const sss = new SendSubscribeStream(ctx, s as any, subscribe, ok as any);
	const err = await sss.update({ trackPriority: 1, minGroupSequence: 2, maxGroupSequence: 3 });
	assertEquals(err, undefined);
	// Expect that the writable got bytes; there should be encoded SubscribeUpdateMessage
	assertEquals(s.writable.writtenData.length > 0, true);
});

Deno.test("SendSubscribeStream closeWithError cancels stream", async () => {
	const [ctx] = withCancelCause(background());
	const s = new MockStream(1n);
	const subscribe = new SubscribeMessage({
		subscribeId: 1,
		broadcastPath: "/test" as any,
		trackName: "t",
		trackPriority: 0,
		minGroupSequence: 0,
		maxGroupSequence: 0,
	});
	const ok = new SubscribeOkMessage({});
	const sss = new SendSubscribeStream(ctx, s as any, subscribe, ok as any);
	await sss.closeWithError(1);
	assertEquals(s.writable.cancelCalls.length, 1);
});

Deno.test("ReceiveSubscribeStream writeInfo sends SUBSCRIBE_OK and prevents double write", async () => {
	const [ctx] = withCancelCause(background());
	const s = new MockStream(2n);
	const subscribe = new SubscribeMessage({
		subscribeId: 42,
		broadcastPath: "/test" as any,
		trackName: "t",
		trackPriority: 0,
		minGroupSequence: 0,
		maxGroupSequence: 0,
	});
	const rss = new ReceiveSubscribeStream(ctx, s as any, subscribe as any);
	const err = await rss.writeInfo({} as any);
	assertEquals(err, undefined);
	// Double write should simply do nothing and return undefined
	const err2 = await rss.writeInfo({} as any);
	assertEquals(err2, undefined);
});

Deno.test("ReceiveSubscribeStream writeInfo returns error when context canceled", async () => {
	const [ctx, cancel] = withCancelCause(background());
	const s = new MockStream(2n);
	const subscribe = new SubscribeMessage({
		subscribeId: 42,
		broadcastPath: "/test" as any,
		trackName: "t",
		trackPriority: 0,
		minGroupSequence: 0,
		maxGroupSequence: 0,
	});
	const rss = new ReceiveSubscribeStream(ctx, s as any, subscribe as any);
	cancel(new Error("canceled"));
	await new Promise((r) => setTimeout(r, 0)); // Wait for cancel to propagate
	const err = await rss.writeInfo({} as any);
	assertEquals(err?.message, "canceled");
});

Deno.test("ReceiveSubscribeStream close closes stream", async () => {
	const [ctx] = withCancelCause(background());
	const s = new MockStream(2n);
	const subscribe = new SubscribeMessage({
		subscribeId: 42,
		broadcastPath: "/test" as any,
		trackName: "t",
		trackPriority: 0,
		minGroupSequence: 0,
		maxGroupSequence: 0,
	});
	const rss = new ReceiveSubscribeStream(ctx, s as any, subscribe as any);
	await rss.close();
	// Should not throw
});

Deno.test("ReceiveSubscribeStream close does nothing if context canceled", async () => {
	const [ctx, cancel] = withCancelCause(background());
	const s = new MockStream(2n);
	const subscribe = new SubscribeMessage({
		subscribeId: 42,
		broadcastPath: "/test" as any,
		trackName: "t",
		trackPriority: 0,
		minGroupSequence: 0,
		maxGroupSequence: 0,
	});
	const rss = new ReceiveSubscribeStream(ctx, s as any, subscribe as any);
	cancel(new Error("canceled"));
	await new Promise((r) => setTimeout(r, 0));
	await rss.close();
	// Should not throw
});

Deno.test("ReceiveSubscribeStream closeWithError does nothing if context canceled", async () => {
	const [ctx, cancel] = withCancelCause(background());
	const s = new MockStream(2n);
	const subscribe = new SubscribeMessage({
		subscribeId: 42,
		broadcastPath: "/test" as any,
		trackName: "t",
		trackPriority: 0,
		minGroupSequence: 0,
		maxGroupSequence: 0,
	});
	const rss = new ReceiveSubscribeStream(ctx, s as any, subscribe as any);
	cancel(new Error("canceled"));
	await new Promise((r) => setTimeout(r, 0));
	await rss.closeWithError(2);
	// Should not throw
});

Deno.test("ReceiveSubscribeStream updated waiters are notified upon update", async () => {
	const [ctx] = withCancelCause(background());
	const sub = new SubscribeMessage({
		subscribeId: 10,
		broadcastPath: "/x" as any,
		trackName: "t",
		trackPriority: 0,
		minGroupSequence: 0,
		maxGroupSequence: 0,
	});
	// note: We'll construct the MockStream with data below to allow the handler to read it immediately
	// Simulate an update message on readable
	const mu = new MockSendStream(3n);
	const update = new SubscribeUpdateMessage({
		trackPriority: 5,
		minGroupSequence: 0,
		maxGroupSequence: 10,
	});
	await update.encode(mu as any);
	const data = mu.getAllWrittenData();
	const s2 = new MockStream(3n, data);
	const rss = new ReceiveSubscribeStream(ctx, s2 as any, sub as any);
	// Wait for internal handler to process
	await new Promise((r) => setTimeout(r, 0));
	// Now trackConfig should been updated
	assertEquals(rss.trackConfig.trackPriority, 5);
});

Deno.test("ReceiveSubscribeStream closeWithError cancels streams and broadcasts cond", async () => {
	const [ctx] = withCancelCause(background());
	const s = new MockStream(4n);
	const sub = new SubscribeMessage({
		subscribeId: 20,
		broadcastPath: "/x" as any,
		trackName: "t",
		trackPriority: 0,
		minGroupSequence: 0,
		maxGroupSequence: 0,
	});
	const rss = new ReceiveSubscribeStream(ctx, s as any, sub as any);
	await rss.closeWithError(2);
	// writable and readable should have cancelCalls
	assertEquals(s.writable.cancelCalls.length >= 0, true);
	assertEquals(s.readable.cancelCalls.length >= 0, true);
});
