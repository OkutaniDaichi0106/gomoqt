import { BytesBuffer, MAX_BYTES_LENGTH } from "../internal/bytes";
import { BytesPool, DefaultBytesPool } from "../internal/bytes_pool";
import { StreamError, StreamErrorCode } from "./error";

let DefaultReadSize: number = 1024; // 1 KB

export class Reader {
    // #byob?: ReadableStreamBYOBReader;
    #pull: ReadableStreamDefaultReader<Uint8Array>;
    #buf: BytesBuffer;
    #pool: BytesPool;
    #closed: Promise<void>;

    constructor(readable: ReadableStream<Uint8Array>, pool: BytesPool = DefaultBytesPool) {
        this.#pool = pool;
        this.#pull = readable.getReader();

        this.#buf = new BytesBuffer(pool.acquire(1024));

        this.#closed = this.#pull.closed.then(() => {
            this.#buf.release();
            return;
        })
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

        const buffer = new Uint8Array(this.#pool.acquire(len));
        const bytes = buffer.subarray(0, len); // Only use the exact length needed

        let n = 0;
        [n, err] = await this.copy(bytes);
        if (err) {
            return [undefined, err];
        }

        // Return only the bytes that were actually read
        if (n < len) {
            return [bytes.subarray(0, n), undefined];
        }

        return [bytes, undefined];
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
            let filled = 0;
            [filled, err] = await this.fill(1);
            if (err) {
                return [0, err];
            }
            if (!filled) {
                err = new Error("Failed to read byte");
                return [0, err];
            }
        }
        const firstByte = this.#buf.readUint8();
        const len = 1 << (firstByte >> 6);
        varint = firstByte & 0x3f;
        const remaining = len - 1; // Remaining bytes to read
        if (this.#buf.size < remaining) {
            let filled = 0;
            [filled, err] = await this.fill(remaining - this.#buf.size);
            if (err) {
                return [0, err];
            }
            if (!filled) {
                err = new Error("Failed to read byte");
                return [0, err];
            }
        }
        for (let i = 0; i < remaining; i++) {
            varint = (varint << 8) | this.#buf.readUint8();
        }
        return [varint, undefined];
    }

    async readBigVarint(): Promise<[bigint, Error?]> {
        let err: Error | undefined = undefined;
        if (this.#buf.size == 0) {
            let filled = 0;
            [filled, err] = await this.fill(1);
            if (err) {
                return [0n, err];
            }
            if (!filled) {
                err = new Error("Failed to read byte");
                return [0n, err];
            }
        }
        const firstByte = this.#buf.readUint8();
        const len = 1 << (firstByte >> 6);
        let bigVarint = BigInt(firstByte & 0x3f);
        const remaining = len - 1; // Remaining bytes to read
        if (this.#buf.size < remaining) {
            let filled = 0;
            [filled, err] = await this.fill(remaining - this.#buf.size);
            if (err) {
                return [0n, err];
            }
            if (!filled) {
                err = new Error("Failed to read byte");
                return [0n, err];
            }
        }
        for (let i = 0; i < remaining; i++) {
            bigVarint = (bigVarint << 8n) | BigInt(this.#buf.readUint8());
        }
        return [bigVarint, undefined];
    }

    async readUint8(): Promise<[number, Error?]> {
        let num: number;
        if (this.#buf.size == 0) {
            const [filled, err] = await this.fill(1);
            if (err) {
                return [0, err];
            }

            if (!filled) {
                return [0, new Error("Failed to read byte")];
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

    async fill(diff: number): Promise<[number, Error?]> {
        let totalFilled = 0;

        while (totalFilled < diff) {
            const {done, value} = await this.#pull.read();
            if (done) {
                return [totalFilled, totalFilled > 0 ? undefined : new Error("Stream closed")];
            }
            if (!value || value.length === 0) {
                break; // No more data to read
            }

            this.#buf.write(value);
            totalFilled += value.length;
       }

        return [totalFilled, undefined];
    }

    async copy(buffer: Uint8Array): Promise<[number, Error?]> {
        // Read existing data into the buffer
        let totalFilled = this.#buf.read(buffer);

        while (totalFilled < buffer.length) {
            const {done, value} = await this.#pull.read();
            if (done) {
                return [totalFilled, totalFilled > 0 ? undefined : new Error("Stream closed")];
            }
            if (!value || value.length === 0) {
                break; // No more data to read
            }

            const needed = buffer.length - totalFilled;
            const len = Math.min(needed, value.length);

            buffer.set(value.subarray(0, len), totalFilled);
            totalFilled += len;

            if (value.length > len) {
                const leftover = value.subarray(len);
                this.#buf.write(leftover);
            }
        }

        return [totalFilled, undefined];
    }

    async cancel(reason: StreamError): Promise<void> {
        this.#pull.cancel(reason)
        this.#buf.release();
    }

    closed(): Promise<void> {
        return this.#closed;
    }
}


