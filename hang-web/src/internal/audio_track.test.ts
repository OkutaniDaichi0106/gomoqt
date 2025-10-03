import { describe, test, expect, beforeEach, afterEach, vi, it } from 'vitest';
import { AudioTrackEncoder, AudioTrackDecoder } from './audio_track';

describe("audio_track", () => {
    afterEach(() => {
        vi.unstubAllGlobals();
        vi.resetModules();
    });

    it("AudioTrackEncoder reports not encoding when no tracks are added", () => {
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
    });

    it("AudioTrackDecoder reports not decoding before a source is set", () => {
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

        expect(decoder.decoding).toBe(false);
    });
});
