import type { Reader, Writer } from "../webtransport/mod.ts";
import { bytesLen, varintLen } from "../webtransport/mod.ts";

export interface SessionClientInit {
	versions?: Set<number>;
	extensions?: Map<number, Uint8Array>;
}

export class SessionClientMessage {
	versions: Set<number>;
	extensions: Map<number, Uint8Array>;

	constructor(init: SessionClientInit) {
		this.versions = init.versions ?? new Set();
		this.extensions = init.extensions ?? new Map();
	}

	get messageLength(): number {
		let length = 0;
		length += varintLen(this.versions.size);
		for (const version of this.versions) {
			length += varintLen(version);
		}
		length += varintLen(this.extensions.size);
		for (const ext of this.extensions.entries()) {
			length += varintLen(ext[0]); // Extension ID length
			length += bytesLen(ext[1]); // Extension data length (includes length prefix)
		}
		return length;
	}

	async encode(writer: Writer): Promise<Error | undefined> {
		writer.writeVarint(this.messageLength);
		writer.writeVarint(this.versions.size);
		for (const version of this.versions) {
			writer.writeVarint(version);
		}
		writer.writeVarint(this.extensions.size);
		for (const ext of this.extensions.entries()) {
			writer.writeVarint(ext[0]); // Write the extension ID
			writer.writeUint8Array(ext[1]); // Write the extension data
		}
		return await writer.flush();
	}

	async decode(reader: Reader): Promise<Error | undefined> {
		let [len, err] = await reader.readVarint();
		if (err) {
			return err;
		}
		let numVersions: number;
		[numVersions, err] = await reader.readVarint();
		if (err) {
			return err;
		}
		if (numVersions < 0) {
			throw new Error("Invalid number of versions for SessionClient");
		}
		if (numVersions > Number.MAX_SAFE_INTEGER) {
			throw new Error("Number of versions exceeds maximum safe integer for SessionClient");
		}
		const versions = new Set<number>();
		for (let i = 0; i < numVersions; i++) {
			let version: number;
			[version, err] = await reader.readVarint();
			if (err) {
				return err;
			}
			versions.add(version);
		}
		let numExtensions: number;
		[numExtensions, err] = await reader.readVarint();
		if (err) {
			return err;
		}
		if (numExtensions === undefined) {
			throw new Error("read numExtensions: number is undefined");
		}
		if (numExtensions < 0) {
			throw new Error("Invalid number of extensions for SessionClient");
		}
		if (numExtensions > Number.MAX_SAFE_INTEGER) {
			throw new Error("Number of extensions exceeds maximum safe integer for SessionClient");
		}
		const extensions = new Map<number, Uint8Array>();

		let extId: number;
		for (let i = 0; i < numExtensions; i++) {
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

		this.versions = versions;
		this.extensions = extensions;

		if (len !== this.messageLength) {
			throw new Error(`message length mismatch: expected ${len}, got ${this.messageLength}`);
		}

		return undefined;
	}
}
