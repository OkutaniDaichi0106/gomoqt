import { EOFError } from "@okudai/golikejs/io";
import type { AnnouncePleaseMessage } from "./internal/message/mod.ts";
import { AnnounceMessage } from "./internal/message/mod.ts";
import { ContextCancelledError, watchPromise, withCancelCause } from "@okudai/golikejs/context";
import type { CancelCauseFunc, Context } from "@okudai/golikejs/context";
import { Cond, Mutex } from "@okudai/golikejs/sync";
import type { TrackPrefix } from "./track_prefix.ts";
import { isValidPrefix, validateTrackPrefix } from "./track_prefix.ts";
import { validateBroadcastPath } from "./broadcast_path.ts";
import type { BroadcastPath } from "./broadcast_path.ts";
import { StreamError } from "./internal/webtransport/error.ts";
import { Queue } from "./internal/queue.ts";
import { AnnounceInitMessage } from "./internal/message/announce_init.ts";
import type { AnnounceErrorCode } from "./error.ts";
import { DuplicatedAnnounceErrorCode } from "./error.ts";
import { Stream } from "./internal/webtransport/stream.ts";

type suffix = string;

export class AnnouncementWriter {
	#stream: Stream;
	readonly prefix: TrackPrefix;
	#announcements: Map<suffix, Announcement> = new Map();
	readonly context: Context;
	#cancelFunc: CancelCauseFunc;
	#ready: Promise<void>;
	#resolveInit?: () => void;

