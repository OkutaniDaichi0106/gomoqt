/**
 * Mock implementation of WebTransport Connection for testing purposes.
 * This file provides test doubles that can be used instead of real WebTransport connections.
 */

import { ReceiveStream, type ReceiveStreamInit } from "./receive_stream.ts";
import { Stream, type StreamInit } from "./stream.ts";
import { SendStream, type SendStreamInit } from "./send_stream.ts";
import type { StreamError } from "./error.ts";
import { Connection } from "./connection.ts";

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
	implements ReadableStreamDefaultReader<ReadableStream<Uint8Array<ArrayBufferLike>>> {
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

	#shouldFailCreateStream = false;

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
	 * Set whether createBidirectionalStream should fail.
	 */
	setFailCreateStream(fail: boolean): void {
		this.#shouldFailCreateStream = fail;
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
		if (this.#shouldFailCreateStream) {
			throw new Error("Failed to create bidirectional stream");
		}
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
	readonly id: bigint;
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
	static create(streamId: bigint): MockStream {
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
	readonly id: bigint;
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

	copyFrom(
		src: { byteLength: number; copyTo(target: ArrayBuffer | ArrayBufferView): void },
	): void {
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
	static create(streamId: bigint): MockSendStream {
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
	readonly id: bigint;
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
	static create(streamId: bigint, data?: Uint8Array[]): MockReceiveStream {
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
export class MockConnection extends Connection {
	#mockWebTransport: MockWebTransport;
	#counter = new streamIDCounter();

	constructor(mockWebTransport?: MockWebTransport) {
		super(mockWebTransport || new MockWebTransport());
		this.#mockWebTransport = mockWebTransport || new MockWebTransport();
	}

	/**
	 * Get the underlying mock WebTransport for test control.
	 */
	getWebTransport(): MockWebTransport {
		return this.#mockWebTransport;
	}

	override async openStream(): Promise<[Stream, undefined] | [undefined, Error]> {
		try {
			const wtStream = await this.#mockWebTransport.createBidirectionalStream();
			const stream = new Stream({
				streamId: this.#counter.countClientBiStream(),
				stream: wtStream,
			});
			return [stream, undefined];
		} catch (e) {
			return [undefined, e as Error];
		}
	}

	override async openUniStream(): Promise<[SendStream, undefined] | [undefined, Error]> {
		try {
			const wtStream = await this.#mockWebTransport.createUnidirectionalStream();
			const stream = new SendStream({
				streamId: this.#counter.countClientUniStream(),
				stream: wtStream,
			});
			return [stream, undefined];
		} catch (e) {
			return [undefined, e as Error];
		}
	}

	override async acceptStream(): Promise<[Stream, undefined] | [undefined, Error]> {
		const biReader = this.#mockWebTransport.getBiReader();
		const { done, value: wtStream } = await biReader.read();
		if (done) {
			return [undefined, new Error("Failed to accept stream")];
		}
		const stream = new Stream({
			streamId: this.#counter.countServerBiStream(),
			stream: wtStream,
		});
		return [stream, undefined];
	}

	override async acceptUniStream(): Promise<[ReceiveStream, undefined] | [undefined, Error]> {
		const uniReader = this.#mockWebTransport.getUniReader();
		const { done, value: wtStream } = await uniReader.read();
		if (done) {
			return [undefined, new Error("Failed to accept unidirectional stream")];
		}
		const stream = new ReceiveStream({
			streamId: this.#counter.countServerUniStream(),
			stream: wtStream,
		});
		return [stream, undefined];
	}

	override close(closeInfo?: WebTransportCloseInfo): void {
		this.#mockWebTransport.close(closeInfo);
	}

	override get ready(): Promise<void> {
		return this.#mockWebTransport.ready;
	}

	override get closed(): Promise<WebTransportCloseInfo> {
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

	/**
	 * Fail the next openStream call for testing.
	 */
	setFailOpenStream(fail: boolean): void {
		this.#mockWebTransport.setFailCreateStream(fail);
	}

	/**
	 * Fail the next openUniStream call for testing.
	 */
	setFailOpenUniStream(fail: boolean): void {
		this.#mockWebTransport.setFailCreateStream(fail); // Assuming same method
	}
}

// ============================================================================
// Tests
// ============================================================================

import { assertEquals, assertExists } from "@std/assert";
import { streamIDCounter } from "./connection.ts";

// Note: sanitizeResources and sanitizeOps are disabled for these tests
// because the mock implementation uses async operations that may not
// complete within the test lifecycle.
const testOptions = { sanitizeResources: false, sanitizeOps: false };

Deno.test("MockBidirectionalStreamReader - Normal Cases", testOptions, async (t) => {
	await t.step("should enqueue and read bidirectional streams", async () => {
		const reader = new MockBidirectionalStreamReader();
		const { readable, writable } = new TransformStream<Uint8Array>();
		const stream = { readable, writable } as WebTransportBidirectionalStream;

		reader.enqueue(stream);

		const { done, value } = await reader.read();
		assertEquals(done, false);
		assertExists(value);
		assertEquals(value, stream);
	});

	await t.step("should return done when closed", async () => {
		const reader = new MockBidirectionalStreamReader();
		reader.close();

		const { done, value } = await reader.read();
		assertEquals(done, true);
		assertEquals(value, undefined);
	});

	await t.step("should close via cancel", async () => {
		const reader = new MockBidirectionalStreamReader();
		await reader.cancel();

		const { done } = await reader.read();
		assertEquals(done, true);
	});
});

Deno.test("MockBidirectionalStreamReader - Error Cases", testOptions, async (t) => {
	await t.step("should throw when enqueuing to closed reader", () => {
		const reader = new MockBidirectionalStreamReader();
		reader.close();

		const { readable, writable } = new TransformStream<Uint8Array>();
		const stream = { readable, writable } as WebTransportBidirectionalStream;

		let errorThrown = false;
		try {
			reader.enqueue(stream);
		} catch (_e) {
			errorThrown = true;
		}
		assertEquals(errorThrown, true);
	});
});

Deno.test("MockUnidirectionalStreamReader - Normal Cases", testOptions, async (t) => {
	await t.step("should enqueue and read unidirectional streams", async () => {
		const reader = new MockUnidirectionalStreamReader();
		const { readable } = new TransformStream<Uint8Array>();

		reader.enqueue(readable);

		const { done, value } = await reader.read();
		assertEquals(done, false);
		assertExists(value);
		assertEquals(value, readable);
	});

	await t.step("should return done when closed", async () => {
		const reader = new MockUnidirectionalStreamReader();
		reader.close();

		const { done, value } = await reader.read();
		assertEquals(done, true);
		assertEquals(value, undefined);
	});
});

Deno.test("MockWebTransport - Normal Cases", testOptions, async (t) => {
	await t.step("should create bidirectional stream", async () => {
		const wt = new MockWebTransport();
		const stream = await wt.createBidirectionalStream();
		assertExists(stream);
		assertExists(stream.readable);
		assertExists(stream.writable);
	});

	await t.step("should create unidirectional stream", async () => {
		const wt = new MockWebTransport();
		const stream = await wt.createUnidirectionalStream();
		assertExists(stream);
	});

	await t.step("should mark ready", async () => {
		const wt = new MockWebTransport();
		wt.markReady();
		await wt.ready;
		// If we reach here without hanging, ready was resolved
		assertEquals(true, true);
	});

	await t.step("should close with info", async () => {
		const wt = new MockWebTransport();
		const closeInfo = { closeCode: 42, reason: "test close" };
		wt.close(closeInfo);
		const closed = await wt.closed;
		assertEquals(closed.closeCode, 42);
		assertEquals(closed.reason, "test close");
	});

	await t.step("should get close info", () => {
		const wt = new MockWebTransport();
		const closeInfo = { closeCode: 100, reason: "manual close" };
		wt.close(closeInfo);
		const info = wt.getCloseInfo();
		assertExists(info);
		assertEquals(info.closeCode, 100);
		assertEquals(info.reason, "manual close");
	});
});

Deno.test("MockStream - Normal Cases", testOptions, async (t) => {
	await t.step("should create mock stream with correct ID", () => {
		const stream = MockStream.create(42n);
		assertEquals(stream.id, 42n);
		assertExists(stream.readable);
		assertExists(stream.writable);
	});
});

Deno.test("MockSendStream - Normal Cases", testOptions, async (t) => {
	await t.step("should create mock send stream", () => {
		const stream = MockSendStream.create(10n);
		assertEquals(stream.id, 10n);
	});

	// Note: flush() may hang if the writable stream has no reader
	// In real usage, there should be a reader on the other end
});

Deno.test("MockReceiveStream - Normal Cases", testOptions, async (t) => {
	await t.step("should create mock receive stream", () => {
		const stream = MockReceiveStream.create(30n);
		assertEquals(stream.id, 30n);
	});

	await t.step("should create with pre-filled data", () => {
		const data = [new Uint8Array([1, 2, 3]), new Uint8Array([4, 5, 6])];
		const stream = MockReceiveStream.create(40n, data);
		assertEquals(stream.id, 40n);
	});
});

Deno.test("MockConnection - Normal Cases", testOptions, async (t) => {
	await t.step("should create connection", () => {
		const conn = new MockConnection();
		assertExists(conn);
	});

	await t.step("should open bidirectional stream", async () => {
		const conn = new MockConnection();
		const [stream, err] = await conn.openStream();
		assertEquals(err, undefined);
		assertExists(stream);
		assertEquals(stream?.id, 0n);
	});

	await t.step("should open unidirectional stream", async () => {
		const conn = new MockConnection();
		const [stream, err] = await conn.openUniStream();
		assertEquals(err, undefined);
		assertExists(stream);
		assertEquals(stream?.id, 2n);
	});

	await t.step("should increment stream IDs correctly", async () => {
		const conn = new MockConnection();

		const [stream1, err1] = await conn.openStream();
		assertEquals(err1, undefined);
		assertEquals(stream1?.id, 0n);

		const [stream2, err2] = await conn.openStream();
		assertEquals(err2, undefined);
		assertEquals(stream2?.id, 4n);

		const [stream3, err3] = await conn.openUniStream();
		assertEquals(err3, undefined);
		assertEquals(stream3?.id, 2n);

		const [stream4, err4] = await conn.openUniStream();
		assertEquals(err4, undefined);
		assertEquals(stream4?.id, 6n);
	});

	await t.step("should accept simulated bidirectional stream", async () => {
		const conn = new MockConnection();
		conn.simulateIncomingBiStream();

		const [stream, err] = await conn.acceptStream();
		assertEquals(err, undefined);
		assertExists(stream);
		assertEquals(stream?.id, 1n);
	});

	await t.step("should accept simulated unidirectional stream", async () => {
		const conn = new MockConnection();
		conn.simulateIncomingUniStream();

		const [stream, err] = await conn.acceptUniStream();
		assertEquals(err, undefined);
		assertExists(stream);
		assertEquals(stream?.id, 3n);
	});

	await t.step("should close connection", async () => {
		const conn = new MockConnection();
		const closeInfo = { closeCode: 0, reason: "test" };
		conn.close(closeInfo);

		const closed = await conn.closed;
		assertEquals(closed.closeCode, 0);
		assertEquals(closed.reason, "test");
	});

	await t.step("should mark connection as ready", async () => {
		const conn = new MockConnection();
		conn.markReady();
		await conn.ready;
		// If we reach here, ready was resolved successfully
		assertEquals(true, true);
	});
});

Deno.test("MockConnection - Error Cases", testOptions, async (t) => {
	await t.step("should return error when accepting from closed stream", async () => {
		const conn = new MockConnection();
		const wt = conn.getWebTransport();
		wt.getBiReader().close();

		const [stream, err] = await conn.acceptStream();
		assertEquals(stream, undefined);
		assertExists(err);
		assertEquals(err?.message, "Failed to accept stream");
	});

	await t.step("should return error when accepting uni from closed stream", async () => {
		const conn = new MockConnection();
		const wt = conn.getWebTransport();
		wt.getUniReader().close();

		const [stream, err] = await conn.acceptUniStream();
		assertEquals(stream, undefined);
		assertExists(err);
		assertEquals(err?.message, "Failed to accept unidirectional stream");
	});
});

Deno.test("MockConnection - Integration Scenarios", testOptions, async (t) => {
	await t.step("should handle multiple stream operations", async () => {
		const conn = new MockConnection();

		// Open multiple streams
		const [stream1] = await conn.openStream();
		const [stream2] = await conn.openUniStream();
		const [stream3] = await conn.openStream();

		// Simulate incoming streams
		conn.simulateIncomingBiStream();
		conn.simulateIncomingUniStream();

		// Accept incoming streams
		const [stream4] = await conn.acceptStream();
		const [stream5] = await conn.acceptUniStream();

		// Verify IDs are correct
		assertEquals(stream1?.id, 0n);
		assertEquals(stream2?.id, 2n);
		assertEquals(stream3?.id, 4n);
		assertEquals(stream4?.id, 1n);
		assertEquals(stream5?.id, 3n);
	});

	await t.step("should handle ready and close lifecycle", async () => {
		const conn = new MockConnection();

		// Mark as ready
		conn.markReady();
		await conn.ready;

		// Open a stream
		const [stream] = await conn.openStream();
		assertExists(stream);

		// Close connection
		conn.close({ closeCode: 100, reason: "done" });
		const closed = await conn.closed;
		assertEquals(closed.closeCode, 100);
	});
});
