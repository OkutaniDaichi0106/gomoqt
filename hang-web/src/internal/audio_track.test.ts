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
});
