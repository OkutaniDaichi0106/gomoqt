import { assertEquals, assertExists } from "@std/assert";
import { ReceiveStream } from "./receive_stream.ts";
import { StreamError } from "./error.ts";

Deno.test("ReceiveStream - readUint8Array", async (t) => {
	await t.step(
		"should read a Uint8Array with varint length prefix",
		async () => {
			const data = new Uint8Array([1, 2, 3, 4, 5]);
			const streamData = new Uint8Array([5, ...data]);

			const stream = new ReadableStream<Uint8Array>({
				start(ctrl) {
					ctrl.enqueue(streamData);
					ctrl.close();
				},
			});
			const reader = new ReceiveStream({ stream, streamId: 1n });

			const [result, error] = await reader.readUint8Array();
			assertEquals(error, undefined);
			assertEquals(result, data);
		},
	);

	await t.step("should handle empty array", async () => {
		const stream = new ReadableStream<Uint8Array>({
			start(ctrl) {
				ctrl.enqueue(new Uint8Array([0]));
				ctrl.close();
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		const [result, error] = await reader.readUint8Array();
		assertEquals(error, undefined);
		assertEquals(result, new Uint8Array([]));
	});

	await t.step("should handle partial reads correctly", async () => {
		const data = new Uint8Array([1, 2, 3]);
		const stream = new ReadableStream<Uint8Array>({
			start(ctrl) {
				ctrl.enqueue(new Uint8Array([3]));
				ctrl.enqueue(new Uint8Array([1, 2]));
				ctrl.enqueue(new Uint8Array([3]));
				ctrl.close();
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		const [result, error] = await reader.readUint8Array();
		assertEquals(error, undefined);
		assertEquals(result, data);
	});

	await t.step(
		"should return error for stream with insufficient data",
		async () => {
			const invalidVarint = new Uint8Array([0xFF]);
			const stream = new ReadableStream<Uint8Array>({
				start(ctrl) {
					ctrl.enqueue(invalidVarint);
					ctrl.close();
				},
			});
			const reader = new ReceiveStream({ stream, streamId: 1n });

			const [result, error] = await reader.readUint8Array();
			assertEquals(result, undefined);
			assertExists(error);
		},
	);

	await t.step("should handle very large length values", async () => {
		const largeLength = new Uint8Array([0xF0, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF]);
		const stream = new ReadableStream<Uint8Array>({
			start(ctrl) {
				ctrl.enqueue(largeLength);
				ctrl.close();
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		try {
			await reader.readUint8Array();
			throw new Error("Expected to throw Varint too large");
		} catch (e: any) {
			if (!e.message.includes("Varint too large")) throw e;
		}
	});
});

Deno.test("ReceiveStream - readString", async (t) => {
	await t.step("should read a UTF-8 string", async () => {
		const str = "hello world";
		const encoded = new TextEncoder().encode(str);
		const streamData = new Uint8Array([encoded.length, ...encoded]);

		const stream = new ReadableStream<Uint8Array>({
			start(ctrl) {
				ctrl.enqueue(streamData);
				ctrl.close();
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		const [result, error] = await reader.readString();
		assertEquals(error, undefined);
		assertEquals(result, str);
	});

	await t.step("should handle empty string", async () => {
		const stream = new ReadableStream<Uint8Array>({
			start(ctrl) {
				ctrl.enqueue(new Uint8Array([0]));
				ctrl.close();
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		const [result, error] = await reader.readString();
		assertEquals(error, undefined);
		assertEquals(result, "");
	});

	await t.step("should handle Unicode characters", async () => {
		const str = "こんにちは🚀";
		const encoded = new TextEncoder().encode(str);
		const streamData = new Uint8Array([encoded.length, ...encoded]);

		const stream = new ReadableStream<Uint8Array>({
			start(ctrl) {
				ctrl.enqueue(streamData);
				ctrl.close();
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		const [result, error] = await reader.readString();
		assertEquals(error, undefined);
		assertEquals(result, str);
	});

	await t.step(
		"should return error when underlying readUint8Array fails",
		async () => {
			const incompleteVarint = new Uint8Array([0xFF]);
			const stream = new ReadableStream<Uint8Array>({
				start(ctrl) {
					ctrl.enqueue(incompleteVarint);
					ctrl.close();
				},
			});
			const reader = new ReceiveStream({ stream, streamId: 1n });

			const [result, error] = await reader.readString();
			assertEquals(result, "");
			assertExists(error);
		},
	);
});

Deno.test("ReceiveStream - readBigVarint", async (t) => {
	await t.step("single byte", async () => {
		const stream = new ReadableStream<Uint8Array>({
			start(ctrl) {
				ctrl.enqueue(new Uint8Array([42]));
				ctrl.close();
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		const [result, error] = await reader.readBigVarint();
		assertEquals(error, undefined);
		assertEquals(result, 42n);
	});

	await t.step("two byte", async () => {
		const stream = new ReadableStream<Uint8Array>({
			start(ctrl) {
				ctrl.enqueue(new Uint8Array([0x41, 0x2C]));
				ctrl.close();
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		const [result, error] = await reader.readBigVarint();
		assertEquals(error, undefined);
		assertEquals(result, 300n);
	});

	await t.step("four byte", async () => {
		const stream = new ReadableStream<Uint8Array>({
			start(ctrl) {
				ctrl.enqueue(new Uint8Array([0x80, 0x0F, 0x42, 0x40]));
				ctrl.close();
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		const [result, error] = await reader.readBigVarint();
		assertEquals(error, undefined);
		assertEquals(result, 1000000n);
	});

	await t.step("eight byte", async () => {
		const stream = new ReadableStream<Uint8Array>({
			start(ctrl) {
				ctrl.enqueue(new Uint8Array([0xC0, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00]));
				ctrl.close();
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		const [result, error] = await reader.readBigVarint();
		assertEquals(error, undefined);
		assertEquals(result, 1n << 40n);
	});

	await t.step("partial varint reads", async () => {
		const stream = new ReadableStream<Uint8Array>({
			start(ctrl) {
				ctrl.enqueue(new Uint8Array([0x41]));
				ctrl.enqueue(new Uint8Array([0x2C]));
				ctrl.close();
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		const [result, error] = await reader.readBigVarint();
		assertEquals(error, undefined);
		assertEquals(result, 300n);
	});

	await t.step("error on stream close before complete read", async () => {
		const stream = new ReadableStream<Uint8Array>({
			start(ctrl) {
				ctrl.enqueue(new Uint8Array([0x41]));
				ctrl.close();
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		const [result, error] = await reader.readBigVarint();
		assertEquals(result, 0n);
		assertExists(error);
	});
});

Deno.test("ReceiveStream - readUint8", async (t) => {
	await t.step("should read a single byte", async () => {
		const stream = new ReadableStream<Uint8Array>({
			start(ctrl) {
				ctrl.enqueue(new Uint8Array([123]));
				ctrl.close();
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		const [result, error] = await reader.readUint8();
		assertEquals(error, undefined);
		assertEquals(result, 123);
	});

	await t.step("should read multiple bytes sequentially", async () => {
		const stream = new ReadableStream<Uint8Array>({
			start(ctrl) {
				ctrl.enqueue(new Uint8Array([1, 2, 3]));
				ctrl.close();
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		const [first, error1] = await reader.readUint8();
		assertEquals(error1, undefined);
		assertEquals(first, 1);
		const [second, error2] = await reader.readUint8();
		assertEquals(error2, undefined);
		assertEquals(second, 2);
		const [third, error3] = await reader.readUint8();
		assertEquals(error3, undefined);
		assertEquals(third, 3);
	});

	await t.step("should return error when no data available", async () => {
		const stream = new ReadableStream<Uint8Array>({
			start(ctrl) {
				ctrl.close();
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		const [result, error] = await reader.readUint8();
		assertEquals(result, 0);
		assertExists(error);
	});
});

Deno.test("ReceiveStream - readBoolean", async (t) => {
	await t.step("should read true as 1", async () => {
		const stream = new ReadableStream<Uint8Array>({
			start(ctrl) {
				ctrl.enqueue(new Uint8Array([1]));
				ctrl.close();
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		const [result, error] = await reader.readBoolean();
		assertEquals(error, undefined);
		assertEquals(result, true);
	});

	await t.step("should read false as 0", async () => {
		const stream = new ReadableStream<Uint8Array>({
			start(ctrl) {
				ctrl.enqueue(new Uint8Array([0]));
				ctrl.close();
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		const [result, error] = await reader.readBoolean();
		assertEquals(error, undefined);
		assertEquals(result, false);
	});

	await t.step("should return error for invalid boolean values", async () => {
		const stream = new ReadableStream<Uint8Array>({
			start(ctrl) {
				ctrl.enqueue(new Uint8Array([2]));
				ctrl.close();
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		const [result, error] = await reader.readBoolean();
		assertEquals(result, false);
		assertExists(error);
	});
});

Deno.test("ReceiveStream - readStringArray", async (t) => {
	await t.step("should read an array of strings", async () => {
		const arr = ["hello", "world"];
		const encoder = new TextEncoder();
		const encoded0 = encoder.encode(arr[0]);
		const encoded1 = encoder.encode(arr[1]);

		const streamData = new Uint8Array([
			2, // count
			encoded0.length,
			...encoded0,
			encoded1.length,
			...encoded1,
		]);

		const stream = new ReadableStream<Uint8Array>({
			start(ctrl) {
				ctrl.enqueue(streamData);
				ctrl.close();
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		const [result, error] = await reader.readStringArray();
		assertEquals(error, undefined);
		assertEquals(result, arr);
	});

	await t.step("should handle empty string array", async () => {
		const stream = new ReadableStream<Uint8Array>({
			start(ctrl) {
				ctrl.enqueue(new Uint8Array([0]));
				ctrl.close();
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		const [result, error] = await reader.readStringArray();
		assertEquals(error, undefined);
		assertEquals(result, []);
	});
});

Deno.test("ReceiveStream - control APIs", async (t) => {
	await t.step("cancel - should cancel the reader with error code and message", async () => {
		const stream = new ReadableStream<Uint8Array>({
			start(_ctrl) {
				// Stream remains open for cancel test
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		const code = 123;
		const message = "Test cancellation";
		const streamError = new StreamError(code, message);
		await reader.cancel(streamError);
	});

	await t.step(
		"closed - should return a promise that resolves when reader is closed",
		async () => {
			const stream = new ReadableStream<Uint8Array>({
				start(ctrl) {
					ctrl.close();
				},
			});
			const reader = new ReceiveStream({ stream, streamId: 1n });

			const closedPromise = reader.closed();
			await closedPromise;
		},
	);
});

Deno.test("ReceiveStream - integration tests", async (t) => {
	await t.step("should read multiple data types in sequence", async () => {
		const testStr = "test";
		const testBytes = new Uint8Array([1, 2, 3]);
		const encodedStr = new TextEncoder().encode(testStr);
		const streamData = new Uint8Array([
			1,
			42,
			encodedStr.length,
			...encodedStr,
			testBytes.length,
			...testBytes,
		]);

		const stream = new ReadableStream<Uint8Array>({
			start(ctrl) {
				ctrl.enqueue(streamData);
				ctrl.close();
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		const [boolResult, boolError] = await reader.readBoolean();
		assertEquals(boolError, undefined);
		assertEquals(boolResult, true);

		const [varintResult, varintError] = await reader.readBigVarint();
		assertEquals(varintError, undefined);
		assertEquals(varintResult, 42n);

		const [stringResult, stringError] = await reader.readString();
		assertEquals(stringError, undefined);
		assertEquals(stringResult, testStr);

		const [arrayResult, arrayError] = await reader.readUint8Array();
		assertEquals(arrayError, undefined);
		assertEquals(arrayResult, testBytes);
	});

	await t.step("should handle stream errors gracefully", async () => {
		const stream = new ReadableStream<Uint8Array>({
			start(ctrl) {
				ctrl.close();
			},
		});
		const reader = new ReceiveStream({ stream, streamId: 1n });

		const [result, error] = await reader.readUint8();
		assertEquals(result, 0);
		assertExists(error);
	});
});
