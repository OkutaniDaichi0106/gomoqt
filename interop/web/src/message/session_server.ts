import { Extensions } from "../internal/extensions";
import { Reader, Writer } from "../io";
import { Version } from "../internal/version";
import { varintLen, bytesLen } from "../io/len";

export class SessionServerMessage {
    version: Version;
    extensions: Extensions;

    constructor(version: Version, extensions: Extensions = new Extensions()) {
        this.version = version;
        this.extensions = extensions;
    }

    length(): number {
        let length = 0;
        length += varintLen(this.version);
        length += varintLen(BigInt(this.extensions.entries.size));
        for (const ext of this.extensions.entries) {
            length += varintLen(ext[0]);
            length += bytesLen(ext[1]);
        }
        return length;
    }

    static async encode(writer: Writer, version: Version, extensions: Extensions = new Extensions()): Promise<[SessionServerMessage?, Error?]> {
        const msg = new SessionServerMessage(version, extensions);
        writer.writeVarint(BigInt(msg.length()));
        writer.writeVarint(version);

        writer.writeVarint(BigInt(extensions.entries.size)); // Write the number of extensions
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


    static async decode(reader: Reader): Promise<[SessionServerMessage?, Error?]> {
        const [len, err] = await reader.readVarint();
        if (err) {
            return [undefined, new Error("Failed to read length for SessionServer")];
        }

        const [version, err2] = await reader.readVarint();
        if (err2) {
            return [undefined, new Error("Failed to read version for SessionServer: " + err2.message)];
        }

        const [extensionCount, err3] = await reader.readVarint();
        if (err3) {
            return [undefined, new Error("Failed to read number of extensions for SessionServer: " + err3.message)];
        }
        if (extensionCount < 0) {
            return [undefined, new Error("Invalid number of extensions for SessionServer")];
        }
        if (extensionCount > BigInt(Number.MAX_SAFE_INTEGER)) {
            return [undefined, new Error("Number of extensions exceeds maximum safe integer for SessionServer")];
        }


        const extensions = new Extensions();
        for (let i = 0; i < Number(extensionCount); i++) {
            let [extId, err3] = await reader.readVarint();
            if (err3) {
                return [undefined, new Error(`Failed to read extension ID for SessionServer`)];
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