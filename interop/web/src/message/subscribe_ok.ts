import { Writer, Reader } from "../io";
import { varintLen } from "../io/len";

export class SubscribeOkMessage {
    groupOrder: bigint;

    constructor(groupOrder: bigint) {
        this.groupOrder = groupOrder;
    }

    length(): number {
        return varintLen(this.groupOrder);
    }

    static async encode(writer: Writer, groupOrder: bigint): Promise<[SubscribeOkMessage?, Error?]> {
        const msg = new SubscribeOkMessage(groupOrder);
        writer.writeVarint(BigInt(msg.length()));
        writer.writeVarint(groupOrder);
        const err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [msg, undefined];
    }

    static async decode(reader: Reader): Promise<[SubscribeOkMessage?, Error?]> {
        const [len, err] = await reader.readVarint();
        if (err) {
            return [undefined, new Error("Failed to read length for SubscribeOkMessage: " + err.message)];
        }

        const [varint, err2] = await reader.readVarint();
        if (err2) {
            return [undefined, new Error("Failed to read groupOrder for SubscribeOkMessage: " + err2.message)];
        }

        return [new SubscribeOkMessage(varint), undefined];
    }
}