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
        length += varintLen(this.extensions.entries.size);
        for (const ext of this.extensions.entries) {
            length += varintLen(ext[0]);
            length += bytesLen(ext[1]);
        }
        return length;
    }

    static async encode(writer: Writer, version: Version, extensions: Extensions = new Extensions()): Promise<[SessionServerMessage?, Error?]> {
        const msg = new SessionServerMessage(version, extensions);
        let err: Error | undefined;
        writer.writeVarint(msg.length());
        writer.writeBigVarint(version);
        writer.writeVarint(extensions.entries.size); // Write the number of extensions
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


    static async decode(reader: Reader): Promise<[SessionServerMessage?, Error?]> {
        let err: Error | undefined;
        [, err] = await reader.readVarint();
            if (err) {
                return [undefined, err];
            }
        let version: bigint;
        [version, err] = await reader.readBigVarint();
        if (err) {
                return [undefined, err];
            }
        let extensionCount: number;
        [extensionCount, err] = await reader.readVarint();
        if (err) {
            return [undefined, err];
        }
        if (extensionCount < 0) {
            throw new Error("Invalid number of extensions for SessionServer");
        }
        if (extensionCount > Number.MAX_SAFE_INTEGER) {
            throw new Error("Number of extensions exceeds maximum safe integer for SessionServer");
        }
        const extensions = new Extensions();

        let extId: number;
        for (let i = 0; i < extensionCount; i++) {
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
        return [new SessionServerMessage(version, extensions), undefined];
    }
}