import { assertEquals, assertInstanceOf } from "@std/assert";
import { Buffer } from "@okdaichi/golikejs/bytes";
import {
	AuthErrorMessage,
	AuthMessage,
	AuthOkMessage,
	AuthReplyTypes,
	readAuthReplyType,
} from "./auth.ts";

// Bytes Go's moqt/internal/message encodes for the same messages, so the two
// stay on one wire format.
const goAuthOk = new Uint8Array([
	0,
	38,
	8,
	47,
	114,
	111,
	111,
	109,
	47,
	49,
	47,
	1,
	6,
	47,
	97,
	108,
	105,
	99,
	101,
	2,
	1,
	47,
	13,
	47,
	123,
	117,
	115,
	101,
	114,
	125,
	47,
	97,
	117,
	100,
	105,
	111,
	128,
	0,
	234,
	96,
]);
const goAuthError = new Uint8Array([
	1,
	16,
	2,
	14,
	117,
	110,
	107,
	110,
	111,
	119,
	110,
	32,
	116,
	105,
	99,
	107,
	101,
	116,
]);
const goAuth = new Uint8Array([6, 5, 114, 111, 111, 109, 49]);

async function encoded(m: { encode(w: Buffer): Promise<Error | undefined> }): Promise<Uint8Array> {
	const buf = Buffer.make(128);
	assertEquals(await m.encode(buf), undefined);
	return buf.bytes();
}

function reader(bytes: Uint8Array): Buffer {
	const buf = Buffer.make(bytes.length);
	buf.write(bytes);
	return buf;
}

Deno.test("AuthOkMessage - matches Go's wire format", async (t) => {
	await t.step("encode", async () => {
		const m = new AuthOkMessage({
			root: "/room/1/",
			publishPatterns: ["/alice"],
			subscribePatterns: ["/", "/{user}/audio"],
			expires: 60000,
		});

		assertEquals(await encoded(m), goAuthOk);
	});

	await t.step("decode", async () => {
		const r = reader(goAuthOk);
		const [type, err] = await readAuthReplyType(r);
		assertEquals(err, undefined);
		assertEquals(type, AuthReplyTypes.Ok);
		const m = new AuthOkMessage();

		assertEquals(await m.decode(r), undefined);
		assertEquals(m.root, "/room/1/");
		assertEquals(m.publishPatterns, ["/alice"]);
		assertEquals(m.subscribePatterns, ["/", "/{user}/audio"]);
		assertEquals(m.expires, 60000);
	});
});

Deno.test("AuthErrorMessage - matches Go's wire format", async (t) => {
	await t.step("encode", async () => {
		const m = new AuthErrorMessage({ code: 0x2, reason: "unknown ticket" });

		assertEquals(await encoded(m), goAuthError);
	});

	await t.step("decode", async () => {
		const r = reader(goAuthError);
		const [type] = await readAuthReplyType(r);
		assertEquals(type, AuthReplyTypes.Error);
		const m = new AuthErrorMessage();

		assertEquals(await m.decode(r), undefined);
		assertEquals(m.code, 0x2);
		assertEquals(new TextDecoder().decode(m.reason), "unknown ticket");
	});
});

Deno.test("AuthMessage - matches Go's wire format", async (t) => {
	await t.step("encode", async () => {
		const m = new AuthMessage({ token: new TextEncoder().encode("room1") });

		assertEquals(await encoded(m), goAuth);
	});

	await t.step("decode", async () => {
		const m = new AuthMessage();

		assertEquals(await m.decode(reader(goAuth)), undefined);
		assertEquals(new TextDecoder().decode(m.token), "room1");
	});
});

Deno.test("AuthOkMessage - decode refuses trailing bytes", async () => {
	// Root "", no patterns, Expires 0, then one byte more than the fields.
	const r = reader(new Uint8Array([5, 0, 0, 0, 0, 0]));
	const m = new AuthOkMessage();

	assertInstanceOf(await m.decode(r), Error);
});
