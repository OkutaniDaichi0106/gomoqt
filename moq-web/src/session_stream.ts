import type { CancelCauseFunc, Context} from "golikejs/context";
import { withCancelCause } from "golikejs/context";
import type { Reader, Writer, } from "./webtransport";
import { EOF } from "golikejs/io";
import { StreamError } from "./webtransport/error";
import { SessionUpdateMessage } from "./message";
import type { SessionClientMessage } from "./message/session_client";
import type { SessionServerMessage } from "./message/session_server";
import { Cond, Mutex } from "golikejs/sync";
import type { Version,Extensions } from "./internal";

interface SessionStreamInit {
    context: Context;
    writer: Writer;
    reader: Reader;
    client: SessionClientMessage;
    server: SessionServerMessage;
    detectFunc: () => Promise<number>;
}

export class SessionStream {
    #writer: Writer;
    #reader: Reader;
    readonly context: Context;
    #mu: Mutex = new Mutex();
    #cond: Cond = new Cond(this.#mu);
	#clientInfo: ClientInfo;
    #serverInfo: ServerInfo;
    readonly streamId: bigint;

    #detectFunc: () => Promise<number>;

    constructor(init: SessionStreamInit) {
        this.#clientInfo = {
            versions: init.client.versions,
            extensions: init.client.extensions,
            bitrate: 0,
        };
        this.#serverInfo = {
            version: init.server.version,
            extensions: init.server.extensions,
            bitrate: 0,
        };
        this.#writer = init.writer;
        this.#reader = init.reader;
        this.context = init.context;
        this.streamId = this.#writer.streamId ?? this.#reader.streamId ?? 0n;
        this.#detectFunc = init.detectFunc;

        // Start handling session updates (fire and forget)
        this.#handleUpdates().catch(err => {
            console.error(`moq: error in handleUpdates: ${err}`);
        });

        // Start detecting bitrate updates (fire and forget)
        this.#detectUpdates().catch(err => {
            console.error(`moq: error in detectUpdates: ${err}`);
        });
    }

    async #detectUpdates(): Promise<void> {
        while (!this.context.err()) {
            const bitrate = await this.#detectFunc();
            if (this.context.err()) {
                break;
            }
            await this.#update(bitrate);

            // Yield control to the event loop to prevent blocking
            await new Promise(resolve => setTimeout(resolve, 0));
        }
    }

    async #handleUpdates(): Promise<void> {
        while (!this.context.err()) {
            const msg = new SessionUpdateMessage({});
            const err = await msg.decode(this.#reader)
            if (err) {
                // if (err !== EOF ) {
                //     console.error(`moq: error reading SESSION_UPDATE message: ${err}`);
                // }
                break;
            }

            console.debug("moq: SESSION_UPDATE message received.",
                {
                    "message": msg
                }
            );

            this.#serverInfo.bitrate = msg.bitrate;
            this.#cond.broadcast();

            // Yield control to the event loop to prevent blocking
            await new Promise(resolve => setTimeout(resolve, 0));
        }
    }

    // #update sends a session update message to the server.
    // It updates the client's bitrate and notifies the server of significant changes.
    // The bitrate should be originated from the WebTransport API.
    // TODO: get bitrate from WebTransport API and detect significant changes.
    async #update(bitrate: number): Promise<void> {
        const msg = new SessionUpdateMessage({ bitrate });
        const err = await msg.encode(this.#writer);
        if (err) {
            throw new Error(`Failed to encode session update message: ${err}`);
        }

        this.#clientInfo.bitrate = msg.bitrate;

        return;
    }

    async updated(): Promise<void> {
        await this.#cond.wait();
    }

    get clientInfo(): ClientInfo {
        return this.#clientInfo;
    }

    get serverInfo(): ServerInfo {
        return this.#serverInfo;
    }
}

type ClientInfo = {
    versions: Set<Version>;
    extensions: Extensions;
    bitrate: number;
};

type ServerInfo = {
    version: Version;
    extensions: Extensions;
    bitrate: number;
};