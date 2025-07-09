import { Queue } from './queue';

describe('Queue', () => {
    let queue: Queue<number>;

    beforeEach(() => {
        queue = new Queue<number>();
    });

    describe('enqueue and dequeue', () => {
        it('should enqueue and dequeue items in FIFO order', async () => {
            await queue.enqueue(1);
            await queue.enqueue(2);
            await queue.enqueue(3);

            const [item1, error1] = await queue.dequeue();
            expect(item1).toBe(1);
            expect(error1).toBeUndefined();

            const [item2, error2] = await queue.dequeue();
            expect(item2).toBe(2);
            expect(error2).toBeUndefined();

            const [item3, error3] = await queue.dequeue();
            expect(item3).toBe(3);
            expect(error3).toBeUndefined();
        });

        it('should handle single item correctly', async () => {
            await queue.enqueue(42);
            
            const [item, error] = await queue.dequeue();
            expect(item).toBe(42);
            expect(error).toBeUndefined();
        });

        it('should handle different data types', async () => {
            const stringQueue = new Queue<string>();
            await stringQueue.enqueue('hello');
            await stringQueue.enqueue('world');

            const [item1] = await stringQueue.dequeue();
            const [item2] = await stringQueue.dequeue();

            expect(item1).toBe('hello');
            expect(item2).toBe('world');
        });

        it('should handle object types', async () => {
            interface TestObject {
                id: number;
                name: string;
            }

            const objectQueue = new Queue<TestObject>();
            const obj1 = { id: 1, name: 'first' };
            const obj2 = { id: 2, name: 'second' };

            await objectQueue.enqueue(obj1);
            await objectQueue.enqueue(obj2);

            const [item1] = await objectQueue.dequeue();
            const [item2] = await objectQueue.dequeue();

            expect(item1).toBe(obj1);
            expect(item2).toBe(obj2);
        });

        it('should block dequeue until item is available', async () => {
            const testItem = 42;
            let dequeueCompleted = false;
            
            // Start dequeue operation (should block)
            const dequeuePromise = queue.dequeue().then(([item, error]) => {
                dequeueCompleted = true;
                expect(item).toBe(testItem);
                expect(error).toBeUndefined();
                return [item, error];
            });
            
            // Wait a bit to ensure dequeue is waiting
            await new Promise(resolve => setTimeout(resolve, 10));
            expect(dequeueCompleted).toBe(false);
            
            // Enqueue an item
            await queue.enqueue(testItem);
            
            // Now dequeue should complete
            await dequeuePromise;
            expect(dequeueCompleted).toBe(true);
        });
    });

    describe('concurrent operations', () => {
        it('should handle concurrent enqueue and dequeue operations', async () => {
            const items = [1, 2, 3, 4, 5];
            const results: number[] = [];
            
            // Start multiple enqueue operations concurrently
            const enqueuePromises = items.map(item => queue.enqueue(item));
            
            // Start multiple dequeue operations concurrently
            const dequeuePromises = items.map(async () => {
                const [item] = await queue.dequeue();
                if (item !== undefined) {
                    results.push(item);
                }
            });
            
            // Wait for all operations to complete
            await Promise.all([...enqueuePromises, ...dequeuePromises]);
            
            // Results should contain all items (order might vary due to concurrency)
            expect(results.sort()).toEqual(items.sort());
        });

        it('should handle rapid enqueue/dequeue operations', async () => {
            const items = Array.from({ length: 100 }, (_, i) => i);
            
            // Enqueue all items
            await Promise.all(items.map(item => queue.enqueue(item)));
            
            // Dequeue all items
            const dequeuedItems: number[] = [];
            for (let i = 0; i < items.length; i++) {
                const [item, error] = await queue.dequeue();
                expect(error).toBeUndefined();
                if (item !== undefined) {
                    dequeuedItems.push(item);
                }
            }
            
            expect(dequeuedItems).toEqual(items);
        });

        it('should handle interleaved enqueue/dequeue operations', async () => {
            const results: number[] = [];
            
            for (let i = 0; i < 10; i++) {
                // Enqueue two items
                await queue.enqueue(i * 2);
                await queue.enqueue(i * 2 + 1);
                
                // Dequeue one item
                const [item, error] = await queue.dequeue();
                expect(error).toBeUndefined();
                if (item !== undefined) {
                    results.push(item);
                }
            }
            
            // Dequeue remaining items
            for (let i = 0; i < 10; i++) {
                const [item, error] = await queue.dequeue();
                expect(error).toBeUndefined();
                if (item !== undefined) {
                    results.push(item);
                }
            }
            
            // Should have all items in correct order
            const expectedItems = Array.from({ length: 20 }, (_, i) => i);
            expect(results).toEqual(expectedItems);
        });
    });

    describe('edge cases', () => {
        it('should handle null and undefined values', async () => {
            const nullQueue = new Queue<null>();
            const undefinedQueue = new Queue<undefined>();
            
            await nullQueue.enqueue(null);
            await undefinedQueue.enqueue(undefined);
            
            const [nullItem, nullError] = await nullQueue.dequeue();
            const [undefinedItem, undefinedError] = await undefinedQueue.dequeue();
            
            expect(nullItem).toBeNull();
            expect(nullError).toBeUndefined();
            expect(undefinedItem).toBeUndefined();
            expect(undefinedError).toBeUndefined();
        });

        it('should handle zero values correctly', async () => {
            await queue.enqueue(0);
            
            const [item, error] = await queue.dequeue();
            expect(item).toBe(0);
            expect(error).toBeUndefined();
        });

        it('should handle false boolean values correctly', async () => {
            const boolQueue = new Queue<boolean>();
            await boolQueue.enqueue(false);
            
            const [item, error] = await boolQueue.dequeue();
            expect(item).toBe(false);
            expect(error).toBeUndefined();
        });

        it('should handle very large queues', async () => {
            const largeSize = 1000;
            
            // Fill queue
            await Promise.all(
                Array.from({ length: largeSize }, (_, i) => queue.enqueue(i))
            );
            
            // Empty queue
            for (let i = 0; i < largeSize; i++) {
                const [item, error] = await queue.dequeue();
                expect(item).toBe(i);
                expect(error).toBeUndefined();
            }
        });
    });
});
