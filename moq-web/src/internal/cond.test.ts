import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { Cond } from './cond';

describe('Cond', () => {
    it('should create a new condition variable', () => {
        const cond = new Cond();
        expect(cond).toBeInstanceOf(Cond);
    });

    it('should wait and broadcast correctly', async () => {
        const cond = new Cond();
        let resolved = false;

        // Start waiting
        const waitPromise = cond.wait().then(() => {
            resolved = true;
        });

        // Should not be resolved yet
        expect(resolved).toBe(false);

        // Broadcast to resolve waiting
        cond.broadcast();

        // Wait for the promise to resolve
        await waitPromise;
        expect(resolved).toBe(true);
    });

    it('should allow multiple waiters', async () => {
        const cond = new Cond();
        let resolved1 = false;
        let resolved2 = false;
        let resolved3 = false;

        // Start multiple waiters
        const wait1 = cond.wait().then(() => { resolved1 = true; });
        const wait2 = cond.wait().then(() => { resolved2 = true; });
        const wait3 = cond.wait().then(() => { resolved3 = true; });

        // None should be resolved yet
        expect(resolved1).toBe(false);
        expect(resolved2).toBe(false);
        expect(resolved3).toBe(false);

        // Broadcast should resolve all waiters
        cond.broadcast();

        await Promise.all([wait1, wait2, wait3]);

        expect(resolved1).toBe(true);
        expect(resolved2).toBe(true);
        expect(resolved3).toBe(true);
    });

    it('should reset after broadcast for new waiters', async () => {
        const cond = new Cond();
        let firstResolved = false;
        let secondResolved = false;

        // First wait and broadcast
        const firstWait = cond.wait().then(() => { firstResolved = true; });
        cond.broadcast();
        await firstWait;
        expect(firstResolved).toBe(true);

        // Second wait should wait for new broadcast
        const secondWait = cond.wait().then(() => { secondResolved = true; });
        
        // Should not be resolved immediately
        expect(secondResolved).toBe(false);

        // Need another broadcast for second waiter
        cond.broadcast();
        await secondWait;
        expect(secondResolved).toBe(true);
    });

    it('should handle immediate broadcast before wait', async () => {
        const cond = new Cond();
        
        // Broadcast first, then wait
        cond.broadcast();
        
        // This should not resolve immediately since broadcast resets the promise
        let resolved = false;
        const waitPromise = cond.wait().then(() => { resolved = true; });
        
        // Should not be resolved yet
        expect(resolved).toBe(false);
        
        // Need another broadcast
        cond.broadcast();
        await waitPromise;
        expect(resolved).toBe(true);
    });

    it('should handle rapid broadcast calls', async () => {
        const cond = new Cond();
        let resolveCount = 0;

        // Multiple rapid broadcasts
        cond.broadcast();
        cond.broadcast();
        cond.broadcast();

        // Start waiting after broadcasts
        const waitPromise = cond.wait().then(() => { resolveCount++; });

        // Should not be resolved yet
        expect(resolveCount).toBe(0);

        // One more broadcast to resolve
        cond.broadcast();
        await waitPromise;
        expect(resolveCount).toBe(1);
    });
});
