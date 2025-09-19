import type { AudioDecodeStream } from "./internal";
import type {
    CancelCauseFunc,
    Context,
} from "@okutanidaichi/moqt/internal";
import {
    withCancelCause,
    background
} from "@okutanidaichi/moqt/internal";

import { DefaultVolume, DefaultMinGain, DefaultFadeTime } from './volume';

export interface AudioOffloaderInit {
    source: AudioDecodeStream;
    decoderConfig: AudioDecoderConfig;
    latency?: number; // Desired latency in milliseconds (default: 100)
    initialVolume?: number; // Optional initial volume (0-1); defaults to 1.0
    volumeRampMs?: number; // Gain ramp duration in ms for user adjustments (default 80ms similar feel to YouTube)
}

export class AudioOffloader {
    source: AudioDecodeStream;
    audioContext: AudioContext;
    worklet?: AudioWorkletNode;
    #muted: boolean = false;
    #unmuteVolume: number;
    #rampMs: number;

    #cancelFunc: CancelCauseFunc
    #ctx: Context;

    gainNode?: GainNode;

    constructor(init: AudioOffloaderInit) {
        this.source = init.source;
        [this.#ctx, this.#cancelFunc] = withCancelCause(background());
        const sampleRate = init.decoderConfig.sampleRate;
        this.audioContext = new AudioContext({
            latencyHint: 'interactive',
            sampleRate: sampleRate,
        });

        this.#rampMs = init.volumeRampMs ?? DefaultFadeTime();
        const desiredInitialVolume = init.initialVolume ?? DefaultVolume();
        const clampedInitial = this.#clamp(desiredInitialVolume);
        this.#unmuteVolume = clampedInitial === 0 ? DefaultVolume() : clampedInitial;

        this.audioContext.audioWorklet.addModule(
            new URL('./internal/audio_renderer.ts', import.meta.url)
        ).then(() => {
            // Create AudioWorkletNode
            const worklet = new AudioWorkletNode(
                this.audioContext,
                'audio-offloader',
                {
                    channelCount: init.decoderConfig.numberOfChannels || 2,
                    numberOfInputs: 0,
                    numberOfOutputs: 1,
                    processorOptions: {
                        sampleRate: sampleRate,
                        latency: init.latency || 100, // Default to 100ms if not specified
                    },
                }
            );

            const gainNode = new GainNode(this.audioContext, { gain: clampedInitial });
            worklet.connect(gainNode);
            gainNode.connect(this.audioContext.destination);

            // Set the nodes to fields
            this.worklet = worklet;
            this.gainNode = gainNode;

            // Ensure clean up on context done
            this.#ctx.done().then(async () => {
                await Promise.all([
                    gainNode.disconnect(),
                    worklet.disconnect(),
                ]);
                this.audioContext.close();
            });
        }).catch((error) => {
            console.error('failed to load AudioWorklet module:', error);
        });

        // Set up the decode callback to send audio data to worklet
        this.source.decodeTo({
            output: ({done, value: frame}) => {
                if (done || !frame) {
                    return;
                }

                this.#emitAudioData(frame);
            },
            error: (error) => {
                // TODO: Handle decoding errors
                console.error('audio decoding error:', error);
            }
        });
    }

    #emitAudioData(frame: AudioData): void {
        if (!this.worklet) {
            frame.close();
            return;
        }

        // No longer drops frames when muted; gain handles silence for continuity.
        const channels: Float32Array[] = [];
        for (let i = 0; i < frame.numberOfChannels; i++) {
            const data = new Float32Array(frame.numberOfFrames);
            frame.copyTo(data, { format: "f32-planar", planeIndex: i });
            channels.push(data);
        }

        // Send AudioData to the worklet
        this.worklet.port.postMessage(
            {
                channels: channels,
                timestamp: frame.timestamp,
            },
            channels.map(d => d.buffer) // Transfer ownership of the buffers
        );

        // Close the frame after sending
        frame.close();
    }

    #clamp(v: number): number {
        return Math.min(1, Math.max(0, isFinite(v) ? v : 1));
    }

    setVolume(v: number) {
        const clamped = this.#clamp(v);
        if (this.gainNode) {
            const now = this.audioContext.currentTime;
            const gainParam = this.gainNode.gain;
            // Cancel scheduled to avoid stacking
            gainParam.cancelScheduledValues(now);
            gainParam.setValueAtTime(gainParam.value, now);
            if (clamped < DefaultMinGain()) {
                gainParam.exponentialRampToValueAtTime(DefaultMinGain(), now + DefaultFadeTime());
                gainParam.setValueAtTime(0, now + DefaultFadeTime() + 0.01);
            } else {
                gainParam.exponentialRampToValueAtTime(clamped, now + DefaultFadeTime());
            }
        }
        if (clamped > 0) {
            this.#unmuteVolume = clamped;
        }
        if (clamped === 0) {
            // Consider this effectively muted but we don't set #muted automatically
        }
    }

    mute(m: boolean) {
        if (m === this.#muted) return;
        this.#muted = m;
        if (!this.gainNode) return;
        const now = this.audioContext.currentTime;
        const gainParam = this.gainNode.gain;
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
        return this.gainNode?.gain.value ?? 1.0;
    }

    destroy() {
        this.#cancelFunc(new Error("AudioEmitter destroyed"));
    }
}
