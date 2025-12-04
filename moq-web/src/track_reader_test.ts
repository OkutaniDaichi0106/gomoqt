import { assertEquals } from "@std/assert";
import { TrackReader } from "./track_reader.ts";
import { Queue } from "./internal/queue.ts";
import { background, withCancelCause } from "@okudai/golikejs/context";
import { MockStream } from "./mock_stream_test.ts";
import { SendSubscribeStream } from "./subscribe_stream.ts";
import { SubscribeMessage, SubscribeOkMessage } from "./internal/message/mod.ts";

Deno.test("TrackReader", async (t) => {
	await t.step(
		"TrackReader.acceptGroup returns a group or error on empty dequeue",
		async () => {
			const [ctx] = withCancelCause(background());
			const stream = new MockStream({ id: 1n });
			const subscribe = new SubscribeMessage({
				subscribeId: 33,
				broadcastPath: "/test",
				trackName: "name",
				trackPriority: 1,
				minGroupSequence: 0,
				maxGroupSequence: 1,
			});
			const ok = new SubscribeOkMessage({});
			const sss = new SendSubscribeStream(ctx, stream, subscribe, ok);
			// Use empty queue which causes acceptGroup to return error due to resolved signal
			const queue = new Queue<[any, any]>();
			const onClose = () => {};
			const tr = new TrackReader("/test", "name", sss, queue, onClose);
			const [grp, err] = await tr.acceptGroup(Promise.resolve());
			assertEquals(grp, undefined);
			assertEquals(err instanceof Error, true);
		},
	);

	await t.step(
		"TrackReader.update proxies to subscribeStream update and readInfo returns info",
		async () => {
			const [ctx] = withCancelCause(background());
			const stream = new MockStream({ id: 2n });
			const subscribe = new SubscribeMessage({
				subscribeId: 21,
				broadcastPath: "/test",
				trackName: "name",
				trackPriority: 0,
				minGroupSequence: 0,
				maxGroupSequence: 0,
			});
			const ok = new SubscribeOkMessage({});
			const sss = new SendSubscribeStream(ctx, stream, subscribe, ok);
			const queue = new Queue<[any, any]>();
			const onClose = () => {};
			const tr = new TrackReader("/test", "name", sss, queue, onClose);
			const err = await tr.update({
				trackPriority: 5,
			});
			assertEquals(err, undefined);
			assertEquals(tr.readInfo(), sss.info);
		},
	);
});
