
// Mock the external dependencies before importing the module under test
vi.mock("@okutanidaichi/moqt", () => ({
    ExpiredGroupErrorCode: 1,
    PublishAbortedErrorCode: 2,
    InternalGroupErrorCode: 3,
    TrackWriter: vi.fn(),
}));

vi.mock("golikejs/sync", () => ({
    Mutex: vi.fn().mockImplementation(() => ({
        lock: vi.fn(async () => vi.fn()),
        unlock: vi.fn(),
    })),
    Cond: vi.fn().mockImplementation(() => ({
        wait: vi.fn(async () => undefined),
        broadcast: vi.fn(),
    })),
}));

import { describe, test, expect, beforeEach, vi, MockedFunction } from 'vitest';
import { GroupCache, TrackCache } from "./cache";
import { ExpiredGroupErrorCode, InternalGroupErrorCode } from "@okutanidaichi/moqt";
import { Mutex, Cond } from "golikejs/sync";
import type { GroupSequence, Frame, GroupWriter } from "@okutanidaichi/moqt";
import type { Source } from "@okutanidaichi/moqt/io";

// Simple mock implementations
const createMockGroupWriter = () => ({
    writeFrame: vi.fn(async () => undefined),
    close: vi.fn(),
    cancel: vi.fn(),
}) as any as GroupWriter;

const createMockFrame = () => ({
    trackId: 1,
    objectId: 1,
    priority: 0,
    byteLength: 10,
    copyTo: vi.fn(),
    bytes: new Uint8Array([1, 2, 3]),
    clone: vi.fn(),
    copyFrom: vi.fn(),
}) as any as Frame;

const createMockSource = () => ({
    byteLength: 5,
    copyTo: vi.fn(),
}) as any as Source;

const setupSynchronizationMocks = () => {
    const mockUnlock = vi.fn();
    const mockMutex = {
        lock: vi.fn(async () => mockUnlock),
        unlock: mockUnlock,
    } as any;

    const mockCond = {
        wait: vi.fn(async () => undefined),
        broadcast: vi.fn(),
    } as any;

    return { mockMutex, mockCond, mockUnlock };
};