	constructor(
		sessCtx: Context,
		stream: Stream,
		req: AnnouncePleaseMessage,
	) {
		this.#stream = stream;
		this.prefix = validateTrackPrefix(req.prefix);

		// const ctx = watchPromise(sessCtx, reader.closed());
		[this.context, this.#cancelFunc] = withCancelCause(sessCtx);
		this.#ready = new Promise<void>((resolve) => {
			this.#resolveInit = resolve;
		});
	}

	async init(anns: Announcement[]): Promise<Error | undefined> {
		// const onEndFuncs:Map<suffix, () => void> = new Map();
		for (const announcement of anns) {
			const path = announcement.broadcastPath;
			const active = announcement.isActive();

			if (!path.startsWith(this.prefix)) {
				return new Error(`Path ${path} does not start with prefix ${this.prefix}`);
			}

			const suffix = path.substring(this.prefix.length);
			const old = this.#announcements.get(suffix);
			if (active) {
				if (old && old.isActive()) {
					return new Error(
						`[AnnouncementWriter] announcement for path ${this.prefix}${suffix} already exists`,
					);
				} else if (old && !old.isActive()) {
					// Delete the old announcement if it is inactive
					this.#announcements.delete(suffix);
				}

				this.#announcements.set(suffix, announcement);

				announcement.ended().then(async () => {
					// When the announcement ends, we remove it from the map
					this.#announcements.delete(suffix);
					const msg = new AnnounceMessage({ suffix, active: false });
					const err = await msg.encode(this.#stream.writable);
					if (err) {
						return err;
					}

					return undefined;
				});
			} else {
				if (!old || (old && !old.isActive())) {
					return new Error(
						`[AnnouncementWriter] announcement to end for path ${this.prefix}${suffix} is not active.`,
					);
				}

				// End the old active announcement
				old.end();
				this.#announcements.delete(suffix);
			}
		}

		const msg = new AnnounceInitMessage({
			suffixes: Array.from(this.#announcements.keys()),
		});
		const err = await msg.encode(this.#stream.writable);
		if (err) {
			return err;
		}

		console.debug(`moq: ANNOUNCE_INIT message sent.`, {
			"prefix": this.prefix,
			"message": msg,
		});

		// Resolve the initialization promise
		this.#resolveInit?.();
		this.#resolveInit = undefined;

		return undefined;
	}

	async send(announcement: Announcement): Promise<Error | undefined> {
		await this.#ready; // Wait for initialization to complete

		const path = announcement.broadcastPath;
		const active = announcement.isActive();

		if (!path.startsWith(this.prefix)) {
			return new Error(`Path ${path} does not start with prefix ${this.prefix}`);
		}

		const suffix = path.substring(this.prefix.length);
		const old = this.#announcements.get(suffix);
		if (active) {
			if (old && old.isActive()) {
				return new Error(
					`[AnnouncementWriter] announcement for path ${suffix} already exists`,
				);
			} else if (old && !old.isActive()) {
				// Delete the old announcement if it is inactive
				this.#announcements.delete(suffix);
			}

			const msg = new AnnounceMessage({ suffix, active });
			let err = await msg.encode(this.#stream.writable);
			if (err) {
				return err;
			}

			console.debug(`moq: ANNOUNCE message sent.`, {
				"prefix": this.prefix,
				"message": msg,
			});

			this.#announcements.set(suffix, announcement);

			announcement.ended().then(async () => {
				this.#announcements.delete(suffix);
				msg.active = false;
				err = await msg.encode(this.#stream.writable);
				if (err) {
					return err;
				}

				return undefined;
			});
		} else {
			if (!old || (old && !old.isActive())) {
				return new Error(
					`[AnnouncementWriter] announcement to end for path ${this.prefix}${suffix} is not active`,
				);
			}

			// End the old active announcement
			old.end();
			this.#announcements.delete(suffix);
		}

		return undefined;
	}

	async close(): Promise<void> {
		if (this.context.err()) {
			// If already closed, do nothing
			return;
		}
		this.#cancelFunc(undefined);
		await this.#stream.writable.close();
		this.#announcements.clear();
		this.#resolveInit?.();
		this.#resolveInit = undefined;
	}

	async closeWithError(code: AnnounceErrorCode, message: string): Promise<void> {
		if (this.context.err()) {
			// If already closed, do nothing
			return;
		}

		const cause = new StreamError(code, message);
		this.#cancelFunc(cause);
		await this.#stream.writable.cancel(cause);
		await this.#stream.readable.cancel(cause);
		this.#announcements.clear();
		this.#resolveInit?.();
		this.#resolveInit = undefined;
	}
}

export class AnnouncementReader {
	#stream: Stream;
	readonly prefix: string;
	#announcements: Map<string, Announcement> = new Map();
	#queue: Queue<Announcement> = new Queue();
	readonly context: Context;
	#cancelFunc: CancelCauseFunc;
	#mu: Mutex = new Mutex();
	#cond: Cond = new Cond(this.#mu);

	constructor(
		sessCtx: Context,
		stream: Stream,
		announcePlease: AnnouncePleaseMessage,
		aim: AnnounceInitMessage,
	) {
		this.#stream = stream;
		const prefix = announcePlease.prefix;
		if (!isValidPrefix(prefix)) {
			throw new Error(`[AnnouncementReader] invalid prefix: ${prefix}.`);
		}
		this.prefix = prefix;
		[this.context, this.#cancelFunc] = withCancelCause(sessCtx);

		// Set initial announcements
		for (const suffix of aim.suffixes) {
			const path = validateBroadcastPath(prefix + suffix);
			const announcement = new Announcement(path, this.context.done());
			this.#announcements.set(suffix, announcement);
			this.#queue.enqueue(announcement);
		}

		// Listen for incoming announcements
		// Start the reading loop
		queueMicrotask(() => this.#readNext());
	}

	async receive(signal: Promise<void>): Promise<[Announcement, undefined] | [undefined, Error]> {
		const ctx = watchPromise(this.context, signal);

		while (true) {
			const announcement = await this.#queue.dequeue();
			if (announcement === undefined) {
				return [undefined, new Error("Queue is closed and empty")];
			}

			if (announcement && announcement.isActive()) {
				return [announcement, undefined];
			}

			const err = ctx.err();
			if (err) {
				return [undefined, err];
			}

			// Wait for either context cancellation or a condition signal.
			// Using Promise.race here is safe because `cond.wait()` is implemented such that
			// it is a lightweight synchronization primitive and does not capture heavy resources.
			// Even if `cond.wait()` loses the race, it does not keep large memory references alive.
			const result = await Promise.race([
				ctx.done().then(() => ctx.err() ?? ContextCancelledError),
				this.#cond.wait(),
			]);

			if (result instanceof Error) {
				return [undefined, result];
			}
		}
	}

	#readNext(): void {
		const msg = new AnnounceMessage({});
		msg.decode(this.#stream.readable).then(async (err) => {
			if (err) {
				// EOFError and ConnectionReset are expected during normal shutdown
				if (err instanceof EOFError) {
					return;
				}
				if (err.message?.includes("ConnectionReset")) {
					console.debug("moq: ANNOUNCE stream closed by peer (normal shutdown)");
					return;
				}
				console.error(`moq: failed to read ANNOUNCE message: ${err}`);
				return;
			}

			console.debug(`moq: ANNOUNCE message received.`, {
				"prefix": this.prefix,
				"message": msg,
			});

			const old = this.#announcements.get(msg.suffix);

			if (msg.active) {
				if (old && old.isActive()) {
					await this.closeWithError(
						DuplicatedAnnounceErrorCode,
						`duplicate announcement for path ${msg.suffix}`,
					);

					return;
				} else if (old && !old.isActive()) {
					this.#announcements.delete(msg.suffix);
				}

				const fullPath = this.prefix + msg.suffix;
				const announcement = new Announcement(
					validateBroadcastPath(fullPath),
					this.context.done(),
				);
				this.#announcements.set(msg.suffix, announcement);
				this.#queue.enqueue(announcement);
			} else {
				if (!old || (old && !old.isActive())) {
					await this.closeWithError(
						DuplicatedAnnounceErrorCode,
						`trying to end non-existent announcement for path ${msg.suffix}`,
					);

					return;
				}

				old.end();
				this.#announcements.delete(msg.suffix);
			}

			this.#cond.broadcast();

			queueMicrotask(() => this.#readNext());
		});
	}

	async close(): Promise<void> {
		if (this.context.err()) {
			// If already closed, do nothing
			return;
		}

		this.#cancelFunc(undefined);

		await this.#stream.writable.close();
		this.#announcements.clear();
		this.#queue.close();
	}

	async closeWithError(code: AnnounceErrorCode, message: string): Promise<void> {
		if (this.context.err()) {
			// If already closed, do nothing
			return;
		}
		const cause = new StreamError(code, message);
		await this.#stream.writable.cancel(cause);
		await this.#stream.readable.cancel(cause);
		this.#announcements.clear();
		this.#queue.close();
	}
}

export class Announcement {
	readonly broadcastPath: BroadcastPath;
	#done: Promise<void>;
	#signalFunc: () => void;
	#active: boolean = true;

	constructor(path: string, signal: Promise<void>) {
		this.broadcastPath = validateBroadcastPath(path);

		let resolveFunc: () => void;
		this.#done = new Promise<void>((resolve) => {
			resolveFunc = resolve;
		});

		this.#signalFunc = () => resolveFunc();

		// Cancel when the signal is done
		signal.then(() => {
			this.end();
		});
	}

	end(): void {
		if (!this.#active) {
			return;
		}
		this.#active = false;
		this.#signalFunc();
	}

	isActive(): boolean {
		return this.#active;
	}

	ended(): Promise<void> {
		return this.#done;
	}

	afterFunc(fn: () => void): () => boolean {
		let executed = false;
		this.#done.then(() => {
			if (executed) return;
			executed = true;
			fn();
		});

		return () => {
			if (executed) {
				return false;
			}
			executed = true;
			return !executed;
		};
	}
}
