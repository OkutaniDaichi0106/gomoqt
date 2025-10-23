import type { CatalogInit } from "../catalog/init";
import { CatalogInitSchema } from "../catalog/init";
import type { TrackDescriptor,CatalogLine } from "../catalog/track";
import { TrackDescriptorSchema,TrackDescriptorsSchema,CatalogLineSchema } from "../catalog/track";
import type { Context,CancelFunc } from "golikejs/context";
import {
    withCancelCause,
    background,
    withCancel,
    ContextCancelledError,
    watchPromise
} from "golikejs/context";
import {
    TrackWriter,
    TrackReader,
    GroupReader,
    GroupWriter,
    InternalSubscribeErrorCode
} from "@okutanidaichi/moqt";
import type {
    BytesFrame
} from "@okutanidaichi/moqt";
import { Channel } from "golikejs/channel";
import { JsonLineDecoder, EncodedJsonChunk,JsonLineEncoder } from "../internal/json";
import type { JsonObject } from "../internal/json";
import type { EncodedChunk, EncodeDestination } from "./container";

export class TrackCatalog {
    readonly descriptor: TrackDescriptor;
    readonly done: Promise<void>;
    #end!: () => void;
    #active: boolean = true;

    constructor(ctx: Promise<void>, descriptor: TrackDescriptor) {
        this.descriptor = descriptor;
        this.done = new Promise<void>((resolve) => {
            this.#end = () => {
                this.#active = false;
                resolve()
            };
        });

        ctx.then(() => {
            this.#end();
        });
    }

    get active(): boolean {
        return this.#active;
    }

    end(): void {
        this.#end();
    }
}

export interface CatalogEncoderInit {
    version: string;
}

export class CatalogEncoder {
    readonly version: string;

    #tracks: Map<string, TrackCatalog> = new Map();

    #channels: Set<Channel<EncodedChunk>> = new Set();

    #encoder: JsonLineEncoder;

    constructor(init: CatalogEncoderInit) {
        this.version = init.version;

        this.#encoder = new JsonLineEncoder();
    }

    async set(tracks: TrackCatalog[]): Promise<Error | undefined> {
        if (tracks.length === 0) {
            return undefined;
        }

        const lines: CatalogLine[] = [];
        const set: Set<string> = new Set();
        for (const track of tracks) {
            if (set.has(track.descriptor.name)) {
                lines.push({active: false, name: track.descriptor.name});
                continue;
            }
            if (!track.active) {
                // Skip ended tracks
                continue;
            }
            set.add(track.descriptor.name);
            this.#tracks.set(track.descriptor.name, track);
            lines.push({active: true, track: track.descriptor});
        }

        const chunk = this.#encoder.encode(lines);

        await Promise.allSettled(
            Array.from(this.#channels, async chan => {
                await chan.send(chunk);
            })
        );
        
        return undefined;
    }

    async encodeTo(dest: EncodeDestination): Promise<Error | undefined> {
        let err: Error | undefined;
        let group: GroupWriter | undefined;

        const initChunk = this.#encoder.encode([{
            version: this.version,
        }]);

        err = await dest.output(initChunk);
        if (err) {
            return new Error("Failed to write init chunk: " + err.message);
        }

        let chunk: EncodedChunk | undefined;

        const existings: CatalogLine[] = Array.from(this.#tracks.values()).map(track => {
            return {active: true, track: track.descriptor};
        });
        if (existings.length > 0) {
            chunk = this.#encoder.encode(existings)
            err = await dest.output(chunk);
            if (err) {
                return new Error("Failed to write existing tracks: " + err.message);
            }
        }

        const chan = new Channel<EncodedChunk>(2);
        this.#channels.add(chan);

        // Integrated encode loop
        try {
            const watchCtx = watchPromise(background(), dest.done);
            let ok: boolean;
            while (true) {
                // Check context cancellation before waiting
                err = watchCtx.err();
                if (err) {
                    return err;
                }

                // Race between context cancellation and channel receive
                const result = await Promise.race([
                    chan.receive(),
                    watchCtx.done().then(() => { return watchCtx.err()!; })
                ]);

                if (result instanceof Error) {
                    return result;
                }

                [chunk, ok] = result;
                if (!ok) {
                    break;
                }

                err = await dest.output(chunk!);
                if (err) {
                    return new Error("Failed to write frame: " + err.message);
                }
            }
        } catch (e) {
            err = e instanceof Error ? e : new Error(String(e));
            return err;
        } finally {
            this.#channels.delete(chan);
            chan.close();
        }
    }
}

