import { assertEquals, assertInstanceOf } from "@std/assert";
import { Buffer } from "@okdaichi/golikejs/bytes";
import { EOFError } from "@okdaichi/golikejs/io";
import { AuthViolation, readAuthReply } from "./auth.ts";
import { AuthError } from "./error.ts";
import { AuthErrorMessage, AuthOkMessage } from "./internal/message/mod.ts";
import { MockReceiveStream, MockStream } from "./mock_stream_test.ts";

// replyStream returns a stream whose reads yield bytes, then EOF.
function replyStream(bytes: Uint8Array): MockStream {
	let offset = 0;
	return new MockStream({
		readable: new MockReceiveStream({
			read: (p) => {
				if (offset >= bytes.length) return Promise.resolve([0, new EOFError()]);
				const n = Math.min(p.length, bytes.length - offset);
				p.set(bytes.subarray(offset, offset + n));
				offset += n;
				return Promise.resolve([n, undefined]);
			},
		}),
	});
}

async function encoded(m: { encode(w: Buffer): Promise<Error | undefined> }): Promise<Uint8Array> {
	const buf = Buffer.make(256);
	assertEquals(await m.encode(buf), undefined);
	return buf.bytes();
}

Deno.test("readAuthReply - accepts a valid grant", async (t) => {
	const cases = {
		"prefix": ["/"],
		"one wildcard": ["/{user}/audio"],
		"rest wildcard": ["/{rest...}"],
		"anchor": ["/{$}"],
		"two names": ["/{a}/{b}"],
	};
	for (const [name, patterns] of Object.entries(cases)) {
		await t.step(name, async () => {
			const bytes = await encoded(
				new AuthOkMessage({ root: "/room/1/", subscribePatterns: patterns }),
			);

			const [grant, err] = await readAuthReply(replyStream(bytes));

			assertEquals(err, undefined);
			assertEquals(grant?.subscribe, patterns);
		});
	}
});

Deno.test("readAuthReply - a malformed answer is a violation", async (t) => {
	const cases: Record<string, () => Promise<Uint8Array>> = {
		"patterns without a root": () => encoded(new AuthOkMessage({ subscribePatterns: ["/"] })),
		"root that is a path": () => encoded(new AuthOkMessage({ root: "/room/1" })),
		"root with a wildcard": () => encoded(new AuthOkMessage({ root: "/room/{id}/" })),
		"pattern without a slash": () =>
			encoded(new AuthOkMessage({ root: "/r/", subscribePatterns: ["x"] })),
		"open wildcard": () =>
			encoded(new AuthOkMessage({ root: "/r/", subscribePatterns: ["/{user"] })),
		"rest not last": () =>
			encoded(new AuthOkMessage({ root: "/r/", subscribePatterns: ["/{a...}/b"] })),
		"name used twice": () =>
			encoded(new AuthOkMessage({ root: "/r/", subscribePatterns: ["/{a}/{a}"] })),
		"name starting with a digit": () =>
			encoded(new AuthOkMessage({ root: "/r/", subscribePatterns: ["/{1a}"] })),
		"wildcard within a segment": () =>
			encoded(new AuthOkMessage({ root: "/r/", subscribePatterns: ["/x{a}"] })),
		// Type OK, a body of 0 bytes: no root length.
		"truncated AUTH_OK": () => Promise.resolve(new Uint8Array([0, 0])),
		// Type OK, root "", then a pattern count far past the body.
		"pattern count past the body": () =>
			Promise.resolve(new Uint8Array([0, 5, 0, 0xc0, 0, 0, 0])),
		"trailing bytes": () => Promise.resolve(new Uint8Array([0, 5, 0, 0, 0, 0, 0])),
		"reason too long": () =>
			encoded(new AuthErrorMessage({ code: 2, reason: "a".repeat(8193) })),
		"reason not UTF-8": () => Promise.resolve(new Uint8Array([1, 3, 2, 1, 0xff])),
		"unknown reply type": () => Promise.resolve(new Uint8Array([7])),
		"expires past 2^53 - 1": () =>
			encoded(new AuthOkMessage({ root: "/r/", expires: 2 ** 53 })),
		"code over 32 bits": () => encoded(new AuthErrorMessage({ code: 2 ** 33, reason: "" })),
	};
	for (const [name, bytes] of Object.entries(cases)) {
		await t.step(name, async () => {
			const [grant, err] = await readAuthReply(replyStream(await bytes()));

			assertEquals(grant, undefined);
			assertInstanceOf(err, AuthViolation);
		});
	}
});

Deno.test("readAuthReply - AUTH_ERROR is an AuthError", async () => {
	const bytes = await encoded(new AuthErrorMessage({ code: 2, reason: "unknown ticket" }));

	const [, err] = await readAuthReply(replyStream(bytes));

	assertInstanceOf(err, AuthError);
	assertEquals(err.reason, "unknown ticket");
});
