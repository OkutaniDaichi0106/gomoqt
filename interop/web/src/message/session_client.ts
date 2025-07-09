import { Extensions } from "../internal/extensions";
import { Version } from "../internal/version";
import { Writer, Reader } from "../internal/io";


export class SessionClientMessage {
    versions: Set<Version>;
    extensions: Extensions;

    constructor(versions: Set<Version>, extensions: Extensions = new Extensions()) {
        this.versions = versions;
        this.extensions = extensions;
    }

    static async encode(writer: Writer, versions: Set<Version>, extensions: Extensions = new Extensions()): Promise<[SessionClientMessage?, Error?]> {
        writer.writeVarint(BigInt(versions.size));
        for (const version of versions) {
            writer.writeVarint(version);
        }
        writer.writeVarint(BigInt(extensions.entries.size));
        for (const ext of extensions.entries) {
            writer.writeVarint(ext[0]); // Write the extension ID
            writer.writeUint8Array(ext[1]); // Write the extension data
        }

        const [_, err] = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [new SessionClientMessage(versions, extensions), undefined];
    }

    static async decode(reader: Reader): Promise<[SessionClientMessage?, Error?]> {
        let [numVersions, err] = await reader.readVarint();
        if (err) {
            return [undefined, new Error("Failed to read number of versions for SessionClient")];
        }
        if (numVersions === undefined) {
            return [undefined, new Error("numVersions is undefined")];
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
                return [undefined, new Error(`Failed to read version for SessionClient`)];
            }
            if (version === undefined) {
                return [undefined, new Error("version is undefined")];
            }
            versions.add(version);
        }
        
        let [numExtensions, err3] = await reader.readVarint();
        if (err3) {
            return [undefined, new Error("Failed to read number of extensions for SessionClient")];
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
                return [undefined, new Error(`Failed to read extension ID for SessionClient`)];
            }
            if (extId === undefined) {
                return [undefined, new Error("extId is undefined")];
            }

            let [extData, err5] = await reader.readUint8Array();
            if (err5) {
                return [undefined, new Error(`Failed to read extension data for SessionClient`)];
            }
            if (extData === undefined) {
                return [undefined, new Error("extData is undefined")];
            }
            extensions.addBytes(extId, extData);
        }

        return [new SessionClientMessage(versions, extensions), undefined];
    }
}

