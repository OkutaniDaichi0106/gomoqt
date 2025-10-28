import type { Version} from "../internal";
import { Extensions,DEFAULT_VERSION } from "../internal";
import type { Reader, Writer } from "../internal/io";
import { varintLen, bytesLen } from "../internal/io";

export interface SessionServerMessageInit {
    version?: Version;
    extensions?: Extensions;
}

export class SessionServerMessage {
    version: Version;
    extensions: Extensions;

    constructor(init: SessionServerMessageInit) {
        this.version = init.version ?? DEFAULT_VERSION;
        this.extensions = init.extensions ?? new Extensions();
    }

    get messageLength(): number {
        let length = 0;
        length += varintLen(this.version);
        length += varintLen(this.extensions.entries.size);
        for (const ext of this.extensions.entries) {
            length += varintLen(ext[0]);
            length += bytesLen(ext[1]);
        }
        return length;
    }

    async encode(writer: Writer): Promise<Error | undefined> {
        writer.writeVarint(this.messageLength);
        writer.writeBigVarint(this.version);
        writer.writeVarint(this.extensions.entries.size); // Write the number of extensions
        for (const ext of this.extensions.entries) {
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

        [this.version, err] = await reader.readBigVarint();
        if (err) {
            return err;
        }

        let extensionCount: number;
        [extensionCount, err] = await reader.readVarint();
        if (err) {
            return err;
        }

        const extensions = new Extensions();

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
            extensions.addBytes(extId, extData);
        }

        this.extensions = extensions;

        if (len !== this.messageLength) {
            throw new Error(`message length mismatch: expected ${len}, got ${this.messageLength}`);
        }

        return undefined;
    }
}