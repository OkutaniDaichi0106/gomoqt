import { AudioTrackDecoder } from "./internal";
import type {
    CancelCauseFunc,
    Context,
} from "golikejs/context";
import {
    withCancelCause,
    background
} from "golikejs/context";

import { DefaultVolume, DefaultMinGain, DefaultFadeTime } from './volume';
import { importUrl as importWorkletUrl } from './internal/audio_offload_worklet';

export interface AudioOffloaderInit {
    latency?: number; // Desired latency in milliseconds (default: 100)
    audioContext?: AudioContext; // Optional external AudioContext to use
    initialVolume?: number; // Optional initial volume (0-1); defaults to 1.0
    volumeRampMs?: number; // Gain ramp duration in ms for user adjustments (default 80ms similar feel to YouTube)
    sampleRate?: number; // Sample rate for audio context
    numberOfChannels?: number; // Number of channels (default: 2)
}

export class AudioOffloader {
    #decoder?: AudioTrackDecoder;
    audioContext: AudioContext;
    // worklet?: AudioWorkletNode;
    #muted: boolean = false;
    #unmuteVolume: number;
    #rampMs: number;

    #initWorklet: () => Promise<AudioWorkletNode>;

    #cancelFunc: CancelCauseFunc
    #ctx: Context;

    gainNode?: GainNode;

    constructor(init: AudioOffloaderInit) {
        [this.#ctx, this.#cancelFunc] = withCancelCause(background());

        this.audioContext = init.audioContext || new AudioContext({
            latencyHint: 'interactive',
            sampleRate: init.sampleRate,
        });

        this.#rampMs = init.volumeRampMs ?? DefaultFadeTime();
        const desiredInitialVolume = init.initialVolume ?? DefaultVolume();
        const clampedInitial = this.#clamp(desiredInitialVolume);
        this.#unmuteVolume = clampedInitial === 0 ? DefaultVolume() : clampedInitial;

        // Initialize worklet asynchronously (similar to VideoRenderer pattern)
        this.#initWorklet = async () => {
            try {
                const workletUrl = importWorkletUrl();
                console.debug("loading audio offload worklet from", workletUrl);
                await this.audioContext.audioWorklet.addModule(workletUrl);

                // Create AudioWorkletNode
                const worklet = new AudioWorkletNode(
                    this.audioContext,
                    'AudioOffloader',
                    {
                        channelCount: init.numberOfChannels || 2,
                        numberOfInputs: 0,
                        numberOfOutputs: 1,
                        processorOptions: {
                            sampleRate: init.sampleRate || this.audioContext.sampleRate,
                            latency: init.latency || 100, // Default to 100ms if not specified
                        },
                    }
                );

                const gainNode = new GainNode(this.audioContext, { gain: clampedInitial });
                worklet.connect(gainNode);
                gainNode.connect(this.audioContext.destination);

                // Set the nodes to fields
                this.gainNode = gainNode;

                // Ensure clean up on context done
                this.#ctx.done().then(async () => {
                    gainNode.disconnect();
                    worklet.disconnect();
                    if (!init.audioContext) { // Only close if we created the context
                        await this.audioContext.close();
                    }
                });

                return worklet;
            } catch (error) {
                console.error('failed to load AudioWorklet module:', error);
                throw error;
            }
        }
    }

    async decoder(): Promise<AudioTrackDecoder> {
        if (!this.#decoder) {
            const worklet = await this.#initWorklet();

            // Create new decoder (similar to VideoRenderer pattern)
            this.#decoder = new AudioTrackDecoder({
                destination: new WritableStream<AudioData>({
                    write: (frame) => {
                        this.#emitAudioData(frame, worklet);
                    }
                })
            });
        }

        return this.#decoder;
    }

    #emitAudioData(frame: AudioData, worklet: AudioWorkletNode): void {
        // No longer drops frames when muted; gain handles silence for continuity.
        const channels: Float32Array[] = [];
        for (let i = 0; i < frame.numberOfChannels; i++) {
            const data = new Float32Array(frame.numberOfFrames);
            frame.copyTo(data, { format: "f32-planar", planeIndex: i });
            channels.push(data);
        }

        // Send AudioData to the worklet
        worklet.port.postMessage(
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
