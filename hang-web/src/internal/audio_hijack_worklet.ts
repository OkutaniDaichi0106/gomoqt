class AudioHijackProcessor extends AudioWorkletProcessor {

	constructor(options: AudioWorkletNodeOptions) {
		super();
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

		const numberOfChannels = channels.length;
		const numberOfFrames = channels[0].length;
		const data = new Float32Array(numberOfChannels * numberOfFrames);
		for (let i = 0; i < numberOfChannels; i++) {
			const arr = channels[i];
			if (!arr || arr.length === 0) {
				data.fill(0, i * numberOfFrames, (i + 1) * numberOfFrames);
				continue;
			}

			data.set(arr, i * numberOfFrames);
		}

		const init: AudioDataInit = {
			format: "f32-planar",
			sampleRate: sampleRate,
			numberOfChannels: numberOfChannels,
			numberOfFrames: numberOfFrames,
			data: data,
			timestamp: Math.round(currentFrame * 1_000_000 / sampleRate),
			transfer: [data.buffer],
		}

		this.port.postMessage(init);

		return true;
	}
}

registerProcessor("AudioHijacker", AudioHijackProcessor);
