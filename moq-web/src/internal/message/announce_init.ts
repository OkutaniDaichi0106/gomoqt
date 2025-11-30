import type { Reader, Writer } from "@okudai/golikejs/io";
import {
	parseStringArray,
	readFull,
	readVarint,
	stringLen,
	varintLen,
	writeStringArray,
	writeVarint,
} from "./message.ts";

export interface AnnounceInitMessageInit {
	suffixes?: string[];
}

export class AnnounceInitMessage {
	suffixes: string[];

	constructor(init: AnnounceInitMessageInit = {}) {
		this.suffixes = init.suffixes ?? [];
	}

	/**
	 * Returns the length of the message body (excluding the length prefix).
	 */
	get len(): number {
		let len = varintLen(this.suffixes.length);
		for (const suffix of this.suffixes) {
			len += stringLen(suffix);
		}
		return len;
	}

	/**
	 * Encodes the message to the writer.
	 */
	async encode(w: Writer): Promise<Error | undefined> {
		const msgLen = this.len;
		let err: Error | undefined;

		[, err] = await writeVarint(w, msgLen);
		if (err) return err;

		[, err] = await writeStringArray(w, this.suffixes);
		if (err) return err;

		return undefined;
	}

	/**
	 * Decodes the message from the reader.
	 */
	async decode(r: Reader): Promise<Error | undefined> {
		const [msgLen, , err1] = await readVarint(r);
		if (err1) return err1;

		const buf = new Uint8Array(msgLen);
		const [, err2] = await readFull(r, buf);
		if (err2) return err2;

		[this.suffixes] = parseStringArray(buf, 0);

		return undefined;
	}
}
