import { Writer, Reader } from "../io";
import { varintLen, stringLen } from "../io/len";

export class AnnounceInitMessage {
    suffixes: string[];

    constructor(suffixes: string[]) {
        this.suffixes = suffixes;
    }

    length(): number {
        let len = 0;
        len += varintLen(this.suffixes.length);
        for (const suffix of this.suffixes) {
            len += stringLen(suffix);
        }
        return len;
    }

    static async encode(writer: Writer, suffixes: string[]): Promise<[AnnounceInitMessage?, Error?]> {
        const msg = new AnnounceInitMessage(suffixes);
        writer.writeVarint(BigInt(msg.length()));
        writer.writeStringArray(suffixes);
        const err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [msg, undefined];
    }

    static async decode(reader: Reader): Promise<[AnnounceInitMessage?, Error?]> {
        const [len, err] = await reader.readVarint();
        if (err) {
            return [undefined, new Error("Failed to read length for AnnounceInit")];
        }
        const [suffixes, err2] = await reader.readStringArray();
        if (err2) {
            return [undefined, new Error("Failed to read suffixes for AnnounceInit")];
        }

        return [new AnnounceInitMessage(suffixes), undefined];
    }
}