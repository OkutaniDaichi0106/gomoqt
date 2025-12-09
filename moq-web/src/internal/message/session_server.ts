import type { Reader, Writer } from "@okdaichi/golikejs/io";
import {
	bytesLen,
	parseBytes,
	parseVarint,
	readFull,
	readUint16,
	varintLen,
	writeBytes,
	writeUint16,
	writeVarint,
} from "./message.ts";

export interface SessionServerMessageInit {
	version?: number;
	extensions?: Map<number, Uint8Array>;
}

export class SessionServerMessage {
	version: number;
	extensions: Map<number, Uint8Array>;

	constructor(init: SessionServerMessageInit = {}) {
		this.version = init.version ?? 0;
		this.extensions = init.extensions ?? new Map();
	}

	/**
	 * Returns the length of the message body (excluding the length prefix).
	 */
	get len(): number {
		let length = 0;
		length += varintLen(this.version);
		length += varintLen(this.extensions.size);
		for (const ext of this.extensions.entries()) {
			length += varintLen(ext[0]);
			length += bytesLen(ext[1]);
		}
		return length;
	}

	/**
	 * Encodes the message to the writer.
	 */
	async encode(w: Writer): Promise<Error | undefined> {
		const msgLen = this.len;
		let err: Error | undefined;

		[, err] = await writeUint16(w, msgLen);
		if (err) return err;

		[, err] = await writeVarint(w, this.version);
		if (err) return err;

		[, err] = await writeVarint(w, this.extensions.size);
		if (err) return err;

		for (const [extId, extData] of this.extensions.entries()) {
			[, err] = await writeVarint(w, extId);
			if (err) return err;
			[, err] = await writeBytes(w, extData);
			if (err) return err;
		}

		return undefined;
	}

	/**
	 * Decodes the message from the reader.
	 */
	async decode(r: Reader): Promise<Error | undefined> {
		const [msgLen, , err1] = await readUint16(r);
		if (err1) return err1;

		const buf = new Uint8Array(msgLen);
		const [, err2] = await readFull(r, buf);
		if (err2) return err2;

		let offset = 0;

		// Read version
		const [version, n1] = parseVarint(buf, offset);
		this.version = version;
		offset += n1;

		// Read extensions
		const [numExtensions, n2] = parseVarint(buf, offset);
		offset += n2;

		const extensions = new Map<number, Uint8Array>();
		for (let i = 0; i < numExtensions; i++) {
			const [extId, n3] = parseVarint(buf, offset);
			offset += n3;
			const [extData, n4] = parseBytes(buf, offset);
			offset += n4;
			extensions.set(extId, extData);
		}
		this.extensions = extensions;

		return undefined;
	}
}
