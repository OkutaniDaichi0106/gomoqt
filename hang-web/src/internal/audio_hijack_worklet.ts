// URL getter function for main thread import
export function importWorkletUrl(): string {
	return new URL('./audio_hijack_worklet.js', import.meta.url).href;
}

// Check if we're in a worklet context
if (typeof AudioWorkletProcessor !== 'undefined') {
	// Worklet code
	class AudioHijackProcessor extends AudioWorkletProcessor {
	#currentFrame: number = 0;
	#sampleRate: number;
	#targetChannels: number;
	
	constructor(options: AudioWorkletNodeOptions) {
		super();
		// Get sampleRate from processorOptions or fall back to global sampleRate
		this.#sampleRate = options.processorOptions?.sampleRate || (globalThis as any).sampleRate;
		// Get target number of channels from processorOptions
		this.#targetChannels = options.processorOptions?.targetChannels || 1;
	}

	process(inputs: Float32Array[][]) {
		if (inputs.length > 1) throw new Error("only one input is supported.");

		// Just take one input channel, the first one.
		// MOQ enables the delivery of audio inputs individually for each track.
		// So do not mix audio from different tracks or different devices.
		const channels = inputs[0];

		if (!channels || channels.length === 0|| !channels[0]) {
			return true
		}

		const inputChannels = channels.length;
		const numberOfFrames = channels[0].length;
		
		// Use target channels from configuration, not input channels
		const numberOfChannels = this.#targetChannels;
		const data = new Float32Array(numberOfChannels * numberOfFrames);
		
		for (let i = 0; i < numberOfChannels; i++) {
			if (i < inputChannels) {
				const inputChannel = channels[i];
				if (inputChannel && inputChannel.length > 0) {
					// Use input channel data
					data.set(inputChannel, i * numberOfFrames);
				} else {
					// Fill with silence if input channel is empty
					data.fill(0, i * numberOfFrames, (i + 1) * numberOfFrames);
				}
			} else if (inputChannels > 0) {
				const firstChannel = channels[0];
				if (firstChannel && firstChannel.length > 0) {
					// If we need more channels than input provides, duplicate the first channel
					data.set(firstChannel, i * numberOfFrames);
				} else {
					// Fill with silence if first channel is empty
					data.fill(0, i * numberOfFrames, (i + 1) * numberOfFrames);
				}
			} else {
				// Fill with silence if no input data
				data.fill(0, i * numberOfFrames, (i + 1) * numberOfFrames);
			}
		}

		const init: AudioDataInit = {
			format: "f32-planar",
			sampleRate: this.#sampleRate,
			numberOfChannels: numberOfChannels,
			numberOfFrames: numberOfFrames,
			data: data,
			timestamp: Math.round(this.#currentFrame * 1_000_000 / this.#sampleRate),
			transfer: [data.buffer],
		}

		this.port.postMessage(init);

		this.#currentFrame += numberOfFrames;

		return true;
	}
}

	registerProcessor("AudioHijacker", AudioHijackProcessor);
}
