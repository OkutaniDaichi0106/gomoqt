import { describe, test, expect, beforeEach, afterEach, vi } from 'vitest';
import { AudioTrackProcessor } from './audio_processor';

vi.mock('./audio_hijack_worklet', () => ({
    importWorkletUrl: vi.fn(() => 'mock-worklet.js'),
}));

describe('AudioTrackProcessor', () => {
    const originalSelf = globalThis.self;
    const originalConsoleWarn = console.warn;

    beforeEach(() => {
        (globalThis as any).self = {};
        console.warn = vi.fn();
    });

    afterEach(() => {
        if (originalSelf === undefined) {
            delete (globalThis as any).self;
        } else {
            (globalThis as any).self = originalSelf;
        }
        console.warn = originalConsoleWarn;
    });

    test('throws when track settings are unavailable', () => {
        const track = {
            getSettings: vi.fn(() => undefined),
        } as unknown as MediaStreamTrack;

    expect(() => new AudioTrackProcessor(track)).toThrow('track has no settings');
        expect(track.getSettings).toHaveBeenCalled();
    });
});
