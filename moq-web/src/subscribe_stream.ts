import { SubscribeMessage, SubscribeOkMessage, SubscribeUpdateMessage } from "./message";
import { Writer, Reader } from "./io"
import { Cond } from "./internal/cond";
import { CancelCauseFunc, Context, withCancelCause } from "./internal/context";
import { StreamError } from "./io/error";
import { Info } from "./info";

export class SendSubscribeStream {
	#config: TrackConfig;
	#id: SubscribeID;
	#reader: Reader
	#writer: Writer
	#ctx: Context;
	#cancelFunc: CancelCauseFunc;

	constructor(sessCtx: Context, writer: Writer, reader: Reader, subscribe: SubscribeMessage, ok: SubscribeOkMessage) {
		[this.#ctx, this.#cancelFunc] = withCancelCause(sessCtx);
		this.#reader = reader;
		this.#writer = writer;
		this.#config = subscribe;
		this.#id = subscribe.subscribeId;
	}

	get subscribeId(): SubscribeID {
		return this.#id;
	}

	get context(): Context {
		return this.#ctx;
	}

	get trackConfig(): TrackConfig {
		return this.#config;
	}

	async update(trackPriority: bigint, minGroupSequence: bigint, maxGroupSequence: bigint): Promise<Error | undefined> {
		const [result, err] = await SubscribeUpdateMessage.encode(this.#writer, trackPriority, minGroupSequence, maxGroupSequence);
		if (err) {
			return new Error(`Failed to write subscribe update: ${err}`);
		}
		this.#config = result!;

		const flushErr = await this.#writer.flush();
		if (flushErr) {
			return new Error(`Failed to flush subscribe update: ${flushErr}`);
		}
	}

	cancel(code: number, message: string): void {
        const err = new StreamError(code, message);
		this.#writer.cancel(err);
		this.#cancelFunc(err);
	}
}

export class ReceiveSubscribeStream {
	#subscribe: SubscribeMessage
	#update?: SubscribeUpdateMessage
	#cond: Cond = new Cond();
	#reader: Reader
	#writer: Writer
	#ok?: SubscribeOkMessage
	#ctx: Context;
	#cancelFunc: CancelCauseFunc;

	constructor(sessCtx: Context, writer: Writer, reader: Reader, subscribe: SubscribeMessage) {
		this.#reader = reader
		this.#writer = writer
		this.#subscribe = subscribe;
		[this.#ctx, this.#cancelFunc] = withCancelCause(sessCtx);

		// The async loop can be cancelled by sessCtx.done.
		(async () => {
			while (true) {
				const [msg, err] = await SubscribeUpdateMessage.decode(reader);
				if (err) {
					return;
				}
				this.#update = msg;
				this.#cond.broadcast();
			}
		})();
	}

	get subscribeId(): SubscribeID {
		return this.#subscribe.subscribeId;
	}

	get trackConfig(): TrackConfig {
		if (this.#update) {
			return this.#update;
		} else {
			return this.#subscribe;
		}
	}

	get context(): Context {
		return this.#ctx;
	}

	async updated(): Promise<void> {
		return this.#cond.wait();
	}

	async accept(info: Info): Promise<Error | undefined> {
		if (this.#ok) {
				return undefined; // Already accepted
			}

			const [msg, err] = await SubscribeOkMessage.encode(this.#writer, BigInt(info.groupOrder));
			if (err) {
				return new Error(`Failed to write subscribe ok: ${err}`);
			}
			if (!msg) {
				return new Error("Failed to encode subscribe ok message");
			}

			console.log(`Accepted subscribe ok with groupOrder: ${msg.groupOrder}`);

			this.#ok = msg;

			const flushErr = await this.#writer.flush();
			if (flushErr) {
				return new Error(`Failed to flush subscribe ok: ${flushErr}`);
			}
	}

	close(): void {
		const ctxErr = this.#ctx.err();
		if (ctxErr !== undefined) {
			// throw ctxErr;
			return;
		}
		this.#writer.close();
		this.#cancelFunc(undefined);
		this.#cond.broadcast(); // Notify any waiting threads that the stream is closed
	}

	closeWithError(code: number, message: string): void {
		const ctxErr = this.#ctx.err();
		if (ctxErr !== undefined) {
			// throw ctxErr;
			return;
		}
		const err = new StreamError(code, message);
		this.#writer.cancel(err);
		this.#cancelFunc(err);
		this.#cond.broadcast(); // Notify any waiting threads that the stream is closed
	}
}

export type TrackConfig = {
	trackPriority: bigint;
    minGroupSequence: bigint;
    maxGroupSequence: bigint;
}

export type SubscribeID = bigint;

