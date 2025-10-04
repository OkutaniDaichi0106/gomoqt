import type { DeviceProps } from "./device";
import { Device } from "./device";
import { VideoTrackEncoder,VideoTrackProcessor } from "../internal";

export interface CameraProps {
    device?: DeviceProps;
    enabled?: boolean;
    constraints?: MediaTrackConstraints;
}

export class Camera {
    device: Device;
    enabled: boolean;
    constraints: MediaTrackConstraints | undefined;
    #stream: MediaStreamTrack | undefined;

    constructor(props?: CameraProps) {
        this.device = new Device("video", props?.device);
        this.enabled = props?.enabled ?? false;
        this.constraints = props?.constraints;
    }

    /**
     * Return a promise that resolves to the current MediaStreamTrack for the camera.
     * If the camera is not started it will call start().
     * On failure this rejects with an Error instead of returning undefined.
     */
    async getVideoTrack(): Promise<MediaStreamTrack> {
        if (!this.enabled) {
            throw new Error("Camera is not enabled");
        }

        if (this.#stream) return this.#stream;

        const track = await this.device.getTrack(this.constraints);
        if (!track) {
            throw new Error("Failed to obtain camera track");
        }

        this.#stream = track;

        return this.#stream;
    }

    // async videoEncoder(config: VideoEncoderConfig, onDecoderConfig: (config: VideoDecoderConfig) => void): Promise<VideoEncodeStream> {
    //     const track = await this.getVideoTrack();
    //     const encoder = new VideoEncodeStream({
    //         source: new VideoTrackProcessor(track).readable,
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