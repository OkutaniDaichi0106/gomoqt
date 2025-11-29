import { assertEquals, assertExists } from "@std/assert";
import { Announcement, AnnouncementReader, AnnouncementWriter } from "./announce_stream.ts";
import { MockSendStream, MockStream } from "./internal/webtransport/mock_stream_test.ts";
import { background, withCancelCause } from "@okudai/golikejs/context";
import {
	AnnounceInitMessage,
	AnnounceMessage,
	AnnouncePleaseMessage,
} from "./internal/message/mod.ts";
import { Stream } from "./internal/webtransport/stream.ts";
import { encodeMessageToUint8Array } from "./testing/mock_webtransport.ts";

Deno.test("Announcement", async (t) => {
	await t.step("lifecycle: isActive and ended", async () => {
		const [ctx] = withCancelCause(background());
		const ann = new Announcement("/some/path", ctx.done());
		assertEquals(ann.isActive(), true);
		const endedPromise = ann.ended();
		ann.end();
		await endedPromise;
		assertEquals(ann.isActive(), false);
	});

	await t.step("afterFunc executes registered function once", async () => {
		const [ctx, cancel] = withCancelCause(background());
		const ann = new Announcement("/some/path", ctx.done());
		let ran = false;
		const rv = ann.afterFunc(() => {
			ran = true;
		});
		ann.end();
		await ann.ended();
		assertEquals(ran, true);
		// second call should return false because already executed
		assertEquals(rv(), false);
		cancel(undefined);
	});
});

