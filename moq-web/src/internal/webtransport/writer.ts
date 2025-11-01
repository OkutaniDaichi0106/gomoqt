import { BytesBuffer, MAX_BYTES_LENGTH } from "./bytes.ts";
import { DefaultBufferPool } from "./buffer_pool.ts";
import type { StreamError } from "./error.ts";
import { MAX_VARINT1, MAX_VARINT2, MAX_VARINT4, MAX_VARINT8 } from "./len.ts";

// /**
//  * Grows the buffer if necessary to accommodate the required size
//  * @param buf The current buffer
//  * @param requiredSize The minimum size needed
//  * @returns The buffer, possibly reallocated
//  */
// function growBuffer(buf: Uint8Array, requiredSize: number): Uint8Array {
//     if (buf.length >= requiredSize) {
//         return buf;
//     }
//     // Go's slice growth algorithm: double the capacity or use required size, whichever is larger
//     const newCapacity = Math.max(buf.length * 2, requiredSize);
//     const newBuf = new Uint8Array(newCapacity);
//     newBuf.set(buf);
//     return newBuf;
// }

// /**
//  * Ensures the buffer has enough capacity for additional bytes
//  * @param buf The current buffer
//  * @param offset The current offset
//  * @param additionalBytes The number of additional bytes needed
//  * @returns The buffer, possibly reallocated
//  */
// function ensureCapacity(buf: Uint8Array, offset: number, additionalBytes: number): Uint8Array {
//     const requiredSize = offset + additionalBytes;
//     if (buf.length >= requiredSize) {
//         return buf;
//     }
//     return growBuffer(buf, requiredSize);
// }

// /**
//  * Writes a varint-encoded number to the byte array at the specified offset
//  * @param buf The byte array to write to
//  * @param offset The offset to start writing at
//  * @param value The number to encode
//  * @returns The modified byte array and the number of bytes written
//  */
// export function writeVarint(buf: Uint8Array, offset: number, value: number): { buf: Uint8Array, wroteLength: number } {
//     // Ensure we have enough space for maximum varint size (5 bytes for 32-bit)
//     buf = ensureCapacity(buf, offset, 5);

//     let written = 0;
//     let v = value;

//     // Optimize: use number instead of BigInt for small values
//     if (v < 0x80) {
//         buf[offset] = v;
//         return { buf, wroteLength: 1 };
//     }

//     while (v >= 0x80) {
//         buf[offset + written] = (v & 0x7F) | 0x80;
//         v >>= 7;
//         written++;
//     }
//     buf[offset + written] = v;
//     written++;

//     return { buf, wroteLength: written };
// }

// /**
//  * Writes a varint-encoded bigint to the byte array at the specified offset
//  * @param buf The byte array to write to
//  * @param offset The offset to start writing at
//  * @param value The bigint to encode
//  * @returns The modified byte array and the number of bytes written
//  */
// export function writeBigVarint(buf: Uint8Array, offset: number, value: bigint): { buf: Uint8Array, wroteLength: number } {
//     // Ensure we have enough space for maximum varint size (10 bytes for 64-bit)
//     buf = ensureCapacity(buf, offset, 10);

//     let written = 0;
//     let v = value;

//     // Optimize: handle small values quickly
//     if (v < 0x80n) {
//         buf[offset] = Number(v);
//         return { buf, wroteLength: 1 };
//     }

//     while (v >= 0x80n) {
//         buf[offset + written] = Number((v & 0x7Fn) | 0x80n);
//         v >>= 7n;
//         written++;
//     }
//     buf[offset + written] = Number(v);
//     written++;

//     return { buf, wroteLength: written };
// }

// /**
//  * Writes a string (UTF-8 encoded) to the byte array at the specified offset
//  * @param buf The byte array to write to
//  * @param offset The offset to start writing at
//  * @param str The string to encode
//  * @returns The modified byte array and the number of bytes written
//  */
// export function writeString(buf: Uint8Array, offset: number, str: string): { buf: Uint8Array, wroteLength: number } {
//     const encoder = new TextEncoder();
//     const strBytes = encoder.encode(str);

//     // Ensure capacity for varint (max 5 bytes) + string bytes
//     buf = ensureCapacity(buf, offset, 5 + strBytes.length);

//     // Write the length as varint
//     const { buf: buf1, wroteLength: lengthWritten } = writeVarint(buf, offset, strBytes.length);
//     buf = buf1;

//     // Write the string bytes efficiently
//     buf.set(strBytes, offset + lengthWritten);

//     return { buf, wroteLength: lengthWritten + strBytes.length };
// }

// /**
//  * Writes bytes to the byte array at the specified offset
//  * @param buf The byte array to write to
//  * @param offset The offset to start writing at
//  * @param bytes The bytes to append
//  * @returns The modified byte array and the number of bytes written
//  */
// export function writeBytes(buf: Uint8Array, offset: number, bytes: Uint8Array): { buf: Uint8Array, wroteLength: number } {
//     buf = ensureCapacity(buf, offset, bytes.length);
//     buf.set(bytes, offset);
//     return { buf, wroteLength: bytes.length };
// }

