import { assertEquals } from "@std/assert";
import { DEFAULT_CLIENT_VERSIONS, Versions } from "./version.ts";

Deno.test("Version constants are defined", () => {
  assertEquals(Versions.DEVELOP, Versions.DEVELOP);
  assertEquals(DEFAULT_CLIENT_VERSIONS instanceof Set, true);
  assertEquals(DEFAULT_CLIENT_VERSIONS.has(Versions.DEVELOP), true);
});
