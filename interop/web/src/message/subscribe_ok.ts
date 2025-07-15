import { Writer, Reader } from "../io";

export class SubscribeOkMessage {
    groupOrder: bigint;

    constructor(groupOrder: bigint) {
        this.groupOrder = groupOrder;
    }

    static async encode(writer: Writer, groupOrder: bigint): Promise<[SubscribeOkMessage?, Error?]> {
        writer.writeVarint(groupOrder);
        const err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [new SubscribeOkMessage(groupOrder), undefined];
    }

    static async decode(reader: Reader): Promise<[SubscribeOkMessage?, Error?]> {
        let [varint, err] = await reader.readVarint();
        if (err) {
            return [undefined, new Error("Failed to read groupOrder for SubscribeOkMessage")];
        }

        return [new SubscribeOkMessage(varint), undefined];
    }
}