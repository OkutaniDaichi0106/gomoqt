import { SubscribeMessage, SubscribeOkMessage, SubscribeUpdateMessage } from "./message";
import { Writer, Reader } from "./io"
import { Cond } from "./internal/cond";
import { CancelCauseFunc, Context, withCancelCause } from "./internal/context";
import { StreamError } from "./io/error";
import { Info } from "./info";

export interface SubscribeController {
	subscribeId: SubscribeID
	subscribeConfig: SubscribeConfig
	update(trackPriority: bigint, min: bigint, max: bigint): Promise<void>
	context: Context
}

export class SendSubscribeStream implements SubscribeController {
	#subscribe: SubscribeMessage
	#ok: SubscribeOkMessage
	#update?: SubscribeUpdateMessage
	#reader: Reader
	#writer: Writer
	#ctx: Context;
	#cancelFunc: CancelCauseFunc;

	constructor(sessCtx: Context, writer: Writer, reader: Reader, subscribe: SubscribeMessage, ok: SubscribeOkMessage) {
		[this.#ctx, this.#cancelFunc] = withCancelCause(sessCtx);
		this.#reader = reader;
		this.#writer = writer;
		this.#subscribe = subscribe;
		this.#ok = ok;
	}

	get subscribeId(): SubscribeID {
		return this.#subscribe.subscribeId;
	}

	get context(): Context {
		return this.#ctx;
	}

	get subscribeConfig(): SubscribeConfig {
		if (this.#update) {
			return this.#update;
		} else {
			return this.#subscribe;
		}
	}
	
	async update(trackPriority: bigint, minGroupSequence: bigint, maxGroupSequence: bigint): Promise<void> {
		const [result, err] = await SubscribeUpdateMessage.encode(this.#writer, trackPriority, minGroupSequence, maxGroupSequence);
		if (err) {
			throw new Error(`Failed to write subscribe update: ${err}`);
		}
		this.#update = result!;

		const flushErr = await this.#writer.flush();
		if (flushErr) {
			throw new Error(`Failed to flush subscribe update: ${flushErr}`);
		}
	}

	cancel(code: number, message: string): void {
        const err = new StreamError(code, message);
		this.#writer.cancel(err);
		this.#cancelFunc(err);
	}
}

export interface PublishController {
	subscribeId: SubscribeID
	updated(): Promise<void>;
	subscribeConfig: SubscribeConfig
	context: Context
	accept(info: Info): Promise<void>;
	close(): void;
	closeWithError(code: number, message: string): void;
}

export class ReceiveSubscribeStream implements PublishController {
	#subscribe: SubscribeMessage
	#update?: SubscribeUpdateMessage
	#cond: Cond = new Cond();
	#reader: Reader
	#writer: Writer
	#acceptFunc: (info: Info) => Promise<[SubscribeOkMessage?, Error?]>
	#ok?: SubscribeOkMessage
	#ctx: Context;
	#cancelFunc: CancelCauseFunc;

	constructor(sessCtx: Context, writer: Writer, reader: Reader, subscribe: SubscribeMessage) {
		this.#reader = reader
		this.#writer = writer
		this.#subscribe = subscribe
		this.#acceptFunc = async (info: Info): Promise<[SubscribeOkMessage?, Error?]> => {
			return await SubscribeOkMessage.encode(this.#writer, info.groupOrder);		
		}
		[this.#ctx, this.#cancelFunc] = withCancelCause(sessCtx);

		async () => {
			for (;;) {
				const [msg, err] = await SubscribeUpdateMessage.decode(this.#reader)
				if (err) {
					// TODO: handle this situation
					break
				}

				this.#update = msg!

				this.#cond.broadcast();
			}
		}
	}

	get subscribeId(): SubscribeID {
		return this.#subscribe.subscribeId;
	}

	get subscribeConfig(): SubscribeConfig {
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

	async accept(info: Info): Promise<void> {
		const [result, err] = await this.#acceptFunc(info);
		if (err) {
			throw new Error(`Failed to write subscribe ok: ${err}`);
		}
		if (!result) {
			throw new Error("Failed to encode subscribe ok message");
		}

		this.#acceptFunc = async () => [result, undefined]; // TODO: No re-accept for updates?

		this.#ok = result;

		const flushErr = await this.#writer.flush();
		if (flushErr) {
			throw new Error(`Failed to flush subscribe ok: ${flushErr}`);
		}
	}

	close(): void {
		if (this.#ctx.err() !== null) {
			throw this.#ctx.err();
		}
		this.#writer.close();
		this.#cancelFunc(null);
		this.#cond.broadcast(); // Notify any waiting threads that the stream is closed
	}

	closeWithError(code: number, message: string): void {
		if (this.#ctx.err() !== null) {
			throw this.#ctx.err();
		}
		const err = new StreamError(code, message);
		this.#writer.cancel(err);
		this.#cancelFunc(err);
		this.#cond.broadcast(); // Notify any waiting threads that the stream is closed
	}
}

export type SubscribeConfig = {
	trackPriority: bigint;
    minGroupSequence: bigint;
    maxGroupSequence: bigint;
}

export type SubscribeID = bigint;

