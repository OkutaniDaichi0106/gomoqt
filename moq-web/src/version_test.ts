import { assertEquals } from "@std/assert";
import { DEFAULT_CLIENT_VERSIONS, Versions } from "./version.ts";

Deno.test("Version constants are defined", () => {
  assertEquals(Versions.LITE_DRAFT_01, 0xff0dad01);
  assertEquals(Versions.LITE_DRAFT_02, 0xff0dad02);
  assertEquals(DEFAULT_CLIENT_VERSIONS instanceof Set, true);
  assertEquals(DEFAULT_CLIENT_VERSIONS.has(Versions.LITE_DRAFT_01), true);
});
