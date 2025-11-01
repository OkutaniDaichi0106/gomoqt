import type { Reader, SendStream } from "../webtransport/mod.ts";
import { varintLen } from "../webtransport/mod.ts";

export interface SubscribeUpdateMessageInit {
	trackPriority?: number;
	minGroupSequence?: bigint;
	maxGroupSequence?: bigint;
}

export class SubscribeUpdateMessage {
	trackPriority: number;
	minGroupSequence: bigint;
	maxGroupSequence: bigint;

	constructor(init: SubscribeUpdateMessageInit) {
		this.trackPriority = init.trackPriority ?? 0;
		this.minGroupSequence = init.minGroupSequence ?? 0n;
		this.maxGroupSequence = init.maxGroupSequence ?? 0n;
	}

	get messageLength(): number {
		return (
			varintLen(this.trackPriority) +
			varintLen(this.minGroupSequence) +
			varintLen(this.maxGroupSequence)
		);
	}

	async encode(writer: SendStream): Promise<Error | undefined> {
		writer.writeVarint(this.messageLength);
		writer.writeVarint(this.trackPriority);
		writer.writeBigVarint(this.minGroupSequence);
		writer.writeBigVarint(this.maxGroupSequence);
		return await writer.flush();
	}

	async decode(reader: Reader): Promise<Error | undefined> {
		let [len, err] = await reader.readVarint();
		if (err) {
			return err;
		}
		[this.trackPriority, err] = await reader.readVarint();
		if (err) {
			return err;
		}
		[this.minGroupSequence, err] = await reader.readBigVarint();
		if (err) {
			return err;
		}
		[this.maxGroupSequence, err] = await reader.readBigVarint();
		if (err) {
			return err;
		}

		if (len !== this.messageLength) {
			throw new Error(`message length mismatch: expected ${len}, got ${this.messageLength}`);
		}

		return undefined;
	}
}
