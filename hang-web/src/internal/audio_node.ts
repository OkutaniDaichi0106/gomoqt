// Audio node API: AudioNode, AudioEncodeNode, AudioDecodeNode
// Based on VideoEncodeNode and VideoDecodeNode patterns
// Uses Web Audio API AudioEncoder/AudioDecoder for encoding/decoding
import { GroupCache } from "./cache";
import { EncodedContainer, cloneChunk } from "./container";
import type { EncodedChunk, EncodeDestination } from "./container";
import { TrackWriter, TrackReader,InternalSubscribeErrorCode } from "@okutanidaichi/moqt";
import { readVarint } from "@okutanidaichi/moqt/io";
import { importWorkletUrl as importOffloadWorkletUrl, workletName as offloadWorkletName} from "./audio_offload_worklet";
import { workletName as hijackWorkletName, importWorkletUrl as importHijackWorkletUrl } from "./audio_hijack_worklet";

export class AudioEncodeNode implements AudioNode {
	#encoder: AudioEncoder;
	context: AudioContext;
	#worklet?: AudioWorkletNode;

	#dests: Set<EncodeDestination> = new Set();

	// Event listeners storage
	#eventListeners: Map<string, Set<EventListenerOrEventListenerObject>> = new Map();


	constructor(context: AudioContext, options?: { startSequence?: bigint; }) {
		this.context = context;

		this.#encoder = new AudioEncoder({
			output: async (chunk) => {
				await Promise.all(Array.from(this.#dests, dest => dest.output(chunk)));
			},
			error: (e) => {
				console.error('AudioEncoder error:', e);
			},
		});

		const self = this;
		const readable = new ReadableStream<AudioData>({
            async start(controller) {
                await context.audioWorklet.addModule(importHijackWorkletUrl());

                const worklet = new AudioWorkletNode(
					context,
					hijackWorkletName,
					{
						numberOfInputs: 1,
						numberOfOutputs: 0,
						channelCount: context.destination.channelCount,
						processorOptions: {
							sampleRate: context.sampleRate,
							targetChannels: context.destination.channelCount || 1,
						},
					}
				);
				self.#worklet = worklet;

                worklet.port.onmessage = ({data}: {data: AudioDataInit}) => {
                    const frame = new AudioData(data);
                    controller.enqueue(frame);
                };
            },
            cancel() {
				// TODO: Clean up worklet if needed
            },
        });

		queueMicrotask(() => this.#next(readable.getReader()));
	}

	// Dummy AudioNode methods for interface compatibility
	connect(destinationNode: AudioNode, output?: number, input?: number): AudioNode;
	connect(destinationParam: AudioParam, output?: number): void;
	connect(destination: AudioNode | AudioParam, output?: number, input?: number): AudioNode | void {
		// This node does not output audio to the graph, so connections are not supported
		throw new Error("AudioEncodeNode does not support connections as it does not output audio");
	}

	disconnect(): void;
	disconnect(output: number): void;
	disconnect(destinationNode: AudioNode): void;
	disconnect(destinationNode: AudioNode, output: number): void;
	disconnect(destinationNode: AudioNode, output: number, input: number): void;
	disconnect(destinationParam: AudioParam): void;
	disconnect(destinationParam: AudioParam, output: number): void;
	disconnect(
		destinationOrOutput?: number | AudioNode | AudioParam,
		output?: number,
		input?: number
	): void {
		// No-op
	}

	addEventListener(type: string, listener: EventListenerOrEventListenerObject, options?: boolean | AddEventListenerOptions): void {
		if (!this.#eventListeners.has(type)) {
			this.#eventListeners.set(type, new Set());
		}
		this.#eventListeners.get(type)!.add(listener);
	}

	removeEventListener(type: string, listener: EventListenerOrEventListenerObject, options?: boolean | EventListenerOptions): void {
		const listeners = this.#eventListeners.get(type);
		if (listeners) {
			listeners.delete(listener);
			if (listeners.size === 0) {
				this.#eventListeners.delete(type);
			}
		}
	}

	dispatchEvent(event: Event): boolean {
		const listeners = this.#eventListeners.get(event.type);
		if (listeners) {
			for (const listener of listeners) {
				if (typeof listener === 'function') {
					listener(event);
				} else {
					listener.handleEvent(event);
				}
			}
		}
		return !event.defaultPrevented;
	}


	configure(config: AudioEncoderConfig): void {
		this.#encoder.configure(config);
	}

	async #next(stream: ReadableStreamDefaultReader<AudioData>): Promise<void> {
		const { done, value } = await stream.read();
		if (done) {
			stream.releaseLock();
			return;
		}

		try {
			this.#encoder.encode(value);
		} catch (e) {
			console.error('encode error', e);
		}

		queueMicrotask(() => this.#next(stream));
	}

	process(input: AudioData): void {
		try {
			this.#encoder.encode(input);
		} catch (e) {
			console.error('encode error', e);
		}

		if (input) {
			try {
				input.close();
			} catch (_) {
				/* ignore */
			}
		}
	}

	async close(): Promise<void> {
		try {
			this.#encoder.close();
		} catch (_) {
			/* ignore */
		}
	}

	// Implement required AudioNode properties for compatibility
	get channelCount(): number {
		return this.#worklet?.channelCount || 1;
	}

	get channelCountMode(): ChannelCountMode {
		return this.#worklet?.channelCountMode || "explicit";
	}

	get channelInterpretation(): ChannelInterpretation {
		return this.#worklet?.channelInterpretation || "speakers";
	}

	get numberOfInputs(): number {
		return this.#worklet?.numberOfInputs || 1;
	}

	get numberOfOutputs(): number {
		return this.#worklet?.numberOfOutputs || 0;
	}

	async encodeTo(callbacks: EncodeDestination): Promise<void> {
		this.#dests.add(callbacks);

		await Promise.race([
			callbacks.done,
		]);

		this.#dests.delete(callbacks);
	}
}

export class AudioDecodeNode implements AudioNode {
	#decoder: AudioDecoder;
	context: AudioContext;
	#worklet?: AudioWorkletNode;

