import { Writer, Reader } from "../internal/io";

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
        const [_, err] = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [new AnnounceMessage(suffix, active), undefined];
    }

    static async decode(reader: Reader): Promise<[AnnounceMessage?, Error?]> {
        let suffix: string | undefined;
        let err: Error | undefined;
        [suffix, err] = await reader.readString();
        if (err) {
            return [undefined, new Error("Failed to read suffix for Announce")];
        }

        let active: boolean | undefined;
        [active, err] = await reader.readBoolean();
        if (err) {
            return [undefined, new Error("Failed to read active for Announce")];
        }
        if (!suffix || !active) {
            return [undefined, new Error("Suffix or active is undefined")];
        }

        return [new AnnounceMessage(suffix, active), undefined];
    }
}