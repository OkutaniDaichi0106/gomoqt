/**
 * Mock implementation of WebTransport Connection for testing purposes.
 * This file provides test doubles that can be used instead of real WebTransport connections.
 */

import { ReceiveStream, type ReceiveStreamInit } from "./reader.ts";
import { Stream, type StreamInit } from "./stream.ts";
import { SendStream, type SendStreamInit } from "./writer.ts";
import type { StreamError } from "./error.ts";

/**
 * Mock implementation of ReadableStreamDefaultReader for bidirectional streams.
 */
export class MockBidirectionalStreamReader
	implements ReadableStreamDefaultReader<WebTransportBidirectionalStream> {
	#queue: WebTransportBidirectionalStream[] = [];
	#closed = false;
	readonly closed: Promise<undefined>;
	#resolveClose!: (value: undefined) => void;

	constructor() {
		this.closed = new Promise((resolve) => {
			this.#resolveClose = resolve;
		});
	}

	/**
	 * Add a mock bidirectional stream to the queue.
	 */
	enqueue(stream: WebTransportBidirectionalStream): void {
		if (this.#closed) {
			throw new Error("Cannot enqueue to closed reader");
		}
		this.#queue.push(stream);
	}

	async read(): Promise<
		ReadableStreamReadResult<WebTransportBidirectionalStream>
	> {
		if (this.#queue.length > 0) {
			const value = this.#queue.shift()!;
			return { done: false, value };
		}
		if (this.#closed) {
			return { done: true, value: undefined };
		}
		// Wait for data or close
		await new Promise((resolve) => setTimeout(resolve, 0));
		return this.read();
	}

	releaseLock(): void {
		// No-op for mock
	}

	cancel(_reason?: unknown): Promise<void> {
		this.#closed = true;
		this.#resolveClose(undefined);
		return Promise.resolve();
	}

	/**
	 * Close the reader for testing purposes.
	 */
	close(): void {
		this.#closed = true;
		this.#resolveClose(undefined);
	}
}

/**
 * Mock implementation of ReadableStreamDefaultReader for unidirectional streams.
 */
export class MockUnidirectionalStreamReader
	implements
		ReadableStreamDefaultReader<ReadableStream<Uint8Array<ArrayBufferLike>>> {
	#queue: ReadableStream<Uint8Array>[] = [];
	#closed = false;
	readonly closed: Promise<undefined>;
	#resolveClose!: (value: undefined) => void;

	constructor() {
		this.closed = new Promise((resolve) => {
			this.#resolveClose = resolve;
		});
	}

	/**
	 * Add a mock unidirectional stream to the queue.
	 */
	enqueue(stream: ReadableStream<Uint8Array>): void {
		if (this.#closed) {
			throw new Error("Cannot enqueue to closed reader");
		}
		this.#queue.push(stream);
	}

	async read(): Promise<
		ReadableStreamReadResult<ReadableStream<Uint8Array<ArrayBufferLike>>>
	> {
		if (this.#queue.length > 0) {
			const value = this.#queue.shift()!;
			return { done: false, value };
		}
		if (this.#closed) {
			return { done: true, value: undefined };
		}
		// Wait for data or close
		await new Promise((resolve) => setTimeout(resolve, 0));
		return this.read();
	}

	releaseLock(): void {
		// No-op for mock
	}

	cancel(_reason?: unknown): Promise<void> {
		this.#closed = true;
		this.#resolveClose(undefined);
		return Promise.resolve();
	}

	/**
	 * Close the reader for testing purposes.
	 */
	close(): void {
		this.#closed = true;
		this.#resolveClose(undefined);
	}
}

/**
 * Mock WebTransport implementation for testing.
 */
