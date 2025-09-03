import { BytesBuffer, MAX_BYTES_LENGTH } from "./bytes";
import { BufferPool, DefaultBufferPool } from "./buffer_pool";
import { StreamError, StreamErrorCode } from "./error";

let DefaultReadSize: number = 1024; // 1 KB

export class Reader {
    // #byob?: ReadableStreamBYOBReader;
    #pull: ReadableStreamDefaultReader<Uint8Array>;
    #buf: BytesBuffer;
    #pool: BufferPool;
    #closed: Promise<void>;

    constructor(readable: ReadableStream<Uint8Array>, pool: BufferPool = DefaultBufferPool) {
        this.#pool = pool;
        this.#pull = readable.getReader();

        this.#buf = new BytesBuffer(pool.acquire(1024));

        this.#closed = this.#pull.closed;
    }

    async readUint8Array(): Promise<[Uint8Array?, Error?]> {
        let err: Error | undefined;
        let len: number;
        [len, err] = await this.readVarint();
        if (err) {
            return [undefined, err];
        }

        if (len > MAX_BYTES_LENGTH) {
            throw new Error("Varint too large");
        }

    // Acquire an underlying ArrayBuffer from the pool but create a view
    // of the exact requested length to avoid returning a larger buffer
    // (which would contain trailing zeros).
    const ab = this.#pool.acquire(len);
    const buffer = new Uint8Array(ab as ArrayBuffer, 0, len);

    err = await this.fillN(buffer, len);
        if (err) {
            return [undefined, err];
        }

        return [buffer, undefined];
    }

    async readString(): Promise<[string, Error?]> {
        let err: Error | undefined = undefined;
        const [bytes, err2] = await this.readUint8Array();
        err = err2;
        if (err) {
            return ["", err];
        }
        if (bytes === undefined) {
            return ["", err];
        }
        const str = new TextDecoder().decode(bytes);
        return [str, undefined];
    }

    async readVarint(): Promise<[number, Error?]> {
        let err: Error | undefined = undefined;
        let varint = 0;
        if (this.#buf.size == 0) {
            err = await this.pushN(1);
            if (err) {
                return [0, err];
            }
        }
        const firstByte = this.#buf.readUint8();
        const len = 1 << (firstByte >> 6);
        varint = firstByte & 0x3f;
        const remaining = len - 1; // Remaining bytes to read
        if (this.#buf.size < remaining) {
            err = await this.pushN(remaining - this.#buf.size);
            if (err) {
                return [0, err];
            }
        }
        for (let i = 0; i < remaining; i++) {
            // Use arithmetic multiplication to avoid 32-bit bitwise overflow in JS
            varint = varint * 256 + this.#buf.readUint8();
        }
        return [varint, undefined];
    }

    async readBigVarint(): Promise<[bigint, Error?]> {
        let err: Error | undefined = undefined;
        if (this.#buf.size == 0) {
            err = await this.pushN(1);
            if (err) {
                return [0n, err];
            }
        }
        const firstByte = this.#buf.readUint8();
        const len = 1 << (firstByte >> 6);
        let bigVarint = BigInt(firstByte & 0x3f);
        const remaining = len - 1; // Remaining bytes to read
        if (this.#buf.size < remaining) {
            let filled = 0;
            err = await this.pushN(remaining - this.#buf.size);
            if (err) {
                return [0n, err];
            }
        }
        for (let i = 0; i < remaining; i++) {
            // Use arithmetic multiplication for bigints to avoid bitwise operations
            bigVarint = bigVarint * 256n + BigInt(this.#buf.readUint8());
        }
        return [bigVarint, undefined];
    }

    async readUint8(): Promise<[number, Error?]> {
        let num: number;
        if (this.#buf.size == 0) {
            const err = await this.pushN(1);
            if (err) {
                return [0, err];
            }
        }

        num = this.#buf.readUint8();
        return [num, undefined];
    }

    async readBoolean(): Promise<[boolean, Error?]> {
        const [num, err] = await this.readUint8();
        if (err) {
            return [false, err];
        }

        if (num < 0 || num > 1) {
            return [false, new Error("Invalid boolean value")];
        }
        return [num === 1, undefined];
    }

    async readStringArray(): Promise<[string[], Error?]> {
        let err: Error | undefined = undefined;
        let bigVarint = 0n;
        [bigVarint, err] = await this.readBigVarint();
        if (err) {
            return [[], err];
        }

        const count = Number(bigVarint);
        if (count > MAX_BYTES_LENGTH) {
            err = new Error("Varint too large");
            return [[], err];
        }

        const strings: string[] = [];
        for (let i = 0; i < count; i++) {
            let str = "";
            [str, err] = await this.readString();
            if (err) {
                return [[], err];
            }
            strings.push(str);
        }

        return [strings, undefined];
    }

    async pushN(n: number): Promise<Error | undefined> {
        let totalFilled = 0;

        while (totalFilled < n) {
            const {done, value} = await this.#pull.read();
            if (done) {
                return new Error("Stream closed");
            }
            if (!value || value.length === 0) {
                continue; // Skip empty values
            }

            this.#buf.write(value);
            totalFilled += value.length;
       }

        return undefined;
    }

    async fillN(buffer: Uint8Array, n: number): Promise<Error | undefined> {

        // Read up to the requested number of bytes from the internal buffer
        let totalFilled = 0;
        if (this.#buf.size > 0) {
            const toRead = Math.min(this.#buf.size, n);
            totalFilled = this.#buf.read(buffer.subarray(0, toRead));
        }

        while (totalFilled < n) {
            const {done, value} = await this.#pull.read();
            if (done) {
                return new Error("Stream closed");
            }
            if (!value || value.length === 0) {
                // No data this iteration; wait for next pull
                continue;
            }

            const needed = n - totalFilled;
            const len = Math.min(needed, value.length);

            buffer.set(value.subarray(0, len), totalFilled);
            totalFilled += len;

            if (value.length > len) {
                // Store leftover bytes to internal buffer for subsequent reads
                const leftover = value.subarray(len);
                this.#buf.write(leftover);
            }
        }

        return undefined;
    }

    async cancel(reason: StreamError): Promise<void> {
        this.#pull.cancel(reason)
    }

    closed(): Promise<void> {
        return this.#closed;
    }
}
