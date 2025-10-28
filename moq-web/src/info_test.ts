import { describe, it, assertEquals, assertExists } from "../deps.ts";
import type { Info } from "./info.ts";

describe("Info", () => {
  it("should be defined as a type", () => {
    // This test ensures the Info interface is properly exported and can be used
    const info: Info = {};

    // Since Info is currently an empty interface, we can only verify it exists
    assertEquals(typeof info, "object");
  });

  it("should allow creating empty Info objects", () => {
    // Test that we can create an empty Info object since it's currently an empty interface
    const info: Info = {};

    assertExists(info);
    assertEquals(info, {});
  });

  it("should be assignable to object type", () => {
    // Verify that Info objects are valid objects
    const info: Info = {};
    const obj: object = info;

    assertEquals(obj, info);
  });
});
