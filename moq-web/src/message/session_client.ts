import { Extensions,Version } from "../internal";
import { Writer, Reader } from "../io";
import { varintLen, bytesLen } from "../io/len";


export class SessionClientMessage {
    versions: Set<Version>;
    extensions: Extensions;

    constructor(versions: Set<Version>, extensions: Extensions = new Extensions()) {
        this.versions = versions;
        this.extensions = extensions;
    }

    length(): number {
        let length = 0;
        length += varintLen(this.versions.size);
        for (const version of this.versions) {
            length += varintLen(version);
        }
        length += varintLen(this.extensions.entries.size);
        for (const ext of this.extensions.entries) {
            length += varintLen(ext[0]); // Extension ID length
            length += bytesLen(ext[1]); // Extension data length (includes length prefix)
        }
        return length;
    }

    static async encode(writer: Writer, versions: Set<Version>, extensions: Extensions = new Extensions()): Promise<[SessionClientMessage?, Error?]> {
        const msg = new SessionClientMessage(versions, extensions);
        let err: Error | undefined;
        writer.writeVarint(msg.length());
        writer.writeVarint(versions.size);
        for (const version of versions) {
            writer.writeBigVarint(version);
        }
        writer.writeVarint(extensions.entries.size);
        for (const ext of extensions.entries) {
            writer.writeVarint(ext[0]); // Write the extension ID
            writer.writeUint8Array(ext[1]); // Write the extension data
        }
        err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [msg, undefined];
    }

    static async decode(reader: Reader): Promise<[SessionClientMessage?, Error?]> {
        let err: Error | undefined;
        [, err] = await reader.readVarint();
        if (err) {
            return [undefined, err];
        }
        let numVersions: number;
        [numVersions, err] = await reader.readVarint();
        if (err) {
            return [undefined, err];
        }
        if (numVersions < 0) {
            throw new Error("Invalid number of versions for SessionClient");
        }
        if (numVersions > Number.MAX_SAFE_INTEGER) {
            throw new Error("Number of versions exceeds maximum safe integer for SessionClient");
        }
        const versions = new Set<Version>();
        for (let i = 0; i < numVersions; i++) {
            let version: bigint;
            [version, err] = await reader.readBigVarint();
            if (err) {
                return [undefined, err];
            }
            versions.add(version);
        }
        let numExtensions: number;
        [numExtensions, err] = await reader.readVarint();
        if (err) {
            return [undefined, err];
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
        const extensions = new Extensions();

        let extId: number;
        for (let i = 0; i < numExtensions; i++) {
            [extId, err] = await reader.readVarint();
            if (err) {
                return [undefined, err];
            }
            let extData: Uint8Array | undefined;
            [extData, err] = await reader.readUint8Array();
            if (err) {
                return [undefined, err];
            }
            if (extData === undefined) {
                throw new Error("read extData: Uint8Array is undefined");
            }
            extensions.addBytes(extId, extData);
        }
        return [new SessionClientMessage(versions, extensions), undefined];
    }
}

