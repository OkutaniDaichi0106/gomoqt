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
export function isValidMinGain(v: number): boolean {
  return typeof v === 'number' && Number.isFinite(v) && v > 0 && v < 0.01; // Reasonable range for min gain
}

export function isValidFadeTime(v: number): boolean {
  return typeof v === 'number' && Number.isFinite(v) && v > 0.01 && v < 1.0; // Reasonable range for fade time
}

export function isValidVolume(v: unknown): v is number {
    return typeof v === 'number' && Number.isFinite(v) && v >= 0 && v <= 1;
}

export function DefaultVolume(): number {
    const injected = (globalThis as unknown as GlobalWithDefaultVolume).__DEFAULT_VOLUME__;
    if (injected !== undefined) {
        if (isValidVolume(injected)) {
            return injected;
        } else {
            console.warn('[volume] __DEFAULT_VOLUME__ is out of range, fallback to 0.5:', injected);
        }
    }
    return FALLBACK_VOLUME;
}


/**
 * Enhanced GainNode with volume control utilities
 */
export class VolumeController extends GainNode {
    #muted: boolean = false;
    #unmuteVolume: number;
    #rampMs: number;

    constructor(audioContext: AudioContext, options?: GainOptions & { initialVolume?: number; fadeTimeMs?: number }) {
        const initialVolume = options?.initialVolume ?? DefaultVolume();
        const clampedInitial = Math.min(1, Math.max(0, isFinite(initialVolume) ? initialVolume : 1));

        super(audioContext, { ...options, gain: clampedInitial });

        this.#rampMs = options?.fadeTimeMs ?? DefaultFadeTime();
        this.#unmuteVolume = clampedInitial === 0 ? DefaultVolume() : clampedInitial;
    }

    #clamp(v: number): number {
        return Math.min(1, Math.max(0, isFinite(v) ? v : 1));
    }

    setVolume(v: number) {
        const clamped = this.#clamp(v);
        const now = this.context.currentTime;
        const gainParam = this.gain;

        // Cancel scheduled to avoid stacking
        gainParam.cancelScheduledValues(now);
        gainParam.setValueAtTime(gainParam.value, now);

        if (clamped < DefaultMinGain()) {
            gainParam.exponentialRampToValueAtTime(DefaultMinGain(), now + DefaultFadeTime());
            gainParam.setValueAtTime(0, now + DefaultFadeTime() + 0.01);
        } else {
            gainParam.exponentialRampToValueAtTime(clamped, now + DefaultFadeTime());
        }

        if (clamped > 0) {
            this.#unmuteVolume = clamped;
        }
    }

    mute(m: boolean) {
        if (m === this.#muted) return;
        this.#muted = m;

        const now = this.context.currentTime;
        const gainParam = this.gain;
        gainParam.cancelScheduledValues(now);
        gainParam.setValueAtTime(gainParam.value, now);

        if (m) {
            // Store previous volume if >0
            const current = gainParam.value;
            if (current > 0.0001) {
                this.#unmuteVolume = current;
            }
            if (current < DefaultMinGain()) {
                gainParam.exponentialRampToValueAtTime(DefaultMinGain(), now + DefaultFadeTime());
                gainParam.setValueAtTime(0, now + DefaultFadeTime() + 0.01);
            } else {
                gainParam.exponentialRampToValueAtTime(0, now + DefaultFadeTime());
            }
        } else {
            const restore = this.#unmuteVolume <= 0 ? DefaultVolume() : this.#unmuteVolume;
            gainParam.exponentialRampToValueAtTime(this.#clamp(restore), now + DefaultFadeTime());
        }
    }

    get muted(): boolean {
        return this.#muted;
    }

    get volume(): number {
        return this.gain.value;
    }
}