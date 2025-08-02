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
        writer.writeVarint(BigInt(msg.length()));
        writer.writeVarint(trackPriority);
        writer.writeVarint(minGroupSequence);
        writer.writeVarint(maxGroupSequence);
        const err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [msg, undefined];
    }

    static async decode(reader: Reader): Promise<[SubscribeUpdateMessage?, Error?]> {
        const [len, err] = await reader.readVarint();
        if (err) {
            return [undefined, new Error("Failed to read length for SubscribeUpdateMessage")];
        }

        const [trackPriority, err2] = await reader.readVarint();
        if (err2) {
            return [undefined, new Error("Failed to read trackPriority for SubscribeUpdateMessage: " + err2.message)];
        }

        const [minGroupSequence, err3] = await reader.readVarint();
        if (err3) {
            return [undefined, new Error("Failed to read minGroupSequence for SubscribeUpdateMessage: " + err3.message)];
        }

        const [maxGroupSequence, err4] = await reader.readVarint();
        if (err4) {
            return [undefined, new Error("Failed to read maxGroupSequence for SubscribeUpdateMessage: " + err4.message)];
        }

        return [new SubscribeUpdateMessage(trackPriority, minGroupSequence, maxGroupSequence), undefined];
    }
}