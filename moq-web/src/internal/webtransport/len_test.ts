import { assertEquals, assertThrows } from "@std/assert";
import {
	bytesLen,
	MAX_VARINT1,
	MAX_VARINT2,
	MAX_VARINT4,
	MAX_VARINT8,
	stringLen,
	varintLen,
} from "./len.ts";

Deno.test("varintLen", async (t) => {
	await t.step("varintLen <= MAX_VARINT1", () => {
		assertEquals(varintLen(0), 1);
		assertEquals(varintLen(63), 1);
		assertEquals(varintLen(MAX_VARINT1), 1);
	});

	await t.step("varintLen <= MAX_VARINT2", () => {
		assertEquals(varintLen(64), 2);
		assertEquals(varintLen(16383), 2);
		assertEquals(varintLen(MAX_VARINT2), 2);
	});

	await t.step("varintLen <= MAX_VARINT4", () => {
		assertEquals(varintLen(16384), 4);
		assertEquals(varintLen(1073741823), 4);
		assertEquals(varintLen(MAX_VARINT4), 4);
	});

	await t.step("varintLen <= MAX_VARINT8", () => {
		assertEquals(varintLen(1073741824), 8);
		assertEquals(varintLen(BigInt("4611686018427387903")), 8);
		assertEquals(varintLen(MAX_VARINT8), 8);
	});

	await t.step("varintLen negative numbers", () => {
		assertThrows(() => varintLen(-1), RangeError);
	});

	await t.step("varintLen > MAX_VARINT8 throws", () => {
		assertThrows(() => varintLen(BigInt("4611686018427387904")), RangeError);
	});
});

Deno.test("stringLen", async (t) => {
	await t.step("empty string", () => {
		assertEquals(stringLen(""), 1);
	});

	await t.step("single character", () => {
		assertEquals(stringLen("a"), 2);
	});

	await t.step("multi-character string", () => {
		const str = "hello world";
		assertEquals(stringLen(str), 1 + str.length);
	});
});

Deno.test("bytesLen", async (t) => {
	await t.step("empty bytes", () => {
		assertEquals(bytesLen(new Uint8Array(0)), 1);
	});

	await t.step("non-empty bytes", () => {
		const bytes = new Uint8Array([1, 2, 3]);
		assertEquals(bytesLen(bytes), 1 + 3);
	});
});
