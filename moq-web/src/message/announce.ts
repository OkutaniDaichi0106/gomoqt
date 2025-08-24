import { Writer, Reader } from "../io";
import { stringLen } from "../io/len";

export class AnnounceMessage {
    suffix: string;
    active: boolean;

    constructor(suffix: string, active: boolean) {
        this.suffix = suffix;
        this.active = active;
    }

    length(): number {
        return stringLen(this.suffix) + 1; // +1 for the boolean
    }

    static async encode(writer: Writer, suffix: string, active: boolean): Promise<[AnnounceMessage?, Error?]> {
        const msg = new AnnounceMessage(suffix, active);
        let err: Error | undefined = undefined;
        writer.writeVarint(msg.length());
        writer.writeBoolean(active);
        writer.writeString(suffix);
        err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [msg, undefined];
    }

    static async decode(reader: Reader): Promise<[AnnounceMessage?, Error?]> {
        let err: Error | undefined = undefined;
        [, err] = await reader.readVarint();
        if (err) {
            return [undefined, err];
        }
        let active = false;
        [active, err] = await reader.readBoolean();
        if (err) {
            return [undefined, err];
        }
        let suffix = "";
        [suffix, err] = await reader.readString();
        if (err) {
            return [undefined, err];
        }
        return [new AnnounceMessage(suffix, active), undefined];
    }
}