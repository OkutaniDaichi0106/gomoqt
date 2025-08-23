import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { Mutex } from './mutex';

describe('Mutex', () => {
    let mutex: Mutex;

    beforeEach(() => {
        mutex = new Mutex();
    });

    describe('constructor', () => {
        it('should create a mutex', () => {
            expect(mutex).toBeDefined();
            expect(typeof mutex.lock).toBe('function');
        });
    });

    describe('lock', () => {
        it('should return an unlock function', async () => {
            const unlock = await mutex.lock();
            expect(typeof unlock).toBe('function');
            unlock();
        });

        it('should provide exclusive access', async () => {
            let counter = 0;
            const promises: Promise<void>[] = [];
            
            // Start multiple concurrent operations
            for (let i = 0; i < 5; i++) {
                promises.push(
                    (async () => {
                        const unlock = await mutex.lock();
                        const startValue = counter;
                        // Simulate some work
                        await new Promise(resolve => setTimeout(resolve, 1));
                        counter = startValue + 1;
                        unlock();
                    })()
                );
            }
            
            await Promise.all(promises);
            expect(counter).toBe(5); // Should be exactly 5 if mutex works correctly
        });

        it('should unlock when unlock function is called', async () => {
            const unlock = await mutex.lock();
            
            let secondLockAcquired = false;
            const secondLockPromise = mutex.lock().then((unlock2) => {
                secondLockAcquired = true;
                unlock2();
            });
            
            // Wait a bit to ensure the second lock is waiting
            await new Promise(resolve => setTimeout(resolve, 10));
            expect(secondLockAcquired).toBe(false);
            
            unlock();
            
            // Now the second lock should be acquired
            await secondLockPromise;
            expect(secondLockAcquired).toBe(true);
        });
    });

    describe('sequential locking', () => {
        it('should allow sequential locks', async () => {
            // First lock
            const unlock1 = await mutex.lock();
            unlock1();
            
            // Wait for unlock to complete
            await new Promise(resolve => setTimeout(resolve, 0));
            
            // Second lock should work
            const unlock2 = await mutex.lock();
            unlock2();
        });
    });

    describe('concurrent locking', () => {
        it('should queue concurrent lock requests', async () => {
            const executionOrder: number[] = [];
            
            const createLockTask = (id: number) => async () => {
                const unlock = await mutex.lock();
                executionOrder.push(id);
                // Hold lock for a short time
                await new Promise(resolve => setTimeout(resolve, 5));
                unlock();
            };
            
            // Start multiple lock tasks concurrently
            const tasks = [
                createLockTask(1),
                createLockTask(2),
                createLockTask(3)
            ];
            
            await Promise.all(tasks.map(task => task()));
            
            // All tasks should have executed
            expect(executionOrder.length).toBe(3);
            expect(executionOrder).toContain(1);
            expect(executionOrder).toContain(2);
            expect(executionOrder).toContain(3);
        });

        it('should maintain order of lock requests', async () => {
            const results: string[] = [];
            
            // Acquire first lock
            const firstUnlock = await mutex.lock();
            
            // Queue several lock requests
            const secondPromise = mutex.lock().then(unlock => {
                results.push('second');
                unlock();
            });
            
            const thirdPromise = mutex.lock().then(unlock => {
                results.push('third');
                unlock();
            });
            
            const fourthPromise = mutex.lock().then(unlock => {
                results.push('fourth');
                unlock();
            });
            
            // Release first lock
            firstUnlock();
            
            // Wait for all to complete
            await Promise.all([secondPromise, thirdPromise, fourthPromise]);
            
            expect(results).toEqual(['second', 'third', 'fourth']);
        });
    });

    describe('error handling', () => {
        it('should handle unlock called multiple times', async () => {
            const unlock = await mutex.lock();
            
            // First unlock should work
            expect(() => unlock()).not.toThrow();
            
            // Second unlock should not cause issues
            expect(() => unlock()).not.toThrow();
        });

        it('should continue working after errors in locked code', async () => {
            try {
                const unlock = await mutex.lock();
                try {
                    throw new Error('Test error');
                } finally {
                    unlock();
                }
            } catch (e) {
                // Expected error
            }
            
            // Mutex should still work after error
            const unlock2 = await mutex.lock();
            expect(typeof unlock2).toBe('function');
            unlock2();
        });
    });

    describe('performance', () => {
        it('should handle rapid lock/unlock cycles', async () => {
            const iterations = 100;
            
            for (let i = 0; i < iterations; i++) {
                const unlock = await mutex.lock();
                unlock();
            }
            
            // If we get here without hanging, the test passes
            expect(true).toBe(true);
        });

        it('should handle many concurrent lock requests', async () => {
            const concurrentRequests = 50;
            let completedRequests = 0;
            
            const promises = Array.from({ length: concurrentRequests }, async (_, index) => {
                const unlock = await mutex.lock();
                completedRequests++;
                // Small delay to ensure some contention
                await new Promise(resolve => setTimeout(resolve, 1));
                unlock();
            });
            
            await Promise.all(promises);
            expect(completedRequests).toBe(concurrentRequests);
        });
    });

    describe('edge cases', () => {
        it('should work with async operations inside lock', async () => {
            let sharedResource = 0;
            
            const task1 = async () => {
                const unlock = await mutex.lock();
                const initialValue = sharedResource;
                await new Promise(resolve => setTimeout(resolve, 10));
                sharedResource = initialValue + 1;
                unlock();
            };
            
            const task2 = async () => {
                const unlock = await mutex.lock();
                const initialValue = sharedResource;
                await new Promise(resolve => setTimeout(resolve, 10));
                sharedResource = initialValue + 1;
                unlock();
            };
            
            await Promise.all([task1(), task2()]);
            expect(sharedResource).toBe(2);
        });

        it('should work with immediate unlock', async () => {
            const unlock = await mutex.lock();
            unlock(); // Immediate unlock
            
            // Should be able to acquire lock again immediately
            const unlock2 = await mutex.lock();
            unlock2();
        });

        it('should handle zero-delay operations', async () => {
            const results: number[] = [];
            
            const createTask = (value: number) => async () => {
                const unlock = await mutex.lock();
                results.push(value);
                unlock();
            };
            
            await Promise.all([
                createTask(1)(),
                createTask(2)(),
                createTask(3)()
            ]);
            
            expect(results.length).toBe(3);
            expect(results).toContain(1);
            expect(results).toContain(2);
            expect(results).toContain(3);
        });
    });
});
