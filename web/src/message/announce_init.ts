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
        let err: Error | undefined = undefined;
        writer.writeVarint(msg.length());
        writer.writeStringArray(suffixes);
        err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [msg, undefined];
    }

    static async decode(reader: Reader): Promise<[AnnounceInitMessage?, Error?]> {
        let err: Error | undefined = undefined;
        [, err] = await reader.readVarint();
        if (err) {
            return [undefined, err];
        }
        let suffixes: string[];
        [suffixes, err] = await reader.readStringArray();
        if (err) {
            return [undefined, err];
        }

        return [new AnnounceInitMessage(suffixes), undefined];
    }
}