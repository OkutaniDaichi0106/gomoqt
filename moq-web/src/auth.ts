import { withCancelCause } from "@okdaichi/golikejs/context";
import type { CancelCauseFunc, Context } from "@okdaichi/golikejs/context";
import { EOFError } from "@okdaichi/golikejs/io";
import type { Stream } from "./internal/webtransport/mod.ts";
import {
	AuthErrorMessage,
	AuthOkMessage,
	AuthReplyTypes,
	readAuthReplyType,
} from "./internal/message/mod.ts";
import { AuthError } from "./error.ts";

/**
 * What the peer allows for a token: a root, the directory of broadcast paths it
 * applies to, and beneath it the path patterns the presenter may publish and
 * subscribe to. Mirrors Go's `moqt.Grant`.
 */
export interface Grant {
	/** The directory the grant applies to, such as "/room/1/"; "" grants nothing. */
	readonly root: string;
	/** Patterns, relative to the root, of the paths the presenter may publish. */
	readonly publish: readonly string[];
	/** Patterns, relative to the root, of the paths the presenter may subscribe to. */
	readonly subscribe: readonly string[];
	/** How long the grant lasts, in milliseconds, or 0 for no limit. */
	readonly maxAge: number;
}

/** An AUTH_OK or AUTH_ERROR that breaks the protocol. */
export class AuthViolation extends Error {
	constructor(reason: string) {
		super(`moq: auth stream protocol violation: ${reason}`);
		this.name = "AuthViolation";
	}
}

const utf8 = new TextDecoder("utf-8", { fatal: true });

/**
 * Reads one acceptor message: the grant of an AUTH_OK, or an {@link AuthError}
 * for AUTH_ERROR. A message that breaks the protocol is an {@link AuthViolation}.
 */
export async function readAuthReply(
	stream: Stream,
): Promise<[Grant, undefined] | [undefined, Error]> {
	const [type, typeErr] = await readAuthReplyType(stream.readable);
	if (typeErr) return [undefined, typeErr];
	switch (type) {
		case AuthReplyTypes.Ok: {
			const m = new AuthOkMessage();
			let err: Error | undefined;
			try {
				err = await m.decode(stream.readable);
			} catch (e) {
				return [undefined, new AuthViolation(`AUTH_OK: ${e}`)];
			}
			if (err) return [undefined, err];
			if (!grantInRoot(m)) {
				return [undefined, new AuthViolation("AUTH_OK patterns outside a root")];
			}
			return [{
				root: m.root,
				publish: m.publishPatterns,
				subscribe: m.subscribePatterns,
				maxAge: m.expires,
			}, undefined];
		}
		case AuthReplyTypes.Error: {
			const m = new AuthErrorMessage();
			let err: Error | undefined;
			try {
				err = await m.decode(stream.readable);
			} catch (e) {
				return [undefined, new AuthViolation(`AUTH_ERROR: ${e}`)];
			}
			if (err) return [undefined, err];
			let reason: string;
			try {
				reason = utf8.decode(m.reason);
			} catch {
				return [undefined, new AuthViolation("AUTH_ERROR reason is not UTF-8")];
			}
			if (m.code > 0xffffffff) {
				return [undefined, new AuthViolation("AUTH_ERROR code over 32 bits")];
			}
			return [undefined, new AuthError(m.code, reason)];
		}
		default:
			return [undefined, new AuthViolation(`auth reply type ${type}`)];
	}
}

// grantInRoot reports whether m's root is a prefix of broadcast paths, or
// empty with no patterns, and each pattern lies beneath it.
function grantInRoot(m: AuthOkMessage): boolean {
	const patterns = [...m.publishPatterns, ...m.subscribePatterns];
	if (m.root === "") return patterns.length === 0;
	if (!m.root.startsWith("/") || !m.root.endsWith("/") || /[{}]/.test(m.root)) return false;
	return patterns.every((p) => p.startsWith("/"));
}

/**
 * A token presented to the peer, and what the peer granted for it. It lasts
 * until {@link close} withdraws the token, the peer revokes it, or the session
 * ends. Mirrors Go's `moqt.Credential`.
 */
export class Credential {
	readonly #stream: Stream;
	#grant: Grant;
	/**
	 * Ends when the grant does. Its error is an {@link AuthError} when the peer
	 * revoked the grant, and an Error otherwise, such as when the token was
	 * withdrawn.
	 */
	readonly context: Context;
	readonly #cancel: CancelCauseFunc;

	/** @internal Made by {@link Session.auth}. */
	constructor(
		stream: Stream,
		grant: Grant,
		sessCtx: Context,
		violation: (reason: string) => void,
	) {
		this.#stream = stream;
		this.#grant = grant;
		[this.context, this.#cancel] = withCancelCause(sessCtx);
		this.#readReplies(violation);
	}

	/** The latest grant the peer sent for the token. */
	grant(): Grant {
		return this.#grant;
	}

	/**
	 * Withdraws the token, which ends its grant on the peer, and ends
	 * {@link context} at once.
	 */
	async close(): Promise<void> {
		this.#cancel(new Error("moq: credential closed"));
		await this.#stream.writable.close().catch(() => {});
	}

	// readReplies follows the peer's answers after the first: a further AUTH_OK
	// replaces the grant, and AUTH_ERROR or the end of the stream ends it.
	async #readReplies(violation: (reason: string) => void): Promise<void> {
		while (true) {
			const [grant, err] = await readAuthReply(this.#stream);
			if (err) {
				let cause: Error | undefined;
				if (err instanceof AuthError) {
					cause = err;
				} else if (err instanceof AuthViolation) {
					violation(err.message);
					cause = err;
				} else if (!(err instanceof EOFError)) {
					cause = err;
				}
				this.#stream.readable.cancel(authStreamCancelled).catch(() => {});
				this.#stream.writable.cancel(authStreamCancelled).catch(() => {});
				this.#cancel(cause ?? new Error("moq: credential ended"));
				return;
			}
			this.#grant = grant;
		}
	}
}

/** Resets the side of an Auth Stream whose token is over. */
export const authStreamCancelled = 0x1;
