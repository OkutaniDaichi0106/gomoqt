import { Writer, Reader } from "../io";

export class SubscribeUpdateMessage {
    trackPriority: bigint;
    minGroupSequence: bigint;
    maxGroupSequence: bigint;

    constructor(trackPriority: bigint, minGroupSequence: bigint, maxGroupSequence: bigint) {
        this.trackPriority = trackPriority;
        this.minGroupSequence = minGroupSequence;
        this.maxGroupSequence = maxGroupSequence;
    }

    static async encode(writer: Writer, trackPriority: bigint, minGroupSequence: bigint, maxGroupSequence: bigint): Promise<[SubscribeUpdateMessage?, Error?]> {
        writer.writeVarint(trackPriority);
        writer.writeVarint(minGroupSequence);
        writer.writeVarint(maxGroupSequence);
        const err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [new SubscribeUpdateMessage(trackPriority, minGroupSequence, maxGroupSequence), undefined];
    }

    static async decode(reader: Reader): Promise<[SubscribeUpdateMessage?, Error?]> {
        let [trackPriority, err] = await reader.readVarint();
        if (err) {
            return [undefined, new Error("Failed to read trackPriority for SubscribeUpdateMessage")];
        }
        if (trackPriority === undefined) {
            return [undefined, new Error("trackPriority is undefined")];
        }

        let [minGroupSequence, err2] = await reader.readVarint();
        if (err2) {
            return [undefined, new Error("Failed to read minGroupSequence for SubscribeUpdateMessage")];
        }
        if (minGroupSequence === undefined) {
            return [undefined, new Error("minGroupSequence is undefined")];
        }

        let [maxGroupSequence, err3] = await reader.readVarint();
        if (err3) {
            return [undefined, new Error("Failed to read maxGroupSequence for SubscribeUpdateMessage")];
        }
        if (maxGroupSequence === undefined) {
            return [undefined, new Error("maxGroupSequence is undefined")];
        }

        return [new SubscribeUpdateMessage(trackPriority, minGroupSequence, maxGroupSequence), undefined];
    }
}