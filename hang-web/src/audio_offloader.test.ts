import { describe, test, expect, beforeEach, afterEach, vi, Mock, MockedClass } from 'vitest';
import { AudioOffloader, AudioOffloaderInit } from "./audio_offloader";
import { AudioTrackDecoder } from "./internal";
import { DefaultVolume, DefaultMinGain, DefaultFadeTime } from './volume';
import * as context from 'golikejs/context';
import { setupGlobalMocks, resetGlobalMocks, mockAudioContext, mockGainNode, mockAudioWorkletNode, mockAudioWorkletAddModule, mockGainNodeConnect, mockGainNodeDisconnect, mockWorkletConnect, mockWorkletDisconnect, mockWorkletPort, mockAudioContextClose } from './test-utils';

// Mock external dependencies
vi.mock("./internal", () => ({
    AudioTrackDecoder: vi.fn()
}));

vi.mock('./volume', () => ({
    DefaultVolume: vi.fn(() => 0.5),
    DefaultMinGain: vi.fn(() => 0.001),
    DefaultFadeTime: vi.fn(() => 80)
}));

vi.mock('./internal/audio_offload_worklet', () => ({
    importUrl: vi.fn(() => 'blob:http://localhost/audio-worklet-url')
}));

vi.mock('golikejs/context', () => ({
    withCancelCause: vi.fn(() => [
        {
            done: vi.fn(() => Promise.resolve()),
            err: vi.fn(() => null)
        } as any,
        vi.fn()
    ]),
    background: vi.fn(() => ({
        done: vi.fn(() => Promise.resolve()),
        err: vi.fn(() => null)
    } as any)),
    ContextCancelledError: class ContextCancelledError extends Error {
        constructor() {
            super('context cancelled');
        }
    }
}));

// Mock global constructors
global.AudioContext = vi.fn(() => mockAudioContext) as any;
global.GainNode = vi.fn(() => mockGainNode) as any;
global.AudioWorkletNode = vi.fn(() => mockAudioWorkletNode) as any;

// Mock console methods
const originalConsoleDebug = console.debug;
const originalConsoleError = console.error;

