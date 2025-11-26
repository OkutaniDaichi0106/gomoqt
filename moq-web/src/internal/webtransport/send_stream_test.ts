import { assertEquals } from "@std/assert";
import { SendStream } from "./send_stream.ts";
import { StreamError } from "./error.ts";

function setupWriter() {
	const writtenData: Uint8Array[] = [];
	const state = { streamClosed: false };
	const writableStream = new WritableStream<Uint8Array>({
		write(chunk) {
			writtenData.push(chunk.slice()); // Clone to preserve data
		},
		close() {
			state.streamClosed = true;
		},
	});
	const writer = new SendStream({ stream: writableStream, streamId: 1n });
	return { writer, writtenData, state };
}

Deno.test("SendStream", async (t) => {
	await t.step("write - should write data to stream", async () => {
		const { writer, writtenData, state } = setupWriter();
		try {
			const data = new Uint8Array([1, 2, 3, 4, 5]);
			const [n, err] = await writer.write(data);

			assertEquals(err, undefined);
			assertEquals(n, 5);
			assertEquals(writtenData.length, 1);
			assertEquals(writtenData[0], data);
		} finally {
			try {
				if (!state.streamClosed) await writer.close();
			} catch (_) { /* ignore cleanup errors */ }
		}
	});

	await t.step("write - should handle empty array", async () => {
		const { writer, writtenData, state } = setupWriter();
		try {
			const data = new Uint8Array([]);
			const [n, err] = await writer.write(data);

			assertEquals(err, undefined);
			assertEquals(n, 0);
			assertEquals(writtenData.length, 1);
			assertEquals(writtenData[0]!.length, 0);
		} finally {
			try {
				if (!state.streamClosed) await writer.close();
			} catch (_) { /* ignore cleanup errors */ }
		}
	});

	await t.step("write - should handle multiple writes", async () => {
		const { writer, writtenData, state } = setupWriter();
		try {
			const data1 = new Uint8Array([1, 2, 3]);
			const data2 = new Uint8Array([4, 5, 6, 7]);

			const [n1, err1] = await writer.write(data1);
			assertEquals(err1, undefined);
			assertEquals(n1, 3);

			const [n2, err2] = await writer.write(data2);
			assertEquals(err2, undefined);
			assertEquals(n2, 4);

			assertEquals(writtenData.length, 2);
			assertEquals(writtenData[0], data1);
			assertEquals(writtenData[1], data2);
		} finally {
			try {
				if (!state.streamClosed) await writer.close();
			} catch (_) { /* ignore cleanup errors */ }
		}
	});

	await t.step("close - should close the stream", async () => {
		const { writer, state } = setupWriter();
		await writer.close();
		assertEquals(state.streamClosed, true);
	});

	await t.step("cancel - should abort the stream with error", async () => {
		const writtenData: Uint8Array[] = [];
		let abortReason: unknown = undefined;
		const writableStream = new WritableStream<Uint8Array>({
			write(chunk) {
				writtenData.push(chunk);
			},
			abort(reason) {
				abortReason = reason;
			},
		});
		const writer = new SendStream({ stream: writableStream, streamId: 1n });

		const error = new StreamError(1, "test error");
		await writer.cancel(error);

		assertEquals(abortReason, error);
	});

	await t.step("id - should return stream id", () => {
		const { writer } = setupWriter();
		assertEquals(writer.id, 1n);
	});

	await t.step("write - should return error on stream failure", async () => {
		const writableStream = new WritableStream<Uint8Array>({
			write(_chunk) {
				throw new Error("Write failed");
			},
		});
		const writer = new SendStream({ stream: writableStream, streamId: 1n });

		const data = new Uint8Array([1, 2, 3]);
		const [n, err] = await writer.write(data);

		assertEquals(n, 0);
		assertEquals(err instanceof Error, true);
		assertEquals(err!.message, "Write failed");
	});
});
