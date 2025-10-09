import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
	AudioEncodeNode,
	AudioDecodeNode,
} from './audio_node';
import { EncodedContainer } from ".";

// Mock implementations for Web APIs
class MockAudioData implements AudioData {
	numberOfFrames: number;
	numberOfChannels: number;
	sampleRate: number;
	format: AudioSampleFormat;
	readonly duration: number;
	readonly timestamp: number;

	constructor(frames: number = 1024, channels: number = 2, sampleRate: number = 44100) {
		this.numberOfFrames = frames;
		this.numberOfChannels = channels;
		this.sampleRate = sampleRate;
		this.format = 'f32';
		this.duration = (frames / sampleRate) * 1000000; // microseconds
		this.timestamp = 0;
	}

	copyTo(destination: AllowSharedBufferSource, options?: AudioDataCopyToOptions): void {
		// Fill with test audio data
		if (destination instanceof Float32Array) {
			for (let i = 0; i < destination.length; i++) {
				destination[i] = Math.sin(i * 0.01); // Simple sine wave
			}
		}
	}

	clone(): AudioData {
		return new MockAudioData(this.numberOfFrames, this.numberOfChannels, this.sampleRate);
	}

	close(): void {
		// Mock close
	}

	allocationSize(options: AudioDataCopyToOptions): number {
		return this.numberOfFrames * this.numberOfChannels * 4; // 4 bytes per float32
	}
}

class MockEncodedAudioChunk implements EncodedAudioChunk {
	timestamp: number;
	duration: number | null;
	type: EncodedAudioChunkType;
	byteLength: number;
	copyTo(destination: AllowSharedBufferSource): void {
		// Mock copy
	}

	constructor(type: EncodedAudioChunkType = 'key', timestamp: number = 0) {
		this.type = type;
		this.timestamp = timestamp;
		this.duration = null;
		this.byteLength = 1024;
	}
}

class MockAudioEncoder {
	configure = vi.fn();
	encode = vi.fn();
	close = vi.fn();
	flush = vi.fn();
}

class MockAudioDecoder {
	configure = vi.fn();
	decode = vi.fn();
	close = vi.fn();
	flush = vi.fn();
}

class MockAudioNode implements AudioNode {
	readonly numberOfInputs: number;
	readonly numberOfOutputs: number;
	readonly channelCount: number;
	readonly channelCountMode: ChannelCountMode;
	readonly channelInterpretation: ChannelInterpretation;
	readonly context: AudioContext;
	readonly inputs: Set<AudioNode>;
	readonly outputs: Set<AudioNode>;

	constructor(options?: { numberOfInputs?: number; numberOfOutputs?: number }) {
		this.numberOfInputs = options?.numberOfInputs ?? 1;
		this.numberOfOutputs = options?.numberOfOutputs ?? 1;
		this.channelCount = 2;
		this.channelCountMode = "explicit";
		this.channelInterpretation = "speakers";
		this.context = {} as AudioContext;
		this.inputs = new Set();
		this.outputs = new Set();
	}

	connect(destinationNode: AudioNode, output?: number, input?: number): AudioNode;
	connect(destinationParam: AudioParam, output?: number): void;
	connect(destination: AudioNode | AudioParam, output?: number, input?: number): AudioNode | void {
		if (destination instanceof (global as any).AudioNode) {
			this.outputs.add(destination as MockAudioNode);
			(destination as MockAudioNode).inputs.add(this);
			return destination as AudioNode;
		} else if (destination instanceof (global as any).AudioParam) {
			return undefined;
		}
		return undefined;
	}

	disconnect(): void;
	disconnect(output: number): void;
	disconnect(destinationNode: AudioNode): void;
	disconnect(destinationNode: AudioNode, output: number): void;
	disconnect(destinationNode: AudioNode, output: number, input: number): void;
	disconnect(destinationParam: AudioParam): void;
	disconnect(destinationParam: AudioParam, output: number): void;
	disconnect(destinationOrOutput?: number | AudioNode | AudioParam, output?: number, input?: number): void {
		if (destinationOrOutput instanceof (global as any).AudioNode) {
			this.outputs.delete(destinationOrOutput as MockAudioNode);
			(destinationOrOutput as MockAudioNode).inputs.delete(this);
		} else if (typeof destinationOrOutput === "number") {
			// disconnect by output index - not implemented in mock
		} else {
			for (const dest of this.outputs) {
				(dest as MockAudioNode).inputs.delete(this);
			}
			this.outputs.clear();
		}
	}

