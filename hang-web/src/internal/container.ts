import { writeVarint, varintLen, readVarint } from "@okutanidaichi/moqt/io";
import type { Frame } from "@okutanidaichi/moqt";

export class EncodedContainer implements Frame {
    chunk: EncodedChunk

    constructor(chunk: EncodedChunk) {
        this.chunk = chunk;
    }

    get byteLength(): number {
        return this.chunk.byteLength;
    }

    copyTo(dest: AllowSharedBufferSource): void {
        // Normalize destination into a Uint8Array view safely so we don't rely
        // on ambiguous constructor overloads.
        // AllowSharedBufferSource may be Uint8Array, ArrayBuffer, or other
        // ArrayBufferView types (e.g. DataView, Int8Array). Handle those.
        let view: Uint8Array;
        // Check total available length where possible
        if ((dest as { byteLength?: number }).byteLength === undefined) {
            // If dest doesn't expose byteLength, we still proceed and rely on
            // runtime checks below when creating views.
        } else if ((dest as { byteLength: number }).byteLength < this.chunk.byteLength + varintLen(this.chunk.byteLength)) {
            throw new Error("Destination buffer is too small");
        }

        if (dest instanceof Uint8Array) {
            view = dest;
        } else if (dest instanceof ArrayBuffer || (typeof SharedArrayBuffer !== "undefined" && dest instanceof SharedArrayBuffer)) {
            // Handle ArrayBuffer and SharedArrayBuffer
            view = new Uint8Array(dest as ArrayBufferLike);
        } else if (ArrayBuffer.isView(dest)) {
            const v = dest as ArrayBufferView;
            // ArrayBufferView always has byteOffset and byteLength
            view = new Uint8Array(v.buffer, v.byteOffset, v.byteLength);
        } else {
            throw new Error("Unsupported destination type");
        }

        writeVarint(view, this.chunk.byteLength);
        this.chunk.copyTo(view.subarray(varintLen(this.chunk.byteLength)));
    }
}

export interface EncodedChunk {
    type: string
    byteLength: number
    timestamp: number
    duration?: number | null
    copyTo(dest: AllowSharedBufferSource): void
}


export function cloneChunk(chunk: EncodedChunk): EncodedChunk {
    const buffer = new Uint8Array(chunk.byteLength);
    chunk.copyTo(buffer);

    const clone = {
        type: chunk.type,
        byteLength: chunk.byteLength,
        timestamp: chunk.timestamp,
        duration: chunk.duration,
        buffer: buffer,
        copyTo(dest: AllowSharedBufferSource): void {
            if (dest.byteLength < this.byteLength) {
                throw new RangeError("Destination buffer is too small");
            }

            let view: Uint8Array;
            if (dest instanceof Uint8Array) {
                view = dest;
            } else if (dest instanceof ArrayBuffer || (typeof SharedArrayBuffer !== "undefined" && dest instanceof SharedArrayBuffer)) {
                view = new Uint8Array(dest as ArrayBufferLike);
            } else if (ArrayBuffer.isView(dest)) {
                const v = dest as ArrayBufferView;
                view = new Uint8Array(v.buffer, v.byteOffset, v.byteLength);
            } else {
                throw new Error("Unsupported destination type");
            }

            view.set(this.buffer);
        }
    };

    return clone;
}

export interface EncodeDestination {
    output: (chunk: EncodedChunk) => Promise<Error | undefined>;
    done: Promise<void>;
}

function decodeContainer(frame: Frame): EncodedChunk {

}