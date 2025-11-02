import { assertEquals, assertExists } from "@std/assert";
import { TrackReader, TrackWriter } from "./track.ts";
import {
	MockReceiveStream,
	MockSendStream,
} from "./internal/webtransport/mock_stream_test.ts";
import type { ReceiveSubscribeStream, SendSubscribeStream } from "./subscribe_stream.ts";
import { background, withCancel } from "@okudai/golikejs/context";
import type { Context } from "@okudai/golikejs/context";
import { GroupMessage } from "./internal/message/mod.ts";
import type { SendStream } from "./internal/webtransport/mod.ts";
import type { Info } from "./info.ts";
import { PublishAbortedErrorCode } from "./error.ts";
import { UniStreamTypes } from "./stream_type.ts";

// Mock ReceiveSubscribeStream
class MockReceiveSubscribeStream {
	public subscribeId: number;
	public writeInfoCalls: Info[] = [];
	public closeCalls = 0;
	public closeWithErrorCalls: Array<{ code: number; message: string }> = [];
	public ctx: Context;
	public trackConfig = {
		trackPriority: 0,
		minGroupSequence: 0n,
		maxGroupSequence: 100n,
	};

	constructor(ctx: Context, subscribeId: number) {
		this.ctx = ctx;
		this.subscribeId = subscribeId;
	}

	get context(): Context {
		return this.ctx;
	}

	async writeInfo(info?: Info): Promise<Error | undefined> {
		this.writeInfoCalls.push(info ?? {});
		return undefined;
	}

	async close(): Promise<void> {
		this.closeCalls++;
	}

	async closeWithError(code: number, message: string): Promise<void> {
		this.closeWithErrorCalls.push({ code, message });
	}

	reset(): void {
		this.writeInfoCalls = [];
		this.closeCalls = 0;
		this.closeWithErrorCalls = [];
	}
}

// Mock SendSubscribeStream
class MockSendSubscribeStream {
	public subscribeId: number;
	public updateCalls: any[] = [];
	public closeWithErrorCalls: Array<{ code: number; message: string }> = [];
	public ctx: Context;
	public _info: Info = {};
	public _config = {
		trackPriority: 0,
		minGroupSequence: 0n,
		maxGroupSequence: 100n,
	};

	constructor(ctx: Context, subscribeId: number) {
		this.ctx = ctx;
		this.subscribeId = subscribeId;
	}

	get context(): Context {
		return this.ctx;
	}

	get info(): Info {
		return this._info;
	}

	get config() {
		return this._config;
	}

	async update(config: any): Promise<Error | undefined> {
		this.updateCalls.push(config);
		this._config = config;
		return undefined;
	}

	async closeWithError(code: number, message: string): Promise<void> {
		this.closeWithErrorCalls.push({ code, message });
	}

	reset(): void {
		this.updateCalls = [];
		this.closeWithErrorCalls = [];
	}
}