	addEventListener(type: string, listener: EventListenerOrEventListenerObject, options?: boolean | AddEventListenerOptions): void {
		// Mock implementation
	}

	removeEventListener(type: string, listener: EventListenerOrEventListenerObject, options?: boolean | EventListenerOptions): void {
		// Mock implementation
	}

	dispatchEvent(event: Event): boolean {
		return !event.defaultPrevented;
	}

	process(input?: AudioData | EncodedContainer): void {
		// Mock implementation
	}

	dispose(): void {
		this.disconnect();
	}
}

// Mock global AudioNode for instanceof checks
(global as any).AudioNode = MockAudioNode;

// Mock AudioWorkletNode
class MockAudioWorkletNode {
	port = { 
		postMessage: vi.fn(),
		onmessage: null as ((event: { data: any }) => void) | null
	};
	connect = vi.fn();
	disconnect = vi.fn();
	addEventListener = vi.fn();
	removeEventListener = vi.fn();
	dispatchEvent = vi.fn().mockReturnValue(true);
}
(global as any).AudioWorkletNode = MockAudioWorkletNode;

// Setup global window for browser environment checks
beforeEach(() => {
	(global as any).window = {
		AudioNode: MockAudioNode,
		AudioParam: function() {},
	};
	(global as any).AudioData = MockAudioData;
});

describe('AudioNode', () => {
	it('should create AudioNode with default options', () => {
		const node = new MockAudioNode();
		expect(node.numberOfInputs).toBe(1);
		expect(node.numberOfOutputs).toBe(1);
	});

	it('should create AudioNode with custom options', () => {
		const node = new MockAudioNode({ numberOfInputs: 2, numberOfOutputs: 3 });
		expect(node.numberOfInputs).toBe(2);
		expect(node.numberOfOutputs).toBe(3);
	});

	it('should connect and disconnect nodes', () => {
		const node1 = new MockAudioNode();
		const node2 = new MockAudioNode();

		node1.connect(node2);
		expect(node1.outputs.has(node2)).toBe(true);
		expect(node2.inputs.has(node1)).toBe(true);

		node1.disconnect(node2);
		expect(node1.outputs.has(node2)).toBe(false);
		expect(node2.inputs.has(node1)).toBe(false);
	});

	it('should disconnect all outputs', () => {
		const node1 = new MockAudioNode();
		const node2 = new MockAudioNode();
		const node3 = new MockAudioNode();

		node1.connect(node2);
		node1.connect(node3);
		expect(node1.outputs.size).toBe(2);

		node1.disconnect();
		expect(node1.outputs.size).toBe(0);
		expect(node2.inputs.has(node1)).toBe(false);
		expect(node3.inputs.has(node1)).toBe(false);
	});

	it('should dispose node', () => {
		const node = new MockAudioNode();
		expect(() => node.dispose()).not.toThrow();
	});
});

