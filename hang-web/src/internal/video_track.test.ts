import { describe, expect, it, afterEach, vi } from 'vitest';
import { VideoTrackEncoder, VideoTrackDecoder } from './video_track';

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

        expect(decoder.decoding).toBe(false);
    });
});
