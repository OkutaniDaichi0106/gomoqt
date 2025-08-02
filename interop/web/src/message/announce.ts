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
        writer.writeVarint(BigInt(msg.length()));
        writer.writeBoolean(active);
        writer.writeString(suffix);
        const err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [msg, undefined];
    }

    static async decode(reader: Reader): Promise<[AnnounceMessage?, Error?]> {
        const [len, err] = await reader.readVarint();
        if (err) {
            return [undefined, new Error("Failed to read length for Announce")];
        }

        const [active, err2] = await reader.readBoolean();
        if (err2) {
            return [undefined, new Error("Failed to read active for Announce")];
        }

        const [suffix, err3] = await reader.readString();
        if (err3) {
            return [undefined, new Error("Failed to read suffix for Announce")];
        }

        return [new AnnounceMessage(suffix, active), undefined];
    }
}