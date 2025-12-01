import { assertEquals, assertExists } from "@std/assert";
import { Extensions } from "./extensions.ts";

Deno.test("internal/extensions - basic operations", async (t) => {
  await t.step("constructor creates empty map", () => {
    const e = new Extensions();
    assertExists(e);
    // entries is a Map
    assertEquals(e.entries instanceof Map, true);
    assertEquals(e.entries.size, 0);
  });

  await t.step("addString / getString / has / delete", () => {
    const e = new Extensions();
    assertEquals(e.has(1), false);

    e.addString(1, "hello");
    assertEquals(e.has(1), true);
    assertEquals(e.getString(1), "hello");

    // overwrite
    e.addString(1, "world");
    assertEquals(e.getString(1), "world");

    // delete
    const deleted = e.delete(1);
    assertEquals(deleted, true);
    assertEquals(e.has(1), false);
    assertEquals(e.getString(1), undefined);
  });

  await t.step("addBytes / getBytes (empty, normal, large)", () => {
    const e = new Extensions();
    const empty = new Uint8Array([]);
    e.addBytes(2, empty);
    assertEquals(e.getBytes(2), empty);

    const data = new Uint8Array([1, 2, 3, 4, 5]);
    e.addBytes(3, data);
    assertEquals(e.getBytes(3), data);

    const large = new Uint8Array(1000).fill(42);
    e.addBytes(4, large);
    assertEquals(e.getBytes(4)?.length, 1000);
  });

  await t.step("addNumber / getNumber and byte-size checks", () => {
    const e = new Extensions();
    const big = 12345678901234567890n;
    e.addNumber(10, big);
    assertEquals(e.getNumber(10), big);

    e.addNumber(11, 42n);
    const bytes = e.getBytes(11);
    // numbers are stored as 8 bytes
    assertEquals(bytes?.length, 8);

    // incorrectly sized bytes should result in undefined number
    e.addBytes(12, new Uint8Array([1, 2, 3]));
    assertEquals(e.getNumber(12), undefined);
  });

  await t.step("addBoolean / getBoolean", () => {
    const e = new Extensions();
    e.addBoolean(20, true);
    assertEquals(e.getBoolean(20), true);

    e.addBoolean(21, false);
    assertEquals(e.getBoolean(21), false);

    // storage as single byte
    const b = e.getBytes(20);
    assertEquals(b?.length, 1);
    assertEquals(b?.[0], 1);
  });

  await t.step("mixed operations and scaling", () => {
    const e = new Extensions();
    e.addString(1, "A");
    e.addNumber(2, 2n);
    e.addBoolean(3, true);
    e.addBytes(4, new Uint8Array([9]));

    assertEquals(e.getString(1), "A");
    assertEquals(e.getNumber(2), 2n);
    assertEquals(e.getBoolean(3), true);
    assertEquals(e.getBytes(4), new Uint8Array([9]));

    // many entries
    for (let i = 0; i < 50; i++) e.addString(i + 100, `val${i}`);
    for (let i = 0; i < 50; i++) assertEquals(e.getString(i + 100), `val${i}`);
  });

  await t.step("constructor with initial entries", () => {
    const initial = new Map<number, Uint8Array>();
    initial.set(1, new Uint8Array([1, 2, 3]));
    const e = new Extensions(initial);
    assertEquals(e.getBytes(1), new Uint8Array([1, 2, 3]));
    assertEquals(e.entries.size, 1);
  });

  await t.step("getBoolean with invalid data", () => {
    const e = new Extensions();
    // length !== 1
    e.addBytes(30, new Uint8Array([1, 2]));
    assertEquals(e.getBoolean(30), undefined);

    // bytes[0] !== 0 and !== 1
    e.addBytes(31, new Uint8Array([2]));
    assertEquals(e.getBoolean(31), false); // since 2 !== 1, returns false
  });
});
