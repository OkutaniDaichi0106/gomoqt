import type { Reader, Writer } from "@okdaichi/golikejs/io";
import { MessageDecoder, MessageEncoder, putVarint, readFull, readVarint } from "./message.ts";

/** The Type that leads each message the acceptor sends on an Auth Stream. */
export const AuthReplyTypes = {
	/** AUTH_OK: the token's grant. */
	Ok: 0x0,
	/** AUTH_ERROR: the token is refused, or its grant revoked. */
	Error: 0x1,
} as const;

/** Reads the Type of the next acceptor message on an Auth Stream. */
export async function readAuthReplyType(
	r: Reader,
): Promise<[number, undefined] | [undefined, Error]> {
	const [t, , err] = await readVarint(r);
	if (err) return [undefined, err];
	return [t, undefined];
}

async function readBody(r: Reader): Promise<[Uint8Array, undefined] | [undefined, Error]> {
	const [len, , err1] = await readVarint(r);
	if (err1) return [undefined, err1];
	const buf = new Uint8Array(len);
	const [, err2] = await readFull(r, buf);
	if (err2) return [undefined, err2];
	return [buf, undefined];
}

async function writeTyped(w: Writer, type: number, frame: Uint8Array): Promise<Error | undefined> {
	const out = new Uint8Array(8 + frame.length);
	const n = putVarint(out, 0, type);
	out.set(frame, n);
	const [, err] = await w.write(out.subarray(0, n + frame.length));
	return err;
}

/**
 * AUTH message, the presenter's token.
 *
 * ```text
 * AUTH Message {
 *   Message Length (i)
 *   Token (b)
 * }
 * ```
 */
export class AuthMessage {
	token: Uint8Array;

	constructor(init: { token?: Uint8Array } = {}) {
		this.token = init.token ?? new Uint8Array(0);
	}

	async encode(w: Writer): Promise<Error | undefined> {
		const e = new MessageEncoder();
		e.bytes(this.token);
		const [, err] = await w.write(e.frame());
		return err;
	}

	async decode(r: Reader): Promise<Error | undefined> {
		const [buf, err] = await readBody(r);
		if (err) return err;
		const d = new MessageDecoder(buf);
		this.token = d.bytes(d.varint()).slice();
		return undefined;
	}
}

/**
 * AUTH_OK message, the token's grant.
 *
 * ```text
 * AUTH_OK Message {
 *   Type (i) = 0x0
 *   Message Length (i)
 *   Root (s)
 *   Publish Count (i)
 *   Publish Pattern (s) ...
 *   Subscribe Count (i)
 *   Subscribe Pattern (s) ...
 *   Expires (i)
 * }
 * ```
 */
export class AuthOkMessage {
	root: string;
	publishPatterns: string[];
	subscribePatterns: string[];
	/** Milliseconds until the grant lapses, or 0 for never. */
	expires: number;

	constructor(
		init: {
			root?: string;
			publishPatterns?: string[];
			subscribePatterns?: string[];
			expires?: number;
		} = {},
	) {
		this.root = init.root ?? "";
		this.publishPatterns = init.publishPatterns ?? [];
		this.subscribePatterns = init.subscribePatterns ?? [];
		this.expires = init.expires ?? 0;
	}

	/** Writes the message with its leading Type. */
	encode(w: Writer): Promise<Error | undefined> {
		const e = new MessageEncoder();
		e.string(this.root);
		e.stringArray(this.publishPatterns);
		e.stringArray(this.subscribePatterns);
		e.varint(this.expires);
		return writeTyped(w, AuthReplyTypes.Ok, e.frame());
	}

	/** Reads the message after its Type (see {@link readAuthReplyType}). */
	async decode(r: Reader): Promise<Error | undefined> {
		const [buf, err] = await readBody(r);
		if (err) return err;
		const d = new MessageDecoder(buf);
		this.root = d.string();
		this.publishPatterns = d.stringArray();
		this.subscribePatterns = d.stringArray();
		this.expires = d.varint();
		if (!d.eof()) return new Error("moq: AUTH_OK longer than its fields");
		return undefined;
	}
}

/**
 * AUTH_ERROR message: the token is refused, or its grant revoked.
 *
 * ```text
 * AUTH_ERROR Message {
 *   Type (i) = 0x1
 *   Message Length (i)
 *   Error Code (i)
 *   Reason Phrase (s)
 * }
 * ```
 */
export class AuthErrorMessage {
	code: number;
	/** The reason's bytes, which a reader checks are UTF-8. */
	reason: Uint8Array;

	constructor(init: { code?: number; reason?: string } = {}) {
		this.code = init.code ?? 0;
		this.reason = new TextEncoder().encode(init.reason ?? "");
	}

	/** Writes the message with its leading Type. */
	encode(w: Writer): Promise<Error | undefined> {
		const e = new MessageEncoder();
		e.varint(this.code);
		e.bytes(this.reason);
		return writeTyped(w, AuthReplyTypes.Error, e.frame());
	}

	/** Reads the message after its Type (see {@link readAuthReplyType}). */
	async decode(r: Reader): Promise<Error | undefined> {
		const [buf, err] = await readBody(r);
		if (err) return err;
		const d = new MessageDecoder(buf);
		this.code = d.varint();
		this.reason = d.bytes(d.varint()).slice();
		if (!d.eof()) return new Error("moq: AUTH_ERROR longer than its fields");
		return undefined;
	}
}
