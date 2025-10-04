import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';

// Mock only the external MOQT dependencies, not internal sync primitives
vi.mock("@okutanidaichi/moqt", () => ({
    ExpiredGroupErrorCode: 1,
    PublishAbortedErrorCode: 2,
    InternalGroupErrorCode: 3,
    TrackWriter: vi.fn(),
}));

// DO NOT mock golikejs/sync - use the real Mutex and Cond implementations!
// This gives us real synchronization behavior and catches integration bugs.

import { GroupCache, TrackCache } from "./cache";
import { ExpiredGroupErrorCode, InternalGroupErrorCode } from "@okutanidaichi/moqt";
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

describe("GroupCache - Using Real Mutex/Cond", () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    afterEach(() => {
        vi.clearAllMocks();
    });

    describe("constructor", () => {
        it("should create GroupCache with sequence and timestamp", () => {
            const sequence: GroupSequence = 123n;
            const timestamp = Date.now();

            const cache = new GroupCache(sequence, timestamp);

            expect(cache.sequence).toBe(sequence);
            expect(cache.timestamp).toBe(timestamp);
            expect(cache.frames).toEqual([]);
            expect(cache.closed).toBe(false);
            expect(cache.expired).toBe(false);
        });

        it("should handle zero sequence", () => {
            const cache = new GroupCache(0n, Date.now());
            expect(cache.sequence).toBe(0n);
        });

        it("should handle large sequence", () => {
            const largeSequence = BigInt(Number.MAX_SAFE_INTEGER);
            const cache = new GroupCache(largeSequence, Date.now());
            expect(cache.sequence).toBe(largeSequence);
        });

        it("should handle zero timestamp", () => {
            const cache = new GroupCache(1n, 0);
            expect(cache.timestamp).toBe(0);
        });

        it("should handle large timestamp", () => {
            const largeTimestamp = Date.now() + 1000000000;
            const cache = new GroupCache(1n, largeTimestamp);
            expect(cache.timestamp).toBe(largeTimestamp);
        });
    });

    describe("append", () => {
        it("should append frame when cache is open", async () => {
            const cache = new GroupCache(1n, Date.now());
            const frame = createMockFrame();

            await cache.append(frame);

            expect(cache.frames).toHaveLength(1);
            expect(cache.frames[0]).toBe(frame);
        });

        it("should append source when cache is open", async () => {
            const cache = new GroupCache(1n, Date.now());
            const source = createMockSource();

            await cache.append(source);

            expect(cache.frames).toHaveLength(1);
            expect(cache.frames[0]).toBe(source);
        });

        it("should not append when cache is closed", async () => {
            const cache = new GroupCache(1n, Date.now());
            const frame = createMockFrame();

            await cache.close();
            await cache.append(frame);

            expect(cache.frames).toHaveLength(0);
        });

        it("should append multiple frames in order", async () => {
            const cache = new GroupCache(1n, Date.now());
            const frame1 = createMockFrame();
            const frame2 = createMockFrame();
            const frame3 = createMockFrame();

            await cache.append(frame1);
            await cache.append(frame2);
            await cache.append(frame3);

            expect(cache.frames).toHaveLength(3);
            expect(cache.frames[0]).toBe(frame1);
            expect(cache.frames[1]).toBe(frame2);
            expect(cache.frames[2]).toBe(frame3);
        });

        it("should handle concurrent append operations with real Mutex", async () => {
            const cache = new GroupCache(1n, Date.now());
            const frames = [createMockFrame(), createMockFrame(), createMockFrame()];

            // Real Mutex should handle concurrent appends correctly
            await Promise.all(frames.map(f => cache.append(f)));

            expect(cache.frames).toHaveLength(3);
        });
    });

    describe("flush with real Cond", () => {
        it("should write initial frames and wait for close signal", async () => {
            const cache = new GroupCache(1n, Date.now());
            const groupWriter = createMockGroupWriter();
            const frame1 = createMockFrame();
            const frame2 = createMockFrame();

            await cache.append(frame1);
            await cache.append(frame2);

            // Start flush in background - it will wait on real Cond
            const flushPromise = cache.flush(groupWriter);
            
            // Give flush time to write initial frames and start waiting
            await new Promise(resolve => setTimeout(resolve, 10));

            // Verify initial frames were written
            expect(groupWriter.writeFrame).toHaveBeenCalledWith(frame1);
            expect(groupWriter.writeFrame).toHaveBeenCalledWith(frame2);

            // Close to trigger Cond.broadcast() and end flush
            await cache.close();
            
            // Wait for flush to complete
            await flushPromise;

            expect(groupWriter.close).toHaveBeenCalled();
        });

        it("should write new frames added during flush", async () => {
            const cache = new GroupCache(1n, Date.now());
            const groupWriter = createMockGroupWriter();
            const frame1 = createMockFrame();
            const frame2 = createMockFrame();
            const frame3 = createMockFrame();

            await cache.append(frame1);
            await cache.append(frame2);

            // Start flush
            const flushPromise = cache.flush(groupWriter);
            
            // Wait for initial writes
            await new Promise(resolve => setTimeout(resolve, 10));

            // Add another frame - this should trigger Cond.broadcast()
            await cache.append(frame3);
            
            // Give time for flush to process new frame
            await new Promise(resolve => setTimeout(resolve, 10));

            // Close to end flush
            await cache.close();
            await flushPromise;

            // All three frames should have been written
            expect(groupWriter.writeFrame).toHaveBeenCalledTimes(3);
            expect(groupWriter.close).toHaveBeenCalled();
        });

        it("should handle expire signal during flush", async () => {
            const cache = new GroupCache(1n, Date.now());
            const groupWriter = createMockGroupWriter();
            const frame = createMockFrame();

            await cache.append(frame);

            const flushPromise = cache.flush(groupWriter);
            
            // Wait for initial write
            await new Promise(resolve => setTimeout(resolve, 10));

            // Expire triggers Cond.broadcast()
            await cache.expire();
            
            await flushPromise;

            expect(groupWriter.cancel).toHaveBeenCalledWith(
                ExpiredGroupErrorCode,
                "cache expired"
            );
        });

        it("should cancel when frame write fails", async () => {
            const cache = new GroupCache(1n, Date.now());
            const groupWriter = createMockGroupWriter();
            const frame = createMockFrame();

            await cache.append(frame);

            // Make write fail
            (groupWriter.writeFrame as any).mockResolvedValueOnce(new Error("Write failed"));

            await cache.flush(groupWriter);

            expect(groupWriter.cancel).toHaveBeenCalledWith(
                InternalGroupErrorCode,
                "failed to write frame: Write failed"
            );
        });
    });

    describe("close", () => {
        it("should close cache and broadcast via real Cond", async () => {
            const cache = new GroupCache(1n, Date.now());

            await cache.close();

            expect(cache.closed).toBe(true);
        });

        it("should be idempotent", async () => {
            const cache = new GroupCache(1n, Date.now());

            await cache.close();
            await cache.close();
            await cache.close();

            expect(cache.closed).toBe(true);
        });

        it("should wake up waiting flush when closed", async () => {
            const cache = new GroupCache(1n, Date.now());
            const groupWriter = createMockGroupWriter();

            // Start flush that will wait
            const flushPromise = cache.flush(groupWriter);
            
            await new Promise(resolve => setTimeout(resolve, 10));

            // Close should wake up the waiting flush
            await cache.close();
            
            // Flush should complete
            await flushPromise;

            expect(groupWriter.close).toHaveBeenCalled();
        });
    });

    describe("expire", () => {
        it("should expire cache and clear frames", async () => {
            const cache = new GroupCache(1n, Date.now());
            const frame = createMockFrame();

            await cache.append(frame);
            expect(cache.frames).toHaveLength(1);

            await cache.expire();

            expect(cache.expired).toBe(true);
            expect(cache.frames).toHaveLength(0);
        });

        it("should be idempotent", async () => {
            const cache = new GroupCache(1n, Date.now());

            await cache.expire();
            await cache.expire();

            expect(cache.expired).toBe(true);
        });

        it("should wake up waiting flush when expired", async () => {
            const cache = new GroupCache(1n, Date.now());
            const groupWriter = createMockGroupWriter();

            const flushPromise = cache.flush(groupWriter);
            
            await new Promise(resolve => setTimeout(resolve, 10));

            // Expire should wake up the waiting flush
            await cache.expire();
            
            await flushPromise;

            expect(groupWriter.cancel).toHaveBeenCalledWith(
                ExpiredGroupErrorCode,
                "cache expired"
            );
        });
    });

    describe("State Management", () => {
        it("should handle sequence immutability", async () => {
            const sequence: GroupSequence = 42n;
            const timestamp = Date.now();
            const cache = new GroupCache(sequence, timestamp);

            expect(cache.sequence).toBe(sequence);
            
            await cache.append(createMockFrame());
            expect(cache.sequence).toBe(sequence);
        });

        it("should handle timestamp immutability", async () => {
            const timestamp = 123456789;
            const cache = new GroupCache(1n, timestamp);

            expect(cache.timestamp).toBe(timestamp);
            
            await cache.append(createMockFrame());
            expect(cache.timestamp).toBe(timestamp);
        });

        it("should maintain frames array reference", async () => {
            const cache = new GroupCache(1n, Date.now());
            const originalRef = cache.frames;

            await cache.append(createMockFrame());
            expect(cache.frames).toBe(originalRef);

            await cache.close();
            expect(cache.frames).toBe(originalRef);

            await cache.expire();
            expect(cache.frames).toBe(originalRef);
        });
    });

    describe("Concurrent Operations with Real Synchronization", () => {
        it("should handle concurrent close and append", async () => {
            const cache = new GroupCache(1n, Date.now());
            const frame = createMockFrame();

            const results = await Promise.allSettled([
                cache.append(frame),
                cache.close(),
            ]);

            expect(cache.closed).toBe(true);
            expect(results.every(r => r.status === 'fulfilled')).toBe(true);
        });

        it("should handle concurrent expire and append", async () => {
            const cache = new GroupCache(1n, Date.now());
            const frame = createMockFrame();

            const results = await Promise.allSettled([
                cache.append(frame),
                cache.expire(),
            ]);

            expect(cache.expired).toBe(true);
            expect(results.every(r => r.status === 'fulfilled')).toBe(true);
        });
    });
});

