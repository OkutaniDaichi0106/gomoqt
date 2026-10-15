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
import { WebTransportStreamError } from "./internal/webtransport/mod.ts";

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

// Limits of the Auth Stream's fields, as in Go.
const maxAuthExpires = Number.MAX_SAFE_INTEGER; // 2^53 - 1 milliseconds
const maxAuthReasonLength = 8192;

const utf8 = new TextDecoder("utf-8", { fatal: true });

// decodeViolation classifies a decode error: the stream ending or being reset
// is not the peer's message, while anything else it sent wrong is a protocol
// violation.
function decodeViolation(what: string, err: Error): Error {
	if (err instanceof EOFError || err instanceof WebTransportStreamError) return err;
	return new AuthViolation(`${what}: ${err.message}`);
}

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
			const err = await m.decode(stream.readable);
			if (err) return [undefined, decodeViolation("AUTH_OK", err)];
			if (m.expires > maxAuthExpires) {
				return [undefined, new AuthViolation(`AUTH_OK expires ${m.expires} ms`)];
			}
			const invalid = invalidGrant(m);
			if (invalid) return [undefined, new AuthViolation(`AUTH_OK ${invalid}`)];
			return [{
				root: m.root,
				publish: m.publishPatterns,
				subscribe: m.subscribePatterns,
				maxAge: m.expires,
			}, undefined];
		}
		case AuthReplyTypes.Error: {
			const m = new AuthErrorMessage();
			const err = await m.decode(stream.readable);
			if (err) return [undefined, decodeViolation("AUTH_ERROR", err)];
			if (m.code > 0xffffffff || m.reason.length > maxAuthReasonLength) {
				return [undefined, new AuthViolation("malformed AUTH_ERROR")];
			}
			let reason: string;
			try {
				reason = utf8.decode(m.reason);
			} catch {
				return [undefined, new AuthViolation("AUTH_ERROR reason is not UTF-8")];
			}
			return [undefined, new AuthError(m.code, reason)];
		}
		default:
			return [undefined, new AuthViolation(`auth reply type ${type}`)];
	}
}

// invalidGrant describes what makes m's grant invalid, as Go's parseGrant
// judges it, or returns undefined for a valid one: a root is a prefix of
// broadcast paths without wildcards, a grant without one has no patterns, and
// each pattern is a path pattern beneath it.
function invalidGrant(m: AuthOkMessage): string | undefined {
	const patterns = [...m.publishPatterns, ...m.subscribePatterns];
	if (m.root === "") {
		return patterns.length === 0 ? undefined : "patterns without a root";
	}
	if (
		!m.root.startsWith("/") || !(m.root === "/" || m.root.endsWith("/")) || /[{}]/.test(m.root)
	) {
		return `root ${JSON.stringify(m.root)}`;
	}
	for (const p of patterns) {
		if (!p.startsWith("/") || !validPattern(m.root.slice(0, -1) + p)) {
			return `pattern ${JSON.stringify(p)}`;
		}
	}
	return undefined;
}

const wildcardName = /^[\p{L}_][\p{L}\p{Nd}_]*$/u;

// validPattern reports whether s is a path pattern as Go's parsePathPattern
// reads it: a path whose wildcards are whole segments, "{name}" anywhere,
// "{name...}" and "{$}" only as the last segment, and each name used once.
function validPattern(s: string): boolean {
	if (!s.startsWith("/")) return false;
	if (!/[{}]/.test(s)) return true;
	const trailing = s.endsWith("/");
	const segs = (trailing ? s.slice(1, -1) : s.slice(1)).split("/");
	const names = new Set<string>();
	for (let i = 0; i < segs.length; i++) {
		const seg = segs[i] ?? "";
		if (!/[{}]/.test(seg)) continue;
		const last = i === segs.length - 1 && !trailing;
		if (!seg.startsWith("{") || !seg.endsWith("}")) return false;
		const inner = seg.slice(1, -1);
		if (inner === "$") {
			if (!last) return false;
			continue;
		}
		const rest = inner.endsWith("...");
		const name = rest ? inner.slice(0, -3) : inner;
		if (rest && !last) return false;
		if (!wildcardName.test(name) || names.has(name)) return false;
		names.add(name);
	}
	return true;
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
