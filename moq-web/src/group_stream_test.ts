import { assertEquals, assertInstanceOf } from "@std/assert";
import { MockReceiveStream, MockSendStream } from "./internal/webtransport/mock_stream_test.ts";
import { GroupReader, GroupSequenceFirst, GroupWriter } from "./group_stream.ts";
import { GroupMessage, writeVarint } from "./internal/message/mod.ts";
import { BytesFrame, Frame } from "./frame.ts";
import { background, withCancelCause } from "@okudai/golikejs/context";
import { GroupErrorCode } from "./error.ts";
import { SendStream } from "./internal/webtransport/mod.ts";
import { ReceiveStream } from "./internal/webtransport/mod.ts";

Deno.test("GroupWriter", async (t) => {
	await t.step("writeFrame writes correct bytes and returns undefined", async () => {
		const [ctx] = withCancelCause(background());
		const writer = new MockSendStream(1n);
		const msg = new GroupMessage({ sequence: 1, subscribeId: 0 });
		const gw = new GroupWriter(ctx, writer as any, msg);
		const frame = new Frame(new Uint8Array([1, 2, 3]));
		const err = await gw.writeFrame(frame);
		assertEquals(err, undefined);
		// There should be two writes: varint length + data
		assertEquals(writer.writtenData.length, 2);
		// The payload should equal the frame data
		assertEquals(
			writer.getAllWrittenData().subarray(writer.getAllWrittenData().length - 3),
			new Uint8Array([1, 2, 3]),
		);
	});

	await t.step("writeFrame returns an error if write fails", async () => {
		const [ctx] = withCancelCause(background());
		const writer = new MockSendStream(1n);
		writer.writeError = new Error("fail");
		const msg = new GroupMessage({ sequence: 1, subscribeId: 0 });
		const gw = new GroupWriter(ctx, writer as any, msg);
		const frame = new Frame(new Uint8Array([1]));
		const err = await gw.writeFrame(frame);
		assertEquals(err instanceof Error, true);
	});

	await t.step(
		"close increments close calls and cancel does not panic when already cancelled",
		async () => {
			const [ctx] = withCancelCause(background());
			const writer = new MockSendStream(2n);
			const msg = new GroupMessage({ sequence: 1, subscribeId: 0 });
			const gw = new GroupWriter(ctx, writer as any, msg);
			await gw.close();
			assertEquals(writer.closeCalls, 1);
			// Cancel again (should do nothing if already closed)
			await gw.cancel(GroupErrorCode.PublishAborted);
			// Canceling after close should not break (cancel calls patched in stream)
		},
	);

	await t.step("cancel doesn't panic when already cancelled", async () => {
		let canceled = false;
		const writer = new SendStream({
			stream: new WritableStream({
				write(_c) {},
				abort(_e) {
					canceled = true;
					return Promise.resolve();
				},
			}),
			streamId: 1n,
		});
		const groupMsg = new GroupMessage({ sequence: GroupSequenceFirst });
		const gw = new GroupWriter(background(), writer, groupMsg);
		await gw.cancel(GroupErrorCode.SubscribeCanceled);
		// second cancel should not throw
		await gw.cancel(GroupErrorCode.SubscribeCanceled);
		assertEquals(canceled, true);
	});

	await t.step("close does nothing when context already has error", async () => {
		const [ctx, cancelFunc] = withCancelCause(background());
		cancelFunc(new Error("already canceled"));
		await new Promise((r) => setTimeout(r, 0));
		const writer = new MockSendStream(5n);
		const msg = new GroupMessage({ sequence: 1, subscribeId: 0 });
		const gw = new GroupWriter(ctx, writer as any, msg);
		await gw.close();
		// Should not call close on stream since context has error
		assertEquals(writer.closeCalls, 0);
	});

	await t.step("cancel does nothing when context already has error", async () => {
		const [ctx, cancelFunc] = withCancelCause(background());
		cancelFunc(new Error("already canceled"));
		await new Promise((r) => setTimeout(r, 0));
		const writer = new MockSendStream(6n);
		const msg = new GroupMessage({ sequence: 1, subscribeId: 0 });
		const gw = new GroupWriter(ctx, writer as any, msg);
		await gw.cancel(GroupErrorCode.SubscribeCanceled);
		// Should not call cancel on stream since context has error
		assertEquals(writer.cancelCalls.length, 0);
	});

	Deno.test("GroupReader", async (t) => {
		await t.step("readFrame reads data without growing buffer when sufficient", async () => {
			const [ctx] = withCancelCause(background());
			const payload = new Uint8Array([10, 20, 30]);
			// Prepare Readable data: varint length then payload
			const ms = new MockSendStream(8n);
			await writeVarint(ms, payload.length);
			await ms.write(payload);
			const data = ms.getAllWrittenData();
			const rs = new MockReceiveStream(8n, data);
			const msg = new GroupMessage({ sequence: 1, subscribeId: 0 });
			const gr = new GroupReader(ctx, rs as any, msg);
			const frame = new Frame(new Uint8Array(10)); // large enough buffer
			const err = await gr.readFrame(frame);
			assertEquals(err, undefined);
			// Now frame.data should contain payload at start
			const readSub = frame.data.subarray(0, payload.length);
			assertEquals(readSub, payload);
			// Buffer should not have been reallocated (same reference)
			assertEquals(frame.data.length, 10);
		});

		await t.step("cancel cancels underlying stream", async () => {
			const [ctx] = withCancelCause(background());
			const rs = new MockReceiveStream(4n, new Uint8Array([]));
			const msg = new GroupMessage({ sequence: 1, subscribeId: 0 });
			const gr = new GroupReader(ctx, rs as any, msg);
			await gr.cancel(GroupErrorCode.ExpiredGroup);
			assertEquals(rs.cancelCalls.length, 1);
		});

		await t.step("cancel does nothing when context already has error", async () => {
			const [ctx, cancelFunc] = withCancelCause(background());
			cancelFunc(new Error("already canceled"));
			const rs = new MockReceiveStream(7n, new Uint8Array([]));
			const msg = new GroupMessage({ sequence: 1, subscribeId: 0 });
			const gr = new GroupReader(ctx, rs as any, msg);
			await gr.cancel(GroupErrorCode.ExpiredGroup);
			// Should not call cancel on stream since context has error
			assertEquals(rs.cancelCalls.length, 0);
		});

		await t.step("readFrame returns error when varint too large", async () => {
			// Construct a readable that returns a very large varint (more than Number.MAX_SAFE_INTEGER)
			// The readVarint reads varint encoded in bytes; to simulate a too large varint,
			// we produce 9 bytes of varint with high bits, value bigger than Number.MAX_SAFE_INTEGER
			const bytes = new Uint8Array([0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01]);
			const readable = new ReadableStream<Uint8Array>({
				start(c) {
					c.enqueue(bytes);
					c.close();
				},
			});

			const reader = new ReceiveStream({ stream: readable, streamId: 1n });
			const gr = new GroupReader(
				background(),
				reader,
				new GroupMessage({ sequence: GroupSequenceFirst }),
			);

			const fr = new BytesFrame(new Uint8Array(1));
			const errRes = await gr.readFrame(fr);
			assertInstanceOf(errRes, Error);
		});

		await t.step(
			"readFrame returns error when readFull returns EOFError due to insufficient data",
			async () => {
				// Prepare a readable where the varint len is 4 but only 2 bytes provided
				const lenBuf = new Uint8Array([0x04]); // varint 4 (single byte, fine)
				const dataBuf = new Uint8Array([1, 2]);
				const total = new Uint8Array([...lenBuf, ...dataBuf]);
				const readable = new ReadableStream<Uint8Array>({
					start(c) {
						c.enqueue(total);
						c.close();
					},
				});

				const reader = new ReceiveStream({ stream: readable, streamId: 1n });
				const gr = new GroupReader(
					background(),
					reader,
					new GroupMessage({ sequence: 1, subscribeId: 0 }),
				);

				const fr = {
					data: new Uint8Array(8),
					byteLength: 8,
					copyTo: (dest: Uint8Array) => dest.set(new Uint8Array(0)),
				} as any;
				const err = await gr.readFrame(fr as any);
				assertInstanceOf(err, Error);
			},
		);
	});
});
