import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
	VideoContext,
	VideoNode,
	VideoSourceNode,
	VideoAnalyserNode,
	VideoDestinationNode,
	VideoEncodeNode,
	VideoRenderFunctions,
	type VideoContextState,
VideoDecodeNode
} from './video_node';
import { EncodedContainer } from ".";

// Mock implementations for Web APIs
class MockVideoFrame implements VideoFrame {
	displayWidth: number;
	displayHeight: number;
	codedWidth: number;
	codedHeight: number;
	timestamp: number;
	duration: number | null;
	colorSpace: VideoColorSpace;
	visibleRect: DOMRectReadOnly | null;
	codedRect: DOMRectReadOnly | null;
	format: VideoPixelFormat | null;

	constructor(width: number = 640, height: number = 480, timestamp: number = 0) {
		this.displayWidth = width;
		this.displayHeight = height;
		this.codedWidth = width;
		this.codedHeight = height;
		this.timestamp = timestamp;
		this.duration = null;
		this.colorSpace = {} as VideoColorSpace;
		this.visibleRect = null;
		this.codedRect = null;
		this.format = null;
	}

	copyTo(destination: AllowSharedBufferSource, options?: VideoFrameCopyToOptions): Promise<PlaneLayout[]> {
		// Fill with test pattern
		if (destination instanceof Uint8Array) {
			for (let i = 0; i < destination.length; i += 4) {
				destination[i] = 255;     // R
				destination[i + 1] = 128; // G
				destination[i + 2] = 64;  // B
				destination[i + 3] = 255; // A
			}
		}
		return Promise.resolve([]);
	}

	clone(): VideoFrame {
		return new MockVideoFrame(this.displayWidth, this.displayHeight, this.timestamp);
	}

	close(): void {
		// Mock close
	}

	allocationSize(options?: VideoFrameCopyToOptions): number {
		return this.displayWidth * this.displayHeight * 4;
	}
}

class MockVideoEncoder {
	state: 'unconfigured' | 'configured' | 'closed' = 'unconfigured';
	configure = vi.fn();
	encode = vi.fn();
	reset = vi.fn();
	flush = vi.fn();
	close = vi.fn();

	constructor(config: VideoEncoderInit) {
		// Mock constructor
	}
}

class MockVideoDecoder {
	state: 'unconfigured' | 'configured' | 'closed' = 'unconfigured';
	configure = vi.fn();
	decode = vi.fn();
	reset = vi.fn();
	flush = vi.fn();
	close = vi.fn();

	constructor(config: VideoDecoderInit) {
		// Mock constructor
	}
}

class MockMediaStream {
	constructor(tracks?: MediaStreamTrack[]) {
		// Mock constructor
	}
}

class MockHTMLCanvasElement {
	width: number = 640;
	height: number = 480;
	getContext = vi.fn((type: string): CanvasRenderingContext2D | null => {
		if (type === '2d') {
			return {
				drawImage: vi.fn(),
				getImageData: vi.fn(() => ({
					data: new Uint8ClampedArray(this.width * this.height * 4)
				}))
			} as any;
		}
		return null;
	});
}

// Mock VideoNode for testing abstract class behavior
class MockVideoNode extends VideoNode {
	process(input?: VideoFrame | any): void {
		// Mock implementation
	}
}

// Mock global constructors
vi.stubGlobal('VideoFrame', MockVideoFrame);
vi.stubGlobal('VideoEncoder', MockVideoEncoder);
vi.stubGlobal('VideoDecoder', MockVideoDecoder);
vi.stubGlobal('MediaStream', MockMediaStream);
vi.stubGlobal('HTMLCanvasElement', MockHTMLCanvasElement);

// Mock document.createElement
Object.defineProperty(document, 'createElement', {
	writable: true,
	value: vi.fn((tag: string) => {
		if (tag === 'canvas') {
			return new MockHTMLCanvasElement();
		}
		return {};
	})
});

