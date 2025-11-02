import { assertEquals, assertExists, assertThrows, fail } from "@std/assert";
import { SendStream } from "./send_stream.ts";
import { StreamError } from "./error.ts";

function setupWriter() {
	const writtenData: Uint8Array[] = [];
	const state = { streamClosed: false };
	const writableStream = new WritableStream<Uint8Array>({
		write(chunk) {
			writtenData.push(chunk);
		},
		close() {
			state.streamClosed = true;
		},
	});
	const writer = new SendStream({ stream: writableStream, streamId: 1n });
	return { writer, writtenData, state };
}

Deno.test("Writer", async (t) => {
	await t.step(
		"writeUint8Array - should write a Uint8Array with varint length prefix",
		async () => {
			const { writer, writtenData, state } = setupWriter();
			try {
				const data = new Uint8Array([1, 2, 3, 4, 5]);
				writer.writeUint8Array(data);
				await writer.flush();

				assertEquals(writtenData.length, 1);
				const written = writtenData[0]!;
				assertEquals(written[0], 5);
				assertEquals(written.slice(1), data);
			} finally {
				try {
					if (!state.streamClosed) await writer.close();
				} catch (_) { /* ignore cleanup errors */ }
			}
		},
	);

	await t.step("writeUint8Array - should throw error for data exceeding maximum length", () => {
		const { writer } = setupWriter();
		const largeData = { length: (1 << 30) + 1 } as Uint8Array;
		assertThrows(
			() => {
				writer.writeUint8Array(largeData);
			},
			Error,
			"Bytes length exceeds maximum limit",
		);
	});

	await t.step("writeUint8Array - should handle empty array", async () => {
		const { writer, writtenData, state } = setupWriter();
		try {
			const data = new Uint8Array([]);
			writer.writeUint8Array(data);
			await writer.flush();

			assertEquals(writtenData.length, 1);
			const written = writtenData[0]!;
			assertEquals(written[0], 0);
			assertEquals(written.length, 1);
		} finally {
			try {
				if (!state.streamClosed) await writer.close();
			} catch (_err) { /* ignore */ }
		}
	});
});

function createWriter(): { writer: SendStream; writtenData: Uint8Array[]; streamClosed: boolean } {
	const writtenData: Uint8Array[] = [];
	let streamClosed = false;
	const writableStream = new WritableStream<Uint8Array>({
		write(chunk) {
			writtenData.push(chunk);
		},
		close() {
			streamClosed = true;
		},
	});
	const writer = new SendStream({ stream: writableStream, streamId: 1n });
	return { writer, writtenData, streamClosed };
}

Deno.test("webtransport/writer - writeUint8Array", async (t) => {
	await t.step("writes Uint8Array with varint length prefix", async () => {
		const { writer, writtenData } = createWriter();
		const data = new Uint8Array([1, 2, 3, 4, 5]);
		writer.writeUint8Array(data);
		await writer.flush();
		assertEquals(writtenData.length, 1);
		const written = writtenData[0];
		assertExists(written);
		assertEquals(written[0], 5);
		assertEquals(written.slice(1), data);
	});

	await t.step("throws for data exceeding maximum length", () => {
		const { writer } = createWriter();
		const largeData = { length: (1 << 30) + 1 } as Uint8Array;
		assertThrows(
			() => {
				writer.writeUint8Array(largeData);
			},
			Error,
			"Bytes length exceeds maximum limit",
		);
	});

	await t.step("handles empty array", async () => {
		const { writer, writtenData } = createWriter();
		writer.writeUint8Array(new Uint8Array([]));
		await writer.flush();
		assertEquals(writtenData.length, 1);
		const written = writtenData[0];
		assertExists(written);
		assertEquals(written[0], 0);
		assertEquals(written.length, 1);
	});
});

Deno.test("webtransport/writer - writeString", async (t) => {
	await t.step("writes UTF-8 string with length prefix", async () => {
		const { writer, writtenData } = createWriter();
		const str = "hello";
		writer.writeString(str);
		await writer.flush();
		assertEquals(writtenData.length, 1);
		const written = writtenData[0];
		assertExists(written);
		assertEquals(written[0], 5);
		const expectedBytes = new TextEncoder().encode(str);
		assertEquals(written.slice(1), expectedBytes);
	});

	await t.step("handles empty string", async () => {
		const { writer, writtenData } = createWriter();
		writer.writeString("");
		await writer.flush();
		assertEquals(writtenData.length, 1);
		const written = writtenData[0];
		assertExists(written);
		assertEquals(written[0], 0);
		assertEquals(written.length, 1);
	});

	await t.step("handles Unicode characters", async () => {
		const { writer, writtenData } = createWriter();
		const str = "こんにちは";
		writer.writeString(str);
		await writer.flush();
		assertEquals(writtenData.length, 1);
		const written = writtenData[0];
		assertExists(written);
		const expectedBytes = new TextEncoder().encode(str);
		assertEquals(written[0], expectedBytes.length);
		assertEquals(written.slice(1), expectedBytes);
	});
});

