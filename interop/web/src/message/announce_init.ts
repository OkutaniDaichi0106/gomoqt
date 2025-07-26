import { Writer, Reader } from "../io";

export class AnnounceInitMessage {
    suffixes: string[];

    constructor(suffixes: string[]) {
        this.suffixes = suffixes;
    }

    static async encode(writer: Writer, suffixes: string[]): Promise<[AnnounceInitMessage?, Error?]> {
        writer.writeStringArray(suffixes);
        const err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [new AnnounceInitMessage(suffixes), undefined];
    }

    static async decode(reader: Reader): Promise<[AnnounceInitMessage?, Error?]> {
        const [suffixes, err] = await reader.readStringArray();
        if (err) {
            return [undefined, new Error("Failed to read suffixes for AnnounceInit")];
        }

        return [new AnnounceInitMessage(suffixes), undefined];
    }
}