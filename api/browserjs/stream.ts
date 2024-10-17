export class Stream {
	#stream: WebTransportBidirectionalStream
	#type: StreamType

	constructor(stream: WebTransportBidirectionalStream, type: StreamType) {
		// Send the stream type
		stream.writable.getWriter().write(new Uint8Array([type]))

		// Register the stream and type 
		this.#stream = stream
		this.#type = type
	}

	async write(data: Uint8Array) {
		this.#stream.writable.getWriter().write(data)
	}
}

export enum StreamType {
	SETUP = 0x00,
	ANNOUNCE = 0x01,
	SUBSCRIBE = 0x02,
	TRACK_STATUS = 0x03,
}