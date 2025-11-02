import { assertEquals, assertExists, assertInstanceOf } from "@std/assert";
import { GroupReader, GroupWriter } from "./group_stream.ts";
import { StreamError } from "./internal/webtransport/error.ts";
import type { GroupMessage } from "./internal/message/mod.ts";
import { BytesFrame } from "./frame.ts";
import { background as createBackground } from "@okudai/golikejs/context";
import { MockReceiveStream, MockSendStream } from "./internal/webtransport/mock_stream_test.ts";

/**
 * Creates a fresh background context for each test.
 * Ensures isolation between tests.
 */
function createTestContext() {
	return createBackground();
}

Deno.test("GroupWriter - Normal Cases", async (t) => {
	await t.step("should create instance with correct sequence", () => {
		const ctx = createTestContext();
		const mockWriter = new MockSendStream();
		const mockGroup = { sequence: 123n } as unknown as GroupMessage;

		const gw = new GroupWriter(ctx, mockWriter as any, mockGroup);

		assertInstanceOf(gw, GroupWriter);
		assertEquals(gw.sequence, 123n);
		assertExists(gw.context);
	});

	await t.step("should write frame successfully", async () => {
		const ctx = createTestContext();
		const mockWriter = new MockSendStream();
		const mockGroup = { sequence: 123n } as unknown as GroupMessage;
		const gw = new GroupWriter(ctx, mockWriter as any, mockGroup);

		const data = new Uint8Array([1, 2, 3, 4]);
		const frame = new BytesFrame(data);

		const err = await gw.writeFrame(frame as any);

		assertEquals(err, undefined);
		// Verify flush was called after write
		assertEquals(mockWriter.flushCalls > 0, true);
	});

	await t.step("should close stream cleanly", async () => {
		const ctx = createTestContext();
		const mockWriter = new MockSendStream();
		const mockGroup = { sequence: 123n } as unknown as GroupMessage;
		const gw = new GroupWriter(ctx, mockWriter as any, mockGroup);

		await gw.close();

		assertExists(gw.context);
		assertEquals(mockWriter.closeCalls > 0, true);
	});
});

Deno.test("GroupWriter - Error Cases", async (t) => {
	await t.step("should return error when flush fails", async () => {
		const ctx = createTestContext();
		const mockWriter = new MockSendStream();
		const mockGroup = { sequence: 123n } as unknown as GroupMessage;
		const gw = new GroupWriter(ctx, mockWriter as any, mockGroup);

		// Simulate flush error
		const flushError = new Error("Flush failed");
		mockWriter.flushError = flushError;

		const data = new Uint8Array([1, 2, 3, 4]);
		const frame = new BytesFrame(data);
		const err = await gw.writeFrame(frame as any);

		assertEquals(err, flushError);
	});

	await t.step("should cancel with StreamError", async () => {
		const ctx = createTestContext();
		const mockWriter = new MockSendStream();
		const mockGroup = { sequence: 123n } as unknown as GroupMessage;
		const gw = new GroupWriter(ctx, mockWriter as any, mockGroup);

		await gw.cancel(404, "Not found");

		assertEquals(mockWriter.cancelCalls.length > 0, true);
		assertInstanceOf(mockWriter.cancelCalls[0], StreamError);
	});
});

Deno.test("GroupReader - Normal Cases", async (t) => {
	await t.step("should create instance with correct sequence", () => {
		const ctx = createTestContext();
		const mockReader = new MockReceiveStream();
		const mockGroup = { sequence: 456n } as unknown as GroupMessage;

		const gr = new GroupReader(ctx, mockReader as any, mockGroup);

		assertInstanceOf(gr, GroupReader);
		assertEquals(gr.sequence, 456n);
		assertExists(gr.context);
	});

	await t.step("should read frame successfully", async () => {
		const ctx = createTestContext();
		const mockReader = new MockReceiveStream();
		const mockGroup = { sequence: 456n } as unknown as GroupMessage;
		const gr = new GroupReader(ctx, mockReader as any, mockGroup);

		const expectedData = new Uint8Array([1, 2, 3, 4]);

		// Mock readVarint to return data length
		mockReader.readVarintImpl = async () =>
			[expectedData.byteLength, undefined] as [number, Error | undefined];

		// Mock fillN to copy expected data
		mockReader.fillNImpl = async (buf: Uint8Array, len: number) => {
			buf.set(expectedData.subarray(0, len));
			return undefined;
		};

		const frame = { data: new Uint8Array() } as any;
		const err = await gr.readFrame(frame);

		assertEquals(err, undefined);
		assertEquals(frame.data.slice(0, expectedData.byteLength), expectedData);
	});
});

