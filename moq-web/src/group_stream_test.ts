import { assertEquals, assertExists, assertInstanceOf } from "@std/assert";
import { stub } from "@std/testing/mock";
import { GroupReader, GroupWriter } from "./group_stream.ts";
import type { Reader, Writer } from "./internal/webtransport/mod.ts";
import { StreamError } from "./internal/webtransport/error.ts";
import type { GroupMessage } from "./internal/message/mod.ts";
import { BytesFrame } from "./frame.ts";
import { background as createBackground } from "@okudai/golikejs/context";

// Create a fresh background context for each test
function createTestContext() {
	return createBackground();
}

// Simple helper to create mock Writer
function createMockWriter() {
	const calls: { flush: any[] } = { flush: [] };
	const obj = {
		writeUint8Array: async () => undefined,
		copyFrom: async () => undefined,
		flush: async (): Promise<Error | undefined> => {
			calls.flush.push({});
			return undefined;
		},
		close: async () => undefined,
		cancel: async () => undefined,
		closed: async () => undefined,
		calls, // expose calls for assertions
	};
	return obj as any as Writer & { calls: any };
}

// Simple helper to create mock Reader
function createMockReader() {
	const obj = {
		readUint8Array: async () => [new Uint8Array(), undefined] as const,
		readVarint: async (): Promise<[number, Error | undefined]> => [0, undefined],
		fillN: async (): Promise<Error | undefined> => undefined,
		cancel: async () => undefined,
		closed: async () => undefined,
	};
	return obj as any as Reader;
}

Deno.test("GroupWriter - constructor and initialization", () => {
	const ctx = createTestContext();
	const mockWriter = createMockWriter();
	const mockGroup = { sequence: 123n } as unknown as GroupMessage;

	const gw = new GroupWriter(ctx, mockWriter, mockGroup);

	assertInstanceOf(gw, GroupWriter);
	assertEquals(gw.sequence, 123n);
	assertExists(gw.context);
});

Deno.test("GroupWriter - writeFrame success", async () => {
	const ctx = createTestContext();
	const mockWriter = createMockWriter();
	const mockGroup = { sequence: 123n } as unknown as GroupMessage;
	const gw = new GroupWriter(ctx, mockWriter, mockGroup);

	const data = new Uint8Array([1, 2, 3, 4]);
	const frame = new BytesFrame(data);

	const err = await gw.writeFrame(frame as any);

	assertEquals(err, undefined);
	assertEquals(mockWriter.calls.flush.length > 0, true);
});

Deno.test("GroupWriter - writeFrame with flush error", async () => {
	const ctx = createTestContext();
	const mockWriter = createMockWriter();
	const mockGroup = { sequence: 123n } as unknown as GroupMessage;
	const gw = new GroupWriter(ctx, mockWriter, mockGroup);

	const flushError = new Error("Flush failed");
	const flushStub = stub(mockWriter, "flush", async () => flushError);

	try {
		const data = new Uint8Array([1, 2, 3, 4]);
		const frame = new BytesFrame(data);
		const err = await gw.writeFrame(frame as any);

		assertEquals(err, flushError);
	} finally {
		flushStub.restore();
	}
});

Deno.test("GroupWriter - close behavior", async () => {
	const ctx = createTestContext();
	const mockWriter = createMockWriter();
	const mockGroup = { sequence: 123n } as unknown as GroupMessage;
	const gw = new GroupWriter(ctx, mockWriter, mockGroup);

	await gw.close();
	// Verify context is set up
	assertExists(gw.context);
});

Deno.test("GroupWriter - cancel with StreamError", async () => {
	const ctx = createTestContext();
	const mockWriter = createMockWriter();
	const mockGroup = { sequence: 123n } as unknown as GroupMessage;
	const gw = new GroupWriter(ctx, mockWriter, mockGroup);

	let cancelCalled = false;
	let cancelError: any;
	const cancelStub = stub(mockWriter, "cancel", async (error: any) => {
		cancelCalled = true;
		cancelError = error;
	});

	try {
		await gw.cancel(404, "Not found");

		assertEquals(cancelCalled, true);
		assertInstanceOf(cancelError, StreamError);
	} finally {
		cancelStub.restore();
	}
});

Deno.test("GroupReader - constructor and sequence", () => {
	const ctx = createTestContext();
	const mockReader = createMockReader();
	const mockGroup = { sequence: 456n } as unknown as GroupMessage;

	const gr = new GroupReader(ctx, mockReader, mockGroup);

	assertInstanceOf(gr, GroupReader);
	assertEquals(gr.sequence, 456n);
	assertExists(gr.context);
});

