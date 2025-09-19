import type {
    TrackWriter,
    TrackReader,
    GroupWriter,
    GroupSequence,
    SubscribeErrorCode,
} from "@okutanidaichi/moqt";
import {
    PublishAbortedErrorCode,
    InternalSubscribeErrorCode,
} from "@okutanidaichi/moqt";
import { readVarint } from "@okutanidaichi/moqt/io";
import {withCancelCause, background,ContextCancelledError,withPromise } from "@okutanidaichi/moqt/internal";
import  type {Context, CancelCauseFunc } from "@okutanidaichi/moqt/internal";
import { EncodedContainer } from "./container";
import { EncodeErrorCode } from "./error";
import type { TrackEncoder, TrackEncoderInit } from "./track_encoder";
import type { TrackCache } from "./cache";
import type { TrackDecoder } from ".";

// Group rollover max latency (microseconds in Video; here we use ms timestamp from AudioEncoder chunk)
const MAX_AUDIO_LATENCY = 100; // 100ms

export class AudioEncodeStream implements TrackEncoder<AudioData> {
    #encoder: AudioEncoder;
    #source: ReadableStreamDefaultReader<AudioData>;
    #decoderConfig: AudioDecoderConfig | undefined = undefined;
    #latestGroupSequence: GroupSequence;
    #latestGroupTimestamp: number = 0;
    #currentGroups: Map<TrackWriter, GroupWriter | undefined> = new Map();

    #previewer?: WritableStreamDefaultWriter<AudioData>;

    cache?: TrackCache;

    #ctx: Context;
    #cancelCtx: CancelCauseFunc;

    constructor(init: TrackEncoderInit<AudioData>) {
        this.#source = init.source;
        this.#latestGroupSequence = init.startGroupSequence ?? 0n;
        this.cache = init.cache ? new init.cache() : undefined;

        [this.#ctx, this.#cancelCtx] = withCancelCause(background());

        this.#encoder = new AudioEncoder({
            output: async (chunk, metadata) => {
                if (metadata?.decoderConfig) {
                    this.#decoderConfig = metadata.decoderConfig;
                }

                // (Original code enforced key-only) Keep for safety
                if (chunk.type !== "key") {
                    console.warn("Ignoring non-key audio chunk");
                    return;
                }

                if (chunk.timestamp - this.#latestGroupTimestamp > MAX_AUDIO_LATENCY) {
                    this.#latestGroupSequence += 1n;
                    this.#latestGroupTimestamp = chunk.timestamp;
                }

                if (this.#currentGroups.size === 0) return;

                const container = new EncodedContainer(chunk);

                const promises: Promise<void>[] = [];
                for (const [writer, group] of this.#currentGroups) {
                    if (!group) {
                        const p = writer.openGroup(this.#latestGroupSequence).then(async ([g, err]) => {
                            if (err) throw err;
                            this.#currentGroups.set(writer, g);
                            err = await g?.writeFrame(container);
                            if (err) console.error("Error writing frame:", err);
                        });
                        promises.push(p);
                    } else if (group.groupSequence === this.#latestGroupSequence) {
                        const p = group.writeFrame(container).then(err => { if (err) console.error("Error writing frame:", err); });
                        promises.push(p);
                    } else if (this.#latestGroupSequence > group.groupSequence) {
                        // Flush cache for old group then close it
                        this.cache?.flush(group).then(() => group.close()).catch(err => group.cancel(InternalSubscribeErrorCode, err.message)).finally(() => {
                            this.#currentGroups.set(writer, undefined);
                        });
                    }
                }

                this.cache?.append(this.#latestGroupSequence, container);
                await Promise.all(promises);
            },
            error: (error) => {
                console.error("Audio encoding error:", error);
                this.closeWithError(EncodeErrorCode, error.message);
            }
        });
    }

    #next(): void {
        if (this.#ctx.err() !== undefined) {
            return;
        }

        Promise.race([
            this.#source.read(),
            this.#ctx.done(),
        ]).then(async (result) => {
            // When context is done, result is undefined
            if (result === undefined) {
                // Context was cancelled
                // Just in case, signal the previewer end of stream
                this.#previewer?.abort(this.#ctx.err() || ContextCancelledError);
                this.#previewer = undefined;
                return;
            }

            const { done, value: frame } = result;
            if (done) {
                this.#previewer?.close();
                this.#previewer = undefined;
                return;
            }

            await this.#previewer?.write(frame).catch((err) => {
                this.#previewer?.abort(err);
                this.#previewer = undefined;
            });

            this.#encoder.encode(frame);

            frame.close();

