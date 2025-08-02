import { BytesBuffer, MAX_BYTES_LENGTH, MAX_UINT } from "../internal/bytes";
import { DefaultBytesPool } from "../internal/bytes_pool";
import { StreamError } from "./error";
import { MAX_VARINT1, MAX_VARINT2, MAX_VARINT4, MAX_VARINT8 } from "./len";

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
        this.writeVarint(BigInt(data.length));
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
        if (num <= MAX_VARINT1) {
            // 1 byte
            this.#buf.writeUint8(Number(num));
        } else if (num <= MAX_VARINT2) {
            // 2 bytes
            this.#buf.writeUint8(Number((num >> 8n) | 0x40n));
            this.#buf.writeUint8(Number(num & 0xFFn));
        } else if (num <= MAX_VARINT4) {
            // 4 bytes
            this.#buf.writeUint8(Number((num >> 24n) | 0x80n));
            this.#buf.writeUint8(Number((num >> 16n) & 0xFFn));
            this.#buf.writeUint8(Number((num >> 8n) & 0xFFn));
            this.#buf.writeUint8(Number(num & 0xFFn));
        } else if (num <= MAX_VARINT8) {
            // 8 bytes
            this.#buf.writeUint8(Number((num >> 56n) | 0xC0n));
            this.#buf.writeUint8(Number((num >> 48n) & 0xFFn));
            this.#buf.writeUint8(Number((num >> 40n) & 0xFFn));
            this.#buf.writeUint8(Number((num >> 32n) & 0xFFn));
            this.#buf.writeUint8(Number((num >> 24n) & 0xFFn));
            this.#buf.writeUint8(Number((num >> 16n) & 0xFFn));
            this.#buf.writeUint8(Number((num >> 8n) & 0xFFn));
            this.#buf.writeUint8(Number(num & 0xFFn));
        } else {
            throw new RangeError("Value exceeds maximum varint size");
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