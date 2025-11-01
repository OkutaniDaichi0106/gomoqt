import { ReceiveStream } from "./reader.ts";
import { SendStream } from "./writer.ts";

export interface StreamInit {
    streamId: number;
    stream : WebTransportBidirectionalStream;
}

export class Stream {
    readonly id: number;
    readonly writable: SendStream;
    readonly readable: ReceiveStream;
    constructor(init: StreamInit) {
        this.id = init.streamId;
        this.writable = new SendStream({
            stream: init.stream.writable,
            streamId: init.streamId,
        });
        this.readable = new ReceiveStream({
            stream: init.stream.readable,
            streamId: init.streamId,
        });
    }
}