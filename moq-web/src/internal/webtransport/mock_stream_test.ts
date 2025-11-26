/**
 * Mock implementations for WebTransport streams used in testing.
 *
 * This module provides mock implementations of:
 * - MockSendStream: Mock for SendStream (implements Writer interface)
 * - MockReceiveStream: Mock for ReceiveStream (implements Reader interface)
 * - MockStream: Mock for bidirectional Stream (combines both)
 *
 * @module
 */

import type { Reader, Writer } from "@okudai/golikejs/io";
import { EOFError } from "@okudai/golikejs/io";
import type { StreamError } from "./error.ts";

/**
 * Mock implementation of SendStream for testing.
 * Implements the Writer interface.
 */
export class MockSendStream implements Writer {
	/** Stream ID (required property) */
	public readonly id: bigint;

	/** Written data chunks */
	public writtenData: Uint8Array[] = [];

	/** Track close method calls */
	public closeCalls: number = 0;

	/** Track cancel method calls */
	public cancelCalls: Array<StreamError> = [];

	/** Error to return from write operations */
	public writeError: Error | undefined = undefined;

	constructor(streamId: bigint = 0n) {
		this.id = streamId;
	}

	/**
	 * Implements Writer.write
	 */
	async write(p: Uint8Array): Promise<[number, Error | undefined]> {
		if (this.writeError) {
			return [0, this.writeError];
		}
		this.writtenData.push(p.slice()); // Clone to preserve data
		return [p.length, undefined];
	}

	/** Close the stream */
	async close(): Promise<void> {
		this.closeCalls++;
	}

	/** Cancel the stream with error */
	async cancel(err: StreamError): Promise<void> {
		this.cancelCalls.push(err);
	}

	/** Returns a resolved promise */
	closed(): Promise<void> {
		return Promise.resolve();
	}

	/** Reset all call tracking */
	public reset(): void {
		this.writtenData = [];
		this.closeCalls = 0;
		this.cancelCalls = [];
		this.writeError = undefined;
	}

	/** Get all written data as a single Uint8Array */
	public getAllWrittenData(): Uint8Array {
		const totalLength = this.writtenData.reduce((sum, chunk) => sum + chunk.length, 0);
		const result = new Uint8Array(totalLength);
		let offset = 0;
		for (const chunk of this.writtenData) {
			result.set(chunk, offset);
			offset += chunk.length;
		}
		return result;
	}
}

/**
 * Mock implementation of ReceiveStream for testing.
 * Implements the Reader interface.
 */
export class MockReceiveStream implements Reader {
	/** Stream ID (required property) */
	public readonly id: bigint;

	/** Mock data to be returned by read operations */
	#data: Uint8Array;
	#offset: number = 0;

	/** Track cancel method calls */
	public cancelCalls: Array<StreamError> = [];

	/** Error to return from read operations */
	public readError: Error | undefined = undefined;

	constructor(streamId: bigint = 0n, data: Uint8Array = new Uint8Array(0)) {
		this.id = streamId;
		this.#data = data;
	}

	/**
	 * Implements Reader.read
	 */
	async read(p: Uint8Array): Promise<[number, Error | undefined]> {
		if (this.readError) {
			return [0, this.readError];
		}
		if (this.#offset >= this.#data.length) {
			return [0, new EOFError()];
		}

		const remaining = this.#data.length - this.#offset;
		const n = Math.min(p.length, remaining);
		p.set(this.#data.subarray(this.#offset, this.#offset + n));
		this.#offset += n;
		return [n, undefined];
	}

	/** Cancel the stream */
	async cancel(reason: StreamError): Promise<void> {
		this.cancelCalls.push(reason);
	}

	/** Returns a resolved promise */
	closed(): Promise<void> {
		return Promise.resolve();
	}

	/** Set the data to be read */
	public setData(data: Uint8Array): void {
		this.#data = data;
		this.#offset = 0;
	}

	/** Reset the read position */
	public reset(): void {
		this.#offset = 0;
		this.cancelCalls = [];
		this.readError = undefined;
	}
}

/**
 * Mock implementation of a bidirectional Stream for testing.
 * Combines MockSendStream (writable) and MockReceiveStream (readable).
 */
export class MockStream {
	/** Stream ID */
	public readonly id: bigint;

	/** Mock writable stream */
	public readonly writable: MockSendStream;

	/** Mock readable stream */
	public readonly readable: MockReceiveStream;

	constructor(streamId: bigint = 0n, readData: Uint8Array = new Uint8Array(0)) {
		this.id = streamId;
		this.writable = new MockSendStream(streamId);
		this.readable = new MockReceiveStream(streamId, readData);
	}

	/** Reset both streams */
	public reset(): void {
		this.writable.reset();
		this.readable.reset();
	}
}
