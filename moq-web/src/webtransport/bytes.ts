import { MAX_VARINT1, MAX_VARINT2, MAX_VARINT4, MAX_VARINT8 } from "./len.ts";

export const MAX_BYTES_LENGTH = 1 << 30; // 1 GiB, maximum length of bytes to read

export const MAX_UINT = 0x3FFFFFFFFFFFFFFFn; // Maximum value for a 62-bit unsigned integer

export class BytesBuffer {
    #buf: Uint8Array;
    #off: number; // read offset
    #len: number; // write offset

    constructor(memory: ArrayBufferLike) {
        this.#buf = new Uint8Array(memory);
        this.#off = 0;
        this.#len = 0; // Start with an empty buffer for writing
    }

    static make(capacity: number): BytesBuffer {
        const buf = new Uint8Array(capacity);
        return new BytesBuffer(buf.buffer);
    }

    bytes(): Uint8Array {
        return this.#buf.subarray(this.#off, this.#len);
    }

    get size(): number {
        return this.#len - this.#off;
    }

    get capacity(): number {
        return this.#buf.length;
    }

    reset() {
        this.#off = 0;
        this.#len = 0;
    }

    read(buf: Uint8Array): number {
        const bytesAvailable = this.size;
        const bytesToRead = Math.min(buf.length, bytesAvailable);
        if (bytesToRead === 0) {
            return 0;
        }
        buf.set(this.#buf.subarray(this.#off, this.#off + bytesToRead));
        this.#off += bytesToRead;
        if (this.#off === this.#len) {
            this.reset();
        }
        return bytesToRead;
    }

    readUint8(): number {
        if (this.size < 1) {
            throw new Error("Not enough data to read a byte");
        }
        const value = this.#buf[this.#off]!;
        this.#off += 1;
        if (this.#off === this.#len) {
            this.reset();
        }
        return value;
    }

    write(data: Uint8Array):number {
        this.grow(data.length);
        this.#buf.set(data, this.#len);
        this.#len += data.length;
        return data.length;
    }

    writeUint8(value: number): void {
        this.grow(1);
        this.#buf[this.#len] = value;
        this.#len += 1;
    }

    grow(n: number) {
        if (n < 0) {
            throw new Error("Cannot grow buffer by a negative size");
        }
        const required = this.size + n;
        if (required > this.capacity) {
            // Create a new buffer having an enough capacity
            const newBuf = new Uint8Array(Math.max(required, this.capacity * 2));
            // Copy the existing data to the new buffer from the head
            newBuf.set(this.bytes());
            this.#buf = newBuf;
        } else if (this.#off > 0) {
            // Slide the buffer to the head
            this.#buf.copyWithin(0, this.#off, this.#len);
        }

        // Adjust the offsets
        this.#len -= this.#off;
        this.#off = 0;
    }

    reserve(n: number): Uint8Array {
        this.grow(n);
        const start = this.#len;
        const end = start + n;
        this.#len = end;
        return this.#buf.subarray(start, end);
    }
}

export function writeVarint(view: Uint8Array, num: number, offset = 0): number  {
    if (num < 0) {
        throw new Error("Varint cannot be negative");
    }


    // Choose encoding length
    if (num <= MAX_VARINT1) {
        if (view.length - offset < 1) throw new RangeError("buffer too small");
        view[offset + 0] = num;
        return 1;
    } else if (num <= MAX_VARINT2) {
        if (view.length - offset < 2) throw new RangeError("buffer too small");
        view[offset + 0] = (num >> 8) | 0x40;
        view[offset + 1] = num & 0xff;
        return 2;
    } else if (num <= MAX_VARINT4) {
        if (view.length - offset < 4) throw new RangeError("buffer too small");
        view[offset + 0] = (num >> 24) | 0x80;
        view[offset + 1] = (num >> 16) & 0xff;
        view[offset + 2] = (num >> 8) & 0xff;
        view[offset + 3] = num & 0xff;
        return 4;
    } else {
        // 8-byte case. For safety require a safe integer and use BigInt to construct bytes.
        if (num > Number.MAX_SAFE_INTEGER) {
            throw new RangeError("Number too large for writeVarint; use writeBigVarint");
        }
        if (view.length - offset < 8) throw new RangeError("buffer too small");
        const bn = BigInt(num);
        view[offset + 0] = Number((bn >> 56n) | 0xc0n);
        view[offset + 1] = Number((bn >> 48n) & 0xffn);
        view[offset + 2] = Number((bn >> 40n) & 0xffn);
        view[offset + 3] = Number((bn >> 32n) & 0xffn);
        view[offset + 4] = Number((bn >> 24n) & 0xffn);
        view[offset + 5] = Number((bn >> 16n) & 0xffn);
        view[offset + 6] = Number((bn >> 8n) & 0xffn);
        view[offset + 7] = Number(bn & 0xffn);
        return 8;
    }
}

export function writeBigVarint(view: Uint8Array, num: bigint, offset = 0): number  {
    if (num < 0n) {
        throw new Error("Varint cannot be negative");
    }

    if (num <= BigInt(MAX_VARINT1)) {
        if (view.length - offset < 1) throw new RangeError("buffer too small");
        view[offset + 0] = Number(num);
        return 1;
    } else if (num <= BigInt(MAX_VARINT2)) {
        if (view.length - offset < 2) throw new RangeError("buffer too small");
        view[offset + 0] = Number((num >> 8n) | 0x40n);
        view[offset + 1] = Number(num & 0xffn);
        return 2;
    } else if (num <= BigInt(MAX_VARINT4)) {
        if (view.length - offset < 4) throw new RangeError("buffer too small");
        view[offset + 0] = Number((num >> 24n) | 0x80n);
        view[offset + 1] = Number((num >> 16n) & 0xffn);
        view[offset + 2] = Number((num >> 8n) & 0xffn);
        view[offset + 3] = Number(num & 0xffn);
        return 4;
    } else if (num <= MAX_VARINT8) {
        if (view.length - offset < 8) throw new RangeError("buffer too small");
        view[offset + 0] = Number((num >> 56n) | 0xc0n);
        view[offset + 1] = Number((num >> 48n) & 0xffn);
        view[offset + 2] = Number((num >> 40n) & 0xffn);
        view[offset + 3] = Number((num >> 32n) & 0xffn);
        view[offset + 4] = Number((num >> 24n) & 0xffn);
        view[offset + 5] = Number((num >> 16n) & 0xffn);
        view[offset + 6] = Number((num >> 8n) & 0xffn);
        view[offset + 7] = Number(num & 0xffn);
        return 8;
    }

    throw new RangeError("Value exceeds maximum varint size");
}

export function writeUint8Array(view: Uint8Array, data: Uint8Array, offset = 0): number {
    const len = data.length;
    // varint may take up to 8 bytes; choose a temporary slice of view to write into
    // We'll attempt to write varint at the start of view and then copy data
    // Caller must ensure view.length >= varintLen + data.length. We simply throw if not enough.
    // Determine needed varint length by checking thresholds
    let headerLen: number;
    if (len <= MAX_VARINT1) {
        headerLen = 1;
    } else if (len <= MAX_VARINT2) {
        headerLen = 2;
    } else if (len <= MAX_VARINT4) {
        headerLen = 4;
    } else {
        headerLen = 8;
    }

    if (view.length - offset < headerLen + len) throw new RangeError("buffer too small");

    // write varint header into view starting at offset
    writeVarint(view, len, offset);
    // copy data after header
    view.set(data, offset + headerLen);
    return headerLen + len;
}

export function writeString(view: Uint8Array, str: string, offset = 0): number {
    const encoder = new TextEncoder();
    const data = encoder.encode(str);
    return writeUint8Array(view, data, offset);
}

export function readVarint(view: Uint8Array, offset = 0): [number, number] {
    if (offset >= view.length) {
        throw new RangeError("offset out of bounds");
    }
    const first = view[offset]!;
    const len = 1 << (first >> 6);
    if (view.length - offset < len) {
        throw new RangeError("buffer too small for varint");
    }
    let value = first & 0x3f;
    for (let i = 1; i < len; i++) {
        value = value * 256 + view[offset + i]!;
    }
    return [value, len];
}

export function readBigVarint(view: Uint8Array, offset = 0): [bigint, number] {
    if (offset >= view.length) {
        throw new RangeError("offset out of bounds");
    }
    const first = view[offset]!;
    const len = 1 << (first >> 6);
    if (view.length - offset < len) {
        throw new RangeError("buffer too small for varint");
    }
    let value = BigInt(first & 0x3f);
    for (let i = 1; i < len; i++) {
        value = value * 256n + BigInt(view[offset + i]!);
    }
    return [value, len];
}

export function readUint8Array(view: Uint8Array, offset = 0): [Uint8Array, number] {
    // read length varint first
    const [len, n] = readVarint(view, offset);
    if (len > MAX_BYTES_LENGTH) throw new RangeError("varint too large");
    const start = offset + n;
    if (view.length - start < len) throw new RangeError("buffer too small for bytes");
    const bytes = view.subarray(start, start + len);
    return [bytes, n + len];
}

export function readString(view: Uint8Array, offset = 0): [string, number] {
    const [bytes, n] = readUint8Array(view, offset);
    const str = new TextDecoder().decode(bytes);
    return [str, n];
}