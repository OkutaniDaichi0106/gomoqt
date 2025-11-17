import type { ReceiveStream, SendStream } from "../webtransport/mod.ts";
import { stringLen, varintLen } from "../webtransport/mod.ts";

export interface AnnounceMessageInit {
	suffix?: string;
	active?: boolean;
}

export class AnnounceMessage {
	suffix: string;
	active: boolean;

	constructor(init: AnnounceMessageInit) {
		this.suffix = init.suffix ?? "";
		this.active = init.active ?? false;
	}

	get messageLength(): number {
		// AnnounceStatus is sent as a varint (not boolean)
		// varint for status + string length
		return varintLen(this.active ? 1 : 0) + stringLen(this.suffix);
	}

	async encode(writer: SendStream): Promise<Error | undefined> {
		writer.writeVarint(this.messageLength);
		// Write AnnounceStatus as varint: 0x0 (ENDED) or 0x1 (ACTIVE)
		writer.writeVarint(this.active ? 1 : 0);
		writer.writeString(this.suffix);
		return await writer.flush();
	}

	async decode(reader: ReceiveStream): Promise<Error | undefined> {
		let [len, err] = await reader.readVarint();
		if (err) {
			return err;
		}
		// Read AnnounceStatus as varint
		let status: number;
		[status, err] = await reader.readVarint();
		if (err) {
			return err;
		}
		this.active = status === 1;
		
		[this.suffix, err] = await reader.readString();
		if (err) {
			return err;
		}

		if (len !== this.messageLength) {
			throw new Error(`message length mismatch: expected ${len}, got ${this.messageLength}`);
		}

		return undefined;
	}
}
