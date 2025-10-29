import { assertEquals, assertThrows } from "../../deps.ts";
import { MAX_VARINT1, MAX_VARINT2, MAX_VARINT4, MAX_VARINT8, varintLen, stringLen, bytesLen } from "./len.ts";

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
        assertThrows(() => { varintLen(c.input as unknown as number | bigint); }, RangeError);
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
    const bytes = new Uint8Array([1,2,3]);
    assertEquals(bytesLen(bytes), 1 + 3);
  });
});
