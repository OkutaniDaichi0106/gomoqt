import type { Reader, Writer } from "@okudai/golikejs/io";
import {
	parseVarint,
	readFull,
	readUint16,
	varintLen,
	writeUint16,
	writeVarint,
} from "./message.ts";

export interface SessionUpdateMessageInit {
	bitrate?: number;
}

export class SessionUpdateMessage {
	bitrate: number;

	constructor(init: SessionUpdateMessageInit = {}) {
		this.bitrate = init.bitrate ?? 0;
	}

	/**
	 * Returns the length of the message body (excluding the length prefix).
	 */
	get len(): number {
		return varintLen(this.bitrate);
	}

	/**
	 * Encodes the message to the writer.
	 */
	async encode(w: Writer): Promise<Error | undefined> {
		const msgLen = this.len;
		let err: Error | undefined;

		[, err] = await writeUint16(w, msgLen);
		if (err) return err;

		[, err] = await writeVarint(w, this.bitrate);
		if (err) return err;

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

		[this.bitrate] = parseVarint(buf, 0);

		return undefined;
	}
}
