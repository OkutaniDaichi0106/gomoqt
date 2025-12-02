import type { Reader, Writer } from "@okudai/golikejs/io";
import {
	parseVarint,
	readFull,
	readUint16,
	varintLen,
	writeUint16,
	writeVarint,
} from "./message.ts";

export interface GroupMessageInit {
	subscribeId?: number;
	sequence?: number;
}

export class GroupMessage {
	subscribeId: number;
	sequence: number;

	constructor(init: GroupMessageInit = {}) {
		this.subscribeId = init.subscribeId ?? 0;
		this.sequence = init.sequence ?? 0;
	}

	/**
	 * Returns the length of the message body (excluding the length prefix).
	 */
	get len(): number {
		return varintLen(this.subscribeId) + varintLen(this.sequence);
	}

	/**
	 * Encodes the message to the writer.
	 */
	async encode(w: Writer): Promise<Error | undefined> {
		const msgLen = this.len;
		let err: Error | undefined;

		[, err] = await writeUint16(w, msgLen);
		if (err) return err;

		[, err] = await writeVarint(w, this.subscribeId);
		if (err) return err;

		[, err] = await writeVarint(w, this.sequence);
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

		[this.subscribeId, offset] = (() => {
			const [val, n] = parseVarint(buf, offset);
			return [val, offset + n];
		})();

		[this.sequence] = parseVarint(buf, offset);

		return undefined;
	}
}
