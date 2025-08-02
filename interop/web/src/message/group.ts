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
        writer.writeVarint(BigInt(msg.length()));
        writer.writeVarint(subscribeId);
        writer.writeVarint(sequence);
        const err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [msg, undefined];
    }

    static async decode(reader: Reader): Promise<[GroupMessage?, Error?]> {
        const [len, err] = await reader.readVarint();
        if (err) {
            return [undefined, new Error("Failed to read length for Group")];
        }

        const [subscribeId, err2] = await reader.readVarint();
        if (err2) {
            return [undefined, new Error("Failed to read subscribeId for Group")];
        }


        const [sequence, err3] = await reader.readVarint();
        if (err3) {
            return [undefined, new Error("Failed to read sequence for Group")];
        }

        return [new GroupMessage(subscribeId, sequence), undefined];
    }
}