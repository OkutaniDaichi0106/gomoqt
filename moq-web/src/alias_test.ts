import { assertExists } from "@std/assert";
import * as alias from "./alias.ts";

Deno.test("alias module loads", () => {
  assertExists(alias);
});
