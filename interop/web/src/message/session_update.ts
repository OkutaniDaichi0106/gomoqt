import { Writer, Reader } from "../io";

export class SessionUpdateMessage {
    bitrate: bigint;

    constructor(bitrate: bigint) {
        this.bitrate = bitrate;
    }

    static async encode(writer: Writer, bitrate: bigint): Promise<[SessionUpdateMessage?, Error?]> {
        writer.writeVarint(bitrate);
        const err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [new SessionUpdateMessage(bitrate), undefined];
    }

    static async decode(reader: Reader): Promise<[SessionUpdateMessage?, Error?]> {
        const [varint, err] = await reader.readVarint();
        if (err) {
            return [undefined, new Error("Failed to read bitrate for SessionUpdateMessage")];
        }

        return [new SessionUpdateMessage(varint), undefined];
    }
}