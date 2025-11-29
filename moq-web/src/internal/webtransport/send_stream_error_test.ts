import { assertEquals, assertInstanceOf } from "@std/assert";
import { SendStream } from "./send_stream.ts";
import { StreamError } from "./error.ts";

Deno.test("SendStream - write handles WebTransportError with null streamErrorCode as EOFError", async () => {
	const writable = new WritableStream<Uint8Array>({
		write(_chunk) {
			// Reject with WebTransportError-like object (non-Error)
			return Promise.reject({ source: "stream", streamErrorCode: null });
		},
	});
	const s = new SendStream({ stream: writable, streamId: 1n });

	const [n, err] = await s.write(new Uint8Array([1, 2, 3]));
	assertEquals(n, 0);
	// err should be instanceof Error (EOFError or Error)
	assertInstanceOf(err, Error);

	// Subsequent writes should return same error
	const [n2, err2] = await s.write(new Uint8Array([4, 5]));
	assertEquals(n2, 0);
	assertInstanceOf(err2, Error);
});

Deno.test("SendStream - write handles WebTransportError with streamErrorCode set as StreamError", async () => {
	const writable = new WritableStream<Uint8Array>({
		write(_chunk) {
			// Reject with WebTransportError-like object (non-Error)
			return Promise.reject({ source: "stream", streamErrorCode: 123 });
		},
	});
	const s = new SendStream({ stream: writable, streamId: 2n });

	const [n, err] = await s.write(new Uint8Array([1, 2, 3]));
	assertEquals(n, 0);
	assertInstanceOf(err, StreamError);

	// Further writes should immediately return the same StreamError
	const [n2, err2] = await s.write(new Uint8Array([4, 5]));
	assertEquals(n2, 0);
	assertInstanceOf(err2, StreamError);
});

Deno.test("SendStream - cancel sets err and aborts", async () => {
	let aborted = false;
	const writable = new WritableStream<Uint8Array>({
		write(_chunk) {/* no-op */},
		abort(_reason) {
			aborted = true;
			return Promise.resolve();
		},
	});
	const s = new SendStream({ stream: writable, streamId: 3n });
	await s.cancel(1);
	const [n, err] = await s.write(new Uint8Array([1]));
	assertEquals(n, 0);
	// err should be a StreamError
	assertInstanceOf(err, StreamError);
	assertEquals(aborted, true);
});
