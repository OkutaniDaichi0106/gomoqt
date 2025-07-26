import { BytesBuffer, MAX_BYTES_LENGTH, MAX_UINT } from "../internal/bytes";
import { DefaultBytesPool } from "../internal/bytes_pool";
import { StreamError } from "./error";

export class Writer {
    #writer: WritableStreamDefaultWriter<Uint8Array>;
    #buf: BytesBuffer;

    constructor(stream: WritableStream<Uint8Array>, buf: ArrayBufferLike = DefaultBytesPool.acquire(1024)) {
        this.#writer = stream.getWriter();
        this.#buf = new BytesBuffer(buf);

        async () => {
            await this.#writer.closed;

            // TODO: Handle stream closure
            this.#buf.release(); // Release the buffer back to the pool
        };
    }

    writeUint8(value: number): void {
        if (value < 0 || value > 255) {
            throw new Error("Uint8 value must be between 0 and 255");
        }
        this.#buf.writeUint8(value);
    }

    writeUint8Array(data: Uint8Array): void {
        if (data.length > MAX_BYTES_LENGTH) {
            throw new Error("Bytes length exceeds maximum limit");
        }
        const len = BigInt(data.length);
        this.writeVarint(len);
        this.#buf.write(data);
    }

    writeString(str: string): void {
        const encoder = new TextEncoder();
        const data = encoder.encode(str);
        this.writeUint8Array(data);
    }

    writeVarint(num: bigint): void {
        if (num < 0) {
            throw new Error("Varint cannot be negative");
        }
        if (num > MAX_UINT) { // MAX_UINT for 62-bit unsigned integer
            throw new Error("Varint exceeds maximum value");
        }
        if (num < (1n << 6n)) {
            // 1 byte
            this.#buf.writeUint8(Number(num));
        } else if (num < (1n << 14n)) {
            // 2 bytes
            this.#buf.writeUint8(Number((num >> 8n) | 0x80n));
            this.#buf.writeUint8(Number(num & 0xFFn));
        } else if (num < (1n << 30n)) {
            // 4 bytes
            this.#buf.writeUint8(Number((num >> 24n) | 0xE0n));
            this.#buf.writeUint8(Number((num >> 16n) & 0xFFn));
            this.#buf.writeUint8(Number((num >> 8n) & 0xFFn));
            this.#buf.writeUint8(Number(num & 0xFFn));
        } else {
            // 8 bytes
            this.#buf.writeUint8(0xF0);
            this.#buf.writeUint8(Number((num >> 56n) & 0xFFn));
            this.#buf.writeUint8(Number((num >> 48n) & 0xFFn));
            this.#buf.writeUint8(Number((num >> 40n) & 0xFFn));
            this.#buf.writeUint8(Number((num >> 32n) & 0xFFn));
            this.#buf.writeUint8(Number((num >> 24n) & 0xFFn));
            this.#buf.writeUint8(Number((num >> 16n) & 0xFFn));
            this.#buf.writeUint8(Number((num >> 8n) & 0xFFn));
        }
    }

    writeBoolean(value: boolean): void {
        this.#buf.writeUint8(value ? 1 : 0);
    }

    writeStringArray(arr: string[]): void {
        this.writeVarint(BigInt(arr.length));
        for (const str of arr) {
            this.writeString(str);
        }
    }

    async flush(): Promise<Error | undefined> {
        console.log(`Flushing buffer of size: ${this.#buf.bytes()}`);
        if (this.#buf.size > 0) {
            try {
                await this.#writer.write(this.#buf.bytes());
            } catch (error) {
                return new Error("Failed to send data to stream");
            } finally {
                this.#buf.reset();
            }
        }

        return undefined;
    }

    async close(): Promise<void> {
        await this.#writer.close();
        this.#buf.release(); // Release the buffer back to the pool
    }

    async cancel(err: StreamError): Promise<void> {
        await this.#writer.abort(err);
        this.#buf.release(); // Release the buffer back to the pool
    }

    closed(): Promise<void> {
        return this.#writer.closed;
    }
}