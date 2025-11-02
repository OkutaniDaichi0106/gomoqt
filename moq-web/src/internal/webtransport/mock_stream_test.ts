/**
 * Mock implementations for WebTransport streams used in testing.
 *
 * This module provides mock implementations of:
 * - MockSendStream: Mock for SendStream (write operations)
 * - MockReceiveStream: Mock for ReceiveStream (read operations)
 * - MockStream: Mock for bidirectional Stream (combines both)
 *
 * All mocks use spy functionality from @std/testing/mock to track method calls
 * and allow customization of return values for testing various scenarios.
 *
 * @example
 * ```ts
 * import { MockStream } from "./internal/webtransport/mock_stream_test.ts";
 *
 * // Create a bidirectional mock stream
 * const mockStream = new MockStream(42n);
 *
 * // Configure mock behavior
 * mockStream.writable.flushError = new Error("Flush failed");
 * mockStream.readable.data = [new Uint8Array([1, 2, 3])];
 *
 * // Use in tests
 * await someFunction(mockStream);
 *
 * // Verify calls
 * assertEquals(mockStream.writable.flushCalls > 0, true);
 * ```
 *
 * @module
 */

import { spy } from "@std/testing/mock";
import type { StreamError } from "./error.ts";
import { EOFError } from "@okudai/golikejs/io";

// Define the minimal interface we need to mock (duck typing approach)
// We don't need to match the full class structure, just the public API

export interface Source {
	byteLength: number;
	copyTo(target: ArrayBuffer | ArrayBufferView<ArrayBufferLike>): void;
}

/**
 * Mock implementation of SendStream for testing.
 * All write methods are spies that can be tracked and verified.
 * This uses duck typing to be compatible with the SendStream class.
 */
export class MockSendStream {
	/** Stream ID (required property) */
	public readonly streamId: bigint = 0n;

	/** Track flush method calls */
	public flushCalls: number = 0;
	/** Track close method calls */
	public closeCalls: number = 0;
	/** Track cancel method calls */
	public cancelCalls: Array<StreamError> = [];

	/** Spy for writeVarint method */
	public writeVarint = spy((_value: number): void => {});

	/** Spy for writeBoolean method */
	public writeBoolean = spy((_value: boolean): void => {});

	/** Spy for writeBigVarint method */
	public writeBigVarint = spy((_value: bigint): void => {});

	/** Spy for writeString method */
	public writeString = spy((_str: string): void => {});

	/** Spy for writeStringArray method */
	public writeStringArray = spy((_strings: string[]): void => {});

	/** Spy for writeUint8Array method */
	public writeUint8Array = spy((_data: Uint8Array): void => {});

	/** Spy for writeUint8 method */
	public writeUint8 = spy((_value: number): void => {});

	/** Spy for copyFrom method */
	public copyFrom = spy((_src: Source): void => {});

	/**
	 * Spy for flush method.
	 * Returns undefined by default (no error).
	 * Override flushError to simulate flush errors.
	 */
	public flushError: Error | undefined = undefined;
	public flush = spy(async (): Promise<Error | undefined> => {
		this.flushCalls++;
		return this.flushError;
	});

	/** Spy for close method */
	public close = spy(async (): Promise<void> => {
		this.closeCalls++;
	});

	/** Spy for cancel method */
	public cancel = spy(async (err: StreamError): Promise<void> => {
		this.cancelCalls.push(err);
	});

	/** Spy for closed method */
	public closed = spy((): Promise<void> => Promise.resolve());

	/** Reset all call tracking */
	public reset(): void {
		this.flushCalls = 0;
		this.closeCalls = 0;
		this.cancelCalls = [];
	}
}

/**
 * Mock implementation of ReceiveStream for testing.
 * All read methods are spies that can be tracked and verified.
 * Use the data array to provide mock data for read operations.
 * This uses duck typing to be compatible with the ReceiveStream class.
 */
export class MockReceiveStream {
	/** Stream ID (required property) */
	public readonly streamId: bigint = 0n;

	/** Mock data to be returned by read operations */
	public data: Uint8Array[] = [];
	/** Current position in the data array */
	private dataIndex = 0;

	/** Track cancel method calls */
	public cancelCalls: Array<StreamError> = [];

	/**
	 * Error to return from read operations.
	 * If set, all read operations will return this error.
	 */
	public readError: Error | undefined = undefined;

	/**
	 * Override to customize readVarint behavior.
	 * By default returns simple varint decoding from data array.
	 */
	public readVarintImpl: () => Promise<[number, Error | undefined]> = async () => {
		if (this.readError) {
			return [0, this.readError];
		}
		if (this.dataIndex >= this.data.length) {
			return [0, new EOFError()];
		}
		const data = this.data[this.dataIndex++];
		// Simple varint decoding (first byte only for testing)
		return [(data?.[0]) || 0, undefined];
	};

