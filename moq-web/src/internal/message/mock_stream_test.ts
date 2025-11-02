/**
 * Mock utilities for testing message encode/decode operations.
 * 
 * This module provides helper functions to create streams for testing
 * message serialization and deserialization.
 * 
 * @example
 * ```ts
 * import { transformStreams } from "./mock_stream_test.ts";
 * 
 * const { writer, reader } = transformStreams();
 * await message.encode(writer);
 * await writer.close();
 * await message.decode(reader);
 * ```
 * 
 * @module
 */

import { SendStream } from "../webtransport/send_stream.ts";
import { ReceiveStream } from "../webtransport/receive_stream.ts";

/**
 * Create a pair of connected send/receive streams for testing encode/decode roundtrip.
 * Uses TransformStream internally to connect the writer and reader.
 * 
 * @returns Object containing writer (SendStream) and reader (ReceiveStream)
 */
export function transformStreams(): {
	writer: SendStream;
	reader: ReceiveStream;
} {
	const { writable, readable } = new TransformStream<Uint8Array>();
	const writer = new SendStream({ stream: writable, transfer: undefined, streamId: 0n });
	const reader = new ReceiveStream({ stream: readable, transfer: undefined, streamId: 0n });
	return { writer, reader };
}
