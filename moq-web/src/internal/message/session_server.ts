import type { ReceiveStream, SendStream } from "../webtransport/mod.ts";
import { bytesLen, varintLen } from "../webtransport/mod.ts";

export interface SessionServerMessageInit {
	version?: number;
	extensions?: Map<number, Uint8Array>;
}

export class SessionServerMessage {
	version: number;
	extensions: Map<number, Uint8Array>;

	constructor(init: SessionServerMessageInit) {
		this.version = init.version ?? 0;
		this.extensions = init.extensions ?? new Map();
	}

	get messageLength(): number {
		let length = 0;
		length += varintLen(this.version);
		length += varintLen(this.extensions.size);
		for (const ext of this.extensions.entries()) {
			length += varintLen(ext[0]);
			length += bytesLen(ext[1]);
		}
		return length;
	}

	async encode(writer: SendStream): Promise<Error | undefined> {
		writer.writeVarint(this.messageLength);
		writer.writeVarint(this.version);
		writer.writeVarint(this.extensions.size); // Write the number of extensions
		for (const ext of this.extensions.entries()) {
			writer.writeVarint(ext[0]); // Write the extension ID
			writer.writeUint8Array(ext[1]); // Write the extension data
		}
		return await writer.flush();
	}

	async decode(reader: ReceiveStream): Promise<Error | undefined> {
		let [len, err] = await reader.readVarint();
		if (err) {
			return err;
		}

		[this.version, err] = await reader.readVarint();
		if (err) {
			return err;
		}

		let extensionCount: number;
		[extensionCount, err] = await reader.readVarint();
		if (err) {
			return err;
		}

		const extensions = new Map<number, Uint8Array>();

		let extId: number;
		for (let i = 0; i < extensionCount; i++) {
			[extId, err] = await reader.readVarint();
			if (err) {
				return err;
			}
			let extData: Uint8Array | undefined;
			[extData, err] = await reader.readUint8Array();
			if (err) {
				return err;
			}
			if (extData === undefined) {
				throw new Error("read extData: Uint8Array is undefined");
			}
			extensions.set(extId, extData);
		}

		this.extensions = extensions;

		if (len !== this.messageLength) {
			throw new Error(`message length mismatch: expected ${len}, got ${this.messageLength}`);
		}

		return undefined;
	}
}
