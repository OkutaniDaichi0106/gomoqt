import type { ReceiveStream, SendStream } from "../webtransport/mod.ts";
import { stringLen } from "../webtransport/mod.ts";

export interface AnnouncePleaseMessageInit {
	prefix?: string;
}

export class AnnouncePleaseMessage {
	prefix: string;

	constructor(init: AnnouncePleaseMessageInit) {
		this.prefix = init.prefix ?? "";
	}

	get messageLength(): number {
		return stringLen(this.prefix);
	}

	async encode(writer: SendStream): Promise<Error | undefined> {
		writer.writeVarint(this.messageLength);
		writer.writeString(this.prefix);
		return await writer.flush();
	}

	async decode(reader: ReceiveStream): Promise<Error | undefined> {
		let [len, err] = await reader.readVarint();
		if (err) {
			return err;
		}

		[this.prefix, err] = await reader.readString();
		if (err) {
			return err;
		}

		if (len !== this.messageLength) {
			throw new Error(`message length mismatch: expected ${len}, got ${this.messageLength}`);
		}

		return undefined;
	}
}
