import { describe, it, expect, jest, beforeEach } from 'vitest';
import { NoOpTrackDecoder } from './track_decoder';

vi.mock('@okutanidaichi/moqt', () => ({
    InternalSubscribeErrorCode: 1,
}));

vi.mock('golikejs/context', () => ({
    withCancelCause: vi.fn(() => [{
        done: vi.fn(() => new Promise(() => {})),
        err: vi.fn(() => undefined),
    }, vi.fn()]),
    background: vi.fn(() => Promise.resolve()),
    ContextCancelledError: undefined,
}));

describe('NoOpTrackDecoder', () => {
    let decoder: NoOpTrackDecoder;
    const mockDestination = {
        getWriter: vi.fn(() => ({
            write: vi.fn(),
            close: vi.fn(),
            releaseLock: vi.fn(),
        })),
    };

    beforeEach(() => {
        decoder = new NoOpTrackDecoder({
            destination: mockDestination as any,
        });
        (decoder as any)['#dests'] = new Map([['test', {}]]);
    });

    describe('constructor', () => {
        it('should create an instance', () => {
            expect(decoder).toBeInstanceOf(NoOpTrackDecoder);
        });
    });

    describe('decoding', () => {
        it('should return true when has destinations', () => {
            expect(decoder.decoding).toBe(true);
        });
    });

    describe('decodeFrom', () => {
        it('should handle decodeFrom', async () => {
            const mockSource = {
                acceptGroup: vi.fn().mockResolvedValue(undefined),
                context: {
                    done: vi.fn(() => new Promise(() => {})),
                    err: vi.fn(() => undefined),
                },
                closeWithError: vi.fn(),
            };

            const result = await decoder.decodeFrom(Promise.resolve(), mockSource as any);
            expect(result).toBeUndefined();
        });
    });

    describe('close', () => {
        it('should close the decoder', async () => {
            await expect(decoder.close()).resolves.not.toThrow();
        });
    });
});
