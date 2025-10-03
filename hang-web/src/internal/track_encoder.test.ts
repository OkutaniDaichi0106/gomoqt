import { describe, test, expect } from 'vitest';
import { cloneChunk } from './track_encoder';
import type { EncodedChunk } from './container';

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

    test('throws when destination buffer is too small', () => {
    const chunk = createChunk([9, 9, 9]);
    const clone = cloneChunk(chunk);

        const dest = new Uint8Array(2);
        expect(() => clone.copyTo(dest)).toThrow(RangeError);
    });
});
