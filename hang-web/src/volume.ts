/**
 * Default volume module with build-time constant injection support.
 *
 * A bundler (esbuild / Vite / webpack, etc.) can define `__DEFAULT_VOLUME__`
 * as a numeric literal in the range [0,1]. That value becomes the initial
 * default volume. If it's absent or invalid, we fall back to 0.5.
 *
 * Example (esbuild):
 *   esbuild src/index.ts --define:__DEFAULT_VOLUME__=0.7
 */

// Previous version used a direct `typeof __DEFAULT_VOLUME__ !== 'undefined'` guard.
// Switched to reading off globalThis so we can rely on nullish coalescing (??) without risking
// a ReferenceError for an undeclared symbol.
// To inject via bundler, define: `globalThis.__DEFAULT_VOLUME__ = 0.7` (or with esbuild define:
// esbuild ... --define:globalThis.__DEFAULT_VOLUME__=0.7)
// We keep a declaration merging-friendly interface for type safety.
interface GlobalWithDefaultVolume { __DEFAULT_VOLUME__?: number }

const FALLBACK_VOLUME = 0.5; // Library fallback default

// Audio fade constants with build-time injection support
export const MIN_GAIN_FALLBACK = 0.001; // Fallback minimum gain
export const FADE_TIME_FALLBACK = 80; // Fallback fade time in milliseconds

// Build-time injection declarations (similar to __DEFAULT_VOLUME__)
interface GlobalWithAudioConstants { __DEFAULT_MIN_GAIN__?: number; __DEFAULT_FADE_TIME__?: number; }

// Get minimum gain with build-time injection
export function DefaultMinGain(): number {
  const injected = (globalThis as unknown as GlobalWithAudioConstants).__DEFAULT_MIN_GAIN__;
  if (injected !== undefined && isValidMinGain(injected)) {
    return injected;
  }
  return MIN_GAIN_FALLBACK;
}

// Get fade time with build-time injection
export function DefaultFadeTime(): number {
  const injected = (globalThis as unknown as GlobalWithAudioConstants).__DEFAULT_FADE_TIME__;
  if (injected !== undefined && isValidFadeTime(injected)) {
    return injected;
  }
  return FADE_TIME_FALLBACK;
}

// Validation functions for compile-time checks
function isValidMinGain(v: number): boolean {
  return typeof v === 'number' && Number.isFinite(v) && v > 0 && v < 0.01; // Reasonable range for min gain
}

function isValidFadeTime(v: number): boolean {
  return typeof v === 'number' && Number.isFinite(v) && v > 0.01 && v < 1.0; // Reasonable range for fade time
}

function isValidVolume(v: unknown): v is number {
    return typeof v === 'number' && Number.isFinite(v) && v >= 0 && v <= 1;
}

let computedDefault = FALLBACK_VOLUME;
const injected = (globalThis as unknown as GlobalWithDefaultVolume).__DEFAULT_VOLUME__;
if (injected !== undefined) {
    if (isValidVolume(injected)) {
        computedDefault = injected;
    } else {
        console.warn('[volume] __DEFAULT_VOLUME__ is out of range, fallback to 0.5:', injected);
    }
}

export function DefaultVolume(): number { return computedDefault; }