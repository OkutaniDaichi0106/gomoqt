import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows } from "../../deps.ts";
import { Queue } from './queue.ts';

describe('Queue', () => {
  let queue: Queue<number>;

  beforeEach(() => {
    queue = new Queue<number>();
  });

  afterEach(() => {
    if (!queue.closed) {
      queue.close();
    }
  });

  it('should be defined', () => {
    assertExists(Queue);
  });

  it('should create an empty queue', () => {
    assertExists(queue);
    assertEquals(queue.closed, false);
  });

  it('should enqueue and dequeue items', async () => {
    await queue.enqueue(1);
    await queue.enqueue(2);
    await queue.enqueue(3);

    const item1 = await queue.dequeue();
    const item2 = await queue.dequeue();
    const item3 = await queue.dequeue();

    assertEquals(item1, 1);
    assertEquals(item2, 2);
    assertEquals(item3, 3);
  });

  it('should handle FIFO order', async () => {
    const items = [10, 20, 30, 40, 50];
    
    // Enqueue all items
    for (const item of items) {
      await queue.enqueue(item);
    }

    // Dequeue all items and verify order
    for (const expectedItem of items) {
      const dequeuedItem = await queue.dequeue();
      assertEquals(dequeuedItem, expectedItem);
    }
  });

  it('should wait for items when queue is empty', async () => {
    let dequeueResult: number | undefined;
    let dequeueError: Error | undefined;

    // Start dequeue operation (will wait)
    const dequeuePromise = queue.dequeue().then(
      result => { dequeueResult = result; },
      error => { dequeueError = error; }
    );

    // Give some time to ensure dequeue is waiting
    await new Promise(resolve => setTimeout(resolve, 10));
    assertEquals(dequeueResult, undefined);
    assertEquals(dequeueError, undefined);

    // Enqueue an item
    await queue.enqueue(42);

    // Wait for dequeue to complete
    await dequeuePromise;

    assertEquals(dequeueResult, 42);
    assertEquals(dequeueError, undefined);
  });

  it('should handle multiple waiters', async () => {
    const results: (number | undefined)[] = [];
    const errors: (Error | undefined)[] = [];

    // Start multiple dequeue operations
    const promises = Array.from({ length: 3 }, (_, i) =>
      queue.dequeue().then(
        result => { results[i] = result; },
        error => { errors[i] = error; }
      )
    );

    // Give some time to ensure all dequeues are waiting
    await new Promise(resolve => setTimeout(resolve, 10));

    // Enqueue items one by one
    await queue.enqueue(100);
    await queue.enqueue(200);
    await queue.enqueue(300);

    // Wait for all dequeue operations to complete
    await Promise.all(promises);

    assertEquals(results, [100, 200, 300]);
    expect(errors.every(err => err === undefined)).toBe(true);
  });

  it('should close the queue', () => {
    assertEquals(queue.closed, false);
    queue.close();
    assertEquals(queue.closed, true);
  });

  it('should throw error when dequeuing from closed empty queue', async () => {
    queue.close();
    
    const val = await queue.dequeue();
    assertEquals(val, undefined);
  });

  it('should allow dequeuing remaining items after close', async () => {
    await queue.enqueue(1);
    await queue.enqueue(2);
    
    queue.close();
    
    const item1 = await queue.dequeue();
    const item2 = await queue.dequeue();
    
    assertEquals(item1, 1);
    assertEquals(item2, 2);
    
    const val = await queue.dequeue();
    assertEquals(val, undefined);
  });

  it('should wake up waiters when closed', async () => {
    let dequeueResult: number | undefined;

    // Start dequeue operation on empty queue
    const dequeuePromise = queue.dequeue().then(result => {
      dequeueResult = result;
    });

    // Give some time to ensure dequeue is waiting
    await new Promise(resolve => setTimeout(resolve, 10));

    // Close the queue
    queue.close();

    // Wait for dequeue to complete
    await dequeuePromise;

    assertEquals(dequeueResult, undefined);
  });

  it('should handle string items', async () => {
    const stringQueue = new Queue<string>();
    
    await stringQueue.enqueue('hello');
    await stringQueue.enqueue('world');
    
    const item1 = await stringQueue.dequeue();
    const item2 = await stringQueue.dequeue();
    
    assertEquals(item1, 'hello');
    assertEquals(item2, 'world');
    
    stringQueue.close();
  });

  it('should handle object items', async () => {
    interface TestObject {
      id: number;
      name: string;
    }
    
    const objectQueue = new Queue<TestObject>();
    const obj1 = { id: 1, name: 'test1' };
    const obj2 = { id: 2, name: 'test2' };
    
    await objectQueue.enqueue(obj1);
    await objectQueue.enqueue(obj2);
    
    const result1 = await objectQueue.dequeue();
    const result2 = await objectQueue.dequeue();
    
    assertEquals(result1, obj1);
    assertEquals(result2, obj2);
    
    objectQueue.close();
  });
});
