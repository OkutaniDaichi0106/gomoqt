import { BytesPool, DefaultBytesPool } from "./bytes_pool";

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
        const value = this.#buf[this.#off];
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

    release(pool: BytesPool = DefaultBytesPool) {
        if (this.#buf.length > 0) { // Only release if there's a buffer to release
            pool.release(this.#buf.buffer);
        }
        this.reset();
        this.#buf = new Uint8Array(0);
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

