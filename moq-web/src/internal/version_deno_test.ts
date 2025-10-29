import { assertEquals, assertExists } from "../../deps.ts";
import type { Version } from './version.ts';
import { Versions, DEFAULT_VERSION } from './version.ts';

Deno.test("internal/version - types and constants", async (t) => {
  await t.step("Version is bigint and accepts values", () => {
    const v: Version = 123n;
    assertEquals(typeof v, 'bigint');
    const zero: Version = 0n;
    const one: Version = 1n;
    assertEquals(zero, 0n);
    assertEquals(one, 1n);
  });

  await t.step("Versions constants are defined and correct", () => {
    assertExists(Versions.DEVELOP);
    assertEquals(typeof Versions.DEVELOP, 'bigint');
    assertEquals(Versions.DEVELOP, 0xffffff00n);
    assertEquals(DEFAULT_VERSION, Versions.DEVELOP);
  });

  await t.step("arithmetic and comparison work", () => {
    const a: Version = 1n;
    const b: Version = 2n;
    assertEquals(a < b, true);
    assertEquals(b > a, true);
    assertEquals((a + 1n), 2n);
  });

  await t.step("string/hex conversions", () => {
    assertEquals(DEFAULT_VERSION.toString(16), 'ffffff00');
    const parsed = BigInt('4294967040');
    assertEquals(parsed, DEFAULT_VERSION);
  });
});
