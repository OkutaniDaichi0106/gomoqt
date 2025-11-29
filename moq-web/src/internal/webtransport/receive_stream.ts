import type { Reader } from "@okudai/golikejs/io";
import { EOFError } from "@okudai/golikejs/io";
import { StreamError, StreamErrorCode, StreamErrorInfo } from "./error.ts";

export interface ReceiveStreamInit {
	stream: ReadableStream<Uint8Array>;
	streamId: bigint;
}

/**
 * ReceiveStream wraps a WebTransport ReadableStream and implements the io.Reader interface.
 * This is a thin wrapper that reads data from the underlying stream with minimal internal buffering.
 */
export class ReceiveStream implements Reader {
	#pull: ReadableStreamDefaultReader<Uint8Array>;
	#buf: Uint8Array = new Uint8Array(0);
	#closed: Promise<void>;
	readonly id: bigint;
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
		this.id = init.streamId;
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
			this.#buf = this.#buf.subarray(n);
			return [n, undefined];
		}

		// Read from the underlying stream
		try {
			const { done, value } = await this.#pull.read();
			if (done) {
				return [0, new EOFError()];
			}
			if (!value || value.length === 0) {
				return [0, undefined];
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

			if (err instanceof Error) {
				return [0, err];
			}
			const wtErr = err as WebTransportError;
			if (wtErr.source === "stream") {
				if (wtErr.streamErrorCode !== null) {
					this.#err = new StreamError(wtErr as StreamErrorInfo, true);
				} else {
					this.#err = new EOFError();
				}
			}
			return [0, new Error(`Failed to read from stream: ${err}`)];
		}
	}

	async cancel(code: StreamErrorCode): Promise<void> {
		if (this.#err) {
			return;
		}
		// Wrap the numeric reason in a StreamError so that consumers can inspect
		// the structured error information (code, message, remote).
		const err = new StreamError({ source: "stream", streamErrorCode: code }, false);
		this.#err = err;
		return this.#pull.cancel(err);
	}

	closed(): Promise<void> {
		return this.#closed;
	}
}