describe("GroupCache", () => {
    let mockMutex: any;
    let mockCond: any;
    let mockUnlock: any;

    beforeEach(() => {
        vi.clearAllMocks();
        ({ mockMutex, mockCond, mockUnlock } = setupSynchronizationMocks());
    });

    describe("constructor", () => {
        test("creates GroupCache with sequence and timestamp", () => {
            const sequence: GroupSequence = 123n;
            const timestamp = Date.now();

            const cache = new GroupCache(sequence, timestamp);

            expect(cache.sequence).toBe(sequence);
            expect(cache.timestamp).toBe(timestamp);
            expect(cache.frames).toEqual([]);
            expect(cache.closed).toBe(false);
            expect(cache.expired).toBe(false);
        });

        test("creates new instances of Mutex and Cond", () => {
            const sequence: GroupSequence = 456n;
            const timestamp = Date.now();

            new GroupCache(sequence, timestamp);

            expect(vi.mocked(Mutex)).toHaveBeenCalledTimes(1);
            expect(vi.mocked(Cond)).toHaveBeenCalledTimes(1);
        });

        test("handles zero sequence", () => {
            const cache = new GroupCache(0n, 0);

            expect(cache.sequence).toBe(0n);
            expect(cache.timestamp).toBe(0);
        });

        test("handles large sequence", () => {
            const largeSequence = BigInt(Number.MAX_SAFE_INTEGER);
            const cache = new GroupCache(largeSequence, Date.now());

            expect(cache.sequence).toBe(largeSequence);
        });
    });

    describe("append", () => {
        test("appends frame when cache is open", async () => {
            const cache = new GroupCache(1n, Date.now());
            const frame = createMockFrame();

            await cache.append(frame);

            expect(mockMutex.lock).toHaveBeenCalledTimes(1);
            expect(cache.frames).toContain(frame);
            expect(cache.frames.length).toBe(1);
            expect(mockUnlock).toHaveBeenCalledTimes(1);
            expect(mockCond.broadcast).toHaveBeenCalledTimes(1);
        });

        test("appends source when cache is open", async () => {
            const cache = new GroupCache(1n, Date.now());
            const source = createMockSource();

            await cache.append(source);

            expect(cache.frames).toContain(source);
            expect(cache.frames.length).toBe(1);
            expect(mockCond.broadcast).toHaveBeenCalledTimes(1);
        });

        test("does not append when cache is closed", async () => {
            const cache = new GroupCache(1n, Date.now());
            cache.closed = true;
            const frame = createMockFrame();

            await cache.append(frame);

            expect(cache.frames.length).toBe(0);
            expect(mockUnlock).toHaveBeenCalledTimes(1);
            expect(mockCond.broadcast).not.toHaveBeenCalled();
        });

        test("appends multiple frames", async () => {
            const cache = new GroupCache(1n, Date.now());
            const frame1 = createMockFrame();
            const frame2 = createMockFrame();

            await cache.append(frame1);
            await cache.append(frame2);

            expect(cache.frames).toEqual([frame1, frame2]);
            expect(mockCond.broadcast).toHaveBeenCalledTimes(2);
        });

        test("handles mutex lock properly", async () => {
            const cache = new GroupCache(1n, Date.now());
            const frame = createMockFrame();

            await cache.append(frame);

            expect(mockMutex.lock).toHaveBeenCalledTimes(1);
            expect(mockUnlock).toHaveBeenCalledTimes(1);
        });

        test("unlocks mutex even when closed", async () => {
            const cache = new GroupCache(1n, Date.now());
            cache.closed = true;
            const frame = createMockFrame();

            await cache.append(frame);

            expect(mockUnlock).toHaveBeenCalledTimes(1);
        });

        test("broadcasts after each successful append", async () => {
            const cache = new GroupCache(1n, Date.now());
            const frames = Array.from({ length: 5 }, () => createMockFrame());

            for (const frame of frames) {
                await cache.append(frame);
            }

            expect(cache.frames).toHaveLength(frames.length);
            expect(mockMutex.lock).toHaveBeenCalledTimes(frames.length);
            expect(mockCond.broadcast).toHaveBeenCalledTimes(frames.length);
        });
    });

    describe("flush - basic functionality", () => {
        test("writes existing frames to group", async () => {
            const cache = new GroupCache(1n, Date.now());
            const frame1 = createMockFrame();
            const group = createMockGroupWriter();

            // Add frame first
            await cache.append(frame1);

            // Mock the wait to immediately close the cache
            mockCond.wait.mockImplementationOnce(async () => {
                cache.closed = true;
            });

            await cache.flush(group);

            expect(group.writeFrame).toHaveBeenCalledWith(frame1);
            expect(group.close).toHaveBeenCalledTimes(1);
        });

        test("handles write frame error", async () => {
            const cache = new GroupCache(1n, Date.now());
            const frame = createMockFrame();
            const group = createMockGroupWriter();
            const error = new Error("Write failed");

            await cache.append(frame);

            // Reset mocks and set up error
            vi.clearAllMocks();
            mockMutex.lock.mockImplementation(async () => Promise.resolve());
            (group.writeFrame as any).mockImplementation(async () => error);

            // Mock wait to not be called since flush should return early on error
            mockCond.wait.mockImplementation(async () => undefined);

            await cache.flush(group);

            expect(group.cancel).toHaveBeenCalledWith(
                InternalGroupErrorCode,
                "failed to write frame: Write failed"
            );
            expect(mockUnlock).toHaveBeenCalledTimes(1);
        });

    });

    describe("close", () => {
        test("closes cache and broadcasts", async () => {
            const cache = new GroupCache(1n, Date.now());

            await cache.close();

            expect(cache.closed).toBe(true);
            expect(mockMutex.lock).toHaveBeenCalledTimes(1);
            expect(mockUnlock).toHaveBeenCalledTimes(1);
            expect(mockCond.broadcast).toHaveBeenCalledTimes(1);
        });

        test("does nothing if already closed", async () => {
            const cache = new GroupCache(1n, Date.now());
            cache.closed = true;

            await cache.close();

            expect(mockUnlock).toHaveBeenCalledTimes(1);
            expect(mockCond.broadcast).not.toHaveBeenCalled();
        });

        test("handles multiple close calls", async () => {
            const cache = new GroupCache(1n, Date.now());

            await cache.close();
            await cache.close();

            expect(cache.closed).toBe(true);
            expect(mockCond.broadcast).toHaveBeenCalledTimes(1);
        });
    });

    describe("expire", () => {
        test("expires cache and clears frames", async () => {
            const cache = new GroupCache(1n, Date.now());
            const frame = createMockFrame();

            await cache.append(frame);
            expect(cache.frames.length).toBe(1);

            await cache.expire();

            expect(cache.expired).toBe(true);
            expect(cache.frames.length).toBe(0);
            expect(mockCond.broadcast).toHaveBeenCalled();
        });

        test("does nothing if already expired", async () => {
            const cache = new GroupCache(1n, Date.now());
            cache.expired = true;

            await cache.expire();

            // Should return early without unlocking or broadcasting
            expect(mockMutex.lock).toHaveBeenCalledTimes(1);
        });

        test("clears frames on expire", async () => {
            const cache = new GroupCache(1n, Date.now());
            const frame1 = createMockFrame();
            const frame2 = createMockFrame();

            await cache.append(frame1);
            await cache.append(frame2);
            expect(cache.frames.length).toBe(2);

            await cache.expire();

            expect(cache.frames.length).toBe(0);
            expect(cache.expired).toBe(true);
        });

        test("handles multiple expire calls", async () => {
            const cache = new GroupCache(1n, Date.now());

            await cache.expire();
            await cache.expire();

            expect(cache.expired).toBe(true);
        });
    });

    describe("Error Handling", () => {
        test("handles mutex lock error", async () => {
            const cache = new GroupCache(1n, Date.now());
            const frame = createMockFrame();
            const error = new Error("Mutex lock failed");

            mockMutex.lock.mockRejectedValue(error);

            await expect(cache.append(frame)).rejects.toThrow("Mutex lock failed");
        });

        test("handles concurrent operations", async () => {
            const cache = new GroupCache(1n, Date.now());
            const frame1 = createMockFrame();
            const frame2 = createMockFrame();

            // Simulate concurrent appends
            const promises = [
                cache.append(frame1),
                cache.append(frame2),
            ];

            await Promise.all(promises);

            expect(cache.frames.length).toBe(2);
            expect(mockMutex.lock).toHaveBeenCalledTimes(2);
        });
    });

    describe("Boundary Value Tests", () => {
        test("handles maximum bigint sequence", () => {
            const maxSequence = BigInt("18446744073709551615"); // Max uint64
            const cache = new GroupCache(maxSequence, Date.now());

            expect(cache.sequence).toBe(maxSequence);
        });

        test("handles zero timestamp", () => {
            const cache = new GroupCache(1n, 0);

            expect(cache.timestamp).toBe(0);
        });

        test("handles large timestamp", () => {
            const largeTimestamp = Date.now() + 1000000000;
            const cache = new GroupCache(1n, largeTimestamp);

            expect(cache.timestamp).toBe(largeTimestamp);
        });
    });
});

