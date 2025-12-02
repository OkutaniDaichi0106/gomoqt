import type { SubscribeMessage } from "./internal/message/mod.ts";
import { SubscribeOkMessage, SubscribeUpdateMessage } from "./internal/message/mod.ts";
import type { Stream } from "./internal/webtransport/mod.ts";
import { EOFError } from "@okudai/golikejs/io";
import { Cond, Mutex } from "@okudai/golikejs/sync";
import type { CancelCauseFunc, Context } from "@okudai/golikejs/context";
import { withCancelCause } from "@okudai/golikejs/context";
import { WebTransportStreamError } from "./internal/webtransport/mod.ts";
import type { Info } from "./info.ts";
import type { GroupSequence, SubscribeID, TrackPriority } from "./alias.ts";
import { SubscribeErrorCode } from "./error.ts";

export interface TrackConfig {
	trackPriority: TrackPriority;
	minGroupSequence: GroupSequence;
	maxGroupSequence: GroupSequence;
}

export class SendSubscribeStream {
	#config: TrackConfig;
	#id: SubscribeID;
	#info: Info;
	#stream: Stream;
	readonly context: Context;
	#cancelFunc: CancelCauseFunc;

	constructor(
		sessCtx: Context,
		stream: Stream,
		subscribe: SubscribeMessage,
		ok: SubscribeOkMessage,
	) {
		[this.context, this.#cancelFunc] = withCancelCause(sessCtx);
		this.#stream = stream;
		this.#config = {
			trackPriority: subscribe.trackPriority,
			minGroupSequence: subscribe.minGroupSequence,
			maxGroupSequence: subscribe.maxGroupSequence,
		};
		this.#id = subscribe.subscribeId;
		this.#info = ok;
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
		const err = await msg.encode(this.#stream.writable);
		if (err) {
			return new Error(`Failed to write subscribe update: ${err}`);
		}
		this.#config = update;

		return undefined;
	}

	async closeWithError(code: SubscribeErrorCode): Promise<void> {
		const err = new WebTransportStreamError({
			source: "stream",
			streamErrorCode: code,
		}, false);
		await this.#stream.writable.cancel(code);
		this.#cancelFunc(err);
	}
}

export class ReceiveSubscribeStream {
	readonly subscribeId: SubscribeID;
	#trackConfig: TrackConfig;
	#mu: Mutex = new Mutex();
	#cond: Cond = new Cond(this.#mu);
	#stream: Stream;
	#info?: Info;
	readonly context: Context;
	#cancelFunc: CancelCauseFunc;

	constructor(
		sessCtx: Context,
		stream: Stream,
		subscribe: SubscribeMessage,
	) {
		this.#stream = stream;
		this.subscribeId = subscribe.subscribeId;
		this.#trackConfig = {
			trackPriority: subscribe.trackPriority,
			minGroupSequence: subscribe.minGroupSequence,
			maxGroupSequence: subscribe.maxGroupSequence,
		};
		[this.context, this.#cancelFunc] = withCancelCause(sessCtx);

		this.#handleUpdates();
	}

	async #handleUpdates(): Promise<void> {
		while (true) {
			const msg = new SubscribeUpdateMessage({});
			const err = await msg.decode(this.#stream.readable);
			if (err) {
				if (err instanceof EOFError) {
					console.error(
						`moq: error reading SUBSCRIBE_UPDATE message for subscribe ID: ${this.subscribeId}: ${err}`,
					);
				}
				return;
			}

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

	async writeInfo(_?: Info): Promise<Error | undefined> {
		if (this.#info) {
			console.warn(
				`Info already written for subscribe ID: ${this.subscribeId}`,
			);
			return undefined; // Info already written
		}

		let err = this.context.err();
		if (err !== undefined) {
			return err;
		}

		const msg = new SubscribeOkMessage({});

		err = await msg.encode(this.#stream.writable);
		if (err) {
			return new Error(`moq: failed to encode SUBSCRIBE_OK message: ${err}`);
		}

		this.#info = msg;

		return undefined;
	}

	async close(): Promise<void> {
		if (this.context.err()) {
			return;
		}
		this.#cancelFunc(undefined);
		await this.#stream.writable.close();

		this.#cond.broadcast();
	}

	async closeWithError(code: SubscribeErrorCode): Promise<void> {
		if (this.context.err()) {
			return;
		}
		const cause = new WebTransportStreamError(
			{ source: "stream", streamErrorCode: code },
			false,
		);
		this.#cancelFunc(cause);
		await this.#stream.writable.cancel(code);
		await this.#stream.readable.cancel(code);

		this.#cond.broadcast();
	}
}
