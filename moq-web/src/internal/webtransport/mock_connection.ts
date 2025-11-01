/**/**

 * Mock implementation for WebTransport Connection used in testing. * Mock implementation for WebTransport Connection used in testing.

 *  * 

 * This module provides a mock implementation of the Connection class that * This module provides a mock implementation of the Connection class that

 * wraps WebTransport functionality for testing purposes. * wraps WebTransport functionality for testing purposes.

 *  * 

 * MockConnection simulates: * MockConnection simulates:

 * - Opening bidirectional and unidirectional streams * - Opening bidirectional and unidirectional streams

 * - Accepting incoming streams * - Accepting incoming streams

 * - Connection lifecycle (ready, closed) * - Connection lifecycle (ready, closed)

 * - Error conditions * - Error conditions

 *  * 

 * @example * @example

 * ```ts * ```ts

 * import { MockConnection } from "./internal/webtransport/mock_connection.ts"; * import { MockConnection } from "./internal/webtransport/mock_connection.ts";

 * import { MockStream } from "./internal/webtransport/mock_stream_test.ts"; * import { MockStream } from "./internal/webtransport/mock_stream_test.ts";

 *  * 

 * // Create a mock connection * // Create a mock connection

 * const mockConn = new MockConnection(); * const mockConn = new MockConnection();

 *  * 

 * // Configure mock streams to be accepted * // Configure mock streams to be accepted

 * const mockStream = new MockStream(1n); * const mockStream = new MockStream(1n);

 * mockConn.incomingBiStreams.push(mockStream); * mockConn.incomingBiStreams.push(mockStream);

 *  * 

 * // Use in tests * // Use in tests

 * const [stream, err] = await mockConn.acceptStream(); * const [stream, err] = await mockConn.acceptStream();

 * assertEquals(err, undefined); * assertEquals(err, undefined);

 * assertEquals(stream?.streamId, 1n); * assertEquals(stream?.streamId, 1n);

 * ``` * ```

 * *

 * @module * @module

 */ */



import { spy } from "@std/testing/mock";import { spy } from "@std/testing/mock";

import type { ReceiveStream } from "./reader.ts";import type { ReceiveStream } from "./reader.ts";

import type { Stream } from "./stream.ts";import type { Stream } from "./stream.ts";

import type { SendStream } from "./writer.ts";import type { SendStream } from "./writer.ts";

import { MockStream, MockSendStream, MockReceiveStream } from "./mock_stream_test.ts";import { MockStream, MockSendStream, MockReceiveStream } from "./mock_stream_test.ts";



