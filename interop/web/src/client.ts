import { Session } from "./session";
import { MOQOptions } from "./options";

const DefaultWebTransportOptions: WebTransportOptions = {
    allowPooling: false,
    congestionControl: "low-latency",
    requireUnreliable: true,
};

const DefaultMOQOptions: MOQOptions = {
    extensions: undefined,
    reconnect: false,
    migrate: (url: URL) => false,
    transport: DefaultWebTransportOptions,
};

export class Client {
    #sessions: Set<Session> = new Set();
    readonly options: MOQOptions;

    constructor(options: MOQOptions = DefaultMOQOptions) {
        this.options = options;
    }

    async dial(url: string | URL): Promise<Session> {
        const session = new Session(new WebTransport(url, this.options.transport));
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

export const MOQ = Client;