            // Schedule next read
            queueMicrotask(() => this.#next());
        }).catch((err) => {
            console.error("audio next error", err);
            this.#previewer?.abort(err);
            this.closeWithError(EncodeErrorCode, err.message ?? String(err));
        });
    }

    configure(config: AudioEncoderConfig): void {
        this.#encoder.configure(config);
    }

    async encodeTo(dest: TrackWriter): Promise<Error | undefined> {
        if (this.#ctx.err() !== undefined) {
            return this.#ctx.err();
        }
        if (this.#currentGroups.has(dest)) {
            console.warn("destination already encoding");
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

    preview(dest: WritableStreamDefaultWriter<AudioData>): Error | undefined {
        if (this.#ctx.err() !== undefined) {
            return this.#ctx.err();
        }
        this.#previewer = dest;
    }

    // close() and closeWithError() do not close the underlying source,
    // Callers should close the source to release resources.
    close(): void {
        if (this.#ctx.err() !== undefined) {
            return;
        }

        const cause = new Error("audio encoder closed");
        this.#cancelCtx(cause);

        this.#encoder.close();

        for (const [tw] of this.#currentGroups) {
            tw.close();
        }

        this.#currentGroups.clear();
        this.cache?.close();
    }

    closeWithError(code: SubscribeErrorCode, reason: string): void {
        if (this.#ctx.err() !== undefined) {
            return;
        }

        const cause = new Error(`audio encoder closed with error: [${code}] ${reason}`);
        this.#cancelCtx(cause);

        this.#encoder.close();

        for (const [tw] of this.#currentGroups) {
            tw.closeWithError(code, `audio encoder closed: ${reason}`);
        }

        this.#currentGroups.clear();
        this.cache?.closeWithError(reason);
    }
}

export class AudioDecodeStream implements TrackDecoder<AudioData> {
    #decoder: AudioDecoder;
    #source: TrackReader;
    #frameCount = 0;
    #dests: Set<WritableStreamDefaultWriter<AudioData>> = new Set();

    #ctx: Context;
    #cancelCtx: CancelCauseFunc;

    constructor(reader: TrackReader) {
        this.#source = reader;
        [this.#ctx, this.#cancelCtx] = withCancelCause(background());
        this.#decoder = new AudioDecoder({
            output: (frame: AudioData) => {
                for (const dest of this.#dests) {
                    dest.write(frame);
                }
            },
            error: (error) => {
                // Log only; avoid invoking close to prevent duplicate cleanup.
                console.error("Audio decoding error (no auto-close):", error);
            }
        });
    }

    #next(): void {
        if (this.#ctx.err() !== undefined) {
            return;
        }

        this.#source.acceptGroup(this.#ctx.done()).then(async (result) => {
            // When context is done, result is undefined
            if (result === undefined) {
                return;
            }

            const [group, err] = result;

            this.#frameCount = 0;

            if (err) {
                console.error("Error accepting group:", err);
                return;
            }

            while (true) {
                const [frame, err] = await group!.readFrame();
                if (err) {
                    console.error("Error reading frame:", err);
                    break;
                }

                this.#frameCount++;

                const [timestamp, headerSize] = readVarint(frame!.bytes);

                const chunk = new EncodedAudioChunk({
                    type: this.#frameCount < 2 ? "key" : "delta",
                    timestamp: timestamp,
                    data: frame!.bytes.subarray(headerSize),
                });

                this.#decoder.decode(chunk);
            }

            if (!this.#ctx.err()) {
                queueMicrotask(() => this.#next());
            }
        }).catch(error => {
            console.error("Error decoding audio:", error);
            this.closeWithError(InternalSubscribeErrorCode, error.message ?? String(error));
        });
    }

    configure(config: AudioDecoderConfig) {
        this.#decoder.configure(config);
    }

    async decodeTo(dest: WritableStreamDefaultWriter<AudioData>): Promise<Error | undefined> {
        let err = this.#ctx.err();
        if (err !== undefined) {
            return err;
        }

        if (this.#dests.has(dest)) {
            console.warn("given WritableStreamDefaultWriter is already being decoded to");
            return;
        }

        this.#dests.add(dest);

        if (this.#dests.size === 1) {
            queueMicrotask(() => this.#next());
        }

        err = await Promise.race([
            dest.closed.catch(e => e),
            this.#ctx.done(),
        ]);

        return this.#ctx.err() || err || ContextCancelledError;
    }

    // close() and closeWithError() do not close the underlying source,
    // Callers should close the source to release resources.
    close(): void {
        if (this.#ctx.err() !== undefined) {
            return;
        }

        const cause = new Error("audio decoder closed");
        this.#cancelCtx(cause);

        this.#decoder.close();

        for (const dest of this.#dests) {
            dest.close();
        }
        this.#dests.clear();
    }

    closeWithError(code: SubscribeErrorCode, reason: string): void {
        if (this.#ctx.err() !== undefined) {
            return;
        }

        const cause = new Error(`audio decoder closed with error: [${code}] ${reason}`);
        this.#cancelCtx(cause);

        this.#decoder.close();

        for (const dest of this.#dests) {
            dest.abort(cause);
        }
        this.#dests.clear();
    }
}