Deno.test("GroupReader - readFrame success", async () => {
	const ctx = createTestContext();
	const mockReader = createMockReader();
	const mockGroup = { sequence: 456n } as unknown as GroupMessage;
	const gr = new GroupReader(ctx, mockReader, mockGroup);

	const expectedData = new Uint8Array([1, 2, 3, 4]);
	const readVarintStub = stub(
		mockReader,
		"readVarint",
		async () => [expectedData.byteLength, undefined] as [number, Error | undefined],
	);
	const fillNStub = stub(mockReader, "fillN", async (buf: Uint8Array, len: number) => {
		buf.set(expectedData.subarray(0, len));
		return undefined;
	});

	try {
		const frame = { data: new Uint8Array() } as any;
		const err = await gr.readFrame(frame);

		assertEquals(frame.data.slice(0, expectedData.byteLength), expectedData);
		assertEquals(err, undefined);
	} finally {
		readVarintStub.restore();
		fillNStub.restore();
	}
});

Deno.test("GroupReader - readFrame read error", async () => {
	const ctx = createTestContext();
	const mockReader = createMockReader();
	const mockGroup = { sequence: 456n } as unknown as GroupMessage;
	const gr = new GroupReader(ctx, mockReader, mockGroup);

	const readErr = new Error("Read failed");
	const readVarintStub = stub(
		mockReader,
		"readVarint",
		async () => [0, readErr] as [number, Error | undefined],
	);

	try {
		const frame = { data: new Uint8Array() } as any;
		const err = await gr.readFrame(frame);

		assertEquals(err, readErr);
	} finally {
		readVarintStub.restore();
	}
});

Deno.test("GroupReader - readFrame fillN error", async () => {
	const ctx = createTestContext();
	const mockReader = createMockReader();
	const mockGroup = { sequence: 456n } as unknown as GroupMessage;
	const gr = new GroupReader(ctx, mockReader, mockGroup);

	const fillErr = new Error("Fill failed");
	const readVarintStub = stub(
		mockReader,
		"readVarint",
		async () => [10, undefined] as [number, Error | undefined],
	);
	const fillNStub = stub(mockReader, "fillN", async () => fillErr);

	try {
		const frame = { data: new Uint8Array() } as any;
		const err = await gr.readFrame(frame);

		assertEquals(err, fillErr);
	} finally {
		readVarintStub.restore();
		fillNStub.restore();
	}
});

Deno.test("GroupReader - varint too large", async () => {
	const ctx = createTestContext();
	const mockReader = createMockReader();
	const mockGroup = { sequence: 456n } as unknown as GroupMessage;
	const gr = new GroupReader(ctx, mockReader, mockGroup);

	const readVarintStub = stub(
		mockReader,
		"readVarint",
		async () => [Number.MAX_SAFE_INTEGER + 1, undefined] as [number, Error | undefined],
	);

	try {
		const frame = { data: new Uint8Array() } as any;
		const err = await gr.readFrame(frame);

		assertInstanceOf(err as Error, Error);
	} finally {
		readVarintStub.restore();
	}
});

Deno.test("GroupReader - reuse buffer on multiple reads", async () => {
	const ctx = createTestContext();
	const mockReader = createMockReader();
	const mockGroup = { sequence: 456n } as unknown as GroupMessage;
	const gr = new GroupReader(ctx, mockReader, mockGroup);

	const data1 = new Uint8Array([1, 2, 3]);
	const data2 = new Uint8Array([4, 5, 6, 7, 8]);

	let readCallCount = 0;
	const readVarintStub = stub(mockReader, "readVarint", async () => {
		readCallCount++;
		if (readCallCount === 1) {
			return [data1.byteLength, undefined] as [number, Error | undefined];
		}
		return [data2.byteLength, undefined] as [number, Error | undefined];
	});

	let fillCallCount = 0;
	const fillNStub = stub(mockReader, "fillN", async (buf: Uint8Array, len: number) => {
		fillCallCount++;
		if (fillCallCount === 1) {
			buf.set(data1.subarray(0, len));
		} else {
			buf.set(data2.subarray(0, len));
		}
		return undefined;
	});

	try {
		const frame = { data: new Uint8Array() } as any;

		const err1 = await gr.readFrame(frame);
		assertEquals(err1, undefined);
		assertEquals(frame.data.slice(0, data1.byteLength), data1);

		// Reset frame buffer for second read
		frame.data = new Uint8Array();
		const err2 = await gr.readFrame(frame);
		assertEquals(err2, undefined);
		assertEquals(frame.data.slice(0, data2.byteLength), data2);
	} finally {
		readVarintStub.restore();
		fillNStub.restore();
	}
});

Deno.test("GroupReader - cancel with StreamError", async () => {
	const ctx = createTestContext();
	const mockReader = createMockReader();
	const mockGroup = { sequence: 456n } as unknown as GroupMessage;
	const gr = new GroupReader(ctx, mockReader, mockGroup);

	let cancelCalled = false;
	let cancelError: any;
	const cancelStub = stub(mockReader, "cancel", async (error: any) => {
		cancelCalled = true;
		cancelError = error;
	});

	try {
		await gr.cancel(404, "Not found");

		assertEquals(cancelCalled, true);
		assertInstanceOf(cancelError, StreamError);
	} finally {
		cancelStub.restore();
	}
});
