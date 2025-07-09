import { Extensions } from "../internal/extensions";
import { Reader, Writer } from "../internal/io";
import { Version } from "../internal/version";

export class SessionServerMessage {
    version: Version;
    extensions: Extensions;

    constructor(version: Version, extensions: Extensions = new Extensions()) {
        this.version = version;
        this.extensions = extensions;
    }

    static async encode(writer: Writer, version: Version, extensions: Extensions = new Extensions()): Promise<[SessionServerMessage?, Error?]> {
        writer.writeVarint(version);

        writer.writeVarint(BigInt(extensions.entries.size)); // Write the number of extensions
        for (const ext of extensions.entries) {
            writer.writeVarint(ext[0]); // Write the extension ID
            writer.writeUint8Array(ext[1]); // Write the extension data
        }

        const [_, err] = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [new SessionServerMessage(version, extensions), undefined];
    }


    static async decode(reader: Reader): Promise<[SessionServerMessage?, Error?]> {
        let [version, err] = await reader.readVarint();
        if (err) {
            return [undefined, new Error("Failed to read version for SessionServer")];
        }
        if (version === undefined) {
            return [undefined, new Error("version is undefined")];
        }

        let [numExtensions, err2] = await reader.readVarint();
        if (err2) {
            return [undefined, new Error("Failed to read number of extensions for SessionServer")];
        }
        if (numExtensions === undefined) {
            return [undefined, new Error("numExtensions is undefined")];
        }
        if (numExtensions < 0) {
            return [undefined, new Error("Invalid number of extensions for SessionServer")];
        }
        if (numExtensions > BigInt(Number.MAX_SAFE_INTEGER)) {
            return [undefined, new Error("Number of extensions exceeds maximum safe integer for SessionServer")];
        }

        const extensions = new Extensions();
        for (let i = 0; i < Number(numExtensions); i++) {
            let [extId, err3] = await reader.readVarint();
            if (err3) {
                return [undefined, new Error(`Failed to read extension ID for SessionServer`)];
            }
            if (extId === undefined) {
                return [undefined, new Error("extId is undefined")];
            }

            let [extData, err4] = await reader.readUint8Array();
            if (err4) {
                return [undefined, new Error(`Failed to read extension data for SessionServer`)];
            }
            if (extData === undefined) {
                return [undefined, new Error("extData is undefined")];
            }

            extensions.addBytes(extId, extData);
        }

        return [new SessionServerMessage(version, extensions), undefined];
    }
}