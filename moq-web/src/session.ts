import { DEFAULT_CLIENT_VERSIONS } from "./version.ts";
import type { Version } from "./version.ts";
import {
	AnnounceInitMessage,
	AnnouncePleaseMessage,
	GroupMessage,
	SessionClientMessage,
	SessionServerMessage,
	SubscribeMessage,
	SubscribeOkMessage,
} from "./internal/message/mod.ts";
import { Stream } from "./internal/webtransport/mod.ts";
import { Extensions } from "./extensions.ts";
import { SessionStream } from "./session_stream.ts";
import { background, withCancelCause } from "@okudai/golikejs/context";
import type { CancelCauseFunc, Context } from "@okudai/golikejs/context";
import { AnnouncementReader, AnnouncementWriter } from "./announce_stream.ts";
import type { TrackPrefix } from "./track_prefix.ts";
import { ReceiveSubscribeStream, SendSubscribeStream } from "./subscribe_stream.ts";
import type { TrackConfig } from "./subscribe_stream.ts";
import type { BroadcastPath } from "./broadcast_path.ts";
import { TrackReader, TrackWriter } from "./track.ts";
import type { TrackMux } from "./track_mux.ts";
import { DefaultTrackMux } from "./track_mux.ts";
import { BiStreamTypes, UniStreamTypes } from "./stream_type.ts";
import { Queue } from "./internal/queue.ts";
import type { SubscribeID, TrackName } from "./alias.ts";
import type { ReceiveStream } from "./internal/webtransport/receive_stream.ts";
import { Connection } from "./internal/webtransport/connection.ts";

export interface SessionInit {
	conn: WebTransport;
	versions?: Set<Version>;
	extensions?: Extensions;
	mux?: TrackMux;
}

export class Session {
	readonly ready: Promise<void>;
	#conn: Connection;
	#sessionStream!: SessionStream;
	#ctx: Context;
	#cancelFunc: CancelCauseFunc;

	#wg: Promise<void>[] = [];

	#subscribeIDCounter: number = 0;

	// #biStreamCounter: number = 0; // client bidirectional stream counter

	// #serverBiStreamCounter: number = 1;

	// #uniStreamCounter: number = 2;

	// #serverUniStreamCounter: number = 3;

	readonly mux: TrackMux;

	#enqueueFuncs: Map<SubscribeID, (stream: ReceiveStream, msg: GroupMessage) => void> = new Map();