export interface CatalogReaderInit {
    version: string;
    reader: TrackReader;
}

export class CatalogDecoder {
    readonly version: string;

    #source: TrackReader;

    #tracks: Map<string, TrackCatalog> = new Map();

    #dests: Set<(tracks: TrackCatalog[]) => void> = new Set();

    #decoder: JsonLineDecoder = new JsonLineDecoder();

    #cancelFunc: CancelFunc;

    constructor(init: CatalogReaderInit) {
        this.version = init.version;
        this.#source = init.reader;

        const [ctx, cancel] = withCancel(this.#source.context);
        this.#cancelFunc = cancel;

        this.#decodeFrom(ctx, this.#source)
    }

    async #decodeFrom(ctx: Context, track: TrackReader): Promise<Error | undefined> {
        while (true) {
            let [group, err] = await track.acceptGroup(ctx.done());

            if (err) {
                return err;
            }

            // Integrated handle logic
            let frame: BytesFrame | undefined;

            try {
                err = ctx.err();
                if (err) {
                    return err;
                }

                let isInit = true;

                while (true) {
                    err = ctx.err();
                    if (err) {
                        return err;
                    }

                    [frame, err] = await group!.readFrame();
                    if (err) {
                        return err;
                    }
                    if (!frame) {
                        // No more frames, exit loop
                        break;
                    }

                    const chunk = new EncodedJsonChunk({
                        type: "jsonl",
                        data: frame.bytes,
                    });

                    let lines: any[];
                    try {
                        lines = this.#decoder.decode(chunk);
                    } catch (e) {
                        err = e instanceof Error ? e : new Error(String(e));
                        break;
                    }

                    if (isInit) {
                        const { success, data: init } = CatalogInitSchema.safeParse(lines.pop());
                        if (!success) {
                            err = new Error("Invalid catalog init data");
                            break;
                        }

                        if (this.version !== init.version) {
                            err = new Error(`Catalog version mismatch: expected ${this.version}, got ${init.version}`);
                            break;
                        }

                        isInit = false;
                    }

                    const catalogs: TrackCatalog[] = [];
                    for (const line of lines) {
                        const { success, data: track } = CatalogLineSchema.safeParse(line);
                        if (!success) {
                            continue;
                        }

                        if (track.active) {
                            const existing = this.#tracks.get(track.track.name);
                            if (existing) {
                                // End the old track since we're replacing it
                                existing.end();
                            }
                            const trackCatalog = new TrackCatalog(ctx.done(), track.track);
                            this.#tracks.set(track.track.name, trackCatalog);
                            catalogs.push(trackCatalog);
                        } else {
                            const existing = this.#tracks.get(track.name);
                            if (existing) {
                                existing.end();
                                this.#tracks.delete(track.name);
                            }
                        }
                    }

                    // Send if there are tracks (skip empty sends after init)
                    if (catalogs.length > 0) {
                        for (const dest of this.#dests) {
                            dest(catalogs);
                        }
                    }
                }

                return undefined;
            } catch (e) {
                err = e instanceof Error ? e : new Error(String(e));
                return err;
            } finally {
                await group!.cancel(InternalSubscribeErrorCode, "CatalogReader is closed");
            }
        }
    }

    async decodeTo(ctx: Promise<void>, dest: (tracks: TrackCatalog[]) => void): Promise<void> {
        this.#dests.add(dest);
        try {
            await ctx;
        } finally {
            this.#dests.delete(dest);
        }
    }

    cancel(): void {
        this.#cancelFunc();
    }
}
