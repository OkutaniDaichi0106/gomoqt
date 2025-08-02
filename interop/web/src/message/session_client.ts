import { Extensions } from "../internal/extensions";
import { Version } from "../internal/version";
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
        length += varintLen(BigInt(this.versions.size));
        for (const version of this.versions) {
            length += varintLen(version);
        }
        length += varintLen(BigInt(this.extensions.entries.size));
        for (const ext of this.extensions.entries) {
            length += varintLen(ext[0]); // Extension ID length
            length += bytesLen(ext[1]); // Extension data length (includes length prefix)
        }
        return length;
    }

    static async encode(writer: Writer, versions: Set<Version>, extensions: Extensions = new Extensions()): Promise<[SessionClientMessage?, Error?]> {
        const msg = new SessionClientMessage(versions, extensions);
        writer.writeVarint(BigInt(msg.length()));
        writer.writeVarint(BigInt(versions.size));
        for (const version of versions) {
            writer.writeVarint(version);
        }
        writer.writeVarint(BigInt(extensions.entries.size));
        for (const ext of extensions.entries) {
            writer.writeVarint(ext[0]); // Write the extension ID
            writer.writeUint8Array(ext[1]); // Write the extension data
        }

        const err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [msg, undefined];
    }

    static async decode(reader: Reader): Promise<[SessionClientMessage?, Error?]> {
        const [len, err] = await reader.readVarint();
        if (err) {
            return [undefined, new Error("Failed to read length for SessionClient: " + err.message)];
        }

        let [numVersions, err2] = await reader.readVarint();
        if (err2) {
            return [undefined, new Error("Failed to read number of versions for SessionClient: " + err2.message)];
        }
        if (numVersions < 0) {
            return [undefined, new Error("Invalid number of versions for SessionClient")];
        }
        if (numVersions > BigInt(Number.MAX_SAFE_INTEGER)) {
            return [undefined, new Error("Number of versions exceeds maximum safe integer for SessionClient")];
        }

        const versions = new Set<Version>();
        for (let i = 0; i < Number(numVersions); i++) {
            let [version, err2] = await reader.readVarint();
            if (err2) {
                return [undefined, new Error(`Failed to read version ${i} for SessionClient: ${err2.message}`)];
            }
            versions.add(version);
        }

        let [numExtensions, err3] = await reader.readVarint();
        if (err3) {
            return [undefined, new Error("Failed to read number of extensions for SessionClient: " + err3.message)];
        }
        if (numExtensions === undefined) {
            return [undefined, new Error("numExtensions is undefined")];
        }
        if (numExtensions < 0) {
            return [undefined, new Error("Invalid number of extensions for SessionClient")];
        }
        if (numExtensions > BigInt(Number.MAX_SAFE_INTEGER)) {
            return [undefined, new Error("Number of extensions exceeds maximum safe integer for SessionClient")];
        }

        const extensions = new Extensions();
        for (let i = 0; i < Number(numExtensions); i++) {
            let [extId, err4] = await reader.readVarint();
            if (err4) {
                return [undefined, new Error(`Failed to read extension ID ${i} for SessionClient: ${err4.message}`)];
            }

            let [extData, err5] = await reader.readUint8Array();
            if (err5) {
                return [undefined, new Error(`Failed to read extension data for ID ${extId} for SessionClient: ${err5.message}`)];
            }
            if (extData === undefined) {
                return [undefined, new Error("extData is undefined")];
            }
            extensions.addBytes(extId, extData);
        }

        return [new SessionClientMessage(versions, extensions), undefined];
    }
}

