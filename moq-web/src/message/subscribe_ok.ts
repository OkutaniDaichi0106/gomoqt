import { Writer, Reader, varintLen } from "../io";
import { GroupPeriod } from "../protocol";

export class SubscribeOkMessage {
    groupPeriod: GroupPeriod;

    constructor(groupPeriod: GroupPeriod) {
        this.groupPeriod = groupPeriod;
    }

    length(): number {
        return varintLen(this.groupPeriod);
    }

    static async encode(writer: Writer, groupPeriod: GroupPeriod): Promise<[SubscribeOkMessage?, Error?]> {
        const msg = new SubscribeOkMessage(groupPeriod);
        let err: Error | undefined = undefined;
        writer.writeVarint(msg.length());
        writer.writeVarint(groupPeriod);
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
        let varint: GroupPeriod;
        [varint, err] = await reader.readVarint();
        if (err) {
            return [undefined, err];
        }
        return [new SubscribeOkMessage(varint), undefined];
    }
}