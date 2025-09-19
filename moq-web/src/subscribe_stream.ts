import type { SubscribeMessage} from "./message";
import { SubscribeOkMessage, SubscribeUpdateMessage } from "./message";
import type { Writer, Reader } from "./io"
import { Cond } from "./internal/cond";
import type { CancelCauseFunc, Context} from "./internal/context";
import { withCancelCause } from "./internal/context";
import { StreamError } from "./io/error";
import type { Info } from "./info";
import type { TrackPriority,GroupSequence,SubscribeID } from ".";

export interface TrackConfig {
	trackPriority: TrackPriority;
    minGroupSequence: GroupSequence;
    maxGroupSequence: GroupSequence;
}

export class SendSubscribeStream {
	#config: TrackConfig;
	#id: SubscribeID;
	#info: Info;
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
		this.#info = ok;
	}

	get subscribeId(): SubscribeID {
		return this.#id;
	}

	get context(): Context {
		return this.#ctx;
	}

	get config(): TrackConfig {
		return this.#config;
	}

	get info(): Info {
		return this.#info;
	}

	async update(update: TrackConfig): Promise<Error | undefined> {
		const msg = new SubscribeUpdateMessage(update);
		const err = await msg.encode(this.#writer);
		if (err) {
			return new Error(`Failed to write subscribe update: ${err}`);
		}
		this.#config = update;

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
	#info?: Info
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
				const msg = new SubscribeUpdateMessage({});
				let err: Error | undefined;
				err = await msg.decode(reader);
				if (err) {
					return; // TODO: Handle decode error
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

	async writeInfo(info?: Info): Promise<Error | undefined> {
		if (this.#info) {
			return undefined; // Info already written
		}

		const msg = new SubscribeOkMessage({});

		const err = await msg.encode(this.#writer);
		if (err) {
			return new Error(`Failed to write subscribe ok: ${err}`);
		}

		this.#info = msg;
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