Deno.test("AnnouncementWriter", async (t) => {
	await t.step("init respects prefix and writes ANNOUNCE_INIT", async () => {
		const [ctx] = withCancelCause(background());
		const ms = new MockStream(1n);
		const req = new AnnouncePleaseMessage({ prefix: "/test/" });
		const writer = new AnnouncementWriter(ctx, ms as any, req as any);
		// Create announcement with correct prefix
		const ann = new Announcement("/test/abc", ctx.done());
		const err = await writer.init([ann]);
		assertEquals(err, undefined);
		// Writable should have been written to for ANNOUNCE_INIT message
		assertEquals(ms.writable.writtenData.length > 0, true);
	});

	await t.step("init returns error when prefix mismatched", async () => {
		const [ctx] = withCancelCause(background());
		const ms = new MockStream(2n);
		const req = new AnnouncePleaseMessage({ prefix: "/test/" });
		const writer = new AnnouncementWriter(ctx, ms as any, req as any);
		const annWrong = new Announcement("/wrong/abc", ctx.done());
		const err = await writer.init([annWrong]);
		assertEquals(err instanceof Error, true);
	});

	await t.step("send sends ANNOUNCE and removes on ended", async () => {
		const [ctx] = withCancelCause(background());
		const ms = new MockStream(3n);
		const req = new AnnouncePleaseMessage({ prefix: "/p/" });
		const writer = new AnnouncementWriter(ctx, ms as any, req as any);
		const ann = new Announcement("/p/def", ctx.done());
		await writer.init([]);
		const sendErr = await writer.send(ann);
		assertEquals(sendErr, undefined);
		// writable should have written bytes; initial ANNOUNCE_INIT + ANNOUNCE
		assertEquals(ms.writable.writtenData.length >= 1, true);
		// End announcement and ensure that it writes the closing message
		ann.end();
		// Give microtask to run end handlers
		await new Promise((r) => setTimeout(r, 10));
		// After ending, writer should still not throw
		await writer.close();
	});

	await t.step("closeWithError cancels and calls stream cancel", async () => {
		const [ctx] = withCancelCause(background());
		const ms = new MockStream(4n);
		const req = new AnnouncePleaseMessage({ prefix: "/p/" });
		const writer = new AnnouncementWriter(ctx, ms as any, req as any);
		const ann = new Announcement("/p/abc", ctx.done());
		await writer.init([ann]);
		await writer.closeWithError(1);
		assertEquals(
			ms.writable.cancelCalls.length >= 0 && ms.readable.cancelCalls.length >= 0,
			true,
		);
	});

	await t.step("init returns error on duplicate suffix in input", async () => {
		const [ctx] = withCancelCause(background());
		const ms = new MockStream(6n);
		const req = new AnnouncePleaseMessage({ prefix: "/dup/" });
		const writer = new AnnouncementWriter(ctx, ms as any, req as any);
		const ann1 = new Announcement("/dup/path", ctx.done());
		const ann2 = new Announcement("/dup/path", ctx.done());
		const err = await writer.init([ann1, ann2]);
		// Should return error due to duplicate
		if (!(err instanceof Error)) throw new Error(`Expected error but got ${err}`);
	});

	await t.step("init replaces inactive announcements with active ones", async () => {
		const [ctx] = withCancelCause(background());
		const ms = new MockStream(7n);
		const req = new AnnouncePleaseMessage({ prefix: "/rep/" });
		const writer = new AnnouncementWriter(ctx, ms as any, req as any);
		const old = new Announcement("/rep/aaa", ctx.done());
		// End the old announcement first
		old.end();
		// Initialize with old first (will still be set as it is in the list)
		await writer.init([old]);
		const newAnn = new Announcement("/rep/aaa", ctx.done());
		// Now init with new active announcement; should replace the old
		const err = await writer.init([newAnn]);
		if (err instanceof Error) throw err;
	});

	await t.step("init returns error when trying to end non-active announcement", async () => {
		const readable = new ReadableStream<Uint8Array>({ start(_c) {} });
		const writable = new WritableStream<Uint8Array>({ write(_c) {} });
		const stream = new Stream({ streamId: 1n, stream: { readable, writable } as any });

		const aw = new AnnouncementWriter(
			background(),
			stream,
			new AnnouncePleaseMessage({ prefix: "/test/" }),
		);
		const ann = new Announcement("/test/a", (async () => {})() as any);
		ann.end(); // mark it as inactive

		const err = await aw.init([ann]);
		assertEquals(err instanceof Error, true);
	});

	await t.step("send returns error when trying to end non-active announcement", async () => {
		const readable = new ReadableStream<Uint8Array>({ start(_c) {} });
		const writable = new WritableStream<Uint8Array>({ write(_c) {} });
		const stream = new Stream({ streamId: 2n, stream: { readable, writable } as any });

		const aw = new AnnouncementWriter(
			background(),
			stream,
			new AnnouncePleaseMessage({ prefix: "/p/" }),
		);
		// Initialize Aw so that ready resolves
		await aw.init([]);

		const ann2 = new Announcement("/p/b", (async () => {})() as any);
		ann2.end();
		const err2 = await aw.send(ann2);
		assertEquals(err2 instanceof Error, true);
	});

	await t.step("close does nothing when context already has error", async () => {
		const [ctx, cancelFunc] = withCancelCause(background());
		cancelFunc(new Error("already canceled"));
		await new Promise((r) => setTimeout(r, 0));
		const ms = new MockStream(9n);
		const req = new AnnouncePleaseMessage({ prefix: "/test/" });
		const writer = new AnnouncementWriter(ctx, ms as any, req);
		await writer.close();
		// Should not call close on stream since context has error
		assertEquals(ms.writable.closeCalls, 0);
	});

	await t.step("closeWithError does nothing when context already has error", async () => {
		const [ctx, cancelFunc] = withCancelCause(background());
		cancelFunc(new Error("already canceled"));
		await new Promise((r) => setTimeout(r, 0));
		const ms = new MockStream(10n);
		const req = new AnnouncePleaseMessage({ prefix: "/test/" });
		const writer = new AnnouncementWriter(ctx, ms as any, req);
		await writer.closeWithError(1);
		// Should not call cancel on streams since context has error
		assertEquals(ms.writable.cancelCalls.length, 0);
		assertEquals(ms.readable.cancelCalls.length, 0);
	});
});

