import { BytesBuffer, MAX_BYTES_LENGTH } from "../internal/bytes";
import { BytesPool, DefaultBytesPool } from "../internal/bytes_pool";
import { CancelCauseFunc, Context, withCancelCause } from "../internal/context";
import { StreamError, StreamErrorCode } from "./error";

let DefaultReadSize: number = 1024; // 1 KB

export class Reader {
    #byob?: ReadableStreamBYOBReader;
    #pull?: ReadableStreamDefaultReader<Uint8Array>;
    #buf: BytesBuffer;
    #pool: BytesPool;
    #closed: Promise<void>;

    constructor(readable: ReadableStream<Uint8Array>, pool: BytesPool = DefaultBytesPool) {
        this.#pool = pool;
        
        let closed: Promise<void>;
        try {
            this.#byob = readable.getReader({ mode: 'byob' });
            closed = this.#byob.closed;
        } catch {
            this.#pull = readable.getReader();
            closed = this.#pull.closed;
        }
        this.#buf = new BytesBuffer(pool.acquire(1024));

        this.#closed = new Promise((resolve) => {
            closed.then(() => {
                this.#buf.release();
                resolve();
            });
        });
        
    }

    async readUint8Array(): Promise<[Uint8Array?, Error?]> {
        let varint: bigint | undefined;
        let err: Error | undefined;
        [varint, err] = await this.readVarint();
        if (err) {
            return [undefined, err];
        }
        if (!varint) {
            return [undefined, new Error("Failed to read varint")];
        }

        const len = Number(varint);
        if (len > MAX_BYTES_LENGTH) {
            return [undefined, new Error("Varint too large")];
        }

        const bytes = new Uint8Array(len);

        let n: number | undefined;
        [n, err] = await this.copy(bytes);
        if (err) {
            return [undefined, err];
        }

        // Return only the bytes that were actually read
        if (n !== undefined && n < len) {
            return [bytes.subarray(0, n), undefined];
        }

        return [bytes, undefined];
    }

    async readString(): Promise<[string?, Error?]> {
        const [bytes, err] = await this.readUint8Array();
        if (err) {
            return [undefined, err];
        }
        if (!bytes) {
            return [undefined, err];
        }
        const str = new TextDecoder().decode(bytes);
        return [str, undefined];
    }

    async readVarint(): Promise<[bigint?, Error?]> {
        if (this.#buf.size == 0) {
            const [filled, err] = await this.fill(1);
            if (err) {
                return [undefined, err];
            }

            if (!filled) {
                return [undefined, new Error("Failed to read byte")];
            }

        }
        const firstByte = this.#buf.readUint8();


        const len = 1 << (firstByte >> 6);

        let value: bigint = BigInt(firstByte & 0x3f);

        if (this.#buf.size < len) {
            const [filled, err] = await this.fill(len-this.#buf.size);
            if (err) {
                return [undefined, err];
            }
            if (!filled) {
                return [undefined, new Error("Failed to read byte")];
            }
        }

        for (let i = 1; i < len; i++) {
            value = value << 8n | BigInt(this.#buf.readUint8());
        }

        return [value, undefined];
    }

    async readUint8(): Promise<[number?, Error?]> {
        let num: number;
        if (this.#buf.size == 0) {
            const [filled, err] = await this.fill(1);
            if (err) {
                return [undefined, err];
            }

            if (!filled) {
                return [undefined, new Error("Failed to read byte")];
            }
        }

        num = this.#buf.readUint8();
        return [num, undefined];
    }

    async readBoolean(): Promise<[boolean?, Error?]> {
        const [num, err] = await this.readUint8();
        if (err) {
            return [undefined, err];
        }
        if (!num) {
            return [undefined, new Error("Failed to read boolean")];
        }
        if (num < 0 || num > 1) {
            return [undefined, new Error("Invalid boolean value")];
        }
        return [num === 1, undefined];
    }

    async fill(diff: number): Promise<[number?, Error?]> {
        let totalFilled = 0;

        if (this.#byob) {
            const remaining = this.#buf.reserve(diff);
            while (totalFilled < diff) {
                const {done, value} = await this.#byob.read(remaining);
                if (done) {
                    return [undefined, new Error("Stream closed")];
                }
                if (value) {
                    totalFilled += value.length;
                } else {
                    break;
                }
            }
        } else if (this.#pull) {
            while (totalFilled < diff) {
                const {done, value} = await this.#pull.read();
                if (done) {
                    return [undefined, new Error("Stream closed")];
                }
                if (!value || value.length === 0) {
                    break; // No more data to read
                }

                this.#buf.write(value);
                totalFilled += value.length;
            }
        } else {
            return Promise.resolve([undefined, new Error("No reader available")]);
        }

        return [totalFilled, undefined];
    }

    async copy(buffer: Uint8Array): Promise<[number?, Error?]> {
        // Read existing data into the buffer
        let totalFilled = this.#buf.read(buffer);

        if (this.#byob) {
            while (totalFilled < buffer.length) {
                const remaining = buffer.subarray(totalFilled);
                const {done, value} = await this.#byob.read(remaining);
                if (done) {
                    return [undefined, new Error("Stream closed")];
                }
                if (value) {
                    totalFilled += value.length;
                } else {
                    break;
                }
            }
        } else if (this.#pull) {
            while (totalFilled < buffer.length) {
                const result = await this.#pull.read();
                if (result.done) {
                    return [undefined, new Error("Stream closed")];
                }
                if (!result.value || result.value.length === 0) {
                    break; // No more data to read
                }

                const needed = buffer.length - totalFilled;
                const len = Math.min(needed, result.value.length);

                buffer.set(result.value.subarray(0, len), totalFilled);
                totalFilled += len;

                if (result.value.length > len) {
                    const leftover = result.value.subarray(len);
                    this.#buf.write(leftover);
                }
            }
        } else {
            return Promise.resolve([undefined, new Error("No reader available")]);
        }

        return [totalFilled, undefined];
    }

    async cancel(code: StreamErrorCode, message: string): Promise<void> {
        const reason = new StreamError(code, message);
        if (this.#byob) {
            this.#byob.cancel(reason)
        } else if (this.#pull) {
            this.#pull.cancel(reason)
        }

        this.#buf.release();
    }

    closed(): Promise<void> {
        return this.#closed;
    }
}