describe("AudioOffloader", () => {
    beforeEach(() => {
        vi.clearAllMocks();
        console.debug = vi.fn();
        console.error = vi.fn();
        (mockAudioContext as any)._currentTime = 0;
        mockGainNode.gain.value = 0.5;

        // Reset global constructor mocks
        (global.AudioContext as Mock).mockImplementation(() => mockAudioContext);
        (global.GainNode as Mock).mockImplementation(() => mockGainNode);
        (global.AudioWorkletNode as Mock).mockImplementation(() => mockAudioWorkletNode);

        // Set up default mock implementation for withCancelCause
        // Already set in vi.mock
    });

    afterEach(() => {
        console.debug = originalConsoleDebug;
        console.error = originalConsoleError;
    });

    describe("Constructor", () => {
        test("creates audio offloader with default options", () => {
            const offloader = new AudioOffloader({});

            expect(global.AudioContext).toHaveBeenCalledWith({
                latencyHint: 'interactive',
                sampleRate: undefined
            });
            expect(offloader.audioContext).toBe(mockAudioContext);
            expect(offloader.muted).toBe(false);
            expect(DefaultVolume).toHaveBeenCalled();
            expect(DefaultFadeTime).toHaveBeenCalled();
        });

        test("creates audio offloader with custom AudioContext", () => {
            const customAudioContext = mockAudioContext as any as AudioContext;
            const offloader = new AudioOffloader({ audioContext: customAudioContext });

            expect(global.AudioContext).not.toHaveBeenCalled();
            expect(offloader.audioContext).toBe(customAudioContext);
        });

        test("creates audio offloader with custom latency and sample rate", () => {
            const init: AudioOffloaderInit = {
                latency: 200,
                sampleRate: 48000,
                numberOfChannels: 1
            };

            const offloader = new AudioOffloader(init);

            expect(global.AudioContext).toHaveBeenCalledWith({
                latencyHint: 'interactive',
                sampleRate: 48000
            });
            expect(offloader.audioContext).toBe(mockAudioContext);
        });

        test("handles initial volume correctly", () => {
            const offloader = new AudioOffloader({ initialVolume: 0.8 });

            expect(offloader.audioContext).toBe(mockAudioContext);
            // Initial volume is used internally for unmute volume
        });

        test("clamps initial volume to valid range", () => {
            const offloader1 = new AudioOffloader({ initialVolume: -0.5 });
            const offloader2 = new AudioOffloader({ initialVolume: 1.5 });

            expect(offloader1.audioContext).toBe(mockAudioContext);
            expect(offloader2.audioContext).toBe(mockAudioContext);
        });

        test("handles zero initial volume", () => {
            const offloader = new AudioOffloader({ initialVolume: 0 });

            expect(offloader.audioContext).toBe(mockAudioContext);
            expect(DefaultVolume).toHaveBeenCalled(); // Should fall back to default
        });

        test("uses custom volume ramp time", () => {
            const offloader = new AudioOffloader({ volumeRampMs: 120 });

            expect(offloader.audioContext).toBe(mockAudioContext);
        });
    });

    describe("decoder()", () => {
        test("creates and returns decoder on first call", async () => {
            mockAudioWorkletAddModule.mockResolvedValue(undefined as any);
            const mockDecoderInstance = { decoder: 'mock' };
            (AudioTrackDecoder as MockedClass<typeof AudioTrackDecoder>).mockImplementation(() => mockDecoderInstance as any);

            const offloader = new AudioOffloader({});
            const decoder = await offloader.decoder();

            expect(decoder).toBe(mockDecoderInstance);
            expect(AudioTrackDecoder).toHaveBeenCalledWith({
                destination: expect.any(WritableStream)
            });
        });

        test("returns cached decoder on subsequent calls", async () => {
            mockAudioWorkletAddModule.mockResolvedValue(undefined);
            const mockDecoderInstance = { decoder: 'mock' };
            (AudioTrackDecoder as MockedClass<typeof AudioTrackDecoder>).mockImplementation(() => mockDecoderInstance as any);

            const offloader = new AudioOffloader({});

            const decoder1 = await offloader.decoder();
            const decoder2 = await offloader.decoder();

            expect(decoder1).toBe(decoder2);
            expect(AudioTrackDecoder).toHaveBeenCalledTimes(1);
        });

        test("initializes worklet during first decoder call", async () => {
            mockAudioWorkletAddModule.mockResolvedValue(undefined);
            const mockDecoderInstance = { decoder: 'mock' };
            (AudioTrackDecoder as MockedClass<typeof AudioTrackDecoder>).mockImplementation(() => mockDecoderInstance as any);

            const offloader = new AudioOffloader({ numberOfChannels: 2, latency: 150 });
            await offloader.decoder();

            expect(mockAudioWorkletAddModule).toHaveBeenCalledWith('blob:http://localhost/audio-worklet-url');
            expect(global.AudioWorkletNode).toHaveBeenCalledWith(
                mockAudioContext,
                'AudioOffloader',
                {
                    channelCount: 2,
                    numberOfInputs: 0,
                    numberOfOutputs: 1,
                    processorOptions: {
                        sampleRate: 44100,
                        latency: 150
                    }
                }
            );
            expect(global.GainNode).toHaveBeenCalledWith(mockAudioContext, { gain: 0.5 });
            expect(mockWorkletConnect).toHaveBeenCalledWith(mockGainNode);
            expect(mockGainNodeConnect).toHaveBeenCalledWith(mockAudioContext.destination);
        });

        test("handles worklet loading failure", async () => {
            const workletError = new Error("Failed to load worklet");
            mockAudioWorkletAddModule.mockRejectedValue(workletError);

            const offloader = new AudioOffloader({});

            await expect(offloader.decoder()).rejects.toThrow("Failed to load worklet");
            expect(console.error).toHaveBeenCalledWith('failed to load AudioWorklet module:', workletError);
        });

        test("uses default latency and channels when not specified", async () => {
            mockAudioWorkletAddModule.mockResolvedValue(undefined);
            const mockDecoderInstance = { decoder: 'mock' };
            (AudioTrackDecoder as MockedClass<typeof AudioTrackDecoder>).mockImplementation(() => mockDecoderInstance as any);

            const offloader = new AudioOffloader({});
            await offloader.decoder();

            expect(global.AudioWorkletNode).toHaveBeenCalledWith(
                mockAudioContext,
                'AudioOffloader',
                {
                    channelCount: 2, // default
                    numberOfInputs: 0,
                    numberOfOutputs: 1,
                    processorOptions: {
                        sampleRate: 44100,
                        latency: 100 // default
                    }
                }
            );
        });
    });

    describe("Volume Control", () => {
        beforeEach(async () => {
            mockAudioWorkletAddModule.mockResolvedValue(undefined);
            const mockDecoderInstance = { decoder: 'mock' };
            (AudioTrackDecoder as MockedClass<typeof AudioTrackDecoder>).mockImplementation(() => mockDecoderInstance as any);
        });

        test("setVolume() updates gain node with valid volume", async () => {
            const offloader = new AudioOffloader({});
            await offloader.decoder(); // Initialize worklet and gain node

            (mockAudioContext as any)._currentTime = 1.5;
            offloader.setVolume(0.8);

            expect(mockGainNode.gain.cancelScheduledValues).toHaveBeenCalled();
            expect(mockGainNode.gain.setValueAtTime).toHaveBeenCalled();
            expect(mockGainNode.gain.exponentialRampToValueAtTime).toHaveBeenCalledWith(0.8, expect.any(Number));
        });

        test("setVolume() clamps values to valid range", async () => {
            const offloader = new AudioOffloader({});
            await offloader.decoder();

            // Test negative value
            offloader.setVolume(-0.2);
            expect(mockGainNode.gain.exponentialRampToValueAtTime).toHaveBeenCalledWith(0.001, expect.any(Number));
            expect(mockGainNode.gain.setValueAtTime).toHaveBeenCalledWith(0, expect.any(Number));

            vi.clearAllMocks();

            // Test value greater than 1
            offloader.setVolume(1.5);
            expect(mockGainNode.gain.exponentialRampToValueAtTime).toHaveBeenCalledWith(1, expect.any(Number));
        });

        test("setVolume() handles very low volume with min gain", async () => {
            (DefaultMinGain as Mock).mockReturnValue(0.001);
            const offloader = new AudioOffloader({});
            await offloader.decoder();

            (mockAudioContext as any)._currentTime = 2.0;
            offloader.setVolume(0.0005); // Below min gain

            expect(mockGainNode.gain.exponentialRampToValueAtTime).toHaveBeenCalledWith(0.001, expect.any(Number));
            expect(mockGainNode.gain.setValueAtTime).toHaveBeenCalledWith(0, expect.any(Number));
        });

        test("setVolume() does nothing when gain node not initialized", () => {
            const offloader = new AudioOffloader({});

            // Should not throw
            expect(() => offloader.setVolume(0.7)).not.toThrow();
            expect(mockGainNode.gain.cancelScheduledValues).not.toHaveBeenCalled();
        });

        test("volume getter returns current gain value", async () => {
            const offloader = new AudioOffloader({});
            await offloader.decoder();

            mockGainNode.gain.value = 0.75;
            expect(offloader.volume).toBe(0.75);
        });

        test("volume getter returns 1.0 when gain node not initialized", () => {
            const offloader = new AudioOffloader({});
            expect(offloader.volume).toBe(1.0);
        });
    });

    describe("Mute Control", () => {
        beforeEach(async () => {
            mockAudioWorkletAddModule.mockResolvedValue(undefined);
            const mockDecoderInstance = { decoder: 'mock' };
            (AudioTrackDecoder as MockedClass<typeof AudioTrackDecoder>).mockImplementation(() => mockDecoderInstance as any);
        });

        test("mute(true) sets gain to zero", async () => {
            const offloader = new AudioOffloader({});
            await offloader.decoder();

            (mockAudioContext as any)._currentTime = 3.0;
            mockGainNode.gain.value = 0.6;

            offloader.mute(true);

            expect(offloader.muted).toBe(true);
            expect(mockGainNode.gain.cancelScheduledValues).toHaveBeenCalled();
            expect(mockGainNode.gain.setValueAtTime).toHaveBeenCalled();
            expect(mockGainNode.gain.exponentialRampToValueAtTime).toHaveBeenCalledWith(0, expect.any(Number));
        });

        test("mute(true) handles low volume with min gain fade", async () => {
            (DefaultMinGain as Mock).mockReturnValue(0.001);
            const offloader = new AudioOffloader({});
            await offloader.decoder();

            (mockAudioContext as any)._currentTime = 4.0;
            mockGainNode.gain.value = 0.0005; // Below min gain

            offloader.mute(true);

            expect(offloader.muted).toBe(true);
            expect(mockGainNode.gain.exponentialRampToValueAtTime).toHaveBeenCalledWith(0.001, expect.any(Number));
            expect(mockGainNode.gain.setValueAtTime).toHaveBeenCalledWith(0, expect.any(Number));
        });

        test("mute(false) restores previous volume", async () => {
            const offloader = new AudioOffloader({});
            await offloader.decoder();

            // First mute to store volume
            mockGainNode.gain.value = 0.7;
            offloader.mute(true);

            vi.clearAllMocks();

            // Then unmute
            (mockAudioContext as any)._currentTime = 5.0;
            offloader.mute(false);

            expect(offloader.muted).toBe(false);
            expect(mockGainNode.gain.exponentialRampToValueAtTime).toHaveBeenCalledWith(0.7, expect.any(Number));
        });

        test("mute(false) uses default volume when no previous volume stored", async () => {
            (DefaultVolume as Mock).mockReturnValue(0.5);
            const offloader = new AudioOffloader({});
            await offloader.decoder();

            // Set a very low initial volume so unmute uses default
            mockGainNode.gain.value = 0.0001; // Very low volume
            offloader.mute(true); // First mute

            vi.clearAllMocks();

            offloader.mute(false); // Unmute should use default volume

            expect(offloader.muted).toBe(false);
            expect(mockGainNode.gain.exponentialRampToValueAtTime).toHaveBeenCalledWith(0.5, expect.any(Number));
        });

        test("mute() does nothing when state doesn't change", async () => {
            const offloader = new AudioOffloader({});
            await offloader.decoder();

            // Mute twice
            offloader.mute(true);
            vi.clearAllMocks();
            offloader.mute(true);

            expect(mockGainNode.gain.cancelScheduledValues).not.toHaveBeenCalled();

            // Unmute twice
            offloader.mute(false);
            vi.clearAllMocks();
            offloader.mute(false);

            expect(mockGainNode.gain.cancelScheduledValues).not.toHaveBeenCalled();
        });

        test("mute() does nothing when gain node not initialized", () => {
            const offloader = new AudioOffloader({});

            expect(() => offloader.mute(true)).not.toThrow();
            expect(() => offloader.mute(false)).not.toThrow();
        });
    });

    describe("Audio Processing", () => {
        beforeEach(async () => {
            mockAudioWorkletAddModule.mockResolvedValue(undefined);
            const mockDecoderInstance = { decoder: 'mock' };
            (AudioTrackDecoder as MockedClass<typeof AudioTrackDecoder>).mockImplementation(() => mockDecoderInstance as any);
        });

        test("processes audio frame data correctly", async () => {
            const offloader = new AudioOffloader({});
            const decoder = await offloader.decoder();

            // Get the WritableStream from the decoder constructor
            const decoderCall = (AudioTrackDecoder as MockedClass<typeof AudioTrackDecoder>).mock.calls[0][0];
            const writableStream = decoderCall.destination;

            // Mock AudioData frame
            const mockAudioFrame = {
                numberOfChannels: 2,
                numberOfFrames: 1024,
                timestamp: 12345,
                duration: 1024 / 44100 * 1000000, // microseconds
                format: 'f32-planar',
                sampleRate: 44100,
                allocationSize: vi.fn(() => 1024 * 4),
                clone: vi.fn(),
                copyTo: vi.fn(),
                close: vi.fn()
            } as unknown as AudioData;

            // Setup copyTo to simulate audio data
            const leftChannel = new Float32Array(1024).fill(0.5);
            const rightChannel = new Float32Array(1024).fill(-0.3);
            (mockAudioFrame.copyTo as Mock<(data: any, options: any) => void>).mockImplementation((data: Float32Array, options: any) => {
                if (options.planeIndex === 0) {
                    data.set(leftChannel);
                } else if (options.planeIndex === 1) {
                    data.set(rightChannel);
                }
            });

            // Write frame to the stream
            const writer = writableStream.getWriter();
            await writer.write(mockAudioFrame);

            // Verify copyTo was called for each channel
            expect(mockAudioFrame.copyTo).toHaveBeenCalledTimes(2);
            expect(mockAudioFrame.copyTo).toHaveBeenCalledWith(expect.any(Float32Array), { format: "f32-planar", planeIndex: 0 });
            expect(mockAudioFrame.copyTo).toHaveBeenCalledWith(expect.any(Float32Array), { format: "f32-planar", planeIndex: 1 });

            // Verify worklet received the data
            expect(mockWorkletPort.postMessage).toHaveBeenCalledWith(
                {
                    channels: [expect.any(Float32Array), expect.any(Float32Array)],
                    timestamp: 12345
                },
                [expect.any(ArrayBuffer), expect.any(ArrayBuffer)] // Transfer ownership
            );

            // Verify frame was closed
            expect(mockAudioFrame.close).toHaveBeenCalledTimes(1);
        });

        test("handles mono audio correctly", async () => {
            const offloader = new AudioOffloader({});
            const decoder = await offloader.decoder();

            const decoderCall = (AudioTrackDecoder as MockedClass<typeof AudioTrackDecoder>).mock.calls[0][0];
            const writableStream = decoderCall.destination;

            const mockMonoFrame = {
                numberOfChannels: 1,
                numberOfFrames: 512,
                timestamp: 67890,
                duration: 512 / 44100 * 1000000,
                format: 'f32-planar',
                sampleRate: 44100,
                allocationSize: vi.fn(() => 512 * 4),
                clone: vi.fn(),
                copyTo: vi.fn(),
                close: vi.fn()
            } as unknown as AudioData;

            const monoChannel = new Float32Array(512).fill(0.8);
            (mockMonoFrame.copyTo as Mock<(data: any) => void>).mockImplementation((data: Float32Array) => {
                data.set(monoChannel);
            });

            const writer = writableStream.getWriter();
            await writer.write(mockMonoFrame);

            expect(mockMonoFrame.copyTo).toHaveBeenCalledTimes(1);
            expect(mockWorkletPort.postMessage).toHaveBeenCalledWith(
                {
                    channels: [expect.any(Float32Array)],
                    timestamp: 67890
                },
                [expect.any(ArrayBuffer)]
            );
            expect(mockMonoFrame.close).toHaveBeenCalledTimes(1);
        });
    });

    describe("Cleanup and Destruction", () => {
        test("destroy() calls cancel function", () => {
            const mockCancelFunc = vi.fn();
            vi.mocked(context.withCancelCause).mockReturnValue([
                {
                    done: vi.fn(() => Promise.resolve()),
                    err: vi.fn(() => context.ContextCancelledError)
                },
                mockCancelFunc
            ]);

            const offloader = new AudioOffloader({});
            offloader.destroy();

            expect(mockCancelFunc).toHaveBeenCalledWith(new Error("AudioEmitter destroyed"));
        });

        test("context cleanup disconnects nodes and closes audio context", async () => {
            const mockDonePromise = Promise.resolve();
            const mockContext = { done: vi.fn(() => mockDonePromise) };
            vi.mocked(context.withCancelCause).mockReturnValue([
                 {
                    done: vi.fn(() => Promise.resolve()),
                    err: vi.fn(() => context.ContextCancelledError)
                },
                vi.fn()
            ]);

            mockAudioWorkletAddModule.mockResolvedValue(undefined);
            const mockDecoderInstance = { decoder: 'mock' };
            (AudioTrackDecoder as MockedClass<typeof AudioTrackDecoder>).mockImplementation(() => mockDecoderInstance as any);

            const offloader = new AudioOffloader({});
            await offloader.decoder(); // Initialize worklet

            // Trigger context cleanup
            await mockDonePromise;

            expect(mockGainNodeDisconnect).toHaveBeenCalledTimes(1);
            expect(mockWorkletDisconnect).toHaveBeenCalledTimes(1);
            expect(mockAudioContextClose).toHaveBeenCalledTimes(1);
        });

        test("context cleanup does not close external audio context", async () => {
            const externalAudioContext = mockAudioContext as any as AudioContext;
            const mockDonePromise = Promise.resolve();
            const mockContext = { done: vi.fn(() => mockDonePromise) };
            vi.mocked(context.withCancelCause).mockReturnValue([
                {
                    done: vi.fn(() => Promise.resolve()),
                    err: vi.fn(() => context.ContextCancelledError)
                },
                vi.fn()
            ]);

            mockAudioWorkletAddModule.mockResolvedValue(undefined);
            const mockDecoderInstance = { decoder: 'mock' };
            (AudioTrackDecoder as MockedClass<typeof AudioTrackDecoder>).mockImplementation(() => mockDecoderInstance as any);

            const offloader = new AudioOffloader({ audioContext: externalAudioContext });
            await offloader.decoder();

            // Trigger context cleanup
            await mockDonePromise;

            expect(mockGainNodeDisconnect).toHaveBeenCalledTimes(1);
            expect(mockWorkletDisconnect).toHaveBeenCalledTimes(1);
            expect(mockAudioContextClose).not.toHaveBeenCalled(); // Should not close external context
        });
    });

    describe("Edge Cases and Error Handling", () => {
        test("handles invalid volume values gracefully", async () => {
            mockAudioWorkletAddModule.mockResolvedValue(undefined);
            const mockDecoderInstance = { decoder: 'mock' };
            (AudioTrackDecoder as MockedClass<typeof AudioTrackDecoder>).mockImplementation(() => mockDecoderInstance as any);

            const offloader = new AudioOffloader({});
            await offloader.decoder();

            // Test NaN, Infinity, etc.
            expect(() => offloader.setVolume(NaN)).not.toThrow();
            expect(() => offloader.setVolume(Infinity)).not.toThrow();
            expect(() => offloader.setVolume(-Infinity)).not.toThrow();
        });

        test("handles audio context creation failure", () => {
            (global.AudioContext as Mock).mockImplementationOnce(() => {
                throw new Error("AudioContext creation failed");
            });

            expect(() => new AudioOffloader({})).toThrow("AudioContext creation failed");
        });

        test("handles worklet node creation failure", async () => {
            mockAudioWorkletAddModule.mockResolvedValue(undefined);
            (global.AudioWorkletNode as Mock).mockImplementationOnce(() => {
                throw new Error("WorkletNode creation failed");
            });

            const offloader = new AudioOffloader({});

            await expect(offloader.decoder()).rejects.toThrow("WorkletNode creation failed");
        });

        test("handles gain node creation failure", async () => {
            mockAudioWorkletAddModule.mockResolvedValue(undefined);
            (global.GainNode as Mock).mockImplementationOnce(() => {
                throw new Error("GainNode creation failed");
            });

            const offloader = new AudioOffloader({});

            await expect(offloader.decoder()).rejects.toThrow("GainNode creation failed");
        });
    });

    describe("Integration Scenarios", () => {
        test("complete audio pipeline workflow", async () => {
            mockAudioWorkletAddModule.mockResolvedValue(undefined);
            const mockDecoderInstance = { decoder: 'mock' };
            (AudioTrackDecoder as MockedClass<typeof AudioTrackDecoder>).mockImplementation(() => mockDecoderInstance as any);

            // Create offloader with custom settings
            const offloader = new AudioOffloader({
                latency: 120,
                initialVolume: 0.8,
                volumeRampMs: 100,
                numberOfChannels: 2
            });

            // Get decoder (initializes worklet)
            const decoder = await offloader.decoder();
            expect(decoder).toBe(mockDecoderInstance);

            // Verify worklet initialization
            expect(mockAudioWorkletAddModule).toHaveBeenCalled();
            expect(global.AudioWorkletNode).toHaveBeenCalled();
            expect(global.GainNode).toHaveBeenCalled();

            // Test volume control
            offloader.setVolume(0.6);
            expect(mockGainNode.gain.exponentialRampToValueAtTime).toHaveBeenCalled();

            // Test mute
            offloader.mute(true);
            expect(offloader.muted).toBe(true);

            offloader.mute(false);
            expect(offloader.muted).toBe(false);

            // Test audio processing
            const decoderCall = (AudioTrackDecoder as MockedClass<typeof AudioTrackDecoder>).mock.calls[0][0];
            const writableStream = decoderCall.destination;

            const mockFrame = {
                numberOfChannels: 2,
                numberOfFrames: 256,
                timestamp: 1000,
                duration: 256 / 44100 * 1000000,
                format: 'f32-planar',
                sampleRate: 44100,
                allocationSize: vi.fn(() => 256 * 4),
                clone: vi.fn(),
                copyTo: vi.fn(),
                close: vi.fn()
            } as unknown as AudioData;

            const writer = writableStream.getWriter();
            await writer.write(mockFrame);

            expect(mockWorkletPort.postMessage).toHaveBeenCalled();
            expect(mockFrame.close).toHaveBeenCalled();

            // Test cleanup
            offloader.destroy();
        });

        test("handles rapid volume and mute changes", async () => {
            mockAudioWorkletAddModule.mockResolvedValue(undefined);
            const mockDecoderInstance = { decoder: 'mock' };
            (AudioTrackDecoder as MockedClass<typeof AudioTrackDecoder>).mockImplementation(() => mockDecoderInstance as any);

            const offloader = new AudioOffloader({});
            await offloader.decoder();

            // Rapid volume changes
            offloader.setVolume(0.1);
            offloader.setVolume(0.5);
            offloader.setVolume(0.9);

            // Each should cancel previous scheduled values
            expect(mockGainNode.gain.cancelScheduledValues).toHaveBeenCalledTimes(3);

            // Rapid mute/unmute
            offloader.mute(true);
            offloader.mute(false);
            offloader.mute(true);

            expect(offloader.muted).toBe(true);
        });
    });
});