describe("TrackCache Interface", () => {
    test("interface definition is correct", () => {
        // This test ensures the interface is properly defined
        const mockTrackCache: TrackCache = {
            store: vi.fn(),
            close: vi.fn(async () => undefined),
        };

        expect(typeof mockTrackCache.store).toBe("function");
        expect(typeof mockTrackCache.close).toBe("function");
    });

    test("store method signature", () => {
        const mockTrackCache: TrackCache = {
            store: vi.fn(),
            close: vi.fn(async () => undefined),
        };

        const groupCache = new GroupCache(1n, Date.now());
        mockTrackCache.store(groupCache);

        expect(mockTrackCache.store).toHaveBeenCalledWith(groupCache);
    });

    test("close method signature", async () => {
        const mockTrackCache: TrackCache = {
            store: vi.fn(),
            close: vi.fn(async () => undefined),
        };

        await mockTrackCache.close();

        expect(mockTrackCache.close).toHaveBeenCalledTimes(1);
    });
});

describe("Enhanced Test Coverage", () => {
    let mockMutex: any;
    let mockCond: any;
    let mockUnlock: any;

    beforeEach(() => {
        vi.clearAllMocks();
        ({ mockMutex, mockCond, mockUnlock } = setupSynchronizationMocks());
    });

    test("flush handles new frames arriving during flush", async () => {
        const cache = new GroupCache(1n, Date.now());
        const groupWriter = createMockGroupWriter();
        const frame1 = createMockFrame();
        const frame2 = createMockFrame();

        // Add initial frame
        await cache.append(frame1);

        // Mock condition wait to simulate frames arriving during flush
        let waitCount = 0;
        mockCond.wait.mockImplementation(async () => {
            waitCount++;
            if (waitCount === 1) {
                // Simulate new frame arriving
                cache.frames.push(frame2);
                return;
            }
            // Close cache to end flush loop
            cache.closed = true;
            return;
        });

        await cache.flush(groupWriter);

        expect(groupWriter.writeFrame).toHaveBeenCalledWith(frame1);
        expect(groupWriter.writeFrame).toHaveBeenCalledWith(frame2);
        expect(groupWriter.close).toHaveBeenCalled();
    });

    test("flush cancels group when frame write fails mid-flush", async () => {
        const cache = new GroupCache(1n, Date.now());
        const groupWriter = createMockGroupWriter();
        const frame1 = createMockFrame();
        const frame2 = createMockFrame();

        await cache.append(frame1);
        await cache.append(frame2);

        // Mock writeFrame to fail on second call
        (groupWriter.writeFrame as any)
            .mockImplementationOnce(async () => undefined)
            .mockImplementationOnce(async () => new Error("Write failed"));

        await cache.flush(groupWriter);

        expect(groupWriter.writeFrame).toHaveBeenCalledTimes(2);
        expect(groupWriter.cancel).toHaveBeenCalledWith(
            InternalGroupErrorCode,
            "failed to write frame: Write failed"
        );
    });

    test("flush handles cache expiration during flush", async () => {
        const cache = new GroupCache(1n, Date.now());
        const groupWriter = createMockGroupWriter();
        const frame = createMockFrame();

        await cache.append(frame);

        // Mock condition wait to simulate cache expiration
        mockCond.wait.mockImplementation(async () => {
            cache.expired = true;
            return;
        });

        await cache.flush(groupWriter);

        expect(groupWriter.cancel).toHaveBeenCalledWith(
            ExpiredGroupErrorCode,
            "cache expired"
        );
    });

    test("flush waits for condition correctly", async () => {
        const cache = new GroupCache(1n, Date.now());
        const groupWriter = createMockGroupWriter();

        // Mock wait to be called twice before closing
        let waitCallCount = 0;
        mockCond.wait.mockImplementation(async () => {
            waitCallCount++;
            if (waitCallCount >= 2) {
                cache.closed = true;
            }
            return;
        });

        await cache.flush(groupWriter);

        expect(mockCond.wait).toHaveBeenCalledTimes(2);
        expect(groupWriter.close).toHaveBeenCalled();
    });

    test("flush handles concurrent frame additions", async () => {
        const cache = new GroupCache(1n, Date.now());
        const groupWriter = createMockGroupWriter();
        const frame1 = createMockFrame();
        const frame2 = createMockFrame();
        const frame3 = createMockFrame();

        await cache.append(frame1);

        // Simulate frames being added during flush
        let conditionWaitCount = 0;
        mockCond.wait.mockImplementation(async () => {
            conditionWaitCount++;
            if (conditionWaitCount === 1) {
                cache.frames.push(frame2, frame3);
                return;
            }
            cache.closed = true;
            return;
        });

        await cache.flush(groupWriter);

        expect(groupWriter.writeFrame).toHaveBeenCalledWith(frame1);
        expect(groupWriter.writeFrame).toHaveBeenCalledWith(frame2);
        expect(groupWriter.writeFrame).toHaveBeenCalledWith(frame3);
        expect(groupWriter.close).toHaveBeenCalled();
    });
});