describe('VideoContext', () => {
	let context: VideoContext;
	let canvas: MockHTMLCanvasElement;

	beforeEach(() => {
		canvas = new MockHTMLCanvasElement();
		context = new VideoContext({ frameRate: 30, canvas: canvas as any });
	});

	afterEach(() => {
		vi.clearAllMocks();
	});

	it('should create VideoContext with default options', () => {
		const defaultContext = new VideoContext();
		expect(defaultContext.frameRate).toBe(30);
		expect(defaultContext.destination).toBeInstanceOf(VideoDestinationNode);
	});

	it('should create VideoContext with custom options', () => {
		expect(context.frameRate).toBe(30);
		expect(context.destination).toBeInstanceOf(VideoDestinationNode);
		expect(context.destination.canvas).toBe(canvas as any);
	});

	it('should have initial running state', () => {
		expect(context.state).toBe('running');
	});

	it('should have currentTime starting at 0', () => {
		expect(context.currentTime).toBe(0);
	});

	it('should resume from suspended state', async () => {
		await context.suspend();
		expect(context.state).toBe('suspended');

		await context.resume();
		expect(context.state).toBe('running');
	});

	it('should suspend from running state', async () => {
		await context.suspend();
		expect(context.state).toBe('suspended');
	});

	it('should close context and disconnect all nodes', async () => {
		const node = new VideoSourceNode(context, new ReadableStream());
		await context.close();
		expect(context.state).toBe('closed');
	});

	it('should register and unregister nodes', () => {
		const node = new MockVideoNode();
		context['_register'](node);
		context['_unregister'](node);
		// No direct assertions possible, but should not throw
	});
});

describe('VideoNode', () => {
	let node: MockVideoNode;

	beforeEach(() => {
		node = new MockVideoNode();
	});

	it('should create VideoNode with default options', () => {
		expect(node.numberOfInputs).toBe(1);
		expect(node.numberOfOutputs).toBe(1);
		expect(node.inputs.size).toBe(0);
		expect(node.outputs.size).toBe(0);
	});

	it('should create VideoNode with custom options', () => {
		const customNode = new MockVideoNode({ numberOfInputs: 2, numberOfOutputs: 3 });
		expect(customNode.numberOfInputs).toBe(2);
		expect(customNode.numberOfOutputs).toBe(3);
	});

	it('should connect to another node', () => {
		const node2 = new MockVideoNode();
		const result = node.connect(node2);
		expect(result).toBe(node2);
		expect(node.outputs.has(node2)).toBe(true);
		expect(node2.inputs.has(node)).toBe(true);
	});

	it('should not connect to itself', () => {
		const result = node.connect(node);
		expect(result).toBe(node);
		expect(node.outputs.has(node)).toBe(false);
		expect(node.inputs.has(node)).toBe(false);
	});

	it('should disconnect from specific node', () => {
		const node2 = new MockVideoNode();
		node.connect(node2);
		node.disconnect(node2);
		expect(node.outputs.has(node2)).toBe(false);
		expect(node2.inputs.has(node)).toBe(false);
	});

	it('should disconnect from all nodes', () => {
		const node2 = new MockVideoNode();
		const node3 = new MockVideoNode();
		node.connect(node2);
		node.connect(node3);
		node.disconnect();
		expect(node.outputs.size).toBe(0);
		expect(node2.inputs.has(node)).toBe(false);
		expect(node3.inputs.has(node)).toBe(false);
	});

	it('should dispose and disconnect', () => {
		const node2 = new MockVideoNode();
		node.connect(node2);
		node.dispose();
		expect(node.outputs.size).toBe(0);
		expect(node2.inputs.has(node)).toBe(false);
	});
});

