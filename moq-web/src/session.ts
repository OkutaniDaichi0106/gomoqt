import { Versions,DEFAULT_CLIENT_VERSIONS } from "./internal";
import type { Version } from "./internal";
import { AnnouncePleaseMessage, AnnounceInitMessage, GroupMessage, SessionClientMessage, SessionServerMessage, SubscribeMessage, SubscribeOkMessage } from "./message";
import { Writer, Reader } from "./io";
import { Extensions } from "./internal/extensions";
import { SessionStream } from "./session_stream";
import { background, withPromise } from "./internal/context";
import type { Context } from "./internal/context";
import { AnnouncementReader, AnnouncementWriter } from "./announce_stream";
import type { TrackPrefix } from "./track_prefix";
import { ReceiveSubscribeStream, SendSubscribeStream } from "./subscribe_stream";
import type { TrackConfig } from "./subscribe_stream";
import type { BroadcastPath } from "./broadcast_path";
import { TrackReader, TrackWriter } from "./track";
import { GroupReader, GroupWriter } from "./group_stream";
import type { TrackMux } from "./track_mux";
import { DefaultTrackMux } from "./track_mux";
import { BiStreamTypes, UniStreamTypes } from "./stream_type";
import { Queue } from "./internal/queue";
import type { Info } from "./info";
import type { TrackName, SubscribeID } from "./protocol";

export interface SessionInit {
	conn: WebTransport;
	versions?: Set<Version>;
	extensions?: Extensions;
	mux?: TrackMux;
}

export class Session {
	readonly ready: Promise<void>
	#conn: WebTransport
	#sessionStream!: SessionStream
	#ctx!: Context;

	#wg: Promise<void>[] = [];

	#subscribeIDCounter: bigint = 0n;

	readonly mux: TrackMux;

	#enqueueFuncs: Map<SubscribeID, (stream: Reader, msg: GroupMessage) => void> = new Map();

