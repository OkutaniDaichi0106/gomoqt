import {Stream, StreamType} from "./stream"

export class Publisher {
	conn: WebTransport

	constructor(conn: WebTransport) {
		this.conn = conn
	}

	async setup(conn: WebTransport) {
		// Create a bidirectional stream to set up moqt
		const stream = await conn.createBidirectionalStream()

		const setupStream = new Stream(stream, StreamType.SETUP)

		setupStream
	}
}