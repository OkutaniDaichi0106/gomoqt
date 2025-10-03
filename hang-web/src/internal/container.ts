import { writeVarint, varintLen, readVarint, type Source } from "@okutanidaichi/moqt/io";

export class EncodedContainer implements Source {
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
    type: "key" | "delta"
    byteLength: number
    timestamp?: number
    copyTo(dest: AllowSharedBufferSource): void
}
