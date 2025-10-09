// Video node API: VideoContext, VideoNode, VideoSourceNode, VideoEncodeNode, VideoSender
// Based on Web Audio API structure: https://developer.mozilla.org/en-US/docs/Web/API/AudioNode
// https://developer.mozilla.org/en-US/docs/Web/API/AudioContext
// https://developer.mozilla.org/en-US/docs/Web/API/AnalyserNode
import { EncodedContainer,type EncodedChunk } from './container';
import { cloneChunk } from './container';
import { TrackWriter,TrackReader } from "@okutanidaichi/moqt";
import { GroupCache } from ".";
import { readVarint } from "@okutanidaichi/moqt/io";

export type VideoContextState = 'running' | 'suspended' | 'closed';

export class VideoContext {
	readonly frameRate: number;
	readonly destination: VideoDestinationNode;
	#nodes: Set<VideoNode> = new Set();
	#state: VideoContextState = 'running';
	#currentTime: number = 0;

	constructor(options?: { frameRate?: number; canvas?: HTMLCanvasElement;}) {
		this.frameRate = options?.frameRate ?? 30;

		this.destination = new VideoDestinationNode(this, options?.canvas ?? document.createElement('canvas'));
	}

	get state(): VideoContextState {
		return this.#state;
	}

	get currentTime(): number {
		return this.#currentTime;
	}

	_register(node: VideoNode): void {
		this.#nodes.add(node);
	}

	_unregister(node: VideoNode): void {
		this.#nodes.delete(node);
	}

