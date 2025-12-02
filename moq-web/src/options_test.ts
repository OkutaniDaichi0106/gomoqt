import { assertEquals, assertExists } from "@std/assert";
import type { MOQOptions } from "./options.ts";
import { Extensions } from "./extensions.ts";

// Test configuration to ignore resource leaks from background operations
const testOptions = {
	sanitizeResources: false,
	sanitizeOps: false,
};

Deno.test("MOQOptions", testOptions, async (t) => {
	await t.step("should define the correct interface structure", () => {
		// This is a type-only test to ensure the interface is correctly defined
		const mockOptions: MOQOptions = {
			versions: new Set([1]),
			extensions: new Extensions(),
		};

		assertExists(mockOptions.extensions);
		assertEquals(mockOptions.extensions instanceof Extensions, true);
		assertEquals(mockOptions.versions instanceof Set, true);
		assertEquals(mockOptions.versions?.has(1), true);
	});

	await t.step("should allow empty options", () => {
		// Extensions should be optional
		const emptyOptions: MOQOptions = {
			versions: new Set(),
		};

		assertEquals(emptyOptions.extensions, undefined);
		assertEquals(emptyOptions.versions instanceof Set, true);
	});

	await t.step("should allow options with extensions", () => {
		const extensions = new Extensions();
		extensions.addString(1, "test");

		const options: MOQOptions = {
			versions: new Set([1]),
			extensions: extensions,
		};

		assertEquals(options.extensions, extensions);
		assertEquals(options.extensions?.getString(1), "test");
		assertEquals(options.versions?.has(1), true);
	});

	await t.step("should support partial assignment", () => {
		// Should be able to create options incrementally
		const options: MOQOptions = {
			versions: new Set(),
		};

		// Initially no extensions
		assertEquals(options.extensions, undefined);
		assertEquals(options.versions instanceof Set, true);

		// Can add extensions later
		options.extensions = new Extensions();
		assertEquals(options.extensions instanceof Extensions, true);
	});

	await t.step(
		"should be compatible with different extension configurations",
		() => {
			const extensions1 = new Extensions();
			extensions1.addBytes(1, new Uint8Array([1, 2, 3]));

			const extensions2 = new Extensions();
			extensions2.addString(2, "test");
			extensions2.addNumber(3, 42n);

			const options1: MOQOptions = {
				versions: new Set([1]),
				extensions: extensions1,
			};
			const options2: MOQOptions = {
				versions: new Set([2]),
				extensions: extensions2,
			};

			assertEquals(options1.extensions?.getBytes(1), new Uint8Array([1, 2, 3]));
			assertEquals(options2.extensions?.getString(2), "test");
			assertEquals(options2.extensions?.getNumber(3), 42n);
			assertEquals(options1.versions?.has(1), true);
			assertEquals(options2.versions?.has(2), true);
		},
	);

	await t.step("should support transportOptions", () => {
		const transportOptions: WebTransportOptions = {
			allowPooling: true,
			congestionControl: "throughput",
		};

		const options: MOQOptions = {
			versions: new Set([1]),
			transportOptions: transportOptions,
		};

		assertEquals(options.transportOptions, transportOptions);
		assertEquals(options.transportOptions?.allowPooling, true);
		assertEquals(options.transportOptions?.congestionControl, "throughput");
	});
});