Deno.test("TrackWriter - Basic Operations", async (t) => {
	const [ctx, cancel] = withCancel(background());

	await t.step("should create TrackWriter with correct properties", () => {
		const mockStream = new MockReceiveSubscribeStream(ctx, 42);
		const openUniStream = async (): Promise<[SendStream, undefined] | [undefined, Error]> => {
			return [new MockSendStream() as unknown as SendStream, undefined];
		};

		const writer = new TrackWriter(
			"/test/path",
			"video",
			mockStream as unknown as ReceiveSubscribeStream,
			openUniStream,
		);

		assertEquals(writer.broadcastPath, "/test/path");
		assertEquals(writer.trackName, "video");
		assertEquals(writer.subscribeId, 42);
		assertEquals(writer.context, ctx);
		assertExists(writer.config);
	});

	await t.step("should open group successfully", async () => {
		const mockStream = new MockReceiveSubscribeStream(ctx, 42);
		const mockWriter = new MockSendStream();
		const openUniStream = async (): Promise<[SendStream, undefined] | [undefined, Error]> => {
			return [mockWriter as unknown as SendStream, undefined];
		};

		const writer = new TrackWriter(
			"/test/path",
			"video",
			mockStream as unknown as ReceiveSubscribeStream,
			openUniStream,
		);

		const [group, err] = await writer.openGroup(100n);

		assertEquals(err, undefined);
		assertExists(group);
		assertEquals(group!.sequence, 100n);
		assertEquals(mockStream.writeInfoCalls.length, 1);
		assertEquals(mockWriter.writeUint8.calls.length, 1);
		if (mockWriter.writeUint8.calls[0]) {
			assertEquals(mockWriter.writeUint8.calls[0].args[0], UniStreamTypes.GroupStreamType);
		}
	});

	await t.step("should handle writeInfo error in openGroup", async () => {
		const mockStream = new MockReceiveSubscribeStream(ctx, 42);
		mockStream.writeInfo = async () => new Error("write failed");

		const writer = new TrackWriter(
			"/test/path",
			"video",
			mockStream as unknown as ReceiveSubscribeStream,
			async () => [new MockSendStream() as unknown as SendStream, undefined],
		);

		const [group, err] = await writer.openGroup(100n);

		assertEquals(group, undefined);
		assertExists(err);
		assertEquals(err!.message, "write failed");
	});

	await t.step("should handle openUniStream error", async () => {
		const mockStream = new MockReceiveSubscribeStream(ctx, 42);
		const openUniStream = async (): Promise<[SendStream, undefined] | [undefined, Error]> => {
			return [undefined, new Error("stream open failed")];
		};

		const writer = new TrackWriter(
			"/test/path",
			"video",
			mockStream as unknown as ReceiveSubscribeStream,
			openUniStream,
		);

		const [group, err] = await writer.openGroup(100n);

		assertEquals(group, undefined);
		assertExists(err);
		assertEquals(err!.message, "stream open failed");
	});

	await t.step("should write info successfully", async () => {
		const mockStream = new MockReceiveSubscribeStream(ctx, 42);

		const writer = new TrackWriter(
			"/test/path",
			"video",
			mockStream as unknown as ReceiveSubscribeStream,
			async () => [new MockSendStream() as unknown as SendStream, undefined],
		);

		const info = {};
		const err = await writer.writeInfo(info);

		assertEquals(err, undefined);
		assertEquals(mockStream.writeInfoCalls.length, 1);
		assertEquals(mockStream.writeInfoCalls[0], info);
	});

	await t.step("should close successfully", async () => {
		const mockStream = new MockReceiveSubscribeStream(ctx, 42);

		const writer = new TrackWriter(
			"/test/path",
			"video",
			mockStream as unknown as ReceiveSubscribeStream,
			async () => [new MockSendStream() as unknown as SendStream, undefined],
		);

		await writer.close();

		assertEquals(mockStream.closeCalls, 1);
	});

	await t.step("should closeWithError successfully", async () => {
		const mockStream = new MockReceiveSubscribeStream(ctx, 42);

		const writer = new TrackWriter(
			"/test/path",
			"video",
			mockStream as unknown as ReceiveSubscribeStream,
			async () => [new MockSendStream() as unknown as SendStream, undefined],
		);

		const errorCode = 0x00; // InternalSubscribeErrorCode
		await writer.closeWithError(errorCode, "test error");

		assertEquals(mockStream.closeWithErrorCalls.length, 1);
		if (mockStream.closeWithErrorCalls[0]) {
			assertEquals(mockStream.closeWithErrorCalls[0].code, errorCode);
			assertEquals(mockStream.closeWithErrorCalls[0].message, "test error");
		}
	});

	await t.step("should cancel all groups when closing with error", async () => {
		const mockStream = new MockReceiveSubscribeStream(ctx, 42);
		const mockWriters: MockSendStream[] = [];

		const openUniStream = async (): Promise<[SendStream, undefined] | [undefined, Error]> => {
			const writer = new MockSendStream();
			mockWriters.push(writer);
			return [writer as unknown as SendStream, undefined];
		};

		const writer = new TrackWriter(
			"/test/path",
			"video",
			mockStream as unknown as ReceiveSubscribeStream,
			openUniStream,
		);

		// Open multiple groups
		await writer.openGroup(1n);
		await writer.openGroup(2n);
		await writer.openGroup(3n);

		const errorCode = 0x00; // InternalSubscribeErrorCode
		await writer.closeWithError(errorCode, "test error");

		// Verify all groups were cancelled
		for (const mockWriter of mockWriters) {
			assertEquals(mockWriter.cancelCalls.length, 1);
			if (mockWriter.cancelCalls[0]) {
				assertEquals(mockWriter.cancelCalls[0].code, PublishAbortedErrorCode);
			}
		}

		// Verify subscribe stream was closed with error
		assertEquals(mockStream.closeWithErrorCalls.length, 1);
	});

	cancel();
});

