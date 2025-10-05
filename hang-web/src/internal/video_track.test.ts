import { describe, expect, it, afterEach, vi } from 'vitest';
import { VideoTrackEncoder, VideoTrackDecoder } from './video_track';
import type { TrackWriter, TrackReader } from '@okutanidaichi/moqt';
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

describe("video_track", () => {
    afterEach(() => {
        vi.unstubAllGlobals();
        vi.resetModules();
    });

    it("VideoTrackEncoder reports not encoding when no tracks are set", () => {
        class FakeVideoEncoder {
            constructor(_: any) {}
            configure(): void {}
            encode(): void {}
            close(): void {}
        }

        vi.stubGlobal('VideoEncoder', FakeVideoEncoder);

        const reader = {
            read: vi.fn(() => Promise.resolve({ done: true })),
            releaseLock: vi.fn(),
        };
        const source = {
            getReader: () => reader,
        } as unknown as ReadableStream<VideoFrame>;

        const encoder = new VideoTrackEncoder({
            source,
        });

        expect(encoder.encoding).toBe(false);
    });

    it("VideoTrackDecoder reports not decoding before a source is provided", () => {
        class FakeVideoEncoder {
            constructor(_: any) {}
            configure(): void {}
            encode(): void {}
            close(): void {}
        }
        class FakeVideoDecoder {
            constructor(_: any) {}
            configure(): void {}
            decode(): void {}
            close(): void {}
        }
        class FakeEncodedVideoChunk {
            constructor(_: any) {}
        }

        vi.stubGlobal('VideoEncoder', FakeVideoEncoder);
        vi.stubGlobal('VideoDecoder', FakeVideoDecoder);
        vi.stubGlobal('EncodedVideoChunk', FakeEncodedVideoChunk);

        const destination = {
            getWriter: () => ({
                ready: Promise.resolve(),
                write: vi.fn(),
                releaseLock: vi.fn(),
            }),
        } as unknown as WritableStream<VideoFrame>;

            const decoder = new VideoTrackDecoder({
                destination,
            });
        });

    it("VideoTrackEncoder encodeTo method", async () => {
        class FakeVideoEncoder {
            constructor(_: any) {}
            configure(): void {}
            encode(): void {}
            close(): void {}
        }

        vi.stubGlobal('VideoEncoder', FakeVideoEncoder);

        const reader = {
            read: vi.fn(() => Promise.resolve({ done: true })),
            releaseLock: vi.fn(),
        };
        const source = {
            getReader: () => reader,
        } as unknown as ReadableStream<VideoFrame>;

        const encoder = new VideoTrackEncoder({
            source,
        });

        const mockTrackWriter = {
            createGroup: vi.fn(),
            context: {
                done: vi.fn(() => Promise.resolve()),
                err: vi.fn(() => undefined),
            },
        } as unknown as TrackWriter;

        const result = await encoder.encodeTo(Promise.resolve(), mockTrackWriter);
        expect(result).toBe(ContextCancelledError);
        expect(encoder.encoding).toBe(false);
    });

    it("VideoTrackEncoder close method", async () => {
        class FakeVideoEncoder {
            constructor(_: any) {}
            configure(): void {}
            encode(): void {}
            close(): void {}
        }

        vi.stubGlobal('VideoEncoder', FakeVideoEncoder);

        const reader = {
            read: vi.fn(() => Promise.resolve({ done: true })),
            releaseLock: vi.fn(),
        };
        const source = {
            getReader: () => reader,
        } as unknown as ReadableStream<VideoFrame>;

        const encoder = new VideoTrackEncoder({
            source,
        });

        await expect(encoder.close()).resolves.toBeUndefined();
    });

    it("VideoTrackDecoder decodeFrom method", async () => {
        class FakeVideoDecoder {
            constructor(_: any) {}
            configure(): void {}
            decode(): void {}
            close(): void {}
        }
        class FakeEncodedVideoChunk {
            constructor(_: any) {}
        }

        vi.stubGlobal('VideoDecoder', FakeVideoDecoder);
        vi.stubGlobal('EncodedVideoChunk', FakeEncodedVideoChunk);

        const destination = {
            getWriter: () => ({
                ready: Promise.resolve(),
                write: vi.fn(),
                releaseLock: vi.fn(),
            }),
        } as unknown as WritableStream<VideoFrame>;

        const decoder = new VideoTrackDecoder({
            destination,
        });

        const mockTrackReader = {
            acceptGroup: vi.fn().mockResolvedValue(undefined),
            context: {
                done: vi.fn(() => Promise.resolve()),
                err: vi.fn(() => undefined),
            },
        } as unknown as TrackReader;

        const result = await decoder.decodeFrom(Promise.resolve(), mockTrackReader);
        expect(result).toBe(ContextCancelledError);
        expect(decoder.decoding).toBe(true);
    });

    it("VideoTrackDecoder close method", async () => {
        class FakeVideoDecoder {
            constructor(_: any) {}
            configure(): void {}
            decode(): void {}
            close(): void {}
        }
        class FakeEncodedVideoChunk {
            constructor(_: any) {}
        }

        vi.stubGlobal('VideoDecoder', FakeVideoDecoder);
        vi.stubGlobal('EncodedVideoChunk', FakeEncodedVideoChunk);

        const destination = {
            getWriter: () => ({
                ready: Promise.resolve(),
                write: vi.fn(),
                releaseLock: vi.fn(),
                close: vi.fn(),
            }),
            close: vi.fn(),
        } as unknown as WritableStream<VideoFrame>;

        const decoder = new VideoTrackDecoder({
            destination,
        });

        // Initialize decoder by calling decodeFrom
        const mockTrackReader = {
            acceptGroup: vi.fn().mockResolvedValue(undefined),
            context: {
                done: vi.fn(() => Promise.resolve()),
                err: vi.fn(() => undefined),
            },
        } as unknown as TrackReader;

        await decoder.decodeFrom(Promise.resolve(), mockTrackReader);

        await expect(decoder.close()).resolves.toBeUndefined();
    });
});



