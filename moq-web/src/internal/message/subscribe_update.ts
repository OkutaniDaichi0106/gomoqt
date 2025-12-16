import type { Reader, Writer } from "@okdaichi/golikejs/io";
import { parseVarint, readFull, readVarint, varintLen, writeVarint } from "./message.ts";

export interface SubscribeUpdateMessageInit {
	trackPriority?: number;
}

export class SubscribeUpdateMessage {
	trackPriority: number;

	constructor(init: SubscribeUpdateMessageInit = {}) {
		this.trackPriority = init.trackPriority ?? 0;
	}

	/**
	 * Returns the length of the message body (excluding the length prefix).
	 */
	get len(): number {
		return (
			varintLen(this.trackPriority)
		);
	}

	/**
	 * Encodes the message to the writer.
	 */
	async encode(w: Writer): Promise<Error | undefined> {
		const msgLen = this.len;
		let err: Error | undefined;

		[, err] = await writeVarint(w, msgLen);
		if (err) return err;

		[, err] = await writeVarint(w, this.trackPriority);
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

		let offset = 0;

		[this.trackPriority, offset] = (() => {
			const [val, n] = parseVarint(buf, offset);
			return [val, offset + n];
		})();

		return undefined;
	}
}