describe('AudioEncodeNode', () => {
	let encodeNode: AudioEncodeNode;
	let mockContext: AudioContext;
	let mockEncoder: MockAudioEncoder;
	let mockWorklet: MockAudioWorkletNode;

	beforeEach(() => {
		mockContext = {
			audioWorklet: { addModule: vi.fn().mockResolvedValue(undefined) },
			destination: { channelCount: 2 },
			sampleRate: 44100
		} as any;
		mockEncoder = new MockAudioEncoder();
		mockWorklet = new MockAudioWorkletNode();
		(global as any).AudioEncoder = vi.fn(() => mockEncoder);
		(global as any).AudioWorkletNode = vi.fn(() => mockWorklet);

		encodeNode = new AudioEncodeNode(mockContext);
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('should create AudioEncodeNode', () => {
		expect(encodeNode).toBeInstanceOf(AudioEncodeNode);
		expect(encodeNode.numberOfInputs).toBe(1);
		expect(encodeNode.numberOfOutputs).toBe(1);
		expect(encodeNode.context).toBe(mockContext);
	});

	it('should start encoding from worklet stream', async () => {
		// Simulate worklet sending audio data
		const audioData = new MockAudioData();
		const audioDataInit = {
			sampleRate: audioData.sampleRate,
			numberOfFrames: audioData.numberOfFrames,
			numberOfChannels: audioData.numberOfChannels,
			format: audioData.format,
			timestamp: audioData.timestamp,
			data: new Float32Array(audioData.numberOfFrames * audioData.numberOfChannels)
		};
		if (mockWorklet.port.onmessage) {
			mockWorklet.port.onmessage({ data: audioDataInit });
		}

		// Wait for encode to be called
		await vi.waitFor(() => {
			expect(mockEncoder.encode).toHaveBeenCalled();
		}, { timeout: 100 });
	});

	it('should configure encoder', () => {
		const config: AudioEncoderConfig = {
			codec: 'opus',
			sampleRate: 44100,
			numberOfChannels: 2
		};

		expect(() => encodeNode.configure(config)).not.toThrow();
		expect(mockEncoder.configure).toHaveBeenCalledWith(config);
	});

	it('should process audio data and encode', () => {
		const config: AudioEncoderConfig = {
			codec: 'opus',
			sampleRate: 44100,
			numberOfChannels: 2
		};
		encodeNode.configure(config);

		const audioData = new MockAudioData();
		expect(() => encodeNode.process(audioData)).not.toThrow();
		expect(mockEncoder.encode).toHaveBeenCalledWith(audioData);
	});

	it('should handle process encode errors gracefully', () => {
		const audioData = new MockAudioData();
		mockEncoder.encode.mockImplementationOnce(() => { throw new Error('encode error'); });
		expect(() => encodeNode.process(audioData)).not.toThrow();
	});

	it('should throw error on connect', () => {
		const mockDestination = {} as AudioNode;
		expect(() => encodeNode.connect(mockDestination)).toThrow('AudioEncodeNode does not support connections as it does not output audio');
	});

	it('should not throw on disconnect', () => {
		expect(() => encodeNode.disconnect()).not.toThrow();
		expect(() => encodeNode.disconnect(0)).not.toThrow();
		expect(() => encodeNode.disconnect({} as AudioNode)).not.toThrow();
	});

	it('should manage event listeners', () => {
		const mockListener = vi.fn();
		const event = new Event('test');

		// Add listener
		encodeNode.addEventListener('test', mockListener);
		encodeNode.dispatchEvent(event);
		expect(mockListener).toHaveBeenCalledWith(event);

		// Remove listener
		encodeNode.removeEventListener('test', mockListener);
		mockListener.mockClear();
		encodeNode.dispatchEvent(event);
		expect(mockListener).not.toHaveBeenCalled();
	});

	it('should handle multiple listeners for same event type', () => {
		const mockListener1 = vi.fn();
		const mockListener2 = vi.fn();
		const event = new Event('test');

		encodeNode.addEventListener('test', mockListener1);
		encodeNode.addEventListener('test', mockListener2);
		encodeNode.dispatchEvent(event);

		expect(mockListener1).toHaveBeenCalledWith(event);
		expect(mockListener2).toHaveBeenCalledWith(event);
	});

	it('should handle EventListenerObject', () => {
		const mockListener = { handleEvent: vi.fn() };
		const event = new Event('test');

		encodeNode.addEventListener('test', mockListener);
		encodeNode.dispatchEvent(event);

		expect(mockListener.handleEvent).toHaveBeenCalledWith(event);
	});

	it('should return correct value from dispatchEvent', () => {
		const normalEvent = new Event('test');
		class TestEvent extends Event {
			defaultPrevented = false;
			preventDefault() {
				this.defaultPrevented = true;
			}
		}
		const preventedEvent = new TestEvent('test');
		preventedEvent.preventDefault();

		expect(encodeNode.dispatchEvent(normalEvent)).toBe(true);
		expect(encodeNode.dispatchEvent(preventedEvent)).toBe(false);
	});

	it('should return correct AudioNode properties', () => {
		expect(encodeNode.numberOfInputs).toBe(1);
		expect(encodeNode.numberOfOutputs).toBe(1);
		expect(encodeNode.channelCount).toBe(1);
		expect(encodeNode.channelCountMode).toBe("explicit");
		expect(encodeNode.channelInterpretation).toBe("speakers");
	});

	it('should close encoder on close', async () => {
		await expect(encodeNode.close()).resolves.not.toThrow();
		expect(mockEncoder.close).toHaveBeenCalled();
	});

	it('should handle close errors gracefully', async () => {
		mockEncoder.close.mockRejectedValueOnce(new Error('close error'));
		await expect(encodeNode.close()).resolves.not.toThrow();
	});

	it('should serve track and manage tracks set', async () => {
		const mockTrack = { context: { done: vi.fn().mockResolvedValue(undefined) } } as any;
		const ctx = Promise.resolve();

		const servePromise = encodeNode.serveTrack(ctx, mockTrack);
		await expect(servePromise).resolves.not.toThrow();
		expect(encodeNode.tracks.has(mockTrack)).toBe(false); // should be removed after completion
	});
});

describe('AudioDecodeNode', () => {
	let decoderNode: AudioDecodeNode;
	let mockContext: AudioContext;
	let mockDecoder: MockAudioDecoder;
	let mockWorklet: any;

	beforeEach(() => {
		mockContext = {
			audioWorklet: { addModule: vi.fn().mockResolvedValue(undefined) },
			destination: { channelCount: 2 },
			sampleRate: 44100
		} as any;
		mockDecoder = new MockAudioDecoder();
		mockWorklet = {
			connect: vi.fn(),
			disconnect: vi.fn(),
			addEventListener: vi.fn(),
			removeEventListener: vi.fn(),
			dispatchEvent: vi.fn().mockReturnValue(false),
			numberOfInputs: 1,
			numberOfOutputs: 1,
			channelCount: 2,
			port: { postMessage: vi.fn() }
		};
		(global as any).AudioDecoder = vi.fn(() => mockDecoder);
		(global as any).AudioWorkletNode = vi.fn(() => mockWorklet);
		(global as any).AudioNode = function() {};
		(global as any).AudioParam = function() {};

		decoderNode = new AudioDecodeNode(mockContext);
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('should create AudioDecodeNode', () => {
		expect(decoderNode).toBeInstanceOf(AudioDecodeNode);
		expect(decoderNode.context).toBe(mockContext);
	});

	it('should configure decoder', () => {
		const config: AudioDecoderConfig = {
			codec: 'opus',
			sampleRate: 44100,
			numberOfChannels: 2
		};

		decoderNode.configure(config);
		expect(mockDecoder.configure).toHaveBeenCalledWith(config);
	});

	it('should connect to AudioNode', () => {
		const mockDestination = new (global as any).AudioNode();
		const result = decoderNode.connect(mockDestination);
		expect(mockWorklet.connect).toHaveBeenCalledWith(mockDestination, undefined, undefined);
		expect(result).toBe(mockDestination);
	});

	it('should connect to AudioParam', () => {
		const mockDestination = new (global as any).AudioParam();
		const result = decoderNode.connect(mockDestination);
		expect(mockWorklet.connect).toHaveBeenCalledWith(mockDestination, undefined);
		expect(result).toBeUndefined();
	});

	it('should throw on invalid connect destination', () => {
		expect(() => decoderNode.connect({} as any)).toThrow('Invalid destination for connect()');
	});

	it('should disconnect worklet', () => {
		decoderNode.disconnect();
		expect(mockWorklet.disconnect).toHaveBeenCalledWith();

		decoderNode.disconnect(0);
		expect(mockWorklet.disconnect).toHaveBeenCalledWith(0);

		const mockNode = new (global as any).AudioNode();
		decoderNode.disconnect(mockNode);
		expect(mockWorklet.disconnect).toHaveBeenCalledWith(mockNode);

		decoderNode.disconnect(mockNode, 1);
		expect(mockWorklet.disconnect).toHaveBeenCalledWith(mockNode, 1);
	});

	it('should delegate event listeners to worklet', () => {
		const mockListener = vi.fn();
		decoderNode.addEventListener('test', mockListener);
		expect(mockWorklet.addEventListener).toHaveBeenCalledWith('test', mockListener, undefined);

		decoderNode.removeEventListener('test', mockListener);
		expect(mockWorklet.removeEventListener).toHaveBeenCalledWith('test', mockListener, undefined);

		const event = new Event('test');
		const result = decoderNode.dispatchEvent(event);
		expect(mockWorklet.dispatchEvent).toHaveBeenCalledWith(event);
		expect(result).toBe(false); // worklet returns undefined, so false
	});

	it('should return worklet properties', () => {
		expect(decoderNode.numberOfInputs).toBe(1);
		expect(decoderNode.numberOfOutputs).toBe(1);
		expect(decoderNode.channelCount).toBe(2);
		expect(decoderNode.channelCountMode).toBe("explicit");
		expect(decoderNode.channelInterpretation).toBe("speakers");
	});

	it('should process AudioData', () => {
		const audioData = new MockAudioData();
		expect(() => decoderNode.process(audioData)).not.toThrow();
		expect(mockWorklet.port.postMessage).toHaveBeenCalled();
	});

	it('should handle process errors gracefully', () => {
		const audioData = new MockAudioData();
		audioData.close = vi.fn().mockImplementation(() => { throw new Error('close error'); });
		expect(() => decoderNode.process(audioData)).not.toThrow();
	});

	it('should decode from track reader', async () => {
		const config: AudioDecoderConfig = {
			codec: 'opus',
			sampleRate: 44100,
			numberOfChannels: 2
		};
		decoderNode.configure(config);

		const mockReader = {
			acceptGroup: vi.fn().mockResolvedValueOnce([{ readFrame: vi.fn().mockResolvedValueOnce([{ bytes: new Uint8Array([0, 0, 0, 0, 1, 2, 3, 4]) }, null]).mockResolvedValueOnce([null, new Error('end')]) }, null]).mockResolvedValueOnce([null, new Error('end')])
		} as any;
		const ctx = Promise.resolve();

		await expect(decoderNode.decodeFrom(ctx, mockReader)).resolves.not.toThrow();
		expect(mockDecoder.decode).toHaveBeenCalled();
	});

	it('should handle decodeFrom errors gracefully', async () => {
		const config: AudioDecoderConfig = {
			codec: 'opus',
			sampleRate: 44100,
			numberOfChannels: 2
		};
		decoderNode.configure(config);

		const mockReader = {
			acceptGroup: vi.fn().mockRejectedValue(new Error('acceptGroup error'))
		} as any;
		const ctx = Promise.resolve();

		await expect(decoderNode.decodeFrom(ctx, mockReader)).resolves.not.toThrow();
	});

	it('should close decoder on close', async () => {
		await expect(decoderNode.close()).resolves.not.toThrow();
		expect(mockDecoder.flush).toHaveBeenCalled();
		expect(mockDecoder.close).toHaveBeenCalled();
	});

	it('should handle flush errors gracefully', async () => {
		mockDecoder.flush.mockRejectedValueOnce(new Error('flush error'));
		await expect(decoderNode.flush()).resolves.not.toThrow();
	});

	it('should handle close errors gracefully', async () => {
		mockDecoder.flush.mockRejectedValueOnce(new Error('flush error'));
		mockDecoder.close.mockRejectedValueOnce(new Error('close error'));
		await expect(decoderNode.close()).resolves.not.toThrow();
	});

	it('should dispose node', () => {
		expect(() => decoderNode.dispose()).not.toThrow();
		expect(mockWorklet.disconnect).toHaveBeenCalled();
	});
});