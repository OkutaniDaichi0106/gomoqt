import { Reader, Writer } from "./io";
import { AnnounceMessage, AnnouncePleaseMessage } from "./message";
import { CancelCauseFunc, CancelFunc, Context, withCancel, withCancelCause } from "./internal/context";
import { Cond } from "./internal/cond";
import { TrackPrefix, isValidPrefix, validateTrackPrefix } from "./track_prefix";
import { BroadcastPath, validateBroadcastPath } from "./broadcast_path";
import { StreamError } from "./io/error";
import { Queue } from "./internal";

export class AnnouncementWriter {
    #writer: Writer;
    #reader: Reader;
    #prefix: TrackPrefix;
    #announcements: Map<string, Announcement> = new Map();
    #ctx: Context;
    #cancelFunc: CancelCauseFunc;

    constructor(sessCtx: Context, writer: Writer, reader: Reader, req: AnnouncePleaseMessage) {
        this.#writer = writer;
        this.#reader = reader;

        this.#prefix = validateTrackPrefix(req.prefix);

        [this.#ctx, this.#cancelFunc] = withCancelCause(sessCtx);

        // Listen for stream closure
        reader.closed().then(() => {
            this.#cancelFunc(undefined);
        });
    }

    async send(announcement: Announcement): Promise<void> {
        const path = announcement.broadcastPath;
        const active = announcement.isActive();

        if (!path.startsWith(this.#prefix)) {
            throw new Error(`Path ${path} does not start with prefix ${this.#prefix}`);
        }

        const suffix = path.substring(this.#prefix.length);
        const old = this.#announcements.get(suffix);
        console.log(`Sending announcement for path: ${suffix}, active: ${active}`);
        if (active) {
            if (old && old.isActive()) {
                throw new Error(`Announcement for path ${suffix} already exists`);
            } else if (old && !old.isActive()) {
                // Delete the old announcement if it is inactive
                this.#announcements.delete(suffix);
            }

            const [_, err] = await AnnounceMessage.encode(this.#writer, suffix, active);
            if (err) {
                throw new Error(`Failed to write announcement: ${err}`);
            }

            console.log(`Announcement sent for path: ${suffix}`);

            const fork = announcement.fork();
            this.#announcements.set(suffix, fork);

            async () => {
                await fork.ended();
                // When the announcement ends, we remove it from the map
                this.#announcements.delete(suffix);
                const [_, err] = await AnnounceMessage.encode(this.#writer, suffix, false);
                if (err) {
                    throw new Error(`Failed to write end of announcement: ${err}`);
                }
            }
        } else {
            if (!old) {
                throw new Error(`Announcement for path ${suffix} does not exist`);
            } else if (old && !old.isActive()) {
                throw new Error(`Announcement for path ${suffix} is already inactive`);
            }

            // End the old active announcement
            old.end();
            this.#announcements.delete(suffix);
        }
    }

    get context(): Context {
        return this.#ctx;
    }

    close(): void {
        const ctxErr = this.#ctx.err();
        if (ctxErr !== undefined) {
            throw ctxErr;
        }
        this.#writer.close();
        this.#announcements.clear();
        this.#cancelFunc(undefined);
    }

    closeWithError(code: number, message: string): void {
        const ctxErr = this.#ctx.err();
        if (ctxErr !== undefined) {
            throw ctxErr;
        }
        const err = new StreamError(code, message);
        this.#writer.cancel(err);
        for (const announcement of this.#announcements.values()) {
            announcement.end();
        }
        this.#announcements.clear();
    }
}

export class AnnouncementReader {
    #writer: Writer;
    #reader: Reader;
    #prefix: string;
    #announcements: Map<string, Announcement> = new Map();
    #queue: Queue<Announcement> = new Queue();
    #ctx: Context;
    #cancelFunc: CancelCauseFunc;
    #cond: Cond = new Cond();


    constructor(sessCtx: Context, writer: Writer, reader: Reader, announcePlease: AnnouncePleaseMessage) {
        this.#writer = writer;
        this.#reader = reader;
        const prefix = announcePlease.prefix;
        if (!isValidPrefix(prefix)) {
            throw new Error(`Invalid prefix: ${prefix}. It must start and end with '/'.`);
        }
        this.#prefix = prefix;
        const [ctx, cancelFunc] = withCancelCause(sessCtx);
        this.#ctx = ctx;
        this.#cancelFunc = cancelFunc;

        // Listen for incoming announcements
        (async () => {
            for (;;) {
                const [msg, err] = await AnnounceMessage.decode(reader);
                if (err) {
                    throw new Error(`Failed to read announcement: ${err}`);
                }
                if (!msg) {
                    throw new Error("Announcement message is undefined after decoding");
                }

                const old = this.#announcements.get(msg.suffix);

                if (msg.active) {
                    if (old && old.isActive()) {
                        throw new Error(`Announcement for path ${msg.suffix} already exists`);
                    } else if (old && !old.isActive()) {
                        // Delete the old announcement if it is inactive
                        this.#announcements.delete(msg.suffix);
                    }

                    const fullPath = prefix + msg.suffix;
                    const announcement = new Announcement(validateBroadcastPath(fullPath), ctx);
                    this.#announcements.set(msg.suffix, announcement);
                    this.#queue.enqueue(announcement);
                } else {
                    if (!old) {
                        throw new Error(`Announcement for path ${msg.suffix} does not exist`);
                    } else if (old && !old.isActive()) {
                        throw new Error(`Announcement for path ${msg.suffix} is already inactive`);
                    }

                    // End the old active announcement
                    old.end();
                    this.#announcements.delete(msg.suffix);
                }

                this.#cond.broadcast();
            }
        })();
    }

    async receive(): Promise<Announcement> {
        while (true) {
            const err = this.#ctx.err();
            if (err !== undefined) {
                throw err;
            }

            const announcement = await this.#queue.dequeue();

            if (announcement && announcement.isActive()) {
                return announcement;
            }

            // Wait for the next announcement
            await this.#cond.wait();
        }
    }

    get context(): Context {
        return this.#ctx;
    }

    close(): void {
        this.#cancelFunc(new Error("ReceiveAnnounceStream closed"));
        this.#announcements.clear();
        this.#queue.close();
    }

    closeWithError(code: number, message: string): void {
        this.#cancelFunc(new Error(`ReceiveAnnounceStream closed with code ${code}`));
        for (const announcement of this.#announcements.values()) {
            announcement.end();
        }
        this.#writer.cancel(new StreamError(code, message)); // TODO:
        this.#announcements.clear();
        this.#queue.close();
    }
}

export class Announcement {
    readonly broadcastPath: BroadcastPath;
    #ctx: Context;
    #cancelFunc: CancelFunc;

    constructor(path: BroadcastPath, ctx: Context) {
        this.broadcastPath = validateBroadcastPath(path);
        [this.#ctx, this.#cancelFunc] = withCancel(ctx);
    }

    end() {
        this.#cancelFunc();
    }

    fork(): Announcement {
        return new Announcement(this.broadcastPath, this.#ctx);
    }

    isActive(): boolean {
        return this.#ctx.err() === undefined;
    }

    ended(): Promise<void>{
        return this.#ctx.done();
    }
}