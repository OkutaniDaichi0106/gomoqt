import { GroupReader, GroupWriter } from "./group_stream.ts";
import type { Info } from "./info.ts";
import type { Context } from "@okudai/golikejs/context";
import { watchPromise } from "@okudai/golikejs/context";
import type {
	ReceiveSubscribeStream,
	SendSubscribeStream,
	TrackConfig,
} from "./subscribe_stream.ts";
import type { ReceiveStream, SendStream } from "./internal/webtransport/mod.ts";
import { UniStreamTypes } from "./stream_type.ts";
import { GroupMessage, writeVarint } from "./internal/message/mod.ts";
import type { BroadcastPath } from "./broadcast_path.ts";
import type { SubscribeErrorCode } from "./error.ts";
import { PublishAbortedErrorCode } from "./error.ts";
import type { GroupSequence } from "./alias.ts";

export class TrackWriter {
	broadcastPath: BroadcastPath;
	trackName: string;
	#subscribeStream: ReceiveSubscribeStream;
	#openUniStreamFunc: () => Promise<[SendStream, undefined] | [undefined, Error]>;
	#groups: GroupWriter[] = [];

	constructor(
		broadcastPath: BroadcastPath,
		trackName: string,
		subscribeStream: ReceiveSubscribeStream,
		openUniStreamFunc: () => Promise<[SendStream, undefined] | [undefined, Error]>,
	) {
		this.broadcastPath = broadcastPath;
		this.trackName = trackName;
		this.#subscribeStream = subscribeStream;
		this.#openUniStreamFunc = openUniStreamFunc;
	}

	get context(): Context {
		return this.#subscribeStream.context;
	}

	get subscribeId(): number {
		return this.#subscribeStream.subscribeId;
	}

	get config(): TrackConfig {
		return this.#subscribeStream.trackConfig;
	}

	async openGroup(
		groupSequence: GroupSequence,
	): Promise<[GroupWriter, undefined] | [undefined, Error]> {
		let err: Error | undefined;
		err = await this.#subscribeStream.writeInfo();
		if (err) {
			return [undefined, err];
		}

		let writer: SendStream | undefined;
		[writer, err] = await this.#openUniStreamFunc();
		if (err) {
			return [undefined, err];
		}

		[, err] = await writeVarint(writer!, UniStreamTypes.GroupStreamType);
		if (err) {
			return [undefined, new Error(`Failed to write stream type: ${err}`)];
		}

		const msg = new GroupMessage({
			subscribeId: this.subscribeId,
			sequence: groupSequence,
		});
		err = await msg.encode(writer!);
		if (err) {
			return [undefined, new Error("Failed to create group message")];
		}

		const group = new GroupWriter(this.context, writer!, msg);

		this.#groups.push(group);

		return [group, undefined];
	}

	async writeInfo(info: Info): Promise<Error | undefined> {
		const err = await this.#subscribeStream.writeInfo(info);
		if (err) {
			return err;
		}

		return undefined;
	}

	async closeWithError(code: SubscribeErrorCode, message: string): Promise<void> {
		// Cancel all groups with the error first
		await Promise.allSettled(this.#groups.map(
			(group) => group.cancel(PublishAbortedErrorCode, message),
		));

		// Then close the subscribe stream with the error
		await this.#subscribeStream.closeWithError(code, message);
	}

	async close(): Promise<void> {
		await Promise.allSettled(this.#groups.map(
			(group) => group.close(),
		));

		await this.#subscribeStream.close();
	}
}

export class TrackReader {
	broadcastPath: BroadcastPath;
	trackName: string;
	#subscribeStream: SendSubscribeStream;
	#acceptFunc: (ctx: Promise<void>) => Promise<[ReceiveStream, GroupMessage] | undefined>;
	#onCloseFunc: () => void;

	constructor(
		broadcastPath: BroadcastPath,
		trackName: string,
		subscribeStream: SendSubscribeStream,
		acceptFunc: (ctx: Promise<void>) => Promise<[ReceiveStream, GroupMessage] | undefined>,
		onCloseFunc: () => void,
	) {
		this.broadcastPath = broadcastPath;
		this.trackName = trackName;
		this.#subscribeStream = subscribeStream;
		this.#acceptFunc = acceptFunc;
		this.#onCloseFunc = onCloseFunc;
	}

	async acceptGroup(
		signal: Promise<void>,
	): Promise<[GroupReader, undefined] | [undefined, Error]> {
		// Check if context is already cancelled
		const err = this.context.err();
		if (err) {
			return [undefined, err];
		}

		const ctx = watchPromise(this.context, signal);

		const dequeued = await this.#acceptFunc(ctx.done());
		if (dequeued === undefined) {
			return [undefined, new Error("[TrackReader] failed to dequeue group message")];
		}

		const [reader, msg] = dequeued;

		const group = new GroupReader(this.context, reader, msg);

		return [group, undefined];
	}

	async update(config: TrackConfig): Promise<Error | undefined> {
		return this.#subscribeStream.update(config);
	}

	readInfo(): Info {
		return this.#subscribeStream.info;
	}

	async closeWithError(code: number, message: string): Promise<void> {
		await this.#subscribeStream.closeWithError(code, message);
		this.#onCloseFunc();
	}

	get trackConfig(): TrackConfig {
		return this.#subscribeStream.config;
	}

	get context(): Context {
		return this.#subscribeStream.context;
	}
}