export class MockWebTransport implements WebTransport {
	readonly closed: Promise<WebTransportCloseInfo>;
	readonly ready: Promise<void>;
	readonly datagrams!: WebTransportDatagramDuplexStream;
	readonly incomingBidirectionalStreams!: ReadableStream<WebTransportBidirectionalStream>;
	readonly incomingUnidirectionalStreams!: ReadableStream<
		ReadableStream<Uint8Array<ArrayBufferLike>>
	>;
	readonly congestionControl!: WebTransportCongestionControl;

	#biReader: MockBidirectionalStreamReader;
	#uniReader: MockUnidirectionalStreamReader;
	#closeInfo: WebTransportCloseInfo | null = null;
	#resolveReady!: () => void;
	#rejectReady!: (reason: Error) => void;
	#resolveClosed!: (value: WebTransportCloseInfo) => void;

	// Counter for generating stream IDs
	#clientBiCounter = 0;
	#clientUniCounter = 2;

	constructor() {
		this.ready = new Promise((resolve, reject) => {
			this.#resolveReady = resolve;
			this.#rejectReady = reject;
		});
		this.closed = new Promise((resolve) => {
			this.#resolveClosed = resolve;
		});

		this.#biReader = new MockBidirectionalStreamReader();
		this.#uniReader = new MockUnidirectionalStreamReader();

		// Create readable streams that use the mock readers
		this.incomingBidirectionalStreams = new ReadableStream({
			start: (controller) => {
				// Delegate to the mock reader
				this.#biReader.read().then((result) => {
					if (result.done) {
						controller.close();
					} else {
						controller.enqueue(result.value);
					}
				});
			},
		});

		this.incomingUnidirectionalStreams = new ReadableStream({
			start: (controller) => {
				// Delegate to the mock reader
				this.#uniReader.read().then((result) => {
					if (result.done) {
						controller.close();
					} else {
						controller.enqueue(result.value);
					}
				});
			},
		});
	}

	/**
	 * Mark the WebTransport as ready.
	 */
	markReady(): void {
		this.#resolveReady();
	}

	/**
	 * Reject the ready promise (for error testing).
	 */
	failReady(error: Error): void {
		this.#rejectReady(error);
	}

	/**
	 * Get the bidirectional stream reader for testing.
	 */
	getBiReader(): MockBidirectionalStreamReader {
		return this.#biReader;
	}

	/**
	 * Get the unidirectional stream reader for testing.
	 */
	getUniReader(): MockUnidirectionalStreamReader {
		return this.#uniReader;
	}

	async createBidirectionalStream(): Promise<WebTransportBidirectionalStream> {
		this.#clientBiCounter += 4;

		const { readable, writable } = new TransformStream<Uint8Array>();

		return {
			readable,
			writable,
		} as WebTransportBidirectionalStream;
	}

	async createUnidirectionalStream(): Promise<WritableStream<Uint8Array>> {
		this.#clientUniCounter += 4;
		const { writable } = new TransformStream<Uint8Array>();
		return writable;
	}

	close(closeInfo?: WebTransportCloseInfo): void {
		const info = closeInfo || { closeCode: 0, reason: "" };
		this.#closeInfo = info;
		this.#biReader.close();
		this.#uniReader.close();
		this.#resolveClosed(info);
	}

	/**
	 * Get the close info for testing.
	 */
	getCloseInfo(): WebTransportCloseInfo | null {
		return this.#closeInfo;
	}
}

/**
 * Mock Stream implementation for testing.
 */
export class MockStream implements Stream {
	readonly id: number;
	readonly writable: SendStream;
	readonly readable: ReceiveStream;

	constructor(init: StreamInit) {
		this.id = init.streamId;
		this.writable = new SendStream({
			stream: init.stream.writable,
			streamId: init.streamId,
		});
		this.readable = new ReceiveStream({
			stream: init.stream.readable,
			streamId: init.streamId,
		});
	}

	/**
	 * Create a mock stream with in-memory buffers.
	 */
	static create(streamId: number): MockStream {
		const { readable, writable } = new TransformStream<Uint8Array>();
		return new MockStream({
			streamId,
			stream: { readable, writable } as WebTransportBidirectionalStream,
		});
	}
}

