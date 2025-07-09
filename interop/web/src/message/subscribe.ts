import { Writer, Reader } from "../internal/io";

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

    static async encode(writer: Writer, subscribeId: bigint, broadcastPath: string, trackName: string,
         trackPriority: bigint, minGroupSequence: bigint, maxGroupSequence: bigint): Promise<[SubscribeMessage?, Error?]> {
        writer.writeVarint(subscribeId);
        writer.writeString(broadcastPath);
        writer.writeString(trackName);
        writer.writeVarint(trackPriority);
        writer.writeVarint(minGroupSequence);
        writer.writeVarint(maxGroupSequence);
        const [_, err] = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [new SubscribeMessage(subscribeId, broadcastPath, trackName, trackPriority, minGroupSequence, maxGroupSequence), undefined];
    }

    static async decode(reader: Reader): Promise<[SubscribeMessage?, Error?]> {
        let [subscribeId, err] = await reader.readVarint();
        if (err) {
            return [undefined, new Error("Failed to read subscribeId for SubscribeMessage")];
        }
        if (!subscribeId) {
            return [undefined, new Error("subscribeId is undefined")];
        }

        let [broadcastPath, err2] = await reader.readString();
        if (err2) {
            return [undefined, new Error("Failed to read broadcastPath for SubscribeMessage")];
        }
        if (!broadcastPath) {
            return [undefined, new Error("broadcastPath is undefined")];
        }

        let [trackName, err3] = await reader.readString();
        if (err3) {
            return [undefined, new Error("Failed to read trackName for SubscribeMessage")];
        }
        if (!trackName) {
            return [undefined, new Error("trackName is undefined")];
        }

        let [trackPriority, err4] = await reader.readVarint();
        if (err4) {
            return [undefined, new Error("Failed to read trackPriority for SubscribeMessage")];
        }
        if (!trackPriority) {
            return [undefined, new Error("trackPriority is undefined")];
        }

        let [minGroupSequence, err5] = await reader.readVarint();
        if (err5) {
            return [undefined, new Error("Failed to read minGroupSequence for SubscribeMessage")];
        }
        if (!minGroupSequence) {
            return [undefined, new Error("minGroupSequence is undefined")];
        }

        let [maxGroupSequence, err6] = await reader.readVarint();
        if (err6) {
            return [undefined, new Error("Failed to read maxGroupSequence for SubscribeMessage")];
        }
        if (!maxGroupSequence) {
            return [undefined, new Error("maxGroupSequence is undefined")];
        }

        return [new SubscribeMessage(subscribeId, broadcastPath, trackName, trackPriority, minGroupSequence, maxGroupSequence), undefined];
    }
}