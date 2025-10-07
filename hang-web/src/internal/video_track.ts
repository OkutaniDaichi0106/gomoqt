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
	ContextCancelledError
} from "golikejs/context";
import {
	Mutex,
} from "golikejs/sync";
import { EncodedContainer } from "./container";
import { EncodeErrorCode } from "./error";
import type { TrackEncoder,TrackEncoderInit } from "./track_encoder";
import { cloneChunk } from "./track_encoder";
import type { TrackCache } from "./cache";
import { GroupCache } from "./cache";
import type { TrackDecoder,TrackDecoderInit } from "./track_decoder";
import type { TrackDescriptor, VideoTrackDescriptor } from "../catalog";
import  { VideoTrackSchema,TrackSchema } from "../catalog";

const GOP_DURATION = 2*1000*1000; // 2 seconds

export class VideoTrackEncoder implements TrackEncoder {
    #encoder: VideoEncoder;

	#resolveConfig?: (config: VideoDecoderConfig) => void;

	#source: ReadableStream<VideoFrame>

	#tracks: Set<TrackWriter> = new Set();

	#latestGroup: GroupCache;
	#trackCache?: TrackCache;

	#mutex: Mutex = new Mutex();

	constructor(init: TrackEncoderInit<VideoFrame>) {
		this.#source = init.source;
		const latestSeq = init.startGroupSequence ?? 1n;
		this.#latestGroup = new GroupCache(latestSeq, 0);
		this.#trackCache = init.cache ? new init.cache() : undefined;

		// Initialize encoder settings
		this.#encoder = new VideoEncoder({
			output: async (chunk, metadata) => {
				if (metadata?.decoderConfig) {
					this.#resolveConfig?.(metadata.decoderConfig);
					console.debug("resolved decoder config");
				}

				if (chunk.type === "key") {
					// Close previous group and start a new one
					this.#latestGroup.close();
					const nextSequence = this.#latestGroup.sequence + 1n;
					this.#latestGroup = new GroupCache(nextSequence, chunk.timestamp);

					// Open new groups for all tracks asynchronously
					for (const track of this.#tracks) {
						track.openGroup(this.#latestGroup.sequence).then(
							([group, err]) => {
								if (err) {
									console.error("moq: failed to open group:", err);
									this.#tracks.delete(track);
									track.closeWithError(InternalSubscribeErrorCode, err.message);
									return;
								}

								// Send frames via latest group cache
								this.#latestGroup.flush(group!);
							}
						);
					}
				}

				// Skip encoding if no current groups
				if (this.#tracks.size === 0) {
					return;
				}

				const container = new EncodedContainer(cloneChunk(chunk));

				await this.#latestGroup.append(container);
			},
			error: (error) => {
				console.error("Video encoding error:", error);

				// Close with error to propagate cancellation
				this.close(error);
			}
		});

		// Start reading frames
		queueMicrotask(() => this.#next(this.#source.getReader()));
	}

	#next(reader: ReadableStreamDefaultReader<VideoFrame>): void {
		if (!this.encoding) {
			// No active tracks to encode to
			// Just release the lock and stop reading from the reader
            reader.releaseLock();
			return;
		}

		reader.read().then(async (result) => {

			const { done, value: frame } = result;
			if (done) {
				return;
			}

			const keyFrame = this.#latestGroup.timestamp - GOP_DURATION < frame.timestamp;

			this.#encoder.encode(frame, { keyFrame });
			frame.close();

			// Continue to the next frame
			queueMicrotask(() => this.#next(reader));
		}).catch(err => {
			console.error("video next error", err);
			this.close(err);
		});
	}

    async configure(config: VideoEncoderConfig): Promise<VideoDecoderConfig> {
		await this.#mutex.lock();

		try {
			const decoderConfig = new Promise<VideoDecoderConfig>((resolve) => {
				this.#resolveConfig = resolve;
				console.debug("set resolveConfig");
			});

			this.#encoder.configure(config);

			console.debug("encoder configured");

			// Wait for the decoder config to be resolved
			return await decoderConfig;
		} finally {
			this.#mutex.unlock();
		}
    }

    async encodeTo(ctx: Promise<void>, dest: TrackWriter): Promise<Error | undefined> {
		if (this.#tracks.has(dest)) {
			console.warn("given TrackWriter is already being encoded to");
			return;
		}

		this.#tracks.add(dest);

		await Promise.race([
			dest.context.done(),
			ctx,
		]);

		this.#tracks.delete(dest);

		return dest.context.err() || ContextCancelledError;
    }

	async close(cause?: Error): Promise<void> {
		await Promise.allSettled([
			this.#trackCache?.close(),
			this.#latestGroup.close(),
		]);

		this.#encoder.close();

		this.#tracks.clear();
	}

	get encoding(): boolean {
		return this.#tracks.size > 0;
	}
}

export class VideoTrackDecoder implements TrackDecoder {
    #source?: TrackReader;

	#decoder: VideoDecoder;

    #frameCount: number = 0;
    #dests: Set<WritableStreamDefaultWriter<VideoFrame>> = new Set();

    constructor(init: TrackDecoderInit<VideoFrame>) {
		this.#dests.add(init.destination.getWriter());

        this.#decoder = new VideoDecoder({
            output: async (frame: VideoFrame) => {
				await Promise.allSettled(Array.from(this.#dests, async (dest) => {
					try {
						await dest.ready;
						await dest.write(frame);
					} catch (err) {
						console.error("VideoTrackDecoder: write error:", err);
						this.#dests.delete(dest);
						dest.releaseLock();
					}
				}));
				frame.close();
            },
            error: (error) => {
                console.error("Video decoding error (no auto-close):", error);
            }
        });
    }

    get decoding(): boolean {
		return this.#source !== undefined;
	}	
	
	#next(): void {
		if (this.#source === undefined) {
			return;
		}

		this.#source.acceptGroup(new Promise(() => {})).then(async (result) => {
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
				const [frame, err] = await group!.readFrame();
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

			queueMicrotask(() => this.#next());
		}).catch(err => {
			console.error("Video decode group error:", err);
		});
	}

    async configure(config: VideoDecoderConfig): Promise<void> {
		// Reset source on configure
		if (this.#source !== undefined) {
			console.warn("[VideoTrackDecoder] source already set. cancelling...");
			await this.#source.closeWithError(InternalSubscribeErrorCode, "codec changed");
			this.#source = undefined;
		}

        this.#decoder.configure(config);
    }

    async decodeFrom(ctx: Promise<void>, source: TrackReader): Promise<Error | undefined> {
        if (this.#source !== undefined) {
            console.warn("[VideoDecodeStream] source already set. replacing...");
            await this.#source.closeWithError(InternalSubscribeErrorCode, "source already set");
        }

		this.#source = source;

        queueMicrotask(() => this.#next());

		await Promise.race([
            source.context.done(),
            ctx,
        ]);

        return source.context.err() || ContextCancelledError;
    }

	async close(cause?: Error): Promise<void> {
		this.#decoder.close();

		await Promise.allSettled(Array.from(this.#dests,
			dest => dest.close()
		));

		this.#dests.clear();
		this.#source = undefined;
	}
}

