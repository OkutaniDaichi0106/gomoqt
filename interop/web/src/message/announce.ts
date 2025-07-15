import { Writer, Reader } from "../io";

export class AnnounceMessage {
    suffix: string;
    active: boolean;

    constructor(suffix: string, active: boolean) {
        this.suffix = suffix;
        this.active = active;
    }

    static async encode(writer: Writer, suffix: string, active: boolean): Promise<[AnnounceMessage?, Error?]> {
        writer.writeString(suffix);
        writer.writeBoolean(active);
        const err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [new AnnounceMessage(suffix, active), undefined];
    }

    static async decode(reader: Reader): Promise<[AnnounceMessage?, Error?]> {
        const [active, err] = await reader.readBoolean();
        if (err) {
            return [undefined, new Error("Failed to read active for Announce")];
        }

        const [suffix, err2] = await reader.readString();
        if (err2) {
            return [undefined, new Error("Failed to read suffix for Announce")];
        }

        return [new AnnounceMessage(suffix, active), undefined];
    }
}