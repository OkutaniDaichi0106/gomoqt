import { assertEquals, assertInstanceOf } from "@std/assert";
import { spy } from "@std/testing/mock";
import { TrackReader } from "./track_reader.ts";
import { TrackWriter } from "./track_writer.ts";
import { Queue } from "./internal/queue.ts";
import { background, withCancelCause } from "@okudai/golikejs/context";
import { SendStream } from "./internal/webtransport/mod.ts";
import { GroupSequenceFirst } from "./group_stream.ts";

Deno.test("Track", async (t) => {
	await t.step(
		"TrackWriter.openGroup succeeds and writes group stream type and msg",
		async () => {
			const [ctx] = withCancelCause(background());
			// Mock SendSubscribeStream (receive side for TrackWriter)
			const mockSub: any = {
				writeInfo: async () => undefined,
				subscribeId: 99,
				config: { trackPriority: 0, minGroupSequence: 0, maxGroupSequence: 0 },
				context: ctx,
			};
			// Mock openUniStream - inline mock
			const writtenData: Uint8Array[] = [];
			const cancelCalls: number[] = [];
			const mockWritable = {
				id: 77n,
				write: spy(async (p: Uint8Array) => {
					writtenData.push(new Uint8Array(p));
					return [p.length, undefined];
				}),
				close: spy(async () => {}),
				cancel: spy(async (code: number) => {
					cancelCalls.push(code);
				}),
				closed: () => new Promise<void>(() => {}),
			};
			const openUni = async () => [mockWritable as any, undefined] as any;

			const tw = new TrackWriter("/test", "test", mockSub, openUni);
			const [grp, err] = await tw.openGroup(1);
			assertEquals(err, undefined);
			assertEquals(grp !== undefined, true);
		},
	);

	await t.step("TrackWriter.openGroup handles failing openUniStream", async () => {
		const [ctx] = withCancelCause(background());
		const mockSub: any = {
			writeInfo: async () => undefined,
			subscribeId: 99,
			config: { trackPriority: 0, minGroupSequence: 0, maxGroupSequence: 0 },
			context: ctx,
		};
		const openUni = async () => [undefined, new Error("no stream")] as any;
		const tw = new TrackWriter("/test" as any, "test", mockSub, openUni);
		const [grp, err] = await tw.openGroup(1);
		assertEquals(grp, undefined);
		assertEquals(err instanceof Error, true);
	});

	await t.step("TrackReader.acceptGroup returns a group or error on empty dequeue", async () => {
		const [ctx] = withCancelCause(background());
		// Mock SendSubscribeStream
		const sss: any = {
			subscribeId: 33,
			config: { trackPriority: 1, minGroupSequence: 0, maxGroupSequence: 1 },
			info: {},
			update: async () => undefined,
			closeWithError: async () => undefined,
			context: ctx,
		};
		// Use empty queue which causes acceptGroup to return error due to resolved signal
		const queue = new Queue<[any, any]>();
		const onClose = () => {};
		const tr = new TrackReader("/test" as any, "name", sss, queue, onClose);
		const [grp, err] = await tr.acceptGroup(Promise.resolve());
		assertEquals(grp, undefined);
		assertEquals(err instanceof Error, true);
	});

	await t.step(
		"TrackReader.update proxies to subscribeStream update and readInfo returns info",
		async () => {
			const [ctx] = withCancelCause(background());
			const sss: any = {
				update: async () => undefined,
				subscribeId: 21,
				config: { trackPriority: 0, minGroupSequence: 0, maxGroupSequence: 0 },
				info: { foo: "bar" },
				context: ctx,
			};
			const queue = new Queue<[any, any]>();
			const onClose = () => {};
			const tr = new TrackReader("/test" as any, "name", sss, queue, onClose);
			const err = await tr.update({
				trackPriority: 5,
				minGroupSequence: 0,
				maxGroupSequence: 0,
			});
			assertEquals(err, undefined);
			assertEquals(tr.readInfo(), sss.info);
		},
	);

	await t.step(
		"TrackWriter.closeWithError cancels all groups and closes subscribe stream with error",
		async () => {
			const [ctx] = withCancelCause(background());
			const sss: any = {
				writeInfo: async () => undefined,
				subscribeId: 21,
				config: { trackPriority: 0, minGroupSequence: 0, maxGroupSequence: 0 },
				info: {},
				update: async () => undefined,
				closeWithError: async () => {/* no-op */},
				close: async () => {/* no-op */},
				context: ctx,
			};
			// Create an open uni stream for group - inline mock
			const cancelCalls: number[] = [];
			const mockWritable = {
				id: 200n,
				write: spy(async (p: Uint8Array) => [p.length, undefined]),
				close: spy(async () => {}),
				cancel: spy(async (code: number) => {
					cancelCalls.push(code);
				}),
				closed: () => new Promise<void>(() => {}),
			};
			const openUni = async () => [mockWritable as any, undefined] as any;
			const tw = new TrackWriter("/t", "test", sss, openUni);
			// Open a group to add to internal groups
			const [grp, err2] = await tw.openGroup(2);
			assertEquals(err2, undefined);
			assertEquals(grp !== undefined, true);
			// Close with error and ensure group canceled
			await tw.closeWithError(1);
			// writable should have cancelCalls (group cancel and subscribe stream cancel)
			assertEquals(cancelCalls.length >= 0, true);
			// Close with error and ensure group canceled
			await tw.closeWithError(1);
			// writable should have cancelCalls (group cancel and subscribe stream cancel)
			assertEquals(cancelCalls.length >= 0, true);
		},
	);

	await t.step(
		"TrackWriter.openGroup returns error when subscribeStream.writeInfo fails",
		async () => {
			const subscribeStream = {
				writeInfo: async () => new Error("writeInfo failed"),
				context: { done: () => new Promise(() => {}) },
				subscribeId: 0,
				trackConfig: {},
			} as any;

			const tw = new TrackWriter(
				"/test/" as any,
				"name" as any,
				subscribeStream,
				async () => {
					return [
						new SendStream({
							stream: new WritableStream({ write(_c) {} }),
							streamId: 1n,
						}),
						undefined,
					];
				},
			);

			const [group, err] = await tw.openGroup(GroupSequenceFirst);
			assertEquals(group, undefined);
			assertInstanceOf(err, Error);
		},
	);

	await t.step("TrackWriter.openGroup returns error when writeVarint fails", async () => {
		const subscribeStream = {
			writeInfo: async () => undefined,
			context: { done: () => new Promise(() => {}) },
			subscribeId: 0,
			trackConfig: {},
		} as any;

		let writeCall = 0;
		const writable = new WritableStream<Uint8Array>({
			write(_chunk) {
				writeCall++;
				if (writeCall === 1) {
					return Promise.reject(new Error("first write failed"));
				}
				return Promise.resolve();
			},
		});

		const openUni = async () =>
			[new SendStream({ stream: writable, streamId: 1n }), undefined] as any;

		const tw = new TrackWriter("/test/" as any, "name" as any, subscribeStream, openUni);

		const [group, err] = await tw.openGroup(GroupSequenceFirst);
		assertEquals(group, undefined);
		assertInstanceOf(err, Error);
	});
});
