import { ReceiveStream } from "./reader.ts";
import { Stream } from "./stream.ts";
import { SendStream } from "./writer.ts";

/**
 * streamIDCounter manages Stream IDs for WebTransport (QUIC) streams.
 * Stream IDs increment by 4 to maintain the initiator and directionality bits.
 */
export class streamIDCounter {
	clientBiStreamCounter: number = 0;        // client bidirectional
	serverBiStreamCounter: number = 1;  // server bidirectional
	clientUniStreamCounter: number = 2;       // client unidirectional
	serverUniStreamCounter: number = 3; // server unidirectional
}

export class Connection {
    #counter: streamIDCounter;
    #webtransport: WebTransport;

    #uniStreams: ReadableStreamDefaultReader<ReadableStream<Uint8Array<ArrayBufferLike>>>;
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
                streamId: this.#counter.clientBiStreamCounter,
                stream: wtStream
            });
            this.#counter.clientBiStreamCounter += 4;
            return [stream, undefined];
        } catch (e) {
            return [undefined, e as Error];
        }
    }

    async openUniStream(): Promise<[SendStream, undefined] | [undefined, Error]> {
        try {
            const wtStream = await this.#webtransport.createUnidirectionalStream();
            const stream = new SendStream({
                streamId: this.#counter.clientUniStreamCounter,
                stream: wtStream
            });
            this.#counter.clientUniStreamCounter += 4;
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
            streamId: this.#counter.serverBiStreamCounter,
            stream: wtStream
        });
        this.#counter.serverBiStreamCounter += 4;
        return [stream, undefined];
    }

    async acceptUniStream(): Promise<[ReceiveStream, undefined] | [undefined, Error]> {
        const { done, value: wtStream } = await this.#uniStreams.read();
        if (done) {
            return [undefined, new Error("Failed to accept unidirectional stream")];
        }
        const stream = new ReceiveStream({ 
            streamId: this.#counter.serverUniStreamCounter, 
            stream: wtStream 
        });
        this.#counter.serverUniStreamCounter += 4;
        return [stream, undefined];
    }

    close(closeInfo?: WebTransportCloseInfo): void {
        this.#webtransport.close(closeInfo);
    }

    get ready(): Promise<void> {
        return this.#webtransport.ready;
    }

    get closed(): Promise<WebTransportCloseInfo> {
        return this.#webtransport.closed;
    }
}