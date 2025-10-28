/**
 * Simple mock utilities for Deno tests
 * Provides basic mocking capabilities similar to Vitest's vi.fn()
 */

export interface MockFunction<T extends (...args: any[]) => any> {
  (...args: Parameters<T>): ReturnType<T>;
  calls: Array<Parameters<T>>;
  results: Array<ReturnType<T>>;
  mockReturnValue(value: ReturnType<T>): this;
  mockResolvedValue(value: Awaited<ReturnType<T>>): this;
  mockImplementation(fn: T): this;
  mockReset(): this;
}

/**
 * Creates a mock function similar to Vitest's vi.fn()
 */
export function createMock<T extends (...args: any[]) => any>(
  implementation?: T,
): MockFunction<T> {
  const calls: Array<Parameters<T>> = [];
  const results: Array<ReturnType<T>> = [];
  let mockImpl: T | undefined = implementation;

  const mockFn = function (...args: Parameters<T>): ReturnType<T> {
    calls.push(args);
    const result = mockImpl ? mockImpl(...args) : undefined;
    results.push(result as ReturnType<T>);
    return result as ReturnType<T>;
  } as MockFunction<T>;

  mockFn.calls = calls;
  mockFn.results = results;

  mockFn.mockReturnValue = function (value: ReturnType<T>) {
    mockImpl = (() => value) as T;
    return this;
  };

  mockFn.mockResolvedValue = function (value: Awaited<ReturnType<T>>) {
    mockImpl = (() => Promise.resolve(value)) as T;
    return this;
  };

  mockFn.mockImplementation = function (fn: T) {
    mockImpl = fn;
    return this;
  };

  mockFn.mockReset = function () {
    calls.length = 0;
    results.length = 0;
    mockImpl = implementation;
    return this;
  };

  return mockFn;
}

/**
 * Creates a spy function that tracks calls but doesn't change behavior
 */
export function createSpy<T extends (...args: any[]) => any>(
  fn: T,
): MockFunction<T> {
  return createMock(fn);
}
