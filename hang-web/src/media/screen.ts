import {
    VideoTrackEncoder,
    VideoTrackProcessor,
    AudioTrackEncoder,
    AudioTrackProcessor,
} from "../internal";

export interface ScreenProps {
    enabled?: boolean;
    constraints?: DisplayMediaStreamOptions;
}

export class Screen {
    enabled: boolean;
    constraints?: DisplayMediaStreamOptions;
    #stream?: {
        video: MediaStreamTrack;
        audio?: MediaStreamTrack;
    }

    constructor(props?: ScreenProps) {
        this.enabled = props?.enabled ?? false;
        this.constraints = props?.constraints;
    }

    async #getTracks(): Promise<{ video: MediaStreamTrack; audio?: MediaStreamTrack }> {
        const stream = await navigator.mediaDevices.getDisplayMedia(this.constraints);
        const video = stream.getVideoTracks()[0];
        if (!video) {
            stream.getTracks().forEach(t => t.stop());
            throw new Error("Failed to obtain display video track");
        }
        const audio = stream.getAudioTracks()[0];

        // Stop any other tracks that may have been included in the stream
        for (const t of stream.getTracks()) {
            if (t !== video && t !== audio) {
                try { t.stop(); } catch {}
            }
        }

        return { video, audio };
    }

    async getVideoTrack(): Promise<MediaStreamTrack> {
        if (!this.enabled) {
            throw new Error("Screen capture is not enabled");
        }

        if (this.#stream) {
            return this.#stream.video
        };

        this.#stream = await this.#getTracks();
        if (!this.#stream.video) {
            throw new Error("No video track available for screen capture");
        }

        return this.#stream.video;
    }

    async getAudioTrack(): Promise<MediaStreamTrack | undefined> {
        if (!this.enabled) {
            throw new Error("Screen capture is not enabled");
        }

        if (this.#stream) {
            return this.#stream.audio
        };

        this.#stream = await this.#getTracks();
        return this.#stream.audio; // May be undefined, which is valid
    }

    // async videoEncoder(config: VideoEncoderConfig, onDecoderConfig: (config: VideoDecoderConfig) => void): Promise<VideoEncodeStream> {
    //     if (!this.enabled) {
    //         throw new Error("Screen capture is not enabled");
    //     }

    //     const video = await this.getVideoTrack();

    //     const encoder = new VideoEncodeStream({
    //         source: new VideoTrackProcessor(video).readable,
    //         onDecoderConfig: onDecoderConfig,
    //     });

    //     encoder.configure(config);

    //     return encoder;
    // }

    // async audioEncoder(config: AudioEncoderConfig, onDecoderConfig: (config: AudioDecoderConfig) => void): Promise<AudioEncodeStream> {
    //     if (!this.enabled) {
    //         throw new Error("Screen capture is not enabled");
    //     }

    //     const audio = await this.getAudioTrack();
    //     if (!audio) {
    //         throw new Error("No audio track available for screen capture");
    //     }

    //     const encoder = new AudioEncodeStream({
    //         source: new AudioTrackProcessor(audio).readable,
    //         onDecoderConfig: onDecoderConfig,
    //     });

    //     encoder.configure(config);

    //     return encoder;
    // }

    async close(): Promise<void> {
        if (!this.#stream) return;
        const tracks = this.#stream;
        this.#stream = undefined;
        try {
            tracks.video?.stop();
        } catch (error) {
            // Ignore errors when stopping video track
        }
        try {
            tracks.audio?.stop();
        } catch (error) {
            // Ignore errors when stopping audio track
        }
    }
}