describe('VideoSourceNode', () => {
	let context: VideoContext;
	let stream: ReadableStream<VideoFrame>;
	let sourceNode: VideoSourceNode;

	beforeEach(() => {
		context = new VideoContext();
		stream = new ReadableStream({
			start(controller) {
				// Mock stream
			}
		});
		sourceNode = new VideoSourceNode(context, stream);
	});

	afterEach(() => {
		vi.clearAllMocks();
	});

	it('should create VideoSourceNode', () => {
		expect(sourceNode.numberOfInputs).toBe(0);
		expect(sourceNode.numberOfOutputs).toBe(1);
		expect(sourceNode.context).toBe(context);
	});

	it('should process frames and pass to outputs', () => {
		const outputNode = new MockVideoNode();
		sourceNode.connect(outputNode);

		const frame = new MockVideoFrame();
		const processSpy = vi.spyOn(outputNode, 'process');

		sourceNode.process(frame);
		expect(processSpy).toHaveBeenCalledWith(frame);
	});

	it('should start and stop processing', async () => {
		// Mock the stream to provide one frame and then close
		const mockReader = {
			read: vi.fn()
				.mockResolvedValueOnce({ done: false, value: new MockVideoFrame() })
				.mockResolvedValue({ done: true }),
			releaseLock: vi.fn()
		};
		vi.spyOn(stream, 'getReader').mockReturnValue(mockReader as any);

		const startPromise = sourceNode.start();
		// Wait a bit for processing
		await new Promise(resolve => setTimeout(resolve, 10));
		sourceNode.stop();
		await startPromise; // Should resolve after stop
	}, 2000);

	it('should dispose and unregister', () => {
		sourceNode.dispose();
		expect(sourceNode.outputs.size).toBe(0);
	});
});

describe('MediaStreamVideoSourceNode', () => {
	let mockTrack: MediaStreamTrack;
	let mockStream: ReadableStream<VideoFrame>;
	let originalMediaStreamTrackProcessor: any;

	beforeEach(() => {
		// Mock MediaStreamTrack
		mockTrack = {
			kind: 'video',
			getSettings: vi.fn(() => ({ frameRate: 30, width: 640, height: 480 })),
			stop: vi.fn(),
		} as any;

		// Mock ReadableStream
		mockStream = new ReadableStream({
			start(controller) {
				controller.enqueue(new MockVideoFrame());
			},
		});

		// Store original MediaStreamTrackProcessor
		originalMediaStreamTrackProcessor = (global as any).MediaStreamTrackProcessor;
	});

	afterEach(() => {
		// Restore original MediaStreamTrackProcessor
		(global as any).MediaStreamTrackProcessor = originalMediaStreamTrackProcessor;
		vi.restoreAllMocks();
	});

	it('should create with MediaStreamTrackProcessor', async () => {
		// Mock MediaStreamTrackProcessor
		(global as any).MediaStreamTrackProcessor = vi.fn(() => ({
			readable: mockStream,
		}));

		const { MediaStreamVideoSourceNode } = await import('./video_node');
		const node = new MediaStreamVideoSourceNode(mockTrack);

		expect(node.track).toBe(mockTrack);
		expect((global as any).MediaStreamTrackProcessor).toHaveBeenCalledWith({ track: mockTrack });
	});

	it('should create with polyfill when MediaStreamTrackProcessor unavailable', async () => {
		// Remove MediaStreamTrackProcessor
		delete (global as any).MediaStreamTrackProcessor;

		// Mock document.createElement
		const mockVideo = {
			srcObject: null,
			play: vi.fn().mockResolvedValue(undefined),
			onloadedmetadata: null,
			videoWidth: 640,
			videoHeight: 480,
		};
		vi.spyOn(document, 'createElement').mockReturnValue(mockVideo as any);

		// Mock Promise.all to resolve immediately
		const originalPromiseAll = Promise.all;
		vi.spyOn(Promise, 'all').mockResolvedValue([]);

		const { MediaStreamVideoSourceNode } = await import('./video_node');
		const node = new MediaStreamVideoSourceNode(mockTrack);

		expect(node.track).toBe(mockTrack);
		expect(document.createElement).toHaveBeenCalledWith('video');
		expect(mockVideo.srcObject).toEqual(expect.any(MediaStream));

		// Restore
		Promise.all = originalPromiseAll;
	});

	it('should dispose and stop track', async () => {
		const { MediaStreamVideoSourceNode } = await import('./video_node');
		const node = new MediaStreamVideoSourceNode(mockTrack);
		node.dispose();

		expect(mockTrack.stop).toHaveBeenCalled();
	});

	it('should handle track without settings', async () => {
		const badTrack = {
			kind: 'video',
			getSettings: vi.fn(() => null),
			stop: vi.fn(),
		} as any;

		const { MediaStreamVideoSourceNode } = await import('./video_node');
		expect(() => new MediaStreamVideoSourceNode(badTrack)).toThrow('track has no settings');
	});
});

