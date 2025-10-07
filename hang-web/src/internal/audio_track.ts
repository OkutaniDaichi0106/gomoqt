import type {
    TrackWriter,
    TrackReader,
    SubscribeErrorCode,
} from "@okutanidaichi/moqt";
import {
    PublishAbortedErrorCode,
    InternalSubscribeErrorCode,
} from "@okutanidaichi/moqt";
import { readVarint } from "@okutanidaichi/moqt/io";
import { ContextCancelledError } from "golikejs/context";
import { Mutex } from "golikejs/sync";
import { EncodedContainer } from "./container";
import { EncodeErrorCode } from "./error";
import type { TrackEncoder, TrackEncoderInit } from "./track_encoder";
import { cloneChunk } from "./track_encoder";
import type { TrackCache } from "./cache";
import { GroupCache } from "./cache";
import type { TrackDecoder,TrackDecoderInit } from "./track_decoder";

// Group rollover max latency (microseconds in Video; here we use ms timestamp from AudioEncoder chunk)
const MAX_AUDIO_LATENCY = 100; // 100ms

export class AudioTrackEncoder implements TrackEncoder {
    #encoder: AudioEncoder;
    #source: ReadableStream<AudioData>;

	#resolveConfig?: (config: AudioDecoderConfig) => void;

	#tracks: Set<TrackWriter> = new Set();

	#latestGroup: GroupCache;
	#trackCache?: TrackCache;

	#mutex: Mutex = new Mutex();

    constructor(init: TrackEncoderInit<AudioData>) {
        this.#source = init.source;
		const latestSeq = init.startGroupSequence ?? 1n;
		this.#latestGroup = new GroupCache(latestSeq, 0);
		this.#trackCache = init.cache ? new init.cache() : undefined;

        this.#encoder = new AudioEncoder({
            output: async (chunk, metadata) => {
                if (metadata?.decoderConfig) {
					console.debug("resolved decoder config");
					this.#resolveConfig?.(metadata.decoderConfig);
					this.#resolveConfig = undefined; // Clear after use
                }

                // (Original code enforced key-only) Keep for safety
                if (chunk.type !== "key") {
                    console.warn("Ignoring non-key audio chunk");
                    return;
                }

				// For audio, we create new groups based on time thresholds
				if (chunk.timestamp - this.#latestGroup.timestamp > MAX_AUDIO_LATENCY) {
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

				// Skip encoding if no current tracks
				if (this.#tracks.size === 0) {
					return;
				}

				const container = new EncodedContainer(cloneChunk(chunk));

				await this.#latestGroup.append(container);
            },
            error: (error) => {
                console.error("Audio encoding error:", error);
                this.close(error);
            }
        });

		// Start encoding loop
		queueMicrotask(() => this.#next(this.#source.getReader()));
    }

	get encoding(): boolean {
		return this.#tracks.size > 0;
	}

    #next(reader: ReadableStreamDefaultReader<AudioData>): void {
        reader.read().then(async (result) => {

            const { done, value: frame } = result;
            if (done) {
                return;
            }

            this.#encoder.encode(frame);

            frame.close();

            // Schedule next read
            queueMicrotask(() => this.#next(reader));
        }).catch((err) => {
            console.error("audio next error", err);
            // this.#previewer?.abort(err);
            this.close(err);
        });
    }

    async configure(config: AudioEncoderConfig): Promise<AudioDecoderConfig> {
		await this.#mutex.lock();

		try {
			const decoderConfig = new Promise<AudioDecoderConfig>((resolve) => {
            	this.#resolveConfig = resolve;
        	});

			this.#encoder.configure(config);

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

		return dest.context.err() || ContextCancelledError;
    }

	async close(cause?: Error): Promise<void> {
		this.#trackCache?.close();

		this.#encoder.close();

		await this.#latestGroup.close();

		// Close all tracks
		await Promise.allSettled(
			Array.from(this.#tracks, dest => dest.close())
		);

		this.#tracks.clear();
	}
}

export class AudioTrackDecoder implements TrackDecoder {
    #decoder: AudioDecoder;
    #source?: TrackReader;
    #frameCount = 0;
    #dests: Set<WritableStreamDefaultWriter<AudioData>> = new Set();

    constructor(init: TrackDecoderInit<AudioData>) {
		this.#dests.add(init.destination.getWriter());

        this.#decoder = new AudioDecoder({
            output: async (frame: AudioData) => {
                await Promise.allSettled(Array.from(this.#dests, async (dest) => {
					try {
						await dest.ready;
						await dest.write(frame);
					} catch (e) {
						// if (!this.#ctx.err()) {
						// 	return;
						// }
						console.error("Audio write error, closing writer:", e);
						this.#dests.delete(dest);
						dest.releaseLock();
					}
				}));
            },
            error: (error) => {
                console.error("Audio decoding error (no auto-close):", error);
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

				const chunk = new EncodedAudioChunk({
					type: this.#frameCount < 2 ? "key" : "delta",
					timestamp,
					data: frame.bytes.subarray(headerSize),
				});

				this.#decoder.decode(chunk);
			}

			queueMicrotask(() => this.#next());
		}).catch(err => {
			console.error("Audio decode group error:", err);
		});
    }

    configure(config: AudioDecoderConfig) {
        this.#decoder.configure(config);
    }

	async decodeFrom(ctx: Promise<void>, source: TrackReader): Promise<Error | undefined> {
        if (this.#source !== undefined) {
            console.warn("[AudioDecodeStream] source is already set, replacing");
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

		for (const writer of this.#dests) {
			writer.releaseLock();
		}

		this.#dests.clear();
	}
}