Deno.test("webtransport/writer - writeBigVarint", async (t) => {
	await t.step("single/two/four/eight byte encodings and errors", async () => {
		let out = createWriter();
		out.writer.writeBigVarint(42n);
		await out.writer.flush();
		assertEquals(out.writtenData.length, 1);
		assertEquals(out.writtenData[0], new Uint8Array([42]));

		out = createWriter();
		out.writer.writeBigVarint(300n);
		await out.writer.flush();
		assertEquals(out.writtenData.length, 1);
		let written = out.writtenData[0];
		assertExists(written);
		assertEquals(written.length, 2);
		assertEquals(written[0], 0x41);
		assertEquals(written[1], 0x2C);

		out = createWriter();
		out.writer.writeBigVarint(1000000n);
		await out.writer.flush();
		written = out.writtenData[0];
		const _w = out.writtenData[0];
		if (!_w) fail("missing written data");
		assertEquals(_w.length, 4);
		const first = _w[0] ?? 0;
		assertEquals(first & 0xC0, 0x80);

		out = createWriter();
		out.writer.writeBigVarint(1n << 40n);
		await out.writer.flush();
		written = out.writtenData[0];
		assertExists(written);
		assertEquals(written.length, 8);
		assertEquals(written[0], 0xC0);

		out = createWriter();
		assertThrows(
			() => {
				out.writer.writeBigVarint(-1n);
			},
			Error,
			"Varint cannot be negative",
		);

		out = createWriter();
		const maxValue = (1n << 62n) - 1n;
		assertThrows(
			() => {
				out.writer.writeBigVarint(maxValue + 1n);
			},
			Error,
			"Value exceeds maximum varint size",
		);
	});
});

Deno.test("webtransport/writer - writeBoolean", async (t) => {
	await t.step("true/false", async () => {
		let out = createWriter();
		out.writer.writeBoolean(true);
		await out.writer.flush();
		assertEquals(out.writtenData.length, 1);
		assertEquals(out.writtenData[0], new Uint8Array([1]));

		out = createWriter();
		out.writer.writeBoolean(false);
		await out.writer.flush();
		assertEquals(out.writtenData.length, 1);
		assertEquals(out.writtenData[0], new Uint8Array([0]));
	});
});

Deno.test("webtransport/writer - flush", async (t) => {
	await t.step("flush success and multiple flushes", async () => {
		let out = createWriter();
		out.writer.writeBoolean(true);
		let err = await out.writer.flush();
		assertEquals(err, undefined);
		assertEquals(out.writtenData.length, 1);

		out = createWriter();
		out.writer.writeBoolean(true);
		await out.writer.flush();
		out.writer.writeBoolean(false);
		err = await out.writer.flush();
		assertEquals(err, undefined);
		assertEquals(out.writtenData.length, 2);

		out = createWriter();
		err = await out.writer.flush();
		assertEquals(err, undefined);
		assertEquals(out.writtenData.length, 0);
	});
});

Deno.test("webtransport/writer - close/cancel/closed", async (t) => {
	await t.step("close behavior", async () => {
		const out = createWriter();
		await out.writer.close();
		// second close should not throw (implementation dependent)
		try {
			await out.writer.close();
		} catch (e) {
			assertExists(e);
		}
	});

	await t.step("cancel resolves", async () => {
		const out = createWriter();
		const err = new StreamError(1, "Test error");
		await out.writer.cancel(err);
	});

	await t.step("closed promise resolves on close", async () => {
		const out = createWriter();
		const p = out.writer.closed();
		await out.writer.close();
		await p;
	});
});

Deno.test("webtransport/writer - integration and string array", async (t) => {
	await t.step("writes multiple data types in sequence", async () => {
		const out = createWriter();
		out.writer.writeBoolean(true);
		out.writer.writeBigVarint(123n);
		out.writer.writeString("test");
		out.writer.writeUint8Array(new Uint8Array([1, 2, 3]));
		await out.writer.flush();
		assertEquals(out.writtenData.length, 1);
		const written = out.writtenData[0];
		assertExists(written);
		if (!(written.length > 10)) fail("expected written length > 10");
		assertEquals(written[0], 1);
	});

	await t.step("writeStringArray", async () => {
		const out = createWriter();
		const arr = ["hello", "world"];
		out.writer.writeStringArray(arr);
		await out.writer.flush();
		assertEquals(out.writtenData.length, 1);
		const written = out.writtenData[0];
		assertExists(written);
		assertEquals(written[0], 2);
		let offset = 1;
		for (const str of arr) {
			const strBytes = new TextEncoder().encode(str);
			assertEquals(written[offset], strBytes.length);
			assertEquals(written.slice(offset + 1, offset + 1 + strBytes.length), strBytes);
			offset += 1 + strBytes.length;
		}

		const out2 = createWriter();
		out2.writer.writeStringArray([]);
		await out2.writer.flush();
		assertEquals(out2.writtenData.length, 1);
		const written2 = out2.writtenData[0];
		assertExists(written2);
		const _w2 = written2!;
		assertEquals(_w2[0], 0);
		assertEquals(_w2.length, 1);
	});
});
