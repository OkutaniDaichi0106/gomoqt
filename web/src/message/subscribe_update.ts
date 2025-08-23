import { Writer, Reader } from "../io";
import { varintLen } from "../io/len";

export class SubscribeUpdateMessage {
    trackPriority: bigint;
    minGroupSequence: bigint;
    maxGroupSequence: bigint;

    constructor(trackPriority: bigint, minGroupSequence: bigint, maxGroupSequence: bigint) {
        this.trackPriority = trackPriority;
        this.minGroupSequence = minGroupSequence;
        this.maxGroupSequence = maxGroupSequence;
    }

    length(): number {
        return varintLen(this.trackPriority) + varintLen(this.minGroupSequence) + varintLen(this.maxGroupSequence);
    }

    static async encode(writer: Writer, trackPriority: bigint, minGroupSequence: bigint, maxGroupSequence: bigint): Promise<[SubscribeUpdateMessage?, Error?]> {
        const msg = new SubscribeUpdateMessage(trackPriority, minGroupSequence, maxGroupSequence);
        let err: Error | undefined = undefined;
        writer.writeVarint(msg.length());
        writer.writeBigVarint(trackPriority);
        writer.writeBigVarint(minGroupSequence);
        writer.writeBigVarint(maxGroupSequence);
        err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [msg, undefined];
    }

    static async decode(reader: Reader): Promise<[SubscribeUpdateMessage?, Error?]> {
        let err: Error | undefined;
        [, err] = await reader.readVarint();
        if (err) {
            return [undefined, err];
        }
        let trackPriority: bigint;
        [trackPriority, err] = await reader.readBigVarint();
        if (err) {
            return [undefined, err];
        }
        let minGroupSequence: bigint;
        [minGroupSequence, err] = await reader.readBigVarint();
        if (err) {
            return [undefined, err];
        }
        let maxGroupSequence: bigint;
        [maxGroupSequence, err] = await reader.readBigVarint();
        if (err) {
            return [undefined, err];
        }
        return [new SubscribeUpdateMessage(trackPriority, minGroupSequence, maxGroupSequence), undefined];
    }
}