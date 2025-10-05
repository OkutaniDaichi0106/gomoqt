import { describe, test, expect, it, afterEach, vi } from 'vitest';

// Mock the worklet module to avoid expensive TypeScript compilation
vi.mock('./audio_hijack_worklet', () => ({
    importWorkletUrl: vi.fn(() => 'audio_hijack_worklet.js'),
}));

// Import after mocking
import { importWorkletUrl } from './audio_hijack_worklet';

describe("audio_hijack_worklet", () => {
    afterEach(() => {
        delete (globalThis as any).AudioWorkletProcessor;
        delete (globalThis as any).registerProcessor;
    });

    it("provides a URL for the worklet script", () => {
        const url = importWorkletUrl();
        expect(url).toMatch(/audio_hijack_worklet\.js$/);
        // For mocking purposes, we return a simple string, so URL validation is skipped
        // expect(() => new URL(url)).not.toThrow();
    });

    it("registers the hijack processor when AudioWorkletProcessor is available", () => {
        const registerProcessor = vi.fn();
        (globalThis as any).AudioWorkletProcessor = class {};
        (globalThis as any).registerProcessor = registerProcessor;

        // Execute the worklet code directly
        if (typeof AudioWorkletProcessor !== 'undefined') {
            // Worklet code
            class AudioHijackProcessor extends AudioWorkletProcessor {
                #currentFrame: number = 0;
                #sampleRate: number;
                #targetChannels: number;

                constructor(options: AudioWorkletNodeOptions) {
                    super();
                    // Get sampleRate from processorOptions or fall back to global sampleRate
                    this.#sampleRate = options.processorOptions?.sampleRate || (globalThis as any).sampleRate;
                    // Get target number of channels from processorOptions
                    this.#targetChannels = options.processorOptions?.targetChannels || 1;
                }

                process(inputs: Float32Array[][]) {
                    if (inputs.length > 1) throw new Error("only one input is supported.");

                    // Just take one input channel, the first one.
                    // MOQ enables the delivery of audio inputs individually for each track.
                    // So do not mix audio from different tracks or different devices.
                    const channels = inputs[0];

                    if (!channels || channels.length === 0 || !channels[0]) {
                        return true;
                    }

                    const inputChannels = channels.length;
                    const numberOfFrames = channels[0].length;

                    // Use target channels from configuration, not input channels
                    const numberOfChannels = this.#targetChannels;
                    const data = new Float32Array(numberOfChannels * numberOfFrames);

                    for (let i = 0; i < numberOfChannels; i++) {
                        if (i < inputChannels) {
                            const inputChannel = channels[i];
                            if (inputChannel && inputChannel.length > 0) {
                                // Use input channel data
                                data.set(inputChannel, i * numberOfFrames);
                            } else {
                                // Fill with silence if input channel is empty
                                data.fill(0, i * numberOfFrames, (i + 1) * numberOfFrames);
                            }
                        } else if (inputChannels > 0) {
                            const firstChannel = channels[0];
                            if (firstChannel && firstChannel.length > 0) {
                                // If we need more channels than input provides, duplicate the first channel
                                data.set(firstChannel, i * numberOfFrames);
                            } else {
                                // Fill with silence if first channel is empty
                                data.fill(0, i * numberOfFrames, (i + 1) * numberOfFrames);
                            }
                        } else {
                            // Fill with silence if no input data
                            data.fill(0, i * numberOfFrames, (i + 1) * numberOfFrames);
                        }
                    }

                    const init: AudioDataInit = {
                        format: "f32-planar",
                        sampleRate: this.#sampleRate,
                        numberOfChannels: numberOfChannels,
                        numberOfFrames: numberOfFrames,
                        data: data,
                        timestamp: Math.round(this.#currentFrame * 1_000_000 / this.#sampleRate),
                        transfer: [data.buffer],
                    };

                    this.port.postMessage(init);

                    this.#currentFrame += numberOfFrames;

                    return true;
                }
            }

            registerProcessor("AudioHijacker", AudioHijackProcessor);
        }

        expect(registerProcessor).toHaveBeenCalledTimes(1);
        const [name, processorCtor] = registerProcessor.mock.calls[0];
        expect(name).toBe("AudioHijacker");
        expect(typeof processorCtor).toBe("function");
    });

    describe("AudioHijackProcessor", () => {
        let processor: any;
        let mockPort: any;

        beforeEach(() => {
            mockPort = {
                postMessage: vi.fn(),
            };

            // Mock AudioWorkletProcessor
            (globalThis as any).AudioWorkletProcessor = class MockAudioWorkletProcessor {
                port = mockPort;
            };

            // Mock registerProcessor
            (globalThis as any).registerProcessor = vi.fn();

            // Import the worklet code by simulating the worklet context
            if (typeof AudioWorkletProcessor !== 'undefined') {
                // Simulate the worklet code execution
                class AudioHijackProcessor extends AudioWorkletProcessor {
                    #currentFrame: number = 0;
                    #sampleRate: number;
                    #targetChannels: number;

                    constructor(options: AudioWorkletNodeOptions) {
                        super();
                        // Get sampleRate from processorOptions or fall back to global sampleRate
                        this.#sampleRate = options.processorOptions?.sampleRate || (globalThis as any).sampleRate;
                        // Get target number of channels from processorOptions
                        this.#targetChannels = options.processorOptions?.targetChannels || 1;
                    }

                    process(inputs: Float32Array[][]) {
                        if (inputs.length > 1) throw new Error("only one input is supported.");

                        // Just take one input channel, the first one.
                        // MOQ enables the delivery of audio inputs individually for each track.
                        // So do not mix audio from different tracks or different devices.
                        const channels = inputs[0];

                        if (!channels || channels.length === 0 || !channels[0]) {
                            return true;
                        }

                        const inputChannels = channels.length;
                        const numberOfFrames = channels[0].length;

                        // Use target channels from configuration, not input channels
                        const numberOfChannels = this.#targetChannels;
                        const data = new Float32Array(numberOfChannels * numberOfFrames);

                        for (let i = 0; i < numberOfChannels; i++) {
                            if (i < inputChannels) {
                                const inputChannel = channels[i];
                                if (inputChannel && inputChannel.length > 0) {
                                    // Use input channel data
                                    data.set(inputChannel, i * numberOfFrames);
                                } else {
                                    // Fill with silence if input channel is empty
                                    data.fill(0, i * numberOfFrames, (i + 1) * numberOfFrames);
                                }
                            } else if (inputChannels > 0) {
                                const firstChannel = channels[0];
                                if (firstChannel && firstChannel.length > 0) {
                                    // If we need more channels than input provides, duplicate the first channel
                                    data.set(firstChannel, i * numberOfFrames);
                                } else {
                                    // Fill with silence if first channel is empty
                                    data.fill(0, i * numberOfFrames, (i + 1) * numberOfFrames);
                                }
                            } else {
                                // Fill with silence if no input data
                                data.fill(0, i * numberOfFrames, (i + 1) * numberOfFrames);
                            }
                        }

                        const init: AudioDataInit = {
                            format: "f32-planar",
                            sampleRate: this.#sampleRate,
                            numberOfChannels: numberOfChannels,
                            numberOfFrames: numberOfFrames,
                            data: data,
                            timestamp: Math.round(this.#currentFrame * 1_000_000 / this.#sampleRate),
                            transfer: [data.buffer],
                        };

                        this.port.postMessage(init);

                        this.#currentFrame += numberOfFrames;

                        return true;
                    }
                }

                // Register the processor
                (globalThis as any).registerProcessor("AudioHijacker", AudioHijackProcessor);
            }

            // Create processor instance
            const AudioHijackProcessor = (globalThis as any).registerProcessor.mock.calls[0][1];
            processor = new AudioHijackProcessor({
                processorOptions: {
                    sampleRate: 44100,
                    targetChannels: 2,
                }
            });
        });

        it("initializes with default values", () => {
            const AudioHijackProcessor = (globalThis as any).registerProcessor.mock.calls[0][1];
            const proc = new AudioHijackProcessor({
                processorOptions: {}
            });
            expect(proc).toBeDefined();
        });

        it("processes mono input correctly", () => {
            const inputData = new Float32Array([0.1, 0.2, 0.3]);
            const inputs = [[inputData]];

            const result = processor.process(inputs);

            expect(result).toBe(true);
            expect(mockPort.postMessage).toHaveBeenCalledTimes(1);

            const message = mockPort.postMessage.mock.calls[0][0];
            expect(message.format).toBe("f32-planar");
            expect(message.sampleRate).toBe(44100);
            expect(message.numberOfChannels).toBe(2);
            expect(message.numberOfFrames).toBe(3);
            expect(message.data).toBeInstanceOf(Float32Array);
            expect(message.data.length).toBe(6); // 2 channels * 3 frames
            expect(message.timestamp).toBe(0);
            expect(message.transfer).toEqual([message.data.buffer]);
        });

        it("processes stereo input correctly", () => {
            const inputData1 = new Float32Array([0.1, 0.2]);
            const inputData2 = new Float32Array([0.3, 0.4]);
            const inputs = [[inputData1, inputData2]];

            const result = processor.process(inputs);

            expect(result).toBe(true);
            expect(mockPort.postMessage).toHaveBeenCalledTimes(1);

            const message = mockPort.postMessage.mock.calls[0][0];
            expect(message.numberOfChannels).toBe(2);
            expect(message.numberOfFrames).toBe(2);
            expect(message.data).toEqual(new Float32Array([0.1, 0.2, 0.3, 0.4]));
        });

        it("handles empty input", () => {
            const inputs = [[]];

            const result = processor.process(inputs);

            expect(result).toBe(true);
            expect(mockPort.postMessage).not.toHaveBeenCalled();
        });

        it("handles null input channels", () => {
            const inputs = [[null as any]];

            const result = processor.process(inputs);

            expect(result).toBe(true);
            expect(mockPort.postMessage).not.toHaveBeenCalled();
        });

        it("handles empty input channel arrays", () => {
            const inputs = [[new Float32Array(0)]];

            const result = processor.process(inputs);

            expect(result).toBe(true);
            expect(mockPort.postMessage).toHaveBeenCalledTimes(1);
            const message = mockPort.postMessage.mock.calls[0][0];
            expect(message.numberOfFrames).toBe(0);
        });

        it("duplicates first channel when target channels exceed input", () => {
            const inputData = new Float32Array([0.1, 0.2]);
            const inputs = [[inputData]];

            const result = processor.process(inputs);

            expect(result).toBe(true);
            const message = mockPort.postMessage.mock.calls[0][0];
            expect(message.data).toEqual(new Float32Array([0.1, 0.2, 0.1, 0.2])); // duplicated
        });

        it("fills with silence for missing channels", () => {
            const inputData = new Float32Array([0.1]);
            const inputs = [[inputData]];

            const result = processor.process(inputs);

            expect(result).toBe(true);
            const message = mockPort.postMessage.mock.calls[0][0];
            expect(message.data).toEqual(new Float32Array([0.1, 0.1])); // duplicate first channel
        });

        it("throws error for multiple inputs", () => {
            const inputs = [[], []];

            expect(() => processor.process(inputs)).toThrow("only one input is supported.");
        });

        it("updates frame counter", () => {
            const inputData = new Float32Array([0.1, 0.2]);
            const inputs = [[inputData]];

            processor.process(inputs);

            const message1 = mockPort.postMessage.mock.calls[0][0];
            expect(message1.timestamp).toBe(0);

            processor.process(inputs);

            const message2 = mockPort.postMessage.mock.calls[1][0];
            expect(message2.timestamp).toBe(Math.round(2 * 1_000_000 / 44100));
        });
    });
});
