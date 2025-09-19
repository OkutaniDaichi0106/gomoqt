export class AudioTrackProcessor {
    readonly readable: ReadableStream<AudioData>;
    readonly gain: GainNode;

    constructor(track: MediaStreamTrack) {
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
                await context.audioWorklet.addModule(
                    new URL("./audio_hijacker.js", import.meta.url)
                );
                const worklet = new AudioWorkletNode(context, "audio-hijacker", {
                    numberOfInputs: 1,
                    numberOfOutputs: 0,
                    channelCount: settings.channelCount,
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