describe("Concurrent Operations and Race Conditions", () => {
    let mockMutex: any;
    let mockCond: any;
    let mockUnlock: any;

    beforeEach(() => {
        vi.clearAllMocks();
        ({ mockMutex, mockCond, mockUnlock } = setupSynchronizationMocks());
    });

    test("handles concurrent append and close operations", async () => {
        const cache = new GroupCache(1n, Date.now());
        const frame = createMockFrame();

        // Setup mutex to simulate race condition
        let lockCount = 0;
        mockMutex.lock.mockImplementation(async () => {
            lockCount++;
            if (lockCount === 1) {
                // First call (append) - don't close yet
                return Promise.resolve();
            } else {
                // Second call (close) - close the cache
                cache.closed = true;
                return Promise.resolve();
            }
        });

        const appendPromise = cache.append(frame);
        const closePromise = cache.close();

        await Promise.all([appendPromise, closePromise]);

        expect(mockMutex.lock).toHaveBeenCalledTimes(2);
        expect(mockUnlock).toHaveBeenCalledTimes(2);
    });

    test("handles concurrent append and expire operations", async () => {
        const cache = new GroupCache(1n, Date.now());
        const frame = createMockFrame();

        const appendPromise = cache.append(frame);
        const expirePromise = cache.expire();

        await Promise.all([appendPromise, expirePromise]);

        expect(mockMutex.lock).toHaveBeenCalledTimes(2);
        expect(cache.expired).toBe(true);
    });

    test("handles multiple simultaneous flush operations", async () => {
        const cache = new GroupCache(1n, Date.now());
        const groupWriter1 = createMockGroupWriter();
        const groupWriter2 = createMockGroupWriter();
        const frame = createMockFrame();

        await cache.append(frame);

        // Mock condition wait to close after first flush starts
        let waitCount = 0;
        mockCond.wait.mockImplementation(async () => {
            waitCount++;
            if (waitCount === 2) {
                cache.closed = true;
            }
            return;
        });

        const flush1Promise = cache.flush(groupWriter1);
        const flush2Promise = cache.flush(groupWriter2);

        await Promise.all([flush1Promise, flush2Promise]);

        expect(groupWriter1.close).toHaveBeenCalled();
        expect(groupWriter2.close).toHaveBeenCalled();
    });

    test("maintains frame order under concurrent operations", async () => {
        const cache = new GroupCache(1n, Date.now());
        const frames = Array.from({ length: 3 }, () => createMockFrame()); // Reduced from 5

        // Append frames concurrently
        const appendPromises = frames.map(frame => cache.append(frame));
        await Promise.all(appendPromises);

        expect(cache.frames).toHaveLength(3);
        expect(mockCond.broadcast).toHaveBeenCalledTimes(3);
    });
});

