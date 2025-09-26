import type { CancelCauseFunc, Context} from "./internal/context";
import { withCancelCause } from "./internal/context";
import type { Reader, Writer, } from "./io";
import { EOF } from "./io";
import { StreamError } from "./io/error";
import { SessionUpdateMessage } from "./message";
import type { SessionClientMessage } from "./message/session_client";
import type { SessionServerMessage } from "./message/session_server";
import { Cond } from "./internal";

export class SessionStream {
    #writer: Writer;
    #reader: Reader;
    #ctx: Context;
    #cond: Cond = new Cond();
	readonly client: SessionClientMessage;
    readonly server: SessionServerMessage;
    #clientInfo!: SessionUpdateMessage;
    #serverInfo!: SessionUpdateMessage;

    constructor(connCtx: Context, writer: Writer, reader: Reader, client: SessionClientMessage, server: SessionServerMessage) {
        this.client = client;
        this.server = server;
        this.#writer = writer;
        this.#reader = reader;
        this.#ctx = connCtx;

        this.#handleUpdates()
    }

    async #handleUpdates(): Promise<void> {
        const msg = new SessionUpdateMessage({});
        let err: Error | undefined;
        while (true) {
            // Check if context is cancelled
            if (this.#ctx.err()) {
                break;
            }
            
            err = await msg.decode(this.#reader)
            if (err) {
                if (err !== EOF ) {
                    console.error(`moq: error reading SESSION_UPDATE message: ${err}`);
                }
                break // TODO: handle this situation
            }

            console.debug("moq: SESSION_UPDATE message received.",
                {
                    "message": msg
                }
            );

            this.#serverInfo = msg;
            this.#cond.broadcast();
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

    get context(): Context {
        return this.#ctx;
    }
}