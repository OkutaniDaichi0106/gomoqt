import { Version, Versions } from "./internal";
import { AnnouncePleaseMessage, AnnounceInitMessage, GroupMessage, SessionClientMessage, SessionServerMessage, SubscribeMessage, SubscribeOkMessage } from "./message";
import { Writer, Reader } from "./io";
import { Extensions } from "./internal/extensions";
import { SessionStream } from "./session_stream";
import { background, Context, withPromise } from "./internal/context";
import { AnnouncementReader, AnnouncementWriter } from "./announce_stream";
import { TrackPrefix } from "./track_prefix";
import { ReceiveSubscribeStream, SendSubscribeStream, TrackConfig, SubscribeID } from "./subscribe_stream";
import { BroadcastPath } from "./broadcast_path";
import { TrackReader, TrackWriter } from "./track";
import { GroupReader, GroupWriter } from "./group_stream";
import { DefaultTrackMux, TrackMux } from "./track_mux";
import { BiStreamTypes, UniStreamTypes } from "./stream_type";
import { Queue } from "./internal/queue";
import { Info } from "./info";
import { subscribe } from "diagnostics_channel";

export class Session {
	readonly ready: Promise<void>
	#conn: WebTransport
	#sessionStream!: SessionStream
	#ctx!: Context;

	#idCounter: bigint = 0n;

	#mux: TrackMux;

	#enqueueFuncs: Map<SubscribeID, (stream: Reader, msg: GroupMessage) => void> = new Map();

	constructor(conn: WebTransport,
		versions: Set<Version> = new Set([Versions.DEVELOP]), extensions: Extensions = new Extensions(),
		mux: TrackMux = DefaultTrackMux) {
		this.#conn = conn;
		this.#mux = mux;
		this.ready = conn.ready.then(async () => {
			const stream = await conn.createBidirectionalStream();
			const baseCtx = withPromise(background(), conn.closed); // TODO: Handle connection closure properly
			const writer = new Writer(stream.writable);
			const reader = new Reader(stream.readable);

			// Send STREAM_TYPE
			writer.writeUint8(BiStreamTypes.SessionStreamType);
			const err = await writer.flush();
			if (err) {
				console.error("Failed to flush writer:", err);
				throw err;
			}

			// Send the session client message
			const [req, reqErr] = await SessionClientMessage.encode(writer, versions, extensions);
			if (reqErr) {
				throw reqErr;
			}
			if (!req) {
				throw new Error("Failed to encode SessionClientMessage");
			}

			// Receive the session server message
			const [rsp, rspErr] = await SessionServerMessage.decode(reader);
			if (rspErr) {
				throw rspErr;
			}
			if (!rsp) {
				throw new Error("Failed to decode SessionServerMessage");
			}

			// TODO: Check the version compatibility
			if (!versions.has(rsp.version)) {
				throw new Error(`Incompatible session version: ${rsp.version}`);
			}

			this.#sessionStream = new SessionStream(baseCtx, writer, reader, req, rsp);
			this.#ctx = this.#sessionStream.context;

			return;
		}).then(() => {
			this.#listenBiStreams();
			this.#listenUniStreams();
		}).catch((error) => {
			this.#conn.close(); // TODO: Specify a proper close code and reason
			throw error;
		});
	}

	update(bitrate: bigint) {
		this.#sessionStream.update(bitrate);
	}

