import { ReceiveStream } from "./receive_stream.ts";
import { Stream } from "./stream.ts";
import { SendStream } from "./send_stream.ts";
import { SessionError, SessionErrorInfo } from "./error.ts";

/**
 * streamIDCounter manages Stream IDs for WebTransport (QUIC) streams.
 * Stream IDs increment by 4 to maintain the initiator and directionality bits.
 */
class streamIDCounter {
	clientBiStreamCounter: bigint = 0n; // client bidirectional
	serverBiStreamCounter: bigint = 1n; // server bidirectional
	clientUniStreamCounter: bigint = 2n; // client unidirectional
	serverUniStreamCounter: bigint = 3n; // server unidirectional

	constructor() {}

	countClientBiStream(): bigint {
		const id = this.clientBiStreamCounter;
		this.clientBiStreamCounter += 4n;
		return id;
	}

	countServerBiStream(): bigint {
		const id = this.serverBiStreamCounter;
		this.serverBiStreamCounter += 4n;
		return id;
	}

	countClientUniStream(): bigint {
		const id = this.clientUniStreamCounter;
		this.clientUniStreamCounter += 4n;
		return id;
	}

	countServerUniStream(): bigint {
		const id = this.serverUniStreamCounter;
		this.serverUniStreamCounter += 4n;
		return id;
	}
}

export interface Session {
	openStream(): Promise<[Stream, undefined] | [undefined, Error]>;
	openUniStream(): Promise<[SendStream, undefined] | [undefined, Error]>;
	acceptStream(): Promise<[Stream, undefined] | [undefined, Error]>;
	acceptUniStream(): Promise<[ReceiveStream, undefined] | [undefined, Error]>;
	close(closeInfo?: WebTransportCloseInfo): void;
	ready: Promise<void>;
	closed: Promise<WebTransportCloseInfo>;
}

type WebTransportUnidirectionalStream = ReadableStream<Uint8Array>;
// TODO: Use proper WebTransport types when available

class SessionImpl implements Session {
	#counter: streamIDCounter;
	#webtransport: WebTransport;

	#uniStreams: ReadableStreamDefaultReader<WebTransportUnidirectionalStream>;
	#biStreams: ReadableStreamDefaultReader<WebTransportBidirectionalStream>;

	constructor(webtransport: WebTransport) {
		this.#counter = new streamIDCounter();
		this.#webtransport = webtransport;
		this.#biStreams = this.#webtransport.incomingBidirectionalStreams.getReader();
		this.#uniStreams = this.#webtransport.incomingUnidirectionalStreams.getReader();
	}

	async openStream(): Promise<[Stream, undefined] | [undefined, Error]> {
		try {
			const wtStream = await this.#webtransport.createBidirectionalStream();
			const stream = new Stream({
				streamId: this.#counter.countClientBiStream(),
				stream: wtStream,
			});
			return [stream, undefined];
		} catch (err) {
			if (err instanceof Error) {
				return [undefined, err];
			}
			const wtErr = err as WebTransportError;
			if (wtErr.source === "session") {
				const info = await this.#webtransport.closed;
				if (info.closeCode !== undefined && info.reason !== undefined) {
					return [undefined, new SessionError(info as SessionErrorInfo, true)];
				}
			}
			return [undefined, err as Error];
		}
	}

	async openUniStream(): Promise<[SendStream, undefined] | [undefined, Error]> {
		try {
			const wtStream = await this.#webtransport.createUnidirectionalStream();
			const stream = new SendStream({
				streamId: this.#counter.countClientUniStream(),
				stream: wtStream,
			});
			return [stream, undefined];
		} catch (e) {
			return [undefined, e as Error];
		}
	}

	async acceptStream(): Promise<[Stream, undefined] | [undefined, Error]> {
		const { done, value: wtStream } = await this.#biStreams.read();
		if (done) {
			return [undefined, new Error("Failed to accept stream")];
		}
		const stream = new Stream({
			streamId: this.#counter.countServerBiStream(),
			stream: wtStream,
		});
		return [stream, undefined];
	}

	async acceptUniStream(): Promise<[ReceiveStream, undefined] | [undefined, Error]> {
		const { done, value: wtStream } = await this.#uniStreams.read();
		if (done) {
			return [undefined, new Error("Failed to accept unidirectional stream")];
		}
		const stream = new ReceiveStream({
			streamId: this.#counter.countServerUniStream(),
			stream: wtStream,
		});
		return [stream, undefined];
	}

	close(closeInfo?: WebTransportCloseInfo): void {
		this.#webtransport.close(closeInfo);
		// Cancel readers to resolve any pending read() calls
		this.#biStreams.cancel().catch(() => {});
		this.#uniStreams.cancel().catch(() => {});
	}

	get ready(): Promise<void> {
		return this.#webtransport.ready;
	}

	get closed(): Promise<WebTransportCloseInfo> {
		return this.#webtransport.closed;
	}
}

export const Session: {
	new (webtransport: WebTransport): Session;
} = SessionImpl;
