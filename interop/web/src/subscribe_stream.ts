import { SubscribeMessage, SubscribeOkMessage, SubscribeUpdateMessage } from "./message";
import { Writer, Reader } from "./internal/io"
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
	#okFunc: (info: Info) => Promise<[SubscribeOkMessage?, Error?]>
	#update?: SubscribeUpdateMessage
	#reader: Reader
	#writer: Writer
	#ctx: Context;
	#cancelFunc: CancelCauseFunc;

	constructor(sessCtx: Context, writer: Writer, reader: Reader, subscribe: SubscribeMessage) {
		[this.#ctx, this.#cancelFunc] = withCancelCause(sessCtx);
		this.#reader = reader;
		this.#writer = writer;
		this.#subscribe = subscribe;
		this.#okFunc = async (info: Info): Promise<[SubscribeOkMessage?, Error?]> => {
			return SubscribeOkMessage.encode(writer, info.groupOrder);
		};
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

		const [_, flushErr] = await this.#writer.flush();
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
}

export class ReceiveSubscribeStream implements PublishController {
	#subscribe: SubscribeMessage
	#ok: SubscribeOkMessage
	#update?: SubscribeUpdateMessage
	#cond: Cond = new Cond();
	#reader: Reader
	#writer: Writer

	constructor(writer: Writer, reader: Reader, subscribe: SubscribeMessage, ok: SubscribeOkMessage) {
		this.#reader = reader
		this.#writer = writer
		this.#subscribe = subscribe
		this.#ok = ok

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

	async updated(): Promise<void> {
		return this.#cond.wait();
	}
}

export type SubscribeConfig = {
	trackPriority: bigint;
    minGroupSequence: bigint;
    maxGroupSequence: bigint;
}

export type SubscribeID = bigint;