	async openAnnounceStream(prefix: TrackPrefix): Promise<AnnouncementReader> {
		const stream = await this.#conn.createBidirectionalStream()
		const writer = new Writer(stream.writable);
		const reader = new Reader(stream.readable);

		// Send STREAM_TYPE
		writer.writeUint8(BiStreamTypes.AnnounceStreamType);
		const err = await writer.flush();
		if (err) {
			throw err;
		}

		// Send AnnouncePleaseMessage
		const [req, reqErr] = await AnnouncePleaseMessage.encode(writer, prefix)
		if (reqErr) {
			throw reqErr;
		}
		if (!req) {
			throw new Error("Failed to encode AnnouncePleaseMessage");
		}

		// Receive AnnounceInitMessage
		const [rsp, rspErr] = await AnnounceInitMessage.decode(reader);
		if (rspErr) {
			throw rspErr;
		}
		if (!rsp) {
			throw new Error("Failed to decode AnnounceInitMessage");
		}


		return new AnnouncementReader(this.#ctx, writer, reader, req, rsp);
	}

	async openTrackStream(path: BroadcastPath, name: string, config: TrackConfig = {
            trackPriority: 0n,
            minGroupSequence: 0n,
            maxGroupSequence: 0n,
        }): Promise<TrackReader> {
		const stream = await this.#conn.createBidirectionalStream()
		const writer = new Writer(stream.writable);
		const reader = new Reader(stream.readable);

		// Send STREAM_TYPE
		writer.writeUint8(BiStreamTypes.SubscribeStreamType);
		const err = await writer.flush();
		if (err) {
			throw err;
		}

		// Send SUBSCRIBE message
		const [req, reqErr] = await SubscribeMessage.encode(writer, this.#idCounter++, path, name,
			config.trackPriority, config.minGroupSequence, config.maxGroupSequence);
		if (reqErr) {
			throw reqErr;
		}
		if (!req) {
			throw new Error("Failed to encode TrackSubscribeMessage");
		}

		const [rsp, rspErr] = await SubscribeOkMessage.decode(reader);
		if (rspErr) {
			throw rspErr;
		}
		if (!rsp) {
			throw new Error("Failed to decode SubscribeOkMessage");
		}

		const subscribeStream = new SendSubscribeStream(this.#ctx, writer, reader, req, rsp);

		const queue = new Queue<GroupReader>();

		// Add the enqueue function to the map
		this.#enqueueFuncs.set(req.subscribeId, (stream, msg) => {
			queue.enqueue(new GroupReader(subscribeStream.context, stream, msg));
		});

		const track = new TrackReader(subscribeStream, queue,
			() => {this.#enqueueFuncs.delete(req.subscribeId);});

		return track;
	}

	async #handleGroupStream(reader: Reader): Promise<void> {
		const [req, err] = await GroupMessage.decode(reader);
		if (err) {
			console.error("Failed to decode GroupMessage:", err);
			return;
		}
		if (!req) {
			console.error("Received empty GroupMessage");
			return;
		}

		const enqueueFunc = this.#enqueueFuncs.get(req.subscribeId);
		if (!enqueueFunc) {
			console.error(`No subscription found for SubscribeID: ${req.subscribeId}`);
			return;
		}

		enqueueFunc(reader, req);
	}

	async #handleSubscribeStream(writer: Writer, reader: Reader): Promise<void> {
		const [req, reqErr] = await SubscribeMessage.decode(reader);
		if (reqErr) {
			console.error("Failed to decode SubscribeMessage:", reqErr);
			return;
		}
		if (!req) {
			console.error("Received empty SubscribeMessage");
			return;
		}

		const id = req.subscribeId;

		const subscribeStream = new ReceiveSubscribeStream(this.#ctx, writer, reader, req);

		const openUniStreamFunc = async (): Promise<[Writer?, Error?]> => {
			try {
				const writer = new Writer(await this.#conn.createUnidirectionalStream());
				return [writer, undefined];
			} catch (err) {
				console.error("Failed to create unidirectional stream:", err);
				return [undefined, new Error(`Failed to create unidirectional stream: ${err}`)];
			}
        };


		const trackWriter = new TrackWriter(req.broadcastPath, req.trackName, subscribeStream, openUniStreamFunc);

		this.#mux.serveTrack(trackWriter);
	}

	async #handleAnnounceStream(writer: Writer, reader: Reader): Promise<void> {
		const [req, err] = await AnnouncePleaseMessage.decode(reader);
		if (err) {
			console.error("Failed to decode AnnouncePleaseMessage:", err);
			return;
		}
		if (!req) {
			console.error("Received empty AnnouncePleaseMessage");
			return;
		}

		console.log(`Received AnnouncePleaseMessage for prefix: ${req.prefix}`);

		const stream = new AnnouncementWriter(this.#ctx, writer, reader, req);

		this.#mux.serveAnnouncement(stream, req.prefix);
	}

	async #listenBiStreams(): Promise<void> {
		const biStreams = this.#conn.incomingBidirectionalStreams.getReader();

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
	}

	async #listenUniStreams(): Promise<void> {
		const uniStreams = this.#conn.incomingUnidirectionalStreams.getReader();
		while (true) {
			const {done, value} = await uniStreams.read();
			// uniStreams.releaseLock(); // Release the lock after reading
			if (done) {
				console.error("Unidirectional stream closed");
				break;
			}
			const readable = value as ReadableStream<Uint8Array<ArrayBufferLike>>;
			const reader = new Reader(readable);

			// Read the first byte to determine the stream type
			const [num, err] = await reader.readUint8();
			if (err) {
				console.error("Failed to read from unidirectional stream:", err);
				continue;
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
	}

	close(): void {
		this.#conn.close({
			closeCode: 0x0, // Normal closure
			reason: "No Error"
		});
	}

	closeWithError(code: number, message: string): void {
		this.#conn.close({
			closeCode: code,
			reason: message,
		});
	}


}