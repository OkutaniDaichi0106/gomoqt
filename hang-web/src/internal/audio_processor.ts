import { importWorkletUrl } from './audio_hijack_worklet';

export interface AudioTrackProcessorOptions {
    targetChannels?: number;
}

export class AudioTrackProcessor {
    readonly readable: ReadableStream<AudioData>;
    readonly gain: GainNode;

    constructor(track: MediaStreamTrack, options: AudioTrackProcessorOptions = {}) {
        // @ts-expect-error No typescript types yet.
        if (self.MediaStreamTrackProcessor) {
            // @ts-expect-error No typescript types yet.
            this.readable = new self.MediaStreamTrackProcessor({ track }).readable;
        }

        console.warn("Using MediaStreamTrackProcessor polyfill; performance might suffer.");

        const settings = track.getSettings();
        if (!settings) {
            throw new Error("track has no settings");
        }

        const context = new AudioContext({
                latencyHint: "interactive",
                sampleRate: settings.sampleRate,
        });

        const source = new MediaStreamAudioSourceNode(context,
            { mediaStream: new MediaStream([track]) });

        const gain = new GainNode(context, { gain: 1.0 });
        this.gain = gain;
        source.connect(gain);

        this.readable = new ReadableStream<AudioData>({
            async start(controller) {
                const workletUrl = importWorkletUrl();
                console.debug("loading audio hijack worklet from", workletUrl);
                await context.audioWorklet.addModule(workletUrl);

                const worklet = new AudioWorkletNode(context, "AudioHijacker", {
                    numberOfInputs: 1,
                    numberOfOutputs: 0,
                    channelCount: settings.channelCount,
                    processorOptions: {
                        sampleRate: context.sampleRate,
                        targetChannels: options.targetChannels || settings.channelCount || 1,
                    },
                });

                gain.connect(worklet);

                worklet.port.onmessage = ({data}: {data: AudioDataInit}) => {
                    const frame = new AudioData(data);
                    controller.enqueue(frame);
                };
            },
            cancel() {
                context.close();
                gain.disconnect();
                source.disconnect();
            },
        });
    }
}