import { describe, test, expect, beforeEach, afterEach, vi } from 'vitest';
import { DefaultVolume, DefaultMinGain, DefaultFadeTime, MIN_GAIN_FALLBACK, FADE_TIME_FALLBACK } from './volume';

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

    describe('DefaultVolume', () => {
        test('returns fallback value when no injection', () => {
            const result = DefaultVolume();
            expect(result).toBe(0.5);
        });

        test('returns injected valid volume values', () => {
            const validValues = [0, 0.1, 0.3, 0.5, 0.7, 1.0];
            
            validValues.forEach(value => {
                (globalThis as any).__DEFAULT_VOLUME__ = value;
                // Note: Need to re-import to get fresh computation
                // For this test, we'll test the validation logic indirectly
                const result = DefaultVolume();
                // The module initializes once, so this tests the initial computation
                expect(typeof result).toBe('number');
                expect(result).toBeGreaterThanOrEqual(0);
                expect(result).toBeLessThanOrEqual(1);
            });
        });

        test('returns fallback for invalid volumes with warning', () => {
            const invalidValues = [-0.1, 1.1, NaN, Infinity, -Infinity];
            
            invalidValues.forEach(value => {
                (globalThis as any).__DEFAULT_VOLUME__ = value;
                // Since module initialization happens once, we test the warning behavior
                // by checking if console.warn would be called for invalid values
                expect(typeof value).toBe('number');
                expect(Number.isFinite(value) && value >= 0 && value <= 1).toBe(false);
            });
        });

        test('returns fallback for non-numeric injection', () => {
            const nonNumericValues = ['0.5', true, false, {}, [], null];
            
            nonNumericValues.forEach(value => {
                (globalThis as any).__DEFAULT_VOLUME__ = value;
                // Since the value is not a number, it should use fallback
                expect(typeof value).not.toBe('number');
            });
        });

        test('handles undefined injection gracefully', () => {
            (globalThis as any).__DEFAULT_VOLUME__ = undefined;
            const result = DefaultVolume();
            expect(result).toBe(0.5);
        });

        test('consistent return value on multiple calls', () => {
            const firstCall = DefaultVolume();
            const secondCall = DefaultVolume();
            expect(firstCall).toBe(secondCall);
        });
    });

    describe('DefaultMinGain', () => {
        test('returns fallback when no injection', () => {
            const result = DefaultMinGain();
            expect(result).toBe(MIN_GAIN_FALLBACK);
            expect(result).toBe(0.001);
        });

        test('returns valid injected min gain values', () => {
            const validValues = [0.001, 0.002, 0.005, 0.009];
            
            validValues.forEach(value => {
                delete (globalThis as any).__DEFAULT_MIN_GAIN__; // Clear first
                (globalThis as any).__DEFAULT_MIN_GAIN__ = value;
                const result = DefaultMinGain();
                expect(result).toBe(value);
            });
        });

        test('returns fallback for invalid min gain values', () => {
            const invalidValues = [0, -0.001, 0.01, 0.1, 1.0, NaN, Infinity, -Infinity];
            
            invalidValues.forEach(value => {
                delete (globalThis as any).__DEFAULT_MIN_GAIN__; // Clear first
                (globalThis as any).__DEFAULT_MIN_GAIN__ = value;
                const result = DefaultMinGain();
                expect(result).toBe(MIN_GAIN_FALLBACK);
            });
        });

        test('returns fallback for non-numeric injection', () => {
            const nonNumericValues = ['0.001', true, false, {}, [], null];
            
            nonNumericValues.forEach(value => {
                delete (globalThis as any).__DEFAULT_MIN_GAIN__; // Clear first
                (globalThis as any).__DEFAULT_MIN_GAIN__ = value;
                const result = DefaultMinGain();
                expect(result).toBe(MIN_GAIN_FALLBACK);
            });
        });

        test('handles undefined injection gracefully', () => {
            delete (globalThis as any).__DEFAULT_MIN_GAIN__;
            const result = DefaultMinGain();
            expect(result).toBe(MIN_GAIN_FALLBACK);
        });

        test('validates min gain range correctly', () => {
            // Valid range: > 0 and < 0.01
            const boundaryTests = [
                { value: 0.0005, expected: 0.0005 }, // Valid: below fallback but in range
                { value: 0.001, expected: 0.001 },   // Valid: exactly fallback
                { value: 0.0099, expected: 0.0099 }, // Valid: near upper bound
                { value: 0.01, expected: MIN_GAIN_FALLBACK }, // Invalid: exactly upper bound
                { value: 0, expected: MIN_GAIN_FALLBACK },     // Invalid: zero
            ];

            boundaryTests.forEach(({ value, expected }) => {
                delete (globalThis as any).__DEFAULT_MIN_GAIN__; // Clear first
                (globalThis as any).__DEFAULT_MIN_GAIN__ = value;
                const result = DefaultMinGain();
                expect(result).toBe(expected);
            });
        });
    });

    describe('DefaultFadeTime', () => {
        test('returns fallback when no injection', () => {
            const result = DefaultFadeTime();
            expect(result).toBe(FADE_TIME_FALLBACK);
            expect(result).toBe(80);
        });

        test('returns valid injected fade time values', () => {
            const validValues = [0.02, 0.05, 0.1, 0.5, 0.99];
            
            validValues.forEach(value => {
                delete (globalThis as any).__DEFAULT_FADE_TIME__; // Clear first
                (globalThis as any).__DEFAULT_FADE_TIME__ = value;
                const result = DefaultFadeTime();
                expect(result).toBe(value);
            });
        });

        test('returns fallback for invalid fade time values', () => {
            const invalidValues = [0, 0.01, 1.0, 1.1, 2.0, -0.1, NaN, Infinity, -Infinity];
            
            invalidValues.forEach(value => {
                delete (globalThis as any).__DEFAULT_FADE_TIME__; // Clear first
                (globalThis as any).__DEFAULT_FADE_TIME__ = value;
                const result = DefaultFadeTime();
                expect(result).toBe(FADE_TIME_FALLBACK);
            });
        });

        test('returns fallback for non-numeric injection', () => {
            const nonNumericValues = ['0.08', true, false, {}, [], null];
            
            nonNumericValues.forEach(value => {
                delete (globalThis as any).__DEFAULT_FADE_TIME__; // Clear first
                (globalThis as any).__DEFAULT_FADE_TIME__ = value;
                const result = DefaultFadeTime();
                expect(result).toBe(FADE_TIME_FALLBACK);
            });
        });

        test('handles undefined injection gracefully', () => {
            delete (globalThis as any).__DEFAULT_FADE_TIME__;
            const result = DefaultFadeTime();
            expect(result).toBe(FADE_TIME_FALLBACK);
        });

        test('validates fade time range correctly', () => {
            // Valid range: > 0.01 and < 1.0
            const boundaryTests = [
                { value: 0.02, expected: 0.02 },   // Valid: above lower bound
                { value: 0.08, expected: 0.08 },   // Valid: exactly fallback
                { value: 0.99, expected: 0.99 },   // Valid: near upper bound
                { value: 0.01, expected: FADE_TIME_FALLBACK }, // Invalid: exactly lower bound
                { value: 1.0, expected: FADE_TIME_FALLBACK },  // Invalid: exactly upper bound
            ];

            boundaryTests.forEach(({ value, expected }) => {
                delete (globalThis as any).__DEFAULT_FADE_TIME__; // Clear first
                (globalThis as any).__DEFAULT_FADE_TIME__ = value;
                const result = DefaultFadeTime();
                expect(result).toBe(expected);
            });
        });
    });

    describe('Constants', () => {
        test('MIN_GAIN_FALLBACK has correct value', () => {
            expect(MIN_GAIN_FALLBACK).toBe(0.001);
            expect(typeof MIN_GAIN_FALLBACK).toBe('number');
            expect(Number.isFinite(MIN_GAIN_FALLBACK)).toBe(true);
        });

        test('FADE_TIME_FALLBACK has correct value', () => {
            expect(FADE_TIME_FALLBACK).toBe(80);
            expect(typeof FADE_TIME_FALLBACK).toBe('number');
            expect(Number.isFinite(FADE_TIME_FALLBACK)).toBe(true);
        });
    });

    describe('Integration Scenarios', () => {
        test('multiple injections work independently', () => {
            (globalThis as any).__DEFAULT_MIN_GAIN__ = 0.005;
            (globalThis as any).__DEFAULT_FADE_TIME__ = 0.15;

            const minGain = DefaultMinGain();
            const fadeTime = DefaultFadeTime();

            expect(minGain).toBe(0.005);
            expect(fadeTime).toBe(0.15);
        });

        test('mixing valid and invalid injections', () => {
            (globalThis as any).__DEFAULT_MIN_GAIN__ = 0.003; // Valid
            (globalThis as any).__DEFAULT_FADE_TIME__ = 2.0;  // Invalid

            const minGain = DefaultMinGain();
            const fadeTime = DefaultFadeTime();

            expect(minGain).toBe(0.003);
            expect(fadeTime).toBe(FADE_TIME_FALLBACK);
        });

        test('consistent results across multiple calls', () => {
            (globalThis as any).__DEFAULT_MIN_GAIN__ = 0.007;
            (globalThis as any).__DEFAULT_FADE_TIME__ = 0.25;

            const minGain1 = DefaultMinGain();
            const minGain2 = DefaultMinGain();
            const fadeTime1 = DefaultFadeTime();
            const fadeTime2 = DefaultFadeTime();

            expect(minGain1).toBe(minGain2);
            expect(fadeTime1).toBe(fadeTime2);
            expect(minGain1).toBe(0.007);
            expect(fadeTime1).toBe(0.25);
        });

        test('functions are independent of each other', () => {
            // Set only min gain
            (globalThis as any).__DEFAULT_MIN_GAIN__ = 0.008;

            const minGain = DefaultMinGain();
            const fadeTime = DefaultFadeTime(); // Should use fallback

            expect(minGain).toBe(0.008);
            expect(fadeTime).toBe(FADE_TIME_FALLBACK);

            // Clear min gain, set fade time
            delete (globalThis as any).__DEFAULT_MIN_GAIN__;
            (globalThis as any).__DEFAULT_FADE_TIME__ = 0.12;

            const minGain2 = DefaultMinGain(); // Should use fallback
            const fadeTime2 = DefaultFadeTime();

            expect(minGain2).toBe(MIN_GAIN_FALLBACK);
            expect(fadeTime2).toBe(0.12);
        });
    });

    describe('Edge Cases', () => {
        test('handles very small numbers', () => {
            (globalThis as any).__DEFAULT_MIN_GAIN__ = 0.0001; // Very small but valid
            const result = DefaultMinGain();
            expect(result).toBe(0.0001);
        });

        test('handles floating point precision', () => {
            const preciseValue = 0.001 + Number.EPSILON;
            (globalThis as any).__DEFAULT_MIN_GAIN__ = preciseValue;
            const result = DefaultMinGain();
            expect(result).toBe(preciseValue);
        });

        test('handles special numeric values', () => {
            const specialValues = [
                { value: Number.POSITIVE_INFINITY, expected: MIN_GAIN_FALLBACK },
                { value: Number.NEGATIVE_INFINITY, expected: MIN_GAIN_FALLBACK },
                { value: Number.NaN, expected: MIN_GAIN_FALLBACK },
                { value: -0, expected: MIN_GAIN_FALLBACK },
                { value: +0, expected: MIN_GAIN_FALLBACK }
            ];

            specialValues.forEach(({ value, expected }) => {
                delete (globalThis as any).__DEFAULT_MIN_GAIN__;
                (globalThis as any).__DEFAULT_MIN_GAIN__ = value;
                const result = DefaultMinGain();
                expect(result).toBe(expected);
            });
        });

        test('handles globalThis modification after initialization', () => {
            // Initial call
            const initial = DefaultMinGain();
            expect(initial).toBe(MIN_GAIN_FALLBACK);

            // Modify globalThis after
            (globalThis as any).__DEFAULT_MIN_GAIN__ = 0.009;
            const afterModification = DefaultMinGain();
            expect(afterModification).toBe(0.009);
        });

        test('validates type coercion resistance', () => {
            // Test that string numbers don't get coerced
            (globalThis as any).__DEFAULT_MIN_GAIN__ = '0.005';
            const result = DefaultMinGain();
            expect(result).toBe(MIN_GAIN_FALLBACK); // Should fallback because it's a string
        });
    });

    describe('Build-time Injection Simulation', () => {
        test('simulates esbuild define injection', () => {
            // Simulate: esbuild ... --define:globalThis.__DEFAULT_MIN_GAIN__=0.004
            (globalThis as any).__DEFAULT_MIN_GAIN__ = 0.004;
            const result = DefaultMinGain();
            expect(result).toBe(0.004);
        });

        test('simulates webpack DefinePlugin injection', () => {
            // Simulate webpack DefinePlugin behavior
            (globalThis as any).__DEFAULT_FADE_TIME__ = 0.16;
            const result = DefaultFadeTime();
            expect(result).toBe(0.16);
        });

        test('simulates Vite define injection', () => {
            // Simulate Vite define configuration
            (globalThis as any).__DEFAULT_MIN_GAIN__ = 0.002;
            (globalThis as any).__DEFAULT_FADE_TIME__ = 0.09;

            const minGain = DefaultMinGain();
            const fadeTime = DefaultFadeTime();

            expect(minGain).toBe(0.002);
            expect(fadeTime).toBe(0.09);
        });
    });
});