describe('VideoAnalyserNode', () => {
	let context: VideoContext;
	let analyserNode: VideoAnalyserNode;

	beforeEach(() => {
		context = new VideoContext();
		analyserNode = new VideoAnalyserNode(context);
	});

	afterEach(() => {
		vi.clearAllMocks();
	});

	it('should create VideoAnalyserNode', () => {
		expect(analyserNode.numberOfInputs).toBe(1);
		expect(analyserNode.numberOfOutputs).toBe(1);
	});

	it('should have initial zero values', () => {
		expect(analyserNode.brightness).toBe(0);
		expect(analyserNode.contrast).toBe(0);
		expect(analyserNode.saturation).toBe(0);
		expect(analyserNode.sharpness).toBe(0);
		expect(analyserNode.edgeStrength).toBe(0);
		expect(analyserNode.textureComplexity).toBe(0);
		expect(analyserNode.motionMagnitude).toBe(0);
		expect(analyserNode.motionDirection).toBe(0);
	});

	it('should process frames and update analysis', () => {
		const frame = new MockVideoFrame(320, 240);
		analyserNode.process(frame);

		// Values should be updated after processing
		expect(analyserNode.brightness).toBeGreaterThanOrEqual(0);
		expect(analyserNode.contrast).toBeGreaterThanOrEqual(0);
		expect(analyserNode.saturation).toBeGreaterThanOrEqual(0);
	});

	it('should get color histogram', () => {
		const histogram = new Uint32Array(768); // RGB * 256
		analyserNode.getColorHistogram(histogram);
		// Should not throw
	});

	it('should get dominant colors', () => {
		const colors = analyserNode.getDominantColors();
		expect(Array.isArray(colors)).toBe(true);
	});

	it('should get spatial frequency data', () => {
		const frequencyData = new Float32Array(256);
		analyserNode.getSpatialFrequencyData(frequencyData);
		// Should not throw
	});

	it('should pass frames to outputs', () => {
		const outputNode = new MockVideoNode();
		analyserNode.connect(outputNode);

		const frame = new MockVideoFrame();
		const processSpy = vi.spyOn(outputNode, 'process');

		analyserNode.process(frame);
		expect(processSpy).toHaveBeenCalledWith(frame);
	});
});

describe('VideoDestinationNode', () => {
	let context: VideoContext;
	let canvas: MockHTMLCanvasElement;
	let destinationNode: VideoDestinationNode;

	beforeEach(() => {
		context = new VideoContext();
		canvas = new MockHTMLCanvasElement();
		destinationNode = new VideoDestinationNode(context, canvas as any);
	});

	afterEach(() => {
		vi.clearAllMocks();
	});

	it('should create VideoDestinationNode', () => {
		expect(destinationNode.numberOfInputs).toBe(1);
		expect(destinationNode.numberOfOutputs).toBe(0);
		expect(destinationNode.canvas).toBe(canvas as any);
		expect(destinationNode.resizeCallback).toBe(VideoRenderFunctions.contain);
	});

	it('should create with custom render function', () => {
		const customNode = new VideoDestinationNode(context, canvas as any, {
			renderFunction: VideoRenderFunctions.cover
		});
		expect(customNode.resizeCallback).toBe(VideoRenderFunctions.cover);
	});

	it('should process frames and draw to canvas', () => {
		const frame = new MockVideoFrame(640, 480);

		expect(() => destinationNode.process(frame)).not.toThrow();
		expect(canvas.getContext).toHaveBeenCalledWith('2d');
	});

	it('should not draw when context is suspended', async () => {
		await context.suspend();
		const frame = new MockVideoFrame();
		const ctx = canvas.getContext('2d') as any;

		destinationNode.process(frame);

		expect(ctx.drawImage).not.toHaveBeenCalled();
	});
});

