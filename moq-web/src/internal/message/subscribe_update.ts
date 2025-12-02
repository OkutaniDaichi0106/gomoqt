import type { Reader, Writer } from "@okudai/golikejs/io";
import {
	parseVarint,
	readFull,
	readUint16,
	varintLen,
	writeUint16,
	writeVarint,
} from "./message.ts";

export interface SubscribeUpdateMessageInit {
	trackPriority?: number;
	minGroupSequence?: number;
	maxGroupSequence?: number;
}

export class SubscribeUpdateMessage {
	trackPriority: number;
	minGroupSequence: number;
	maxGroupSequence: number;

	constructor(init: SubscribeUpdateMessageInit = {}) {
		this.trackPriority = init.trackPriority ?? 0;
		this.minGroupSequence = init.minGroupSequence ?? 0;
		this.maxGroupSequence = init.maxGroupSequence ?? 0;
	}

	/**
	 * Returns the length of the message body (excluding the length prefix).
	 */
	get len(): number {
		return (
			varintLen(this.trackPriority) +
			varintLen(this.minGroupSequence) +
			varintLen(this.maxGroupSequence)
		);
	}

	/**
	 * Encodes the message to the writer.
	 */
	async encode(w: Writer): Promise<Error | undefined> {
		const msgLen = this.len;
		let err: Error | undefined;

		[, err] = await writeUint16(w, msgLen);
		if (err) return err;

		[, err] = await writeVarint(w, this.trackPriority);
		if (err) return err;

		[, err] = await writeVarint(w, this.minGroupSequence);
		if (err) return err;

		[, err] = await writeVarint(w, this.maxGroupSequence);
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

		[this.trackPriority, offset] = (() => {
			const [val, n] = parseVarint(buf, offset);
			return [val, offset + n];
		})();

		[this.minGroupSequence, offset] = (() => {
			const [val, n] = parseVarint(buf, offset);
			return [val, offset + n];
		})();

		[this.maxGroupSequence] = (() => {
			const [val, n] = parseVarint(buf, offset);
			return [val, offset + n];
		})();

		return undefined;
	}
}
