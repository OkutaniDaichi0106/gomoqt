import { assertEquals, assertExists, assertInstanceOf, assertThrows } from "@std/assert";
import type { StreamErrorCode } from "./error.ts";
import { StreamError } from "./error.ts";

Deno.test("StreamError - constructor - should create StreamError with required parameters", () => {
	const code: StreamErrorCode = 404;
	const message = "Not found";
	const error = new StreamError(code, message);
	assertEquals(error.code, code);
	assertEquals(error.message, message);
	assertEquals(error.remote, false);
	// name for Error is usually 'Error'
	assertEquals(error.name, "Error");
	assertInstanceOf(error, Error);
	assertInstanceOf(error, StreamError);
});

Deno.test("StreamError - constructor - should create StreamError with remote flag", () => {
	const code: StreamErrorCode = 500;
	const message = "Internal server error";
	const remote = true;
	const error = new StreamError(code, message, remote);
	assertEquals(error.code, code);
	assertEquals(error.message, message);
	assertEquals(error.remote, remote);
});

Deno.test("StreamError - constructor - should default remote to false when not specified", () => {
	const error = new StreamError(200, "OK");
	assertEquals(error.remote, false);
});

Deno.test("StreamError - prototype chain - should maintain proper prototype chain", () => {
	const error = new StreamError(123, "Test error");
	assertInstanceOf(error, StreamError);
	assertInstanceOf(error, Error);
	assertEquals(Object.getPrototypeOf(error), StreamError.prototype);
});

Deno.test("StreamError - prototype chain - instanceof after JSON serialization", () => {
	const original = new StreamError(456, "Original error", true);
	const recreated = Object.create(StreamError.prototype);
	Object.assign(recreated, original as any);
	assertInstanceOf(recreated, StreamError);
	assertEquals((recreated as any).code, 456);
	// Error.message is not enumerable by default
	assertEquals((recreated as any).message, "");
	assertEquals((recreated as any).remote, true);
});

Deno.test("StreamError - error codes - various codes", () => {
	const testCases: Array<[StreamErrorCode, string]> = [
		[0, "Success"],
		[400, "Bad Request"],
		[401, "Unauthorized"],
		[403, "Forbidden"],
		[404, "Not Found"],
		[500, "Internal Server Error"],
		[503, "Service Unavailable"],
		[-1, "Custom negative code"],
		[999999, "Large error code"],
	];
	testCases.forEach(([code, message]) => {
		const error = new StreamError(code, message);
		assertEquals(error.code, code);
		assertEquals(error.message, message);
	});
});

Deno.test("StreamError - message handling - empty message", () => {
	const error = new StreamError(1, "");
	assertEquals(error.message, "");
	assertEquals(error.code, 1);
});

Deno.test("StreamError - message handling - unicode", () => {
	const message = "エラーが発生しました 🚨";
	const error = new StreamError(2, message);
	assertEquals(error.message, message);
});

Deno.test("StreamError - message handling - very long messages", () => {
	const longMessage = "A".repeat(10000);
	const error = new StreamError(3, longMessage);
	assertEquals(error.message, longMessage);
	assertEquals(error.message.length, 10000);
});

Deno.test("StreamError - remote flag behavior", () => {
	const localError = new StreamError(1, "Local error", false);
	const remoteError = new StreamError(2, "Remote error", true);
	assertEquals(localError.remote, false);
	assertEquals(remoteError.remote, true);
});

Deno.test("StreamError - boolean conversion", () => {
	const truthyError = new StreamError(1, "Test", true);
	const falsyError = new StreamError(2, "Test", false);
	assertEquals(!!truthyError.remote, true);
	assertEquals(!!falsyError.remote, false);
});

Deno.test("StreamError - throwable and catchable", () => {
	const code = 418;
	const message = "I'm a teapot";
	assertThrows(() => {
		throw new StreamError(code, message);
	}, StreamError as any);
	try {
		throw new StreamError(code, message);
	} catch (error) {
		assertInstanceOf(error, StreamError);
		if (error instanceof StreamError) {
			assertEquals(error.code, code);
			assertEquals(error.message, message);
		}
	}
});

Deno.test("StreamError - preserve stack trace", () => {
	const error = new StreamError(500, "Stack trace test");
	assertExists(error.stack);
	assertEquals(typeof error.stack, "string");
	if (error.stack) {
		const ok = error.stack.includes("error.test.ts") ||
			error.stack.includes("Object.<anonymous>");
		assertEquals(ok, true);
	}
});

Deno.test("StreamError - serialization - JSON serializable", () => {
	const error = new StreamError(123, "Serialization test", true);
	const serialized = JSON.stringify(error);
	const parsed = JSON.parse(serialized);
	assertEquals(parsed.code, 123);
	assertEquals(parsed.message, undefined);
	assertEquals(parsed.remote, true);
	const manualSerialized = JSON.stringify({
		code: error.code,
		message: error.message,
		remote: error.remote,
	});
	const manualParsed = JSON.parse(manualSerialized);
	assertEquals(manualParsed.code, 123);
	assertEquals(manualParsed.message, "Serialization test");
	assertEquals(manualParsed.remote, true);
});

Deno.test("StreamError - serialization - circular references", () => {
	const error = new StreamError(456, "Circular test");
	(error as any).self = error;
	assertThrows(() => {
		JSON.stringify(error);
	});
});