	constructor(init: SessionInit) {
		this.#conn = init.conn;
		this.mux = init.mux ?? DefaultTrackMux;
		this.ready = this.#setup(
			init.versions ?? DEFAULT_CLIENT_VERSIONS,
			init.extensions ?? new Extensions()
		);
	}

	async #setup(versions: Set<Version>, extensions: Extensions): Promise<void> {
		await this.#conn.ready;

		const stream = await this.#conn.createBidirectionalStream();
		const writer = new Writer(stream.writable);
		const reader = new Reader(stream.readable);

		// Send STREAM_TYPE
		writer.writeUint8(BiStreamTypes.SessionStreamType);
		let err = await writer.flush();
		if (err) {
			console.error("moq: failed to open session stream:", err);
			throw err;
		}

		// Send the session client message
		const req = new SessionClientMessage({
			versions,
			extensions,
		});
		err = await req.encode(writer);
		if (err) {
			console.error("moq: failed to send SESSION_CLIENT message:", err);
			throw err;
		}

		console.debug("moq: SESSION_CLIENT message sent.",
			{
				"message": req,
			}
		);

		// Receive the session server message
		const rsp = new SessionServerMessage({});
		err = await rsp.decode(reader);
		if (err) {
			console.error("moq: failed to receive SESSION_SERVER message:", err);
			throw err;
		}

		console.debug("moq: SESSION_SERVER message received.",
			{
				"message": rsp,
			}
		);

		// TODO: Check the version compatibility
		if (!versions.has(rsp.version)) {
			throw new Error(`Incompatible session version: ${rsp.version}`);
		}

		const connCtx = withPromise(background(), this.#conn.closed); // TODO: Handle connection closure properly

		this.#sessionStream = new SessionStream(connCtx, writer, reader, req, rsp);

		this.#ctx = this.#sessionStream.context;

		// Start listening for incoming streams
		this.#wg.push(this.#listenBiStreams());
		this.#wg.push(this.#listenUniStreams());

		return;
	}

	update(bitrate: bigint) {
		this.#sessionStream.update(bitrate);
	}

	async acceptAnnounce(prefix: TrackPrefix): Promise<[AnnouncementReader, undefined] | [undefined, Error]> {
		const stream = await this.#conn.createBidirectionalStream()
		const writer = new Writer(stream.writable);
		const reader = new Reader(stream.readable);

		// Send STREAM_TYPE
		writer.writeUint8(BiStreamTypes.AnnounceStreamType);
		let err = await writer.flush();
		if (err) {
			console.error("moq: failed to open announce stream:", err);
			return [undefined, err];
		}

		// Send AnnouncePleaseMessage
		const req = new AnnouncePleaseMessage({ prefix });
		err = await req.encode(writer);
		if (err) {
			console.error("moq: failed to send ANNOUNCE_PLEASE message:", err);
			return [undefined, err];
		}

		console.debug(`moq: ANNOUNCE_PLEASE message sent.`,
			{
				"message": req,
			}
		)

		// Receive AnnounceInitMessage
		const rsp = new AnnounceInitMessage({});
		err = await rsp.decode(reader);
		if (err) {
			console.error("moq: failed to receive ANNOUNCE_INIT message:", err);
			return [undefined, err];
		}

		console.debug(`moq: ANNOUNCE_INIT message received.`,
			{
				"prefix": prefix,
				"message": rsp,
			}
		)

		return [new AnnouncementReader(this.#ctx, writer, reader, req, rsp), undefined];
	}

	async subscribe(path: BroadcastPath, name: TrackName, config?: TrackConfig): Promise<[TrackReader, undefined] | [undefined, Error]> {
		const stream = await this.#conn.createBidirectionalStream()
		const writer = new Writer(stream.writable);
		const reader = new Reader(stream.readable);

		// Send STREAM_TYPE
		writer.writeUint8(BiStreamTypes.SubscribeStreamType);
		let err = await writer.flush();
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
			maxGroupSequence: config?.maxGroupSequence ?? 0n
		});
		err = await req.encode(writer);
		if (err) {
			console.error("moq: failed to send SUBSCRIBE message:", err);
			return [undefined, err];
		}

		console.debug(`moq: SUBSCRIBE message sent.`,
			{
				"message": req,
			}
		)

		const rsp = new SubscribeOkMessage({});
		err = await rsp.decode(reader);
		if (err) {
			console.error("moq: failed to receive SUBSCRIBE_OK message:", err);
			return [undefined, err];
		}

		console.debug(`moq: SUBSCRIBE_OK message received.`,
			{
				"subscribeId": req.subscribeId,
				"message": req,
			}
		)

		const subscribeStream = new SendSubscribeStream(this.#ctx, writer, reader, req, rsp);

		const queue = new Queue<[Reader, GroupMessage]>();

		// Add the enqueue function to the map
		this.#enqueueFuncs.set(req.subscribeId, (stream, msg) => {
			queue.enqueue([stream, msg]);
		});

		const track = new TrackReader(
			subscribeStream,
			queue.dequeue,
			() => {this.#enqueueFuncs.delete(req.subscribeId);}
		);

		return [track, undefined];
	}

	async #handleGroupStream(reader: Reader): Promise<void> {
		const req = new GroupMessage({});
		const err = await req.decode(reader);
		if (err) {
			console.error("Failed to decode GroupMessage:", err);
			return;
		}

		console.debug("moq: GROUP message received.",
			{
				"message": req,
			}
		)

		const enqueueFunc = this.#enqueueFuncs.get(req.subscribeId);
		if (!enqueueFunc) {
			console.error(`moq: no subscription found for Subscribe ID: ${req.subscribeId}`);
			return;
		}

		enqueueFunc(reader, req);
	}

	async #handleSubscribeStream(writer: Writer, reader: Reader): Promise<void> {
		const req = new SubscribeMessage({});
		const reqErr = await req.decode(reader);
		if (reqErr) {
			console.error("Failed to decode SubscribeMessage:", reqErr);
			return;
		}

		console.debug("moq: SUBSCRIBE message received.",
			{
				"message": req,
			}
		);

		const subscribeStream = new ReceiveSubscribeStream(this.#ctx, writer, reader, req);

		const openUniStreamFunc = async (): Promise<[Writer, undefined] | [undefined, Error]> => {
			try {
				const writer = new Writer(await this.#conn.createUnidirectionalStream());
				return [writer, undefined];
			} catch (err) {
				console.error("moq: failed to create unidirectional stream:", err);
				return [undefined, new Error(`moq: failed to create unidirectional stream: ${err}`)];
			}
		};

		const trackWriter = new TrackWriter(
			req.broadcastPath as BroadcastPath,
			req.trackName,
			subscribeStream,
			openUniStreamFunc,
		);

		this.mux.serveTrack(trackWriter);
	}

	async #handleAnnounceStream(writer: Writer, reader: Reader): Promise<void> {
		const req = new AnnouncePleaseMessage({});
		const err = await req.decode(reader);
		if (err) {
			console.error("Failed to decode AnnouncePleaseMessage:", err);
			return;
		}

		console.debug("moq: ANNOUNCE message received.",
			{
				"message": req,
			}
		);

		const stream = new AnnouncementWriter(this.#ctx, writer, reader, req);

		this.mux.serveAnnouncement(stream, stream.prefix);
	}

	async #listenBiStreams(): Promise<void> {
		const biStreams = this.#conn.incomingBidirectionalStreams.getReader();

		try {
			// Handle incoming streams
			let num: number | undefined;
			let err: Error | undefined;
			while (true) {
				const {done, value} = await biStreams.read();
				// biStreams.releaseLock(); // Release the lock after reading
				if (done) {
					console.error("Bidirectional stream closed");
					break;
				}
				const stream = value as WebTransportBidirectionalStream;
				const writer = new Writer(stream.writable);
				const reader = new Reader(stream.readable);
				[num, err] = await reader.readUint8();
				if (err) {
					console.error("Failed to read from bidirectional stream:", err);
					continue;
				}

				switch (num) {
					case BiStreamTypes.SubscribeStreamType:
						this.#handleSubscribeStream(writer, reader);
						break;
					case BiStreamTypes.AnnounceStreamType:
						this.#handleAnnounceStream(writer, reader);
						break;
					default:
						console.warn(`Unknown bidirectional stream type: ${num}`);
						break; // Ignore unknown stream types
				}
			}
		} catch (error) {
			console.error("Error in listenBiStreams:", error);
		} finally {
			biStreams.releaseLock();
		}
	}

	async #listenUniStreams(): Promise<void> {
		const uniStreams = this.#conn.incomingUnidirectionalStreams.getReader();
		try {
			let num: number | undefined;
			let err: Error | undefined;
			while (true) {
				const {done, value} = await uniStreams.read();
				if (done) {
					console.log("Unidirectional stream reader closed");
					break;
				}

				const reader = new Reader(value as ReadableStream<Uint8Array<ArrayBufferLike>>);


				// Read the first byte to determine the stream type
				[num, err] = await reader.readUint8();
				if (err) {
					console.error("Failed to read from unidirectional stream:", err);
					return;
				}

				switch (num) {
					case UniStreamTypes.GroupStreamType:
						await this.#handleGroupStream(reader);
						break;
					default:
						console.warn(`Unknown unidirectional stream type: ${num}`);
						break; // Ignore unknown stream types
				}
			}
		} catch (error) {
			console.error("Error in listenUniStreams:", error);
		} finally {
			uniStreams.releaseLock();
		}
	}

	async close(): Promise<void> {
		if (this.#ctx.err()) {
			return;
		}

		this.#conn.close({
			closeCode: 0x0, // Normal closure
			reason: "No Error"
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