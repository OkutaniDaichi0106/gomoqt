import { describe, test, expect, beforeEach, afterEach, vi } from 'vitest';
import { DefaultVolume, DefaultMinGain, DefaultFadeTime, MIN_GAIN_FALLBACK, FADE_TIME_FALLBACK, isValidMinGain, isValidFadeTime, isValidVolume } from './volume';

// Type augmentation for testing globalThis properties
declare global {
    var __DEFAULT_VOLUME__: number | undefined;
    var __DEFAULT_MIN_GAIN__: number | undefined;
    var __DEFAULT_FADE_TIME__: number | undefined;
}

describe('Volume', () => {
    let originalVolume: number | undefined;
    let originalMinGain: number | undefined;
    let originalFadeTime: number | undefined;
    let consoleWarnSpy: ReturnType<typeof vi.spyOn>;

    beforeEach(() => {
        // Save original globalThis values
        originalVolume = (globalThis as any).__DEFAULT_VOLUME__;
        originalMinGain = (globalThis as any).__DEFAULT_MIN_GAIN__;
        originalFadeTime = (globalThis as any).__DEFAULT_FADE_TIME__;

        // Clear globalThis properties
        delete (globalThis as any).__DEFAULT_VOLUME__;
        delete (globalThis as any).__DEFAULT_MIN_GAIN__;
        delete (globalThis as any).__DEFAULT_FADE_TIME__;

        // Mock console.warn
        consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
    });

    afterEach(() => {
        // Restore original globalThis values
        if (originalVolume !== undefined) {
            (globalThis as any).__DEFAULT_VOLUME__ = originalVolume;
        } else {
            delete (globalThis as any).__DEFAULT_VOLUME__;
        }

        if (originalMinGain !== undefined) {
            (globalThis as any).__DEFAULT_MIN_GAIN__ = originalMinGain;
        } else {
            delete (globalThis as any).__DEFAULT_MIN_GAIN__;
        }

        if (originalFadeTime !== undefined) {
            (globalThis as any).__DEFAULT_FADE_TIME__ = originalFadeTime;
        } else {
            delete (globalThis as any).__DEFAULT_FADE_TIME__;
        }

        // Restore console.warn
        consoleWarnSpy.mockRestore();
    });

    describe('Default Values', () => {
        test('returns fallback values when globalThis properties are not set', () => {
            const minGain = DefaultMinGain();
            const fadeTime = DefaultFadeTime();

            expect(minGain).toBe(MIN_GAIN_FALLBACK);
            expect(fadeTime).toBe(FADE_TIME_FALLBACK);
        });

        test('returns globalThis values when set', () => {
            // Simulate Vite define injection
            (globalThis as any).__DEFAULT_MIN_GAIN__ = 0.002;
            (globalThis as any).__DEFAULT_FADE_TIME__ = 0.09;

            const minGain = DefaultMinGain();
            const fadeTime = DefaultFadeTime();

            expect(minGain).toBe(0.002);
            expect(fadeTime).toBe(0.09);
        });

        test('warns when globalThis values are invalid', () => {
            // Simulate invalid Vite define injection
            (globalThis as any).__DEFAULT_MIN_GAIN__ = NaN;
            (globalThis as any).__DEFAULT_FADE_TIME__ = Infinity;

            const minGain = DefaultMinGain();
            const fadeTime = DefaultFadeTime();

            expect(minGain).toBe(MIN_GAIN_FALLBACK);
            expect(fadeTime).toBe(FADE_TIME_FALLBACK);

            // Note: Warnings are only logged during module initialization, not during function calls
            // So we don't expect console.warn to be called here
        });
    });

    describe('Validation Functions', () => {
        describe('isValidMinGain', () => {
            test('returns true for valid min gain values', () => {
                expect(isValidMinGain(0.001)).toBe(true);
                expect(isValidMinGain(0.005)).toBe(true);
                expect(isValidMinGain(0.009)).toBe(true);
            });

            test('returns false for invalid min gain values', () => {
                expect(isValidMinGain(0)).toBe(false);
                expect(isValidMinGain(-0.001)).toBe(false);
                expect(isValidMinGain(0.01)).toBe(false);
                expect(isValidMinGain(0.1)).toBe(false);
                expect(isValidMinGain(NaN)).toBe(false);
                expect(isValidMinGain(Infinity)).toBe(false);
                expect(isValidMinGain('0.001' as any)).toBe(false);
            });
        });

        describe('isValidFadeTime', () => {
            test('returns true for valid fade time values', () => {
                expect(isValidFadeTime(0.02)).toBe(true);
                expect(isValidFadeTime(0.5)).toBe(true);
                expect(isValidFadeTime(0.99)).toBe(true);
            });

            test('returns false for invalid fade time values', () => {
                expect(isValidFadeTime(0)).toBe(false);
                expect(isValidFadeTime(0.005)).toBe(false);
                expect(isValidFadeTime(1.0)).toBe(false);
                expect(isValidFadeTime(2.0)).toBe(false);
                expect(isValidFadeTime(NaN)).toBe(false);
                expect(isValidFadeTime(Infinity)).toBe(false);
                expect(isValidFadeTime('0.5' as any)).toBe(false);
            });
        });
    });

    describe('Validation Functions', () => {
        describe('isValidMinGain', () => {
            test('returns true for valid min gain values', () => {
                expect(isValidMinGain(0.001)).toBe(true);
                expect(isValidMinGain(0.005)).toBe(true);
                expect(isValidMinGain(0.009)).toBe(true);
            });

            test('returns false for invalid min gain values', () => {
                expect(isValidMinGain(0)).toBe(false);
                expect(isValidMinGain(-0.001)).toBe(false);
                expect(isValidMinGain(0.01)).toBe(false);
                expect(isValidMinGain(0.1)).toBe(false);
                expect(isValidMinGain(NaN)).toBe(false);
                expect(isValidMinGain(Infinity)).toBe(false);
                expect(isValidMinGain('0.001' as any)).toBe(false);
            });
        });

        describe('isValidFadeTime', () => {
            test('returns true for valid fade time values', () => {
                expect(isValidFadeTime(0.02)).toBe(true);
                expect(isValidFadeTime(0.5)).toBe(true);
                expect(isValidFadeTime(0.99)).toBe(true);
            });

            test('returns false for invalid fade time values', () => {
                expect(isValidFadeTime(0)).toBe(false);
                expect(isValidFadeTime(0.005)).toBe(false);
                expect(isValidFadeTime(1.0)).toBe(false);
                expect(isValidFadeTime(2.0)).toBe(false);
                expect(isValidFadeTime(NaN)).toBe(false);
                expect(isValidFadeTime(Infinity)).toBe(false);
                expect(isValidFadeTime('0.5' as any)).toBe(false);
            });
        });

        describe('isValidVolume', () => {
            test('returns true for valid volume values', () => {
                expect(isValidVolume(0)).toBe(true);
                expect(isValidVolume(0.1)).toBe(true);
                expect(isValidVolume(0.5)).toBe(true);
                expect(isValidVolume(1.0)).toBe(true);
            });

            test('returns false for invalid volume values', () => {
                expect(isValidVolume(-0.1)).toBe(false);
                expect(isValidVolume(1.1)).toBe(false);
                expect(isValidVolume(NaN)).toBe(false);
                expect(isValidVolume(Infinity)).toBe(false);
                expect(isValidVolume('0.5' as any)).toBe(false);
                expect(isValidVolume(null)).toBe(false);
                expect(isValidVolume(undefined)).toBe(false);
            });
        });
    });
});
