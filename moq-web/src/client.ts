import { Session } from "./session.ts";
import type { MOQOptions } from "./options.ts";
import { Extensions,DEFAULT_CLIENT_VERSIONS } from "./internal.ts";
import { DefaultTrackMux, TrackMux } from "./track_mux.ts";

const DefaultWebTransportOptions: WebTransportOptions = {
    allowPooling: false,
    congestionControl: "low-latency",
    requireUnreliable: true,
};

const DefaultMOQOptions: MOQOptions = {
    versions: DEFAULT_CLIENT_VERSIONS,
    extensions: undefined,
    reconnect: false, // TODO: Implement reconnect logic
    // migrate: (url: URL) => false,
    transportOptions: DefaultWebTransportOptions,
};

export class Client {
    #sessions?: Set<Session> = new Set();
    readonly options: MOQOptions;
    #mux: TrackMux;

    constructor(options: MOQOptions = DefaultMOQOptions, mux?: TrackMux) {
        this.options = options;
        this.#mux = mux || new TrackMux();
    }

    async dial(url: string | URL, mux: TrackMux = DefaultTrackMux): Promise<Session> {
        if (this.#sessions === undefined) {
            return Promise.reject(new Error("Client is closed"));
        }

        const transport = new WebTransport(url, this.options.transportOptions);
        const session = new Session({
            conn: transport,
            extensions: this.options.extensions,
            mux
        });
        await session.ready;
        this.#sessions.add(session);
        return session;
    }

    async close(): Promise<void> {
        if (this.#sessions === undefined) {
            return Promise.resolve();
        }

        await Promise.allSettled(Array.from(this.#sessions).map(
            session => session.close()
        ));
        this.#sessions = new Set();
    }

    async abort(): Promise<void> {
        if (this.#sessions === undefined) {
            return;
        }

        await Promise.allSettled(Array.from(this.#sessions).map(
            session => session.close()
        ));

        this.#sessions = new Set();
    }
}

export const MOQ = Client;