export interface SendStreamInit {
	stream: WritableStream<Uint8Array>;
	transfer?: ArrayBufferLike;
	streamId: number;
}

export class SendStream {
	#writer: WritableStreamDefaultWriter<Uint8Array>;
	#buf: BytesBuffer;
	readonly id: number;

	constructor(init: SendStreamInit) {
		this.#writer = init.stream.getWriter();
		this.#buf = new BytesBuffer(init.transfer || DefaultBufferPool.acquire(1024));

		this.id = init.streamId;
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
		this.writeVarint(data.length);
		this.#buf.write(data);
	}

	writeString(str: string): void {
		const encoder = new TextEncoder();
		const data = encoder.encode(str);
		this.writeUint8Array(data);
	}

	writeVarint(num: number): void {
		if (num < 0) {
			throw new Error("Varint cannot be negative");
		}
		if (num <= MAX_VARINT1) {
			// 1 byte
			this.#buf.write(Uint8Array.of(num));
		} else if (num <= MAX_VARINT2) {
			// 2 bytes
			const out = new Uint8Array(2);
			out[0] = (num >> 8) | 0x40;
			out[1] = num & 0xFF;
			this.#buf.write(out);
		} else if (num <= MAX_VARINT4) {
			// 4 bytes
			const out = new Uint8Array(4);
			out[0] = (num >> 24) | 0x80;
			out[1] = (num >> 16) & 0xFF;
			out[2] = (num >> 8) & 0xFF;
			out[3] = num & 0xFF;
			this.#buf.write(out);
		} else if (num <= MAX_VARINT8) {
			// 8 bytes
			const out = new Uint8Array(8);
			out[0] = (num >> 56) | 0xC0;
			out[1] = (num >> 48) & 0xFF;
			out[2] = (num >> 40) & 0xFF;
			out[3] = (num >> 32) & 0xFF;
			out[4] = (num >> 24) & 0xFF;
			out[5] = (num >> 16) & 0xFF;
			out[6] = (num >> 8) & 0xFF;
			out[7] = num & 0xFF;
			this.#buf.write(out);
		} else {
			throw new RangeError("Value exceeds maximum varint size");
		}
	}

	writeBigVarint(num: bigint): void {
		if (num < 0n) {
			throw new Error("Varint cannot be negative");
		}
		if (num <= MAX_VARINT1) {
			// 1 byte
			this.#buf.write(Uint8Array.of(Number(num)));
		} else if (num <= MAX_VARINT2) {
			// 2 bytes
			const out = new Uint8Array(2);
			out[0] = Number((num >> 8n) | 0x40n);
			out[1] = Number(num & 0xFFn);
			this.#buf.write(out);
		} else if (num <= MAX_VARINT4) {
			// 4 bytes
			const out = new Uint8Array(4);
			out[0] = Number((num >> 24n) | 0x80n);
			out[1] = Number((num >> 16n) & 0xFFn);
			out[2] = Number((num >> 8n) & 0xFFn);
			out[3] = Number(num & 0xFFn);
			this.#buf.write(out);
		} else if (num <= MAX_VARINT8) {
			// 8 bytes
			const out = new Uint8Array(8);
			out[0] = Number((num >> 56n) | 0xC0n);
			out[1] = Number((num >> 48n) & 0xFFn);
			out[2] = Number((num >> 40n) & 0xFFn);
			out[3] = Number((num >> 32n) & 0xFFn);
			out[4] = Number((num >> 24n) & 0xFFn);
			out[5] = Number((num >> 16n) & 0xFFn);
			out[6] = Number((num >> 8n) & 0xFFn);
			out[7] = Number(num & 0xFFn);
			this.#buf.write(out);
		} else {
			throw new RangeError("Value exceeds maximum varint size");
		}
	}

	writeBoolean(value: boolean): void {
		this.#buf.writeUint8(value ? 1 : 0);
	}

	writeStringArray(arr: string[]): void {
		this.writeVarint(arr.length);
		for (const str of arr) {
			this.writeString(str);
		}
	}

	copyFrom(src: Source): void {
		this.writeVarint(src.byteLength);
		src.copyTo(this.#buf.reserve(src.byteLength));
	}

	async flush(): Promise<Error | undefined> {
		if (this.#buf.size > 0) {
			try {
				await this.#writer.write(this.#buf.bytes());
			} catch (error) {
				return new Error(`Failed to write to stream: ${error}`);
			} finally {
				this.#buf.reset();
			}
		}

		return undefined;
	}

	async close(): Promise<void> {
		await this.#writer.close();
	}

	async cancel(err: StreamError): Promise<void> {
		await this.#writer.abort(err);
	}

	closed(): Promise<void> {
		return this.#writer.closed;
	}
}

export interface Source {
	byteLength: number;
	copyTo(target: ArrayBuffer | ArrayBufferView<ArrayBufferLike>): void;
}
