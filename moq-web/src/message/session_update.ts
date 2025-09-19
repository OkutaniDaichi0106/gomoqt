import type { Writer, Reader } from "../io";
import { varintLen } from "../io/len";

export interface SessionUpdateMessageInit {
    bitrate?: bigint;
}

export class SessionUpdateMessage {
    bitrate: bigint;

    constructor(init: SessionUpdateMessageInit) {
        this.bitrate = init.bitrate ?? 0n;
    }

    get messageLength(): number {
        return varintLen(this.bitrate);
    }

    async encode(writer: Writer): Promise<Error | undefined> {
        writer.writeVarint(this.messageLength);
        writer.writeBigVarint(this.bitrate);
        return await writer.flush();
    }

    async decode(reader: Reader): Promise<Error | undefined> {
        let [len, err] = await reader.readVarint();
        if (err) {
            return err;
        }
        [this.bitrate, err] = await reader.readBigVarint();
        if (err) {
            return err;
        }

        if (len !== this.messageLength) {
            throw new Error(`message length mismatch: expected ${len}, got ${this.messageLength}`);
        }

        return undefined;
    }
}