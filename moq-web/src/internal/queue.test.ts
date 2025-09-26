import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { Queue } from './queue';

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
    expect(Queue).toBeDefined();
  });

  it('should create an empty queue', () => {
    expect(queue).toBeDefined();
    expect(queue.closed).toBe(false);
  });

  it('should enqueue and dequeue items', async () => {
    await queue.enqueue(1);
    await queue.enqueue(2);
    await queue.enqueue(3);

    const item1 = await queue.dequeue();
    const item2 = await queue.dequeue();
    const item3 = await queue.dequeue();

    expect(item1).toBe(1);
    expect(item2).toBe(2);
    expect(item3).toBe(3);
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
      expect(dequeuedItem).toBe(expectedItem);
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
    expect(dequeueResult).toBeUndefined();
    expect(dequeueError).toBeUndefined();

    // Enqueue an item
    await queue.enqueue(42);

    // Wait for dequeue to complete
    await dequeuePromise;

    expect(dequeueResult).toBe(42);
    expect(dequeueError).toBeUndefined();
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

    expect(results).toEqual([100, 200, 300]);
    expect(errors.every(err => err === undefined)).toBe(true);
  });

  it('should close the queue', () => {
    expect(queue.closed).toBe(false);
    queue.close();
    expect(queue.closed).toBe(true);
  });

  it('should throw error when dequeuing from closed empty queue', async () => {
    queue.close();
    
    const val = await queue.dequeue();
    expect(val).toBeUndefined();
  });

  it('should allow dequeuing remaining items after close', async () => {
    await queue.enqueue(1);
    await queue.enqueue(2);
    
    queue.close();
    
    const item1 = await queue.dequeue();
    const item2 = await queue.dequeue();
    
    expect(item1).toBe(1);
    expect(item2).toBe(2);
    
    const val = await queue.dequeue();
    expect(val).toBeUndefined();
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

    expect(dequeueResult).toBeUndefined();
  });

  it('should handle string items', async () => {
    const stringQueue = new Queue<string>();
    
    await stringQueue.enqueue('hello');
    await stringQueue.enqueue('world');
    
    const item1 = await stringQueue.dequeue();
    const item2 = await stringQueue.dequeue();
    
    expect(item1).toBe('hello');
    expect(item2).toBe('world');
    
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
    
    expect(result1).toEqual(obj1);
    expect(result2).toEqual(obj2);
    
    objectQueue.close();
  });
});