import { assertEquals, assertInstanceOf } from "@std/assert";
import { WebTransportSession } from "./connection.ts";
import { WebTransportSessionError } from "./error.ts";

class FailingMockWebTransport {
	ready = Promise.resolve();
	closed = Promise.resolve({ closeCode: 123, reason: "fail" });
	incomingBidirectionalStreams = new ReadableStream({ start(_c) {} });
	incomingUnidirectionalStreams = new ReadableStream({ start(_c) {} });
	async createBidirectionalStream() {
		const err = { source: "session" as const } as any; // not an Error
		throw err;
	}
}

Deno.test("SessionImpl.openStream maps WebTransportError session source to SessionError", async () => {
	const session = new WebTransportSession(
		(new FailingMockWebTransport()) as any,
	);
	const [stream, err] = await session.openStream();
	assertEquals(stream, undefined);
	assertInstanceOf(err, WebTransportSessionError);
});

Deno.test("SessionImpl.acceptStream returns error when incoming reader yields done", async () => {
	const mock = {
		incomingBidirectionalStreams: new ReadableStream({
			start(controller) {
				controller.close();
			},
		}),
		incomingUnidirectionalStreams: new ReadableStream({
			start(controller) {
				controller.close();
			},
		}),
		ready: Promise.resolve(),
		closed: Promise.resolve({ closeCode: undefined, reason: undefined }),
	} as any;

	const session = new WebTransportSession(mock);

	const [s1, e1] = await session.acceptStream();
	assertEquals(s1, undefined);
	assertInstanceOf(e1, Error);

	const [s2, e2] = await session.acceptUniStream();
	assertEquals(s2, undefined);
	assertInstanceOf(e2, Error);
});

Deno.test("SessionImpl.openStream returns error when createBidirectionalStream throws Error", async () => {
	const mock = {
		incomingBidirectionalStreams: new ReadableStream({ start(_c) {} }),
		incomingUnidirectionalStreams: new ReadableStream({ start(_c) {} }),
		ready: Promise.resolve(),
		closed: Promise.resolve({ closeCode: undefined, reason: undefined }),
		async createBidirectionalStream() {
			throw new Error("connection failed");
		},
	} as any;

	const session = new WebTransportSession(mock);
	const [stream, err] = await session.openStream();
	assertEquals(stream, undefined);
	assertInstanceOf(err, Error);
	assertEquals(err?.message, "connection failed");
});

Deno.test("SessionImpl.openStream handles non-session WebTransportError", async () => {
	const mock = {
		incomingBidirectionalStreams: new ReadableStream({ start(_c) {} }),
		incomingUnidirectionalStreams: new ReadableStream({ start(_c) {} }),
		ready: Promise.resolve(),
		closed: Promise.resolve({ closeCode: undefined, reason: undefined }),
		async createBidirectionalStream() {
			const err = { source: "stream" } as any;
			throw err;
		},
	} as any;

	const session = new WebTransportSession(mock);
	const [stream, err] = await session.openStream();
	assertEquals(stream, undefined);
	assertEquals((err as any).source, "stream");
});

Deno.test("SessionImpl.openUniStream returns error when createUnidirectionalStream throws", async () => {
	const mock = {
		incomingBidirectionalStreams: new ReadableStream({ start(_c) {} }),
		incomingUnidirectionalStreams: new ReadableStream({ start(_c) {} }),
		ready: Promise.resolve(),
		closed: Promise.resolve({ closeCode: undefined, reason: undefined }),
		async createUnidirectionalStream() {
			throw new Error("uni stream failed");
		},
	} as any;

	const session = new WebTransportSession(mock);
	const [stream, err] = await session.openUniStream();
	assertEquals(stream, undefined);
	assertInstanceOf(err, Error);
	assertEquals(err?.message, "uni stream failed");
});

Deno.test("SessionImpl.openStream succeeds with valid bidirectional stream", async () => {
	const mockWritableStream = {
		getWriter() {
			return {
				ready: Promise.resolve(),
				write: async () => {},
				close: async () => {},
				abort: async () => {},
				releaseLock: () => {},
				closed: Promise.resolve(),
			};
		},
	};
	const mockReadableStream = {
		getReader() {
			return {
				read: async () => ({ done: true, value: undefined }),
				cancel: async () => {},
				releaseLock: () => {},
			};
		},
	};

	const mock = {
		incomingBidirectionalStreams: new ReadableStream({ start(_c) {} }),
		incomingUnidirectionalStreams: new ReadableStream({ start(_c) {} }),
		ready: Promise.resolve(),
		closed: Promise.resolve({ closeCode: undefined, reason: undefined }),
		async createBidirectionalStream() {
			return {
				readable: mockReadableStream,
				writable: mockWritableStream,
			};
		},
	} as any;

	const session = new WebTransportSession(mock);
	const [stream, err] = await session.openStream();
	assertEquals(err, undefined);
	assertEquals(typeof stream?.id, "bigint");
});

Deno.test("SessionImpl.openUniStream succeeds with valid unidirectional stream", async () => {
	const mockWritableStream = {
		getWriter() {
			return {
				ready: Promise.resolve(),
				write: async () => {},
				close: async () => {},
				abort: async () => {},
				releaseLock: () => {},
				closed: Promise.resolve(),
			};
		},
	};

	const mock = {
		incomingBidirectionalStreams: new ReadableStream({ start(_c) {} }),
		incomingUnidirectionalStreams: new ReadableStream({ start(_c) {} }),
		ready: Promise.resolve(),
		closed: Promise.resolve({ closeCode: undefined, reason: undefined }),
		async createUnidirectionalStream() {
			return mockWritableStream;
		},
	} as any;

	const session = new WebTransportSession(mock);
	const [stream, err] = await session.openUniStream();
	assertEquals(err, undefined);
	assertEquals(typeof stream?.id, "bigint");
});

Deno.test("SessionImpl.close cancels readers", async () => {
	let closeCalled = false;
	const mock = {
		incomingBidirectionalStreams: new ReadableStream({ start(_c) {} }),
		incomingUnidirectionalStreams: new ReadableStream({ start(_c) {} }),
		ready: Promise.resolve(),
		closed: Promise.resolve({ closeCode: undefined, reason: undefined }),
		close(_info?: any) {
			closeCalled = true;
		},
	} as any;

	const session = new WebTransportSession(mock);
	session.close({ closeCode: 0, reason: "test" });
	assertEquals(closeCalled, true);
});

Deno.test("SessionImpl.ready and closed return promises from underlying transport", async () => {
	const mock = {
		incomingBidirectionalStreams: new ReadableStream({ start(_c) {} }),
		incomingUnidirectionalStreams: new ReadableStream({ start(_c) {} }),
		ready: Promise.resolve(),
		closed: Promise.resolve({ closeCode: 0, reason: "closed" }),
	} as any;

	const session = new WebTransportSession(mock);
	await session.ready;
	const closeInfo = await session.closed;
	assertEquals(closeInfo.closeCode, 0);
	assertEquals(closeInfo.reason, "closed");
});
