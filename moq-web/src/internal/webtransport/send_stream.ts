import { EOFError, type Writer } from "@okudai/golikejs/io";
import { StreamError, StreamErrorInfo } from "./error.ts";
import { StreamErrorCode } from "./mod.ts";

export interface SendStreamInit {
	stream: WritableStream<Uint8Array>;
	streamId: bigint;
}

/**
 * SendStream wraps a WebTransport WritableStream and implements the io.Writer interface.
 * This is a thin wrapper that passes data directly to the underlying stream without internal buffering.
 */
export class SendStream implements Writer {
	#writer: WritableStreamDefaultWriter<Uint8Array>;
	readonly id: bigint;

	#err?: Error;

	constructor(init: SendStreamInit) {
		this.#writer = init.stream.getWriter();
		this.id = init.streamId;
		try {
			(globalThis as any).__moq_openSendStreams = (globalThis as any).__moq_openSendStreams ??
				0;
			(globalThis as any).__moq_openSendStreams++;
			this.closed().then(() => {
				(globalThis as any).__moq_openSendStreams--;
			}).catch(() => {
				// Ignore close errors - we don't want unhandled rejections during tests
			});
		} catch (_e) {
			// ignore
		}
	}

	/**
	 * Writes p to the underlying stream.
	 * Returns the number of bytes written and any error encountered.
	 * Implements io.Writer interface.
	 */
	async write(p: Uint8Array): Promise<[number, Error | undefined]> {
		if (this.#err) {
			return [0, this.#err];
		}
		try {
			await this.#writer.write(p);
			return [p.length, undefined];
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
			return [0, this.#err];
		}
	}

	async close(): Promise<void> {
		await this.#writer.close();
	}

	async cancel(code: StreamErrorCode): Promise<void> {
		if (this.#err) {
			return;
		}
		const wtErr: StreamErrorInfo = {
			source: "stream",
			streamErrorCode: code,
		};
		const err = new StreamError(wtErr, false);
		this.#err = err;
		await this.#writer.abort(err);
	}

	closed(): Promise<void> {
		return this.#writer.closed;
	}
}

export interface Source {
	byteLength: number;
	copyTo(target: ArrayBuffer | ArrayBufferView): void;
}