/**
 * Mock SendStream implementation for testing.
 * This wraps a real SendStream for easier testing.
 */
export class MockSendStream {
	readonly id: number;
	#stream: SendStream;

	constructor(init: SendStreamInit) {
		this.id = init.streamId;
		this.#stream = new SendStream(init);
	}

	writeUint8(value: number): void {
		this.#stream.writeUint8(value);
	}

	writeUint8Array(data: Uint8Array): void {
		this.#stream.writeUint8Array(data);
	}

	writeString(str: string): void {
		this.#stream.writeString(str);
	}

	writeVarint(num: number): void {
		this.#stream.writeVarint(num);
	}

	writeBigVarint(num: bigint): void {
		this.#stream.writeBigVarint(num);
	}

	writeBoolean(value: boolean): void {
		this.#stream.writeBoolean(value);
	}

	writeStringArray(arr: string[]): void {
		this.#stream.writeStringArray(arr);
	}

	copyFrom(src: { byteLength: number; copyTo(target: ArrayBuffer | ArrayBufferView): void }): void {
		this.#stream.copyFrom(src);
	}

	async flush(): Promise<Error | undefined> {
		return this.#stream.flush();
	}

	async close(): Promise<void> {
		return this.#stream.close();
	}

	async cancel(err: StreamError): Promise<void> {
		return this.#stream.cancel(err);
	}

	closed(): Promise<void> {
		return this.#stream.closed();
	}

	/**
	 * Create a mock send stream with an in-memory buffer.
	 */
	static create(streamId: number): MockSendStream {
		const { writable } = new TransformStream<Uint8Array>();
		return new MockSendStream({
			streamId,
			stream: writable,
		});
	}
}

/**
 * Mock ReceiveStream implementation for testing.
 * This wraps a real ReceiveStream for easier testing.
 */
export class MockReceiveStream {
	readonly id: number;
	#stream: ReceiveStream;

	constructor(init: ReceiveStreamInit) {
		this.id = init.streamId;
		this.#stream = new ReceiveStream(init);
	}

	async readUint8Array(
		transfer?: ArrayBufferLike,
	): Promise<[Uint8Array, undefined] | [undefined, Error]> {
		return this.#stream.readUint8Array(transfer);
	}

	async readString(): Promise<[string, Error | undefined]> {
		return this.#stream.readString();
	}

	async readVarint(): Promise<[number, Error | undefined]> {
		return this.#stream.readVarint();
	}

	async readBigVarint(): Promise<[bigint, Error | undefined]> {
		return this.#stream.readBigVarint();
	}

	async readUint8(): Promise<[number, Error | undefined]> {
		return this.#stream.readUint8();
	}

	async readBoolean(): Promise<[boolean, Error | undefined]> {
		return this.#stream.readBoolean();
	}

	async readStringArray(): Promise<[string[], Error | undefined]> {
		return this.#stream.readStringArray();
	}

	async cancel(reason: StreamError): Promise<void> {
		return this.#stream.cancel(reason);
	}

	closed(): Promise<void> {
		return this.#stream.closed();
	}

	/**
	 * Create a mock receive stream with pre-filled data.
	 */
	static create(streamId: number, data?: Uint8Array[]): MockReceiveStream {
		const { readable, writable } = new TransformStream<Uint8Array>();
		
		// If data is provided, write it to the stream
		if (data && data.length > 0) {
			const writer = writable.getWriter();
			(async () => {
				for (const chunk of data) {
					await writer.write(chunk);
				}
				await writer.close();
			})();
		}

		return new MockReceiveStream({
			streamId,
			stream: readable,
		});
	}
}

/**
 * Mock Connection implementation for testing.
 * This provides a test double for the Connection class.
 */