	constructor(init: SessionInit) {
		this.#conn = new Connection(init.conn);
		this.mux = init.mux ?? DefaultTrackMux;
		const [ctx, cancel] = withCancelCause(background());
		this.#conn.closed.finally(() => {
			cancel(new Error("webtransport: connection closed"));
		});
		this.#ctx = ctx;
		this.#cancelFunc = cancel;

		this.ready = this.#setup(
			init.versions ?? DEFAULT_CLIENT_VERSIONS,
			init.extensions ?? new Extensions(),
		);
	}

	async #setup(versions: Set<Version>, extensions: Extensions): Promise<void> {
		await this.#conn.ready;

		const [stream, openErr] = await this.#conn.openStream();
		if (openErr) {
			console.error("moq: failed to open session stream:", openErr);
			throw openErr;
		}
		// Send STREAM_TYPE
		stream.writable.writeUint8(BiStreamTypes.SessionStreamType);
		let err = await stream.writable.flush();
		if (err) {
			console.error("moq: failed to open session stream:", err);
			throw err;
		}

		// Send the session client message
		const req = new SessionClientMessage({
			versions,
			extensions: extensions.entries,
		});
		err = await req.encode(stream.writable);
		if (err) {
			console.error("moq: failed to send SESSION_CLIENT message:", err);
			throw err;
		}

		console.debug("moq: SESSION_CLIENT message sent.", {
			"message": req,
			"streamId": stream.id,
		});

		// Receive the session server message
		const rsp = new SessionServerMessage({});
		err = await rsp.decode(stream.readable);
		if (err) {
			console.error("moq: failed to receive SESSION_SERVER message:", err);
			throw err;
		}

		console.debug("moq: SESSION_SERVER message received.", {
			"message": rsp,
		});

		// TODO: Check the version compatibility
		if (!versions.has(rsp.version)) {
			throw new Error(`Incompatible session version: ${rsp.version}`);
		}


		this.#sessionStream = new SessionStream({
			context: this.#ctx,
			stream: stream,
			client: req,
			server: rsp,
			detectFunc: async () => {
				// Block until the connection is closed
				// TODO: Implement actual bitrate detection logic
				await this.#ctx.done();
				return 0; // Placeholder for bitrate detection logic
			},
		});

		this.#sessionStream.context.done().then(()=>{
			this.#cancelFunc(new Error("moq: session stream closed"));
		});

		// Start listening for incoming streams
		this.#wg.push(this.#listenBiStreams());
		this.#wg.push(this.#listenUniStreams());

		return;
	}

	async acceptAnnounce(
		prefix: TrackPrefix,
	): Promise<[AnnouncementReader, undefined] | [undefined, Error]> {
		const [stream, openErr] = await this.#conn.openStream();
		if (openErr) {
			console.error("moq: failed to open announce stream:", openErr);
			return [undefined, openErr];
		}
		// Send STREAM_TYPE
		stream.writable.writeUint8(BiStreamTypes.AnnounceStreamType);
		let err = await stream.writable.flush();
		if (err) {
			console.error("moq: failed to open announce stream:", err);
			return [undefined, err];
		}

		// Send AnnouncePleaseMessage
		const req = new AnnouncePleaseMessage({ prefix });
		err = await req.encode(stream.writable);
		if (err) {
			console.error("moq: failed to send ANNOUNCE_PLEASE message:", err);
			return [undefined, err];
		}

		console.debug(`moq: ANNOUNCE_PLEASE message sent.`, {
			"message": req,
			"streamId": stream.id,
		});

		// Receive AnnounceInitMessage
		const rsp = new AnnounceInitMessage({});
		err = await rsp.decode(stream.readable);
		if (err) {
			console.error("moq: failed to receive ANNOUNCE_INIT message:", err);
			return [undefined, err];
		}

		console.debug(`moq: ANNOUNCE_INIT message received.`, {
			"prefix": prefix,
			"message": rsp,
		});

		return [new AnnouncementReader(this.#ctx, stream, req, rsp), undefined];
	}

	async subscribe(
		path: BroadcastPath,
		name: TrackName,
		config?: TrackConfig,
	): Promise<[TrackReader, undefined] | [undefined, Error]> {
		const [stream, openErr] = await this.#conn.openStream();
		if (openErr) {
			console.error("moq: failed to open subscribe stream:", openErr);
			return [undefined, openErr];
		}
		// Send STREAM_TYPE
		stream.writable.writeUint8(BiStreamTypes.SubscribeStreamType);
		let err = await stream.writable.flush();
		if (err) {
			console.error("moq: failed to open subscribe stream:", err);
			return [undefined, err];
		}

		// Send SUBSCRIBE message
		const req = new SubscribeMessage({
			subscribeId: this.#subscribeIDCounter++,
			broadcastPath: path,
			trackName: name,
			trackPriority: config?.trackPriority ?? 0,
			minGroupSequence: config?.minGroupSequence ?? 0n,
			maxGroupSequence: config?.maxGroupSequence ?? 0n,
		});
		err = await req.encode(stream.writable);
		if (err) {
			console.error("moq: failed to send SUBSCRIBE message:", err);
			return [undefined, err];
		}

		console.debug(`moq: SUBSCRIBE message sent.`, {
			"message": req,
			"streamId": stream.id,
		});

		const rsp = new SubscribeOkMessage({});
		err = await rsp.decode(stream.readable);
		if (err) {
			console.error("moq: failed to receive SUBSCRIBE_OK message:", err);
			return [undefined, err];
		}

		console.debug(`moq: SUBSCRIBE_OK message received.`, {
			"subscribeId": req.subscribeId,
			"message": req,
		});

		const subscribeStream = new SendSubscribeStream(this.#ctx, stream, req, rsp);

		const queue = new Queue<[ReceiveStream, GroupMessage]>();

		// Add the enqueue function to the map
		this.#enqueueFuncs.set(req.subscribeId, (stream, msg) => {
			queue.enqueue([stream, msg]);
		});

		const track = new TrackReader(
			subscribeStream,
			async (ctx: Promise<void>) => {
				const result = await Promise.race([
					ctx,
					this.#ctx.done(),
					queue.dequeue(),
				]);

				if (!result) {
					return undefined;
				}

				return result;
			},
			() => {
				this.#enqueueFuncs.delete(req.subscribeId);
			},
		);

		return [track, undefined];
	}

	async #handleGroupStream(reader: ReceiveStream): Promise<void> {
		const req = new GroupMessage({});
		const err = await req.decode(reader);
		if (err) {
			console.error("Failed to decode GroupMessage:", err);
			return;
		}

		console.debug("moq: GROUP message received.", {
			"message": req,
			"streamId": reader.id,
		});

		const enqueueFunc = this.#enqueueFuncs.get(req.subscribeId);
		if (!enqueueFunc) {
			console.error(`moq: no subscription found for Subscribe ID: ${req.subscribeId}`);
			return;
		}

		enqueueFunc(reader, req);
	}

	async #handleSubscribeStream(stream: Stream): Promise<void> {
		const req = new SubscribeMessage({});
		const reqErr = await req.decode(stream.readable);
		if (reqErr) {
			console.error("Failed to decode SubscribeMessage:", reqErr);
			return;
		}

		console.debug("moq: SUBSCRIBE message received.", {
			"message": req,
			"streamId": stream.id,
		});

		const subscribeStream = new ReceiveSubscribeStream(this.#ctx, stream, req);

		// const openUniStreamFunc = async (): Promise<[SendStream, undefined] | [undefined, Error]> => {
		// 	try {
		// 		const stream = await this.#conn.openUniStream();
		// 		return [stream, undefined];
		// 	} catch (err) {
		// 		console.error("moq: failed to create unidirectional stream:", err);
		// 		return [
		// 			undefined,
		// 			new Error(`moq: failed to create unidirectional stream: ${err}`),
		// 		];
		// 	}
		// };

		const trackWriter = new TrackWriter(
			req.broadcastPath as BroadcastPath,
			req.trackName,
			subscribeStream,
			this.#conn.openUniStream.bind(this.#conn),
		);

		this.mux.serveTrack(trackWriter);
	}

	async #handleAnnounceStream(stream: Stream): Promise<void> {
		const req = new AnnouncePleaseMessage({});
		const err = await req.decode(stream.readable);
		if (err) {
			console.error("Failed to decode AnnouncePleaseMessage:", err);
			return;
		}

		console.debug("moq: ANNOUNCE_PLEASE message received.", {
			"message": req,
			"streamId": stream.id,
		});

		const aw = new AnnouncementWriter(this.#ctx, stream, req);

		this.mux.serveAnnouncement(aw, aw.prefix);
	}

	async #listenBiStreams(): Promise<void> {
		try {
			// Handle incoming streams
			let num: number | undefined;
			let err: Error | undefined;
			while (true) {
				const [stream, acceptErr] = await this.#conn.acceptStream();
				// biStreams.releaseLock(); // Release the lock after reading
				if (acceptErr) {
					console.error("Bidirectional stream closed", acceptErr);
					break;
				}
				[num, err] = await stream.readable.readUint8();
				if (err) {
					console.error("Failed to read from bidirectional stream:", err);
					continue;
				}

				switch (num) {
					case BiStreamTypes.SubscribeStreamType:
						this.#handleSubscribeStream(stream);
						break;
					case BiStreamTypes.AnnounceStreamType:
						this.#handleAnnounceStream(stream);
						break;
					default:
						console.warn(`Unknown bidirectional stream type: ${num}`);
						break; // Ignore unknown stream types
				}
			}
		} catch (error) {
			console.error("Error in listenBiStreams:", error);
		}
	}

	async #listenUniStreams(): Promise<void> {
		try {
			let num: number | undefined;
			let err: Error | undefined;
			while (true) {
				const [stream, acceptErr] = await this.#conn.acceptUniStream();
				if (acceptErr) {
					console.error("Unidirectional stream closed", acceptErr);
					break;
				}

				// Read the first byte to determine the stream type
				[num, err] = await stream.readUint8();
				if (err) {
					console.error("Failed to read from unidirectional stream:", err);
					return;
				}

				switch (num) {
					case UniStreamTypes.GroupStreamType:
						await this.#handleGroupStream(stream);
						break;
					default:
						console.warn(`Unknown unidirectional stream type: ${num}`);
						break; // Ignore unknown stream types
				}
			}
		} catch (error) {
			console.error("Error in listenUniStreams:", error);
		}
	}

	async close(): Promise<void> {
		if (this.#ctx.err()) {
			return;
		}

		this.#conn.close({
			closeCode: 0x0, // Normal closure
			reason: "No Error",
		});

		await Promise.allSettled(this.#wg);
		this.#wg = [];
	}

	async closeWithError(code: number, message: string): Promise<void> {
		if (this.#ctx.err()) {
			return;
		}

		this.#conn.close({
			closeCode: code,
			reason: message,
		});

		await Promise.allSettled(this.#wg);
		this.#wg = [];
	}
}