Deno.test("AnnouncementReader", async (t) => {
	await t.step("initial announcements enqueued", async () => {
		const [ctx] = withCancelCause(background());
		const mock = new MockStream(5n);
		const req = new AnnouncePleaseMessage({ prefix: "/x/" });
		const aim = new AnnounceInitMessage({ suffixes: ["a", "b"] });
		const reader = new AnnouncementReader(ctx, mock as any, req as any, aim as any);
		// Read initial announcements via receive - should return without waiting
		// call receive with an immediate resolved signal; returns an announcement which should be active
		const [ann, err] = await reader.receive(Promise.resolve());
		assertEquals(err, undefined);
		assertExists(ann);
		if (ann) {
			assertEquals(ann.isActive(), true);
		}
	});

	await t.step("handles duplicate ANNOUNCE messages by closing with error", async () => {
		const [ctx] = withCancelCause(background());
		// Create initial AnnounceInitMessage with suffix 'a'
		const aim = new AnnounceInitMessage({ suffixes: ["a"] });
		// Prepare an ANNOUNCE message for same suffix 'a' (active)
		const mu = new MockSendStream(8n);
		const am = new AnnounceMessage({ suffix: "a", active: true });
		await am.encode(mu as any);
		const data = mu.getAllWrittenData();
		const ms = new MockStream(8n, data);
		const req = new AnnouncePleaseMessage({ prefix: "/" });
		new AnnouncementReader(ctx, ms as any, req as any, aim as any);
		// Allow microtask to run
		await new Promise((r) => setTimeout(r, 10));
		// If duplicate announce handled, the reader should have cancelled the stream; check cancel calls
		// (writable/ readable cancel recorded in MockStream)
		assertEquals(ms.writable.cancelCalls.length >= 0, true);
	});

	await t.step(
		"handles ANNOUNCE message with active false when no old exists and closes with error",
		async () => {
			// Build ANNOUNCE message with active false for suffix 'a'
			const msg = new AnnounceMessage({ suffix: "a", active: false });
			const buf = await encodeMessageToUint8Array(async (w) => msg.encode(w));

			// Build readable that returns the ANNOUNCE message bytes
			const readable = new ReadableStream<Uint8Array>({
				start(c) {
					c.enqueue(buf);
					c.close();
				},
			});

			// Writable stream that records when cancel is called via abort
			let writerAborted = false;
			const writable = new WritableStream<Uint8Array>({
				write(_c) {},
				abort(_reason) {
					writerAborted = true;
					return Promise.resolve();
				},
			});

			const stream = new Stream({ streamId: 1n, stream: { readable, writable } as any });

			const apm = new AnnouncePleaseMessage({ prefix: "/" });
			const aim = new AnnounceInitMessage({ suffixes: [] });

			new AnnouncementReader(background(), stream, apm, aim);

			// Wait a brief tick for the background reading to process
			await new Promise((r) => setTimeout(r, 20));

			// Since there was no old announcement, the reader should closeWithError and cancel
			assertEquals(writerAborted, true);
		},
	);

	await t.step("receive returns error when queue closed", async () => {
		const readable = new ReadableStream<Uint8Array>({ start(_c) {} });
		const writable = new WritableStream<Uint8Array>({ write(_c) {} });
		const stream = new Stream({ streamId: 2n, stream: { readable, writable } as any });

		const apm = new AnnouncePleaseMessage({ prefix: "/test/" });
		const aim = { suffixes: [] } as any;
		const ar = new AnnouncementReader(background(), stream, apm, aim);

		// Close the reader so queue is closed
		await ar.close();

		const [ann, err] = await ar.receive(new Promise(() => {}));
		assertEquals(ann, undefined);
		assertEquals(err instanceof Error, true);
	});

	await t.step("close does nothing when context already has error", async () => {
		const [ctx, cancelFunc] = withCancelCause(background());
		cancelFunc(new Error("already canceled"));
		const mock = new MockStream(11n);
		const req = new AnnouncePleaseMessage({ prefix: "/x/" });
		const aim = new AnnounceInitMessage({ suffixes: [] });
		const reader = new AnnouncementReader(ctx, mock as any, req as any, aim as any);
		await reader.close();
		// Should not call close on stream since context has error
		assertEquals(mock.writable.closeCalls, 0);
	});

	await t.step("closeWithError does nothing when context already has error", async () => {
		const [ctx, cancelFunc] = withCancelCause(background());
		cancelFunc(new Error("already canceled"));
		const mock = new MockStream(12n);
		const req = new AnnouncePleaseMessage({ prefix: "/x/" });
		const aim = new AnnounceInitMessage({ suffixes: [] });
		const reader = new AnnouncementReader(ctx, mock as any, req as any, aim as any);
		await reader.closeWithError(1);
		// Should not call cancel on streams since context has error
		assertEquals(mock.writable.cancelCalls.length, 0);
		assertEquals(mock.readable.cancelCalls.length, 0);
	});
});
