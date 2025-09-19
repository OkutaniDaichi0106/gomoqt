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
import { withCancelCause, background,ContextCancelledError } from "@okutanidaichi/moqt/internal";
import type { Context, CancelCauseFunc, } from "@okutanidaichi/moqt/internal";
import { EncodedContainer, type EncodedChunk } from "./container";
import type { TrackEncoder, TrackEncoderInit } from "./track_encoder";
import { EncodeErrorCode } from "./error";
import type { z } from "zod";
import type { TrackCache } from "./cache";
import type { TrackDecoder } from ".";
import { th } from "zod/v4/locales";

export class JsonEncoder {
    #output: (chunk: EncodedJsonChunk, metadata?: EncodedJsonChunkMetadata) => void;
    #error: (error: Error) => void;
    #replacer?: (key: string, value: any) => any;
    #space?: string | number;
    #meta?: EncodedJsonChunkMetadata;
    #textEncoder: TextEncoder = new TextEncoder();
    #buffer: Uint8Array | undefined;

    constructor(init: JsonEncoderInit) {
        this.#output = init.output;
        this.#error = init.error;
    }

    configure(config: JsonEncoderConfig): void {
        this.#replacer = config.replacer?.callback ?? this.#replacer;
        this.#space = config.space ?? this.#space;
        this.#meta = {
            replaceRule: config.replacer?.rule,
            space: this.#space,
        }
    }

    encode(value: any, patch: boolean = false): void {
        const type = patch ? "delta" : "key";
        const str = JSON.stringify(value, this.#replacer, this.#space);

        // Allocate or resize buffer efficiently, following Go's allocation rules
        if (!this.#buffer) {
            // Initialize buffer with required size
            this.#buffer = new Uint8Array(str.length);
        } else if (this.#buffer.length < str.length) {
            // Grow buffer: use at least double the current size or required size
            const newSize = Math.max(this.#buffer.length * 2, str.length);
            this.#buffer = new Uint8Array(newSize);
        }

        // Encode into the buffer
        const {written} = this.#textEncoder.encodeInto(str, this.#buffer);
        const chunk = new EncodedJsonChunk({ type, data: this.#buffer.subarray(0, written) });
        this.#output(chunk, this.#meta);

        // Reset metadata
        if (this.#meta) {
            this.#meta = undefined;
        }
    }

    close(): void {
        // No resources to release in current implementation
    }
}

export interface JsonEncoderInit {
    output: (chunk: EncodedJsonChunk, metadata?: any) => void;
    error: (error: Error) => void;
}

export interface JsonEncoderConfig {
    space?: string | number;
    replacer?: {
        rule: string;
        callback: (key: string, value: any) => any
    };
    nextFile?: boolean;
}

export class EncodedJsonChunk {
    type: "key" | "delta";
    data: Uint8Array;

    constructor(init: EncodedJsonChunkInit) {
        this.type = init.type;
        this.data = init.data;
    }

    get byteLength() {
        return this.data.byteLength;
    }

    copyTo(target: Uint8Array) {
        target.set(this.data);
    }
}

export interface EncodedJsonChunkInit {
    type: "key" | "delta";
    data: Uint8Array;
}

export interface EncodedJsonChunkMetadata {
    replaceRule?: string;
    space?: string | number;
}

export class JsonDecoder {
    #output: (chunk: any) => void;
    #error: (error: Error) => void;
    #config?: JsonDecoderConfig;
    #decoder: TextDecoder = new TextDecoder();

    constructor(init: JsonDecoderInit) {
        this.#output = init.output;
        this.#error = init.error;
    }

    configure(config: JsonDecoderConfig): void {
        this.#config = config;
    }

    decode(chunk: EncodedJsonChunk): void {
        try {
            const json = JSON.parse(this.#decoder.decode(chunk.data), this.#config?.reviver);
            this.#output(json);
        } catch (error) {
            if (error instanceof Error) {
                this.#error(error);
            } else {
                this.#error(new Error(String(error)));
            }
        }
    }

    close(): void {

    }
}

export interface JsonDecoderInit {
    output: (chunk: string) => void;
    error: (error: Error) => void;
    reviver?: (key: string, value: any) => any;
}

export interface JsonDecoderConfig {
    reviver?: (key: string, value: any) => any;
}

export function replaceBigInt(key: string, value: any): any {
	if (typeof value === "bigint") {
		return value.toString();
	}
	return value;
}

export function reviveBigInt(key: string, value: any): any {
	if (typeof value === "string" && /^\d+$/.test(value)) {
		try {
			return BigInt(value);
		} catch {
			return value;
		}
	}
	return value;
}

export class JsonEncodeStream implements TrackEncoder<z.ZodType<any>> {
    encoder: JsonEncoder;
    #source: ReadableStreamDefaultReader<z.ZodType<any>>;
    #latestGroupSequence: GroupSequence;
    #currentGroups: Map<TrackWriter, GroupWriter | undefined> = new Map();

    #previewer?: WritableStreamDefaultWriter<z.ZodType<any>>;

    cache?: TrackCache;

    #ctx: Context;
    #cancelCtx: CancelCauseFunc;

    constructor(init: TrackEncoderInit<z.ZodType<any>>) {
        this.#source = init.source;
        this.#latestGroupSequence = init.startGroupSequence ?? 0n;
        this.cache = init.cache ? new init.cache() : undefined;
        [this.#ctx, this.#cancelCtx] = withCancelCause(background());

        this.encoder = new JsonEncoder({
            output: async (chunk) => {
                if (this.#ctx.err() !== undefined) return;
                if (this.#currentGroups.size === 0) return;

                if (chunk.type === 'key') {
                    this.#latestGroupSequence += 1n;
                }

                const promises: Promise<void>[] = [];
                const encodedChunk: EncodedChunk = {
                    byteLength: chunk.byteLength,
                    timestamp: Date.now(),
                    copyTo: (dest: AllowSharedBufferSource) => {
                        let view: Uint8Array;
                        if (dest instanceof Uint8Array) {
                            view = dest;
                        } else if (dest instanceof ArrayBuffer || (typeof SharedArrayBuffer !== 'undefined' && dest instanceof SharedArrayBuffer)) {
                            view = new Uint8Array(dest as ArrayBufferLike);
                        } else if (ArrayBuffer.isView(dest)) {
                            const v = dest as ArrayBufferView;
                            view = new Uint8Array(v.buffer, v.byteOffset, v.byteLength);
                        } else {
                            throw new Error('Unsupported destination type');
                        }
                        chunk.copyTo(view);
                    }
                };
                const container = new EncodedContainer(encodedChunk);
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
                        const p = group.writeFrame(container).then((err) => {
                            if (err) console.error("Error writing frame:", err);
                        });
                        promises.push(p);
                    } else if (this.#latestGroupSequence > group.groupSequence) {
                        this.cache?.flush(group).then(() => {
                            group.close();
                        }).catch((err) => {
                            group.cancel(InternalSubscribeErrorCode, err.message);
                        }).finally(() => {
                            this.#currentGroups.set(writer, undefined);
                        });
                    }
                }

                this.cache?.append(this.#latestGroupSequence, container);
                await Promise.all(promises);
            },
            error: (error) => {
                console.error("JSON encoding error:", error);
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
        ]).then(async (result: any) => {
            if (result === undefined) {
                // Context was cancelled
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

            await this.#previewer?.write(frame).catch((err)=>{
                this.#previewer?.abort(err);
                this.#previewer = undefined;
            });

            // Pass through to encoder (JsonEncoder handles stringify)
            this.encoder.encode(frame, false);

            if (!this.#ctx.err()) {
                queueMicrotask(() => this.#next());
            }
        }).catch(err => {
            console.error("json next error", err);
            this.#previewer?.abort(err);
            this.closeWithError(EncodeErrorCode, err.message ?? String(err));
        });
    }

    configure(config: JsonEncoderConfig) {
        this.encoder.configure(config);
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

    preview(dest?: WritableStreamDefaultWriter<z.ZodType<any>>): void {
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

        const cause = new Error("json stream encoder closed");
        this.#cancelCtx(cause);

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

        const cause = new Error(`json stream encoder closed with error: [${code}] ${reason}`);
        this.#cancelCtx(cause);

        for (const [tw] of this.#currentGroups) {
            tw.closeWithError(code, reason);
        }

        this.#currentGroups.clear();
        this.cache?.closeWithError(reason);
    }
}

export class JsonDecodeStream implements TrackDecoder<z.ZodType<any>> {
    #decoder: JsonDecoder;
    #source: TrackReader;
    #frameCount: number = 0;
    #dests: Set<WritableStreamDefaultWriter<any>> = new Set();

    #ctx: Context;
    #cancelFunc: CancelCauseFunc;

    constructor(reader: TrackReader) {
        this.#source = reader;
        [this.#ctx, this.#cancelFunc] = withCancelCause(background());

        this.#decoder = new JsonDecoder({
            output: (chunk: any) => {
                for (const dest of this.#dests) {
                    dest.write(chunk);
                }
            },
            error: (error) => {
                console.error("JSON decoding error (no auto-close):", error);
            }
        });
    }

    #next(): void {
        if (this.#ctx.err() !== undefined) {
            return;
        }

        this.#source.acceptGroup(this.#ctx.done()).then(async (result) => {
            if (result === undefined) {
                return;
            }

            const [group, err] = result;
            if (err) {
                console.error("Error accepting group:", err);
                return;
            }
            this.#frameCount = 0;

            while (true) {
                const [frame, err] = await group!.readFrame();
                if (err) {
                    console.error("Error reading frame:", err);
                    break;
                }

                this.#frameCount++;

                const chunk = new EncodedJsonChunk({
                    type: this.#frameCount < 2 ? "key" : "delta",
                    data: frame!.bytes,
                });

                this.#decoder.decode(chunk);
            }

            if (!this.#ctx.err()) {
                queueMicrotask(() => this.#next());
            }
        }).catch(error => {
            console.error("Error decoding json:", error);
            this.closeWithError(InternalSubscribeErrorCode, error.message);
        });
    }

    configure(config: JsonDecoderConfig) {
        this.#decoder.configure(config);
    }

    async decodeTo(dest: WritableStreamDefaultWriter<any>): Promise<Error | undefined> {
        let err = this.#ctx.err();
        if (err) {
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

        const cause = new Error("json stream decoder closed");
        this.#cancelFunc(cause);

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

        const cause = new Error(`json stream decoder closed with error: [${code}] ${reason}`);
        this.#cancelFunc(cause);

        this.#decoder.close();
        for (const dest of this.#dests) {
            dest.abort(cause);
        }
        this.#dests.clear();
    }
}