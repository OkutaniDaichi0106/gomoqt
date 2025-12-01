import {
  assertEquals,
  assertExists,
  assertInstanceOf,
  assertThrows,
  fail,
} from "@std/assert";
import type { WebTransportStreamErrorCode } from "./error.ts";
import { WebTransportStreamError } from "./error.ts";

Deno.test("webtransport/error - StreamError behavior", async (t) => {
  await t.step("constructor sets fields correctly", () => {
    const code: WebTransportStreamErrorCode = 404;
    const message = `stream was reset with code ${code}`;

    const error = new WebTransportStreamError(
      { source: "stream", streamErrorCode: code },
      false,
    );

    assertEquals(error.code, code);
    assertEquals(error.message, message);
    assertEquals(error.remote, false);
    assertEquals(error.name, "Error");
    assertEquals(error instanceof Error, true);
    assertEquals(error instanceof WebTransportStreamError, true);
  });

  await t.step("constructor accepts remote flag", () => {
    const code: WebTransportStreamErrorCode = 500;
    const message = `stream was reset with code ${code}`;
    const remote = true;

    const error = new WebTransportStreamError(
      { source: "stream", streamErrorCode: code },
      remote,
    );

    assertEquals(error.code, code);
    assertEquals(error.message, message);
    assertEquals(error.remote, remote);
  });

  await t.step("defaults remote to false", () => {
    const error = new WebTransportStreamError(
      { source: "stream", streamErrorCode: 200 },
      false,
    );
    assertEquals(error.remote, false);
  });

  await t.step("prototype chain and instanceof", () => {
    const error = new WebTransportStreamError(
      { source: "stream", streamErrorCode: 123 },
      false,
    );
    assertEquals(error instanceof WebTransportStreamError, true);
    assertEquals(error instanceof Error, true);
    assertEquals(
      Object.getPrototypeOf(error),
      WebTransportStreamError.prototype,
    );

    const original = new WebTransportStreamError(
      { source: "stream", streamErrorCode: 456 },
      true,
    );
    const recreated = Object.create(WebTransportStreamError.prototype);
    Object.assign(recreated, original as unknown as Record<string, unknown>);
    assertEquals(recreated instanceof WebTransportStreamError, true);
    // Object.assign doesn't copy non-enumerable Error.message
    // we expect assigned properties to match what's enumerable
    assertEquals((recreated as any).code, 456);
    assertEquals((recreated as any).remote, true);
  });

  await t.step("handles various error codes", () => {
    const testCases: Array<[WebTransportStreamErrorCode, string]> = [
      [0, `stream was reset with code 0`],
      [400, `stream was reset with code 400`],
      [401, `stream was reset with code 401`],
      [403, `stream was reset with code 403`],
      [404, `stream was reset with code 404`],
      [500, `stream was reset with code 500`],
      [503, `stream was reset with code 503`],
      [-1, `stream was reset with code -1`],
      [999999, `stream was reset with code 999999`],
    ];

    for (const [code, message] of testCases) {
      const error = new WebTransportStreamError(
        { source: "stream", streamErrorCode: code },
        false,
      );
      assertEquals(error.code, code);
      assertEquals(error.message, message);
    }
  });

  await t.step("message handling", () => {
    const error1 = new WebTransportStreamError({
      source: "stream",
      streamErrorCode: 1,
    }, false);
    assertEquals(error1.message, `stream was reset with code 1`);
    assertEquals(error1.code, 1);

    const error2 = new WebTransportStreamError({
      source: "stream",
      streamErrorCode: 2,
    }, false);
    assertEquals(error2.message, `stream was reset with code 2`);
    const error3 = new WebTransportStreamError({
      source: "stream",
      streamErrorCode: 3,
    }, false);
    assertEquals(error3.message, `stream was reset with code 3`);
    // remote flag checks moved to the remote flag behavior test

    const truthyError = new WebTransportStreamError(
      { source: "stream", streamErrorCode: 1 },
      true,
    );
    const falsyError = new WebTransportStreamError(
      { source: "stream", streamErrorCode: 2 },
      false,
    );
    assertEquals(!!truthyError.remote, true);
    assertEquals(!!falsyError.remote, false);
  });

  await t.step("throwing and catching", () => {
    const code = 418;
    const message = `stream was reset with code ${code}`;

    assertThrows(() => {
      throw new WebTransportStreamError({
        source: "stream",
        streamErrorCode: code,
      }, false);
    });

    try {
      throw new WebTransportStreamError({
        source: "stream",
        streamErrorCode: code,
      }, false);
    } catch (error) {
      assertInstanceOf(
        error,
        WebTransportStreamError as unknown as new (...args: any[]) => Error,
      );
      if (error instanceof WebTransportStreamError) {
        assertEquals(error.code, code);
        assertEquals(error.message, message);
      }
    }
  });

  await t.step("preserves stack trace", () => {
    const error = new WebTransportStreamError(
      { source: "stream", streamErrorCode: 500 },
      false,
    );
    assertExists(error.stack);
    assertEquals(typeof error.stack, "string");
    if (error.stack) {
      const ok = error.stack.includes("error_test.ts") ||
        error.stack.includes("error.test.ts") ||
        error.stack.includes("Object.<anonymous>");
      if (!ok) fail("stack trace does not include expected markers");
    }
  });

  await t.step("serialization", () => {
    const error = new WebTransportStreamError({
      source: "stream",
      streamErrorCode: 123,
    }, true);
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
    assertEquals(manualParsed.message, "stream was reset with code 123");
    assertEquals(manualParsed.remote, true);
  });

  await t.step("circular references throw on JSON.stringify", () => {
    const error = new WebTransportStreamError(
      { source: "stream", streamErrorCode: 456 },
      false,
    );
    (error as any).self = error;
    assertThrows(() => {
      JSON.stringify(error);
    });
  });
});
