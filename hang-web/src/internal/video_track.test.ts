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

            expect(decoder.decoding).toBe(false);
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
        expect(encoder.encoding).toBe(false);
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

        await expect(decoder.close(new Error("test"))).resolves.toBeUndefined();
        await new Promise<void>(resolve => queueMicrotask(() => resolve()));
        expect(decoder.decoding).toBe(false);
    });

    it("VideoTrackEncoder handles encoder output with decoderConfig", async () => {
        let outputCallback: ((chunk: EncodedVideoChunk, metadata?: any) => void) | undefined;
        
        class FakeVideoEncoder {
            constructor(init: any) {
                outputCallback = init.output;
            }
            configure(_config: any): void {
                queueMicrotask(() => {
                    const mockChunk = {
                        type: 'key',
                        timestamp: 0,
                        duration: 33333,
                        byteLength: 500,
                        copyTo: vi.fn(),
                    } as unknown as EncodedVideoChunk;

                    const metadata = {
                        decoderConfig: {
                            codec: 'vp09.00.10.08',
                            codedWidth: 1920,
                            codedHeight: 1080,
                        }
                    };

                    outputCallback!(mockChunk, metadata);
                });
            }
            encode(): void {}
            close(): void {}
        }

        vi.stubGlobal('VideoEncoder', FakeVideoEncoder);

        const reader = {
            read: vi.fn(() => new Promise(() => {})),
            releaseLock: vi.fn(),
        };
        const source = {
            getReader: () => reader,
        } as unknown as ReadableStream<VideoFrame>;

        const encoder = new VideoTrackEncoder({
            source,
        });

        const configPromise = encoder.configure({ 
            codec: 'vp09.00.10.08',
            width: 1920,
            height: 1080,
        });

        const decoderConfig = await configPromise;
        expect(decoderConfig).toEqual({ 
            codec: 'vp09.00.10.08', 
            codedWidth: 1920, 
            codedHeight: 1080 
        });
    });

    it("VideoTrackEncoder handles key frames and group rollover", async () => {
        let outputCallback: ((chunk: EncodedVideoChunk, metadata?: any) => void) | undefined;
        
        class FakeVideoEncoder {
            constructor(init: any) {
                outputCallback = init.output;
            }
            configure(): void {}
            encode(): void {}
            close(): void {}
        }

        class FakeEncodedVideoChunk {
            type: string;
            timestamp: number;
            constructor(init: any) {
                this.type = init.type;
                this.timestamp = init.timestamp;
            }
            copyTo(_buffer: Uint8Array): void {}
        }

        vi.stubGlobal('VideoEncoder', FakeVideoEncoder);
        vi.stubGlobal('EncodedVideoChunk', FakeEncodedVideoChunk);

        const reader = {
            read: vi.fn(() => new Promise(() => {})),
            releaseLock: vi.fn(),
        };
        const source = {
            getReader: () => reader,
        } as unknown as ReadableStream<VideoFrame>;

        const encoder = new VideoTrackEncoder({
            source,
            startGroupSequence: 1n,
        });

        const mockGroup = {
            close: vi.fn().mockResolvedValue(undefined),
            writeFrame: vi.fn().mockResolvedValue(undefined),
        };

        const openGroupMock = vi.fn().mockResolvedValue([mockGroup, undefined]);

        const mockTrackWriter = {
            openGroup: openGroupMock,
            closeWithError: vi.fn(),
            context: {
                done: vi.fn(() => new Promise(() => {})),
                err: vi.fn(() => undefined),
            },
        } as unknown as TrackWriter;

        encoder.encodeTo(Promise.resolve(), mockTrackWriter);

        // Send key frame - should trigger group rollover
        const keyChunk = new FakeEncodedVideoChunk({
            type: 'key',
            timestamp: 1000000,
        }) as unknown as EncodedVideoChunk;

        await outputCallback!(keyChunk, {});

        // Wait for async operations
        await new Promise(resolve => setTimeout(resolve, 10));

        expect(openGroupMock).toHaveBeenCalledWith(2n);
    });

    it("VideoTrackEncoder handles encoding errors", async () => {
        let errorCallback: ((error: Error) => void) | undefined;
        
        class FakeVideoEncoder {
            constructor(init: any) {
                errorCallback = init.error;
            }
            configure(): void {}
            encode(): void {}
            close(): void {}
        }

        vi.stubGlobal('VideoEncoder', FakeVideoEncoder);

        const reader = {
            read: vi.fn(() => new Promise(() => {})),
            releaseLock: vi.fn(),
        };
        const source = {
            getReader: () => reader,
        } as unknown as ReadableStream<VideoFrame>;

        const encoder = new VideoTrackEncoder({
            source,
        });

        const closeSpy = vi.spyOn(encoder, 'close');

        const error = new Error("Encoding failed");
        errorCallback!(error);

        expect(closeSpy).toHaveBeenCalledWith(error);
    });

    it("VideoTrackEncoder skips encoding when no tracks", async () => {
        let outputCallback: ((chunk: EncodedVideoChunk, metadata?: any) => void) | undefined;
        
        class FakeVideoEncoder {
            constructor(init: any) {
                outputCallback = init.output;
            }
            configure(): void {}
            encode(): void {}
            close(): void {}
        }

        vi.stubGlobal('VideoEncoder', FakeVideoEncoder);

        const reader = {
            read: vi.fn(() => new Promise(() => {})),
            releaseLock: vi.fn(),
        };
        const source = {
            getReader: () => reader,
        } as unknown as ReadableStream<VideoFrame>;

        const encoder = new VideoTrackEncoder({
            source,
        });

        const mockChunk = {
            type: 'delta',
            timestamp: 0,
            copyTo: vi.fn(),
        } as unknown as EncodedVideoChunk;

        await expect(outputCallback!(mockChunk, {})).resolves.toBeUndefined();
    });

    it("VideoTrackEncoder handles openGroup errors", async () => {
        let outputCallback: ((chunk: EncodedVideoChunk, metadata?: any) => void) | undefined;
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
        
        class FakeVideoEncoder {
            constructor(init: any) {
                outputCallback = init.output;
            }
            configure(): void {}
            encode(): void {}
            close(): void {}
        }

        class FakeEncodedVideoChunk {
            type: string;
            timestamp: number;
            constructor(init: any) {
                this.type = init.type;
                this.timestamp = init.timestamp;
            }
            copyTo(_buffer: Uint8Array): void {}
        }

        vi.stubGlobal('VideoEncoder', FakeVideoEncoder);
        vi.stubGlobal('EncodedVideoChunk', FakeEncodedVideoChunk);

        const reader = {
            read: vi.fn(() => new Promise(() => {})),
            releaseLock: vi.fn(),
        };
        const source = {
            getReader: () => reader,
        } as unknown as ReadableStream<VideoFrame>;

        const encoder = new VideoTrackEncoder({
            source,
            startGroupSequence: 1n,
        });

        const error = new Error("Failed to open group");
        const closeWithErrorMock = vi.fn();
        const openGroupMock = vi.fn().mockResolvedValue([undefined, error]);

        const mockTrackWriter = {
            openGroup: openGroupMock,
            closeWithError: closeWithErrorMock,
            context: {
                done: vi.fn(() => new Promise(() => {})),
                err: vi.fn(() => undefined),
            },
        } as unknown as TrackWriter;

        encoder.encodeTo(Promise.resolve(), mockTrackWriter);

        const chunk = new FakeEncodedVideoChunk({
            type: 'key',
            timestamp: 1000000,
        }) as unknown as EncodedVideoChunk;

        await outputCallback!(chunk, {});

        await new Promise(resolve => setTimeout(resolve, 10));

        expect(consoleErrorSpy).toHaveBeenCalledWith("moq: failed to open group:", error);
        expect(closeWithErrorMock).toHaveBeenCalled();
        
        consoleErrorSpy.mockRestore();
    });

    it("VideoTrackDecoder handles frame decoding", async () => {
        let outputCallback: ((frame: VideoFrame) => void) | undefined;
        
        class FakeVideoDecoder {
            constructor(init: any) {
                outputCallback = init.output;
            }
            configure(): void {}
            decode(): void {}
            close(): void {}
        }

        vi.stubGlobal('VideoDecoder', FakeVideoDecoder);

        const writeMock = vi.fn().mockResolvedValue(undefined);
        const destination = {
            getWriter: () => ({
                ready: Promise.resolve(),
                write: writeMock,
                releaseLock: vi.fn(),
            }),
        } as unknown as WritableStream<VideoFrame>;

        const decoder = new VideoTrackDecoder({
            destination,
        });

        const mockFrame = {
            timestamp: 33333,
            codedWidth: 1920,
            codedHeight: 1080,
            close: vi.fn(),
        } as unknown as VideoFrame;

        await outputCallback!(mockFrame);

        expect(writeMock).toHaveBeenCalledWith(mockFrame);
    });

    it("VideoTrackDecoder handles write errors", async () => {
        let outputCallback: ((frame: VideoFrame) => void) | undefined;
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
        
        class FakeVideoDecoder {
            constructor(init: any) {
                outputCallback = init.output;
            }
            configure(): void {}
            decode(): void {}
            close(): void {}
        }

        vi.stubGlobal('VideoDecoder', FakeVideoDecoder);

        const releaseLockMock = vi.fn();
        const destination = {
            getWriter: () => ({
                ready: Promise.resolve(),
                write: vi.fn().mockRejectedValue(new Error("Write failed")),
                releaseLock: releaseLockMock,
            }),
        } as unknown as WritableStream<VideoFrame>;

        const decoder = new VideoTrackDecoder({
            destination,
        });

        const mockFrame = {
            timestamp: 33333,
            close: vi.fn(),
        } as unknown as VideoFrame;

        await outputCallback!(mockFrame);

        await new Promise(resolve => setTimeout(resolve, 10));

        expect(consoleErrorSpy).toHaveBeenCalled();
        expect(releaseLockMock).toHaveBeenCalled();
        
        consoleErrorSpy.mockRestore();
    });

    it("VideoTrackDecoder handles decoding errors", async () => {
        let errorCallback: ((error: Error) => void) | undefined;
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
        
        class FakeVideoDecoder {
            constructor(init: any) {
                errorCallback = init.error;
            }
            configure(): void {}
            decode(): void {}
            close(): void {}
        }

        vi.stubGlobal('VideoDecoder', FakeVideoDecoder);

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

        const error = new Error("Decoding failed");
        errorCallback!(error);

        expect(consoleErrorSpy).toHaveBeenCalledWith("Video decoding error (no auto-close):", error);
        
        consoleErrorSpy.mockRestore();
    });
});



