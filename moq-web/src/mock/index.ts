/**
 * Mock utilities for Deno tests using std/testing/mock
 * Provides wrapper functions for easier mocking with compatibility in mind
 */

export {
	assertSpyCall,
	assertSpyCallArg,
	assertSpyCallArgs,
	assertSpyCallAsync,
	assertSpyCalls,
	type ConstructorSpy,
	type MethodSpy,
	mockSession,
	mockSessionAsync,
	resolvesNext,
	restore,
	returnsArg,
	returnsArgs,
	returnsNext,
	returnsThis,
	type Spy,
	spy,
	type SpyCall,
	type Stub,
	stub,
} from "@std/testing/mock";

/**
 * Creates a mock function with built-in call tracking and return value manipulation
 * This is a convenience wrapper around std/testing/mock spy for simple cases
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
 * Creates a simple mock function compatible with the old test-utils/mock API
 * Uses std/testing/mock internally for the implementation
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
 * This is a compatibility wrapper that returns a simple mock function
 */
export function createSpy<T extends (...args: any[]) => any>(
	fn: T,
): MockFunction<T> {
	return createMock(fn);
}
