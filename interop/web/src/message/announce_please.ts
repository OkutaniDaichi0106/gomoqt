import { Writer, Reader } from "../io";

export class AnnouncePleaseMessage {
    prefix: string;

    constructor(prefix: string) {
        this.prefix = prefix;
    }

    static async encode(writer: Writer, prefix: string): Promise<[AnnouncePleaseMessage?, Error?]> {
        writer.writeString(prefix);
        const err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [new AnnouncePleaseMessage(prefix), undefined];
    }

    static async decode(reader: Reader): Promise<[AnnouncePleaseMessage?, Error?]> {
        const [str, err] = await reader.readString();
        if (err) {
            return [undefined, new Error("Failed to read prefix for AnnouncePlease")];
        }

        return [new AnnouncePleaseMessage(str), undefined];
    }
}