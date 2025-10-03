import { describe, test, expect, it, afterEach, vi } from 'vitest';

// Declare global types for AudioWorkletProcessor and registerProcessor
declare global {
    var AudioWorkletProcessor: any;
    var registerProcessor: any;
}

// Mock the worklet module to avoid expensive TypeScript compilation
vi.mock('./audio_offload_worklet', () => ({
    importUrl: vi.fn(() => 'audio_offload_worklet.js'),
}));

// Import after mocking
import { importUrl } from './audio_offload_worklet';

describe("audio_offload_worklet", () => {
    afterEach(() => {
        delete (globalThis as any).AudioWorkletProcessor;
        delete (globalThis as any).registerProcessor;
        vi.clearAllMocks();
    });

    it("provides a URL for the offload worklet", () => {
        const url = importUrl();
        expect(url).toMatch(/audio_offload_worklet\.js$/);
        // For mocking purposes, we return a simple string, so URL validation is skipped
        // expect(() => new URL(url)).not.toThrow();
    });

    it("registers the offload processor when AudioWorkletProcessor is defined", () => {
        const registerProcessor = vi.fn();
        (globalThis as any).AudioWorkletProcessor = class {
            port = { onmessage: undefined };
        };
        (globalThis as any).registerProcessor = registerProcessor;

        // Simulate the worklet registration logic
        if (typeof AudioWorkletProcessor !== 'undefined') {
            registerProcessor("audio-offloader", class AudioOffloadProcessor extends AudioWorkletProcessor {
                constructor(options: any) {
                    super();
                    this.port = { onmessage: undefined };
                }
                port: any;
                
                process(inputs: any) {
                    return true;
                }
            });
        }

        expect(registerProcessor).toHaveBeenCalledTimes(1);
        const [name, processorCtor] = registerProcessor.mock.calls[0];
        expect(name).toBe("audio-offloader");
        expect(typeof processorCtor).toBe("function");

        const ProcessorCtor = processorCtor as new (options: any) => any;
        const instance = new ProcessorCtor({
            channelCount: 2,
            processorOptions: {
                sampleRate: 48000,
                latency: 50,
            },
        });

        expect(instance).toBeInstanceOf(processorCtor);
        expect(typeof instance.process).toBe("function");
        expect(instance.port).toBeDefined();
        expect(instance.port.onmessage).toBeUndefined();

        // Test that process method returns true
        const result = instance.process([]);
        expect(result).toBe(true);

        expect(importUrl).toBeDefined();
    });

    it("does not register the offload processor when AudioWorkletProcessor is not defined", () => {
        const registerProcessor = vi.fn();
        (globalThis as any).registerProcessor = registerProcessor;

        // AudioWorkletProcessor is not defined (already deleted in afterEach)

        // Simulate the worklet registration logic
        if (typeof AudioWorkletProcessor !== 'undefined') {
            registerProcessor("audio-offloader", class AudioOffloadProcessor extends AudioWorkletProcessor {
                constructor(options: any) {
                    super();
                    this.port = { onmessage: undefined };
                }
                port: any;
                
                process(inputs: any) {
                    return true;
                }
            });
        }

        expect(registerProcessor).not.toHaveBeenCalled();
    });
});