describe("TrackCache Interface", () => {
    it("should have correct store method signature", () => {
        const mockTrackCache: TrackCache = {
            store: vi.fn(),
            close: vi.fn(async () => undefined),
        };

        const groupCache = new GroupCache(1n, Date.now());
        mockTrackCache.store(groupCache);

        expect(mockTrackCache.store).toHaveBeenCalledWith(groupCache);
        expect(mockTrackCache.store).toHaveBeenCalledTimes(1);
    });

    it("should have async close method", async () => {
        const mockTrackCache: TrackCache = {
            store: vi.fn(),
            close: vi.fn(async () => undefined),
        };

        await mockTrackCache.close();

        expect(mockTrackCache.close).toHaveBeenCalledTimes(1);
    });
});

describe("Integration Scenarios with Real Synchronization", () => {
    it("should handle typical broadcast workflow", async () => {
        const cache = new GroupCache(1n, Date.now());
        const groupWriter = createMockGroupWriter();

        // Producer adds frames
        await cache.append(createMockFrame());
        await cache.append(createMockFrame());

        // Consumer starts reading
        const flushPromise = cache.flush(groupWriter);
        
        // Wait for initial writes
        await new Promise(resolve => setTimeout(resolve, 10));

        // More frames arrive - real Cond.broadcast() wakes up flush
        await cache.append(createMockFrame());
        
        await new Promise(resolve => setTimeout(resolve, 10));

        // Close - real Cond.broadcast() signals flush to end
        await cache.close();
        
        await flushPromise;

        expect(groupWriter.writeFrame).toHaveBeenCalledTimes(3);
        expect(groupWriter.close).toHaveBeenCalled();
    });

    it("should handle complex state transitions", async () => {
        const cache = new GroupCache(1n, Date.now());

        // Initial state
        expect(cache.closed).toBe(false);
        expect(cache.expired).toBe(false);
        expect(cache.frames).toHaveLength(0);

        // Add frames
        await cache.append(createMockFrame());
        await cache.append(createMockFrame());
        expect(cache.frames).toHaveLength(2);

        // Close
        await cache.close();
        expect(cache.closed).toBe(true);
        expect(cache.frames).toHaveLength(2);

        // Expire after close
        await cache.expire();
        expect(cache.expired).toBe(true);
        expect(cache.frames).toHaveLength(0);

        // Try to append after close and expire
        await cache.append(createMockFrame());
        expect(cache.frames).toHaveLength(0);
    });
});
