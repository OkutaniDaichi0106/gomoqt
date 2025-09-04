import { Writer, Reader } from "../io";
import { varintLen, stringLen } from "../io/len";
import { GroupSequence, TrackPriority } from "../protocol";

export interface SubscribeMessageInit {
    subscribeId?: bigint;
    broadcastPath?: string;
    trackName?: string;
    trackPriority?: TrackPriority;
    minGroupSequence?: GroupSequence;
    maxGroupSequence?: GroupSequence;
}

export class SubscribeMessage {
    subscribeId: bigint;
    broadcastPath: string;
    trackName: string;
    trackPriority: TrackPriority;
    minGroupSequence: GroupSequence;
    maxGroupSequence: GroupSequence;

    constructor(init: SubscribeMessageInit) {
        this.subscribeId = init.subscribeId ?? 0n;
        this.broadcastPath = init.broadcastPath ?? "";
        this.trackName = init.trackName ?? "";
        this.trackPriority = init.trackPriority ?? 0;
        this.minGroupSequence = init.minGroupSequence ?? 0n;
        this.maxGroupSequence = init.maxGroupSequence ?? 0n;
    }

    get messageLength(): number {
        return (
            varintLen(this.subscribeId)
            + stringLen(this.broadcastPath)
            + stringLen(this.trackName)
            + varintLen(this.trackPriority)
            + varintLen(this.minGroupSequence)
            + varintLen(this.maxGroupSequence)
        );
    }


    async encode(writer: Writer): Promise<Error | undefined> {
        let err: Error | undefined = undefined;
        writer.writeVarint(this.messageLength + varintLen(this.messageLength));
        writer.writeBigVarint(this.subscribeId);
        writer.writeString(this.broadcastPath);
        writer.writeString(this.trackName);
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
        [this.subscribeId, err] = await reader.readBigVarint();
        if (err) {
            return err;
        }
        [this.broadcastPath, err] = await reader.readString();
        if (err) {
            return err;
        }
        [this.trackName, err] = await reader.readString();
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