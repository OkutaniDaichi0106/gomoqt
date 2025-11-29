import { assertEquals, assertInstanceOf } from "@std/assert";
import { ReceiveStream } from "./receive_stream.ts";
import { StreamError } from "./error.ts";

Deno.test("ReceiveStream - read captures WebTransportError with null streamErrorCode as EOFError-like", async () => {
	const readable = new ReadableStream<Uint8Array>({
		pull(_controller) {
			return Promise.reject({ source: "stream", streamErrorCode: null });
		},
	});
	const r = new ReceiveStream({ stream: readable, streamId: 1n });

	const p = new Uint8Array(4);
	const [n, err] = await r.read(p);
	assertEquals(n, 0);
	assertInstanceOf(err, Error);
});

Deno.test("ReceiveStream - cancel sets error and subsequent read returns error", async () => {
	let canceled = false;
	const readable = new ReadableStream<Uint8Array>({
		start(_controller) {},
		cancel(_reason) {
			canceled = true;
			return Promise.resolve();
		},
	});
	const r = new ReceiveStream({ stream: readable, streamId: 2n });
	await r.cancel(1 as any);

	const p = new Uint8Array(4);
	const [n, err] = await r.read(p);
	assertEquals(n, 0);
	assertInstanceOf(err, StreamError);
	assertEquals(canceled, true);
});
