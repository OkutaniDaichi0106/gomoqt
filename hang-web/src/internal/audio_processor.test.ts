import { describe, test, expect, beforeEach, afterEach, vi } from 'vitest';
import { AudioTrackProcessor } from './audio_processor';

vi.mock('./audio_hijack_worklet', () => ({
    importWorkletUrl: vi.fn(() => 'mock-worklet.js'),
}));

describe('AudioTrackProcessor', () => {
    const originalSelf = globalThis.self;
    const originalConsoleWarn = console.warn;
    const originalConsoleDebug = console.debug;
    const originalAudioContext = globalThis.AudioContext;
    const originalMediaStream = globalThis.MediaStream;
    const originalMediaStreamAudioSourceNode = globalThis.MediaStreamAudioSourceNode;
    const originalGainNode = globalThis.GainNode;
    const originalAudioWorkletNode = globalThis.AudioWorkletNode;
    const originalAudioData = globalThis.AudioData;

    beforeEach(() => {
        (globalThis as any).self = {};
        console.warn = vi.fn();
        console.debug = vi.fn();
        (globalThis as any).AudioContext = vi.fn();
        (globalThis as any).MediaStream = vi.fn();
        (globalThis as any).MediaStreamAudioSourceNode = vi.fn();
        (globalThis as any).GainNode = vi.fn();
        (globalThis as any).AudioWorkletNode = vi.fn();
        (globalThis as any).AudioData = vi.fn();
    });

    afterEach(() => {
        if (originalSelf === undefined) {
            delete (globalThis as any).self;
        } else {
            (globalThis as any).self = originalSelf;
        }
        console.warn = originalConsoleWarn;
        console.debug = originalConsoleDebug;
        if (originalAudioContext === undefined) {
            delete (globalThis as any).AudioContext;
        } else {
            (globalThis as any).AudioContext = originalAudioContext;
        }
        if (originalMediaStream === undefined) {
            delete (globalThis as any).MediaStream;
        } else {
            (globalThis as any).MediaStream = originalMediaStream;
        }
        if (originalMediaStreamAudioSourceNode === undefined) {
            delete (globalThis as any).MediaStreamAudioSourceNode;
        } else {
            (globalThis as any).MediaStreamAudioSourceNode = originalMediaStreamAudioSourceNode;
        }
        if (originalGainNode === undefined) {
            delete (globalThis as any).GainNode;
        } else {
            (globalThis as any).GainNode = originalGainNode;
        }
        if (originalAudioWorkletNode === undefined) {
            delete (globalThis as any).AudioWorkletNode;
        } else {
            (globalThis as any).AudioWorkletNode = originalAudioWorkletNode;
        }
        if (originalAudioData === undefined) {
            delete (globalThis as any).AudioData;
        } else {
            (globalThis as any).AudioData = originalAudioData;
        }
    });

    test('throws when track settings are unavailable', () => {
        const track = {
            getSettings: vi.fn(() => undefined),
        } as unknown as MediaStreamTrack;

        expect(() => new AudioTrackProcessor(track)).toThrow('track has no settings');
        expect(track.getSettings).toHaveBeenCalled();
    });

    test('creates processor with worklet', async () => {
        const mockContext = {
            audioWorklet: {
                addModule: vi.fn().mockResolvedValue(undefined),
            },
            close: vi.fn(),
            sampleRate: 44100,
        };
        const mockWorkletNode = {
            port: {
                onmessage: vi.fn(),
            },
        };
        const mockGain = {
            connect: vi.fn(),
            disconnect: vi.fn(),
        };
        const mockSource = {
            connect: vi.fn(),
            disconnect: vi.fn(),
        };

        (globalThis as any).AudioContext = vi.fn(() => mockContext);
        (globalThis as any).MediaStream = vi.fn(() => ({}));
        (globalThis as any).MediaStreamAudioSourceNode = vi.fn(() => mockSource);
        (globalThis as any).GainNode = vi.fn(() => mockGain);
        (globalThis as any).AudioWorkletNode = vi.fn(() => mockWorkletNode);
        (globalThis as any).AudioData = vi.fn(() => ({}));

        const track = {
            getSettings: vi.fn(() => ({
                sampleRate: 44100,
                channelCount: 2,
            })),
        } as unknown as MediaStreamTrack;

        const processor = new AudioTrackProcessor(track);

        expect(processor.gain).toBe(mockGain);
        expect(mockContext.audioWorklet.addModule).toHaveBeenCalledWith('mock-worklet.js');
        expect(mockSource.connect).toHaveBeenCalledWith(mockGain);
        expect(console.warn).toHaveBeenCalledWith('Using AudioWorklet polyfill; performance might suffer.');

        // Test readable stream cancel
        const reader = processor.readable.getReader();
        await reader.cancel();
        expect(mockContext.close).toHaveBeenCalled();
        expect(mockGain.disconnect).toHaveBeenCalled();
        expect(mockSource.disconnect).toHaveBeenCalled();
    });

    test('handles worklet messages correctly', async () => {
        const mockContext = {
            audioWorklet: {
                addModule: vi.fn().mockResolvedValue(undefined),
            },
            close: vi.fn(),
            sampleRate: 44100,
        };
        const mockWorkletNode = {
            port: {
                onmessage: null as any,
            },
        };
        const mockGain = {
            connect: vi.fn(),
            disconnect: vi.fn(),
        };
        const mockSource = {
            connect: vi.fn(),
            disconnect: vi.fn(),
        };

        (globalThis as any).AudioContext = vi.fn(() => mockContext);
        (globalThis as any).MediaStream = vi.fn(() => ({}));
        (globalThis as any).MediaStreamAudioSourceNode = vi.fn(() => mockSource);
        (globalThis as any).GainNode = vi.fn(() => mockGain);
        (globalThis as any).AudioWorkletNode = vi.fn(() => mockWorkletNode);
        (globalThis as any).AudioData = vi.fn(() => 'mock-audio-data');

        const track = {
            getSettings: vi.fn(() => ({
                sampleRate: 44100,
                channelCount: 2,
            })),
        } as unknown as MediaStreamTrack;

        const processor = new AudioTrackProcessor(track);

        // Get the reader to start the stream and set up the worklet
        const reader = processor.readable.getReader();

        // Wait for the worklet to be set up (onmessage should be assigned)
        await new Promise(resolve => setTimeout(resolve, 0));

        // Simulate worklet message by calling the onmessage handler that was set
        const mockMessageData = { format: 'f32' as const, sampleRate: 44100, numberOfFrames: 1024, numberOfChannels: 2, data: new Float32Array(2048), timestamp: 0 };
        if (mockWorkletNode.port.onmessage) {
            mockWorkletNode.port.onmessage({ data: mockMessageData });
        }

        // Check that AudioData was created with the message data
        expect((globalThis as any).AudioData).toHaveBeenCalledWith(mockMessageData);

        // Clean up
        await reader.cancel();
    });
});
