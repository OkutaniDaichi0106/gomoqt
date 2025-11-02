import { assertEquals, assertNotStrictEquals, assertStrictEquals } from "@std/assert";
import { BufferPool } from "./buffer_pool.ts";

Deno.test("webtransport/buffer_pool", async (t) => {
	await t.step("acquire and release bytes", () => {
		const pool = new BufferPool({ min: 1, middle: 10, max: 100 });
		const bytes = pool.acquire(10);
		assertEquals(bytes.byteLength, 10);
		pool.release(bytes);
	});

	await t.step("reuse bytes from pool", () => {
		const pool = new BufferPool({ min: 1, middle: 10, max: 100 });
		const bytes1 = pool.acquire(10);
		pool.release(bytes1);
		const bytes2 = pool.acquire(10);
		// Ensure we reused the same buffer instance
		assertStrictEquals(bytes2, bytes1);
	});

	await t.step("not reuse when capacity too small", () => {
		const pool = new BufferPool({ min: 1, middle: 10, max: 100 });
		const bytes1 = pool.acquire(10);
		pool.release(bytes1);
		const bytes2 = pool.acquire(20);
		// Different sized buckets should not return the same buffer instance
		assertNotStrictEquals(bytes2, bytes1);
	});

	await t.step("cleanup old bytes", async () => {
		const pool = new BufferPool({
			min: 1,
			middle: 10,
			max: 100,
			options: { maxPerBucket: 10, maxTotalBytes: 10 },
		});
		const bytes1 = pool.acquire(10);
		pool.release(bytes1);
		await new Promise((resolve) => setTimeout(resolve, 20));
		pool.cleanup();
		const bytes2 = pool.acquire(10);
		// After cleanup we should not get back the same buffer instance
		assertNotStrictEquals(bytes2, bytes1);
	});

	await t.step("create new buffer when capacity exceeds max", () => {
		const pool = new BufferPool({ min: 1, middle: 10, max: 100 });
		const bytes = pool.acquire(200);
		assertEquals(bytes.byteLength, 200);
	});

	await t.step("handle empty bucket", () => {
		const pool = new BufferPool({ min: 1, middle: 10, max: 100 });
		for (let i = 0; i < 6; i++) {
			const bytes = pool.acquire(10);
			if (i < 5) pool.release(bytes);
		}
		const bytes = pool.acquire(10);
		assertEquals(bytes.byteLength, 10);
	});

	await t.step("not release when maxTotalBytes exceeded", () => {
		const pool = new BufferPool({
			min: 1,
			middle: 10,
			max: 100,
			options: { maxTotalBytes: 10 },
		});
		const bytes1 = pool.acquire(10);
		pool.release(bytes1);
		const bytes2 = pool.acquire(10);
		pool.release(bytes2);
		const bytes3 = new ArrayBuffer(10);
		pool.release(bytes3);
		const bytes4 = pool.acquire(10);
		// Ensure capacity bookkeeping yields expected sizing behavior (instance-level)
		assertStrictEquals(bytes4.byteLength, bytes2.byteLength);
		assertNotStrictEquals(bytes4, bytes3);
	});

	await t.step("not release non-matching sizes", () => {
		const pool = new BufferPool({ min: 1, middle: 10, max: 100 });
		const bytes = new ArrayBuffer(50);
		pool.release(bytes);
		const acquired = pool.acquire(10);
		assertEquals(acquired.byteLength, 10);
	});

	await t.step("cleanup buckets", () => {
		const pool = new BufferPool({ min: 1, middle: 10, max: 100 });
		const bytes = pool.acquire(10);
		pool.release(bytes);
		pool.cleanup();
		const bytes2 = pool.acquire(10);
		assertNotStrictEquals(bytes2, bytes);
	});
});