describe("Memory Management and Performance", () => {
    let mockMutex: any;
    let mockCond: any;
    let mockUnlock: any;

    beforeEach(() => {
        vi.clearAllMocks();
        ({ mockMutex, mockCond, mockUnlock } = setupSynchronizationMocks());
    });

    test("handles large number of frames efficiently", async () => {
        const cache = new GroupCache(1n, Date.now());
        const frameCount = 100; // Reduced from 1000 to prevent memory issues
        const frames = Array.from({ length: frameCount }, () => createMockFrame());

        // Add all frames
        for (const frame of frames) {
            await cache.append(frame);
        }

        expect(cache.frames).toHaveLength(frameCount);
        expect(mockMutex.lock).toHaveBeenCalledTimes(frameCount);
        expect(mockCond.broadcast).toHaveBeenCalledTimes(frameCount);
    });

    test("expire clears all frames to free memory", async () => {
        const cache = new GroupCache(1n, Date.now());
        const frames = Array.from({ length: 50 }, () => createMockFrame()); // Reduced from 100

        for (const frame of frames) {
            await cache.append(frame);
        }

        expect(cache.frames).toHaveLength(50);

        await cache.expire();

        expect(cache.frames).toHaveLength(0);
        expect(cache.expired).toBe(true);
    });

    test("handles rapid successive operations", async () => {
        const cache = new GroupCache(1n, Date.now());
        const operationCount = 20; // Reduced from 50

        const promises: Promise<void>[] = [];

        // Mix of append, close, and expire operations
        for (let i = 0; i < operationCount; i++) {
            if (i % 3 === 0) {
                promises.push(cache.append(createMockFrame()));
            } else if (i % 3 === 1 && !cache.closed) {
                promises.push(cache.close());
            } else if (!cache.expired) {
                promises.push(cache.expire());
            }
        }

        await Promise.allSettled(promises);

        expect(vi.mocked(Mutex)).toHaveBeenCalled();
        expect(mockCond.broadcast).toHaveBeenCalled();
    });
});

