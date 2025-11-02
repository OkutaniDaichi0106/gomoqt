import type { ReceiveStream, SendStream } from "../webtransport/mod.ts";

// deno-lint-ignore no-empty-interface
export interface SubscribeOkMessageInit {}

export class SubscribeOkMessage {
	constructor(_: SubscribeOkMessageInit) {
	}

	get messageLength(): number {
		return 0;
	}

	async encode(writer: SendStream): Promise<Error | undefined> {
		writer.writeVarint(this.messageLength);
		return await writer.flush();
	}

	async decode(reader: ReceiveStream): Promise<Error | undefined> {
		const [len, err] = await reader.readVarint();
		if (err) {
			return err;
		}

		if (len !== this.messageLength) {
			throw new Error(`message length mismatch: expected ${len}, got ${this.messageLength}`);
		}

		return undefined;
	}
}
