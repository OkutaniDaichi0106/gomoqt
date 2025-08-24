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
        let err: Error | undefined;
        writer.writeVarint(msg.length());
        writer.writeBigVarint(bitrate);
        err = await writer.flush();
        if (err) {
            return [undefined, err];
        }
        return [msg, undefined];
    }

    static async decode(reader: Reader): Promise<[SessionUpdateMessage?, Error?]> {
        let err: Error | undefined;
        [, err] = await reader.readVarint();
        if (err) {
            return [undefined, err];
        }
        let varint: bigint;
        [varint, err] = await reader.readBigVarint();
        if (err) {
            return [undefined, err];
        }
        return [new SessionUpdateMessage(varint), undefined];
    }
}