import { assertEquals, assertExists, assertThrows } from "@std/assert";
import type { BroadcastPath } from "./broadcast_path.ts";
import { extension, isValidBroadcastPath, validateBroadcastPath } from "./broadcast_path.ts";

Deno.test("BroadcastPath - isValidBroadcastPath", async (t) => {
	await t.step("returns true for valid paths", () => {
		assertEquals(isValidBroadcastPath("/"), true);
		assertEquals(isValidBroadcastPath("/test"), true);
		assertEquals(isValidBroadcastPath("/test/path"), true);
		assertEquals(isValidBroadcastPath("/alice.json"), true);
		assertEquals(isValidBroadcastPath("/video/stream"), true);
		assertEquals(isValidBroadcastPath("/path/with/multiple/segments"), true);
	});

	await t.step("returns false for invalid paths", () => {
		assertEquals(isValidBroadcastPath(""), false);
		assertEquals(isValidBroadcastPath("test"), false);
		assertEquals(isValidBroadcastPath("test/path"), false);
		assertEquals(isValidBroadcastPath("alice.json"), false);
	});
});

Deno.test("BroadcastPath - validateBroadcastPath", async (t) => {
	await t.step("returns path for valid paths", () => {
		assertEquals(validateBroadcastPath("/"), "/");
		assertEquals(validateBroadcastPath("/test"), "/test");
		assertEquals(validateBroadcastPath("/test/path"), "/test/path");
		assertEquals(validateBroadcastPath("/alice.json"), "/alice.json");
	});

	await t.step("throws error for invalid paths", () => {
		assertThrows(
			() => validateBroadcastPath(""),
			Error,
			'Invalid broadcast path: "". Must start with "/"',
		);
		assertThrows(
			() => validateBroadcastPath("test"),
			Error,
			'Invalid broadcast path: "test". Must start with "/"',
		);
		assertThrows(
			() => validateBroadcastPath("test/path"),
			Error,
			'Invalid broadcast path: "test/path". Must start with "/"',
		);
		assertThrows(
			() => validateBroadcastPath("alice.json"),
			Error,
			'Invalid broadcast path: "alice.json". Must start with "/"',
		);
	});
});

Deno.test("BroadcastPath - extension", async (t) => {
	await t.step("returns correct extension for paths with extensions", () => {
		assertEquals(extension("/alice.json" as BroadcastPath), ".json");
		assertEquals(extension("/video/stream.mp4" as BroadcastPath), ".mp4");
		assertEquals(extension("/file.min.js" as BroadcastPath), ".js");
		assertEquals(extension("/test/path.backup.mp4" as BroadcastPath), ".mp4");
		assertEquals(extension("/test/.hidden.txt" as BroadcastPath), ".txt");
		assertEquals(extension("/test/path." as BroadcastPath), ".");
		assertEquals(extension("file.txt" as BroadcastPath), ".txt");
	});

	await t.step("returns empty string for paths without extensions", () => {
		assertEquals(extension("/test/path" as BroadcastPath), "");
		assertEquals(extension("/video/stream" as BroadcastPath), "");
		assertEquals(extension("/" as BroadcastPath), "");
		assertEquals(extension("" as BroadcastPath), "");
	});

	await t.step("handles edge cases correctly", () => {
		assertEquals(extension("/test.dir/file" as BroadcastPath), "");
		assertEquals(extension("/test.dir/file.ext" as BroadcastPath), ".ext");
		assertEquals(extension("/.hidden" as BroadcastPath), "");
	});
});

Deno.test("BroadcastPath - type safety and utilities", async (t) => {
	await t.step("BroadcastPath can be used as string and validated", () => {
		const path: BroadcastPath = validateBroadcastPath("/test/path");
		assertEquals(typeof path, "string");
		// basic runtime checks
		assertExists(path.startsWith);
		assertEquals(path.startsWith("/"), true);
		assertEquals(path.length > 0, true);
	});

	await t.step("validateBroadcastPath throws on invalid and returns on valid", () => {
		assertThrows(() => validateBroadcastPath("no-slash"), Error);
		assertEquals(validateBroadcastPath("/ok"), "/ok");
	});

	await t.step("extension extraction cases", () => {
		assertEquals(extension("/alice.hang"), ".hang");
		assertEquals(extension("/path/to/file"), "");
		assertEquals(extension("/.hidden"), "");
		assertEquals(extension("/dir.name/file.txt"), ".txt");
	});
});
