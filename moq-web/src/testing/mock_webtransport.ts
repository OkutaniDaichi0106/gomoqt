import { SendStream } from "../internal/webtransport/mod.ts";
// SessionServerMessage import not needed here

export async function encodeMessageToUint8Array(
	encodeFn: (writer: SendStream) => Promise<Error | undefined>,
): Promise<Uint8Array> {
	const chunks: Uint8Array[] = [];
	const writable = new WritableStream<Uint8Array>({
		write(chunk) {
			chunks.push(chunk.slice());
		},
	});
	const writer = new SendStream({ stream: writable, streamId: 0n });
	const err = await encodeFn(writer);
	if (err) throw err;
	const total = chunks.reduce((s, c) => s + c.length, 0);
	const combined = new Uint8Array(total);
	let off = 0;
	for (const c of chunks) {
		combined.set(c, off);
		off += c.length;
	}
	return combined;
}

export class MockWebTransport implements WebTransport {
	ready: Promise<void>;
	closed: Promise<WebTransportCloseInfo>;
	#closeResolve?: (info: WebTransportCloseInfo) => void;
	// Minimal datagrams implementation to satisfy WebTransport type
	// Use `any` to avoid having to implement the full DOM WebTransportDatagramDuplexStream
	datagrams: any;
	private _incomingBidirectionalStreams: ReadableStream<WebTransportBidirectionalStream>;
	private _incomingUnidirectionalStreams: ReadableStream<ReadableStream<Uint8Array>>;
	private _assignedIncomingBidirectionalStreams: ReadableStream<
		WebTransportBidirectionalStream
	>[] = [];
	private _assignedIncomingUnidirectionalStreams: ReadableStream<ReadableStream<Uint8Array>>[] =
		[];
	serverBytesQueue: Uint8Array[] = [];
	keepStreamsOpen: boolean = false;
	#biController?: ReadableStreamDefaultController<WebTransportBidirectionalStream>;
	#uniController?: ReadableStreamDefaultController<ReadableStream<Uint8Array>>;

	static allMocks: MockWebTransport[] = [];

	constructor(serverSessionBytesQueue: Uint8Array[] = [], opts?: { keepStreamsOpen?: boolean }) {
		MockWebTransport.allMocks.push(this);
		// Minimal datagram implementation
		this.datagrams = {
			incoming: new ReadableStream<Uint8Array>({
				start() {},
			}),
			outgoing: {
				async write(_b: Uint8Array) {},
			},
		} as any;
		this.ready = Promise.resolve();
		this.closed = new Promise((resolve) => {
			this.#closeResolve = resolve;
		});
		this.serverBytesQueue = serverSessionBytesQueue.slice();
		this.keepStreamsOpen = opts?.keepStreamsOpen ?? false;
		this._incomingBidirectionalStreams = new ReadableStream<WebTransportBidirectionalStream>({
			start: (controller) => {
				this.#biController = controller;
			},
		});
		this._incomingUnidirectionalStreams = new ReadableStream<ReadableStream<Uint8Array>>({
			start: (controller) => {
				this.#uniController = controller;
			},
		});
	}

	async createBidirectionalStream(): Promise<WebTransportBidirectionalStream> {
		const chunks: Uint8Array[] = [];
		const writable = new WritableStream<Uint8Array>({
			write(chunk) {
				chunks.push(chunk.slice());
			},
		});
		const serverBytes = this.serverBytesQueue.shift() ?? new Uint8Array([]);
		const readable = new ReadableStream<Uint8Array>({
			start: (controller) => {
				controller.enqueue(serverBytes);
				if (!this.keepStreamsOpen) {
					controller.close();
				}
			},
		});
		// Keep a test-only field for verifying written chunks
		(writable as unknown as { writtenChunks?: Uint8Array[] }).writtenChunks = chunks;
		return { writable, readable };
	}

	async createUnidirectionalStream(): Promise<WritableStream<Uint8Array>> {
		const writable = new WritableStream<Uint8Array>({ write(_c) {} });
		return writable;
	}

	close(_closeInfo?: WebTransportCloseInfo) {
		if (this.#closeResolve) {
			this.#closeResolve({ closeCode: _closeInfo?.closeCode, reason: _closeInfo?.reason });
		}
		if (this.#biController) this.#biController.close();
		if (this.#uniController) this.#uniController.close();
		// Remove from tracked mocks
		const idx = MockWebTransport.allMocks.indexOf(this);
		if (idx >= 0) MockWebTransport.allMocks.splice(idx, 1);
		// debug log
		try {
			console.log("MockWebTransport.close", this.serverBytesQueue.length);
		} catch (e) {}
	}

	static closeAll() {
		for (const m of MockWebTransport.allMocks.slice()) {
			try {
				m.close();
			} catch (_) {
				// ignore
			}
		}
		MockWebTransport.allMocks.length = 0;
	}

	// Accessors for assigned incoming streams so tests can set them without
	// losing the ability to cancel them on close().
	public get incomingBidirectionalStreams(): ReadableStream<WebTransportBidirectionalStream> {
		return this._incomingBidirectionalStreams;
	}

	public set incomingBidirectionalStreams(v: ReadableStream<WebTransportBidirectionalStream>) {
		// When user assigns a custom incoming stream, tee it so MockWebTransport
		// keeps a branch we can cancel on close() without interfering with the
		// SessionImpl's reader.
		try {
			const [a, b] = (v as any).tee();
			this._assignedIncomingBidirectionalStreams.push(
				a as ReadableStream<WebTransportBidirectionalStream>,
			);
			this._incomingBidirectionalStreams = b as ReadableStream<
				WebTransportBidirectionalStream
			>;
		} catch (_e) {
			// If tee() is not available, fall back to assigning directly so tests
			// still function, but we won't be able to cancel the assigned stream.
			this._incomingBidirectionalStreams = v;
			this._assignedIncomingBidirectionalStreams.push(v);
		}
	}

	public get incomingUnidirectionalStreams(): ReadableStream<ReadableStream<Uint8Array>> {
		return this._incomingUnidirectionalStreams;
	}

	public set incomingUnidirectionalStreams(v: ReadableStream<ReadableStream<Uint8Array>>) {
		try {
			const [a, b] = (v as any).tee();
			// a: our tracked branch, b: delivered to SessionImpl
			this._assignedIncomingUnidirectionalStreams.push(
				a as ReadableStream<ReadableStream<Uint8Array>>,
			);
			this._incomingUnidirectionalStreams = b as ReadableStream<ReadableStream<Uint8Array>>;
		} catch (_e) {
			this._incomingUnidirectionalStreams = v;
			this._assignedIncomingUnidirectionalStreams.push(v);
		}
	}
}
