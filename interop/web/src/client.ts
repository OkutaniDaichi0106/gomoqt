import { Session } from "./session";
import { MOQOptions } from "./options";

let DefaultWebTransportOptions: WebTransportOptions = {
    allowPooling: false,
    congestionControl: "low-latency",
    requireUnreliable: true,
};
    
export class Client {
    transportOptions: WebTransportOptions;
    sessions: Set<Session> = new Set();
    closed: boolean = false;

    constructor(transportOptions: WebTransportOptions = DefaultWebTransportOptions) {
        this.transportOptions = transportOptions;
    }

    dial(url: string | URL, options?: MOQOptions): Promise<Session> {
        return new Promise((resolve, reject) => {
            const conn = new WebTransport(url, this.transportOptions);
            conn.ready.then(() => {
                const session = new Session(conn);

                this.sessions.add(session);
                resolve(session);
            }).catch(reject);
        });
    }

    close(): void {
        for (const session of this.sessions) {
            session.close();
        }

        this.sessions = new Set();
    }

    abort(): void {
        for (const session of this.sessions) {
            session.close();
        }

        this.sessions = new Set();
    }
}