	/**
	 * Spy for readVarint method.
	 * Returns [0, readError] or [0, EOF] when no data available.
	 */
	public readVarint = spy(async (): Promise<[number, Error | undefined]> => {
		return this.readVarintImpl();
	});

	/** Spy for readBoolean method */
	public readBoolean = spy(async (): Promise<[boolean, Error | undefined]> => {
		if (this.readError) {
			return [false, this.readError];
		}
		if (this.dataIndex >= this.data.length) {
			return [false, new EOFError()];
		}
		const data = this.data[this.dataIndex++];
		return [(data?.[0]) === 1, undefined];
	});

	/** Spy for readBigVarint method */
	public readBigVarint = spy(async (): Promise<[bigint, Error | undefined]> => {
		if (this.readError) {
			return [0n, this.readError];
		}
		if (this.dataIndex >= this.data.length) {
			return [0n, new EOFError()];
		}
		const data = this.data[this.dataIndex++];
		return [BigInt((data?.[0]) || 0), undefined];
	});

	/** Spy for readString method */
	public readString = spy(async (): Promise<[string, Error | undefined]> => {
		if (this.readError) {
			return ["", this.readError];
		}
		if (this.dataIndex >= this.data.length) {
			return ["", new EOFError()];
		}
		const data = this.data[this.dataIndex++];
		if (!data) {
			return ["", new EOFError()];
		}
		const decoder = new TextDecoder();
		return [decoder.decode(data), undefined];
	});

	/** Spy for readStringArray method */
	public readStringArray = spy(async (): Promise<[string[], Error | undefined]> => {
		if (this.readError) {
			return [[], this.readError];
		}
		return [[], undefined];
	});

	/** Spy for readUint8Array method */
	public readUint8Array = spy(
		async (_transfer?: ArrayBufferLike): Promise<
			[Uint8Array<ArrayBufferLike>, undefined] | [undefined, Error]
		> => {
			if (this.readError) {
				return [undefined, this.readError];
			}
			if (this.dataIndex >= this.data.length) {
				return [undefined, new EOFError()];
			}
			const data = this.data[this.dataIndex++];
			if (!data) {
				return [undefined, new EOFError()];
			}
			return [data, undefined];
		},
	);

	/** Spy for readUint8 method */
	public readUint8 = spy(async (): Promise<[number, Error | undefined]> => {
		if (this.readError) {
			return [0, this.readError];
		}
		if (this.dataIndex >= this.data.length) {
			return [0, new EOFError()];
		}
		const data = this.data[this.dataIndex++];
		return [(data?.[0]) || 0, undefined];
	});

	/** Spy for pushN method (simulates reading data into internal buffer) */
	public pushN = spy(async (_n: number): Promise<Error | undefined> => {
		if (this.readError) {
			return this.readError;
		}
		return undefined;
	});

	/**
	 * Override to customize fillN behavior.
	 * By default returns undefined (no error).
	 */
	public fillNImpl: (buffer: Uint8Array, n: number) => Promise<Error | undefined> = async (
		_buffer: Uint8Array,
		_n: number,
	) => {
		if (this.readError) {
			return this.readError;
		}
		return undefined;
	};

	/** Spy for fillN method (fills n bytes into buffer) */
	public fillN = spy(async (buffer: Uint8Array, n: number): Promise<Error | undefined> => {
		return this.fillNImpl(buffer, n);
	});

	/** Spy for cancel method */
	public cancel = spy(async (reason: StreamError): Promise<void> => {
		this.cancelCalls.push(reason);
	});

	/** Spy for closed method */
	public closed = spy((): Promise<void> => Promise.resolve());

	/** Reset data index and call tracking */
	public reset(): void {
		this.dataIndex = 0;
		this.cancelCalls = [];
	}
}

/**
 * Mock implementation of Stream (bidirectional stream).
 * Combines MockSendStream and MockReceiveStream with a stream ID.
 * This is compatible with the Stream class used in the application.
 */
export class MockStream {
	/** Stream ID */
	public readonly streamId: bigint;
	/** Writable stream (send) */
	public readonly writable: MockSendStream;
	/** Readable stream (receive) */
	public readonly readable: MockReceiveStream;

	/**
	 * Create a new MockStream with the specified stream ID.
	 * @param streamId The stream ID (default: 0n)
	 */
	constructor(streamId: bigint = 0n) {
		this.streamId = streamId;
		this.writable = new MockSendStream();
		this.readable = new MockReceiveStream();
		// Set the same streamId on the child streams
		(this.writable as any).streamId = streamId;
		(this.readable as any).streamId = streamId;
	}

	/**
	 * Reset all tracking on both writable and readable streams.
	 */
	public reset(): void {
		this.writable.reset();
		this.readable.reset();
	}
}
