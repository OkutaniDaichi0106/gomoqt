import {
	afterEach,
	assertEquals,
	assertExists,
	assertThrows,
	beforeEach,
	describe,
	it,
} from "../deps.ts";
import type { MOQOptions } from "./options.ts";
import { Extensions } from "./internal/extensions.ts";

describe("MOQOptions", () => {
	it("should define the correct interface structure", () => {
		// This is a type-only test to ensure the interface is correctly defined
		const mockOptions: MOQOptions = {
			versions: new Set([1n]),
			extensions: new Extensions(),
		};

		assertExists(mockOptions.extensions);
		assertInstanceOf(mockOptions.extensions, Extensions);
		assertInstanceOf(mockOptions.versions, Set);
		expect(mockOptions.versions?.has(1n)).toBe(true);
	});

	it("should allow empty options", () => {
		// Extensions should be optional
		const emptyOptions: MOQOptions = {
			versions: new Set(),
		};

		assertEquals(emptyOptions.extensions, undefined);
		assertInstanceOf(emptyOptions.versions, Set);
	});

	it("should allow options with extensions", () => {
		const extensions = new Extensions();
		extensions.addString(1, "test");

		const options: MOQOptions = {
			versions: new Set([1n]),
			extensions: extensions,
		};

		assertEquals(options.extensions, extensions);
		expect(options.extensions?.getString(1)).toBe("test");
		expect(options.versions?.has(1n)).toBe(true);
	});

	it("should support partial assignment", () => {
		// Should be able to create options incrementally
		const options: MOQOptions = {
			versions: new Set(),
		};

		// Initially no extensions
		assertEquals(options.extensions, undefined);
		assertInstanceOf(options.versions, Set);

		// Can add extensions later
		options.extensions = new Extensions();
		assertInstanceOf(options.extensions, Extensions);
	});

	it("should be compatible with different extension configurations", () => {
		const extensions1 = new Extensions();
		extensions1.addBytes(1, new Uint8Array([1, 2, 3]));

		const extensions2 = new Extensions();
		extensions2.addString(2, "test");
		extensions2.addNumber(3, 42n);

		const options1: MOQOptions = { versions: new Set([1n]), extensions: extensions1 };
		const options2: MOQOptions = { versions: new Set([2n]), extensions: extensions2 };

		expect(options1.extensions?.getBytes(1)).toEqual(new Uint8Array([1, 2, 3]));
		expect(options2.extensions?.getString(2)).toBe("test");
		expect(options2.extensions?.getNumber(3)).toBe(42n);
		expect(options1.versions?.has(1n)).toBe(true);
		expect(options2.versions?.has(2n)).toBe(true);
	});

	it("should support transportOptions", () => {
		const transportOptions: WebTransportOptions = {
			allowPooling: true,
			congestionControl: "throughput",
		};

		const options: MOQOptions = {
			versions: new Set([1n]),
			transportOptions: transportOptions,
		};

		assertEquals(options.transportOptions, transportOptions);
		assertEquals(options.transportOptions?.allowPooling, true);
		assertEquals(options.transportOptions?.congestionControl, "throughput");
	});
});
