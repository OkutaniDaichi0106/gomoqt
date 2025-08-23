import { Writer, Reader } from "../io";

export class FrameMessage {
    data: Uint8Array;

    constructor(data: Uint8Array) {
        this.data = data;
    }

    length(): number {
        return this.data.length;
    }

    static async encode(writer: Writer, data: Uint8Array): Promise<[FrameMessage?, Error?]> {
        writer.writeUint8Array(data);
        const err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [new FrameMessage(data), undefined];
    }

    static async decode(reader: Reader): Promise<[FrameMessage?, Error?]> {
        let arr: Uint8Array<ArrayBufferLike> | undefined;
        let err: Error | undefined;
        [arr, err] = await reader.readUint8Array();
        if (err) {
            return [undefined, err];
        }
        if (arr === undefined) {
            throw new Error("read extData: Uint8Array is undefined");
        }
        return [new FrameMessage(arr), undefined];
    }
}