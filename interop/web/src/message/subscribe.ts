import { Writer, Reader } from "../io";
import { varintLen, stringLen } from "../io/len";

export class SubscribeMessage {
    subscribeId: bigint;
    broadcastPath: string;
    trackName: string;
    trackPriority: bigint;
    minGroupSequence: bigint;
    maxGroupSequence: bigint;

    constructor(subscribeId: bigint, broadcastPath: string, trackName: string,
         trackPriority: bigint, minGroupSequence: bigint, maxGroupSequence: bigint) {
        this.subscribeId = subscribeId;
        this.broadcastPath = broadcastPath;
        this.trackName = trackName;
        this.trackPriority = trackPriority;
        this.minGroupSequence = minGroupSequence;
        this.maxGroupSequence = maxGroupSequence;
    }

    length(): number {
        return (
            varintLen(this.subscribeId) +
            stringLen(this.broadcastPath) +
            stringLen(this.trackName) +
            varintLen(this.trackPriority) +
            varintLen(this.minGroupSequence) +
            varintLen(this.maxGroupSequence)
        );
    }


    static async encode(writer: Writer, subscribeId: bigint, broadcastPath: string, trackName: string,
         trackPriority: bigint, minGroupSequence: bigint, maxGroupSequence: bigint): Promise<[SubscribeMessage?, Error?]> {
        const msg = new SubscribeMessage(subscribeId, broadcastPath, trackName, trackPriority, minGroupSequence, maxGroupSequence);
        writer.writeVarint(BigInt(msg.length()));
        writer.writeVarint(subscribeId);
        writer.writeString(broadcastPath);
        writer.writeString(trackName);
        writer.writeVarint(trackPriority);
        writer.writeVarint(minGroupSequence);
        writer.writeVarint(maxGroupSequence);
        const err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [msg, undefined];
    }

    static async decode(reader: Reader): Promise<[SubscribeMessage?, Error?]> {
        const [len, err] = await reader.readVarint();
        if (err) {
            return [undefined, new Error("Failed to read length for SubscribeMessage: " + err.message)];
        }

        const [subscribeId, err2] = await reader.readVarint();
        if (err2) {
            return [undefined, new Error("Failed to read subscribeId for SubscribeMessage")];
        }

        const [broadcastPath, err3] = await reader.readString();
        if (err3) {
            return [undefined, new Error("Failed to read broadcastPath for SubscribeMessage")];
        }

        const [trackName, err4] = await reader.readString();
        if (err4) {
            return [undefined, new Error("Failed to read trackName for SubscribeMessage")];
        }

        const [trackPriority, err5] = await reader.readVarint();
        if (err5) {
            return [undefined, new Error("Failed to read trackPriority for SubscribeMessage")];
        }

        const [minGroupSequence, err6] = await reader.readVarint();
        if (err6) {
            return [undefined, new Error("Failed to read minGroupSequence for SubscribeMessage")];
        }

        const [maxGroupSequence, err7] = await reader.readVarint();
        if (err7) {
            return [undefined, new Error("Failed to read maxGroupSequence for SubscribeMessage")];
        }

        return [new SubscribeMessage(subscribeId, broadcastPath, trackName, trackPriority, minGroupSequence, maxGroupSequence), undefined];
    }
}