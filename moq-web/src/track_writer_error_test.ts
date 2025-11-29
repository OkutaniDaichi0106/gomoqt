import { assertEquals, assertInstanceOf } from "@std/assert";
import { TrackWriter } from "./track.ts";
import { SendStream } from "./internal/webtransport/mod.ts";
import { GroupSequenceFirst } from "./group_stream.ts";

// GroupMessage not used in this test file
Deno.test("TrackWriter.openGroup returns error when subscribeStream.writeInfo fails", async () => {
	const subscribeStream = {
		writeInfo: async () => new Error("writeInfo failed"),
		context: { done: () => new Promise(() => {}) },
		subscribeId: 0,
		trackConfig: {},
	} as any;

	const tw = new TrackWriter("/test/" as any, "name" as any, subscribeStream, async () => {
		return [
			new SendStream({ stream: new WritableStream({ write(_c) {} }), streamId: 1n }),
			undefined,
		];
	});

	const [group, err] = await tw.openGroup(GroupSequenceFirst);
	assertEquals(group, undefined);
	assertInstanceOf(err, Error);
});

Deno.test("TrackWriter.openGroup returns error when writeVarint fails", async () => {
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
