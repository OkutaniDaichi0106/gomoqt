import type { Writer, Reader } from "../internal/io";
import { stringLen,varintLen } from "../internal/io";

export interface AnnounceMessageInit {
    suffix?: string;
    active?: boolean;
}

export class AnnounceMessage {
    suffix: string;
    active: boolean;

    constructor(init: AnnounceMessageInit) {
        this.suffix = init.suffix ?? "";
        this.active = init.active ?? false;
    }

    get messageLength(): number {
        return stringLen(this.suffix) + 1; // +1 for the boolean
    }

    async encode(writer: Writer): Promise<Error | undefined> {
        writer.writeVarint(this.messageLength);
        writer.writeBoolean(this.active);
        writer.writeString(this.suffix);
        return await writer.flush();
    }

    async decode(reader: Reader): Promise<Error | undefined> {
        let [len, err] = await reader.readVarint();
        if (err) {
            return err;
        }
        [this.active, err] = await reader.readBoolean();
        if (err) {
            return err;
        }
        [this.suffix, err] = await reader.readString();
        if (err) {
            return err;
        }

        if (len !== this.messageLength) {
            throw new Error(`message length mismatch: expected ${len}, got ${this.messageLength}`);
        }

        return undefined;
    }
}