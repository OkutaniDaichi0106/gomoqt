import { Version, Versions } from "./internal/version";
import { AnnouncePleaseMessage, GroupMessage, SessionClientMessage, SessionServerMessage, SubscribeMessage, SubscribeOkMessage } from "./message";
import { Writer, Reader } from "./internal/io";
import { Extensions } from "./internal/extensions";
import { SessionStream } from "./session_stream";
import { background, Context, withPromise } from "./internal/context";
import { AnnouncementReader } from "./announce_stream";
import { TrackPrefix } from "./track_prefix";
import { ReceiveSubscribeStream, SendSubscribeStream, SubscribeConfig, SubscribeID } from "./subscribe_stream";
import { Subscriber as Subscription } from "./subscriber";
import { BroadcastPath } from "./broadcast_path";
import { TrackReader, TrackWriter } from "./track";
import { GroupReader } from "./group_stream";
import { Queue } from "./internal/queue";
import { TrackMux } from "./track_mux";

export class Session {
	readonly ready: Promise<void>
	#conn: WebTransport
	#sessionStream!: SessionStream
	#ctx!: Context;

	#idCounter: bigint = 0n;

	#mux: TrackMux = new TrackMux();

	#subscriptions: Map<SubscribeID, [Context, Queue<GroupReader>]> = new Map();

	constructor(conn: WebTransport, versions: Set<Version> = new Set([Versions.DEVELOP]), extensions: Extensions = new Extensions()) {
		this.#conn = conn;
		this.ready = conn.ready.then(async () => {
			const stream = await conn.createBidirectionalStream();
			const baseCtx = withPromise(background(), conn.closed); // TODO: Handle connection closure properly
			const writer = new Writer(stream.writable);
			const reader = new Reader(stream.readable);

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

		}).catch((error) => {
			console.error("Error during session initialization:", error);
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
		const [req, err] = await AnnouncePleaseMessage.encode(writer, prefix)
		if (err) {
			throw err;
		}
		if (!req) {
			throw new Error("Failed to encode AnnouncePleaseMessage");
		}

		return new AnnouncementReader(this.#ctx, writer, reader, req);
	}

	async openTrackStream(path: BroadcastPath, name: string, config: SubscribeConfig): Promise<Subscription> {
		const stream = await this.#conn.createBidirectionalStream()
		const writer = new Writer(stream.writable);
		const reader = new Reader(stream.readable);
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

		const controller = new SendSubscribeStream(this.#ctx, writer, reader, req, rsp);
		const trackCtx = controller.context;

		const queue = new Queue<GroupReader>();
		this.#subscriptions.set(req.subscribeId, [trackCtx, queue]);

		const acceptFunc = (): Promise<[GroupReader?, Error?]> => {
			return new Promise((resolve) => {
				const [reader, err] = queue.dequeue();
				if (err) {
					resolve([undefined, err]);
					return;
				}
				if (!reader) {
					resolve([undefined, new Error("No group message available")]);
					return;
				}
				resolve([reader, undefined]);
			});
		}

		const track = new TrackReader(trackCtx, acceptFunc)

		return {
			broadcastPath: path,
			trackName: name,
			subscribeId: req.subscribeId,
			controller,
			trackReader: track,
		};
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

		const subscription = this.#subscriptions.get(req.subscribeId);
		if (!subscription) {
			console.error(`No subscription found for SubscribeID: ${req.subscribeId}`);
			return;
		}

		const [trackCtx, queue] = subscription;
		if (trackCtx.err()) {
			console.error(`Track context for SubscribeID ${req.subscribeId} has an error: ${trackCtx.err()}`);
			return;
		}

		const groupReader = new GroupReader(trackCtx, reader, req);
		queue.enqueue(groupReader);
	}

	async #handleSubscribeStream(writer: Writer, reader: Reader): Promise<void> {
		const [req, err] = await SubscribeMessage.decode(reader);
		if (err) {
			console.error("Failed to decode SubscribeMessage:", err);
			return;
		}
		if (!req) {
			console.error("Received empty SubscribeMessage");
			return;
		}

		

		const subscribeStream = new SendSubscribeStream(this.#ctx, writer, reader, req);
		const subscriber = new Subscriber(subscribeStream);
		this.#sessionStream.addSubscriber(subscriber);
		subscribeStream.start();
		this.#handleAnnounceStream(writer, reader);
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

		const stream = new AnnouncementReader(this.#ctx, writer, reader, req);

		this.#mux.serveAnnouncement(stream, req.prefix);
	}

	close(): void {
		this.#conn.close();
	}
}

export function dial(url: string | URL, versions?: Set<Version>, extensions?: Extensions): Promise<Session> {
	return new Promise((resolve, reject) => {
		const conn = new WebTransport(url);
		conn.ready.then(() => {
			resolve(new Session(conn, versions, extensions));
		}).catch(reject);
	});
}