/**/**

 * Mock implementation of Connection for testing. * Mock implementation of Connection for testing.

 * Simulates WebTransport connection behavior without requiring actual network connectivity. * Simulates WebTransport connection behavior without requiring actual network connectivity.

 * Uses duck typing to be compatible with the Connection class. */

 */export class MockConnection {

export class MockConnection {	/** Track stream ID counters */

	/** Track stream ID counters */	private clientBiStreamCounter: number = 0;

	private clientBiStreamCounter: number = 0;	private serverBiStreamCounter: number = 1;

	private serverBiStreamCounter: number = 1;	private clientUniStreamCounter: number = 2;

	private clientUniStreamCounter: number = 2;	private serverUniStreamCounter: number = 3;

	private serverUniStreamCounter: number = 3;

	/** Queue of incoming bidirectional streams */

	/** Queue of incoming bidirectional streams */	public incomingBiStreams: MockStream[] = [];

	public incomingBiStreams: MockStream[] = [];	/** Queue of incoming unidirectional streams */

	/** Queue of incoming unidirectional streams */	public incomingUniStreams: MockReceiveStream[] = [];

	public incomingUniStreams: MockReceiveStream[] = [];

	/** Control whether operations should fail */

	/** Control whether operations should fail */	public openStreamError: Error | undefined = undefined;

	public openStreamError: Error | undefined = undefined;	public openUniStreamError: Error | undefined = undefined;

	public openUniStreamError: Error | undefined = undefined;	public acceptStreamError: Error | undefined = undefined;

	public acceptStreamError: Error | undefined = undefined;	public acceptUniStreamError: Error | undefined = undefined;

	public acceptUniStreamError: Error | undefined = undefined;

	/** Promise that resolves when the connection is ready */

	/** Promise that resolves when the connection is ready */	public ready: Promise<void> = Promise.resolve();

	public ready: Promise<void> = Promise.resolve();	/** Promise that resolves when the connection is closed */

	/** Promise that resolves when the connection is closed */	private _closed: Promise<WebTransportCloseInfo>;

	private _closed: Promise<WebTransportCloseInfo>;	private _closeResolver?: (value: WebTransportCloseInfo) => void;

	private _closeResolver?: (value: WebTransportCloseInfo) => void;

	/** Track close method calls */

	/** Track close method calls */	public closeCalls: Array<WebTransportCloseInfo | undefined> = [];

	public closeCalls: Array<WebTransportCloseInfo | undefined> = [];

	/**

	/**	 * Create a new MockConnection.

	 * Create a new MockConnection.	 */

	 */	constructor() {

	constructor() {		this._closed = new Promise((resolve) => {

		this._closed = new Promise((resolve) => {			this._closeResolver = resolve;

			this._closeResolver = resolve;		});

		});	}

	}

	/**

	/**	 * Spy for openStream method.

	 * Spy for openStream method.	 * Opens a client-initiated bidirectional stream.

	 * Opens a client-initiated bidirectional stream.	 */

	 */	public openStream = spy(

	public openStream = spy(		async (): Promise<[Stream, undefined] | [undefined, Error]> => {

		async (): Promise<[Stream, undefined] | [undefined, Error]> => {			if (this.openStreamError) {

			if (this.openStreamError) {				return [undefined, this.openStreamError];

				return [undefined, this.openStreamError];			}

			}

			const streamId = BigInt(this.clientBiStreamCounter);

			const streamId = BigInt(this.clientBiStreamCounter);			this.clientBiStreamCounter += 4;

			this.clientBiStreamCounter += 4;

			const mockStream = new MockStream(streamId);

			const mockStream = new MockStream(streamId);			return [mockStream as unknown as Stream, undefined];

			return [mockStream as unknown as Stream, undefined];		},

		},	);

	);

	/**

	/**	 * Spy for openUniStream method.

	 * Spy for openUniStream method.	 * Opens a client-initiated unidirectional stream.

	 * Opens a client-initiated unidirectional stream.	 */

	 */	public openUniStream = spy(

	public openUniStream = spy(		async (): Promise<[SendStream, undefined] | [undefined, Error]> => {

		async (): Promise<[SendStream, undefined] | [undefined, Error]> => {			if (this.openUniStreamError) {

			if (this.openUniStreamError) {				return [undefined, this.openUniStreamError];

				return [undefined, this.openUniStreamError];			}

			}

			const streamId = BigInt(this.clientUniStreamCounter);

			const streamId = BigInt(this.clientUniStreamCounter);			this.clientUniStreamCounter += 4;

			this.clientUniStreamCounter += 4;

			const mockSendStream = new MockSendStream();

			const mockSendStream = new MockSendStream();			(mockSendStream as any).streamId = streamId;

			(mockSendStream as any).streamId = streamId;			return [mockSendStream as unknown as SendStream, undefined];

			return [mockSendStream as unknown as SendStream, undefined];		},

		},	);

	);

	/**

	/**	 * Spy for acceptStream method.

	 * Spy for acceptStream method.	 * Accepts a server-initiated bidirectional stream.

	 * Accepts a server-initiated bidirectional stream.	 */

	 */	public acceptStream = spy(

	public acceptStream = spy(		async (): Promise<[Stream, undefined] | [undefined, Error]> => {

		async (): Promise<[Stream, undefined] | [undefined, Error]> => {			if (this.acceptStreamError) {

			if (this.acceptStreamError) {				return [undefined, this.acceptStreamError];

				return [undefined, this.acceptStreamError];			}

			}

			if (this.incomingBiStreams.length === 0) {

			if (this.incomingBiStreams.length === 0) {				return [undefined, new Error("No incoming bidirectional streams available")];

				return [undefined, new Error("No incoming bidirectional streams available")];			}

			}

			const mockStream = this.incomingBiStreams.shift();

			const mockStream = this.incomingBiStreams.shift();			if (!mockStream) {

			if (!mockStream) {				return [undefined, new Error("Failed to accept stream")];

				return [undefined, new Error("Failed to accept stream")];			}

			}

			return [mockStream as unknown as Stream, undefined];

			return [mockStream as unknown as Stream, undefined];		},

		},	);

	);

	/**

	/**	 * Spy for acceptUniStream method.

	 * Spy for acceptUniStream method.	 * Accepts a server-initiated unidirectional stream.

	 * Accepts a server-initiated unidirectional stream.	 */

	 */	public acceptUniStream = spy(

	public acceptUniStream = spy(		async (): Promise<[ReceiveStream, undefined] | [undefined, Error]> => {

		async (): Promise<[ReceiveStream, undefined] | [undefined, Error]> => {			if (this.acceptUniStreamError) {

			if (this.acceptUniStreamError) {				return [undefined, this.acceptUniStreamError];

				return [undefined, this.acceptUniStreamError];			}

			}

			if (this.incomingUniStreams.length === 0) {

			if (this.incomingUniStreams.length === 0) {				return [undefined, new Error("No incoming unidirectional streams available")];

				return [undefined, new Error("No incoming unidirectional streams available")];			}

			}

			const mockReceiveStream = this.incomingUniStreams.shift();

			const mockReceiveStream = this.incomingUniStreams.shift();			if (!mockReceiveStream) {

			if (!mockReceiveStream) {				return [undefined, new Error("Failed to accept unidirectional stream")];

				return [undefined, new Error("Failed to accept unidirectional stream")];			}

			}

			return [mockReceiveStream as unknown as ReceiveStream, undefined];

			return [mockReceiveStream as unknown as ReceiveStream, undefined];		},

		},	);

	);

	/**

	/**	 * Spy for close method.

	 * Spy for close method.	 * Closes the connection with optional close info.

	 * Closes the connection with optional close info.	 */

	 */	public close = spy((closeInfo?: WebTransportCloseInfo): void => {

	public close = spy((closeInfo?: WebTransportCloseInfo): void => {		this.closeCalls.push(closeInfo);

		this.closeCalls.push(closeInfo);		if (this._closeResolver) {

		if (this._closeResolver) {			this._closeResolver(closeInfo || { closeCode: 0, reason: "" });

			this._closeResolver(closeInfo || { closeCode: 0, reason: "" });		}

		}	});

	});

	/**

	/**	 * Get the closed promise.

	 * Get the closed promise.	 */

	 */	public get closed(): Promise<WebTransportCloseInfo> {

	public get closed(): Promise<WebTransportCloseInfo> {		return this._closed;

		return this._closed;	}

	}

	/**

	/**	 * Reset all call tracking and queues.

	 * Reset all call tracking and queues.	 */

	 */	public reset(): void {

	public reset(): void {		this.clientBiStreamCounter = 0;

		this.clientBiStreamCounter = 0;		this.serverBiStreamCounter = 1;

		this.serverBiStreamCounter = 1;		this.clientUniStreamCounter = 2;

		this.clientUniStreamCounter = 2;		this.serverUniStreamCounter = 3;

		this.serverUniStreamCounter = 3;		this.incomingBiStreams = [];

		this.incomingBiStreams = [];		this.incomingUniStreams = [];

		this.incomingUniStreams = [];		this.closeCalls = [];

		this.closeCalls = [];		this.openStreamError = undefined;

		this.openStreamError = undefined;		this.openUniStreamError = undefined;

		this.openUniStreamError = undefined;		this.acceptStreamError = undefined;

		this.acceptStreamError = undefined;		this.acceptUniStreamError = undefined;

		this.acceptUniStreamError = undefined;		this._closed = new Promise((resolve) => {

		this._closed = new Promise((resolve) => {			this._closeResolver = resolve;

			this._closeResolver = resolve;		});

		});	}

	}

	/**

	/**	 * Helper method to add an incoming bidirectional stream.

	 * Helper method to add an incoming bidirectional stream.	 * @param stream Optional stream to add (creates a new one if not provided)

	 * @param stream Optional stream to add (creates a new one if not provided)	 * @returns The stream that was added

	 * @returns The stream that was added	 */

	 */	public addIncomingBiStream(stream?: MockStream): MockStream {

	public addIncomingBiStream(stream?: MockStream): MockStream {		const mockStream = stream || new MockStream(BigInt(this.serverBiStreamCounter));

		const mockStream = stream || new MockStream(BigInt(this.serverBiStreamCounter));		if (!stream) {

		if (!stream) {			this.serverBiStreamCounter += 4;

			this.serverBiStreamCounter += 4;		}

		}		this.incomingBiStreams.push(mockStream);

		this.incomingBiStreams.push(mockStream);		return mockStream;

		return mockStream;	}

	}

	/**

	/**	 * Helper method to add an incoming unidirectional stream.

	 * Helper method to add an incoming unidirectional stream.	 * @param stream Optional stream to add (creates a new one if not provided)

	 * @param stream Optional stream to add (creates a new one if not provided)	 * @returns The stream that was added

	 * @returns The stream that was added	 */

	 */	public addIncomingUniStream(stream?: MockReceiveStream): MockReceiveStream {

	public addIncomingUniStream(stream?: MockReceiveStream): MockReceiveStream {		const mockReceiveStream = stream || new MockReceiveStream();

		const mockReceiveStream = stream || new MockReceiveStream();		if (!stream) {

		if (!stream) {			(mockReceiveStream as any).streamId = BigInt(this.serverUniStreamCounter);

			(mockReceiveStream as any).streamId = BigInt(this.serverUniStreamCounter);			this.serverUniStreamCounter += 4;

			this.serverUniStreamCounter += 4;		}

		}		this.incomingUniStreams.push(mockReceiveStream);

		this.incomingUniStreams.push(mockReceiveStream);		return mockReceiveStream;

		return mockReceiveStream;	}

	}}

}
