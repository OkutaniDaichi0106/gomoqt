import { assertEquals, assertExists, assertArrayIncludes } from "../../deps.ts";
import * as index from './index.ts';

Deno.test("internal/index - re-exports and module shape", async (t) => {
  await t.step("exports include Extensions and Queue", () => {
    const keys = Object.keys(index);
    assertArrayIncludes(keys, ['Extensions', 'Queue']);
    assertExists(index.Extensions);
    assertExists(index.Queue);
  });

  await t.step("module is an object and not empty", () => {
    assertExists(index);
    assertEquals(typeof index, 'object');
    const count = Object.keys(index).length;
    // Expect at least 2 exports
    assertEquals(count >= 2, true);
  });

  await t.step("Queue is constructible", () => {
    const Q = index.Queue as unknown as { new(): unknown };
    // construct to ensure the export is a class/constructor
    const q = new Q();
    assertExists(q);
  });
});
