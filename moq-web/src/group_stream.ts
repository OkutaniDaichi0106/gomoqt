import type { ReceiveStream, SendStream } from "./internal/webtransport/mod.ts";
import { withCancelCause } from "@okudai/golikejs/context";
import type { CancelCauseFunc, Context } from "@okudai/golikejs/context";
import { StreamError } from "./internal/webtransport/error.ts";
import type { GroupMessage } from "./internal/message/mod.ts";
import type { Frame } from "./frame.ts";
import type { GroupErrorCode } from "./error.ts";
import { PublishAbortedErrorCode, SubscribeCanceledErrorCode } from "./error.ts";

export class GroupWriter {
	readonly sequence: bigint;
	#stream: SendStream;
	readonly context: Context;
	#cancelFunc: CancelCauseFunc;

	constructor(trackCtx: Context, writer: SendStream, group: GroupMessage) {
		this.sequence = group.sequence;
		this.#stream = writer;
		[this.context, this.#cancelFunc] = withCancelCause(trackCtx);

		trackCtx.done().then(() => {
			this.cancel(SubscribeCanceledErrorCode, "track was closed");
		});
	}

	async writeFrame(src: Frame): Promise<Error | undefined> {
		this.#stream.copyFrom(src);
		const err = await this.#stream.flush();
		if (err) {
			return err;
		}

		return undefined;
	}

	async close(): Promise<void> {
		if (this.context.err()) {
			return;
		}
		this.#cancelFunc(undefined); // Notify the context about closure
		await this.#stream.close();
	}

	async cancel(code: GroupErrorCode, message: string): Promise<void> {
		if (this.context.err()) {
			// Do nothing if already cancelled
			return;
		}
		const cause = new StreamError(code, message);
		this.#cancelFunc(cause); // Notify the context about cancellation
		await this.#stream.cancel(cause);
	}
}

export class GroupReader {
	readonly sequence: bigint;
	#reader: ReceiveStream;
	readonly context: Context;
	#cancelFunc: CancelCauseFunc;
	// #frame?: BytesFrame;

	constructor(trackCtx: Context, reader: ReceiveStream, group: GroupMessage) {
		this.sequence = group.sequence;
		this.#reader = reader;
		[this.context, this.#cancelFunc] = withCancelCause(trackCtx);

		trackCtx.done().then(() => {
			this.cancel(PublishAbortedErrorCode, "track was closed");
		});
	}

	async readFrame(frame: Frame): Promise<Error | undefined> {
		let err: Error | undefined;
		let len: number;
		[len, err] = await this.#reader.readVarint();
		if (err) {
			return err;
		}

		if (len > Number.MAX_SAFE_INTEGER) {
			return new Error("Varint too large");
		}

		if (frame.data.byteLength < len) {
			const currentSize = frame.data.byteLength || 0;
			const cap = Math.max(currentSize * 2, len);
			// Swap buffers
			frame.data = new Uint8Array(cap);
		}

		err = await this.#reader.fillN(frame.data, len);
		if (err) {
			return err;
		}

		return undefined;
	}

	async cancel(code: GroupErrorCode, message: string): Promise<void> {
		if (this.context.err()) {
			// Do nothing if already cancelled
			return;
		}
		const reason = new StreamError(code, message);
		this.#cancelFunc(reason);
		await this.#reader.cancel(reason);
	}
}
