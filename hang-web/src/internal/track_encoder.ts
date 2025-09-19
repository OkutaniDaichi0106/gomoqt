import type {
    TrackWriter,
    GroupSequence,
    GroupWriter,
    SubscribeErrorCode,
} from "@okutanidaichi/moqt";
import {
    InternalSubscribeErrorCode,
} from "@okutanidaichi/moqt";
import type { TrackCache } from "./cache";
import { withCancelCause, background,ContextCancelledError } from "@okutanidaichi/moqt/internal";
import type { Context, CancelCauseFunc } from "@okutanidaichi/moqt/internal";
import type { GroupedFrame } from ".";

export interface TrackEncoder<T> {
    encodeTo(dest: TrackWriter): Promise<Error | undefined>;
    preview(dest?: WritableStreamDefaultWriter<T>): void;
    close(): void;
    closeWithError(code: SubscribeErrorCode, message: string): void;
}

export interface TrackEncoderInit<T> {
    source: ReadableStreamDefaultReader<T>;
    startGroupSequence?: GroupSequence;
    cache?: new () => TrackCache;
}

export class NoOpTrackEncoder implements TrackEncoder<GroupedFrame> {
    #source: ReadableStreamDefaultReader<GroupedFrame>;

    #latestGroupSequence: GroupSequence;
    #currentGroups: Map<TrackWriter, GroupWriter | undefined> = new Map();
    cache?: TrackCache;

    #previewer?: WritableStreamDefaultWriter<GroupedFrame>;

    #ctx: Context;
    #cancelCtx: CancelCauseFunc;

    constructor(init: TrackEncoderInit<GroupedFrame>) {
        this.#source = init.source;
        this.#latestGroupSequence = init.startGroupSequence ?? 0n;
        this.cache = init.cache ? new init.cache() : undefined;
        [this.#ctx, this.#cancelCtx] = withCancelCause(background());// TODO: need?
    }

    #next(): void {
        if (this.#ctx.err() !== undefined) return;

        Promise.race<ReadableStreamReadResult<GroupedFrame> | void>([
            this.#source.read(),
            this.#ctx.done(),
        ]).then(async (result) => {
            if (result === undefined) {
                // Context was cancelled
                throw this.#ctx.err() || ContextCancelledError;
            }

            const { done, value: data } = result;
            if (done) {
                this.#previewer?.close();
                return;
            }

            await this.#previewer?.write(data).catch(err => {
                this.#previewer?.abort(err);
                this.#previewer = undefined;
            });

            if (data.groupSequence > this.#latestGroupSequence) {
                this.#latestGroupSequence = data.groupSequence;
            }

            const promises: Promise<void>[] = [];
            for (const [writer, group] of this.#currentGroups) {
                if (!group) {
                    const p = writer.openGroup(this.#latestGroupSequence).then(async ([g, err]) => {
                        if (err) throw err;
                        this.#currentGroups.set(writer, g);
                        err = await g?.writeFrame(data.frame);
                        if (err) console.error("Error writing frame:", err);
                    });
                    promises.push(p);
                } else if (group.groupSequence === this.#latestGroupSequence) {
                    const p = group.writeFrame(data.frame).then(err => { if (err) console.error("Error writing frame:", err); });
                    promises.push(p);
                } else if (this.#latestGroupSequence > group.groupSequence) {
                    this.cache?.flush(group).then(() => group.close()).catch(err => group.cancel(InternalSubscribeErrorCode, err.message)).finally(() => {
                        this.#currentGroups.set(writer, undefined);
                    });
                }
            }

            this.cache?.append(this.#latestGroupSequence, data.frame);

            await Promise.all(promises);

            if (!this.#ctx.err()) {
                queueMicrotask(() => this.#next());
            }
        }).catch((err) => {
            this.#previewer?.abort(err);
            this.closeWithError(InternalSubscribeErrorCode, err.message ?? String(err));
        });
    }

    async encodeTo(dest: TrackWriter): Promise<Error | undefined> {
        if (this.#ctx.err() !== undefined) {
            return this.#ctx.err();
        }
        if (this.#currentGroups.has(dest)) {
            console.warn("given TrackWriter is already being encoded to");
            return;
        }

        this.#currentGroups.set(dest, undefined);
        if (this.#currentGroups.size === 1) {
            queueMicrotask(() => this.#next());
        }

        await Promise.race([
            dest.context.done(),
            this.#ctx.done(),
        ]);

        return this.#ctx.err() || dest.context.err() || ContextCancelledError;
    }

    preview(dest?: WritableStreamDefaultWriter<GroupedFrame>): void {
        if (this.#ctx.err() !== undefined) {
            return;
        }
        this.#previewer = dest;
    }

    // close() and closeWithError() do not close the underlying source,
    // Callers should close the source to release resources.
    close(): void {
        if (this.#ctx.err() !== undefined) {
            return;
        }

        const cause = new Error("no-op encoder closed");
        this.#cancelCtx(cause);
        for (const [tw] of this.#currentGroups) {
            tw.close();
        }

        this.#currentGroups.clear();
        this.cache?.close();
    }

    closeWithError(code: SubscribeErrorCode, message: string): void {
        if (this.#ctx.err() !== undefined) {
            return;
        }

        const cause = new Error(`no-op encoder closed: [${code}] ${message}`);
        this.#cancelCtx(cause);

        for (const [tw] of this.#currentGroups) {
            tw.closeWithError(code, message);
        }

        this.#currentGroups.clear();
        this.cache?.closeWithError(message);
    }
}