import { describe, it, expect, vi, beforeEach } from 'vitest';
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
    ContextCancelledError: new Error("Context cancelled"),
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

            const result = await decoder.decodeFrom(Promise.resolve(undefined), mockSource as any);
            expect(result).toBeUndefined();
        });
    });

    describe('tee', () => {
        it('should add a new destination', async () => {
            const mockDest = {
                getWriter: vi.fn(() => ({
                    write: vi.fn(),
                    close: vi.fn(),
                    releaseLock: vi.fn(),
                    closed: Promise.resolve(),
                })),
            };

            const result = await decoder.tee(mockDest as any);
            expect(result).toBeUndefined();
            expect(mockDest.getWriter).toHaveBeenCalled();
        });

        it('should return error if destination already set', async () => {
            const mockDest = {
                getWriter: vi.fn(() => ({
                    write: vi.fn(),
                    close: vi.fn(),
                    releaseLock: vi.fn(),
                    closed: Promise.resolve(),
                })),
            };

            // First tee
            await decoder.tee(mockDest as any);

            // Second tee with same dest
            const result = await decoder.tee(mockDest as any);
            expect(result).toBeInstanceOf(Error);
            expect((result as Error).message).toBe('destination already set');
        });

        it('should handle close method', async () => {
            const mockWriter = {
                write: vi.fn(),
                close: vi.fn(),
                releaseLock: vi.fn(),
                closed: Promise.resolve(),
            };
            const mockDest = {
                getWriter: vi.fn(() => mockWriter),
                close: vi.fn(),
            };

            await decoder.tee(mockDest as any);
            await expect(decoder.close()).resolves.toBeUndefined();
            expect(mockWriter.releaseLock).toHaveBeenCalled();
        });

        it('should handle writer closed with error in tee', async () => {
            const mockWriter = {
                write: vi.fn(),
                close: vi.fn(),
                releaseLock: vi.fn(),
                closed: Promise.reject(new Error('Writer closed with error')),
            };
            const mockDest = {
                getWriter: vi.fn(() => mockWriter),
            };

            const result = await decoder.tee(mockDest as any);
            expect(result).toBeInstanceOf(Error);
            expect((result as Error).message).toBe('destination closed with error');
        });

        it('should handle decodeFrom replacing existing source', async () => {
            const consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
            
            const mockSource1 = {
                acceptGroup: vi.fn().mockResolvedValue(undefined),
                context: {
                    done: vi.fn(() => Promise.resolve()),
                    err: vi.fn(() => undefined),
                },
                closeWithError: vi.fn(),
            };

            const mockSource2 = {
                acceptGroup: vi.fn().mockResolvedValue(undefined),
                context: {
                    done: vi.fn(() => Promise.resolve()),
                    err: vi.fn(() => undefined),
                },
                closeWithError: vi.fn(),
            };

            // First decode
            const promise1 = decoder.decodeFrom(Promise.resolve(), mockSource1 as any);

            // Second decode (should replace first)
            const promise2 = decoder.decodeFrom(Promise.resolve(), mockSource2 as any);

            await Promise.all([promise1, promise2]);

            expect(consoleWarnSpy).toHaveBeenCalledWith('[NoOpTrackDecoder] source already set. replacing...');
            expect(mockSource1.closeWithError).toHaveBeenCalled();
            
            consoleWarnSpy.mockRestore();
        });

    });
});
