import { Session } from "./session";
import { MOQOptions } from "./options";
import { Extensions } from "./internal";
import { DefaultTrackMux, TrackMux } from "./track_mux";

const DefaultWebTransportOptions: WebTransportOptions = {
    allowPooling: false,
    congestionControl: "low-latency",
    requireUnreliable: true,
};

const DefaultMOQOptions: MOQOptions = {
    extensions: undefined,
    reconnect: false, // TODO: Implement reconnect logic
    // migrate: (url: URL) => false,
    transport: DefaultWebTransportOptions,
};

export class Client {
    #sessions: Set<Session> = new Set();
    readonly options: MOQOptions;
    #mux: TrackMux;

    constructor(options: MOQOptions = DefaultMOQOptions, mux?: TrackMux) {
        this.options = options;
        this.#mux = mux || new TrackMux();
    }

    async dial(url: string | URL, mux: TrackMux = DefaultTrackMux): Promise<Session> {
        const transport = new WebTransport(url, this.options.transport);
        const session = new Session(transport, undefined, this.options.extensions, mux);
        await session.ready;
        this.#sessions.add(session);
        return session;
    }

    close(): void {
        for (const session of this.#sessions) {
            session.close();
        }

        this.#sessions = new Set();
    }

    abort(): void {
        for (const session of this.#sessions) {
            session.close();
        }

        this.#sessions = new Set();
    }
}