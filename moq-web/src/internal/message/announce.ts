import type { Reader, Writer } from "@okdaichi/golikejs/io";
import {
	parseString,
	parseVarint,
	readFull,
	readUint16,
	stringLen,
	varintLen,
	writeString,
	writeUint16,
	writeVarint,
} from "./message.ts";

export interface AnnounceMessageInit {
	suffix?: string;
	active?: boolean;
}

export class AnnounceMessage {
	suffix: string;
	active: boolean;

	constructor(init: AnnounceMessageInit = {}) {
		this.suffix = init.suffix ?? "";
		this.active = init.active ?? false;
	}

	/**
	 * Returns the length of the message body (excluding the length prefix).
	 */
	get len(): number {
		// AnnounceStatus is sent as a varint (not boolean)
		return varintLen(this.active ? 1 : 0) + stringLen(this.suffix);
	}

	/**
	 * Encodes the message to the writer.
	 */
	async encode(w: Writer): Promise<Error | undefined> {
		const msgLen = this.len;
		let err: Error | undefined;

		[, err] = await writeUint16(w, msgLen);
		if (err) return err;

		// Write AnnounceStatus as varint: 0x0 (ENDED) or 0x1 (ACTIVE)
		[, err] = await writeVarint(w, this.active ? 1 : 0);
		if (err) return err;

		[, err] = await writeString(w, this.suffix);
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

		let offset = 0;

		// Read AnnounceStatus as varint
		const [status, n1] = parseVarint(buf, offset);
		this.active = status === 1;
		offset += n1;

		[this.suffix] = parseString(buf, offset);

		return undefined;
	}
}
