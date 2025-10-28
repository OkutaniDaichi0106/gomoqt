import { describe, it, assertEquals, assertThrows } from "../deps.ts";
import { isValidPrefix, validateTrackPrefix } from "./track_prefix.ts";

describe("TrackPrefix", () => {
  describe("isValidPrefix", () => {
    it("should return true for valid prefixes", () => {
      assertEquals(isValidPrefix("/foo/"), true);
      assertEquals(isValidPrefix("//"), true);
      assertEquals(isValidPrefix("/"), true);
    });

    it("should return false for invalid prefixes", () => {
      assertEquals(isValidPrefix("foo/"), false);
      assertEquals(isValidPrefix("/foo"), false);
      assertEquals(isValidPrefix("foo"), false);
      assertEquals(isValidPrefix(""), false);
    });
  });

  describe("validateTrackPrefix", () => {
    it("should return the prefix for valid prefixes", () => {
      assertEquals(validateTrackPrefix("/foo/"), "/foo/");
      assertEquals(validateTrackPrefix("//"), "//");
      assertEquals(validateTrackPrefix("/"), "/");
    });

    it("should throw an error for invalid prefixes", () => {
      assertThrows(() => validateTrackPrefix("foo/"));
      assertThrows(() => validateTrackPrefix("/foo"));
      assertThrows(() => validateTrackPrefix("foo"));
      assertThrows(() => validateTrackPrefix(""));
    });
  });
});
