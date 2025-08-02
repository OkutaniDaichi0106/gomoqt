import { Writer, Reader } from "../io";
import { varintLen } from "../io/len";

export class SessionUpdateMessage {
    bitrate: bigint;

    constructor(bitrate: bigint) {
        this.bitrate = bitrate;
    }

    length(): number {
        return varintLen(this.bitrate);
    }

    static async encode(writer: Writer, bitrate: bigint): Promise<[SessionUpdateMessage?, Error?]> {
        const msg = new SessionUpdateMessage(bitrate);
        writer.writeVarint(BigInt(msg.length()));
        writer.writeVarint(bitrate);
        const err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [msg, undefined];
    }

    static async decode(reader: Reader): Promise<[SessionUpdateMessage?, Error?]> {
        const [len, err] = await reader.readVarint();
        if (err) {
            return [undefined, new Error("Failed to read length for SessionUpdateMessage")];
        }
        const [varint, err2] = await reader.readVarint();
        if (err2) {
            return [undefined, new Error("Failed to read bitrate for SessionUpdateMessage: " + err2.message)];
        }

        return [new SessionUpdateMessage(varint), undefined];
    }
}