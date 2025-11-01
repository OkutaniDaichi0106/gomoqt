/**
 * Mock test file demonstrating usage of mocks with std/testing/mock
 * Shows examples of creating mocks, spies, stubs, and using assertions
 */

import {
  assertEquals,
} from "@std/assert";
import {
  spy,
  stub,
  returnsNext,
  assertSpyCall,
  assertSpyCalls,
} from "@std/testing/mock";
import {
  createMock,
  createSpy,
  type MockFunction,
} from "./src/mock/index.ts";

// Example 1: createMock - should track function calls
Deno.test("createMock - should track function calls", () => {
  const mockFn = createMock<(x: number) => number>();

  mockFn(5);
  mockFn(10);

  assertEquals(mockFn.calls.length, 2);
  assertEquals(mockFn.calls[0], [5]);
  assertEquals(mockFn.calls[1], [10]);
});

// Example 2: createMock - should support mockReturnValue
Deno.test("createMock - should support mockReturnValue", () => {
  const mockFn = createMock<() => string>();
  mockFn.mockReturnValue("hello");

  const result = mockFn();

  assertEquals(result, "hello");
  assertEquals(mockFn.results.length, 1);
  assertEquals(mockFn.results[0], "hello");
});

// Example 3: createMock - should support mockResolvedValue
Deno.test("createMock - should support mockResolvedValue", async () => {
  const mockFn = createMock<() => Promise<number>>();
  mockFn.mockResolvedValue(42);

  const result = await mockFn();

  assertEquals(result, 42);
});

// Example 4: createMock - should support mockImplementation
Deno.test("createMock - should support mockImplementation", () => {
  const mockFn = createMock<(x: number) => number>();
  mockFn.mockImplementation((x) => x * 2);

  const result = mockFn(5);

  assertEquals(result, 10);
});

// Example 5: createMock - should support mockReset
Deno.test("createMock - should support mockReset", () => {
  const mockFn = createMock<() => string>();
  mockFn.mockReturnValue("hello");

  mockFn();
  assertEquals(mockFn.calls.length, 1);

  mockFn.mockReset();
  assertEquals(mockFn.calls.length, 0);
  assertEquals(mockFn.results.length, 0);
});

// Example 6: createSpy - should spy on function calls
Deno.test("createSpy - should spy on function calls", () => {
  const originalFn = (x: number) => x * 2;
  const spyFn = createSpy(originalFn);

  const result = spyFn(5);

  assertEquals(result, 10);
  assertEquals(spyFn.calls.length, 1);
  assertEquals(spyFn.calls[0], [5]);
});

// Example 7: createSpy - should preserve original behavior
Deno.test("createSpy - should preserve original behavior", () => {
  const add = (a: number, b: number) => a + b;
  const addSpy = createSpy(add);

  assertEquals(addSpy(2, 3), 5);
  assertEquals(addSpy(10, 20), 30);
  assertEquals(addSpy.calls.length, 2);
});

// Example 8: spy - should create a spy on a function
Deno.test("spy - should create a spy on a function", () => {
  const multiply = (a: number, b: number) => a * b;
  const multiplySpy = spy(multiply);

  const result = multiplySpy(5, 6);

  assertEquals(result, 30);
  assertEquals(multiplySpy.calls.length, 1);
  assertSpyCall(multiplySpy, 0, {
    args: [5, 6],
    returned: 30,
  });
});

// Example 9: spy - should spy on object methods
Deno.test("spy - should spy on object methods", () => {
  const obj = {
    getValue: () => 42,
  };

  using methodSpy = spy(obj, "getValue");

  const result = obj.getValue();

  assertEquals(result, 42);
  assertSpyCalls(methodSpy, 1);
});

// Example 10: stub - should stub a method with custom implementation
Deno.test("stub - should stub a method with custom implementation", () => {
  const obj = {
    calculate: (a: number, b: number) => a + b,
  };

  using calculateStub = stub(obj, "calculate", () => 100);

  const result = obj.calculate(5, 5);

  assertEquals(result, 100);
  assertSpyCalls(calculateStub, 1);
});

// Example 11: stub - should stub with returnsNext helper
Deno.test("stub - should stub with returnsNext helper", () => {
  const obj = {
    randomValue: () => Math.random(),
  };

  using valueStub = stub(obj, "randomValue", returnsNext([10, 20, 30]));

  assertEquals(obj.randomValue(), 10);
  assertEquals(obj.randomValue(), 20);
  assertEquals(obj.randomValue(), 30);

  assertSpyCalls(valueStub, 3);
});

// Example 12: Complex example - test UserService with mocked logger
Deno.test("Complex example - test UserService with mocked logger", () => {
  interface Logger {
    log(message: string): void;
    error(message: string): void;
  }

  interface UserService {
    logger: Logger;
    getUser(id: number): { id: number; name: string } | null;
  }

  const mockLogger: Logger = {
    log: createMock<(message: string) => void>(),
    error: createMock<(message: string) => void>(),
  };

  const userService: UserService = {
    logger: mockLogger,
    getUser: (id: number) => {
      mockLogger.log(`Fetching user ${id}`);
      if (id === 1) {
        return { id: 1, name: "Alice" };
      }
      mockLogger.error(`User ${id} not found`);
      return null;
    },
  };

  const user = userService.getUser(1);

  assertEquals(user, { id: 1, name: "Alice" });
  assertEquals((mockLogger.log as MockFunction<any>).calls.length, 1);
  assertEquals(
    (mockLogger.log as MockFunction<any>).calls[0],
    ["Fetching user 1"],
  );
});

// Example 13: Complex example - test with std spy on object methods
Deno.test("Complex example - test with std spy on object methods", () => {
  const logger = {
    log: (_message: string) => {
      // noop
    },
    error: (_message: string) => {
      // noop
    },
  };

  using logSpy = spy(logger, "log");
  using errorSpy = spy(logger, "error");

  logger.log("Test message");
  logger.error("Error message");

  assertSpyCalls(logSpy, 1);
  assertSpyCalls(errorSpy, 1);
  assertSpyCall(logSpy, 0, { args: ["Test message"] });
  assertSpyCall(errorSpy, 0, { args: ["Error message"] });
});

// Example 14: Async mocking - should mock async functions
Deno.test("Async mocking - should mock async functions", async () => {
  const mockFetch = createMock<
    (url: string) => Promise<{ status: number }>
  >();
  mockFetch.mockResolvedValue({ status: 200 });

  const result = await mockFetch("https://example.com");

  assertEquals(result.status, 200);
  assertEquals(mockFetch.calls.length, 1);
  assertEquals(mockFetch.calls[0], ["https://example.com"]);
});

// Example 15: Async mocking - should handle multiple async calls
Deno.test("Async mocking - should handle multiple async calls", async () => {
  const apiCall = createMock<(endpoint: string) => Promise<unknown>>();
  apiCall.mockResolvedValue({ data: "success" });

  await apiCall("/users");
  await apiCall("/posts");

  assertEquals(apiCall.calls.length, 2);
  assertEquals(apiCall.calls[0], ["/users"]);
  assertEquals(apiCall.calls[1], ["/posts"]);
});
