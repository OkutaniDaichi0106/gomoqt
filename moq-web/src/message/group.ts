import { Writer, Reader } from "../io";
import { varintLen } from "../io/len";
import { GroupSequence } from "../protocol";

export interface GroupMessageInit {
    subscribeId?: bigint;
    sequence?: GroupSequence;
}

export class GroupMessage {
    subscribeId: bigint;
    sequence: GroupSequence;

    constructor(init: GroupMessageInit) {
        this.subscribeId = init.subscribeId ?? 0n;
        this.sequence = init.sequence ?? 0n;
    }

    get messageLength(): number {
        return varintLen(this.subscribeId) + varintLen(this.sequence);
    }

    async encode(writer: Writer): Promise<Error | undefined> {
        writer.writeVarint(this.messageLength + varintLen(this.messageLength));
        writer.writeBigVarint(this.subscribeId);
        writer.writeBigVarint(this.sequence);
        return await writer.flush();
    }

    async decode(reader: Reader): Promise<Error | undefined> {
        let [, err] = await reader.readVarint();
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

        return undefined;
    }
}