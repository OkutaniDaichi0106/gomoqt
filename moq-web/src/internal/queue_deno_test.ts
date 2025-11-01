import { assertEquals, assertExists } from "@std/assert";
import { Queue } from "./queue.ts";

Deno.test("internal/queue - basic enqueue/dequeue and close behavior", async (t) => {
	await t.step("constructor and basic properties", () => {
		const q = new Queue<number>();
		assertExists(q);
		assertEquals(q.closed, false);
		q.close();
		assertEquals(q.closed, true);
	});

	await t.step("enqueue and dequeue FIFO order", async () => {
		const q = new Queue<number>();
		await q.enqueue(1);
		await q.enqueue(2);
		await q.enqueue(3);

		const a = await q.dequeue();
		const b = await q.dequeue();
		const c = await q.dequeue();

		assertEquals(a, 1);
		assertEquals(b, 2);
		assertEquals(c, 3);
		q.close();
	});

	await t.step("wait for items when empty", async () => {
		const q = new Queue<number>();
		let result: number | undefined;

		const p = q.dequeue().then((v) => {
			result = v;
		});

		// enqueue after a tick
		setTimeout(() => {
			q.enqueue(42);
		}, 10);

		await p;
		assertEquals(result, 42);
		q.close();
	});

	await t.step("multiple waiters receive items in order", async () => {
		const q = new Queue<number>();
		const results: (number | undefined)[] = [];

		const promises = Array.from({ length: 3 }).map((_, i) =>
			q.dequeue().then((v) => {
				results[i] = v;
			})
		);

		// enqueue values
		await q.enqueue(100);
		await q.enqueue(200);
		await q.enqueue(300);

		await Promise.all(promises);
		assertEquals(results, [100, 200, 300]);
		q.close();
	});

	await t.step("dequeue after close returns undefined when empty", async () => {
		const q = new Queue<number>();
		q.close();
		const v = await q.dequeue();
		assertEquals(v, undefined);
	});

	await t.step("remaining items can be dequeued after close", async () => {
		const q = new Queue<number>();
		await q.enqueue(1);
		await q.enqueue(2);
		q.close();
		const a = await q.dequeue();
		const b = await q.dequeue();
		const c = await q.dequeue();
		assertEquals(a, 1);
		assertEquals(b, 2);
		assertEquals(c, undefined);
	});
});
