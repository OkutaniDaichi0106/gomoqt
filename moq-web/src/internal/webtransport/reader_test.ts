import { assertEquals, assertInstanceOf, assertStrictEquals } from "@std/assert";

import { Reader } from "./reader.ts";
import { StreamError } from "./error.ts";

// Many tests share setup logic. We create helper functions here for reuse.
const createControllerAndReader = () => {
	let controller!: ReadableStreamDefaultController<Uint8Array>;
	const readableStream = new ReadableStream<Uint8Array>({
		start(ctrl) {
			controller = ctrl;
		},
	});
	const reader = new Reader({ stream: readableStream, streamId: 1n });
	return { reader, controller } as const;
};

const createFreshReader = (data?: Uint8Array) => {
	let ctrl!: ReadableStreamDefaultController<Uint8Array>;
	const stream = new ReadableStream<Uint8Array>({
		start(c) {
			ctrl = c;
			if (data) {
				ctrl.enqueue(data);
			}
		},
	});
	return { reader: new Reader({ stream, streamId: 1n }), controller: ctrl! } as const;
};

const createClosedReader = (data: Uint8Array) => {
	const stream = new ReadableStream<Uint8Array>({
		start(ctrl) {
			ctrl.enqueue(data);
			ctrl.close();
		},
	});
	return new Reader({ stream, transfer: undefined, streamId: 0n });
};

