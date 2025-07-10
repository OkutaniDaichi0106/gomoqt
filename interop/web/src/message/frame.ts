import { Writer, Reader } from "../io";

export class FrameMessage {
    data: Uint8Array;

    constructor(data: Uint8Array) {
        this.data = data;
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
        const [dataResult, err] = await reader.readUint8Array();
        if (err) {
            return [undefined, new Error("Failed to read data for FrameMessage")];
        }
        if (!dataResult) {
            return [undefined, new Error("data is undefined")];
        }
        return [new FrameMessage(dataResult), undefined];
    }
}