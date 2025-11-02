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
		"zero": { input: 0, expected: 1 },
		"63": { input: 63, expected: 1 },
		"MAX_VARINT1": { input: MAX_VARINT1, expected: 1 },
		"64": { input: 64, expected: 2 },
		"16383": { input: 16383, expected: 2 },
		"MAX_VARINT2": { input: MAX_VARINT2, expected: 2 },
		"16384": { input: 16384, expected: 4 },
		"1073741823": { input: 1073741823, expected: 4 },
		"MAX_VARINT4": { input: MAX_VARINT4, expected: 4 },
		"1073741824": { input: 1073741824, expected: 8 },
		"large bigint": { input: BigInt("4611686018427387903"), expected: 8 },
		"MAX_VARINT8": { input: MAX_VARINT8, expected: 8 },
		"negative": { input: -1, throws: true },
		"too large": { input: BigInt("4611686018427387904"), throws: true },
	};

	for (const [name, c] of Object.entries(cases)) {
		await t.step(name, () => {
			if (c.throws) {
				assertThrows(() => {
					varintLen(c.input);
				}, RangeError);
			} else {
				assertEquals(varintLen(c.input), c.expected);
			}
		});
	}
});

Deno.test("webtransport/len - stringLen and bytesLen", async (t) => {
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

	await t.step("empty bytes", () => {
		assertEquals(bytesLen(new Uint8Array(0)), 1);
	});

	await t.step("non-empty bytes", () => {
		const bytes = new Uint8Array([1, 2, 3]);
		assertEquals(bytesLen(bytes), 1 + 3);
	});
});
