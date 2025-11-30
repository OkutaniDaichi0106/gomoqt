import { assertEquals, assertExists, assertInstanceOf, assertThrows, fail } from "@std/assert";
import type { StreamErrorCode } from "./error.ts";
import { StreamError } from "./error.ts";

Deno.test("webtransport/error - StreamError behavior", async (t) => {
	await t.step("constructor sets fields correctly", () => {
		const code: StreamErrorCode = 404;
		const message = "Not found";

		const error = new StreamError(code, message);

		assertEquals(error.code, code);
		assertEquals(error.message, message);
		assertEquals(error.remote, false);
		assertEquals(error.name, "Error");
		assertEquals(error instanceof Error, true);
		assertEquals(error instanceof StreamError, true);
	});

	await t.step("constructor accepts remote flag", () => {
		const code: StreamErrorCode = 500;
		const message = "Internal server error";
		const remote = true;

		const error = new StreamError(code, message, remote);

		assertEquals(error.code, code);
		assertEquals(error.message, message);
		assertEquals(error.remote, remote);
	});

	await t.step("defaults remote to false", () => {
		const error = new StreamError(200, "OK");
		assertEquals(error.remote, false);
	});

	await t.step("prototype chain and instanceof", () => {
		const error = new StreamError(123, "Test error");
		assertEquals(error instanceof StreamError, true);
		assertEquals(error instanceof Error, true);
		assertEquals(Object.getPrototypeOf(error), StreamError.prototype);

		const original = new StreamError(456, "Original error", true);
		const recreated = Object.create(StreamError.prototype);
		Object.assign(recreated, original as unknown as Record<string, unknown>);
		assertEquals(recreated instanceof StreamError, true);
		// Object.assign doesn't copy non-enumerable Error.message
		// we expect assigned properties to match what's enumerable
		assertEquals((recreated as any).code, 456);
		assertEquals((recreated as any).remote, true);
	});

	await t.step("handles various error codes", () => {
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

		for (const [code, message] of testCases) {
			const error = new StreamError(code, message);
			assertEquals(error.code, code);
			assertEquals(error.message, message);
		}
	});

	await t.step("message handling", () => {
		const error1 = new StreamError(1, "");
		assertEquals(error1.message, "");
		assertEquals(error1.code, 1);

		const message = "エラーが発生しました 🚨";
		const error2 = new StreamError(2, message);
		assertEquals(error2.message, message);

		const longMessage = "A".repeat(10000);
		const error3 = new StreamError(3, longMessage);
		assertEquals(error3.message, longMessage);
		assertEquals(error3.message.length, 10000);
	});

	await t.step("remote flag behavior", () => {
		const localError = new StreamError(1, "Local error", false);
		const remoteError = new StreamError(2, "Remote error", true);
		assertEquals(localError.remote, false);
		assertEquals(remoteError.remote, true);

		const truthyError = new StreamError(1, "Test", true);
		const falsyError = new StreamError(2, "Test", false);
		assertEquals(!!truthyError.remote, true);
		assertEquals(!!falsyError.remote, false);
	});

	await t.step("throwing and catching", () => {
		const code = 418;
		const message = "I'm a teapot";

		assertThrows(() => {
			throw new StreamError(code, message);
		});

		try {
			throw new StreamError(code, message);
		} catch (error) {
			assertInstanceOf(error, StreamError as unknown as new (...args: any[]) => Error);
			if (error instanceof StreamError) {
				assertEquals(error.code, code);
				assertEquals(error.message, message);
			}
		}
	});

	await t.step("preserves stack trace", () => {
		const error = new StreamError(500, "Stack trace test");
		assertExists(error.stack);
		assertEquals(typeof error.stack, "string");
		if (error.stack) {
			const ok = error.stack.includes("error_test.ts") ||
				error.stack.includes("error.test.ts") || error.stack.includes("Object.<anonymous>");
			if (!ok) fail("stack trace does not include expected markers");
		}
	});

	await t.step("serialization", () => {
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

	await t.step("circular references throw on JSON.stringify", () => {
		const error = new StreamError(456, "Circular test");
		(error as any).self = error;
		assertThrows(() => {
			JSON.stringify(error);
		});
	});
});
