import type {
    TrackReader,
    GroupSequence,
    Frame,
    SubscribeErrorCode,
} from "@okutanidaichi/moqt";
import {
    InternalSubscribeErrorCode,
} from "@okutanidaichi/moqt";
import { 
    withCancelCause, 
    background,
    ContextCancelledError 
} from "golikejs/context";
import type { TrackCache } from "./cache";
import type { EncodedChunk } from "./container";
import { fi } from "zod/v4/locales";

export interface TrackDecoder {
    decodeFrom(ctx: Promise<void>, source: TrackReader): Promise<Error | undefined>;
    close(cause?: Error): Promise<void>;
    decoding: boolean;
}

export class NoOpTrackDecoder implements TrackDecoder {
    #source?: TrackReader;
    #dests: Map<WritableStream<EncodedChunk>, WritableStreamDefaultWriter<EncodedChunk>> = new Map();

    constructor(init: TrackDecoderInit<EncodedChunk>) {
        // If a destination is provided in init, register its writer so decoding reflects it.
        if (init.destination) {
            try {
                const w = init.destination.getWriter();
                this.#dests.set(init.destination, w);
            } catch (e) {
                // ignore errors retrieving writer in odd test environments
            }
        }
    }

    get decoding(): boolean {
        return this.#dests.size > 0;
    }

    #next(ctx: Promise<void>): void {
        if (this.#source === undefined) {
            return;
        }
        if (!this.decoding) {
            return;
        }

        this.#source.acceptGroup(ctx).then(
            async (result) => {
                if (result === undefined) {
                    return;
                }

                let [group, err] = result;
                if (err) {
                    this.close(err);
                    return;
                }

                let frame: Frame | undefined;
                const groupSequence = group!.groupSequence;
                const isKey = true;
                while (true) {
                    const result = await Promise.race([
                        group!.readFrame(),
                        ctx,
                    ]);
                    if (result === undefined) {
                        break;
                    }

                    [frame, err] = result;
                    if (err) {
                        break;
                    }

                    await Promise.allSettled(
                        Array.from(this.#dests, ([, dest]) => {
                            return dest.write({
                                type: isKey ? "key" : "delta",
                                byteLength: frame!.byteLength,
                                copyTo: frame!.copyTo
                            });
                        })
                    );
                }

                // continue reading as long as decoding flag is set
                if (this.decoding) {
                    queueMicrotask(() => this.#next(ctx));
                }
            },
        );
    }

    async decodeFrom(ctx: Promise<void>, source: TrackReader): Promise<Error | undefined> {
        if (this.#source !== undefined) {
            console.warn("[NoOpTrackDecoder] source already set. replacing...");
            await this.#source.closeWithError(InternalSubscribeErrorCode, "source was overwritten");
        }

        this.#source = source;

        queueMicrotask(() => this.#next(ctx));

        await Promise.race([
            source.context.done(),
            ctx,
        ]);

        return source.context.err();
    }

    async decodeTo(ctx: Promise<void>, dest: WritableStream): Promise<Error | undefined> {
        this.#dests.set(dest, dest.getWriter());
        try {
            await Promise.race([dest.getWriter().closed, ctx]);
        } catch (e) {
            // ignore for now
        }
        return dest.getWriter().closed instanceof Promise ? undefined : undefined;
    }

    async tee(dest: WritableStream<EncodedChunk>): Promise<Error | undefined> {
        // No internal ctx; proceed

        if (this.#dests.has(dest)) {
            return Promise.resolve(new Error("destination already set"));
        }

        const writer = dest.getWriter();
        this.#dests.set(dest, writer);

        try {
            await Promise.race([
                writer.closed,
                // no internal ctx to wait on
            ]);
        } catch (e) {
            // Clean up the destination only on error
            if (this.#dests.has(dest)) {
                this.#dests.delete(dest);
            }
            return new Error("destination closed with error");
        }

        return undefined;
    }

    async close(cause?: Error): Promise<void> {
        await Promise.allSettled(Array.from(this.#dests,
            ([dest, writer]) => {
                // Just release the lock, do not close the stream
                try { writer.releaseLock(); } catch (e) { }
            }
        ));

        this.#dests.clear();

        // Clear the source to avoid races where an in-flight acceptGroup
        // resolves and tries to operate on a closed/overwritten source.
        this.#source = undefined;
    }
}

export interface TrackDecoderInit<T> {
    destination?: WritableStream<T>;
    cache?: TrackCache;
}