Deno.test("GroupReader - Error Cases", async (t) => {
	// Table-driven tests for multiple error scenarios
	const errorCases = {
		"readVarint error": {
			setup: (mockReader: MockReceiveStream) => {
				const readErr = new Error("Read failed");
				mockReader.readVarintImpl = async () => [0, readErr] as [number, Error | undefined];
			},
			expectedError: "Read failed",
		},
		"fillN error": {
			setup: (mockReader: MockReceiveStream) => {
				const fillErr = new Error("Fill failed");
				mockReader.readVarintImpl = async () =>
					[10, undefined] as [number, Error | undefined];
				mockReader.fillNImpl = async () => fillErr;
			},
			expectedError: "Fill failed",
		},
		"varint too large": {
			setup: (mockReader: MockReceiveStream) => {
				// Varint exceeds MAX_SAFE_INTEGER (boundary case)
				mockReader.readVarintImpl = async () =>
					[Number.MAX_SAFE_INTEGER + 1, undefined] as [number, Error | undefined];
			},
			expectedError: undefined, // Just verify error is returned
		},
	};

	for (const [name, testCase] of Object.entries(errorCases)) {
		await t.step(`should handle ${name}`, async () => {
			const ctx = createTestContext();
			const mockReader = new MockReceiveStream();
			const mockGroup = { sequence: 456n } as unknown as GroupMessage;
			const gr = new GroupReader(ctx, mockReader as any, mockGroup);

			testCase.setup(mockReader);

			const frame = { data: new Uint8Array() } as any;
			const err = await gr.readFrame(frame);

			assertInstanceOf(err, Error);
			if (testCase.expectedError) {
				assertEquals((err as Error).message, testCase.expectedError);
			}
		});
	}
});

Deno.test("GroupReader - Buffer Management", async (t) => {
	await t.step("should reuse buffer on multiple reads", async () => {
		const ctx = createTestContext();
		const mockReader = new MockReceiveStream();
		const mockGroup = { sequence: 456n } as unknown as GroupMessage;
		const gr = new GroupReader(ctx, mockReader as any, mockGroup);

		const data1 = new Uint8Array([1, 2, 3]);
		const data2 = new Uint8Array([4, 5, 6, 7, 8]);

		let readCallCount = 0;
		mockReader.readVarintImpl = async () => {
			readCallCount++;
			if (readCallCount === 1) {
				return [data1.byteLength, undefined] as [number, Error | undefined];
			}
			return [data2.byteLength, undefined] as [number, Error | undefined];
		};

		let fillCallCount = 0;
		mockReader.fillNImpl = async (buf: Uint8Array, len: number) => {
			fillCallCount++;
			if (fillCallCount === 1) {
				buf.set(data1.subarray(0, len));
			} else {
				buf.set(data2.subarray(0, len));
			}
			return undefined;
		};

		const frame = { data: new Uint8Array() } as any;

		// First read
		const err1 = await gr.readFrame(frame);
		assertEquals(err1, undefined);
		assertEquals(frame.data.slice(0, data1.byteLength), data1);

		// Second read - buffer should be reused
		frame.data = new Uint8Array();
		const err2 = await gr.readFrame(frame);
		assertEquals(err2, undefined);
		assertEquals(frame.data.slice(0, data2.byteLength), data2);
	});

	await t.step("should cancel with StreamError", async () => {
		const ctx = createTestContext();
		const mockReader = new MockReceiveStream();
		const mockGroup = { sequence: 456n } as unknown as GroupMessage;
		const gr = new GroupReader(ctx, mockReader as any, mockGroup);

		await gr.cancel(404, "Not found");

		assertEquals(mockReader.cancelCalls.length > 0, true);
		assertInstanceOf(mockReader.cancelCalls[0], StreamError);
	});
});
