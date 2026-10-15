import type { Reader } from "@okdaichi/golikejs/io";
import { EOFError } from "@okdaichi/golikejs/io";
import { WebTransportStreamError, WebTransportStreamErrorCode } from "./error.ts";

export interface ReceiveStream extends Reader {
	cancel(code: WebTransportStreamErrorCode): Promise<void>;

	closed(): Promise<void>;
}
export interface ReceiveStreamInit {
	stream: ReadableStream<Uint8Array>;
}

/**
 * ReceiveStream wraps a WebTransport ReadableStream and implements the io.Reader interface.
 * This is a thin wrapper that reads data from the underlying stream with minimal internal buffering.
 */
class ReceiveStreamClass implements ReceiveStream {
	#pull: ReadableStreamDefaultReader<Uint8Array>;
	#buf: Uint8Array = new Uint8Array(0);
	#closed: Promise<void>;
	#err?: Error;

	constructor(init: ReceiveStreamInit) {
		this.#pull = init.stream.getReader();
		this.#closed = this.#pull.closed;
		// Track active ReceiveStreams for debugging
		try {
			(globalThis as any).__moq_openReceiveStreams =
				(globalThis as any).__moq_openReceiveStreams ?? 0;
			(globalThis as any).__moq_openReceiveStreams++;
			this.#closed.then(() => {
				(globalThis as any).__moq_openReceiveStreams--;
			}).catch(() => {
				// Ignore close errors for debugging counters
			});
		} catch (_e) {
			// ignore if not present
		}
	}

	/**
	 * Reads up to p.length bytes into p.
	 * Returns the number of bytes read and any error encountered.
	 * When EOF is reached, returns [n, EOFError].
	 * Implements io.Reader interface.
	 */
	async read(p: Uint8Array): Promise<[number, Error | undefined]> {
		if (this.#err) {
			return [0, this.#err];
		}

		// If we have buffered data, use it first
		if (this.#buf.length > 0) {
			const n = Math.min(p.length, this.#buf.length);
			p.set(this.#buf.subarray(0, n));
			// Release backing buffer when fully consumed (memory optimization)
			this.#buf = (n >= this.#buf.length) ? new Uint8Array(0) : this.#buf.subarray(n);
			return [n, undefined];
		}

		// Read from the underlying stream
		try {
			const value = await this.#nextChunk();
			if (value === undefined) {
				return [0, new EOFError()];
			}

			const n = Math.min(p.length, value.length);
			p.set(value.subarray(0, n));

			// Buffer any remaining data
			if (value.length > n) {
				this.#buf = value.subarray(n);
			}

			return [n, undefined];
		} catch (err) {
			if (this.#err) {
				return [0, this.#err];
			}
			this.#err = streamReadError(err);
			return [0, this.#err];
		}
	}

	// nextChunk returns the next non-empty chunk, or undefined at the end of
	// the stream. An empty chunk carries nothing, and returning 0 bytes with no
	// error for it would read as the end to readers such as readFull.
	async #nextChunk(): Promise<Uint8Array | undefined> {
		while (true) {
			const { done, value } = await this.#pull.read();
			if (done) {
				return undefined;
			}
			if (value.length > 0) {
				return value;
			}
		}
	}

	async cancel(code: WebTransportStreamErrorCode): Promise<void> {
		if (this.#err) {
			return;
		}
		// Wrap the numeric reason in a StreamError so that consumers can inspect
		// the structured error information (code, message, remote).
		const err = new WebTransportStreamError({
			source: "stream",
			streamErrorCode: code,
		}, false);
		this.#err = err;
		return this.#pull.cancel(err);
	}

	closed(): Promise<void> {
		return this.#closed;
	}
}

export const ReceiveStream: {
	new (init: ReceiveStreamInit): ReceiveStream;
} = ReceiveStreamClass;

// The first HTTP/3 error code WebTransport maps application stream error codes
// onto (draft-ietf-webtrans-http3, "Resetting Data Streams").
const http3ErrorCodeFirst = 0x52e4a40fa8db;

/**
 * Turns the error a stream read failed with into the error the reader
 * returns: a reset by the peer is a {@link WebTransportStreamError} with its
 * application code, whether the runtime reports it as a WebTransportError
 * (browsers) or as an error naming the HTTP/3 code (Deno).
 */
export function streamReadError(err: unknown): Error {
	if (typeof err === "object" && err !== null && "source" in err && err.source === "stream") {
		const code = (err as { streamErrorCode?: number | null }).streamErrorCode;
		if (typeof code === "number") {
			return new WebTransportStreamError({ source: "stream", streamErrorCode: code }, true);
		}
	}
	if (err instanceof Error) {
		const match = /stream reset by peer: error (\d+)/.exec(err.message);
		if (match) {
			const h3 = Number(match[1]);
			if (h3 >= http3ErrorCodeFirst) {
				const d = h3 - http3ErrorCodeFirst;
				// Every 0x1f-th code point is reserved (a GREASE value) and skipped.
				const code = d - Math.floor(d / 0x1f);
				return new WebTransportStreamError(
					{ source: "stream", streamErrorCode: code },
					true,
				);
			}
		}
		return err;
	}
	return new Error(`Failed to read from stream: ${err}`);
}
