import { assertEquals } from "@std/assert";
import { encodeMessageToUint8Array, MockWebTransport } from "../testing/mock_webtransport.ts";
import { SessionServerMessage } from "../internal/message/mod.ts";
import { DEFAULT_CLIENT_VERSIONS } from "../version.ts";

Deno.test("encodeMessageToUint8Array returns combined bytes of multiple writes", async () => {
	const msg = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
	const bytes = await encodeMessageToUint8Array(async (w: any) => {
		return await msg.encode(w);
	});
	assertEquals(bytes instanceof Uint8Array, true);
	assertEquals(bytes.length > 0, true);
});

Deno.test("MockWebTransport.createBidirectionalStream returns readable and writable streams", async () => {
	const mock = new MockWebTransport([]);
	const stream = await mock.createBidirectionalStream();
	assertEquals(typeof stream.readable.getReader, "function");
	assertEquals(typeof stream.writable.getWriter, "function");
});

Deno.test("MockWebTransport - readable contains server bytes and closed behavior", async () => {
	const serverBytes = new Uint8Array([11, 22, 33]);
	const mock = new MockWebTransport([serverBytes], { keepStreamsOpen: false });
	const s = await mock.createBidirectionalStream();
	const reader = s.readable.getReader();
	const v = await reader.read();
	assertEquals(new Uint8Array(v.value), serverBytes);
	const next = await reader.read();
	assertEquals(next.done, true);
});

Deno.test("MockWebTransport - createUnidirectionalStream works and datagrams exists", async () => {
	const mock = new MockWebTransport([]);
	// datagrams should be present with incoming and outgoing
	assertEquals(typeof mock.datagrams, "object");
	assertEquals(typeof mock.datagrams.incoming?.getReader, "function");
	assertEquals(typeof mock.datagrams.outgoing?.write, "function");

	const w = await mock.createUnidirectionalStream();
	const writer = w.getWriter();
	await writer.write(new Uint8Array([99]));
	await writer.close();
});

Deno.test("MockWebTransport - closed resolves when close is called", async () => {
	const mock = new MockWebTransport([]);
	let info: WebTransportCloseInfo | undefined;
	const p = mock.closed.then((i) => (info = i));
	mock.close({ closeCode: 55, reason: "bye" });
	await p;
	assertEquals(info?.closeCode, 55);
	assertEquals(info?.reason, "bye");
});
