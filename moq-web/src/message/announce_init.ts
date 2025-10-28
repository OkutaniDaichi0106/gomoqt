import type { Writer, Reader } from "../internal/io";
import { varintLen, stringLen } from "../internal/io";

export interface AnnounceInitMessageInit {
    suffixes?: string[];
}

export class AnnounceInitMessage {
    suffixes: string[];

    constructor(init: AnnounceInitMessageInit) {
        this.suffixes = init.suffixes ?? [];
    }

    get messageLength(): number {
        let len = 0;
        len += varintLen(this.suffixes.length);
        for (const suffix of this.suffixes) {
            len += stringLen(suffix);
        }
        return len;
    }

    async encode(writer: Writer): Promise<Error | undefined> {
        writer.writeVarint(this.messageLength);
        writer.writeStringArray(this.suffixes);
        return await writer.flush();
    }

    async decode(reader: Reader): Promise<Error | undefined> {
        let [len, err] = await reader.readVarint();
        if (err) {
            return err;
        }
        [this.suffixes, err] = await reader.readStringArray();
        if (err) {
            return err;
        }

        if (len !== this.messageLength) {
            throw new Error(`message length mismatch: expected ${len}, got ${this.messageLength}`);
        }

        return undefined;
    }
}