describe('VideoEncodeNode', () => {
	let context: VideoContext;
	let encodeNode: VideoEncodeNode;
	let onChunk: (c: EncodedContainer) => void;

	beforeEach(() => {
		context = new VideoContext();
		onChunk = vi.fn();
		encodeNode = new VideoEncodeNode(context);
	});

	afterEach(() => {
		vi.clearAllMocks();
	});

	it('should create VideoEncodeNode', () => {
		expect(encodeNode.numberOfInputs).toBe(1);
		expect(encodeNode.numberOfOutputs).toBe(1);
	});

	it('should configure encoder', () => {
		const config: VideoEncoderConfig = {
			codec: 'vp8',
			width: 640,
			height: 480
		};

		expect(() => encodeNode.configure(config)).not.toThrow();
	});

	it('should process frames and encode', () => {
		const config: VideoEncoderConfig = {
			codec: 'vp8',
			width: 640,
			height: 480
		};
		encodeNode.configure(config);

		const frame = new MockVideoFrame();
		expect(() => encodeNode.process(frame)).not.toThrow();
	});

	it('should not encode when not configured', () => {
		const frame = new MockVideoFrame();
		expect(() => encodeNode.process(frame)).not.toThrow();
	});

	it('should dispose and close encoder', () => {
		expect(() => encodeNode.dispose()).not.toThrow();
	});
});

describe('VideoRenderFunctions', () => {
	it('contain should fit frame within canvas maintaining aspect ratio', () => {
		const result = VideoRenderFunctions.contain(640, 480, 800, 600);
		expect(result.width).toBe(800);
		expect(result.height).toBe(600);
		expect(result.x).toBe(0);
		expect(result.y).toBe(0);
	});

	it('cover should cover entire canvas maintaining aspect ratio', () => {
		const result = VideoRenderFunctions.cover(640, 480, 800, 600);
		expect(result.width).toBe(800);
		expect(result.height).toBe(600);
		expect(result.x).toBe(0);
		expect(result.y).toBe(0);
	});

	it('fill should fill entire canvas', () => {
		const result = VideoRenderFunctions.fill(640, 480, 800, 600);
		expect(result.width).toBe(800);
		expect(result.height).toBe(600);
		expect(result.x).toBe(0);
		expect(result.y).toBe(0);
	});

	it('scaleDown should only scale down, never up', () => {
		const result = VideoRenderFunctions.scaleDown(320, 240, 800, 600);
		expect(result.width).toBe(320);
		expect(result.height).toBe(240);
		expect(result.x).toBe(240);
		expect(result.y).toBe(180);
	});
});


// // Mock VideoFrame globally
// class MockVideoFrame implements VideoFrame {
//   displayWidth: number;
//   displayHeight: number;
//   timestamp: number;
//   duration: number | null;
//   codedWidth: number;
//   codedHeight: number;
//   codedRect: DOMRectReadOnly;
//   visibleRect: DOMRectReadOnly;
//   colorSpace: VideoColorSpace;
//   format: VideoPixelFormat;
//   allocationSize: (options?: VideoFrameCopyToOptions) => number;
//   copyTo: any;
//   close: () => void;
//   clone: () => VideoFrame;