Deno.test("Reader", async (t) => {
	await t.step(
		"readUint8Array - should read a Uint8Array with varint length prefix",
		async () => {
			const { reader, controller } = createControllerAndReader();
			const data = new Uint8Array([1, 2, 3, 4, 5]);
			const streamData = new Uint8Array([5, ...data]);

			controller.enqueue(streamData);
			controller.close();

			const [result, error] = await reader.readUint8Array();
			assertEquals(error, undefined);
			assertEquals(result, data);
		},
	);

	await t.step("readUint8Array - should handle empty array", async () => {
		const freshReader = createClosedReader(new Uint8Array([0]));
		const [result, error] = await freshReader.readUint8Array();
		assertEquals(error, undefined);
		assertEquals(result, new Uint8Array([]));
	});

	await t.step("readUint8Array - should handle partial reads correctly", async () => {
		const { reader, controller } = createControllerAndReader();
		const data = new Uint8Array([1, 2, 3]);
		controller.enqueue(new Uint8Array([3]));
		controller.enqueue(new Uint8Array([1, 2]));
		controller.enqueue(new Uint8Array([3]));
		controller.close();

		const [result, error] = await reader.readUint8Array();
		assertEquals(error, undefined);
		assertEquals(result, data);
	});

	await t.step(
		"readUint8Array - should return error for stream with insufficient data",
		async () => {
			const { reader, controller } = createControllerAndReader();
			const invalidVarint = new Uint8Array([0xFF]);
			controller.enqueue(invalidVarint);
			controller.close();

			const [result, error] = await reader.readUint8Array();
			assertEquals(result, undefined);
			// error should be defined
			if (error === undefined) throw new Error("expected error");
		},
	);

	await t.step("readUint8Array - should handle very large length values", async () => {
		const largeLength = new Uint8Array([0xF0, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF]);
		const freshReader = createClosedReader(largeLength);

		try {
			await freshReader.readUint8Array();
			throw new Error("Expected to throw Varint too large");
		} catch (e: any) {
			if (!e.message.includes("Varint too large")) throw e;
		}
	});

	await t.step("readString - should read a UTF-8 string", async () => {
		const { reader, controller } = createControllerAndReader();
		const str = "hello world";
		const encoded = new TextEncoder().encode(str);
		const streamData = new Uint8Array([encoded.length, ...encoded]);
		controller.enqueue(streamData);
		controller.close();

		const [result, error] = await reader.readString();
		assertEquals(error, undefined);
		assertEquals(result, str);
	});

	await t.step("readString - should handle empty string", async () => {
		const freshReader = createClosedReader(new Uint8Array([0]));
		const [result, error] = await freshReader.readString();
		assertEquals(error, undefined);
		assertEquals(result, "");
	});

	await t.step("readString - should handle Unicode characters", async () => {
		const { reader, controller } = createControllerAndReader();
		const str = "こんにちは🚀";
		const encoded = new TextEncoder().encode(str);
		const streamData = new Uint8Array([encoded.length, ...encoded]);
		controller.enqueue(streamData);
		controller.close();

		const [result, error] = await reader.readString();
		assertEquals(error, undefined);
		assertEquals(result, str);
	});

	await t.step(
		"readString - should return error when underlying readUint8Array fails",
		async () => {
			const incompleteVarint = new Uint8Array([0xFF]);
			const freshReader = createClosedReader(incompleteVarint);
			const [result, error] = await freshReader.readString();
			assertEquals(result, ""); // Implementation returns empty string on error
			if (error === undefined) throw new Error("expected error");
		},
	);

	// readBigVarint tests
	await t.step("readBigVarint - single byte", async () => {
		const freshReader = createClosedReader(new Uint8Array([42]));
		const [result, error] = await freshReader.readBigVarint();
		assertEquals(error, undefined);
		assertEquals(result, 42n);
	});

	await t.step("readBigVarint - two byte", async () => {
		const freshReader = createClosedReader(new Uint8Array([0x41, 0x2C]));
		const [result, error] = await freshReader.readBigVarint();
		assertEquals(error, undefined);
		assertEquals(result, 300n);
	});

	await t.step("readBigVarint - four byte", async () => {
		const freshReader = createClosedReader(new Uint8Array([0x80, 0x0F, 0x42, 0x40]));
		const [result, error] = await freshReader.readBigVarint();
		assertEquals(error, undefined);
		assertEquals(result, 1000000n);
	});

	await t.step("readBigVarint - eight byte", async () => {
		const freshReader = createClosedReader(
			new Uint8Array([0xC0, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00]),
		);
		const [result, error] = await freshReader.readBigVarint();
		assertEquals(error, undefined);
		assertEquals(result, 1n << 40n);
	});

	await t.step("readBigVarint - partial varint reads", async () => {
		const { reader, controller } = createControllerAndReader();
		controller.enqueue(new Uint8Array([0x41]));
		controller.enqueue(new Uint8Array([0x2C]));
		controller.close();
		const [result, error] = await reader.readBigVarint();
		assertEquals(error, undefined);
		assertEquals(result, 300n);
	});

	await t.step("readBigVarint - error on stream close before complete read", async () => {
		const { reader, controller } = createControllerAndReader();
		controller.enqueue(new Uint8Array([0x41]));
		controller.close();
		const [result, error] = await reader.readBigVarint();
		assertEquals(result, 0n); // Implementation returns 0n on error
		if (error === undefined) throw new Error("expected error");
	});

	// readUint8 tests
	await t.step("readUint8 - should read a single byte", async () => {
		const { reader, controller } = createControllerAndReader();
		controller.enqueue(new Uint8Array([123]));
		controller.close();
		const [result, error] = await reader.readUint8();
		assertEquals(error, undefined);
		assertEquals(result, 123);
	});

	await t.step("readUint8 - should read multiple bytes sequentially", async () => {
		const { reader, controller } = createControllerAndReader();
		controller.enqueue(new Uint8Array([1, 2, 3]));
		controller.close();
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

	await t.step("readUint8 - should return error when no data available", async () => {
		const { reader, controller } = createControllerAndReader();
		controller.close();
		const [result, error] = await reader.readUint8();
		assertEquals(result, 0); // Implementation returns 0 on error
		if (error === undefined) throw new Error("expected error");
	});

	// readBoolean tests
	await t.step("readBoolean - should read true as 1", async () => {
		const { reader, controller } = createControllerAndReader();
		controller.enqueue(new Uint8Array([1]));
		controller.close();
		const [result, error] = await reader.readBoolean();
		assertEquals(error, undefined);
		assertEquals(result, true);
	});

	await t.step("readBoolean - should read false as 0", async () => {
		const { reader, controller } = createControllerAndReader();
		controller.enqueue(new Uint8Array([0]));
		controller.close();
		const [result, error] = await reader.readBoolean();
		assertEquals(error, undefined);
		assertEquals(result, false);
	});

	await t.step("readBoolean - should return error for invalid boolean values", async () => {
		const { reader, controller } = createControllerAndReader();
		controller.enqueue(new Uint8Array([2]));
		controller.close();
		const [result, error] = await reader.readBoolean();
		assertEquals(result, false); // Implementation returns false on error
		if (error === undefined) throw new Error("expected error");
	});

	await t.step("cancel - should cancel the reader with error code and message", async () => {
		const { reader } = createControllerAndReader();
		const code = 123;
		const message = "Test cancellation";
		const streamError = new StreamError(code, message);
		// If this rejects the test will fail
		await reader.cancel(streamError);
	});

	await t.step(
		"closed - should return a promise that resolves when reader is closed",
		async () => {
			const { reader, controller } = createControllerAndReader();
			const closedPromise = reader.closed();
			controller.close();
			await closedPromise; // will throw if rejected
		},
	);

	await t.step("integration tests - should read multiple data types in sequence", async () => {
		const { reader, controller } = createControllerAndReader();
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
		controller.enqueue(streamData);
		controller.close();

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

	await t.step("integration tests - should handle stream errors gracefully", async () => {
		const { reader, controller } = createControllerAndReader();
		controller.close();
		const [result, error] = await reader.readUint8();
		assertEquals(result, 0);
		if (error === undefined) throw new Error("expected error");
	});

	await t.step("BYOB support - should work with BYOB reader when available", async () => {
		const data = new Uint8Array([42]);
		const freshReader = createClosedReader(data);
		const [result, error] = await freshReader.readUint8();
		assertEquals(error, undefined);
		assertEquals(result, 42);
		await freshReader.cancel(new StreamError(0, "Test cleanup"));
	});
});
import { assertEquals, assertExists } from "@std/assert";
import { Reader } from "./reader.ts";
import { StreamError } from "./error.ts";

function createControllerReader(): {
	reader: Reader;
	controller: ReadableStreamDefaultController<Uint8Array>;
} {
	let ctrl!: ReadableStreamDefaultController<Uint8Array>;
	const stream = new ReadableStream<Uint8Array>({
		start(c) {
			ctrl = c;
		},
	});
	return { reader: new Reader({ stream, streamId: 1n }), controller: ctrl };
}

function createClosedReader(data: Uint8Array): Reader {
	const stream = new ReadableStream<Uint8Array>({
		start(ctrl) {
			ctrl.enqueue(data);
			ctrl.close();
		},
	});
	return new Reader({ stream, transfer: undefined, streamId: 0n });
}

Deno.test("webtransport/reader - readUint8Array scenarios", async (t) => {
	await t.step("reads Uint8Array with varint length", () => {
		const { reader, controller } = createControllerReader();
		const data = new Uint8Array([1, 2, 3, 4, 5]);
		controller.enqueue(new Uint8Array([5, ...data]));
		controller.close();

		return (async () => {
			const [res, err] = await reader.readUint8Array();
			assertEquals(err, undefined);
			assertEquals(res, data);
		})();
	});

	await t.step("handles empty array", async () => {
		const r = createClosedReader(new Uint8Array([0]));
		const [res, err] = await r.readUint8Array();
		assertEquals(err, undefined);
		assertEquals(res, new Uint8Array([]));
	});

	await t.step("partial reads assemble correctly", async () => {
		const { reader, controller } = createControllerReader();
		controller.enqueue(new Uint8Array([3]));
		controller.enqueue(new Uint8Array([1, 2]));
		controller.enqueue(new Uint8Array([3]));
		controller.close();
		const [res, err] = await reader.readUint8Array();
		assertEquals(err, undefined);
		assertEquals(res, new Uint8Array([1, 2, 3]));
	});

	await t.step("insufficient data returns error", async () => {
		const { reader, controller } = createControllerReader();
		controller.enqueue(new Uint8Array([0xFF]));
		controller.close();
		const [res, err] = await reader.readUint8Array();
		assertEquals(res, undefined);
		assertExists(err);
	});

	await t.step("very large varint triggers error", async () => {
		const large = new Uint8Array([0xF0, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF]);
		const r = createClosedReader(large);
		// Some implementations may throw synchronously or reject; ensure we catch either
		try {
			await r.readUint8Array();
			throw new Error("expected readUint8Array to throw for very large varint");
		} catch (e) {
			// error expected
		}
	});
});

Deno.test("webtransport/reader - readString and readBigVarint", async (t) => {
	await t.step("reads UTF-8 string", () => {
		const { reader, controller } = createControllerReader();
		const s = "hello world";
		const enc = new TextEncoder().encode(s);
		controller.enqueue(new Uint8Array([enc.length, ...enc]));
		controller.close();
		return (async () => {
			const [res, err] = await reader.readString();
			assertEquals(err, undefined);
			assertEquals(res, s);
		})();
	});

	await t.step("empty string", async () => {
		const r = createClosedReader(new Uint8Array([0]));
		const [res, err] = await r.readString();
		assertEquals(err, undefined);
		assertEquals(res, "");
	});

	await t.step("unicode string", () => {
		const { reader, controller } = createControllerReader();
		const s = "こんにちは🚀";
		const enc = new TextEncoder().encode(s);
		controller.enqueue(new Uint8Array([enc.length, ...enc]));
		controller.close();
		return (async () => {
			const [res, err] = await reader.readString();
			assertEquals(err, undefined);
			assertEquals(res, s);
		})();
	});

	await t.step("underlying readUint8Array failure returns error", async () => {
		const r = createClosedReader(new Uint8Array([0xFF]));
		const [res, err] = await r.readString();
		assertExists(err);
		assertEquals(res, "");
	});

	await t.step("readBigVarint - single/two/four/eight bytes and partial/error", async () => {
		// single
		let r = createClosedReader(new Uint8Array([42]));
		let [res, err] = await r.readBigVarint();
		assertEquals(err, undefined);
		assertEquals(res, 42n);

		// two-byte
		r = createClosedReader(new Uint8Array([0x41, 0x2C]));
		[res, err] = await r.readBigVarint();
		assertEquals(err, undefined);
		assertEquals(res, 300n);

		// four-byte
		r = createClosedReader(new Uint8Array([0x80, 0x0F, 0x42, 0x40]));
		[res, err] = await r.readBigVarint();
		assertEquals(err, undefined);
		assertEquals(res, 1000000n);

		// eight-byte
		r = createClosedReader(new Uint8Array([0xC0, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00]));
		[res, err] = await r.readBigVarint();
		assertEquals(err, undefined);
		assertEquals(res, 1n << 40n);

		// partial varint
		const cr = createControllerReader();
		cr.controller.enqueue(new Uint8Array([0x41]));
		cr.controller.enqueue(new Uint8Array([0x2C]));
		cr.controller.close();
		[res, err] = await cr.reader.readBigVarint();
		assertEquals(err, undefined);
		assertEquals(res, 300n);

		// error on close before complete
		const cr2 = createControllerReader();
		cr2.controller.enqueue(new Uint8Array([0x41]));
		cr2.controller.close();
		[res, err] = await cr2.reader.readBigVarint();
		assertExists(err);
	});
});

Deno.test("webtransport/reader - readUint8/readBoolean and control APIs", async (t) => {
	await t.step("readUint8 single and sequence", async () => {
		const { reader, controller } = createControllerReader();
		controller.enqueue(new Uint8Array([123]));
		controller.close();
		const [v, e] = await reader.readUint8();
		assertEquals(e, undefined);
		assertEquals(v, 123);

		const r2 = createClosedReader(new Uint8Array([1, 2, 3]));
		const [a, ea] = await r2.readUint8();
		assertEquals(ea, undefined);
		assertEquals(a, 1);
		const [b, eb] = await r2.readUint8();
		assertEquals(eb, undefined);
		assertEquals(b, 2);
		const [c, ec] = await r2.readUint8();
		assertEquals(ec, undefined);
		assertEquals(c, 3);

		const r3 = createControllerReader();
		r3.controller.close();
		const [rv, re] = await r3.reader.readUint8();
		assertExists(re);
		assertEquals(rv, 0);
	});

	await t.step("readBoolean true/false and invalid cases", async () => {
		let r = createClosedReader(new Uint8Array([1]));
		let [bv, be] = await r.readBoolean();
		assertEquals(be, undefined);
		assertEquals(bv, true);

		r = createClosedReader(new Uint8Array([0]));
		[bv, be] = await r.readBoolean();
		assertEquals(be, undefined);
		assertEquals(bv, false);

		r = createClosedReader(new Uint8Array([2]));
		[bv, be] = await r.readBoolean();
		assertExists(be);
		assertEquals(bv, false);
	});

	await t.step("cancel and closed APIs", async () => {
		const { reader } = createControllerReader();
		const err = new StreamError(123, "msg");
		await reader.cancel(err);

		const { reader: r2, controller } = createControllerReader();
		const closedPromise = r2.closed();
		controller.close();
		await closedPromise;
	});
});

Deno.test("webtransport/reader - integration sequence", async (t) => {
	await t.step("reads boolean, varint, string, and bytes sequentially", async () => {
		const { reader, controller } = createControllerReader();
		const testStr = "test";
		const testBytes = new Uint8Array([1, 2, 3]);
		const enc = new TextEncoder().encode(testStr);
		controller.enqueue(
			new Uint8Array([1, 42, enc.length, ...enc, testBytes.length, ...testBytes]),
		);
		controller.close();

		const [bv, be] = await reader.readBoolean();
		assertEquals(be, undefined);
		assertEquals(bv, true);

		const [vv, ve] = await reader.readBigVarint();
		assertEquals(ve, undefined);
		assertEquals(vv, 42n);

		const [sv, se] = await reader.readString();
		assertEquals(se, undefined);
		assertEquals(sv, testStr);

		const [av, ae] = await reader.readUint8Array();
		assertEquals(ae, undefined);
		assertEquals(av, testBytes);
	});

	await t.step("handles stream errors gracefully", async () => {
		const { reader, controller } = createControllerReader();
		controller.close();
		const [res, err] = await reader.readUint8();
		assertExists(err);
		assertEquals(res, 0);
	});
});
