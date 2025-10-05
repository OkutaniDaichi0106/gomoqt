import { describe, test, expect, beforeEach, afterEach, vi, it } from 'vitest';
import { AudioTrackEncoder, AudioTrackDecoder } from './audio_track';
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

describe("audio_track", () => {
    afterEach(() => {
        vi.unstubAllGlobals();
        vi.resetModules();
    });

    it("AudioTrackEncoder reports encoding when tracks are added", async () => {
        class FakeAudioEncoder {
            constructor(_: any) {}
            configure(): void {}
            encode(): void {}
            close(): void {}
        }

        vi.stubGlobal('AudioEncoder', FakeAudioEncoder);

        const reader = {
            read: vi.fn(() => Promise.resolve({ done: true })),
            releaseLock: vi.fn(),
        };
        const source = {
            getReader: () => reader,
        } as unknown as ReadableStream<AudioData>;

        const encoder = new AudioTrackEncoder({
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
        expect(encoder.encoding).toBe(true);
    });

    it("AudioTrackEncoder close method", async () => {
        class FakeAudioEncoder {
            constructor(_: any) {}
            configure(): void {}
            encode(): void {}
            close(): void {}
        }

        vi.stubGlobal('AudioEncoder', FakeAudioEncoder);

        const reader = {
            read: vi.fn(() => Promise.resolve({ done: true })),
            releaseLock: vi.fn(),
        };
        const source = {
            getReader: () => reader,
        } as unknown as ReadableStream<AudioData>;

        const encoder = new AudioTrackEncoder({
            source,
        });

        await expect(encoder.close()).resolves.toBeUndefined();
    });

    it("AudioTrackDecoder decodeFrom method", async () => {
        class FakeAudioEncoder {
            constructor(_: any) {}
            configure(): void {}
            encode(): void {}
            close(): void {}
        }
        class FakeAudioDecoder {
            constructor(_: any) {}
            configure(): void {}
            decode(): void {}
            close(): void {}
        }
        class FakeEncodedAudioChunk {
            constructor(_: any) {}
        }

        vi.stubGlobal('AudioEncoder', FakeAudioEncoder);
        vi.stubGlobal('AudioDecoder', FakeAudioDecoder);
        vi.stubGlobal('EncodedAudioChunk', FakeEncodedAudioChunk);

        const destination = {
            getWriter: () => ({
                ready: Promise.resolve(),
                write: vi.fn(),
                releaseLock: vi.fn(),
            }),
        } as unknown as WritableStream<AudioData>;

        const decoder = new AudioTrackDecoder({
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

    // it("AudioTrackEncoder configure method", async () => {
    //     let configureCalled = false;
    //     let outputCalled = false;
    //     class FakeAudioEncoder {
    //         output?: (chunk: EncodedAudioChunk, metadata?: any) => void;
    //         constructor(_: any) {
    //             // Set output in constructor to simulate the real encoder
    //             this.output = (chunk: EncodedAudioChunk, metadata?: any) => {
    //                 if (metadata?.decoderConfig) {
    //                     // We need to access the resolveConfig from the encoder instance
    //                     // This is tricky, so we'll use a global variable for the test
    //                     if ((global as any).testResolveConfig) {
    //                         (global as any).testResolveConfig(metadata.decoderConfig);
    //                     }
    //                 }
    //             };
    //         }
    //         configure(config: any): void {
    //             configureCalled = true;
    //             // Call output directly
    //             if (this.output) {
    //                 this.output(new EncodedAudioChunk({ type: 'key', timestamp: 0, data: new Uint8Array() }), { decoderConfig: { codec: 'opus', sampleRate: 48000, numberOfChannels: 2 } });
    //             }
    //         }
    //         encode(): void {}
    //         close(): void {}
    //     }

    //     vi.stubGlobal('AudioEncoder', FakeAudioEncoder);

    //     const reader = {
    //         read: vi.fn(() => Promise.resolve({ done: true })),
    //         releaseLock: vi.fn(),
    //     };
    //     const source = {
    //         getReader: () => reader,
    //     } as unknown as ReadableStream<AudioData>;

    //     const encoder = new AudioTrackEncoder({
    //         source,
    //     });

    //     // Set up a way to resolve the config
    //     (global as any).testResolveConfig = (config: any) => {
    //         (encoder as any).#resolveConfig(config);
    //     };

    //     const config: AudioEncoderConfig = { codec: 'opus', sampleRate: 48000, numberOfChannels: 2 };
    //     const decoderConfig = await encoder.configure(config);
    //     expect(decoderConfig).toEqual({ codec: 'opus', sampleRate: 48000, numberOfChannels: 2 });
    // //     expect(decoderConfig).toEqual({ codec: 'opus', sampleRate: 48000, numberOfChannels: 2 });
    // // });
    // });

    it("AudioTrackDecoder configure method", async () => {
        let configureCalled = false;
        class FakeAudioDecoder {
            constructor(_: any) {}
            configure(config: any): void {
                configureCalled = true;
            }
            decode(): void {}
            close(): void {}
        }

        vi.stubGlobal('AudioDecoder', FakeAudioDecoder);

        const destination = {
            getWriter: () => ({
                ready: Promise.resolve(),
                write: vi.fn(),
                releaseLock: vi.fn(),
            }),
        } as unknown as WritableStream<AudioData>;

        const decoder = new AudioTrackDecoder({
            destination,
        });

        const config: AudioDecoderConfig = { codec: 'opus', sampleRate: 48000, numberOfChannels: 2 };
        decoder.configure(config);
        expect(configureCalled).toBe(true);
    });

    it("AudioTrackEncoder encoding getter", async () => {
        class FakeAudioEncoder {
            constructor(_: any) {}
            configure(): void {}
            encode(): void {}
            close(): void {}
        }

        vi.stubGlobal('AudioEncoder', FakeAudioEncoder);

        const reader = {
            read: vi.fn(() => Promise.resolve({ done: true })),
            releaseLock: vi.fn(),
        };
        const source = {
            getReader: () => reader,
        } as unknown as ReadableStream<AudioData>;

        const encoder = new AudioTrackEncoder({
            source,
        });

        expect(encoder.encoding).toBe(false);

        const mockTrackWriter = {
            createGroup: vi.fn(),
            context: {
                done: vi.fn(() => Promise.resolve()),
                err: vi.fn(() => undefined),
            },
        } as unknown as TrackWriter;

        // Add a track
        encoder.encodeTo(Promise.resolve(), mockTrackWriter);
        expect(encoder.encoding).toBe(true);
    });

    it("AudioTrackDecoder decoding getter", async () => {
        class FakeAudioDecoder {
            constructor(_: any) {}
            configure(): void {}
            decode(): void {}
            close(): void {}
        }

        vi.stubGlobal('AudioDecoder', FakeAudioDecoder);

        const destination = {
            getWriter: () => ({
                ready: Promise.resolve(),
                write: vi.fn(),
                releaseLock: vi.fn(),
            }),
        } as unknown as WritableStream<AudioData>;

        const decoder = new AudioTrackDecoder({
            destination,
        });

        expect(decoder.decoding).toBe(false);

        const mockTrackReader = {
            acceptGroup: vi.fn().mockResolvedValue(undefined),
            context: {
                done: vi.fn(() => Promise.resolve()),
                err: vi.fn(() => undefined),
            },
        } as unknown as TrackReader;

        // Set source
        decoder.decodeFrom(Promise.resolve(), mockTrackReader);
        expect(decoder.decoding).toBe(true);
    });

    it("AudioTrackEncoder handles encoder output with decoderConfig", async () => {
        let outputCallback: ((chunk: EncodedAudioChunk, metadata?: any) => void) | undefined;
        
        class FakeAudioEncoder {
            constructor(init: any) {
                outputCallback = init.output;
            }
            configure(_config: any): void {
                // Simulate encoder calling output with metadata after configure
                queueMicrotask(() => {
                    const mockChunk = {
                        type: 'key',
                        timestamp: 0,
                        duration: 20000,
                        byteLength: 100,
                        copyTo: vi.fn(),
                    } as unknown as EncodedAudioChunk;

                    const metadata = {
                        decoderConfig: {
                            codec: 'opus',
                            sampleRate: 48000,
                            numberOfChannels: 2,
                        }
                    };

                    outputCallback!(mockChunk, metadata);
                });
            }
            encode(): void {}
            close(): void {}
        }

        vi.stubGlobal('AudioEncoder', FakeAudioEncoder);

        const reader = {
            read: vi.fn(() => new Promise(() => {})),
            releaseLock: vi.fn(),
        };
        const source = {
            getReader: () => reader,
        } as unknown as ReadableStream<AudioData>;

        const encoder = new AudioTrackEncoder({
            source,
        });

        // Trigger configure to set resolveConfig
        const configPromise = encoder.configure({ 
            codec: 'opus', 
            sampleRate: 48000, 
            numberOfChannels: 2 
        });

        const decoderConfig = await configPromise;
        expect(decoderConfig).toEqual({ 
            codec: 'opus', 
            sampleRate: 48000, 
            numberOfChannels: 2 
        });
    });

    it("AudioTrackEncoder handles group rollover based on timestamp", async () => {
        let outputCallback: ((chunk: EncodedAudioChunk, metadata?: any) => void) | undefined;
        
        class FakeAudioEncoder {
            constructor(init: any) {
                outputCallback = init.output;
            }
            configure(): void {}
            encode(): void {}
            close(): void {}
        }

        class FakeEncodedAudioChunk {
            type: string;
            timestamp: number;
            constructor(init: any) {
                this.type = init.type;
                this.timestamp = init.timestamp;
            }
            copyTo(_buffer: Uint8Array): void {}
        }

        vi.stubGlobal('AudioEncoder', FakeAudioEncoder);
        vi.stubGlobal('EncodedAudioChunk', FakeEncodedAudioChunk);

        const reader = {
            read: vi.fn(() => new Promise(() => {})),
            releaseLock: vi.fn(),
        };
        const source = {
            getReader: () => reader,
        } as unknown as ReadableStream<AudioData>;

        const encoder = new AudioTrackEncoder({
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

        // First chunk at timestamp 0
        const chunk1 = new FakeEncodedAudioChunk({
            type: 'key',
            timestamp: 0,
        }) as unknown as EncodedAudioChunk;

        await outputCallback!(chunk1, {});

        // Second chunk at timestamp 150ms (exceeds MAX_AUDIO_LATENCY of 100ms)
        const chunk2 = new FakeEncodedAudioChunk({
            type: 'key',
            timestamp: 150,
        }) as unknown as EncodedAudioChunk;

        await outputCallback!(chunk2, {});

        // Wait for async operations
        await new Promise(resolve => setTimeout(resolve, 10));

        // Should have opened a new group due to timestamp exceeding threshold
        expect(openGroupMock).toHaveBeenCalledWith(2n);
    });

    it("AudioTrackEncoder handles non-key chunks", async () => {
        let outputCallback: ((chunk: EncodedAudioChunk, metadata?: any) => void) | undefined;
        const consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
        
        class FakeAudioEncoder {
            constructor(init: any) {
                outputCallback = init.output;
            }
            configure(): void {}
            encode(): void {}
            close(): void {}
        }

        vi.stubGlobal('AudioEncoder', FakeAudioEncoder);

        const reader = {
            read: vi.fn(() => new Promise(() => {})),
            releaseLock: vi.fn(),
        };
        const source = {
            getReader: () => reader,
        } as unknown as ReadableStream<AudioData>;

        const encoder = new AudioTrackEncoder({
            source,
        });

        const mockChunk = {
            type: 'delta',
            timestamp: 0,
        } as EncodedAudioChunk;

        await outputCallback!(mockChunk, {});

        expect(consoleWarnSpy).toHaveBeenCalledWith("Ignoring non-key audio chunk");
        consoleWarnSpy.mockRestore();
    });

    it("AudioTrackEncoder handles encoding errors", async () => {
        let errorCallback: ((error: Error) => void) | undefined;
        
        class FakeAudioEncoder {
            constructor(init: any) {
                errorCallback = init.error;
            }
            configure(): void {}
            encode(): void {}
            close(): void {}
        }

        vi.stubGlobal('AudioEncoder', FakeAudioEncoder);

        const reader = {
            read: vi.fn(() => new Promise(() => {})),
            releaseLock: vi.fn(),
        };
        const source = {
            getReader: () => reader,
        } as unknown as ReadableStream<AudioData>;

        const encoder = new AudioTrackEncoder({
            source,
        });

        const closeSpy = vi.spyOn(encoder, 'close');

        const error = new Error("Encoding failed");
        errorCallback!(error);

        expect(closeSpy).toHaveBeenCalledWith(error);
    });

    it("AudioTrackDecoder handles frame decoding in #next", async () => {
        let outputCallback: ((frame: AudioData) => void) | undefined;
        
        class FakeAudioDecoder {
            constructor(init: any) {
                outputCallback = init.output;
            }
            configure(): void {}
            decode(): void {}
            close(): void {}
        }

        class FakeEncodedAudioChunk {
            constructor(_: any) {}
        }

        vi.stubGlobal('AudioDecoder', FakeAudioDecoder);
        vi.stubGlobal('EncodedAudioChunk', FakeEncodedAudioChunk);

        const writeMock = vi.fn().mockResolvedValue(undefined);
        const destination = {
            getWriter: () => ({
                ready: Promise.resolve(),
                write: writeMock,
                releaseLock: vi.fn(),
            }),
        } as unknown as WritableStream<AudioData>;

        const decoder = new AudioTrackDecoder({
            destination,
        });

        const mockFrame = {
            timestamp: 1000,
        } as AudioData;

        await outputCallback!(mockFrame);

        expect(writeMock).toHaveBeenCalledWith(mockFrame);
    });

    it("AudioTrackDecoder handles write errors", async () => {
        let outputCallback: ((frame: AudioData) => void) | undefined;
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
        
        class FakeAudioDecoder {
            constructor(init: any) {
                outputCallback = init.output;
            }
            configure(): void {}
            decode(): void {}
            close(): void {}
        }

        vi.stubGlobal('AudioDecoder', FakeAudioDecoder);

        const releaseLockMock = vi.fn();
        const destination = {
            getWriter: () => ({
                ready: Promise.resolve(),
                write: vi.fn().mockRejectedValue(new Error("Write failed")),
                releaseLock: releaseLockMock,
            }),
        } as unknown as WritableStream<AudioData>;

        const decoder = new AudioTrackDecoder({
            destination,
        });

        const mockFrame = {
            timestamp: 1000,
        } as AudioData;

        await outputCallback!(mockFrame);

        // Wait for error handling
        await new Promise(resolve => setTimeout(resolve, 10));

        expect(consoleErrorSpy).toHaveBeenCalled();
        expect(releaseLockMock).toHaveBeenCalled();
        
        consoleErrorSpy.mockRestore();
    });

    it("AudioTrackDecoder handles decoding errors", async () => {
        let errorCallback: ((error: Error) => void) | undefined;
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
        
        class FakeAudioDecoder {
            constructor(init: any) {
                errorCallback = init.error;
            }
            configure(): void {}
            decode(): void {}
            close(): void {}
        }

        vi.stubGlobal('AudioDecoder', FakeAudioDecoder);

        const destination = {
            getWriter: () => ({
                ready: Promise.resolve(),
                write: vi.fn(),
                releaseLock: vi.fn(),
            }),
        } as unknown as WritableStream<AudioData>;

        const decoder = new AudioTrackDecoder({
            destination,
        });

        const error = new Error("Decoding failed");
        errorCallback!(error);

        expect(consoleErrorSpy).toHaveBeenCalledWith("Audio decoding error (no auto-close):", error);
        
        consoleErrorSpy.mockRestore();
    });

    it("AudioTrackEncoder skips encoding when no tracks", async () => {
        let outputCallback: ((chunk: EncodedAudioChunk, metadata?: any) => void) | undefined;
        
        class FakeAudioEncoder {
            constructor(init: any) {
                outputCallback = init.output;
            }
            configure(): void {}
            encode(): void {}
            close(): void {}
        }

        vi.stubGlobal('AudioEncoder', FakeAudioEncoder);

        const reader = {
            read: vi.fn(() => new Promise(() => {})),
            releaseLock: vi.fn(),
        };
        const source = {
            getReader: () => reader,
        } as unknown as ReadableStream<AudioData>;

        const encoder = new AudioTrackEncoder({
            source,
        });

        const mockChunk = {
            type: 'key',
            timestamp: 0,
        } as EncodedAudioChunk;

        // Should return immediately without error since no tracks
        await expect(outputCallback!(mockChunk, {})).resolves.toBeUndefined();
    });

    it("AudioTrackEncoder handles openGroup errors", async () => {
        let outputCallback: ((chunk: EncodedAudioChunk, metadata?: any) => void) | undefined;
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
        
        class FakeAudioEncoder {
            constructor(init: any) {
                outputCallback = init.output;
            }
            configure(): void {}
            encode(): void {}
            close(): void {}
        }

        class FakeEncodedAudioChunk {
            type: string;
            timestamp: number;
            constructor(init: any) {
                this.type = init.type;
                this.timestamp = init.timestamp;
            }
            copyTo(_buffer: Uint8Array): void {}
        }

        vi.stubGlobal('AudioEncoder', FakeAudioEncoder);
        vi.stubGlobal('EncodedAudioChunk', FakeEncodedAudioChunk);

        const reader = {
            read: vi.fn(() => new Promise(() => {})),
            releaseLock: vi.fn(),
        };
        const source = {
            getReader: () => reader,
        } as unknown as ReadableStream<AudioData>;

        const encoder = new AudioTrackEncoder({
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

        // Trigger group rollover
        const chunk = new FakeEncodedAudioChunk({
            type: 'key',
            timestamp: 150,
        }) as unknown as EncodedAudioChunk;

        await outputCallback!(chunk, {});

        // Wait for async operations
        await new Promise(resolve => setTimeout(resolve, 10));

        expect(consoleErrorSpy).toHaveBeenCalledWith("moq: failed to open group:", error);
        expect(closeWithErrorMock).toHaveBeenCalled();
        
        consoleErrorSpy.mockRestore();
    });

    it("AudioTrackDecoder close method", async () => {
        class FakeAudioDecoder {
            constructor(_: any) {}
            configure(): void {}
            decode(): void {}
            close(): void {}
        }

        vi.stubGlobal('AudioDecoder', FakeAudioDecoder);

        const releaseLockMock = vi.fn();
        const destination = {
            getWriter: () => ({
                ready: Promise.resolve(),
                write: vi.fn(),
                releaseLock: releaseLockMock,
            }),
        } as unknown as WritableStream<AudioData>;

        const decoder = new AudioTrackDecoder({
            destination,
        });

        await decoder.close();

        expect(releaseLockMock).toHaveBeenCalled();
    });

    it("AudioTrackDecoder decodeFrom replaces existing source", async () => {
        const consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
        
        class FakeAudioDecoder {
            constructor(_: any) {}
            configure(): void {}
            decode(): void {}
            close(): void {}
        }

        vi.stubGlobal('AudioDecoder', FakeAudioDecoder);

        const destination = {
            getWriter: () => ({
                ready: Promise.resolve(),
                write: vi.fn(),
                releaseLock: vi.fn(),
            }),
        } as unknown as WritableStream<AudioData>;

        const decoder = new AudioTrackDecoder({
            destination,
        });

        const closeWithErrorMock = vi.fn().mockResolvedValue(undefined);
        const mockTrackReader1 = {
            acceptGroup: vi.fn().mockResolvedValue(undefined),
            closeWithError: closeWithErrorMock,
            context: {
                done: vi.fn(() => Promise.resolve()),
                err: vi.fn(() => undefined),
            },
        } as unknown as TrackReader;

        // Set first source
        const promise1 = decoder.decodeFrom(Promise.resolve(), mockTrackReader1);

        // Set second source (should replace the first)
        const mockTrackReader2 = {
            acceptGroup: vi.fn().mockResolvedValue(undefined),
            context: {
                done: vi.fn(() => Promise.resolve()),
                err: vi.fn(() => undefined),
            },
        } as unknown as TrackReader;

        const promise2 = decoder.decodeFrom(Promise.resolve(), mockTrackReader2);

        await Promise.all([promise1, promise2]);

        expect(consoleWarnSpy).toHaveBeenCalledWith("[AudioDecodeStream] source is already set, replacing");
        expect(closeWithErrorMock).toHaveBeenCalled();
        
        consoleWarnSpy.mockRestore();
    });

    it("AudioTrackDecoder close does nothing when already cancelled", async () => {
        class FakeAudioDecoder {
            constructor(_: any) {}
            configure(): void {}
            decode(): void {}
            close(): void {}
        }

        vi.stubGlobal('AudioDecoder', FakeAudioDecoder);

        const releaseLockMock = vi.fn();
        const destination = {
            getWriter: () => ({
                ready: Promise.resolve(),
                write: vi.fn(),
                releaseLock: releaseLockMock,
            }),
        } as unknown as WritableStream<AudioData>;

        const decoder = new AudioTrackDecoder({
            destination,
        });

        // Close once
        await decoder.close();
        expect(releaseLockMock).toHaveBeenCalledTimes(1);

        // Close again - should not call releaseLock again
        await decoder.close();
        expect(releaseLockMock).toHaveBeenCalledTimes(1);
    });
});