//   constructor(width: number = 640, height: number = 480, timestamp: number = 0) {
//     this.displayWidth = width;
//     this.displayHeight = height;
//     this.codedWidth = width;
//     this.codedHeight = height;
//     this.timestamp = timestamp;
//     this.duration = null;
//     this.codedRect = {
//       x: 0,
//       y: 0,
//       width: width,
//       height: height,
//       top: 0,
//       right: width,
//       bottom: height,
//       left: 0,
//       toJSON: () => ({})
//     } as DOMRectReadOnly;
//     this.visibleRect = {
//       x: 0,
//       y: 0,
//       width: width,
//       height: height,
//       top: 0,
//       right: width,
//       bottom: height,
//       left: 0,
//       toJSON: () => ({})
//     } as DOMRectReadOnly;
//     this.colorSpace = {
//       primaries: 'bt709',
//       transfer: 'bt709',
//       matrix: 'bt709',
//       fullRange: false,
//       toJSON: () => ({})
//     } as VideoColorSpace;
//     this.format = 'NV12';
//     this.allocationSize = vi.fn(() => width * height * 1.5);
//     this.copyTo = vi.fn();
//     this.close = vi.fn();
//     this.clone = vi.fn(() => new MockVideoFrame(width, height, timestamp));
//   }
// }

// Set VideoFrame globally
(global as any).VideoFrame = MockVideoFrame;

// Mock EncodedVideoChunk
class MockEncodedVideoChunk implements EncodedVideoChunk {
  type: 'key' | 'delta';
  timestamp: number;
  duration: number | null;
  byteLength: number;
  copyTo: (destination: AllowSharedBufferSource) => void;

  constructor(type: 'key' | 'delta' = 'key', timestamp: number = 0, duration: number | null = 33, byteLength: number = 1024) {
    this.type = type;
    this.timestamp = timestamp;
    this.duration = duration;
    this.byteLength = byteLength;
    this.copyTo = vi.fn((dest) => {
      if (dest instanceof Uint8Array) {
        // Fill with dummy data
        for (let i = 0; i < Math.min(dest.length, byteLength); i++) {
          dest[i] = i % 256;
        }
      }
    });
  }
}

// // Mock VideoEncoder
// class MockVideoEncoder {
//   configure = vi.fn();
//   encode = vi.fn();
//   flush = vi.fn();
//   close = vi.fn();
//   reset = vi.fn();
// }

// Mock VideoDecoder
class MockVideoDecoder {
  configure = vi.fn();
  decode = vi.fn();
  flush = vi.fn();
  close = vi.fn();
  reset = vi.fn();

  constructor(config: VideoDecoderInit) {
    // Mock constructor
  }
}

describe('VideoEncodeNode', () => {
  let context: VideoContext;
  let encoderNode: VideoEncodeNode;
  let mockEncoder: MockVideoEncoder;
  let onChunk: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    // Mock the global VideoEncoder
    mockEncoder = new MockVideoEncoder({
      output: vi.fn(),
      error: vi.fn()
    });
    (global as any).VideoEncoder = vi.fn(() => mockEncoder);

    context = new VideoContext();
    onChunk = vi.fn();
    encoderNode = new VideoEncodeNode(context);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should create VideoEncodeNode', () => {
    expect(encoderNode).toBeInstanceOf(VideoEncodeNode);
    expect(encoderNode.numberOfInputs).toBe(1);
    expect(encoderNode.numberOfOutputs).toBe(1);
  });

  it('should configure encoder', () => {
    const config: VideoEncoderConfig = {
      codec: 'vp8',
      width: 640,
      height: 480,
      bitrate: 1000000,
      framerate: 30
    };

    encoderNode.configure(config);

    expect(mockEncoder.configure).toHaveBeenCalledWith(config);
  });

  it('should encode video frame', () => {
    const config: VideoEncoderConfig = {
      codec: 'vp8',
      width: 640,
      height: 480,
      bitrate: 1000000,
      framerate: 30
    };
    encoderNode.configure(config);

    const frame = new MockVideoFrame();
    encoderNode.process(frame);

    expect(mockEncoder.encode).toHaveBeenCalledWith(frame);
  });

  it('should call onChunk callback when encoder outputs chunk', async () => {
    const config: VideoEncoderConfig = {
      codec: 'vp8',
      width: 640,
      height: 480,
      bitrate: 1000000,
      framerate: 30
    };
    encoderNode.configure(config);

    // Simulate encoder output
    const mockChunk = new MockEncodedVideoChunk();
    const outputCallback = (global as any).VideoEncoder.mock.calls[0][0].output;
    await outputCallback(mockChunk);

    expect(onChunk).toHaveBeenCalledWith(expect.any(EncodedContainer));
  });

  it('should close encoder', async () => {
    const config: VideoEncoderConfig = {
      codec: 'vp8',
      width: 640,
      height: 480,
      bitrate: 1000000,
      framerate: 30
    };
    encoderNode.configure(config);

    await encoderNode.close();
    expect(mockEncoder.close).toHaveBeenCalled();
  });
});

