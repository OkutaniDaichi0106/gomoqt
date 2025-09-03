import { Writer, Reader } from "../io";
import { varintLen, stringLen } from "../io/len";
import { GroupSequence } from "../protocol";

export class SubscribeMessage {
    subscribeId: bigint;
    broadcastPath: string;
    trackName: string;
    trackPriority: bigint;
    minGroupSequence: GroupSequence;
    maxGroupSequence: GroupSequence;

    constructor(subscribeId: bigint, broadcastPath: string, trackName: string,
         trackPriority: bigint, minGroupSequence: GroupSequence, maxGroupSequence: GroupSequence) {
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
         trackPriority: bigint, minGroupSequence: GroupSequence, maxGroupSequence: GroupSequence): Promise<[SubscribeMessage?, Error?]> {
        const msg = new SubscribeMessage(subscribeId, broadcastPath, trackName, trackPriority, minGroupSequence, maxGroupSequence);
        let err: Error | undefined = undefined;
        writer.writeVarint(msg.length());
        writer.writeBigVarint(subscribeId);
        writer.writeString(broadcastPath);
        writer.writeString(trackName);
        writer.writeBigVarint(trackPriority);
        writer.writeBigVarint(minGroupSequence);
        writer.writeBigVarint(maxGroupSequence);
        err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [msg, undefined];
    }

    static async decode(reader: Reader): Promise<[SubscribeMessage?, Error?]> {
        let err: Error | undefined;
        [, err] = await reader.readVarint();
        if (err) {
            return [undefined, err];
        }
        let subscribeId: bigint;
        [subscribeId, err] = await reader.readBigVarint();
        if (err) {
            return [undefined, err];
        }
        let broadcastPath: string;
        [broadcastPath, err] = await reader.readString();
        if (err) {
            return [undefined, err];
        }
        let trackName: string;
        [trackName, err] = await reader.readString();
        if (err) {
            return [undefined, err];
        }
        let trackPriority: bigint;
        [trackPriority, err] = await reader.readBigVarint();
        if (err) {
            return [undefined, err];
        }
        let minGroupSequence: GroupSequence;
        [minGroupSequence, err] = await reader.readBigVarint();
        if (err) {
            return [undefined, err];
        }
        let maxGroupSequence: GroupSequence;
        [maxGroupSequence, err] = await reader.readBigVarint();
        if (err) {
            return [undefined, err];
        }
        return [new SubscribeMessage(subscribeId, broadcastPath, trackName, trackPriority, minGroupSequence, maxGroupSequence), undefined];
    }
}