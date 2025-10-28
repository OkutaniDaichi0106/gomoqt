import type { Writer, Reader } from "../internal/io";
import { varintLen } from "../internal/io";
import type * as internal from "../internal";

export interface GroupMessageInit {
    subscribeId?: bigint;
    sequence?: internal.GroupSequence;
}

export class GroupMessage {
    subscribeId: bigint;
    sequence: internal.GroupSequence;

    constructor(init: GroupMessageInit) {
        this.subscribeId = init.subscribeId ?? 0n;
        this.sequence = init.sequence ?? 0n;
    }

    get messageLength(): number {
        return varintLen(this.subscribeId) + varintLen(this.sequence);
    }

    async encode(writer: Writer): Promise<Error | undefined> {
        writer.writeVarint(this.messageLength);
        writer.writeBigVarint(this.subscribeId);
        writer.writeBigVarint(this.sequence);
        return await writer.flush();
    }

    async decode(reader: Reader): Promise<Error | undefined> {
        let [len, err] = await reader.readVarint();
        if (err) {
            return err;
        }

        [this.subscribeId, err] = await reader.readBigVarint();
        if (err) {
            return err;
        }

        [this.sequence, err] = await reader.readBigVarint();
        if (err) {
            return err;
        }

        if (len !== this.messageLength) {
            throw new Error(`message length mismatch: expected ${len}, got ${this.messageLength}`);
        }

        return undefined;
    }
}