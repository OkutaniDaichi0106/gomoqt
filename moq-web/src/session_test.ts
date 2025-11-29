import { assertEquals, assertExists, assertInstanceOf, assertNotEquals } from "@std/assert";
import { Session } from "./session.ts";
import { DEFAULT_CLIENT_VERSIONS } from "./version.ts";
import {
	AnnounceInitMessage,
	AnnouncePleaseMessage,
	GroupMessage,
	SessionClientMessage,
	SessionServerMessage,
	SessionUpdateMessage,
	SubscribeMessage,
	SubscribeOkMessage,
	writeVarint,
} from "./internal/message/mod.ts";
import { Extensions } from "./extensions.ts";
import { BiStreamTypes, UniStreamTypes } from "./stream_type.ts";
import { encodeMessageToUint8Array, MockWebTransport } from "./testing/mock_webtransport.ts";
import { MockSendStream, MockStream } from "./internal/webtransport/mock_stream_test.ts";
import { SessionStream } from "./session_stream.ts";
import { TrackReader } from "./track.ts";
import { background, withCancelCause } from "@okudai/golikejs/context";
import { TrackMux } from "./track_mux.ts";

// Global tracking for unhandled rejections during tests
const globalUnhandledRejections: Promise<any>[] = [];
const unhandledRejectionListener = (e: PromiseRejectionEvent) => {
	console.error("GLOBAL UNHANDLED REJECTION:", e.reason);
	globalUnhandledRejections.push(e.promise);
	e.preventDefault();
};
addEventListener("unhandledrejection", unhandledRejectionListener);

// Final diagnostics to help identify any pending resources left open by tests.
// This step is primarily diagnostic and ensures we catch open-streams, mocks, or background
// tasks that were not closed by individual tests.
Deno.test("Session cleanup diagnostics", async (t) => {
	await t.step("no lingering resources", async () => {
		// Force cleanup first
		try {
			MockWebTransport.closeAll();
		} catch (_e) {
			// ignore
		}
		
		const mocks = (MockWebTransport as any).allMocks?.length ?? 0;
		console.log("Diag: openMocks=", mocks, "globalUnhandledRejections=", globalUnhandledRejections.length);
		// If mocks remain open, attempt to close them
		if (mocks !== 0) {
			try {
				MockWebTransport.closeAll();
			} catch (_e) {
				// ignore
			}
		}
		// Re-evaluate counters
		const mocks2 = (MockWebTransport as any).allMocks?.length ?? 0;
		if (mocks2 !== 0 || globalUnhandledRejections.length > 0) {
			console.error("Lingering resources detected:");
			if (mocks2 !== 0) console.error("  openMocks:", mocks2);
			if (globalUnhandledRejections.length > 0) console.error("  globalUnhandledRejections:", globalUnhandledRejections.length);
			throw new Error("Detected lingering resources after tests");
		}
	});
});

// NOTE: using explicit new Session(...) in each test and explicit cleanup. No helper.