	async resume(): Promise<void> {
		if (this.#state === 'closed') return;
		this.#state = 'running';
	}

	async suspend(): Promise<void> {
		if (this.#state === 'closed') return;
		this.#state = 'suspended';
	}

	async close(): Promise<void> {
		if (this.#state === 'closed') return;
		this.#state = 'closed';

		for (const n of Array.from(this.#nodes)) {
			try {
				n.disconnect();
			} catch (_) {
				/* ignore */
			}
		}

		this.#nodes.clear();
	}
}

export abstract class VideoNode {
	readonly numberOfInputs: number;
	readonly numberOfOutputs: number;
	readonly inputs: Set<VideoNode> = new Set();
	readonly outputs: Set<VideoNode> = new Set();

	constructor(options?: { numberOfInputs?: number; numberOfOutputs?: number }) {
		this.numberOfInputs = options?.numberOfInputs ?? 1;
		this.numberOfOutputs = options?.numberOfOutputs ?? 1;
	}

	connect(destination: VideoNode): VideoNode {
		if (destination === this) return destination;
		this.outputs.add(destination);
		destination.inputs.add(this);
		return destination;
	}

	disconnect(destination?: VideoNode): void {
		if (destination) {
			this.outputs.delete(destination);
			destination.inputs.delete(this);
		} else {
			for (const o of Array.from(this.outputs)) {
				this.outputs.delete(o);
				o.inputs.delete(this);
			}
		}
	}

	abstract process(input?: VideoFrame | EncodedContainer): void;

	dispose(): void {
		this.disconnect();
	}
}

export class VideoSourceNode extends VideoNode {
	#stream: ReadableStream<VideoFrame>;
	#running: boolean = false;
	#reader?: ReadableStreamDefaultReader<VideoFrame>;
	readonly context: VideoContext;

	constructor(context: VideoContext, stream: ReadableStream<VideoFrame>) {
		super({ numberOfInputs: 0, numberOfOutputs: 1 });
		this.context = context;
		this.#stream = stream;
		this.context._register(this);
	}

	process(input: VideoFrame): void {
		// Pass context to next nodes
		for (const out of Array.from(this.outputs)) {
			void out.process(input);
		}
	}

	async start(): Promise<void> {
		if (this.#running) return;
		this.#running = true;

		// Note: In a real implementation, this would process frames from the MediaStreamTrack
		// For now, this is a placeholder
		try {
			this.#reader = this.#stream.getReader();

			// Simulate frame processing
			while (this.#running && this.context.state === 'running') {
				const {done, value: frame} = await this.#reader.read()
				if (done) break;
				void this.process(frame);
			}
		} catch (e) {
			console.error('VideoSourceNode error:', e);
		} finally {
			this.#running = false;
		}
	}

	stop(): void {
		this.#running = false;
	}

	dispose(): void {
		this.stop();
		// Resource cleanup (do not cancel external streams)
		if (this.#reader) {
			this.#reader.releaseLock();
			this.#reader = undefined;
		}
		this.disconnect();
		this.context._unregister(this);
	}
}

export class MediaStreamVideoSourceNode extends VideoSourceNode {
	readonly track: MediaStreamTrack;
	#stream: ReadableStream<VideoFrame>;

	constructor(track: MediaStreamTrack) {
		let stream: ReadableStream<VideoFrame>;
		const context = new VideoContext({ frameRate: track.getSettings()?.frameRate ?? 30 });

		if (self.MediaStreamTrackProcessor) {
			// @ts-ignore
			stream = new self.MediaStreamTrackProcessor({ track }).readable;
		} else {
			console.warn("Using MediaStreamTrackProcessor polyfill; performance might suffer.");

			const settings = track.getSettings();
			if (!settings) {
				throw new Error("track has no settings");
			}

			const video = document.createElement("video") as HTMLVideoElement;
			let last: DOMHighResTimeStamp = performance.now();

			stream = new ReadableStream<VideoFrame>({
				async start(controller) {
					video.srcObject = new MediaStream([track]);
					await Promise.all([
						video!.play(),
						new Promise((r) => {
							video!.onloadedmetadata = r;
						}),
					]);

					last = performance.now();
				},
				async pull(controller) {
					while (performance.now() - last < 1000 / context.frameRate) {
						await new Promise((r) => requestAnimationFrame(r));
					}
					last = performance.now();
					controller.enqueue(new VideoFrame(video!, { timestamp: last * 1000 }));
				},
				cancel() {
					video.srcObject = null;
				},
			});
		}

		super(context, stream);
		this.#stream = stream;
		this.track = track;
	}

	dispose(): void {
		this.stop();
		this.track.stop();
		this.#stream.cancel();
		super.dispose();
	}
}

export class VideoAnalyserNode extends VideoNode {
	// Basic image statistics
	#brightness: number = 0;
	#contrast: number = 0;
	#saturations: number = 0;

	// Color analysis
	#colorHistogram: Uint32Array;
	#dominantColors: Array<{r: number, g: number, b: number, count: number}> = [];

	// Spatial features
	#sharpness: number = 0;
	#edgeStrength: number = 0;
	#textureComplexity: number = 0;

	// Motion features (requires frame comparison)
	#motionMagnitude: number = 0;
	#motionDirection: number = 0;
	#previousFrameData: Uint8Array | null = null;

	// Frequency domain features
	#spatialFrequency: Float32Array;
	#context: VideoContext;

	constructor(context: VideoContext, options?: {
		histogramBins?: number;
		enableMotionDetection?: boolean;
		enableContentDetection?: boolean;
	}) {
		super({ numberOfInputs: 1, numberOfOutputs: 1 });
		this.#context = context;
		this.#context._register(this);

		const histogramBins = options?.histogramBins ?? 256;
		this.#colorHistogram = new Uint32Array(histogramBins * 3); // RGB
		this.#spatialFrequency = new Float32Array(histogramBins);
	}

	// Basic statistics getters
	get brightness(): number { return this.#brightness; }
	get contrast(): number { return this.#contrast; }
	get saturation(): number { return this.#saturations; }

	// Color analysis getters
	getColorHistogram(array: Uint32Array): void {
		const length = Math.min(array.length, this.#colorHistogram.length);
		for (let i = 0; i < length; i++) {
			array[i] = this.#colorHistogram[i] ?? 0;
		}
	}

	getDominantColors(): ReadonlyArray<{r: number, g: number, b: number, count: number}> {
		return [...this.#dominantColors];
	}

	// Spatial features getters
	get sharpness(): number { return this.#sharpness; }
	get edgeStrength(): number { return this.#edgeStrength; }
	get textureComplexity(): number { return this.#textureComplexity; }

	// Motion features getters
	get motionMagnitude(): number { return this.#motionMagnitude; }
	get motionDirection(): number { return this.#motionDirection; }

	// Frequency domain getters
	getSpatialFrequencyData(array: Float32Array): void {
		const length = Math.min(array.length, this.#spatialFrequency.length);
		for (let i = 0; i < length; i++) {
			array[i] = this.#spatialFrequency[i] ?? 0;
		}
	}

	process(input: VideoFrame): void {
		// Analyze the video frame and update internal data
		void this.#analyzeFrame(input);


		// Pass context to connected outputs
		for (const output of Array.from(this.outputs)) {
			try {
				void output.process(input);
			} catch (e) {
				console.error('VideoAnalyserNode process error', e);
			}
		}

		if (input) {
			try {
				input.close();
			} catch (_) {
				/* ignore */
			}
		}
	}

	async #analyzeFrame(frame: VideoFrame): Promise<void> {
		// Extract video features from the frame
		const width = frame.displayWidth;
		const height = frame.displayHeight;

		// Performance optimization: downsample to reasonable size for analysis
		const sampleWidth = Math.min(width, 320);
		const sampleHeight = Math.min(height, 240);
		const sampleStepX = width / sampleWidth;
		const sampleStepY = height / sampleHeight;

		// Get pixel data from VideoFrame (RGBA format)
		const pixelData = new Uint8Array(sampleWidth * sampleHeight * 4);
		try {
			// Create a temporary canvas for downsampling
			const canvas = new OffscreenCanvas(sampleWidth, sampleHeight);
			const ctx = canvas.getContext('2d');
			if (!ctx) return;

			// Draw frame to canvas (this handles format conversion and downsampling)
			ctx.drawImage(frame, 0, 0, sampleWidth, sampleHeight);

			// Get image data
			const imageData = ctx.getImageData(0, 0, sampleWidth, sampleHeight);
			pixelData.set(imageData.data);
		} catch (e) {
			// Fallback: try direct copyTo if canvas fails
			try {
				const fullData = new Uint8Array(width * height * 4);
				frame.copyTo(fullData);

				// Manual downsampling with bounds checking
				for (let y = 0; y < sampleHeight; y++) {
					for (let x = 0; x < sampleWidth; x++) {
						const srcX = Math.floor(x * sampleStepX);
						const srcY = Math.floor(y * sampleStepY);
						const srcIdx = (srcY * width + srcX) * 4;
						const dstIdx = (y * sampleWidth + x) * 4;

						if (srcIdx + 3 < fullData.length && dstIdx + 3 < pixelData.length) {
							pixelData[dstIdx] = fullData[srcIdx]!;
							pixelData[dstIdx + 1] = fullData[srcIdx + 1]!;
							pixelData[dstIdx + 2] = fullData[srcIdx + 2]!;
							pixelData[dstIdx + 3] = fullData[srcIdx + 3]!;
						}
					}
				}
			} catch (fallbackError) {
				console.warn('Failed to analyze frame:', fallbackError);
				return;
			}
		}

		// Calculate all features from pixel data
		this.#calculateBasicStats(pixelData, sampleWidth, sampleHeight);
		this.#calculateColorHistogram(pixelData, sampleWidth, sampleHeight);
		this.#calculateDominantColors();
		this.#calculateSpatialFeatures(pixelData, sampleWidth, sampleHeight);
		this.#calculateMotionFeatures(pixelData, sampleWidth, sampleHeight);
		this.#calculateSpatialFrequency(pixelData, sampleWidth, sampleHeight);
	}

	#calculateBasicStats(pixelData: Uint8Array, width: number, height: number): void {
		let sumBrightness = 0;
		let sumSaturation = 0;
		const pixelCount = width * height;

		for (let i = 0; i < pixelData.length; i += 4) {
			const r = pixelData[i]!;
			const g = pixelData[i + 1]!;
			const b = pixelData[i + 2]!;

			// Brightness (luminance)
			const brightness = (0.299 * r + 0.587 * g + 0.114 * b) / 255;
			sumBrightness += brightness;

			// Saturation
			const max = Math.max(r, g, b);
			const min = Math.min(r, g, b);
			const saturation = max > 0 ? (max - min) / max : 0;
			sumSaturation += saturation;
		}

		this.#brightness = sumBrightness / pixelCount;

		// Contrast (standard deviation of brightness)
		let sumSquaredDiff = 0;
		for (let i = 0; i < pixelData.length; i += 4) {
			const r = pixelData[i]!;
			const g = pixelData[i + 1]!;
			const b = pixelData[i + 2]!;
			const brightness = (0.299 * r + 0.587 * g + 0.114 * b) / 255;
			sumSquaredDiff += Math.pow(brightness - this.#brightness, 2);
		}
		this.#contrast = Math.sqrt(sumSquaredDiff / pixelCount);
		this.#saturations = sumSaturation / pixelCount;
	}

	#calculateColorHistogram(pixelData: Uint8Array, width: number, height: number): void {
		// Reset histogram
		this.#colorHistogram.fill(0);

		for (let i = 0; i < pixelData.length; i += 4) {
			const r = pixelData[i]!;
			const g = pixelData[i + 1]!;
			const b = pixelData[i + 2]!;

			// Update histogram bins
			this.#colorHistogram[r] = (this.#colorHistogram[r] ?? 0) + 1;
			this.#colorHistogram[256 + g] = (this.#colorHistogram[256 + g] ?? 0) + 1;
			this.#colorHistogram[512 + b] = (this.#colorHistogram[512 + b] ?? 0) + 1;
		}
	}

	#calculateDominantColors(): void {
		// Simple approach: find peaks in histogram
		const colorCounts: Map<string, {r: number, g: number, b: number, count: number}> = new Map();

		// Sample colors from histogram peaks
		const peaks = this.#findHistogramPeaks();

		this.#dominantColors = peaks.slice(0, 5).map(peak => ({
			r: peak.r,
			g: peak.g,
			b: peak.b,
			count: peak.count
		}));
	}

	#findHistogramPeaks(): Array<{r: number, g: number, b: number, count: number}> {
		const peaks: Array<{r: number, g: number, b: number, count: number}> = [];

		// Find local maxima in each color channel
		for (let r = 1; r < 255; r++) {
			const countR = this.#colorHistogram[r] ?? 0;
			if (countR > (this.#colorHistogram[r - 1] ?? 0) && countR > (this.#colorHistogram[r + 1] ?? 0)) {
				for (let g = 1; g < 255; g++) {
					const countG = this.#colorHistogram[256 + g] ?? 0;
					if (countG > (this.#colorHistogram[256 + g - 1] ?? 0) && countG > (this.#colorHistogram[256 + g + 1] ?? 0)) {
						for (let b = 1; b < 255; b++) {
							const countB = this.#colorHistogram[512 + b] ?? 0;
							if (countB > (this.#colorHistogram[512 + b - 1] ?? 0) && countB > (this.#colorHistogram[512 + b + 1] ?? 0)) {
								peaks.push({r, g, b, count: countR + countG + countB});
							}
						}
					}
				}
			}
		}

		return peaks.sort((a, b) => b.count - a.count);
	}

	#calculateSpatialFeatures(pixelData: Uint8Array, width: number, height: number): void {
		// Convert to grayscale for spatial analysis
		const gray = new Uint8Array(width * height);
		for (let i = 0; i < pixelData.length; i += 4) {
			const r = pixelData[i]!;
			const g = pixelData[i + 1]!;
			const b = pixelData[i + 2]!;
			gray[i / 4] = Math.round(0.299 * r + 0.587 * g + 0.114 * b);
		}

		// Sharpness (variance of Laplacian)
		this.#sharpness = this.#calculateSharpness(gray, width, height);

		// Edge strength (Sobel operator)
		this.#edgeStrength = this.#calculateEdgeStrength(gray, width, height);

		// Texture complexity (entropy)
		this.#textureComplexity = this.#calculateTextureComplexity(gray);
	}

	#calculateSharpness(gray: Uint8Array, width: number, height: number): number {
		let sum = 0;
		let count = 0;

		// Laplacian kernel
		for (let y = 1; y < height - 1; y++) {
			for (let x = 1; x < width - 1; x++) {
				const idx = y * width + x;
				const laplacian =
					-4 * gray[idx]! +
					gray[(y-1) * width + x]! +
					gray[y * width + (x-1)]! +
					gray[y * width + (x+1)]! +
					gray[(y+1) * width + x]!;

				sum += laplacian * laplacian;
				count++;
			}
		}

		return count > 0 ? Math.sqrt(sum / count) / 128 : 0; // Normalize to 0-1
	}

	#calculateEdgeStrength(gray: Uint8Array, width: number, height: number): number {
		let sum = 0;
		let count = 0;

		// Sobel operator
		for (let y = 1; y < height - 1; y++) {
			for (let x = 1; x < width - 1; x++) {
				const idx = y * width + x;
				const gx =
					-1 * gray[(y-1) * width + (x-1)]! + 1 * gray[(y-1) * width + (x+1)]! +
					-2 * gray[y * width + (x-1)]! + 2 * gray[y * width + (x+1)]! +
					-1 * gray[(y+1) * width + (x-1)]! + 1 * gray[(y+1) * width + (x+1)]!;

				const gy =
					-1 * gray[(y-1) * width + (x-1)]! - 2 * gray[(y-1) * width + x]! - 1 * gray[(y-1) * width + (x+1)]! +
					1 * gray[(y+1) * width + (x-1)]! + 2 * gray[(y+1) * width + x]! + 1 * gray[(y+1) * width + (x+1)]!;

				sum += Math.sqrt(gx * gx + gy * gy);
				count++;
			}
		}

		return count > 0 ? (sum / count) / 1442 : 0; // Normalize to 0-1 (max Sobel response)
	}

	#calculateTextureComplexity(gray: Uint8Array): number {
		// Simple entropy calculation
		const hist = new Uint32Array(256);
		for (let i = 0; i < gray.length; i++) {
			hist[gray[i]!] = (hist[gray[i]!] ?? 0) + 1;
		}

		let entropy = 0;
		const total = gray.length;
		for (let i = 0; i < 256; i++) {
			if (hist[i]! > 0) {
				const p = hist[i]! / total;
				entropy -= p * Math.log2(p);
			}
		}

		return entropy / 8; // Normalize to 0-1 (max entropy for 8-bit)
	}

	#calculateMotionFeatures(pixelData: Uint8Array, width: number, height: number): void {
		if (!this.#previousFrameData) {
			this.#previousFrameData = new Uint8Array(pixelData);
			this.#motionMagnitude = 0;
			this.#motionDirection = 0;
			return;
		}

		let sumDiff = 0;
		let sumX = 0;
		let sumY = 0;
		let count = 0;

		// Calculate motion vectors using block matching
		const blockSize = 8;
		for (let by = 0; by < height - blockSize; by += blockSize) {
			for (let bx = 0; bx < width - blockSize; bx += blockSize) {
				const motion = this.#findBlockMotion(pixelData, this.#previousFrameData, width, height, bx, by, blockSize);
				if (motion) {
					sumDiff += motion.magnitude;
					sumX += motion.dx;
					sumY += motion.dy;
					count++;
				}
			}
		}

		if (count > 0) {
			this.#motionMagnitude = (sumDiff / count) / (255 * 3); // Normalize to 0-1
			this.#motionDirection = Math.atan2(sumY / count, sumX / count);
		}

		// Update previous frame
		this.#previousFrameData.set(pixelData);
	}

	#findBlockMotion(curr: Uint8Array, prev: Uint8Array, width: number, height: number, bx: number, by: number, blockSize: number):
		{magnitude: number, dx: number, dy: number} | null {

		let minSAD = Infinity;
		let bestDx = 0;
		let bestDy = 0;

		// Search in a small window around current position
		const searchRange = 4;
		for (let dy = -searchRange; dy <= searchRange; dy++) {
			for (let dx = -searchRange; dx <= searchRange; dx++) {
				if (bx + dx < 0 || bx + dx + blockSize >= width ||
					by + dy < 0 || by + dy + blockSize >= height) continue;

				let sad = 0;
				for (let y = 0; y < blockSize; y++) {
					for (let x = 0; x < blockSize; x++) {
						const currIdx = ((by + y) * width + (bx + x)) * 4;
						const prevIdx = ((by + dy + y) * width + (bx + dx + x)) * 4;

						const currR = curr[currIdx]!;
						const currG = curr[currIdx + 1]!;
						const currB = curr[currIdx + 2]!;

						const prevR = prev[prevIdx]!;
						const prevG = prev[prevIdx + 1]!;
						const prevB = prev[prevIdx + 2]!;

						sad += Math.abs(currR - prevR) + Math.abs(currG - prevG) + Math.abs(currB - prevB);
					}
				}

				if (sad < minSAD) {
					minSAD = sad;
					bestDx = dx;
					bestDy = dy;
				}
			}
		}

		return {
			magnitude: Math.sqrt(bestDx * bestDx + bestDy * bestDy),
			dx: bestDx,
			dy: bestDy
		};
	}

	#calculateSpatialFrequency(pixelData: Uint8Array, width: number, height: number): void {
		// Convert to grayscale
		const gray = new Uint8Array(width * height);
		for (let i = 0; i < pixelData.length; i += 4) {
			const r = pixelData[i]!;
			const g = pixelData[i + 1]!;
			const b = pixelData[i + 2]!;
			gray[i / 4] = Math.round(0.299 * r + 0.587 * g + 0.114 * b);
		}

		// Simple DFT for horizontal frequencies
		const numFreqs = Math.min(this.#spatialFrequency.length, width / 2);
		for (let freq = 0; freq < numFreqs; freq++) {
			let real = 0;
			let imag = 0;

			for (let x = 0; x < width; x++) {
				// Average across all rows for this frequency
				let rowSum = 0;
				for (let y = 0; y < height; y++) {
					rowSum += gray[y * width + x]!;
				}
				const avg = rowSum / height;

				const angle = -2 * Math.PI * freq * x / width;
				real += avg * Math.cos(angle);
				imag += avg * Math.sin(angle);
			}

			this.#spatialFrequency[freq] = Math.sqrt(real * real + imag * imag) / width;
		}

		// Normalize
		const maxFreq = Math.max(...this.#spatialFrequency);
		if (maxFreq > 0) {
			for (let i = 0; i < this.#spatialFrequency.length; i++) {
				this.#spatialFrequency[i] = this.#spatialFrequency[i]! / maxFreq;
			}
		}
	}

	// Content detection getters (removed - use separate AI nodes)
	// get hasFaces(): boolean { return this.#hasFaces; }
	// get hasText(): boolean { return this.#hasText; }
	// get sceneChange(): boolean { return this.#sceneChange; }
}

export type VideoRenderFunction = (
	frameWidth: number,
	frameHeight: number,
	canvasWidth: number,
	canvasHeight: number
) => { x: number; y: number; width: number; height: number };
export const VideoRenderFunctions = {
	contain: (
		frameWidth: number,
		frameHeight: number,
		canvasWidth: number,
		canvasHeight: number
	): { x: number; y: number; width: number; height: number } => {
		const frameAspect = frameWidth / frameHeight;
		const canvasAspect = canvasWidth / canvasHeight;

		if (frameAspect > canvasAspect) {
			// Frame is wider, fit to width
			const height = canvasWidth / frameAspect;
			const y = (canvasHeight - height) / 2;
			return { x: 0, y, width: canvasWidth, height };
		} else {
			// Frame is taller, fit to height
			const width = canvasHeight * frameAspect;
			const x = (canvasWidth - width) / 2;
			return { x, y: 0, width, height: canvasHeight };
		}
	},

	cover: (
		frameWidth: number,
		frameHeight: number,
		canvasWidth: number,
		canvasHeight: number
	): { x: number; y: number; width: number; height: number } => {
		const frameAspect = frameWidth / frameHeight;
		const canvasAspect = canvasWidth / canvasHeight;

		if (frameAspect > canvasAspect) {
			// Frame is wider, fit to height
			const width = canvasHeight * frameAspect;
			const x = (canvasWidth - width) / 2;
			return { x, y: 0, width, height: canvasHeight };
		} else {
			// Frame is taller, fit to width
			const height = canvasWidth / frameAspect;
			const y = (canvasHeight - height) / 2;
			return { x: 0, y, width: canvasWidth, height };
		}
	},

	fill: (
		frameWidth: number,
		frameHeight: number,
		canvasWidth: number,
		canvasHeight: number
	): { x: number; y: number; width: number; height: number } => {
		// Fill entire canvas, may distort
		return { x: 0, y: 0, width: canvasWidth, height: canvasHeight };
	},

	scaleDown: (
		frameWidth: number,
		frameHeight: number,
		canvasWidth: number,
		canvasHeight: number
	): { x: number; y: number; width: number; height: number } => {
		// Only scale down, never up
		if (frameWidth <= canvasWidth && frameHeight <= canvasHeight) {
			// No scaling needed
			const x = (canvasWidth - frameWidth) / 2;
			const y = (canvasHeight - frameHeight) / 2;
			return { x, y, width: frameWidth, height: frameHeight };
		} else {
			// Scale down using contain logic
			return VideoRenderFunctions.contain(frameWidth, frameHeight, canvasWidth, canvasHeight);
		}
	}
};

export class VideoDestinationNode extends VideoNode {
	resizeCallback: VideoRenderFunction;
	canvas: HTMLCanvasElement;
	#context: VideoContext;
	#animateId?: number;
	delayFunc?: () => Promise<void>;
	#isVisible: boolean = true;

	constructor(
		context: VideoContext,
		canvas: HTMLCanvasElement,
		options?: {
			renderFunction?: VideoRenderFunction;
		}
	) {
		super({ numberOfInputs: 1, numberOfOutputs: 0 });
		this.#context = context;
		this.#context._register(this);
		this.canvas = canvas;

		// Set render function
		this.resizeCallback = options?.renderFunction ?? VideoRenderFunctions.contain;
	}

	process(input: VideoFrame): void {
		if (this.#context.state !== 'running') {
			try {
				input.close();
			} catch (_) {
				/* ignore */
			}
			return;
		}

		// Cancel any previously scheduled frame
		if (this.#animateId) cancelAnimationFrame(this.#animateId);

		// Schedule the next frame
		this.#animateId = requestAnimationFrame(() => this.#renderVideoFrame(input));
	}

	async #renderVideoFrame(frame?: VideoFrame): Promise<void> {
		if (!frame) {
			return;
		}

		// Skip rendering if canvas is not visible
		if (!this.#isVisible) {
			try {
				frame.close();
			} catch (e) {
				console.error('VideoDestinationNode frame close error:', e);
			}
			return;
		}

		// Check if delay function is defined
		if (this.delayFunc) {
			console.log('Rendering delayed');
			try {
				await this.delayFunc();
			} catch (error) {
				console.warn('Error during rendering delay:', error);
			}
		}

		// Calculate rendering dimensions using render function
		const { x, y, width, height } = this.resizeCallback(
			frame.displayWidth, frame.displayHeight, this.canvas.width, this.canvas.height
		);

		// Get 2D context
		const ctx = this.canvas.getContext('2d');
		if (!ctx) return;

		// Clear the canvas
		ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);

		// Draw frame to the provided context
		ctx.drawImage(frame, x, y, width, height);

		// Close the original frame after processing
		try {
			frame.close();
		} catch (e) {
			console.error('[VideoDestinationNode] frame close error:', e);
		}
	}

	setVisible(visible: boolean): void {
		this.#isVisible = visible;
	}

	dispose(): void {
		// Cancel any scheduled animation
		if (this.#animateId) {
			cancelAnimationFrame(this.#animateId);
			this.#animateId = undefined;
		}
		this.disconnect();
		this.#context._unregister(this);
	}
}

export class VideoEncodeNode extends VideoNode {
	#encoder: VideoEncoder;
	#context: VideoContext;

	// #latestGroup: GroupCache = new GroupCache(0n, 0);

	#dests: Set<(frame: EncodedChunk) => Promise<void>> = new Set();

	constructor(context: VideoContext, options?: { startSequence?: bigint; }) {
		super({ numberOfInputs: 1, numberOfOutputs: 1 });
		this.#context = context;
		this.#context._register(this);

		this.#encoder = new VideoEncoder({
			output: async (chunk) => {
				// Pass encoded chunk to all registered destinations
				await Promise.allSettled(Array.from(this.#dests, dest => dest(chunk)));
			},
			error: (e) => {
				console.error('VideoEncoder error:', e);
			},
		});
	}

	configure(config: VideoEncoderConfig): void {
		this.#encoder.configure(config);
	}

	process(input: VideoFrame): void {
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

	dispose(): void {
		this.disconnect();
		this.#context._unregister(this);
	}

	async encodeTo(ctx: Promise<void>, dest: (frame: EncodedChunk) => Promise<void>): Promise<void> {
		this.#dests.add(dest);

		await Promise.allSettled([
			ctx,
		]);

		this.#dests.delete(dest);
	}
}

export class VideoDecodeNode extends VideoNode {
	#decoder: VideoDecoder;
	#context: VideoContext;

	constructor(context: VideoContext) {
		super({ numberOfInputs: 1, numberOfOutputs: 1 });
		this.#context = context;
		this.#context._register(this);

		this.#decoder = new VideoDecoder({
			output: async (frame) => {
				// Pass decoded frame to next nodes
				this.process(frame);
			},
			error: (e) => {
				console.error('VideoDecoder error:', e);
			},
		});
	}

	configure(config: VideoDecoderConfig): void {
		this.#decoder.configure(config);
	}

	async decodeFrom(ctx: Promise<void>, reader: TrackReader): Promise<void> {
		try {
			while (this.#context.state === 'running') {
				const [group, err] = await reader.acceptGroup(ctx);
				if (err) {
					break;
				}

				let isKey = true;
				while (true) {
					const [frame, err] = await group.readFrame();
					if (err) {
						break;
					}

					// Read varint timestamp from the header
					const [timestamp, headerSize] = readVarint(frame.bytes);
					const chunk = new EncodedVideoChunk({
						type: isKey ? 'key' : 'delta',
						timestamp: timestamp,
						data: frame.bytes.subarray(headerSize),
					});

					// Decode the chunk
					this.#decoder.decode(chunk);

					if (isKey) isKey = false;

					// Note: We don't pass the container to next nodes as it's consumed
					// The decoded frame is passed in the decoder's output callback
				}
			}
		} catch (e) {
			console.error('decodeFrom error:', e);
		}
	}

	process(input: VideoFrame): void {
		// Pass frame to next nodes
		for (const out of Array.from(this.outputs)) {
			try {
				void out.process(input);
			} catch (e) {
				console.error('VideoDecodeNode process error:', e);
			}
		}

		// Close the frame after processing
		try {
			input.close();
		} catch (_) {
			/* ignore */
		}
	}

	async flush(): Promise<void> {
		try {
			await this.#decoder?.flush();
		} catch (e) {
			console.error('VideoDecoder flush error:', e);
		}
	}

	async close(): Promise<void> {
		try {
			await this.#decoder?.flush();
			this.#decoder?.close();
		} catch (_) {
			/* ignore */
		}
	}

	dispose(): void {
		this.disconnect();
		this.#context._unregister(this);
	}
}

export class VideoObserveNode extends VideoNode {
	#observer?: IntersectionObserver;
	#isVisible: boolean = true;

	constructor(context: VideoContext, options?: { threshold?: number; enableBackground?: boolean }) {
		super({ numberOfInputs: 1, numberOfOutputs: 1 });
		const threshold = options?.threshold ?? 0.01;
		const enableBackground = options?.enableBackground ?? false;
		if (!enableBackground) {
			this.#observer = new IntersectionObserver(
				(entries) => {
					const entry = entries[0];
					if (entry) {
						this.#isVisible = entry.isIntersecting;
					}
				},
				{ threshold }
			);
			// this.#observer.observe(context.destination.canvas);
		} else {
			this.#isVisible = true;
		}
		context._register(this);
	}

	observe(element: Element): void {
		this.#observer?.observe(element);
	}

	get isVisible(): boolean {
		return this.#isVisible;
	}

	process(input: VideoFrame): void {
		// Only pass to next nodes if visible
		if (this.#isVisible) {
			for (const out of Array.from(this.outputs)) {
				try {
					void out.process(input);
				} catch (e) {
					console.error('VideoObserveNode process error:', e);
				}
			}
		}
	}

	dispose(): void {
		this.#observer?.disconnect();
		this.disconnect();
	}
}