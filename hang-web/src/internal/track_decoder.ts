import type {
    TrackReader,
    GroupSequence,
    Frame,
    SubscribeErrorCode,
} from "@okutanidaichi/moqt";
import {
    InternalSubscribeErrorCode,
} from "@okutanidaichi/moqt";
import { withCancelCause, background } from "@okutanidaichi/moqt/internal";
import type { Context, CancelCauseFunc } from "@okutanidaichi/moqt/internal";
import type { TrackCache } from "./cache";

export interface TrackDecoder<T> {
    decodeTo(dest: WritableStreamDefaultWriter<T>): Promise<Error | undefined>;
    close(): void;
    closeWithError(code: SubscribeErrorCode, reason: string): void;
}

export class NoOpTrackDecoder implements TrackDecoder<GroupedFrame> {
    #source: TrackReader;
    #dests: Set<WritableStreamDefaultWriter<GroupedFrame>> = new Set();

    #ctx: Context;
    #cancelCtx: CancelCauseFunc;

    constructor(source: TrackReader) {
        this.#source = source;
        [this.#ctx, this.#cancelCtx] = withCancelCause(background());

        queueMicrotask(() => this.#next());
    }

    #next(): void {
        if (this.#ctx.err() !== undefined) {
            return;
        }

        this.#source.acceptGroup(this.#ctx.done()).then(
            async (result) => {
                if (result === undefined) {
                    return;
                }

                let [group, err] = result;
                if (err) {
                    this.closeWithError(InternalSubscribeErrorCode, err.message);
                    return;
                }

                let frame: Frame | undefined;
                const groupSequence = group!.groupSequence;
                while (true) {
                    const result = await Promise.race([
                        group!.readFrame(),
                        this.#ctx.done(),
                    ]);
                    if (result === undefined) {
                        break;
                    }

                    [frame, err] = result;
                    if (err || !frame) {
                        break;
                    }

                    for (const dest of this.#dests) {
                        dest.write({ groupSequence, frame });
                    }
                }

                if (!this.#ctx.err()) {
                    queueMicrotask(() => this.#next());
                }
            },
        );
    }

    async decodeTo(dest: WritableStreamDefaultWriter<GroupedFrame>): Promise<Error | undefined> {
        const err = this.#ctx.err();
        if (err !== undefined) {
            return err;
        }

        this.#dests.add(dest);
    }

    // close() and closeWithError() do not close the underlying source,
    // Callers should close the source to release resources.
    close(): void {
        if (this.#ctx.err() !== undefined) {
            return;
        }

        const cause = new Error("no-op decoder closed");
        this.#cancelCtx(cause);

        const doneFrame: ReadableStreamReadResult<GroupedFrame> = {
            value: undefined,
            done: true,
        };

        for (const dest of this.#dests) {
            dest.close();
        }
        this.#dests.clear();
    }

    closeWithError(code: SubscribeErrorCode, reason: string): void {
        if (this.#ctx.err() !== undefined) {
            return;
        }

        const cause = new Error(`no-op decoder closed: [${code}] ${reason}`);
        this.#cancelCtx(cause);

        const doneFrame: ReadableStreamReadResult<GroupedFrame> = {
            value: undefined,
            done: true,
        };

        for (const dest of this.#dests) {
            dest.abort(cause);
        }
        this.#dests.clear();
    }
}

export interface GroupedFrame {
    groupSequence: GroupSequence;
    frame: Frame;
}