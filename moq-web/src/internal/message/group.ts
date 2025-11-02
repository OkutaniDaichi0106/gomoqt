import type { ReceiveStream, SendStream } from "../webtransport/mod.ts";
import { varintLen } from "../webtransport/mod.ts";

export interface GroupMessageInit {
	subscribeId?: number;
	sequence?: bigint;
}

export class GroupMessage {
	subscribeId: number;
	sequence: bigint;

	constructor(init: GroupMessageInit) {
		this.subscribeId = init.subscribeId ?? 0;
		this.sequence = init.sequence ?? 0n;
	}

	get messageLength(): number {
		return varintLen(this.subscribeId) + varintLen(this.sequence);
	}

	async encode(writer: SendStream): Promise<Error | undefined> {
		writer.writeVarint(this.messageLength);
		writer.writeVarint(this.subscribeId);
		writer.writeBigVarint(this.sequence);
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

		[this.sequence, err] = await reader.readBigVarint();
		if (err) {
			return err;
		}

		if (len !== this.messageLength) {
			throw new Error(`message length mismatch: expected ${len}, got ${this.messageLength}`);
		}

		return undefined;
	}
}