Deno.test("TrackReader - Basic Operations", async (t) => {
	const [ctx, cancel] = withCancel(background());

	await t.step("should create TrackReader with correct properties", () => {
		const mockStream = new MockSendSubscribeStream(ctx, 42);
		const acceptFunc = async (_: Promise<void>) => {
			return undefined;
		};
		const onCloseFunc = () => {};

		const reader = new TrackReader(
			mockStream as unknown as SendSubscribeStream,
			acceptFunc,
			onCloseFunc,
		);

		assertEquals(reader.context, ctx);
		assertEquals(reader.trackConfig, mockStream.config);
	});

	await t.step("should accept group successfully", async () => {
		const mockStream = new MockSendSubscribeStream(ctx, 42);
		const mockReader = new MockReceiveStream();
		const groupMsg = new GroupMessage({ subscribeId: 42, sequence: 100n });

		const acceptFunc = async (_: Promise<void>) => {
			return [mockReader, groupMsg] as [any, GroupMessage];
		};

		const reader = new TrackReader(
			mockStream as unknown as SendSubscribeStream,
			acceptFunc,
			() => {},
		);

		const [group, err] = await reader.acceptGroup(Promise.resolve());

		assertEquals(err, undefined);
		assertExists(group);
		assertEquals(group!.sequence, 100n);
	});

	await t.step("should handle acceptGroup error when dequeue fails", async () => {
		const mockStream = new MockSendSubscribeStream(ctx, 42);
		const acceptFunc = async (_: Promise<void>) => {
			return undefined;
		};

		const reader = new TrackReader(
			mockStream as unknown as SendSubscribeStream,
			acceptFunc,
			() => {},
		);

		const [group, err] = await reader.acceptGroup(Promise.resolve());

		assertEquals(group, undefined);
		assertExists(err);
		assertEquals(err!.message, "[TrackReader] failed to dequeue group message");
	});

	await t.step("should handle acceptGroup error when context is cancelled", async () => {
		const [localCtx, localCancel] = withCancel(background());
		const mockStream = new MockSendSubscribeStream(localCtx, 42);
		const acceptFunc = async (_: Promise<void>) => {
			return undefined;
		};

		const reader = new TrackReader(
			mockStream as unknown as SendSubscribeStream,
			acceptFunc,
			() => {},
		);

		// Cancel the context before accepting
		localCancel();
		await new Promise((resolve) => setTimeout(resolve, 10));

		const [group, err] = await reader.acceptGroup(Promise.resolve());

		assertEquals(group, undefined);
		assertExists(err);
	});

	await t.step("should update track config successfully", async () => {
		const mockStream = new MockSendSubscribeStream(ctx, 42);
		const reader = new TrackReader(
			mockStream as unknown as SendSubscribeStream,
			async (_) => undefined,
			() => {},
		);

		const newConfig = {
			trackPriority: 10,
			minGroupSequence: 50n,
			maxGroupSequence: 150n,
		};

		const err = await reader.update(newConfig);

		assertEquals(err, undefined);
		assertEquals(mockStream.updateCalls.length, 1);
		assertEquals(mockStream.updateCalls[0], newConfig);
	});

	await t.step("should read info successfully", () => {
		const mockStream = new MockSendSubscribeStream(ctx, 42);
		mockStream._info = { someField: "value" } as any;

		const reader = new TrackReader(
			mockStream as unknown as SendSubscribeStream,
			async (_) => undefined,
			() => {},
		);

		const info = reader.readInfo();

		assertEquals(info, mockStream._info);
	});

	await t.step("should closeWithError successfully", async () => {
		const mockStream = new MockSendSubscribeStream(ctx, 42);
		let onCloseCalled = false;
		const onCloseFunc = () => {
			onCloseCalled = true;
		};

		const reader = new TrackReader(
			mockStream as unknown as SendSubscribeStream,
			async (_) => undefined,
			onCloseFunc,
		);

		const errorCode = 0x00; // InternalSubscribeErrorCode
		await reader.closeWithError(errorCode, "test error");

		assertEquals(mockStream.closeWithErrorCalls.length, 1);
		if (mockStream.closeWithErrorCalls[0]) {
			assertEquals(mockStream.closeWithErrorCalls[0].code, errorCode);
			assertEquals(mockStream.closeWithErrorCalls[0].message, "test error");
		}
		assertEquals(onCloseCalled, true);
	});

	await t.step("should get trackConfig from subscribeStream", () => {
		const mockStream = new MockSendSubscribeStream(ctx, 42);
		mockStream._config = {
			trackPriority: 20,
			minGroupSequence: 10n,
			maxGroupSequence: 200n,
		};

		const reader = new TrackReader(
			mockStream as unknown as SendSubscribeStream,
			async (_) => undefined,
			() => {},
		);

		assertEquals(reader.trackConfig, mockStream._config);
	});

	cancel();
});

