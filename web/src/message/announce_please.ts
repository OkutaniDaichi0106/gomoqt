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
        const msg = new AnnouncePleaseMessage(prefix);
        let err: Error | undefined;
        writer.writeVarint(msg.length());
        writer.writeString(prefix);
        err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [msg, undefined];
    }

    static async decode(reader: Reader): Promise<[AnnouncePleaseMessage?, Error?]> {
        let err: Error | undefined;
        [, err] = await reader.readVarint();
        if (err) {
            return [undefined, err];
        }
        let str: string;
        [str, err] = await reader.readString();
        if (err) {
            return [undefined, err];
        }
        return [new AnnouncePleaseMessage(str), undefined];
    }
}