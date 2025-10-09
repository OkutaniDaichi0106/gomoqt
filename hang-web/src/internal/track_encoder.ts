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
import { GroupCache } from "./cache";
import { withCancelCause, background,ContextCancelledError } from "golikejs/context";
import type { Context, CancelCauseFunc } from "golikejs/context";
// import type { EncodedChunk } from ".";
import { EncodedContainer,type EncodedChunk } from "./container";
import { EncodeErrorCode } from ".";
import type { TrackDescriptor } from "../catalog";

export interface TrackEncoder {
    encodeTo(ctx: Promise<void>, dest: TrackWriter): Promise<Error | undefined>;
    close(cause?: Error): Promise<void>;
    encoding: boolean;
}

export interface TrackEncoderInit<T> {
    source: ReadableStream<T>;
    startGroupSequence?: GroupSequence;
    cache?: new () => TrackCache;
}

// export class NoOpTrackEncoder implements TrackEncoder {
//     #source: ReadableStream<EncodedChunk>;

//     #latestGroup: GroupCache;
//     #tracks: Set<TrackWriter> = new Set();
//     cache?: TrackCache;

//     // #previewer?: WritableStreamDefaultWriter<EncodedChunk>;

//     #ctx: Context;
//     #cancelCtx: CancelCauseFunc;

//     constructor(init: TrackEncoderInit<EncodedChunk>) {
//         this.#source = init.source;
//         this.#latestGroup = new GroupCache(init.startGroupSequence ?? 0n, 0);
//         this.cache = init.cache ? new init.cache() : undefined;
//         const [ctx, cancelCtx] = withCancelCause(background());// TODO: need?
//         this.#ctx = ctx;
//         this.#cancelCtx = cancelCtx;
//     }

//     get encoding(): boolean {
//         return this.#tracks.size > 0;
//     }

//     #next(reader: ReadableStreamDefaultReader<EncodedChunk>): void {
//         if (this.#ctx.err()) {
// 			return;
// 		}
//         if (!this.encoding) {
// 			// No active tracks to encode to
//             // Just release the lock and stop reading from the reader
//             reader.releaseLock();
// 			return;
// 		}


// 		Promise.race([
// 			reader.read(),
// 			this.#ctx.done(),
// 		]).then(async (result) => {
// 			if (result === undefined) {
// 				// Context was cancelled
// 				return;
// 			}

// 			const { done, value: chunk } = result;
// 			if (done) {
// 				return;
// 			}

//             if (chunk.type === "key") {
//                 // Close previous group and start a new one
//                 this.#latestGroup.close();
//                 const nextSequence = this.#latestGroup.sequence + 1n;
//                 this.#latestGroup = new GroupCache(nextSequence, 0); // TODO: timestamp?

//                 // Open new groups for all tracks asynchronously
//                 for (const track of this.#tracks) {
//                     track.openGroup(this.#latestGroup.sequence).then(
//                         ([group, err]) => {
//                             if (err) {
//                                 console.error("moq: failed to open group:", err);
//                                 this.#tracks.delete(track);
//                                 track.closeWithError(InternalSubscribeErrorCode, err.message);
//                                 return;
//                             }

//                             // Send frames via latest group cache
//                             this.#latestGroup.flush(group!);
//                         }
//                     );
//                 }
//             }

//             // Skip encoding if no current groups
//             if (this.#tracks.size === 0) {
//                 return;
//             }

//             const container = new EncodedContainer(cloneChunk(chunk));

//             await this.#latestGroup.append(container);

// 			// Continue to the next frame
// 			if (!this.#ctx.err()) {
// 				queueMicrotask(() => this.#next(reader));
// 			}
// 		}).catch(err => {
// 			console.error("video next error", err);
// 			this.close(err);
// 		});
//     }

//     async encodeTo(ctx: Promise<void>, dest: TrackWriter): Promise<Error | undefined> {
//         if (this.#ctx.err()) {
//             return this.#ctx.err()!;
//         }

//         if (this.#tracks.has(dest)) {
//             console.warn("given TrackWriter is already being encoded to");
//             return;
//         }

//         this.#tracks.add(dest);

//         if (this.#tracks.size === 1) {
//             const reader = this.#source.getReader();
//             queueMicrotask(() => this.#next(reader));
//         }

//         await Promise.race([
//             dest.context.done(),
//             this.#ctx.done(),
//             ctx,
//         ]);

//         return this.#ctx.err() || dest.context.err() || ContextCancelledError;
//     }

//     async close(cause?: Error): Promise<void> {
//         if (!this.#ctx.err()) {
//             this.#cancelCtx(cause);
//         }


//         await Promise.allSettled(Array.from(this.#tracks,
//             (tw) => {
//                 if (cause) {
//                     return tw.closeWithError(InternalSubscribeErrorCode, cause.message);
//                 } else {
//                     tw.close();
//                 }
//             }
//         ));

//         this.#tracks.clear();
//         this.cache?.close();
//     }
// }
