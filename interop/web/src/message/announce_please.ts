import { Writer, Reader } from "../io";
import { stringLen } from "../io/len";

export class AnnouncePleaseMessage {
    prefix: string;

    constructor(prefix: string) {
        this.prefix = prefix;
    }

    length(): number {
        return stringLen(this.prefix);
    }

    static async encode(writer: Writer, prefix: string): Promise<[AnnouncePleaseMessage?, Error?]> {
        const msg = new AnnouncePleaseMessage(prefix)
        writer.writeVarint(BigInt(msg.length()));
        writer.writeString(prefix);
        const err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [msg, undefined];
    }

    static async decode(reader: Reader): Promise<[AnnouncePleaseMessage?, Error?]> {
        const [len, err] = await reader.readVarint();
        if (err) {
            return [undefined, new Error("Failed to read length for AnnouncePlease")];
        }
        const [str, err2] = await reader.readString();
        if (err2) {
            return [undefined, new Error("Failed to read prefix for AnnouncePlease")];
        }

        return [new AnnouncePleaseMessage(str), undefined];
    }
}