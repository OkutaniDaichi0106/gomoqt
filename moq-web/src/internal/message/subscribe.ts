import type { ReceiveStream, SendStream } from "../webtransport/mod.ts";
import { stringLen, varintLen } from "../webtransport/mod.ts";

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

	constructor(init: SubscribeMessageInit) {
		this.subscribeId = init.subscribeId ?? 0;
		this.broadcastPath = init.broadcastPath ?? "";
		this.trackName = init.trackName ?? "";
		this.trackPriority = init.trackPriority ?? 0;
		this.minGroupSequence = init.minGroupSequence ?? 0;
		this.maxGroupSequence = init.maxGroupSequence ?? 0;
	}

	get messageLength(): number {
		return (
			varintLen(this.subscribeId) +
			stringLen(this.broadcastPath) +
			stringLen(this.trackName) +
			varintLen(this.trackPriority) +
			varintLen(this.minGroupSequence) +
			varintLen(this.maxGroupSequence)
		);
	}

	async encode(writer: SendStream): Promise<Error | undefined> {
		writer.writeVarint(this.messageLength);
		writer.writeVarint(this.subscribeId);
		writer.writeString(this.broadcastPath);
		writer.writeString(this.trackName);
		writer.writeVarint(this.trackPriority);
		writer.writeVarint(this.minGroupSequence);
		writer.writeVarint(this.maxGroupSequence);
		return await writer.flush();
	}

	async decode(reader: ReceiveStream): Promise<Error | undefined> {
		let [len, err] = await reader.readVarint();
		if (err) {
			return err;
		}
		[this.subscribeId, err] = await reader.readVarint();
		if (err) {
			return err;
		}
		[this.broadcastPath, err] = await reader.readString();
		if (err) {
			return err;
		}
		[this.trackName, err] = await reader.readString();
		if (err) {
			return err;
		}
		[this.trackPriority, err] = await reader.readVarint();
		if (err) {
			return err;
		}
		[this.minGroupSequence, err] = await reader.readVarint();
		if (err) {
			return err;
		}
		[this.maxGroupSequence, err] = await reader.readVarint();
		if (err) {
			return err;
		}

		if (len !== this.messageLength) {
			throw new Error(`message length mismatch: expected ${len}, got ${this.messageLength}`);
		}

		return undefined;
	}
}
