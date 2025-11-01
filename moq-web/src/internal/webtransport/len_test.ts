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

Deno.test("len utilities", async (t) => {
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

	await t.step("stringLen", () => {
		assertEquals(stringLen(""), 1);
		assertEquals(stringLen("a"), 2);
		const str = "hello world";
		assertEquals(stringLen(str), 1 + str.length);
	});

	await t.step("bytesLen", () => {
		assertEquals(bytesLen(new Uint8Array(0)), 1);
		const bytes = new Uint8Array([1, 2, 3]);
		assertEquals(bytesLen(bytes), 1 + 3);
	});
});
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

Deno.test("webtransport/len - varintLen behavior", async (t) => {
	const cases: Record<string, { input: number | bigint; expected?: number; throws?: boolean }> = {
		"small 0": { input: 0, expected: 1 },
		"small 63": { input: 63, expected: 1 },
		"boundary1": { input: MAX_VARINT1, expected: 1 },
		"start2": { input: 64, expected: 2 },
		"boundary2": { input: MAX_VARINT2, expected: 2 },
		"start4": { input: 16384, expected: 4 },
		"boundary4": { input: MAX_VARINT4, expected: 4 },
		"start8": { input: 1073741824, expected: 8 },
		"boundary8": { input: MAX_VARINT8, expected: 8 },
		"negative": { input: -1, throws: true },
		"tooLarge": { input: BigInt("4611686018427387904"), throws: true },
	};

	for (const [name, c] of Object.entries(cases)) {
		await t.step(name, () => {
			if (c.throws) {
				assertThrows(() => {
					varintLen(c.input as unknown as number | bigint);
				}, RangeError);
			} else {
				assertEquals(varintLen(c.input as unknown as number | bigint), c.expected);
			}
		});
	}
});

Deno.test("webtransport/len - stringLen and bytesLen", async (t) => {
	await t.step("empty string", () => {
		assertEquals(stringLen(""), 1);
	});

	await t.step("short string", () => {
		assertEquals(stringLen("a"), 2);
	});

	await t.step("bytes length", () => {
		const bytes = new Uint8Array([1, 2, 3]);
		assertEquals(bytesLen(bytes), 1 + 3);
	});
});
