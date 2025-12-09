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

export interface SubscribeMessageInit {
	subscribeId?: number;
	broadcastPath?: string;
	trackName?: string;
	trackPriority?: number;
}

export class SubscribeMessage {
	subscribeId: number;
	broadcastPath: string;
	trackName: string;
	trackPriority: number;

	constructor(init: SubscribeMessageInit = {}) {
		this.subscribeId = init.subscribeId ?? 0;
		this.broadcastPath = init.broadcastPath ?? "";
		this.trackName = init.trackName ?? "";
		this.trackPriority = init.trackPriority ?? 0;
	}

	/**
	 * Returns the length of the message body (excluding the length prefix).
	 */
	get len(): number {
		return (
			varintLen(this.subscribeId) +
			stringLen(this.broadcastPath) +
			stringLen(this.trackName) +
			varintLen(this.trackPriority)
		);
	}

	/**
	 * Encodes the message to the writer.
	 * Go-style: encode(w io.Writer) error
	 */
	async encode(w: Writer): Promise<Error | undefined> {
		const msgLen = this.len;
		let err: Error | undefined;

		[, err] = await writeUint16(w, msgLen);
		if (err) return err;

		[, err] = await writeVarint(w, this.subscribeId);
		if (err) return err;

		[, err] = await writeString(w, this.broadcastPath);
		if (err) return err;

		[, err] = await writeString(w, this.trackName);
		if (err) return err;

		[, err] = await writeVarint(w, this.trackPriority);
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
		[msgLen, , err] = await readUint16(r);
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

		return undefined;
	}
}
