import type { ReceiveStream, SendStream } from "./internal/webtransport/mod.ts";
import { withCancelCause } from "@okudai/golikejs/context";
import type { CancelCauseFunc, Context } from "@okudai/golikejs/context";
import { WebTransportStreamError } from "./internal/webtransport/error.ts";
import type { GroupMessage } from "./internal/message/mod.ts";
import { readFull, readVarint, writeVarint } from "./internal/message/mod.ts";
import type { Frame } from "./frame.ts";
import { GroupErrorCode } from "./error.ts";
import { GroupSequence } from "./alias.ts";

export const GroupSequenceFirst: GroupSequence = 1;

export class GroupWriter {
	readonly sequence: GroupSequence;
	#stream: SendStream;
	readonly context: Context;
	#cancelFunc: CancelCauseFunc;

	constructor(trackCtx: Context, writer: SendStream, group: GroupMessage) {
		this.sequence = group.sequence;
		this.#stream = writer;
		[this.context, this.#cancelFunc] = withCancelCause(trackCtx);

		trackCtx.done().then(() => {
			this.cancel(GroupErrorCode.SubscribeCanceled);
		});
	}

	async writeFrame(src: Frame): Promise<Error | undefined> {
		// Write length prefix
		let [, err] = await writeVarint(this.#stream, src.data.byteLength);
		if (err) {
			return err;
		}
		// Write data
		[, err] = await this.#stream.write(src.data);
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

	async cancel(code: GroupErrorCode): Promise<void> {
		if (this.context.err()) {
			// Do nothing if already cancelled
			return;
		}
		const cause = new WebTransportStreamError(
			{ source: "stream", streamErrorCode: code },
			false,
		);
		this.#cancelFunc(cause); // Notify the context about cancellation
		await this.#stream.cancel(code);
	}
}

export class GroupReader {
	readonly sequence: GroupSequence;
	#reader: ReceiveStream;
	readonly context: Context;
	#cancelFunc: CancelCauseFunc;

	constructor(trackCtx: Context, reader: ReceiveStream, group: GroupMessage) {
		this.sequence = group.sequence;
		this.#reader = reader;
		[this.context, this.#cancelFunc] = withCancelCause(trackCtx);

		trackCtx.done().then(() => {
			this.cancel(GroupErrorCode.PublishAborted);
		});
	}

	async readFrame(frame: Frame): Promise<Error | undefined> {
		const [len, , err1] = await readVarint(this.#reader);
		if (err1) {
			return err1;
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

		const [, err2] = await readFull(this.#reader, frame.data.subarray(0, len));
		if (err2) {
			return err2;
		}

		return undefined;
	}

	async cancel(code: GroupErrorCode): Promise<void> {
		if (this.context.err()) {
			// Do nothing if already cancelled
			return;
		}
		const reason = new WebTransportStreamError(
			{ source: "stream", streamErrorCode: code },
			false,
		);
		this.#cancelFunc(reason);
		await this.#reader.cancel(code);
	}
}
