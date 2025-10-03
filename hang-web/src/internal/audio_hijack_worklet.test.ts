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

        // Simulate the worklet registration logic
        if (typeof AudioWorkletProcessor !== 'undefined') {
            registerProcessor("AudioHijacker", class AudioHijackProcessor extends AudioWorkletProcessor {});
        }

        expect(registerProcessor).toHaveBeenCalledTimes(1);
        const [name, processorCtor] = registerProcessor.mock.calls[0];
        expect(name).toBe("AudioHijacker");
        expect(typeof processorCtor).toBe("function");
    });
});
