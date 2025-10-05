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
import type { Context, CancelCauseFunc } from "golikejs/context";
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

    #ctx: Context;
    #cancelCtx: CancelCauseFunc;

    constructor(init: TrackDecoderInit<EncodedChunk>) {
        this.#dests.set(init.destination, init.destination.getWriter());
        const [ctx, cancelCtx] = withCancelCause(background());
        this.#ctx = ctx;
        this.#cancelCtx = cancelCtx;
    }

    get decoding(): boolean {
        return this.#dests.size > 0;
    }

    #next(): void {
        if (this.#ctx.err()) {
            return;
        }
        if (this.#source === undefined) {
            return;
        }
        if (!this.decoding) {
            return;
        }

        this.#source.acceptGroup(this.#ctx.done()).then(
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
                        this.#ctx.done(),
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

                if (!this.#ctx.err()) {
                    queueMicrotask(() => this.#next());
                }
            },
        );
    }

    async decodeFrom(ctx: Promise<void>, source: TrackReader): Promise<Error | undefined> {
        if (this.#ctx.err()) {
            return this.#ctx.err()!;
        }

        if (this.#source !== undefined) {
            console.warn("[NoOpTrackDecoder] source already set. replacing...");

            await this.#source.closeWithError(InternalSubscribeErrorCode, "source was overwritten");
        }

        this.#source = source;

        queueMicrotask(() => this.#next());

        await Promise.race([
            source.context.done(),
            this.#ctx.done(),
            ctx,
        ]);

        return this.#ctx.err() || source.context.err() || undefined;
    }

    async tee(dest: WritableStream<EncodedChunk>): Promise<Error | undefined> {
        const err = this.#ctx.err();
        if (err !== undefined) {
            return Promise.resolve(err);
        }

        if (this.#dests.has(dest)) {
            return Promise.resolve(new Error("destination already set"));
        }

        const writer = dest.getWriter();
        this.#dests.set(dest, writer);

        try {
            await Promise.race([
                writer.closed,
                this.#ctx.done(),
            ]);
        } catch (e) {
            // Clean up the destination only on error
            if (this.#dests.has(dest)) {
                this.#dests.delete(dest);
            }
            return new Error("destination closed with error");
        }

        return this.#ctx.err();
    }

    async close(cause?: Error): Promise<void> {
        if (!this.#ctx.err()) {
            this.#cancelCtx(cause);
        }

        await Promise.allSettled(Array.from(this.#dests,
            ([dest, writer]) => {
                // Just release the lock, do not close the stream
                writer.releaseLock();
            }
        ));

        this.#dests.clear();

        // Clear the source to avoid races where an in-flight acceptGroup
        // resolves and tries to operate on a closed/overwritten source.
        this.#source = undefined;
    }
}

export interface TrackDecoderInit<T> {
    destination: WritableStream<T>;
    // cache?: TrackCache;
}