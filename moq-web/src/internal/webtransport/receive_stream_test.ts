import { assertEquals, assertInstanceOf } from "@std/assert";
import { ReceiveStream } from "./receive_stream.ts";
import { EOFError } from "@okudai/golikejs/io";
import { StreamError } from "./error.ts";

function setupReader(data: Uint8Array[]) {
	let index = 0;
	const readableStream = new ReadableStream<Uint8Array>({
		pull(controller) {
			if (index < data.length) {
				controller.enqueue(data[index]!);
				index++;
			} else {
				controller.close();
			}
		},
	});
	const reader = new ReceiveStream({ stream: readableStream, streamId: 1n });
	return { reader };
}

Deno.test("ReceiveStream", async (t) => {
	await t.step("read - should read data from stream", async () => {
		const { reader } = setupReader([new Uint8Array([1, 2, 3, 4, 5])]);

		const buf = new Uint8Array(5);
		const [n, err] = await reader.read(buf);

		assertEquals(err, undefined);
		assertEquals(n, 5);
		assertEquals(buf, new Uint8Array([1, 2, 3, 4, 5]));
	});

	await t.step("read - should return EOFError when stream is empty", async () => {
		const { reader } = setupReader([]);

		const buf = new Uint8Array(5);
		const [n, err] = await reader.read(buf);

		assertEquals(n, 0);
		assertInstanceOf(err, EOFError);
	});

	await t.step("read - should handle partial reads", async () => {
		const { reader } = setupReader([new Uint8Array([1, 2, 3, 4, 5])]);

		// Read only 3 bytes
		const buf1 = new Uint8Array(3);
		const [n1, err1] = await reader.read(buf1);

		assertEquals(err1, undefined);
		assertEquals(n1, 3);
		assertEquals(buf1, new Uint8Array([1, 2, 3]));

		// Read remaining 2 bytes
		const buf2 = new Uint8Array(2);
		const [n2, err2] = await reader.read(buf2);

		assertEquals(err2, undefined);
		assertEquals(n2, 2);
		assertEquals(buf2, new Uint8Array([4, 5]));
	});

	await t.step("read - should handle multiple chunks", async () => {
		const { reader } = setupReader([
			new Uint8Array([1, 2]),
			new Uint8Array([3, 4, 5]),
		]);

		// Read all 5 bytes across chunks
		const buf1 = new Uint8Array(2);
		const [n1, err1] = await reader.read(buf1);
		assertEquals(err1, undefined);
		assertEquals(n1, 2);
		assertEquals(buf1, new Uint8Array([1, 2]));

		const buf2 = new Uint8Array(3);
		const [n2, err2] = await reader.read(buf2);
		assertEquals(err2, undefined);
		assertEquals(n2, 3);
		assertEquals(buf2, new Uint8Array([3, 4, 5]));
	});

	await t.step("read - should buffer excess data", async () => {
		const { reader } = setupReader([new Uint8Array([1, 2, 3, 4, 5])]);

		// Read only 2 bytes, leaving 3 in buffer
		const buf1 = new Uint8Array(2);
		const [n1, err1] = await reader.read(buf1);
		assertEquals(err1, undefined);
		assertEquals(n1, 2);
		assertEquals(buf1, new Uint8Array([1, 2]));

		// Read the remaining 3 bytes from buffer
		const buf2 = new Uint8Array(3);
		const [n2, err2] = await reader.read(buf2);
		assertEquals(err2, undefined);
		assertEquals(n2, 3);
		assertEquals(buf2, new Uint8Array([3, 4, 5]));
	});

	await t.step("id - should return stream id", () => {
		const { reader } = setupReader([]);
		assertEquals(reader.id, 1n);
	});

	await t.step("cancel - should cancel the stream", async () => {
		let cancelReason: unknown = undefined;
		const readableStream = new ReadableStream<Uint8Array>({
			cancel(reason) {
				cancelReason = reason;
			},
		});
		const reader = new ReceiveStream({ stream: readableStream, streamId: 1n });

		const error = new StreamError(1, "test error");
		await reader.cancel(error);

		assertEquals(cancelReason, error);
	});

	await t.step("read - should handle large buffer request", async () => {
		const { reader } = setupReader([new Uint8Array([1, 2, 3])]);

		// Request more than available
		const buf = new Uint8Array(10);
		const [n, err] = await reader.read(buf);

		assertEquals(err, undefined);
		assertEquals(n, 3);
		assertEquals(buf.subarray(0, 3), new Uint8Array([1, 2, 3]));
	});
});
