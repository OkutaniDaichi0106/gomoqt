import type { Writer } from "@okudai/golikejs/io";
import type { StreamError } from "./error.ts";

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

	constructor(init: SendStreamInit) {
		this.#writer = init.stream.getWriter();
		this.id = init.streamId;
	}

	/**
	 * Writes p to the underlying stream.
	 * Returns the number of bytes written and any error encountered.
	 * Implements io.Writer interface.
	 */
	async write(p: Uint8Array): Promise<[number, Error | undefined]> {
		try {
			await this.#writer.write(p);
			return [p.length, undefined];
		} catch (error) {
			if (error instanceof Error) {
				return [0, error];
			}
			return [0, new Error(`Failed to write to stream: ${error}`)];
		}
	}

	async close(): Promise<void> {
		await this.#writer.close();
	}

	async cancel(err: StreamError): Promise<void> {
		await this.#writer.abort(err);
	}

	closed(): Promise<void> {
		return this.#writer.closed;
	}
}

export interface Source {
	byteLength: number;
	copyTo(target: ArrayBuffer | ArrayBufferView<ArrayBufferLike>): void;
}
