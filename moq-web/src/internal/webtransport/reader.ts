import { BytesBuffer, MAX_BYTES_LENGTH } from "./bytes.ts";
import { DefaultBufferPool } from "./buffer_pool.ts";
import type { StreamError } from "./error.ts";

const DefaultReadSize: number = 1024; // 1 KB

export const EOF = new Error("EOF");

// /**
//  * Reads a varint-encoded number from the byte array at the specified offset
//  * @param buf The byte array to read from
//  * @param offset The offset to start reading at
//  * @returns The decoded number and the number of bytes read
//  */
// export function readVarint(buf: Uint8Array, offset: number): { value: number, readLength: number } {
//     let result = 0;
//     let shift = 0;
//     let currentOffset = offset;

//     // Optimize: handle common cases
//     if (currentOffset >= buf.length) {
//         throw new Error("Buffer underflow while reading varint");
//     }

//     let byte = buf[currentOffset++];
//     result = byte & 0x7F;
//     if ((byte & 0x80) === 0) {
//         return { value: result, readLength: 1 };
//     }

//     // Continue reading
//     while (currentOffset < buf.length) {
//         byte = buf[currentOffset++];
//         result |= (byte & 0x7F) << shift;
//         if ((byte & 0x80) === 0) {
//             return { value: result, readLength: currentOffset - offset };
//         }
//         shift += 7;
//         if (shift >= 32) {
//             throw new Error("Varint too long for number");
//         }
//     }

//     throw new Error("Buffer underflow while reading varint");
// }

// /**
//  * Reads a varint-encoded bigint from the byte array at the specified offset
//  * @param buf The byte array to read from
//  * @param offset The offset to start reading at
//  * @returns The decoded bigint and the number of bytes read
//  */
// export function readBigVarint(buf: Uint8Array, offset: number): { value: bigint, readLength: number } {
//     let result = 0n;
//     let shift = 0;
//     let currentOffset = offset;

//     // Optimize: handle common cases
//     if (currentOffset >= buf.length) {
//         throw new Error("Buffer underflow while reading big varint");
//     }

//     let byte = buf[currentOffset++];
//     result = BigInt(byte & 0x7F);
//     if ((byte & 0x80) === 0) {
//         return { value: result, readLength: 1 };
//     }

//     // Continue reading
//     while (currentOffset < buf.length) {
//         byte = buf[currentOffset++];
//         result |= BigInt(byte & 0x7F) << BigInt(shift);
//         if ((byte & 0x80) === 0) {
//             return { value: result, readLength: currentOffset - offset };
//         }
//         shift += 7;
//         if (shift >= 64) {
//             throw new Error("Varint too long for bigint");
//         }
//     }

//     throw new Error("Buffer underflow while reading big varint");
// }

// /**
//  * Reads a string (UTF-8 encoded) from the byte array at the specified offset
//  * @param buf The byte array to read from
//  * @param offset The offset to start reading at
//  * @returns The decoded string and the number of bytes read
//  */
// export function readString(buf: Uint8Array, offset: number): { value: string, readLength: number } {
//     // First read the length as varint
//     const { value: length, readLength: varintLength } = readVarint(buf, offset);
//     if (offset + varintLength + length > buf.length) {
//         throw new Error("Buffer underflow while reading string");
//     }
//     // Then read the string bytes
//     const strBytes = buf.subarray(offset + varintLength, offset + varintLength + length);
//     const decoder = new TextDecoder();
//     const str = decoder.decode(strBytes);
//     return { value: str, readLength: varintLength + length };
// }

// /**
//  * Reads bytes from the byte array at the specified offset
//  * @param buf The byte array to read from
//  * @param offset The offset to start reading at
//  * @param length The number of bytes to read
//  * @returns The read bytes and the number of bytes read
//  */
// export function readUint8Array(buf: Uint8Array, offset: number): { value: Uint8Array, readLength: number } {
//     const { value: length, readLength } = readVarint(buf, offset);
//     offset += readLength;
//     if (offset + length > buf.length) {
//         throw new Error("Buffer underflow while reading bytes");
//     }
//     const bytes = buf.subarray(offset, offset + length);
//     return { value: bytes, readLength: length + readLength };
// }

export interface ReceiveStreamInit {
	stream: ReadableStream<Uint8Array>;
	transfer?: ArrayBufferLike;
	streamId: number;
}

export class ReceiveStream {
	// #byob?: ReadableStreamBYOBReader;
	#pull: ReadableStreamDefaultReader<Uint8Array>;
	#buf: BytesBuffer;
	#closed: Promise<void>;
	readonly id: number;

	constructor(init: ReceiveStreamInit) {
		this.#pull = init.stream.getReader();

		this.#buf = new BytesBuffer(init.transfer || DefaultBufferPool.acquire(DefaultReadSize));

		this.#closed = this.#pull.closed;
		this.id = init.streamId;
	}

	async readUint8Array(
		transfer?: ArrayBufferLike,
	): Promise<[Uint8Array, undefined] | [undefined, Error]> {
		let err: Error | undefined;
		let len: number;
		[len, err] = await this.readVarint();
		if (err) {
			return [undefined, err];
		}

		if (len > MAX_BYTES_LENGTH) {
			throw new Error("Varint too large");
		}

		let buffer: Uint8Array;
		if (transfer && transfer.byteLength >= len) {
			buffer = new Uint8Array(transfer, 0, len);
		} else {
			buffer = new Uint8Array(len);
		}

		err = await this.fillN(buffer, len);
		if (err) {
			return [undefined, err];
		}

		return [buffer, undefined];
	}

	async readString(): Promise<[string, Error | undefined]> {
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

	async readVarint(): Promise<[number, Error | undefined]> {
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

	async readBigVarint(): Promise<[bigint, Error | undefined]> {
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

	async readUint8(): Promise<[number, Error | undefined]> {
		if (this.#buf.size == 0) {
			const err = await this.pushN(1);
			if (err) {
				return [0, err];
			}
		}

		const num = this.#buf.readUint8();
		return [num, undefined];
	}

	async readBoolean(): Promise<[boolean, Error | undefined]> {
		const [num, err] = await this.readUint8();
		if (err) {
			return [false, err];
		}

		if (num < 0 || num > 1) {
			return [false, new Error("Invalid boolean value")];
		}
		return [num === 1, undefined];
	}

	async readStringArray(): Promise<[string[], Error | undefined]> {
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
		if (n <= 0) {
			return undefined;
		}

		let totalFilled = 0;

		while (totalFilled < n) {
			const { done, value } = await this.#pull.read();
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
			const { done, value } = await this.#pull.read();
			if (done) {
				return EOF;
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
		return this.#pull.cancel(reason);
	}

	closed(): Promise<void> {
		return this.#closed;
	}
}
