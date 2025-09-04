import { Writer, Reader, varintLen } from "../io";
import { GroupPeriod } from "../protocol";

export interface SubscribeOkMessageInit {
    groupPeriod?: GroupPeriod;
}

export class SubscribeOkMessage {
    groupPeriod: GroupPeriod;

    constructor(init: SubscribeOkMessageInit) {
        this.groupPeriod = init.groupPeriod ?? 0;
    }

    get messageLength(): number {
        return varintLen(this.groupPeriod);
    }

    async encode(writer: Writer): Promise<Error | undefined> {
        let err: Error | undefined = undefined;
        writer.writeVarint(this.messageLength + varintLen(this.messageLength));
        writer.writeVarint(this.groupPeriod);
        return await writer.flush();
    }

    async decode(reader: Reader): Promise<Error | undefined> {
        let [, err] = await reader.readVarint();
        if (err) {
            return err;
        }
        [this.groupPeriod, err] = await reader.readVarint();
        if (err) {
            return err;
        }

        return undefined;
    }
}