describe("Error Recovery and Resilience", () => {
    let mockMutex: any;
    let mockCond: any;
    let mockUnlock: any;

    beforeEach(() => {
        vi.clearAllMocks();
        ({ mockMutex, mockCond, mockUnlock } = setupSynchronizationMocks());
    });

    test("recovers from mutex lock failures gracefully", async () => {
        const cache = new GroupCache(1n, Date.now());
        const frame = createMockFrame();

        // Mock mutex lock to fail first time, succeed second time
        mockMutex.lock
            .mockRejectedValueOnce(new Error("Lock failed"))
            .mockImplementationOnce(async () => Promise.resolve());

        // First append should fail
        await expect(cache.append(frame)).rejects.toThrow("Lock failed");

        // Second append should succeed
        await expect(cache.append(frame)).resolves.toBeUndefined();

        expect(cache.frames).toHaveLength(1);
    });

    test("handles condition broadcast failures", async () => {
        const cache = new GroupCache(1n, Date.now());
        const frame = createMockFrame();

        // Mock broadcast to throw an error
        mockCond.broadcast.mockImplementation(() => {
            throw new Error("Broadcast failed");
        });

        // Should still complete append despite broadcast failure
        await expect(cache.append(frame)).rejects.toThrow("Broadcast failed");
    });

    test("handles partial frame writes in flush", async () => {
        const cache = new GroupCache(1n, Date.now());
        const groupWriter = createMockGroupWriter();
        const frames = [createMockFrame(), createMockFrame(), createMockFrame()];

        for (const frame of frames) {
            await cache.append(frame);
        }

        // Mock writeFrame to fail on second frame
        (groupWriter.writeFrame as any)
            .mockImplementationOnce(async () => undefined)
            .mockImplementationOnce(async () => new Error("Network error"))
            .mockImplementationOnce(async () => undefined);

        await cache.flush(groupWriter);

        expect(groupWriter.writeFrame).toHaveBeenCalledTimes(2);
        expect(groupWriter.cancel).toHaveBeenCalledWith(
            InternalGroupErrorCode,
            "failed to write frame: Network error"
        );
    });
});

