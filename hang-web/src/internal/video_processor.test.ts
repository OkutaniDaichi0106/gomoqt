import { describe, test, expect, beforeEach, afterEach, vi } from 'vitest';
import { VideoTrackProcessor } from './video_processor';

describe('VideoTrackProcessor', () => {
    const originalSelf = globalThis.self;
    const originalDocument = globalThis.document;
    const originalPerformance = globalThis.performance;
    const originalRequestAnimationFrame = globalThis.requestAnimationFrame;
    const originalConsoleWarn = console.warn;

    beforeEach(() => {
        (globalThis as any).self = {};
        (globalThis as any).document = {
            createElement: vi.fn(() => ({
                play: vi.fn(async () => undefined),
                srcObject: null,
                onloadedmetadata: null,
                videoWidth: 0,
                videoHeight: 0,
            })),
        };
        (globalThis as any).performance = {
            now: vi.fn(() => 0),
        };
        (globalThis as any).requestAnimationFrame = vi.fn();
        console.warn = vi.fn();
    });

    afterEach(() => {
        if (originalSelf === undefined) {
            delete (globalThis as any).self;
        } else {
            (globalThis as any).self = originalSelf;
        }

        if (originalDocument === undefined) {
            delete (globalThis as any).document;
        } else {
            (globalThis as any).document = originalDocument;
        }

        if (originalPerformance === undefined) {
            delete (globalThis as any).performance;
        } else {
            (globalThis as any).performance = originalPerformance;
        }

        if (originalRequestAnimationFrame === undefined) {
            delete (globalThis as any).requestAnimationFrame;
        } else {
            (globalThis as any).requestAnimationFrame = originalRequestAnimationFrame;
        }

        console.warn = originalConsoleWarn;
    });

    test('throws when track settings are unavailable', () => {
        const track = {
            getSettings: vi.fn(() => undefined),
        } as unknown as MediaStreamTrack;

        expect(() => new VideoTrackProcessor(track)).toThrow('track has no settings');
        expect(track.getSettings).toHaveBeenCalled();
    });
});
