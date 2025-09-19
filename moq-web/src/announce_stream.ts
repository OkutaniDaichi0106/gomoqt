import type { Reader, Writer } from "./io";
import type { AnnouncePleaseMessage } from "./message";
import { AnnounceMessage } from "./message";
import { withCancel, withCancelCause, withPromise, background, ContextCancelledError } from "./internal/context";
import type { CancelCauseFunc, CancelFunc, Context } from "./internal/context";
import { Cond } from "./internal/cond";
import type { TrackPrefix } from "./track_prefix";
import { isValidPrefix, validateTrackPrefix } from "./track_prefix";
import { validateBroadcastPath } from "./broadcast_path";
import type{  BroadcastPath } from "./broadcast_path";
import { StreamError } from "./io/error";
import { Queue } from "./internal";
import { AnnounceInitMessage } from "./message/announce_init";
import type { AnnounceErrorCode } from ".";

type suffix = string;

export class AnnouncementWriter {
    #writer: Writer;
    #reader: Reader;
    readonly prefix: TrackPrefix;
    #announcements: Map<suffix, Announcement> = new Map();
    #ctx: Context;
    #cancelFunc: CancelCauseFunc;
    #ready: Promise<void>;
    #resolveInit!: () => void;

    constructor(sessCtx: Context, writer: Writer, reader: Reader,
         req: AnnouncePleaseMessage) {
        this.#writer = writer;
        this.#reader = reader;

        this.prefix = validateTrackPrefix(req.prefix);

        [this.#ctx, this.#cancelFunc] = withCancelCause(sessCtx);
        this.#ready = new Promise<void>((resolve) => {
            this.#resolveInit = resolve;
        });

        // Listen for stream closure
        reader.closed().then(() => {
            this.#cancelFunc(undefined);
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
                    return new Error(`Announcement for path ${suffix} already exists`);
                } else if (old && !old.isActive()) {
                    // Delete the old announcement if it is inactive
                    this.#announcements.delete(suffix);
                }

                this.#announcements.set(suffix, announcement);

                announcement.ended().then(async ()=>{
                    // When the announcement ends, we remove it from the map
                    this.#announcements.delete(suffix);
                    const msg = new AnnounceMessage({
                        suffix,
                        active: false
                    });
                    const err = await msg.encode(this.#writer);
                    if (err) {
                        return new Error(`Failed to write end of announcement: ${err}`);
                    }
                });
            } else {
                if (!old) {
                    return new Error(`Announcement for path ${suffix} does not exist`);
                } else if (old && !old.isActive()) {
                    return new Error(`Announcement for path ${suffix} is already inactive`);
                }

                // End the old active announcement
                old.end();
                this.#announcements.delete(suffix);
            }
        }

        const msg = new AnnounceInitMessage({
            suffixes: Array.from(this.#announcements.keys())
        });
        const err = await msg.encode(this.#writer);
        if (err) {
            return new Error(`Failed to write init message: ${err}`);
        }

        // Resolve the initialization promise
        this.#resolveInit();

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
                return new Error(`Announcement for path ${suffix} already exists`);
            } else if (old && !old.isActive()) {
                // Delete the old announcement if it is inactive
                this.#announcements.delete(suffix);
            }

            const msg = new AnnounceMessage({
                suffix,
                active
            });
            let err = await msg.encode(this.#writer);
            if (err) {
                return new Error(`Failed to write announcement: ${err}`);
            }

            console.log(`Announcement sent for path: ${suffix}`);

            this.#announcements.set(suffix, announcement);

            announcement.ended().then(async () => {
                this.#announcements.delete(suffix);
                msg.active = false;
                err = await msg.encode(this.#writer);
                if (err) {
                    return new Error(`Failed to write end of announcement: ${err}`);
                }
            });
        } else {
            if (!old) {
                return new Error(`Announcement for path ${suffix} does not exist`);
            } else if (old && !old.isActive()) {
                return new Error(`Announcement for path ${suffix} is already inactive`);
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

    closeWithError(code: AnnounceErrorCode, message: string): void {
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
    prefix: string;
    #announcements: Map<string, Announcement> = new Map();
    #queue: Queue<Announcement> = new Queue();
    #ctx: Context;
    #cancelFunc: CancelCauseFunc;
    #cond: Cond = new Cond();


    constructor(sessCtx: Context, writer: Writer, reader: Reader,
        announcePlease: AnnouncePleaseMessage, aim: AnnounceInitMessage) {
        this.#writer = writer;
        this.#reader = reader;
        const prefix = announcePlease.prefix;
        if (!isValidPrefix(prefix)) {
            throw new Error(`Invalid prefix: ${prefix}. It must start and end with '/'.`);
        }
        this.prefix = prefix;
        [this.#ctx, this.#cancelFunc] = withCancelCause(sessCtx);

        // Set initial announcements
        for (const suffix of aim.suffixes) {
            const path = validateBroadcastPath(prefix + suffix);
            const announcement = new Announcement(path, this.#ctx.done());
            this.#announcements.set(suffix, announcement);
            this.#queue.enqueue(announcement);
        }

        // Listen for incoming announcements
        // Start the reading loop
        queueMicrotask(this.#readNext.bind(this));
    }

    async receive(signal: Promise<void>): Promise<[Announcement | undefined, Error | undefined]> {
        const ctx = withPromise(this.context, signal);

        while (true) {
            const announcement = await this.#queue.dequeue();

            if (announcement && announcement.isActive()) {
                return [announcement, undefined];
            }

            if (ctx.err() !== undefined) {
                return [undefined, ctx.err()];
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
        msg.decode(this.#reader).then((err) => {
            if (err) {
                console.error(`Failed to read announcement: ${err}`);
                return;
            }

            const old = this.#announcements.get(msg.suffix);

            if (msg.active) {
                if (old && old.isActive()) {
                    console.error(`Announcement for path ${msg.suffix} already exists`);
                    return;
                } else if (old && !old.isActive()) {
                    this.#announcements.delete(msg.suffix);
                }

                const fullPath = this.prefix + msg.suffix;
                const announcement = new Announcement(validateBroadcastPath(fullPath), this.#ctx.done());
                this.#announcements.set(msg.suffix, announcement);
                this.#queue.enqueue(announcement);
            } else {
                if (!old) {
                    console.error(`Announcement for path ${msg.suffix} does not exist`);
                    return;
                } else if (old && !old.isActive()) {
                    console.error(`Announcement for path ${msg.suffix} is already inactive`);
                    return;
                }

                old.end();
                this.#announcements.delete(msg.suffix);
            }

            this.#cond.broadcast();

            queueMicrotask(this.#readNext.bind(this));
        });
    }

    get context(): Context {
        return this.#ctx;
    }

    close(): void {
        this.#cancelFunc(new Error("AnnouncementReader closed"));
        this.#announcements.clear();
        this.#queue.close();
    }

    closeWithError(code: AnnounceErrorCode, message: string): void {
        this.#cancelFunc(new Error(`AnnouncementReader closed with code ${code}`));
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

    constructor(path: string, context: Promise<void>) {
        this.broadcastPath = validateBroadcastPath(path);
        const ctx = withPromise(background(), context);
        [this.#ctx, this.#cancelFunc] = withCancel(ctx);
    }

    end(): void {
        this.#cancelFunc();
    }

    isActive(): boolean {
        return this.#ctx.err() === undefined;
    }

    ended(): Promise<void>{
        return this.#ctx.done();
    }
}