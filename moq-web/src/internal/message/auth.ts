import type { Reader, Writer } from "@okdaichi/golikejs/io";
import { MessageEncoder, parseVarint, putVarint, readFull, readVarint } from "./message.ts";

/**
 * Reads the fields of a message body, throwing a RangeError for a field that
 * runs past the body rather than reading zeros, as MessageDecoder does.
 */
class BodyReader {
	readonly #buf: Uint8Array;
	#offset = 0;

	constructor(buf: Uint8Array) {
		this.#buf = buf;
	}

	#need(n: number): void {
		if (this.#offset + n > this.#buf.length) {
			throw new RangeError("field runs past the message");
		}
	}

	varint(): number {
		this.#need(1);
		this.#need(1 << ((this.#buf[this.#offset] ?? 0) >> 6));
		const [v, n] = parseVarint(this.#buf, this.#offset);
		this.#offset += n;
		return v;
	}

	bytes(): Uint8Array {
		const n = this.varint();
		this.#need(n);
		const out = this.#buf.slice(this.#offset, this.#offset + n);
		this.#offset += n;
		return out;
	}

	string(): string {
		return new TextDecoder().decode(this.bytes());
	}

	stringArray(): string[] {
		const count = this.varint();
		// Each string takes at least its length byte.
		this.#need(count);
		const out: string[] = [];
		for (let i = 0; i < count; i++) {
			out.push(this.string());
		}
		return out;
	}

	eof(): boolean {
		return this.#offset >= this.#buf.length;
	}
}

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

// The largest message body read, as Go's message.MaxMessageSize.
const maxMessageSize = 64 * 1024;

async function readBody(r: Reader): Promise<[Uint8Array, undefined] | [undefined, Error]> {
	const [len, , err1] = await readVarint(r);
	if (err1) return [undefined, err1];
	if (len > maxMessageSize) {
		return [undefined, new Error(`moq: message of ${len} bytes is too large`)];
	}
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
		try {
			const d = new BodyReader(buf);
			this.token = d.bytes();
			if (!d.eof()) return new Error("moq: AUTH longer than its fields");
		} catch (e) {
			return e instanceof Error ? e : new Error(String(e));
		}
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
		try {
			const d = new BodyReader(buf);
			this.root = d.string();
			this.publishPatterns = d.stringArray();
			this.subscribePatterns = d.stringArray();
			this.expires = d.varint();
			if (!d.eof()) return new Error("moq: AUTH_OK longer than its fields");
		} catch (e) {
			return e instanceof Error ? e : new Error(String(e));
		}
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
		try {
			const d = new BodyReader(buf);
			this.code = d.varint();
			this.reason = d.bytes();
			if (!d.eof()) return new Error("moq: AUTH_ERROR longer than its fields");
		} catch (e) {
			return e instanceof Error ? e : new Error(String(e));
		}
		return undefined;
	}
}
