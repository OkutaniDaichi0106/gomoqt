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
        let err: Error | undefined = undefined;
        writer.writeVarint(msg.length());
        writer.writeBigVarint(groupOrder);
        err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [msg, undefined];
    }

    static async decode(reader: Reader): Promise<[SubscribeOkMessage?, Error?]> {
        let err: Error | undefined;
        [, err] = await reader.readVarint();
        if (err) {
            return [undefined, err];
        }
        let varint: bigint;
        [varint, err] = await reader.readBigVarint();
        if (err) {
            return [undefined, err];
        }
        return [new SubscribeOkMessage(varint), undefined];
    }
}