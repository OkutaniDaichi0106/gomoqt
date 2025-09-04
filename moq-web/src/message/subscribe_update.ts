import { Writer, Reader } from "../io";
import { varintLen } from "../io/len";
import { GroupSequence } from "../protocol";

export interface SubscribeUpdateMessageInit {
    trackPriority?: number;
    minGroupSequence?: GroupSequence;
    maxGroupSequence?: GroupSequence;
}

export class SubscribeUpdateMessage {
    trackPriority: number;
    minGroupSequence: GroupSequence;
    maxGroupSequence: GroupSequence;

    constructor(init: SubscribeUpdateMessageInit) {
        this.trackPriority = init.trackPriority ?? 0;
        this.minGroupSequence = init.minGroupSequence ?? 0n;
        this.maxGroupSequence = init.maxGroupSequence ?? 0n;
    }

    get messageLength(): number {
        return (
            varintLen(this.trackPriority)
             + varintLen(this.minGroupSequence)
             + varintLen(this.maxGroupSequence)
        );
    }

    async encode(writer: Writer): Promise<Error | undefined> {
        // const msg = new SubscribeUpdateMessage(trackPriority, minGroupSequence, maxGroupSequence);
        let err: Error | undefined;
        writer.writeVarint(this.messageLength + varintLen(this.messageLength));
        writer.writeVarint(this.trackPriority);
        writer.writeBigVarint(this.minGroupSequence);
        writer.writeBigVarint(this.maxGroupSequence);
        return await writer.flush();
    }

    async decode(reader: Reader): Promise<Error | undefined> {
        let [, err] = await reader.readVarint();
        if (err) {
            return err;
        }
        [this.trackPriority, err] = await reader.readVarint();
        if (err) {
            return err;
        }
        [this.minGroupSequence, err] = await reader.readBigVarint();
        if (err) {
            return err;
        }
        [this.maxGroupSequence, err] = await reader.readBigVarint();
        if (err) {
            return err;
        }

        return undefined;
    }
}