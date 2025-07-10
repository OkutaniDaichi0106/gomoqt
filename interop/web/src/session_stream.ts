import { CancelCauseFunc, Context, withCancelCause } from "./internal/context";
import { Reader, Writer } from "./io";
import { StreamError } from "./io/error";
import { SessionUpdateMessage } from "./message";
import { SessionClientMessage } from "./message/session_client";
import { SessionServerMessage } from "./message/session_server";

export class SessionStream {
    #writer: Writer;
    #reader: Reader;
    #ctx: Context;
    #cancelFunc: CancelCauseFunc;
	client: SessionClientMessage;
    server: SessionServerMessage;
    clientInfo!: SessionUpdateMessage;
    serverInfo!: SessionUpdateMessage;

    constructor(ctx: Context, writer: Writer, reader: Reader, client: SessionClientMessage, server: SessionServerMessage) {
        this.client = client;
        this.server = server;
        this.#writer = writer;
        this.#reader = reader;
        [this.#ctx, this.#cancelFunc] = withCancelCause(ctx);

        // Listen for incoming messages
        async () => {
            for (;;) {
                const [result, err] = await SessionUpdateMessage.decode(this.#reader)
                if (err) {
                    // TODO: handle this situation
                    break
                }

                this.serverInfo = result!
            }
        }
    }

    async update(bitrate: bigint): Promise<void> {
        const [result, err] = await SessionUpdateMessage.encode(this.#writer, bitrate);
        if (err) {
            throw new Error(`Failed to encode session update message: ${err}`);
        }

        this.clientInfo = result!;

        return
    }

    close(): void {
        this.#cancelFunc(new Error("SessionStream closed"));
    }

    closeWithError(code: number, message: string): void {
        const reason = new StreamError(code, message);
        this.#cancelFunc(reason);
    }

    get context(): Context {
        return this.#ctx;
    }
}