export class MockConnection {
	#mockWebTransport: MockWebTransport;
	#clientBiStreamCounter = 0;
	#serverBiStreamCounter = 1;
	#clientUniStreamCounter = 2;
	#serverUniStreamCounter = 3;

	constructor(mockWebTransport?: MockWebTransport) {
		this.#mockWebTransport = mockWebTransport || new MockWebTransport();
	}

	/**
	 * Get the underlying mock WebTransport for test control.
	 */
	getWebTransport(): MockWebTransport {
		return this.#mockWebTransport;
	}

	async openStream(): Promise<[Stream, undefined] | [undefined, Error]> {
		try {
			const wtStream = await this.#mockWebTransport.createBidirectionalStream();
			const stream = new Stream({
				streamId: this.#clientBiStreamCounter,
				stream: wtStream,
			});
			this.#clientBiStreamCounter += 4;
			return [stream, undefined];
		} catch (e) {
			return [undefined, e as Error];
		}
	}

	async openUniStream(): Promise<[SendStream, undefined] | [undefined, Error]> {
		try {
			const wtStream = await this.#mockWebTransport.createUnidirectionalStream();
			const stream = new SendStream({
				streamId: this.#clientUniStreamCounter,
				stream: wtStream,
			});
			this.#clientUniStreamCounter += 4;
			return [stream, undefined];
		} catch (e) {
			return [undefined, e as Error];
		}
	}

	async acceptStream(): Promise<[Stream, undefined] | [undefined, Error]> {
		const biReader = this.#mockWebTransport.getBiReader();
		const { done, value: wtStream } = await biReader.read();
		if (done) {
			return [undefined, new Error("Failed to accept stream")];
		}
		const stream = new Stream({
			streamId: this.#serverBiStreamCounter,
			stream: wtStream,
		});
		this.#serverBiStreamCounter += 4;
		return [stream, undefined];
	}

	async acceptUniStream(): Promise<[ReceiveStream, undefined] | [undefined, Error]> {
		const uniReader = this.#mockWebTransport.getUniReader();
		const { done, value: wtStream } = await uniReader.read();
		if (done) {
			return [undefined, new Error("Failed to accept unidirectional stream")];
		}
		const stream = new ReceiveStream({
			streamId: this.#serverUniStreamCounter,
			stream: wtStream,
		});
		this.#serverUniStreamCounter += 4;
		return [stream, undefined];
	}

	close(closeInfo?: WebTransportCloseInfo): void {
		this.#mockWebTransport.close(closeInfo);
	}

	get ready(): Promise<void> {
		return this.#mockWebTransport.ready;
	}

	get closed(): Promise<WebTransportCloseInfo> {
		return this.#mockWebTransport.closed;
	}

	/**
	 * Simulate incoming bidirectional stream for testing.
	 */
	simulateIncomingBiStream(): WebTransportBidirectionalStream {
		const { readable, writable } = new TransformStream<Uint8Array>();
		const stream = { readable, writable } as WebTransportBidirectionalStream;
		this.#mockWebTransport.getBiReader().enqueue(stream);
		return stream;
	}

	/**
	 * Simulate incoming unidirectional stream for testing.
	 */
	simulateIncomingUniStream(data?: Uint8Array[]): ReadableStream<Uint8Array> {
		const { readable, writable } = new TransformStream<Uint8Array>();
		
		// If data is provided, write it to the stream
		if (data && data.length > 0) {
			const writer = writable.getWriter();
			(async () => {
				for (const chunk of data) {
					await writer.write(chunk);
				}
				await writer.close();
			})();
		}
		
		this.#mockWebTransport.getUniReader().enqueue(readable);
		return readable;
	}

	/**
	 * Mark the connection as ready for testing.
	 */
	markReady(): void {
		this.#mockWebTransport.markReady();
	}

	/**
	 * Fail the connection ready promise for error testing.
	 */
	failReady(error: Error): void {
		this.#mockWebTransport.failReady(error);
	}
}
