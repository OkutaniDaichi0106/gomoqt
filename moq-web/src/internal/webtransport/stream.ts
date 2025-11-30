import { ReceiveStream } from "./receive_stream.ts";
import { SendStream } from "./send_stream.ts";

export interface Stream {
	readonly id: bigint;
	readonly writable: SendStream;
	readonly readable: ReceiveStream;
}
export interface StreamInit {
	streamId: bigint;
	stream: WebTransportBidirectionalStream;
}

class StreamClass {
	readonly id: bigint;
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

export const Stream: {
	new (init: StreamInit): Stream;
} = StreamClass;
