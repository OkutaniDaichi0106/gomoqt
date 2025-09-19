import {
	type TrackWriter,
	type TrackReader,
	type GroupWriter,
	type GroupReader,
	type Frame,
	type GroupSequence,
	type SubscribeErrorCode,
	PublishAbortedErrorCode,
	InternalGroupErrorCode,
	InternalSubscribeErrorCode,
} from "@okutanidaichi/moqt";
import { readVarint } from "@okutanidaichi/moqt/io";
import {
	type Context,
	type CancelCauseFunc,
	withCancelCause,
	withPromise,
	background,
ContextCancelledError
} from "@okutanidaichi/moqt/internal";
import { EncodedContainer } from "./container";
import { EncodeErrorCode } from "./error";
import { isFirefox } from "./browser";
import type { TrackEncoder,TrackEncoderInit } from "./track_encoder";
import type { TrackCache } from "./cache";
import type { TrackDecoder } from ".";
import { throwStatement } from "@babel/types";
import { th,tr } from "zod/v4/locales";

const GOP_DURATION = 2*1000*1000; // 2 seconds

export class VideoEncodeStream implements TrackEncoder<VideoFrame> {
    #encoder: VideoEncoder;

    #decoderConfig: VideoDecoderConfig | undefined = undefined; // TODO: Expose this if needed

	#source: ReadableStreamDefaultReader<VideoFrame>

	#latestGroupSequence: GroupSequence;
	#latestGroupTimestamp: number = 0;
	#currentGroups: Map<TrackWriter, GroupWriter | undefined> = new Map();

	#previewer?: WritableStreamDefaultWriter<VideoFrame>;

	cache?: TrackCache;

	#ctx: Context;
	#cancelCtx: CancelCauseFunc;

