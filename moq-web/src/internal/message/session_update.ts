import type { Reader, Writer } from "../webtransport/mod.ts";
import { varintLen } from "../webtransport/mod.ts";

export interface SessionUpdateMessageInit {
	bitrate?: number;
}

export class SessionUpdateMessage {
	bitrate: number;

	constructor(init: SessionUpdateMessageInit) {
		this.bitrate = init.bitrate ?? 0;
	}

	get messageLength(): number {
		return varintLen(this.bitrate);
	}

	async encode(writer: Writer): Promise<Error | undefined> {
		writer.writeVarint(this.messageLength);
		writer.writeVarint(this.bitrate);
		return await writer.flush();
	}

	async decode(reader: Reader): Promise<Error | undefined> {
		let [len, err] = await reader.readVarint();
		if (err) {
			return err;
		}
		[this.bitrate, err] = await reader.readVarint();
		if (err) {
			return err;
		}

		if (len !== this.messageLength) {
			throw new Error(`message length mismatch: expected ${len}, got ${this.messageLength}`);
		}

		return undefined;
	}
}