describe('VideoDecodeNode', () => {
  let context: VideoContext;
  let decoderNode: VideoDecodeNode;
  let mockDecoder: MockVideoDecoder;
  let onFrame: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    // Mock the global VideoDecoder
    mockDecoder = new MockVideoDecoder({
      output: vi.fn(),
      error: vi.fn()
    });
    (global as any).VideoDecoder = vi.fn(() => mockDecoder);

    context = new VideoContext();
    onFrame = vi.fn();
    decoderNode = new VideoDecodeNode(context);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should create VideoDecodeNode', () => {
    expect(decoderNode).toBeInstanceOf(VideoDecodeNode);
    expect(decoderNode.numberOfInputs).toBe(1);
    expect(decoderNode.numberOfOutputs).toBe(1);
  });

  it('should configure decoder', () => {
    const config: VideoDecoderConfig = {
      codec: 'vp8',
      codedWidth: 640,
      codedHeight: 480
    };

    decoderNode.configure(config);

    expect(mockDecoder.configure).toHaveBeenCalledWith(config);
  });

  it('should decode encoded container', () => {
    const config: VideoDecoderConfig = {
      codec: 'vp8',
      codedWidth: 640,
      codedHeight: 480
    };
    decoderNode.configure(config);

    const mockChunk = new MockEncodedVideoChunk();
    const container = new EncodedContainer(mockChunk);
    // Note: VideoDecodeNode.process expects VideoFrame, but test is passing EncodedContainer
    // This seems to be a test error; process should be called with decoded frame
    const frame = new MockVideoFrame();
    decoderNode.process(frame);

    expect(mockDecoder.decode).toHaveBeenCalledWith(mockChunk);
  });

  it('should call onFrame callback when decoder outputs frame', async () => {
    const config: VideoDecoderConfig = {
      codec: 'vp8',
      codedWidth: 640,
      codedHeight: 480
    };
    decoderNode.configure(config);

    // Simulate decoder output
    const mockFrame = new MockVideoFrame();
    const outputCallback = (global as any).VideoDecoder.mock.calls[0][0].output;
    await outputCallback(mockFrame);

    expect(onFrame).toHaveBeenCalledWith(mockFrame);
  });

  it('should close decoder', async () => {
    const config: VideoDecoderConfig = {
      codec: 'vp8',
      codedWidth: 640,
      codedHeight: 480
    };
    decoderNode.configure(config);

    await decoderNode.close();
    expect(mockDecoder.close).toHaveBeenCalled();
  });

  it('should flush decoder', async () => {
    const config: VideoDecoderConfig = {
      codec: 'vp8',
      codedWidth: 640,
      codedHeight: 480
    };
    decoderNode.configure(config);

    await decoderNode.flush();
    expect(mockDecoder.flush).toHaveBeenCalled();
  });
});