	constructor(init: TrackEncoderInit<VideoFrame>) {
		this.#source = init.source;
		this.#latestGroupSequence = init.startGroupSequence ?? 0n;
		this.cache = init.cache ? new init.cache() : undefined;

		[this.#ctx, this.#cancelCtx] = withCancelCause(background());

		// Initialize encoder settings
		this.#encoder = new VideoEncoder({
			output: async (chunk, metadata) => {
				if (metadata?.decoderConfig) {
					this.#decoderConfig = metadata.decoderConfig;
				}

				const isKey = chunk.type === "key";
				if (isKey && chunk.timestamp - this.#latestGroupTimestamp > GOP_DURATION) {
					this.#latestGroupSequence += 1n;
					this.#latestGroupTimestamp = chunk.timestamp;
				}

				// Skip encoding if no current groups
				if (this.#currentGroups.size === 0) {
					return;
				}

				const container = new EncodedContainer(chunk);

				const promises: Promise<void>[] = [];
				for (const [writer, group] of this.#currentGroups) {
					if (!group) {
						// Open a new group if none exists
						const p = writer.openGroup(this.#latestGroupSequence).then(async ([g, err]) => {
							if (err) throw err;
							this.#currentGroups.set(writer, g);

							err = await g?.writeFrame(container);
							if (err) console.error("Error writing frame:", err);

							return;
						});
						// Add to promises to ensure order
						promises.push(p);
					} else if (group.groupSequence === this.#latestGroupSequence) {
						// Write to existing group
						const p = group.writeFrame(container).then((err) => {
							if (err) console.error("Error writing frame:", err);
						});
						// Add to promises to ensure order
						promises.push(p);
					} else if (this.#latestGroupSequence > group.groupSequence) {
						// Remove old group first
						this.#currentGroups.set(writer, undefined);

						this.cache?.flush(group).then(() => {
							group.close();
						}).catch((err) => {
							group.cancel(InternalGroupErrorCode, err.message);
						});
					}
				}

				this.cache?.append(this.#latestGroupSequence, container);

				await Promise.all(promises);
			},
			error: (error) => {
				console.error("Video encoding error:", error);

				// Close with error to propagate cancellation
				this.closeWithError(EncodeErrorCode, error.message);
			}
		});
	}

	#next(): void {
		if (this.#ctx.err() !== undefined) {
			return;
		}

		Promise.race([
			this.#source.read(),
			this.#ctx.done(),
		]).then(async (result) => {
			if (result === undefined) {
				// Context was cancelled
				this.#previewer?.abort(this.#ctx.err() || ContextCancelledError);
				this.#previewer = undefined;
				return;
			}

			const { done, value: frame } = result;
			if (done) {
				this.#previewer?.close();
				this.#previewer = undefined;
				return;
			}

			await this.#previewer?.write(frame).catch((err) => {
                this.#previewer?.abort(err);
				this.#previewer = undefined;
			});

			const keyFrame = this.#latestGroupTimestamp - GOP_DURATION < frame.timestamp;
			if (keyFrame) this.#latestGroupTimestamp = frame.timestamp;

			this.#encoder.encode(frame, { keyFrame });
			frame.close();

			// Continue to the next frame
			if (!this.#ctx.err()) {
				queueMicrotask(() => this.#next());
			}
		}).catch(err => {
			console.error("video next error", err);
			this.closeWithError(EncodeErrorCode, err.message ?? String(err));
		});
	}

    configure(config: VideoEncoderConfig) {
        this.#encoder.configure(config);
    }

    async encodeTo(dest: TrackWriter): Promise<Error | undefined> {
		if (this.#ctx.err() !== undefined) {
			return this.#ctx.err();
		}
		if (this.#currentGroups.has(dest)) {
			console.warn("given TrackWriter is already being encoded to");
			return;
		}

		this.#currentGroups.set(dest, undefined);

		// Start encoding if this is the first destination
		if (this.#currentGroups.size === 1) {
			queueMicrotask(() => this.#next());
		}

		await Promise.race([
			dest.context.done(),
			this.#ctx.done(),
		]);

		return this.#ctx.err() || dest.context.err() || ContextCancelledError;
    }

	preview(dest?: WritableStreamDefaultWriter<VideoFrame>): void {
		if (this.#ctx.err() !== undefined) {
			return;
		}

		this.#previewer = dest;
	}

	// close() and closeWithError() do not close the underlying source,
	// Callers should close the source to release resources.
	close(): void {
		if (this.#ctx.err() !== undefined) return;

		const cause = new Error("video stream encoder closed");
		this.#cancelCtx(cause);

		this.#encoder.close();

		for (const [tw] of this.#currentGroups) {
			tw.close();
		}
		this.#currentGroups.clear();

		this.cache?.close();
	}

	closeWithError(code: SubscribeErrorCode, reason: string): void {
		if (this.#ctx.err() !== undefined) {
			return;
		}

		const cause = new Error(`video stream encoder closed with error: [${code}] ${reason}`);
		this.#cancelCtx(cause);

		this.#encoder.close();

		for (const [tw] of this.#currentGroups) {
			tw.closeWithError(code, reason);
		}

		this.#currentGroups.clear();

		this.cache?.closeWithError(reason);
	}
}

export class VideoDecodeStream implements TrackDecoder<VideoFrame> {
    #source: TrackReader;

	#decoder: VideoDecoder;

    #frameCount: number = 0;
    #dests: Set<WritableStreamDefaultWriter<VideoFrame>> = new Set();

	#ctx: Context;
	#cancelCtx: CancelCauseFunc;

    constructor(source: TrackReader) {
        this.#source = source;
		[this.#ctx, this.#cancelCtx] = withCancelCause(background());

        this.#decoder = new VideoDecoder({
            output: (frame: VideoFrame) => {
                for (const dest of this.#dests) {
                    dest.write(frame);
                }
            },
            error: (error) => {
                console.error("Video decoding error (no auto-close):", error);
            }
        });

        queueMicrotask(() => this.#next());
    }

	#next(): void {
		if (this.#ctx.err() !== undefined) {
			return;
		}

		this.#source.acceptGroup(this.#ctx.done()).then(async (result) => {
			if (result === undefined) {
				// Context was cancelled
				//
				return;
			}

			const [group, err] = result;
			this.#frameCount = 0;
			if (err) {
				console.error("Error accepting group:", err);
				return;
			}

			while (true) {
				const result = await Promise.race([
					group!.readFrame(),
					this.#ctx.done(),
				]);
				if (result === undefined) {
					// Context was cancelled
					return;
				}

				const [frame, err] = result;
				if (err || !frame) {
					console.error("Error reading frame:", err);
					break;
				}

				this.#frameCount++;

				const [timestamp, headerSize] = readVarint(frame.bytes);

				const chunk = new EncodedVideoChunk({
					type: this.#frameCount < 2 ? "key" : "delta",
					timestamp,
					data: frame.bytes.subarray(headerSize),
				});

				this.#decoder.decode(chunk);
			}

			if (this.#ctx.err() === undefined) {
				queueMicrotask(() => this.#next());
			}
		}).catch(err => {
			if (this.#ctx.err() !== undefined) {
				return;
			}
			console.error("Video decode group error:", err);
		});
	}

    configure(config: VideoDecoderConfig) {
        this.#decoder.configure(config);
    }

    async decodeTo(dest: WritableStreamDefaultWriter<VideoFrame>): Promise<Error | undefined> {
		let err = this.#ctx.err();
		if (err) {
			return err;
		}

		if (this.#dests.has(dest)) {
			console.warn("given WritableStreamDefaultWriter is already being decoded to");
			return;
		}

		this.#dests.add(dest);

		// Start encoding if this is the first destination
		if (this.#dests.size === 1) {
			queueMicrotask(() => this.#next());
		}

		err = await Promise.race([
            dest.closed.catch(e => e),
            this.#ctx.done(),
        ]);

        return this.#ctx.err() || err || ContextCancelledError;
    }

	// close() and closeWithError() do not close the underlying source,
	// Callers should close the source to release resources.
	close(): void {
		if (this.#ctx.err() !== undefined) {
			return;
		}

		const cause = new Error("video decoder closed");
		this.#cancelCtx(cause);

		this.#decoder.close();

		for (const dest of this.#dests) {
			dest.close();
		}
		this.#dests.clear();
	}

	closeWithError(code: SubscribeErrorCode, reason: string): void {
		if (this.#ctx.err() !== undefined) {
			return;
		}

		const cause = new Error(`video decoder closed: [${code}] ${reason}`);
		this.#cancelCtx(cause);

		this.#decoder.close();

		for (const dest of this.#dests) {
			dest.abort(cause);
		}
		this.#dests.clear();
	}
}