	constructor(context: AudioContext, init: { latency?: number; } = {}) {
		this.context = context;

        context.audioWorklet.addModule(importOffloadWorkletUrl()).then(() => {
			// Create AudioWorkletNode
			this.#worklet = new AudioWorkletNode(
				context,
				offloadWorkletName,
				{
					channelCount: context.destination.channelCount,
					numberOfInputs: 0,
					numberOfOutputs: 1,
					processorOptions: {
						sampleRate: context.sampleRate,
						latency: init.latency || 100, // Default to 100ms if not specified
					},
				}
			);
		}).catch((error) => {
			console.error('failed to load AudioWorklet module:', error);
		});

		this.#decoder = new AudioDecoder({
			output: async (frame) => {
				// Pass audio frame
                this.process(frame);
			},
			error: (e) => {
				console.error('AudioDecoder error:', e);
			},
		});
	}


	// Implement missing AudioNode methods
	disconnect(): void;
	disconnect(output: number): void;
	disconnect(destinationNode: AudioNode): void;
	disconnect(destinationNode: AudioNode, output: number): void;
	disconnect(destinationNode: AudioNode, output: number, input: number): void;
	disconnect(destinationParam: AudioParam): void;
	disconnect(destinationParam: AudioParam, output: number): void;
	disconnect(
		destinationOrOutput?: number | AudioNode | AudioParam,
		output?: number,
		input?: number
	): void {
		if (arguments.length === 0) {
			this.#worklet?.disconnect();
		} else if (typeof destinationOrOutput === "number") {
			this.#worklet?.disconnect(destinationOrOutput);
		} else if (
			typeof window !== "undefined" &&
			typeof AudioNode !== "undefined" &&
			destinationOrOutput instanceof AudioNode
		) {
			if (output !== undefined && input !== undefined) {
				this.#worklet?.disconnect(destinationOrOutput, output, input);
			} else if (output !== undefined) {
				this.#worklet?.disconnect(destinationOrOutput, output);
			} else {
				this.#worklet?.disconnect(destinationOrOutput);
			}
		} else if (
			typeof window !== "undefined" &&
			typeof AudioParam !== "undefined" &&
			destinationOrOutput instanceof AudioParam
		) {
			if (output !== undefined) {
				this.#worklet?.disconnect(destinationOrOutput, output);
			} else {
				this.#worklet?.disconnect(destinationOrOutput);
			}
		} else {
			this.#worklet?.disconnect();
		}
	}

	addEventListener(type: string, listener: EventListenerOrEventListenerObject, options?: boolean | AddEventListenerOptions): void {
		this.#worklet?.addEventListener(type, listener, options);
	}

	removeEventListener(type: string, listener: EventListenerOrEventListenerObject, options?: boolean | EventListenerOptions): void {
		this.#worklet?.removeEventListener(type, listener, options);
	}

	dispatchEvent(event: Event): boolean {
		return this.#worklet?.dispatchEvent(event) || false;
	}

	get numberOfInputs(): number {
		return this.#worklet?.numberOfInputs || 1;
	}

	get numberOfOutputs(): number {
		return this.#worklet?.numberOfOutputs || 1;
	}

	get channelCount(): number {
		return this.#worklet?.channelCount || 1;
	}

	get channelCountMode(): ChannelCountMode {
		return "explicit";
	}

	get channelInterpretation(): ChannelInterpretation {
		return "speakers";
	}

	connect(destinationNode: AudioNode, output?: number, input?: number): AudioNode;
	connect(destinationParam: AudioParam, output?: number): void;
	connect(destination: AudioNode | AudioParam, output?: number, input?: number): AudioNode | void {
		if ((window.AudioNode && destination instanceof window.AudioNode) || (typeof AudioNode !== "undefined" && destination instanceof AudioNode)) {
			// Connect to another AudioNode
			this.#worklet?.connect(destination, output, input);
			return destination as AudioNode;
		} else if ((window.AudioParam && destination instanceof window.AudioParam) || (typeof AudioParam !== "undefined" && destination instanceof AudioParam)) {
			// Connect to an AudioParam
			this.#worklet?.connect(destination, output);
			return;
		} else {
			throw new TypeError("Invalid destination for connect()");
		}
	}


	configure(config: AudioDecoderConfig): void {
		this.#decoder.configure(config);
	}

	async decodeFrom(stream: ReadableStream<EncodedAudioChunk>): Promise<void> {
		try {
			const reader = stream.getReader();

			const { done, value: chunk } = await reader.read();
			if (done) {
				reader.releaseLock();
				return;
			}

			this.#decoder.decode(chunk);
		} catch (e) {
			console.error('AudioDecodeNode decodeFrom error:', e);
		}
	}

	process(input: AudioData): void {
        // No longer drops frames when muted; gain handles silence for continuity.
        const channels: Float32Array[] = [];
        for (let i = 0; i < input.numberOfChannels; i++) {
            const data = new Float32Array(input.numberOfFrames);
            input.copyTo(data, { format: "f32-planar", planeIndex: i });
            channels.push(data);
        }

        // Send AudioData to the worklet
        this.#worklet?.port.postMessage(
            {
                channels: channels,
                timestamp: input.timestamp,
            },
            channels.map(d => d.buffer) // Transfer ownership of the buffers
        );

        // Close the frame after sending
        try {
			input.close();
		} catch (_) {
			/* ignore */
		}
	}

	async flush(): Promise<void> {
		try {
			await this.#decoder.flush();
		} catch (e) {
			console.error('AudioDecoder flush error:', e);
		}
	}

	async close(): Promise<void> {
		try {
			await this.#decoder.flush();
			this.#decoder.close();
		} catch (_) {
			/* ignore */
		}
	}

	dispose(): void {
		this.disconnect();
	}
}