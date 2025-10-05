import { describe, test, expect, vi } from 'vitest';
import { cloneChunk, NoOpTrackEncoder } from './track_encoder';
import type { EncodedChunk } from './container';
import { ContextCancelledError } from 'golikejs/context';

vi.mock('golikejs/context', () => ({
    withCancelCause: vi.fn(() => [{
        done: vi.fn(() => new Promise(() => {})),
        err: vi.fn(() => undefined),
    }, vi.fn()]),
    background: vi.fn(() => ({
        done: vi.fn(() => new Promise(() => {})),
        err: vi.fn(() => undefined),
    })),
    ContextCancelledError: class ContextCancelledError extends Error {
        constructor() {
            super("Context cancelled");
        }
    },
}));

type TestChunk = EncodedChunk & { data: Uint8Array; timestamp: number };

function createChunk(bytes: number[]): TestChunk {
    const buffer = new Uint8Array(bytes);

    return {
        type: 'key',
        byteLength: buffer.byteLength,
        timestamp: 1234,
        data: buffer,
        copyTo(dest: AllowSharedBufferSource) {
            if (dest instanceof Uint8Array) {
                dest.set(buffer);
            } else if (dest instanceof ArrayBuffer || (typeof SharedArrayBuffer !== 'undefined' && dest instanceof SharedArrayBuffer)) {
                new Uint8Array(dest as ArrayBufferLike).set(buffer);
            } else if (ArrayBuffer.isView(dest)) {
                const view = dest as ArrayBufferView;
                new Uint8Array(view.buffer, view.byteOffset, view.byteLength).set(buffer);
            } else {
                throw new Error('Unsupported destination type');
            }
        },
    };
}

describe('cloneChunk', () => {
    test('copies data into a Uint8Array destination', () => {
    const chunk = createChunk([1, 2, 3, 4]);
    const clone = cloneChunk(chunk);

        const dest = new Uint8Array(4);
        clone.copyTo(dest);

        expect(dest).toEqual(chunk.data);
        expect(clone.type).toBe(chunk.type);
        expect(clone.byteLength).toBe(chunk.byteLength);
        expect(clone.timestamp).toBe(chunk.timestamp);
    });

    test('copies data into an ArrayBuffer destination', () => {
    const chunk = createChunk([5, 6, 7]);
    const clone = cloneChunk(chunk);

        const destBuffer = new ArrayBuffer(chunk.byteLength);
        clone.copyTo(destBuffer);

        expect(new Uint8Array(destBuffer)).toEqual(chunk.data);
    });

    test('copies data into a typed array view', () => {
    const chunk = createChunk([0, 10, 20, 30]);
    const clone = cloneChunk(chunk);

        const dest = new Uint8Array(new ArrayBuffer(chunk.byteLength + 4), 2, chunk.byteLength);
        clone.copyTo(dest);

        expect(dest).toEqual(chunk.data);
    });

    test('throws when destination type is unsupported', () => {
    const chunk = createChunk([1, 2, 3]);
    const clone = cloneChunk(chunk);

        expect(() => clone.copyTo({} as any)).toThrow('Unsupported destination type');
    });
});

describe('NoOpTrackEncoder', () => {
    test('encoding is false when no tracks', () => {
        const source = {
            getReader: vi.fn(() => ({
                read: vi.fn(() => Promise.resolve({ done: true })),
                releaseLock: vi.fn(),
            })),
        } as unknown as ReadableStream<EncodedChunk>;
        const encoder = new NoOpTrackEncoder({ source });
        expect(encoder.encoding).toBe(false);
    });

    test('encoding is true when tracks are added', async () => {
        const source = {
            getReader: vi.fn(() => ({
                read: vi.fn(() => Promise.resolve({ done: true })),
                releaseLock: vi.fn(),
            })),
        } as unknown as ReadableStream<EncodedChunk>;
        const encoder = new NoOpTrackEncoder({ source });
        const mockTrackWriter = {
            createGroup: vi.fn(),
            context: {
                done: vi.fn(() => new Promise(() => {})),
                err: vi.fn(() => undefined),
            },
        } as any;
        const result = await encoder.encodeTo(Promise.resolve(), mockTrackWriter);
        expect(result).toBe(ContextCancelledError);
        expect(encoder.encoding).toBe(true);
    });

    test('close method', async () => {
        const source = {
            getReader: vi.fn(() => ({
                read: vi.fn(() => Promise.resolve({ done: true })),
                releaseLock: vi.fn(),
            })),
        } as unknown as ReadableStream<EncodedChunk>;
        const encoder = new NoOpTrackEncoder({ source });
        await expect(encoder.close()).resolves.toBeUndefined();
    });

    test('encodeTo handles key chunks', async () => {
        const chunks = [
            createChunk([1, 2, 3]),
            createChunk([4, 5, 6]),
        ];
        chunks[0].type = 'key';
        chunks[1].type = 'delta';

        const source = {
            getReader: vi.fn(() => ({
                read: vi.fn()
                    .mockResolvedValueOnce({ done: false, value: chunks[0] })
                    .mockResolvedValueOnce({ done: false, value: chunks[1] })
                    .mockResolvedValueOnce({ done: true }),
                releaseLock: vi.fn(),
            })),
        } as unknown as ReadableStream<EncodedChunk>;

        const mockGroup = {
            writeFrame: vi.fn().mockResolvedValue(undefined),
        };
        const mockWriter = {
            createGroup: vi.fn(),
            context: {
                done: vi.fn(() => new Promise(() => {})),
                err: vi.fn(() => undefined),
            },
            openGroup: vi.fn().mockResolvedValue([mockGroup, null]),
            closeWithError: vi.fn(),
            close: vi.fn(),
        } as any;

        const encoder = new NoOpTrackEncoder({ source });
        const ctx = { done: () => Promise.resolve(), err: () => undefined };
        await encoder.encodeTo(ctx as any, mockWriter);

        // Should have called openGroup for key chunk
        expect(mockWriter.openGroup).toHaveBeenCalled();
    });
});


describe('cloneChunk', () => {
  it('clones buffer and copyTo works', () => {
    const data = new Uint8Array([1,2,3,4]);
    const chunk = {
      type: 'key' as const,
      byteLength: data.byteLength,
      timestamp: Date.now(),
      copyTo(dest: Uint8Array) { dest.set(data); }
    };

    const cloned = cloneChunk(chunk as any);
    expect(cloned.byteLength).toBe(4);
    const target = new Uint8Array(4);
    cloned.copyTo(target);
    expect(Array.from(target)).toEqual([1,2,3,4]);
  });

  it('copyTo throws when dest too small', () => {
    const data = new Uint8Array([1,2,3,4]);
    const chunk = {
      type: 'delta' as const,
      byteLength: data.byteLength,
      timestamp: Date.now(),
      copyTo(dest: Uint8Array) { dest.set(data); }
    };

    const cloned = cloneChunk(chunk as any);
    const small = new Uint8Array(2);
    expect(() => cloned.copyTo(small)).toThrow();
  });
});
