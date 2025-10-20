import type { CatalogInit } from "../catalog/init";
import { CatalogInitSchema } from "../catalog/init";
import type { TrackDescriptor,CatalogLine } from "../catalog/track";
import { TrackDescriptorSchema,TrackDescriptorsSchema,CatalogLineSchema } from "../catalog/track";
import type { Context } from "golikejs/context";
import { withCancelCause, background,withCancel,ContextCancelledError,watchPromise } from "golikejs/context";
import { TrackWriter,TrackReader,GroupReader,Frame,GroupWriter,InternalSubscribeErrorCode } from "@okutanidaichi/moqt";
import { Channel, select, send, default_ } from "golikejs/channel";
import { JsonLineDecoder, EncodedJsonChunk,JsonLineEncoder } from "../internal/json";
import type { JsonObject } from "../internal/json";
import type { EncodedChunk } from ".";

export class TrackCatalog {
    readonly descriptor: TrackDescriptor;
    #done: Promise<void>;
    #end!: () => void;
    #active: boolean = true;

    constructor(ctx: Context, descriptor: TrackDescriptor) {
        this.descriptor = descriptor;
        this.#done = new Promise<void>((resolve) => {
            this.#end = () => {
                this.#active = false;
                resolve()
            };
        });
    }

    get active(): boolean {
        return this.#active;
    }

    end(): void {
        this.#end();
    }

    async done(): Promise<void> {
        return this.#done;
    }
}

export interface CatalogEncoderInit {
    version: string;
    description?: string;
}

export class CatalogEncoder {
    readonly version: string;
    readonly description?: string;

    #tracks: Map<string, TrackCatalog> = new Map();

    #channels: Set<Channel<EncodedChunk>> = new Set();

    #encoder: JsonLineEncoder;

    constructor(init: CatalogEncoderInit) {
        this.version = init.version;
        this.description = init.description;

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
                await send(chan, chunk);
            })
        );
        
        return undefined;
    }

    async encodeTo(ctx: Promise<void>, track: TrackWriter): Promise<Error | undefined> {
        let err: Error | undefined;
        let group: GroupWriter | undefined;

        [group, err] = await track.openGroup(1n)
        if (err) {
            return new Error("Failed to open catalog group: " + err.message);
        }

        if (!group) {
            return new Error("No group returned from openGroup");
        }

        const initChunk = this.#encoder.encode([{
            version: this.version,
            description: this.description,
        }]);

        err = await group.writeFrame(initChunk)
        if (err) {
            return new Error("Failed to write catalog init: " + err.message);
        }

        let chunk: EncodedChunk | undefined;

        const existings: CatalogLine[] = Array.from(this.#tracks.values()).map(track => {
            return {active: true, track: track.descriptor};
        });
        if (existings.length > 0) {
            chunk = this.#encoder.encode(existings)
            err = await group.writeFrame(chunk);
            if (err) {
                return new Error("Failed to write existing tracks: " + err.message);
            }
        }

        const chan = new Channel<EncodedChunk>(2);
        this.#channels.add(chan);

        // Integrated encode loop
        try {
            const watchCtx = watchPromise(background(), ctx);
            let ok: boolean;
            while (true) {
                // Check context cancellation before waiting
                err = watchCtx.err();
                if (err) {
                    return err;
                }

                // Race between context cancellation and channel receive
                let chunkOrErr: [EncodedChunk | undefined, boolean] | Error;
                try {
                    chunkOrErr = await Promise.race([
                        chan.receive(),
                        ctx.then(() => new Error("context cancelled"))
                    ]);
                } catch (e) {
                    err = e instanceof Error ? e : new Error(String(e));
                    return err;
                }

                if (chunkOrErr instanceof Error) {
                    return chunkOrErr;
                }

                [chunk, ok] = chunkOrErr;
                if (!ok) {
                    break;
                }

                err = await group.writeFrame(chunk!);
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

            await group.close();
        }
    }
}

export interface CatalogReaderInit {
    version: string;
    reader: TrackReader;
}

export class CatalogDecoder {
    readonly version: string;

    #source?: TrackReader;

    #tracks: Map<string, TrackCatalog> = new Map();

    // #chan: Channel<TrackCatalog[]> = new Channel(2);
    #dests: Set<(tracks: TrackCatalog[]) => void> = new Set();

    #decoder: JsonLineDecoder = new JsonLineDecoder();

    constructor(init: CatalogReaderInit) {
        this.version = init.version;
    }

    async decodeFrom(ctx: Promise<void>, track: TrackReader): Promise<Error | undefined> {
        if (this.#source) {
            this.#source.closeWithError(0, "replaced by new source"); // TODO: use proper error code
            this.#source = undefined;
        }

        const [group, err] = await track.acceptGroup(ctx)

        if (err) {
            return err;
        }

        this.#source = track;

        // Integrated handle logic
        let error: Error | undefined;
        let frame: Frame | undefined;

        try {
            const watchCtx = watchPromise(background(), ctx);
            error = watchCtx.err();
            if (error) {
                return error;
            }

            let isInit = true;

            while (true) {
                error = watchCtx.err();
                if (error) {
                    return error;
                }

                [frame, error] = await group!.readFrame();
                if (error) {
                    return error;
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
                    return e instanceof Error ? e : new Error(String(e));
                }

                if (isInit) {
                    const { success, data: init } = CatalogInitSchema.safeParse(lines.pop());
                    if (!success) {
                        return new Error("Invalid catalog init data");
                    }

                    if (this.version !== init.version) {
                        return new Error(`Catalog version mismatch: expected ${this.version}, got ${init.version}`);
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
                        const trackCatalog = new TrackCatalog(group.context, track.track);
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
            error = e instanceof Error ? e : new Error(String(e));
            return error;
        } finally {
            await group.cancel(InternalSubscribeErrorCode, "CatalogReader is closed");
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

    // async accept(ctx: Promise<void>): Promise<TrackCatalog[] | Error> {
    //     // Use select-like behavior: wait for either context cancellation or channel receive
    //     try {
    //         return await Promise.race([
    //             ctx.then(() => ContextCancelledError),
    //             this.#chan.receive().then(([desc, ok]) => {
    //                 if (!ok) {
    //                     return new Error("CatalogDecoder channel closed");
    //                 }
    //                 return desc;
    //             })
    //         ]);
    //     } catch (e) {
    //         return e instanceof Error ? e : new Error(String(e));
    //     }
    // }
}
