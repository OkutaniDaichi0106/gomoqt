import type { SubscribeMessage} from "./message.ts";
import { SubscribeOkMessage, SubscribeUpdateMessage } from "./message.ts";
import type { Writer, Reader } from "./io.ts"
import { EOF } from "./io.ts"
import { Cond, Mutex } from "golikejs/sync";
import type { CancelCauseFunc, Context} from "golikejs/context";
import { withCancelCause } from "golikejs/context";
import { StreamError } from "./io/error.ts";
import type { Info } from "./info.ts";
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
	readonly context: Context;
	#cancelFunc: CancelCauseFunc;
	readonly streamId: bigint;

	constructor(sessCtx: Context, writer: Writer, reader: Reader, subscribe: SubscribeMessage, ok: SubscribeOkMessage) {
		[this.context, this.#cancelFunc] = withCancelCause(sessCtx);
		this.#reader = reader;
		this.#writer = writer;
		this.#config = {
			trackPriority: subscribe.trackPriority,
			minGroupSequence: subscribe.minGroupSequence,
			maxGroupSequence: subscribe.maxGroupSequence,
		};
		this.#id = subscribe.subscribeId;
		this.#info = ok;
		this.streamId = writer.streamId ?? reader.streamId ?? 0n;
	}

	get subscribeId(): SubscribeID {
		return this.#id;
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

	async closeWithError(code: number, message: string): Promise<void> {
        const err = new StreamError(code, message);
		await this.#writer.cancel(err);
		this.#cancelFunc(err);
	}
}

export class ReceiveSubscribeStream {
	readonly subscribeId: SubscribeID;
	#trackConfig: TrackConfig;
	#mu: Mutex = new Mutex();
	#cond: Cond = new Cond(this.#mu);
	#reader: Reader
	#writer: Writer
	#info?: Info
	readonly context: Context;
	#cancelFunc: CancelCauseFunc;
	readonly streamId: bigint;


	constructor(
		sessCtx: Context,
		writer: Writer,
		reader: Reader,
		subscribe: SubscribeMessage
	) {
		this.#reader = reader
		this.#writer = writer
		this.subscribeId = subscribe.subscribeId;
		this.#trackConfig = {
			trackPriority: subscribe.trackPriority,
			minGroupSequence: subscribe.minGroupSequence,
			maxGroupSequence: subscribe.maxGroupSequence,
		};
		this.streamId = writer.streamId ?? reader.streamId ?? 0n;
		[this.context, this.#cancelFunc] = withCancelCause(sessCtx);

		this.#handleUpdates();
	}

	async #handleUpdates(): Promise<void> {
		while (true) {
			const msg = new SubscribeUpdateMessage({});
			const err = await msg.decode(this.#reader);
			if (err) {
				if (err !== EOF ) {
					console.error(`moq: error reading SUBSCRIBE_UPDATE message for subscribe ID: ${this.subscribeId}: ${err}`);
				}
				return;
			}

			console.debug(`moq: SUBSCRIBE_UPDATE message received.`,
				{
					"subscribeId": this.subscribeId,
					"message": msg
				}
			);

			this.#trackConfig = {
				trackPriority: msg.trackPriority,
				minGroupSequence: msg.minGroupSequence,
				maxGroupSequence: msg.maxGroupSequence,
			};

			this.#cond.broadcast();
		}
	}

	get trackConfig(): TrackConfig {
		return this.#trackConfig;
	}

	async updated(): Promise<void> {
		return this.#cond.wait();
	}

	async writeInfo(info?: Info): Promise<Error | undefined> {
		if (this.#info) {
			console.warn(`Info already written for subscribe ID: ${this.subscribeId}`);
			return undefined; // Info already written
		}

		let err = this.context.err();
		if (err !== undefined) {
			return err;
		}

		const msg = new SubscribeOkMessage({});

		err = await msg.encode(this.#writer);
		if (err) {
			return new Error(`moq: failed to encode SUBSCRIBE_OK message: ${err}`);
		}

		console.debug(`moq: SUBSCRIBE_OK message sent.`,
			{
				"subscribeId": this.subscribeId,
				"message": msg
			}
		);

		this.#info = msg;
	}

	async close(): Promise<void> {
		if (this.context.err()) {
			return;
		}
		this.#cancelFunc(undefined);
		await this.#writer.close();

		this.#cond.broadcast();
	}

	async closeWithError(code: number, message: string): Promise<void> {
		if (this.context.err()) {
			return;
		}
		const cause = new StreamError(code, message);
		this.#cancelFunc(cause);
		await this.#writer.cancel(cause);
		await this.#reader.cancel(cause);

		this.#cond.broadcast();
	}
}