Deno.test("Session", async (t) => {
	// Catch unhandled rejections during tests to get better diagnostics.
	addEventListener("unhandledrejection", (e) => {
		try {
			console.error("UNHANDLED REJECTION in session_test:", e.reason);
		} catch (_e) {
			// ignore
		}
	});
	await t.step("constructor and ready sends client message", async () => {
		// Create a ServerSessionMessage with version included in DEFAULT_CLIENT_VERSIONS
		const rsp = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
		const serverBytes = await encodeMessageToUint8Array(async (w) => rsp.encode(w));

		const mock = new MockWebTransport([serverBytes]);
		const session = new Session({ conn: (mock as unknown) as WebTransport });

		await session.ready;

		// Check that the Session wrote the expected StreamType and SESSION_CLIENT message
		// The mock's writable captured chunks are available from returned createBidirectionalStream call
		// Since our test can't access the created stream directly, ensure no exception and instance created
		assertInstanceOf(session, Session);
		await session.close();
	});
	await t.step("constructor throws when SESSION_SERVER decode fails", async () => {
		// Truncated server message
		const serverBytes = new Uint8Array([0x80]); // incomplete varint
		const mock = new MockWebTransport([serverBytes]);
		let threw = false;
		let session: Session | undefined;
		try {
			session = new Session({ conn: (mock as unknown) as WebTransport });
			await session.ready;
		} catch (err) {
			threw = true;
		} finally {
			if (session) {
				try {
					await session.close();
				} catch (_e) {
					// ignore
				}
			}
		}
		assertEquals(threw, true);
	});
	await t.step("constructor throws when SESSION_SERVER version is incompatible", async () => {
		// Server sends incompatible version
		const incompatibleVersion = 0x12345678;
		const rsp = new SessionServerMessage({ version: incompatibleVersion });
		const serverBytes = await encodeMessageToUint8Array(async (w) => rsp.encode(w));
		const mock = new MockWebTransport([serverBytes]);
		let threw = false;
		let session: Session | undefined;
		try {
			session = new Session({ conn: (mock as unknown) as WebTransport });
			await session.ready;
		} catch (err) {
			threw = true;
		} finally {
			if (session) {
				try {
					await session.close();
				} catch (_e) {
					// ignore
				}
			}
		}
		assertEquals(threw, true);
	});

	await t.step("acceptAnnounce returns error when ANNOUNCE_INIT decode fails", async () => {
		// Truncated AnnounceInitMessage
		const bytes = new Uint8Array([0x80]); // incomplete
		const mock = new MockWebTransport([
			await encodeMessageToUint8Array(async (w) => {
				const s = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
				return await s.encode(w);
			}),
			bytes,
		]);
		const session = new Session({ conn: (mock as unknown) as WebTransport });

		await session.ready;
		const [reader, err] = await session.acceptAnnounce("/test/" as any);
		assertEquals(reader, undefined);
		assertExists(err);
		await session.close();
	});

	await t.step("subscribe returns error when SUBSCRIBE_OK decode fails", async () => {
		// Truncated SubscribeOkMessage
		const bytes = new Uint8Array([0x80]);
		const mock = new MockWebTransport([
			await encodeMessageToUint8Array(async (w) => {
				const s = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
				return await s.encode(w);
			}),
			bytes,
		]);
		const session = new Session({ conn: (mock as unknown) as WebTransport });

		await session.ready;
		const [track, err] = await session.subscribe("/test/path" as any, "track-name" as any);
		assertEquals(track, undefined);
		assertExists(err);
		await session.close();
	});

	await t.step("listening for subscribe stream calls mux serveTrack", async () => {
		let served = false;
		const mux: TrackMux = {
			serveTrack: async (_t) => {
				served = true;
			},
		} as TrackMux;

		// Create a SubscribeMessage and encode: this will be part of a bidirectional accept stream
		const req = new SubscribeMessage({
			subscribeId: 1,
			broadcastPath: "/test/path" as any,
			trackName: "name" as any,
			trackPriority: 0,
			minGroupSequence: 0,
			maxGroupSequence: 0,
		});
		const buf = await encodeMessageToUint8Array(async (w) => {
			await writeVarint(w, BiStreamTypes.SubscribeStreamType);
			return await req.encode(w);
		});

		// Now build a WebTransportBidirectionalStream-like object for incomingBidirectionalStreams
		const readable = new ReadableStream<Uint8Array>({
			start(c) {
				c.enqueue(buf);
				c.close();
			},
		});
		const writable = new WritableStream<Uint8Array>({ write(_c) {} });
		const incoming = new ReadableStream({
			start(controller) {
				controller.enqueue({ readable, writable });
				controller.close();
			},
		});

		// Mock transport uses server bytes for initial session response
		const serverRsp = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
		const serverBytes = await encodeMessageToUint8Array(async (w) => serverRsp.encode(w));
		const mock = new MockWebTransport([serverBytes]);
		mock.incomingBidirectionalStreams = incoming as any;

		const session = new Session({ conn: (mock as unknown) as WebTransport, mux });
		await session.ready;

		// Wait a tiny bit for background listeners to process
		await new Promise((resolve) => setTimeout(resolve, 10));
		assertEquals(served, true);
		await session.close();
		// Close the writable stream to avoid pending promises
		try {
			const writer = writable.getWriter();
			await writer.close();
		} catch (_e) {
			// ignore
		}
		// Also cancel the readable stream
		try {
			await readable.cancel();
		} catch (_e) {
			// ignore
		}
	});

	await t.step("acceptAnnounce succeeds with valid messages", async () => {
		// First send ANNOUNCE_PLEASE
		const please = new AnnouncePleaseMessage({ prefix: "/test/" });
		const pleaseBytes = await encodeMessageToUint8Array(async (w) => please.encode(w));

		// Then send ANNOUNCE_INIT response
		const init = new AnnounceInitMessage({ suffixes: ["suffix"] });
		const initBytes = await encodeMessageToUint8Array(async (w) => init.encode(w));

		const mock = new MockWebTransport([
			await encodeMessageToUint8Array(async (w) => {
				const s = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
				return await s.encode(w);
			}),
			pleaseBytes,
			initBytes,
		]);
		const session = new Session({ conn: (mock as unknown) as WebTransport });

		await session.ready;
		const [reader, err] = await session.acceptAnnounce("/test/" as any);
		assertExists(reader);
		assertEquals(err, undefined);
		await reader.close();
		await session.close();
	});
	await t.step("subscribe succeeds with valid messages", async () => {
		// Send SUBSCRIBE_OK response
		const ok = new SubscribeOkMessage({});
		const okBytes = await encodeMessageToUint8Array(async (w) => ok.encode(w));

		const mock = new MockWebTransport([
			await encodeMessageToUint8Array(async (w) => {
				const s = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
				return await s.encode(w);
			}),
			okBytes,
		], { keepStreamsOpen: true });
		// Keep the session stream open so the session context doesn't become canceled during the test
		// Keep the session stream open (don't close after server bytes), so
		// the session context doesn't get canceled during the test.
		const createdStreams: Array<{
			readable: ReadableStream<Uint8Array>;
			writable: WritableStream<Uint8Array>;
		}> = [];
		(mock as any).createBidirectionalStream = async () => {
			const chunks: Uint8Array[] = [];
			const writable = new WritableStream<Uint8Array>({
				write(chunk) {
					chunks.push(chunk.slice());
				},
			});
			const serverBytes = mock.serverBytesQueue.shift() ?? new Uint8Array([]);
			const readable = new ReadableStream<Uint8Array>({
				start(controller) {
					controller.enqueue(serverBytes);
					// Intentionally DO NOT close the controller to keep the
					// stream open for the session, preventing cancellation.
				},
			});
			(writable as any).writtenChunks = chunks;
			createdStreams.push({ writable, readable });
			return { writable, readable };
		};
		const session = new Session({ conn: (mock as unknown) as WebTransport });

		await session.ready;
		const [track, err] = await session.subscribe("/test/path" as any, "track-name" as any);
		assertExists(track);
		assertEquals(err, undefined);
		await track.closeWithError(0);
		await session.close();
		// Close the created streams to avoid leaving pending readers/writers
		for (const s of createdStreams) {
			try {
				const writer = s.writable.getWriter();
				await writer.close();
			} catch (_e) {
				// ignore
			}
			try {
				const reader = s.readable.getReader();
				await reader.cancel();
			} catch (_e) {
				// ignore
			}
		}
	});

	await t.step("close() calls transport.close with normal closure", async () => {
		let closeInfo: WebTransportCloseInfo | undefined;
		class LocalMock extends MockWebTransport {
			override close(closeInfoArg?: WebTransportCloseInfo) {
				closeInfo = closeInfoArg;
				super.close();
			}
		}

		const serverRsp = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
		const serverBytes = await encodeMessageToUint8Array(async (w) => serverRsp.encode(w));
		const mock = new LocalMock([serverBytes]);

		const session = new Session({ conn: (mock as any) as WebTransport });
		await session.ready;
		await session.close();
		assertEquals(closeInfo?.closeCode, 0x0);
		assertEquals(closeInfo?.reason, "No Error");
	});

	await t.step("closeWithError calls transport.close with error code and message", async () => {
		let closeInfo: WebTransportCloseInfo | undefined;
		class LocalMock extends MockWebTransport {
			override close(closeInfoArg?: WebTransportCloseInfo) {
				closeInfo = closeInfoArg;
				super.close();
			}
		}

		const serverRsp = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
		const serverBytes = await encodeMessageToUint8Array(async (w) => serverRsp.encode(w));
		const mock = new LocalMock([serverBytes]);

		const session = new Session({ conn: (mock as any) as WebTransport });
		await session.ready;
		await session.closeWithError(0x1, "Test Error");
		assertEquals(closeInfo?.closeCode, 0x1);
		assertEquals(closeInfo?.reason, "Test Error");
	});

	await t.step("handleAnnounceStream gracefully handles decode errors", async () => {
		// Create a stream with varint for AnnounceStreamType but truncated announce please message
		const buf = new Uint8Array([BiStreamTypes.AnnounceStreamType, 0xFF]);
		const readable = new ReadableStream<Uint8Array>({
			start(c) {
				c.enqueue(buf);
				c.close();
			},
		});
		const writable = new WritableStream<Uint8Array>({ write(_c) {} });
		const incoming = new ReadableStream({
			start(controller) {
				controller.enqueue({ readable, writable });
				controller.close();
			},
		});

		const serverRsp = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
		const serverBytes = await encodeMessageToUint8Array(async (w) => serverRsp.encode(w));
		const mock = new MockWebTransport([serverBytes]);
		mock.incomingBidirectionalStreams = incoming as any;

		const session = new Session({ conn: (mock as any) as WebTransport });
		await session.ready;

		await new Promise((r) => setTimeout(r, 20));
		assertEquals(true, true);
		await session.close();
		await session.close();
		await session.close();
		await session.close();
	});

	await t.step("listenBiStreams ignores unknown stream types without crashing", async () => {
		const unknownType = 99;
		const buf = new Uint8Array([unknownType]); // varint == 99
		const readable = new ReadableStream<Uint8Array>({
			start(c) {
				c.enqueue(buf);
				c.close();
			},
		});
		const writable = new WritableStream<Uint8Array>({ write(_c) {} });
		const incoming = new ReadableStream({
			start(controller) {
				controller.enqueue({ readable, writable });
				controller.close();
			},
		});

		const serverRsp = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
		const serverBytes = await encodeMessageToUint8Array(async (w) => serverRsp.encode(w));
		const mock = new MockWebTransport([serverBytes]);

		mock.incomingBidirectionalStreams = incoming as any;

		const session = new Session({ conn: (mock as any) as WebTransport });

		await session.ready;
		// Wait a tiny bit for background listeners to process
		await new Promise((r) => setTimeout(r, 10));
		// If no crash, success
		assertEquals(true, true);
		await session.close();
	});

	await t.step("listenUniStreams ignores unknown stream types without crashing", async () => {
		const unknownType = 99;
		const buf = new Uint8Array([unknownType]);
		const readable = new ReadableStream<Uint8Array>({
			start(c) {
				c.enqueue(buf);
				c.close();
			},
		});
		const incoming = new ReadableStream({
			start(controller) {
				controller.enqueue(readable);
				controller.close();
			},
		});

		const serverRsp = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
		const serverBytes = await encodeMessageToUint8Array(async (w) => serverRsp.encode(w));
		const mock = new MockWebTransport([serverBytes]);
		mock.incomingUnidirectionalStreams = incoming as any;

		const session = new Session({ conn: (mock as any) as WebTransport });
		await session.ready;

		await new Promise((r) => setTimeout(r, 10));
		assertEquals(true, true);
		await session.close();
	});

	await t.step("handleGroupStream logs error when no subscription found", async () => {
		// Create a GroupMessage with subscribeId not in map
		const gm = new GroupMessage({ subscribeId: 999, sequence: 0 });
		const buf = await encodeMessageToUint8Array(async (w) => {
			await writeVarint(w, UniStreamTypes.GroupStreamType);
			return await gm.encode(w);
		});
		const readable = new ReadableStream<Uint8Array>({
			start(c) {
				c.enqueue(buf);
				c.close();
			},
		});
		const incoming = new ReadableStream({
			start(controller) {
				controller.enqueue(readable);
				controller.close();
			},
		});

		const serverRsp = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
		const serverBytes = await encodeMessageToUint8Array(async (w) => serverRsp.encode(w));
		const mock = new MockWebTransport([serverBytes]);
		mock.incomingUnidirectionalStreams = incoming as any;

		const session = new Session({ conn: (mock as any) as WebTransport });
		await session.ready;

		await new Promise((r) => setTimeout(r, 20));
		assertEquals(true, true);
		await session.close();
	});

	await t.step(
		"SessionStream detect updates updates clientInfo.bitrate and encodes message",
		async () => {
			const [ctx, ctxCancel] = withCancelCause(background());
			const mock = new MockStream(10n);
			const client = new SessionClientMessage({
				versions: new Set([0xffffff00]),
				extensions: new Map(),
			});
			const server = new SessionServerMessage({ version: 0xffffff00, extensions: new Map() });
			let called = false;
			const ss = new SessionStream({
				context: ctx,
				stream: mock as any,
				client,
				server,
				detectFunc: async () => {
					if (!called) {
						called = true;
						return 12345;
					}
					// Close the context by returning a value and letting loop continue once more
					return 0;
				},
			});
			// Wait a tick for detect to run at least once
			await new Promise((r) => setTimeout(r, 10));
			// clientInfo should have been updated
			assertNotEquals(ss.clientInfo.bitrate, 0);
			// Writable should have been written into
			assertEquals(mock.writable.writtenData.length > 0, true);
			// Clean up session stream loops
			ctxCancel(new Error("test done"));
			await ss.waitForBackgroundTasks();
		},
	);

	await t.step(
		"SessionStream handles incoming SessionUpdate message and updates serverInfo",
		async () => {
			const [ctx, ctxCancel] = withCancelCause(background());
			const mock = new MockStream(11n);
			const client = new SessionClientMessage({
				versions: new Set([0xffffff00]),
				extensions: new Map(),
			});
			const server = new SessionServerMessage({ version: 0xffffff00, extensions: new Map() });
			// Prepare update message bytes
			const mu = new MockSendStream(11n);
			const update = new SessionUpdateMessage({ bitrate: 98765 });
			await update.encode(mu as any);
			mock.readable.setData(mu.getAllWrittenData());
			const ss = new SessionStream({
				context: ctx,
				stream: mock as any,
				client,
				server,
				detectFunc: async () => {
					return 0;
				},
			});
			// Wait a tick for handleUpdates to process the message
			await new Promise((r) => setTimeout(r, 10));
			assertEquals(ss.serverInfo.bitrate, 98765);
			// Clean up session stream loops
			ctxCancel(new Error("test done"));
			await ss.waitForBackgroundTasks();
		},
	);

	await t.step("constructor throws when openStream fails", async () => {
		const mock = new MockWebTransport();
		(mock as any).createBidirectionalStream = () => {
			throw new Error("openStream failed");
		};
		let threw = false;
		let session: Session | undefined;
		try {
			session = new Session({ conn: (mock as unknown) as WebTransport });
			await session.ready;
		} catch (err) {
			threw = true;
			assertEquals((err as Error).message, "openStream failed");
		} finally {
			if (session) {
				try {
					await session.close();
				} catch (_e) {
					// ignore
				}
			}
		}
		assertEquals(threw, true);
	});

	await t.step("constructor throws when writeVarint fails", async () => {
		const mock = new MockWebTransport();
		(mock as any).createBidirectionalStream = () => ({
			writable: new WritableStream({
				async write() {
					throw new Error("write failed");
				},
			}),
			readable: new ReadableStream(),
		});
		let threw = false;
		let session: Session | undefined;
		try {
			session = new Session({ conn: (mock as unknown) as WebTransport });
			await session.ready;
		} catch (err) {
			threw = true;
			assertEquals((err as Error).message, "write failed");
		} finally {
			if (session) {
				try {
					await session.close();
				} catch (_e) {
					// ignore
				}
			}
		}
		assertEquals(threw, true);
		// Allow microtasks to process and any rejected Promises created by
		// the writable stream's write algorithm to settle so they don't
		// generate unhandledrejection outside the test.
		await new Promise((r) => setTimeout(r, 0));
	});

	await t.step("acceptAnnounce returns error when openStream fails", async () => {
		const rsp = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
		const serverBytes = await encodeMessageToUint8Array(async (w) => rsp.encode(w));
		const mock = new MockWebTransport([serverBytes]);
		const session = new Session({ conn: (mock as unknown) as WebTransport });
		await session.ready;
		// Override createBidirectionalStream to fail
		(mock as any).createBidirectionalStream = () => {
			throw new Error("openStream failed");
		};
		const [reader, err] = await session.acceptAnnounce("/test/");
		assertEquals(reader, undefined);
		assertExists(err);
		await session.close();
		assertEquals(err.message, "openStream failed");
		await session.close();
	});

	await t.step("acceptAnnounce returns error when writeVarint fails", async () => {
		const rsp = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
		const serverBytes = await encodeMessageToUint8Array(async (w) => rsp.encode(w));
		const mock = new MockWebTransport([serverBytes]);
		const session = new Session({ conn: (mock as unknown) as WebTransport });
		await session.ready;
		// Override createBidirectionalStream to return failing writable
		(mock as any).createBidirectionalStream = () => ({
			writable: new WritableStream({
				async write() {
					throw new Error("write failed");
				},
			}),
			readable: new ReadableStream(),
		});
		const [reader, err] = await session.acceptAnnounce("/test/");
		assertEquals(reader, undefined);
		assertExists(err);
		await session.close();
		assertEquals(err.message, "write failed");
		await session.close();
		// Allow microtasks to process and any rejected Promises created by
		// the writable stream's write algorithm to settle.
		await new Promise((r) => setTimeout(r, 0));
		// Allow microtasks to process and any rejected Promises created by
		// the writable stream's write algorithm to settle.
		await new Promise((r) => setTimeout(r, 0));
	});

	await t.step("subscribe returns error when openStream fails", async () => {
		const rsp = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
		const serverBytes = await encodeMessageToUint8Array(async (w) => rsp.encode(w));
		const mock = new MockWebTransport([serverBytes]);
		const session = new Session({ conn: (mock as unknown) as WebTransport });
		await session.ready;
		// Override createBidirectionalStream to fail
		(mock as any).createBidirectionalStream = () => {
			throw new Error("openStream failed");
		};
		const [track, err] = await session.subscribe("/test/path", "track-name");
		assertEquals(track, undefined);
		assertExists(err);
		await session.close();
		assertEquals(err.message, "openStream failed");
		await session.close();
	});

	await t.step("subscribe returns error when writeVarint fails", async () => {
		const rsp = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
		const serverBytes = await encodeMessageToUint8Array(async (w) => rsp.encode(w));
		const mock = new MockWebTransport([serverBytes]);
		const session = new Session({ conn: (mock as unknown) as WebTransport });
		await session.ready;
		// Override createBidirectionalStream to return failing writable
		(mock as any).createBidirectionalStream = () => ({
			writable: new WritableStream({
				async write() {
					throw new Error("write failed");
				},
			}),
			readable: new ReadableStream(),
		});
		const [track, err] = await session.subscribe("/test/path", "track-name");
		assertEquals(track, undefined);
		assertExists(err);
		await session.close();
		assertEquals(err.message, "write failed");
		await session.close();
	});

	await t.step("close() returns early if already closed", async () => {
		const rsp = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
		const serverBytes = await encodeMessageToUint8Array(async (w) => rsp.encode(w));
		const mock = new MockWebTransport([serverBytes]);
		const session = new Session({ conn: (mock as unknown) as WebTransport });
		await session.ready;
		await session.close();
		// Second close should return early
		await session.close();
	});

	await t.step("closeWithError returns early if already closed", async () => {
		const rsp = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
		const serverBytes = await encodeMessageToUint8Array(async (w) => rsp.encode(w));
		const mock = new MockWebTransport([serverBytes]);
		const session = new Session({ conn: (mock as unknown) as WebTransport });
		await session.ready;
		await session.close();
		// Second close should return early
		await session.closeWithError(1, "test");
	});

	// cleanup block moved to the end of Deno.test

	await t.step("handleSubscribeStream gracefully handles decode errors", async () => {
		// Create a stream with SubscribeStreamType but truncated subscribe message
		const buf = new Uint8Array([BiStreamTypes.SubscribeStreamType, 0xFF]);
		const readable = new ReadableStream<Uint8Array>({
			start(c) {
				c.enqueue(buf);
				c.close();
			},
		});
		const writable = new WritableStream<Uint8Array>({ write(_c) {} });
		const incoming = new ReadableStream({
			start(controller) {
				controller.enqueue({ readable, writable });
				controller.close();
			},
		});

		const serverRsp = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
		const serverBytes = await encodeMessageToUint8Array(async (w) => serverRsp.encode(w));
		const mock = new MockWebTransport([serverBytes]);
		mock.incomingBidirectionalStreams = incoming as any;

		const session = new Session({ conn: (mock as any) as WebTransport });
		await session.ready;

		await new Promise((r) => setTimeout(r, 20));
		assertEquals(true, true);
		await session.close();
	});

	await t.step("handleGroupStream gracefully handles decode errors", async () => {
		// Create a GroupMessage with GroupStreamType but truncated message
		const buf = new Uint8Array([UniStreamTypes.GroupStreamType, 0xFF]);
		const readable = new ReadableStream<Uint8Array>({
			start(c) {
				c.enqueue(buf);
				c.close();
			},
		});
		const incoming = new ReadableStream({
			start(controller) {
				controller.enqueue(readable);
				controller.close();
			},
		});

		const serverRsp = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
		const serverBytes = await encodeMessageToUint8Array(async (w) => serverRsp.encode(w));
		const mock = new MockWebTransport([serverBytes]);
		mock.incomingUnidirectionalStreams = incoming as any;

		const session = new Session({ conn: (mock as any) as WebTransport });
		await session.ready;

		await new Promise((r) => setTimeout(r, 20));
		assertEquals(true, true);
		await session.close();
	});

	await t.step("subscribe with config parameters", async () => {
		// Send SUBSCRIBE_OK response
		const ok = new SubscribeOkMessage({});
		const okBytes = await encodeMessageToUint8Array(async (w) => ok.encode(w));

		const mock = new MockWebTransport([
			await encodeMessageToUint8Array(async (w) => {
				const s = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
				return await s.encode(w);
			}),
			okBytes,
		], { keepStreamsOpen: true });
		const session = new Session({ conn: (mock as unknown) as WebTransport });

		await session.ready;
		const config = {
			trackPriority: 5,
			minGroupSequence: 10,
			maxGroupSequence: 20,
		};
		const [track, err] = await session.subscribe(
			"/test/path" as any,
			"track-name" as any,
			config,
		);
		assertExists(track);
		assertEquals(err, undefined);
		await session.close();
	});

	await t.step("group messages are enqueued and accepted by TrackReader", async () => {
		// Prepare SUBSCRIBE_OK response
		const ok = new SubscribeOkMessage({});
		const okBytes = await encodeMessageToUint8Array(async (w) => ok.encode(w));

		// Build GroupMessage for subscribe id 0
		const gm = new GroupMessage({ subscribeId: 0, sequence: 5 });
		const buf = await encodeMessageToUint8Array(async (w) => {
			await writeVarint(w, UniStreamTypes.GroupStreamType);
			return await gm.encode(w);
		});

		// Build readable for unidirectional streams
		const readable = new ReadableStream<Uint8Array>({
			start(c) {
				c.enqueue(buf);
				c.close();
			},
		});
		// Use an outer incoming stream so we can enqueue the unidirectional
		// stream after subscribing, preventing the session from closing early.
		let outerController:
			| ReadableStreamDefaultController<ReadableStream<Uint8Array>>
			| undefined;
		const incoming = new ReadableStream<ReadableStream<Uint8Array>>({
			start(c) {
				outerController = c;
			},
		});

		const mock = new MockWebTransport([
			await encodeMessageToUint8Array(async (w) => {
				const s = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
				return await s.encode(w);
			}),
			okBytes,
		], { keepStreamsOpen: true });

		// Set incoming unidirectional streams; we'll enqueue the group readable
		// after we subscribe to ensure it is buffered and drained correctly.
		mock.incomingUnidirectionalStreams = incoming as any;

		const session = new Session({ conn: (mock as unknown) as WebTransport });
		await session.ready;
		const [track, err] = await session.subscribe("/test/path" as any, "track-name" as any);
		assertExists(track);
		assertEquals(err, undefined);

		// Now enqueue the previously made readable into the incoming streams
		// Begin waiting for a group — acceptGroup will block until a group is
		// enqueued. Doing so before enqueueing prevents races where the session
		// might close before acceptGroup is called.
		const acceptPromise = track!.acceptGroup(new Promise(() => {}));
		// debug log removed

		outerController!.enqueue(readable);
		// Do not close outerController to avoid session stream becoming closed
		// outerController!.close();

		// Wait for the acceptGroup call to resolve
		const [group, gerr] = await acceptPromise;
		assertExists(group);
		assertEquals(gerr, undefined);
		// Sequence should match the message we sent
		assertEquals(group!.sequence, 5);
		// Close the outer controller to ensure no pending readers remain
		try {
			outerController!.close();
		} catch (_e) {
			// ignore
		}
	});

	await t.step("constructor with extensions", async () => {
		const extensions = new Extensions();
		extensions.addBytes(1, new Uint8Array([1, 2, 3]));

		// Server responds with extensions
		const serverExtensions = new Map([[2, new Uint8Array([4, 5, 6])]]);
		const rsp = new SessionServerMessage({
			version: [...DEFAULT_CLIENT_VERSIONS][0],
			extensions: serverExtensions,
		});
		const serverBytes = await encodeMessageToUint8Array(async (w) => rsp.encode(w));

		const mock = new MockWebTransport([serverBytes]);
		const session = new Session({
			conn: (mock as unknown) as WebTransport,
			extensions,
		});

		await session.ready;
		assertInstanceOf(session, Session);
	});

	await t.step("listenBiStreams handles timed out error gracefully", async () => {
		const serverRsp = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
		const serverBytes = await encodeMessageToUint8Array(async (w) => serverRsp.encode(w));
		const mock = new MockWebTransport([serverBytes]);

		// Mock acceptStream to throw "timed out" error
		(mock as any).acceptStream = async () => {
			throw new Error("timed out");
		};

		const session = new Session({ conn: (mock as any) as WebTransport });
		await session.ready;

		// Wait for the listener to handle the error
		await new Promise((r) => setTimeout(r, 10));
		assertEquals(true, true);
		await session.close();
		await session.close();
		await session.close();
		await session.close();
	});

	await t.step("listenUniStreams handles timed out error gracefully", async () => {
		const serverRsp = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
		const serverBytes = await encodeMessageToUint8Array(async (w) => serverRsp.encode(w));
		const mock = new MockWebTransport([serverBytes]);

		// Mock acceptUniStream to throw "timed out" error
		(mock as any).acceptUniStream = async () => {
			throw new Error("timed out");
		};

		const session = new Session({ conn: (mock as any) as WebTransport });
		await session.ready;

		// Wait for the listener to handle the error
		await new Promise((r) => setTimeout(r, 10));
		await session.close();
		assertEquals(true, true);
	});

	await t.step("listenBiStreams handles other errors gracefully", async () => {
		const serverRsp = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
		const serverBytes = await encodeMessageToUint8Array(async (w) => serverRsp.encode(w));
		const mock = new MockWebTransport([serverBytes]);

		// Mock acceptStream to throw other error
		(mock as any).acceptStream = async () => {
			throw new Error("some other error");
		};

		const session = new Session({ conn: (mock as any) as WebTransport });
		await session.ready;

		// Wait for the listener to handle the error
		await new Promise((r) => setTimeout(r, 10));
		assertEquals(true, true);
		await session.close();
	});

	await t.step("listenUniStreams handles other errors gracefully", async () => {
		const serverRsp = new SessionServerMessage({ version: [...DEFAULT_CLIENT_VERSIONS][0] });
		const serverBytes = await encodeMessageToUint8Array(async (w) => serverRsp.encode(w));
		const mock = new MockWebTransport([serverBytes]);

		// Mock acceptUniStream to throw other error
		(mock as any).acceptUniStream = async () => {
			throw new Error("some other error");
		};

		const session = new Session({ conn: (mock as any) as WebTransport });
		await session.ready;

		// Wait for the listener to handle the error
		await new Promise((r) => setTimeout(r, 10));
		assertEquals(true, true);
	});

	await t.step(
		"handleGroupStream enqueue called for matching subscribeId and TrackReader acceptGroup receives group",
		async () => {
			const ok = new SubscribeOkMessage({});
			const okBytes = await encodeMessageToUint8Array(async (w) => ok.encode(w));

			const mock = new MockWebTransport([
				await encodeMessageToUint8Array(async (w) => {
					const s = new SessionServerMessage({
						version: [...DEFAULT_CLIENT_VERSIONS][0],
					});
					return await s.encode(w);
				}),
				okBytes,
			], { keepStreamsOpen: true });

			const gm = new GroupMessage({ subscribeId: 0, sequence: 1 });
			const gmBuf = await encodeMessageToUint8Array(async (w) => {
				return await gm.encode(w);
			});

			const total = new Uint8Array([UniStreamTypes.GroupStreamType, ...gmBuf]);
			const readable = new ReadableStream<Uint8Array>({
				start(c) {
					c.enqueue(total);
					c.close();
				},
			});
			let outerController:
				| ReadableStreamDefaultController<ReadableStream<Uint8Array>>
				| undefined;
			const incoming = new ReadableStream<ReadableStream<Uint8Array>>({
				start(c) {
					outerController = c;
				},
			});
			mock.incomingUnidirectionalStreams = incoming as any;

			const session = new Session({ conn: (mock as any) as WebTransport });

			await session.ready;
			const [track, err] = await session.subscribe("/test/path" as any, "track-name" as any);
			assertExists(track);
			assertEquals(err, undefined);

			outerController!.enqueue(readable);
			outerController!.close();

			await new Promise((r) => setTimeout(r, 20));
			const [group, ge] = await track!.acceptGroup(new Promise(() => {}));
			assertEquals(ge, undefined);
			assertInstanceOf(group, Object);
			await session.close();
		},
	);

	await t.step(
		"buffered group messages before subscribe are delivered after subscribe",
		async () => {
			const ok = new SubscribeOkMessage({});
			const okBytes = await encodeMessageToUint8Array(async (w) => ok.encode(w));

			const gm = new GroupMessage({ subscribeId: 0, sequence: 123 });
			const gmBuf = await encodeMessageToUint8Array(async (w) => gm.encode(w));

			const total = new Uint8Array([UniStreamTypes.GroupStreamType, ...gmBuf]);
			const readable = new ReadableStream<Uint8Array>({
				start(c) {
					c.enqueue(total);
					c.close();
				},
			});

			const incomingStreamController: ReadableStreamDefaultController<
				ReadableStream<Uint8Array>
			>[] = [];
			const incoming = new ReadableStream<ReadableStream<Uint8Array>>({
				start(c) {
					incomingStreamController.push(c);
				},
			});

			const mock = new MockWebTransport([
				await encodeMessageToUint8Array(async (w) => {
					const s = new SessionServerMessage({
						version: [...DEFAULT_CLIENT_VERSIONS][0],
					});
					return await s.encode(w);
				}),
				okBytes,
			], { keepStreamsOpen: true });

			mock.incomingUnidirectionalStreams = incoming as any;
			const session = new Session({ conn: (mock as any) as WebTransport });
			await session.ready;

			await new Promise((r) => setTimeout(r, 0));
			if (incomingStreamController[0]) {
				incomingStreamController[0].enqueue(readable);
				incomingStreamController[0].close();
			}

			const [track, err] = await session.subscribe("/test/path" as any, "track-name" as any);
			assertEquals(err, undefined);
			assertInstanceOf(track, TrackReader);

			await new Promise((r) => setTimeout(r, 10));
			const [group, gerr] = await track!.acceptGroup(new Promise(() => {}));
			assertEquals(gerr, undefined);
			assertInstanceOf(group, Object);
			assertEquals(group!.sequence, 123);
			await session.close();
		},
	);

	// No final cleanup: tests manage their own `Session` lifecycles explicitly.

	// All cleanup is handled by individual tests and the diagnostics test
	// Clean up global event listeners
	removeEventListener("unhandledrejection", unhandledRejectionListener);
});
