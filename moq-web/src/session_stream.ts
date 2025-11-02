import { type CancelCauseFunc, type Context, withCancelCause } from "@okudai/golikejs/context";
import type { Stream } from "./internal/webtransport/mod.ts";
import { SessionUpdateMessage } from "./internal/message/mod.ts";
import type { SessionClientMessage, SessionServerMessage } from "./internal/message/mod.ts";
import { Cond, Mutex } from "@okudai/golikejs/sync";
import type { Version } from "./version.ts";
import { Extensions } from "./extensions.ts";

interface SessionStreamInit {
	context: Context;
	stream: Stream;
	client: SessionClientMessage;
	server: SessionServerMessage;
	detectFunc: () => Promise<number>;
}

export class SessionStream {
	#stream: Stream;
	readonly context: Context;
	#cancelFunc: CancelCauseFunc;
	#mu: Mutex = new Mutex();
	#cond: Cond = new Cond(this.#mu);
	#clientInfo: ClientInfo;
	#serverInfo: ServerInfo;

	#detectFunc: () => Promise<number>;

	constructor(init: SessionStreamInit) {
		this.#stream = init.stream;
		this.#clientInfo = {
			versions: init.client.versions,
			extensions: new Extensions(init.client.extensions),
			bitrate: 0,
		};
		this.#serverInfo = {
			version: init.server.version,
			extensions: new Extensions(init.server.extensions),
			bitrate: 0,
		};
		[this.context, this.#cancelFunc] = withCancelCause(init.context);
		this.#detectFunc = init.detectFunc;

		// Start handling session updates (fire and forget)
		this.#handleUpdates().catch((err) => {
			console.error(`moq: error in handleUpdates: ${err}`);
		});

		// Start detecting bitrate updates (fire and forget)
		this.#detectUpdates().catch((err) => {
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
			await new Promise((resolve) => setTimeout(resolve, 0));
		}
	}

	async #handleUpdates(): Promise<void> {
		while (!this.context.err()) {
			const msg = new SessionUpdateMessage({});
			const err = await msg.decode(this.#stream.readable);
			if (err) {
				this.#cancelFunc(new Error(`moq: failed to decode session update message: ${err}`));
				break;
			}

			console.debug("moq: SESSION_UPDATE message received.", {
				"message": msg,
			});

			this.#serverInfo.bitrate = msg.bitrate;
			this.#cond.broadcast();

			// Yield control to the event loop to prevent blocking
			await new Promise((resolve) => setTimeout(resolve, 0));
		}
	}

	// #update sends a session update message to the server.
	// It updates the client's bitrate and notifies the server of significant changes.
	// The bitrate should be originated from the WebTransport API.
	// TODO: get bitrate from WebTransport API and detect significant changes.
	async #update(bitrate: number): Promise<void> {
		const msg = new SessionUpdateMessage({ bitrate });
		const err = await msg.encode(this.#stream.writable);
		if (err) {
			this.#cancelFunc(new Error(`moq: failed to encode session update message: ${err}`));
			return;
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
