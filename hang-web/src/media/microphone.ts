import { Device } from "./device";
import type { DeviceProps } from "./device";

export interface MicrophoneProps {
    device?: DeviceProps;
    enabled?: boolean;
    constraints?: MediaTrackConstraints;
}

export class Microphone {
    device: Device;
    enabled: boolean;
    constraints: MediaTrackConstraints | undefined;
    #stream: MediaStreamTrack | undefined;

    constructor(props?: MicrophoneProps) {
        this.device = new Device("audio", props?.device);
        this.enabled = props?.enabled ?? false;
        this.constraints = props?.constraints;
    }

    /**
     * Return a promise that resolves to the current MediaStreamTrack for the microphone.
     * If the microphone is not started it will call start(). Caller should await
     * this instead of accessing `.stream` directly.
     */
    async getAudioTrack(): Promise<MediaStreamTrack> {
        if (!this.enabled) {
            throw new Error("Microphone is not enabled");
        }

        if (this.#stream) return this.#stream;

        const track = await this.device.getTrack(this.constraints);
        if (!track) {
            throw new Error("Failed to obtain microphone track");
        }

        this.#stream = track;
        return this.#stream;
    }

    async getSettings(): Promise<MediaTrackSettings> {
        const track = await this.getAudioTrack();
        return track.getSettings();
    }

    // async audioEncoder(config: AudioEncoderConfig, onDecoderConfig: (config: AudioDecoderConfig) => void): Promise<AudioEncodeStream> {
    //     const track = await this.getAudioTrack();
    //     const encoder = new AudioEncodeStream({
    //         source: new AudioTrackProcessor(track).readable,
    //         onDecoderConfig: onDecoderConfig,
    //     });

    //     encoder.configure(config);

    //     return encoder;
    // }


    close(): void {
        if (this.#stream) {
            try {
                this.#stream.stop();
            } catch (error) {
                // Ignore errors when stopping track
            }
            this.#stream = undefined;
        }
        try {
            this.device.close();
        } catch (error) {
            // Ignore errors when closing device
        }
    }
}