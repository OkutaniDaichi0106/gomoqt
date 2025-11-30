import type { Reader, Writer } from "@okudai/golikejs/io";
import {
	parseString,
	parseVarint,
	readFull,
	readVarint,
	stringLen,
	varintLen,
	writeString,
	writeVarint,
} from "./message.ts";

export interface SubscribeMessageInit {
	subscribeId?: number;
	broadcastPath?: string;
	trackName?: string;
	trackPriority?: number;
	minGroupSequence?: number;
	maxGroupSequence?: number;
}

export class SubscribeMessage {
	subscribeId: number;
	broadcastPath: string;
	trackName: string;
	trackPriority: number;
	minGroupSequence: number;
	maxGroupSequence: number;

	constructor(init: SubscribeMessageInit = {}) {
		this.subscribeId = init.subscribeId ?? 0;
		this.broadcastPath = init.broadcastPath ?? "";
		this.trackName = init.trackName ?? "";
		this.trackPriority = init.trackPriority ?? 0;
		this.minGroupSequence = init.minGroupSequence ?? 0;
		this.maxGroupSequence = init.maxGroupSequence ?? 0;
	}

	/**
	 * Returns the length of the message body (excluding the length prefix).
	 */
	get len(): number {
		return (
			varintLen(this.subscribeId) +
			stringLen(this.broadcastPath) +
			stringLen(this.trackName) +
			varintLen(this.trackPriority) +
			varintLen(this.minGroupSequence) +
			varintLen(this.maxGroupSequence)
		);
	}

	/**
	 * Encodes the message to the writer.
	 * Go-style: encode(w io.Writer) error
	 */
	async encode(w: Writer): Promise<Error | undefined> {
		const msgLen = this.len;
		let err: Error | undefined;

		[, err] = await writeVarint(w, msgLen);
		if (err) return err;

		[, err] = await writeVarint(w, this.subscribeId);
		if (err) return err;

		[, err] = await writeString(w, this.broadcastPath);
		if (err) return err;

		[, err] = await writeString(w, this.trackName);
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
	 * Go-style: decode(r io.Reader) error
	 */
	async decode(r: Reader): Promise<Error | undefined> {
		let err: Error | undefined;

		// Read message length
		let msgLen: number;
		[msgLen, , err] = await readVarint(r);
		if (err) return err;

		// Read message body into a buffer
		const buf = new Uint8Array(msgLen);
		[, err] = await readFull(r, buf);
		if (err) return err;

		// Parse fields from the buffer
		let offset = 0;

		// subscribeId
		const [subscribeId, n1] = parseVarint(buf, offset);
		this.subscribeId = subscribeId;
		offset += n1;

		// broadcastPath
		const [broadcastPath, n2] = parseString(buf, offset);
		this.broadcastPath = broadcastPath;
		offset += n2;

		// trackName
		const [trackName, n3] = parseString(buf, offset);
		this.trackName = trackName;
		offset += n3;

		// trackPriority
		const [trackPriority, n4] = parseVarint(buf, offset);
		this.trackPriority = trackPriority;
		offset += n4;

		// minGroupSequence
		const [minGroupSequence, n5] = parseVarint(buf, offset);
		this.minGroupSequence = minGroupSequence;
		offset += n5;

		// maxGroupSequence
		const [maxGroupSequence, _n6] = parseVarint(buf, offset);
		this.maxGroupSequence = maxGroupSequence;

		return undefined;
	}
}