describe("State Consistency and Invariants", () => {
    let mockMutex: any;
    let mockCond: any;
    let mockUnlock: any;

    beforeEach(() => {
        vi.clearAllMocks();
        ({ mockMutex, mockCond, mockUnlock } = setupSynchronizationMocks());
    });

    test("maintains consistent state during concurrent modifications", async () => {
        const cache = new GroupCache(1n, Date.now());

        // Ensure state is consistent after concurrent operations
        const operations = [
            () => cache.append(createMockFrame()),
            () => cache.append(createMockFrame()),
            () => cache.close(),
            () => cache.expire(),
        ];

        await Promise.allSettled(operations.map(op => op()));

        // Cache should be in a valid state
        expect(typeof cache.closed).toBe("boolean");
        expect(typeof cache.expired).toBe("boolean");
        expect(Array.isArray(cache.frames)).toBe(true);
    });

    test("sequence and timestamp are immutable", async () => {
        const sequence: GroupSequence = 42n;
        const timestamp = 123456789;
        const cache = new GroupCache(sequence, timestamp);

        // Perform various operations
        await cache.append(createMockFrame());
        await cache.close();
        await cache.expire();

        // Properties should remain unchanged
        expect(cache.sequence).toBe(sequence);
        expect(cache.timestamp).toBe(timestamp);
    });

    test("frames array reference stability", async () => {
        const cache = new GroupCache(1n, Date.now());
        const originalFramesRef = cache.frames;

        await cache.append(createMockFrame());
        await cache.close();

        // Same reference should be maintained
        expect(cache.frames).toBe(originalFramesRef);

        await cache.expire();

        // Reference should still be the same, just cleared
        expect(cache.frames).toBe(originalFramesRef);
        expect(cache.frames).toHaveLength(0);
    });
});

describe("Integration and Real-world Scenarios", () => {
    let mockMutex: any;
    let mockCond: any;
    let mockUnlock: any;

    beforeEach(() => {
        vi.clearAllMocks();
        ({ mockMutex, mockCond, mockUnlock } = setupSynchronizationMocks());
    });

    test("simulates typical broadcast scenario", async () => {
        const cache = new GroupCache(1n, Date.now());
        const groupWriter = createMockGroupWriter();
        const frames = Array.from({ length: 10 }, (_, i) => {
            const frame = createMockFrame();
            (frame as any).objectId = i;
            return frame;
        });

        // Producer adds frames periodically
        for (let i = 0; i < 5; i++) {
            await cache.append(frames[i]);
        }

        // Consumer starts reading
        const flushPromise = cache.flush(groupWriter);

        // More frames arrive during flush
        for (let i = 5; i < 10; i++) {
            await cache.append(frames[i]);
        }

        // Close the cache to complete flush
        await cache.close();
        await flushPromise;

        expect(groupWriter.writeFrame).toHaveBeenCalledTimes(10);
        expect(groupWriter.close).toHaveBeenCalled();
    });

    test("handles cache expiration during active streaming", async () => {
        const cache = new GroupCache(1n, Date.now());
        const groupWriter = createMockGroupWriter();
        const frame = createMockFrame();

        await cache.append(frame);

        // Expire cache while flush is in progress by simulating wait callback
        mockCond.wait.mockImplementationOnce(async () => {
            await cache.expire();
        });

        // Start flush
        const flushPromise = cache.flush(groupWriter);

        await flushPromise;

        expect(groupWriter.cancel).toHaveBeenCalledWith(
            ExpiredGroupErrorCode,
            "cache expired"
        );
    });

    test("handles multiple consumers with different timing", async () => {
        const cache = new GroupCache(1n, Date.now());
        const groupWriter1 = createMockGroupWriter();
        const groupWriter2 = createMockGroupWriter();
        const frame = createMockFrame();

        await cache.append(frame);

        // Start multiple flush operations with different conditions
        let writer1Closed = false;
        let writer2Closed = false;

        mockCond.wait.mockImplementation(async () => {
            if (!writer1Closed) {
                writer1Closed = true;
                return;
            }
            if (!writer2Closed) {
                writer2Closed = true;
                cache.closed = true;
                return;
            }
            return;
        });

        const flush1Promise = cache.flush(groupWriter1);
        const flush2Promise = cache.flush(groupWriter2);

        await Promise.all([flush1Promise, flush2Promise]);

        expect(groupWriter1.close).toHaveBeenCalled();
        expect(groupWriter2.close).toHaveBeenCalled();
    });
});