// Based on: https://github.com/kixelated/moq/blob/main/js/hang/src/publish/video/index.ts
export async function encoderConfig(width: number, height: number, frameRate: number): Promise<VideoEncoderConfig> {
    // TARGET BITRATE CALCULATION (h264)
	// 480p@30 = 1.0mbps
	// 480p@60 = 1.5mbps
	// 720p@30 = 2.5mbps
	// 720p@60 = 3.5mpbs
	// 1080p@30 = 4.5mbps
	// 1080p@60 = 6.0mbps
	const pixels = width * height;

	// 30fps is the baseline, applying a multiplier for higher framerates.
	// Framerate does not cause a multiplicative increase in bitrate because of delta encoding.
	// TODO Make this better.
	const framerateFactor = 30.0 + (frameRate - 30) / 2;
	const bitrate = Math.round(pixels * 0.07 * framerateFactor);

	// ACTUAL BITRATE CALCULATION
	// 480p@30 = 409920 * 30 * 0.07 = 0.9 Mb/s
	// 480p@60 = 409920 * 45 * 0.07 = 1.3 Mb/s
	// 720p@30 = 921600 * 30 * 0.07 = 1.9 Mb/s
	// 720p@60 = 921600 * 45 * 0.07 = 2.9 Mb/s
	// 1080p@30 = 2073600 * 30 * 0.07 = 4.4 Mb/s
	// 1080p@60 = 2073600 * 45 * 0.07 = 6.5 Mb/s

    // A list of codecs to try, in order of preference.
	const HARDWARE_CODECS = [
		// VP9
		// More likely to have hardware decoding, but hardware encoding is less likely.
		"vp09.00.10.08",
		"vp09", // Browser's choice

        // H.264
		// Almost always has hardware encoding and decoding.
		"avc1.640028",
		"avc1.4D401F",
		"avc1.42E01E",
		"avc1",

		// AV1
		// One day will get moved higher up the list, but hardware decoding is rare.
		"av01.0.08M.08",
		"av01",

		// HEVC (aka h.265)
		// More likely to have hardware encoding, but less likely to be supported (licensing issues).
		// Unfortunately, Firefox doesn't support decoding so it's down here at the bottom.
		"hev1.1.6.L93.B0",
		"hev1", // Browser's choice

		// VP8
		// A terrible codec but it's easy.
		"vp8",
	];

	const SOFTWARE_CODECS = [
		// Now try software encoding for simple enough codecs.
		// H.264
		"avc1.640028", // High
		"avc1.4D401F", // Main
		"avc1.42E01E", // Baseline
		"avc1",

        // VP8
		"vp8",

		// VP9
		// It's a bit more expensive to encode so we shy away from it.
		"vp09.00.10.08",
		"vp09",

		// HEVC (aka h.265)
		// This likely won't work because of licensing issues.
		"hev1.1.6.L93.B0",
		"hev1", // Browser's choice

		// AV1
		// Super expensive to encode so it's our last choice.
		"av01.0.08M.08",
		"av01",
	];

	const baseConfig: VideoEncoderConfig = {
		codec: "none",
		width,
		height,
		bitrate,
		latencyMode: "realtime",
		framerate: frameRate,
	};

	// Try hardware encoding first.
	// We can't reliably detect hardware encoding on Firefox: https://github.com/w3c/webcodecs/issues/896
	if (!isFirefox) {
		for (const codec of HARDWARE_CODECS) {
			const config = upgradeEncoderConfig(baseConfig, codec, bitrate, true);
			const { supported, config: hardwareConfig } = await VideoEncoder.isConfigSupported(config);
			if (supported && hardwareConfig) {
				console.debug("using hardware encoding: ", hardwareConfig);
				return hardwareConfig;
			}
		}
	} else {
		console.warn("Cannot detect hardware encoding on Firefox.");
	}

    // Try software encoding.
	for (const codec of SOFTWARE_CODECS) {
		const config = upgradeEncoderConfig(baseConfig, codec, bitrate, false);
		const { supported, config: softwareConfig } = await VideoEncoder.isConfigSupported(config);
		if (supported && softwareConfig) {
			console.debug("using software encoding: ", softwareConfig);
			return softwareConfig;
		}
	}

	throw new Error("no supported codec");
}

export function upgradeEncoderConfig(base: VideoEncoderConfig, codec: string, bitrate: number, hardware: boolean): VideoEncoderConfig {
    const config: VideoEncoderConfig = {
		...base,
		codec,
		hardwareAcceleration: hardware ? "prefer-hardware" : undefined,
	};

	// We scale the bitrate for more efficient codecs.
	// TODO This shouldn't be linear, as the efficiency is very similar at low bitrates.
	if (config.codec.startsWith("avc1")) {
		// Annex-B allows changing the resolution without nessisarily updating the catalog (description).
		config.avc = { format: "annexb" };
	} else if (config.codec.startsWith("hev1")) {
		// Annex-B allows changing the resolution without nessisarily updating the catalog (description).
		// @ts-expect-error Typescript needs to be updated.
		config.hevc = { format: "annexb" };
	} else if (config.codec.startsWith("vp09")) {
		config.bitrate = bitrate * 0.8;
	} else if (config.codec.startsWith("av01")) {
		config.bitrate = bitrate * 0.6;
	} else if (config.codec === "vp8") {
		// Worse than H.264 but it's a backup plan.
		config.bitrate = bitrate * 1.1;
	}

	return config;
}
