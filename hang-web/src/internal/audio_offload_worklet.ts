// URL getter function for main thread import
export function importWorkletUrl(): string {
	return new URL('./audio_offload_worklet.js', import.meta.url).href;
}

export const workletName: string = 'audio-offloader';

// Check if we're in a worklet context
if (typeof AudioWorkletProcessor !== 'undefined') {
	// AudioWorkletProcessor for AudioEmitter
	class AudioOffloadProcessor extends AudioWorkletProcessor {
        #channelsBuffer: Float32Array[] = [];

        #readIndex: number = 0;
        #writeIndex: number = 0;

        constructor(options: AudioWorkletNodeOptions) {
            super();
            if (!options.processorOptions) {
                throw new Error("processorOptions is required");
            }

            const channelCount = options.channelCount;
            if (!channelCount || channelCount <= 0) {
                throw new Error("invalid channelCount");
            }

            const sampleRate = options.processorOptions.sampleRate;
            if (!sampleRate || sampleRate <= 0) {
                throw new Error("invalid sampleRate");
            }

            const latency = options.processorOptions.latency;
            if (!latency || latency <= 0) {
                throw new Error("invalid latency");
            }

            const bufferingSamples = Math.ceil(sampleRate * latency / 1000);

            for (let i = 0; i < channelCount; i++) {
                this.#channelsBuffer[i] = new Float32Array(bufferingSamples);
            }

            this.port.onmessage = ({data}: {data: {channels: Float32Array[], timestamp: number}}) => {
                this.append(data.channels);
                // We do not use timestamp for now
                // TODO: handle timestamp and sync if needed
            };
        }

        append(channels: Float32Array[]): void {
            if (!channels.length || !channels[0] || channels[0].length === 0) {
                return;
            }

            // Not initialized yet. Skip
            if (
                this.#channelsBuffer === undefined
                || this.#channelsBuffer.length === 0
                || this.#channelsBuffer[0] === undefined
            ) return;

            const numberOfFrames = channels[0].length;

            // Advance read index for discarded samples
            const discard = this.#writeIndex - this.#readIndex + numberOfFrames - this.#channelsBuffer[0].length;
            if (discard >= 0) {
                this.#readIndex += discard;
            }

            // Write new samples to buffer
            for (let channel = 0; channel < this.#channelsBuffer.length; channel++) {
                const src = channels[channel];
                const dst = this.#channelsBuffer[channel];

                if (!dst) continue;
                if (!src) {
                    dst.fill(0, 0, numberOfFrames);
                    continue;
                }

                let readPos = this.#writeIndex % dst.length;
                let offset = 0;

                let n: number;
                while (numberOfFrames - offset > 0) { // Still data remaining to copy
                    n = Math.min(numberOfFrames - offset, numberOfFrames - readPos);
                    dst.set(src.subarray(readPos, readPos + n), offset);
                    readPos = (readPos + n) % numberOfFrames;
                    offset += n;
                }
            }

            this.#writeIndex += numberOfFrames;
        }

        process(inputs: Float32Array[][], outputs: Float32Array[][]): boolean {
            // No output to write to
            if (
                outputs === undefined
                || outputs.length === 0
                || outputs[0] === undefined
                || outputs[0]?.length === 0
            ) return true;

            // Not initialized yet
            if (this.#channelsBuffer.length === 0 || this.#channelsBuffer[0] === undefined) return true;

            const available = (this.#writeIndex - this.#readIndex + this.#channelsBuffer[0].length) % this.#channelsBuffer[0].length;
            const numberOfFrames = Math.min(available, outputs[0].length);

            // No data to read
            if (numberOfFrames <= 0) return true;

            for (const output of outputs) {
                for (let channel = 0; channel < output.length; channel++) {
                    const src = this.#channelsBuffer[channel];
                    const dst = output[channel];
                    if (!dst) continue;
                    if (!src) {
                        dst.fill(0, 0, numberOfFrames);
                        continue;
                    };

                    let readPos = this.#readIndex;
                    let offset = 0;

                    let n: number;
                    while (numberOfFrames - offset > 0) { // Still data remaining to copy
                        n = Math.min(numberOfFrames - offset, numberOfFrames - readPos);
                        dst.set(src.subarray(readPos, readPos + n), offset);
                        readPos = (readPos + n) % numberOfFrames;
                        offset += n;
                    }
                }
            }

            // Advance read index
            this.#readIndex += numberOfFrames;
            if (this.#readIndex >= this.#channelsBuffer[0].length) {
                this.#readIndex -= this.#channelsBuffer[0].length;
                this.#writeIndex -= this.#channelsBuffer[0].length;
            }

            return true;
        }
    }

	registerProcessor(workletName, AudioOffloadProcessor);
}
