import { Writer, Reader } from "../io";
import { varintLen } from "../io/len";

export class GroupMessage {
    subscribeId: bigint;
    sequence: bigint;

    constructor(subscribeId: bigint, sequence: bigint) {
        this.subscribeId = subscribeId;
        this.sequence = sequence;
    }

    length(): number {
        return varintLen(this.subscribeId) + varintLen(this.sequence);
    }

    static async encode(writer: Writer, subscribeId: bigint, sequence: bigint): Promise<[GroupMessage?, Error?]> {
        const msg = new GroupMessage(subscribeId, sequence);
        writer.writeVarint(msg.length());
        writer.writeBigVarint(subscribeId);
        writer.writeBigVarint(sequence);
        const err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [msg, undefined];
    }

    static async decode(reader: Reader): Promise<[GroupMessage?, Error?]> {
        let err: Error | undefined;
        [, err] = await reader.readVarint();
        if (err) {
            return [undefined, err];
        }

        let subscribeId: bigint;
        [subscribeId, err] = await reader.readBigVarint();
        if (err) {
            return [undefined, err];
        }

        let sequence: bigint;
        [sequence, err] = await reader.readBigVarint();
        if (err) {
            return [undefined, err];
        }

        return [new GroupMessage(subscribeId, sequence), undefined];
    }
}