Deno.test("TrackWriter - Multiple Groups Management", async (t) => {
	const [ctx, cancel] = withCancel(background());

	await t.step("should handle multiple groups correctly", async () => {
		const mockStream = new MockReceiveSubscribeStream(ctx, 42);
		const mockWriters: MockSendStream[] = [];

		const openUniStream = async (): Promise<[SendStream, undefined] | [undefined, Error]> => {
			const writer = new MockSendStream();
			mockWriters.push(writer);
			return [writer as unknown as SendStream, undefined];
		};

		const writer = new TrackWriter(
			"/test/path",
			"video",
			mockStream as unknown as ReceiveSubscribeStream,
			openUniStream,
		);

		const sequences = [1n, 2n, 3n, 4n, 5n];
		const groups = [];

		for (const seq of sequences) {
			const [group, err] = await writer.openGroup(seq);
			assertEquals(err, undefined);
			assertExists(group);
			groups.push(group!);
		}

		assertEquals(groups.length, 5);
		assertEquals(mockWriters.length, 5);

		for (let i = 0; i < sequences.length; i++) {
			const group = groups[i];
			if (group) {
				assertEquals(group.sequence, sequences[i]);
			}
		}
	});

	await t.step("should close all groups when writer closes", async () => {
		const mockStream = new MockReceiveSubscribeStream(ctx, 42);
		const mockWriters: MockSendStream[] = [];

		const openUniStream = async (): Promise<[SendStream, undefined] | [undefined, Error]> => {
			const writer = new MockSendStream();
			mockWriters.push(writer);
			return [writer as unknown as SendStream, undefined];
		};

		const writer = new TrackWriter(
			"/test/path",
			"video",
			mockStream as unknown as ReceiveSubscribeStream,
			openUniStream,
		);

		// Open multiple groups
		await writer.openGroup(1n);
		await writer.openGroup(2n);
		await writer.openGroup(3n);

		await writer.close();

		// Verify all groups were closed
		for (const mockWriter of mockWriters) {
			assertEquals(mockWriter.closeCalls, 1);
		}

		// Verify subscribe stream was closed
		assertEquals(mockStream.closeCalls, 1);
	});

	cancel();
});

Deno.test("TrackReader - Context Cancellation", async (t) => {
	await t.step("should respect context cancellation in acceptGroup", async () => {
		const [ctx, cancel] = withCancel(background());
		const mockStream = new MockSendSubscribeStream(ctx, 42);

		const acceptFunc = async (signal: Promise<void>) => {
			// Wait for cancellation
			await signal;
			return undefined;
		};

		const reader = new TrackReader(
			mockStream as unknown as SendSubscribeStream,
			acceptFunc,
			() => {},
		);

		// Cancel context before accept completes
		const acceptPromise = reader.acceptGroup(Promise.resolve());
		cancel();

		const [group, err] = await acceptPromise;

		assertEquals(group, undefined);
		assertExists(err);
	});
});

Deno.test("TrackWriter - Edge Cases", async (t) => {
	const [ctx, cancel] = withCancel(background());

	await t.step("should handle writeInfo errors gracefully", async () => {
		const mockStream = new MockReceiveSubscribeStream(ctx, 42);
		mockStream.writeInfo = async () => new Error("network error");

		const writer = new TrackWriter(
			"/test/path",
			"video",
			mockStream as unknown as ReceiveSubscribeStream,
			async () => [new MockSendStream() as unknown as SendStream, undefined],
		);

		const err = await writer.writeInfo({});

		assertExists(err);
		assertEquals(err!.message, "network error");
	});

	await t.step("should handle multiple writeInfo calls", async () => {
		const mockStream = new MockReceiveSubscribeStream(ctx, 42);

		const writer = new TrackWriter(
			"/test/path",
			"video",
			mockStream as unknown as ReceiveSubscribeStream,
			async () => [new MockSendStream() as unknown as SendStream, undefined],
		);

		const info1 = { field: "value1" } as any;
		const info2 = { field: "value2" } as any;

		await writer.writeInfo(info1);
		await writer.writeInfo(info2);

		assertEquals(mockStream.writeInfoCalls.length, 2);
		assertEquals(mockStream.writeInfoCalls[0], info1);
		assertEquals(mockStream.writeInfoCalls[1], info2);
	});

	await t.step("should handle close after closeWithError", async () => {
		const mockStream = new MockReceiveSubscribeStream(ctx, 42);

		const writer = new TrackWriter(
			"/test/path",
			"video",
			mockStream as unknown as ReceiveSubscribeStream,
			async () => [new MockSendStream() as unknown as SendStream, undefined],
		);

		const errorCode = 0x00; // InternalSubscribeErrorCode
		await writer.closeWithError(errorCode, "error");
		await writer.close();

		// Both close operations should have been called
		assertEquals(mockStream.closeWithErrorCalls.length, 1);
		assertEquals(mockStream.closeCalls, 1);
	});

	cancel();
});
