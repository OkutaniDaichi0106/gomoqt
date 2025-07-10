import { Writer, Reader } from "../io";

export class GroupMessage {
    subscribeId: bigint;
    sequence: bigint;

    constructor(subscribeId: bigint, sequence: bigint) {
        this.subscribeId = subscribeId;
        this.sequence = sequence;
    }

    static async encode(writer: Writer, subscribeId: bigint, sequence: bigint): Promise<[GroupMessage?, Error?]> {
        writer.writeVarint(subscribeId);
        writer.writeVarint(sequence);
        const err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [new GroupMessage(subscribeId, sequence), undefined];
    }

    static async decode(reader: Reader): Promise<[GroupMessage?, Error?]> {
        let [subscribeId, err] = await reader.readVarint();
        if (err) {
            return [undefined, new Error("Failed to read subscribeId for Group")];
        }
        if (subscribeId === undefined) {
            return [undefined, new Error("subscribeId is undefined")];
        }

        let [sequence, err2] = await reader.readVarint();
        if (err2) {
            return [undefined, new Error("Failed to read sequence for Group")];
        }
        if (sequence === undefined) {
            return [undefined, new Error("sequence is undefined")];
        }

        return [new GroupMessage(subscribeId, sequence), undefined];
    }
}