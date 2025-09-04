import { CancelCauseFunc, Context, withCancelCause } from "./internal/context";
import { Reader, Writer } from "./io";
import { StreamError } from "./io/error";
import { SessionUpdateMessage } from "./message";
import { SessionClientMessage } from "./message/session_client";
import { SessionServerMessage } from "./message/session_server";
import { Cond } from "./internal";

export class SessionStream {
    #writer: Writer;
    #reader: Reader;
    #ctx: Context;
    #cancelFunc: CancelCauseFunc;
    #cond: Cond = new Cond();
	readonly client: SessionClientMessage;
    readonly server: SessionServerMessage;
    #clientInfo!: SessionUpdateMessage;
    #serverInfo!: SessionUpdateMessage;

    constructor(ctx: Context, writer: Writer, reader: Reader, client: SessionClientMessage, server: SessionServerMessage) {
        this.client = client;
        this.server = server;
        this.#writer = writer;
        this.#reader = reader;
        [this.#ctx, this.#cancelFunc] = withCancelCause(ctx);

        // Listen for incoming messages
        async () => {
            const msg = new SessionUpdateMessage({});
            let err: Error | undefined;
            for (;;) {
                err = await msg.decode(this.#reader)
                if (err) {
                    break // TODO: handle this situation
                }

                this.#serverInfo = msg;
                this.#cond.broadcast();
            }
        }
    }

    async update(bitrate: bigint): Promise<void> {
        const msg = new SessionUpdateMessage({ bitrate });
        const err = await msg.encode(this.#writer);
        if (err) {
            throw new Error(`Failed to encode session update message: ${err}`);
        }

        this.#clientInfo = msg;

        return;
    }

    async updated(): Promise<void> {
        await this.#cond.wait();
    }

    get clientInfo(): SessionUpdateMessage {
        return this.#clientInfo;
    }

    get serverInfo(): SessionUpdateMessage {
        return this.#serverInfo;
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