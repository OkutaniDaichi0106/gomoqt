import {
	PublishAbortedErrorCode,
	InternalGroupErrorCode,
	InternalSubscribeErrorCode,
} from "@okutanidaichi/moqt";
import type {
    TrackWriter,
	TrackReader,
	GroupWriter,
	GroupReader,
	Frame,
	GroupSequence,
	SubscribeErrorCode,
} from "@okutanidaichi/moqt";
import { readVarint } from "@okutanidaichi/moqt/io";
import {
	withCancelCause,
	background,
	ContextCancelledError
} from "golikejs/context";
import type {
	Context,
	CancelCauseFunc,
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
import type { TrackDescriptor, VideoTrackDescriptor,CatalogRoot,TrackPatch } from "../catalog";
import  { VideoTrackSchema,TrackSchema,isEqualTrack,DEFAULT_CATALOG_VERSION,RootSchema,TrackPatchSchema } from "../catalog";
import { JsonEncoder,JsonDecoder,EncodedJsonChunk } from "./json";
import type { JsonEncoderConfig, JsonDecoderConfig } from "./json";
import type { JsonPatch,JsonValue } from "./json_patch";
import { JsonPatchSchema } from "./json_patch";
import { de } from "zod/v4/locales";

const GOP_DURATION = 2*1000*1000; // 2 seconds

export interface CatalogTrackEncoderInit {
	version?: string;
	description?: string;
}

export interface CatalogTrackDecoderInit {
	version?: string;
	description?: string;
}

/**
 * CatalogTrackEncoder manages the catalog track for broadcasting track metadata.
 * 
 * API Design:
 * - setTrack(): Adds or updates tracks in the root catalog and creates patches
 * - removeTrack(): Removes tracks from the root catalog and creates patches  
 * - sync(): Flushes pending patches to all connected track writers
 * 
 * This encoder only handles root catalog editing operations. The corresponding
 * CatalogTrackDecoder.nextTrack() will only be triggered by "add" operations,
 * maintaining the design where nextTrack exposes new tracks only.
 */
export class CatalogTrackEncoder implements TrackEncoder {
    #root: CatalogRoot;
    #patches: TrackPatch[] = [];

    #encoder: JsonEncoder;

	#resolveConfig?: (config: JsonDecoderConfig) => void;

	#tracks: Map<TrackWriter, GroupWriter | undefined> = new Map();

	#mutex: Mutex = new Mutex();

	constructor(init: CatalogTrackEncoderInit) {
		this.#root = {
			version: init.version ?? DEFAULT_CATALOG_VERSION,
			description: init.description ?? "",
			tracks: new Map(),
		}

		// Initialize encoder settings
		this.#encoder = new JsonEncoder({
			output: async (chunk, metadata) => {
				if (metadata?.decoderConfig) {
					this.#resolveConfig?.(metadata.decoderConfig);
				}

				if (chunk.type === "key") {
					// Close previous group and start a new one

					// Open new groups for all tracks asynchronously
                    await Promise.allSettled(
                        Array.from(this.#tracks, async ([track, prevGroup]) => {
                            await prevGroup?.close();

                            const nextSequence = prevGroup ? prevGroup.groupSequence + 1n : 1n;

                            const [nextGroup, err] = await track.openGroup(nextSequence);
                            if (err) {
                                console.error("moq: failed to open group:", err);
                                this.#tracks.delete(track);
                                await track.closeWithError(InternalSubscribeErrorCode, err.message);
                                return;
                            }

                            this.#tracks.set(track, prevGroup!);
                        })
                    );
				}

				// Skip encoding if no current groups
				if (this.#tracks.size === 0) {
					return;
				}

				const container = new EncodedContainer(cloneChunk(chunk));

				// Write to all current groups asynchronously
				await Promise.allSettled(
					Array.from(this.#tracks, async ([track, group]) => {
						const err = await group?.writeFrame(container);
						if (err) {
							console.error("moq: failed to write frame:", err);
							await group?.cancel(InternalGroupErrorCode, err.message);
							this.#tracks.set(track, undefined);
							return;
						}
					})
				);
			},
			error: (error) => {
				console.error("Video encoding error:", error);

				// Close with error to propagate cancellation
				this.close(error);
			}
		});
	}

    async configure(config: JsonEncoderConfig): Promise<JsonDecoderConfig> {
		await this.#mutex.lock();

		try {
			const decoderConfig = new Promise<JsonDecoderConfig>((resolve) => {
				this.#resolveConfig = resolve;
			});

			this.#encoder.configure(config);

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

		this.#tracks.set(dest, undefined);

		const cause = await Promise.race([
			dest.context.done().then(()=>{return dest.context.err();}),
			ctx.then(()=>{return ContextCancelledError;}),
		]);

		this.#tracks.delete(dest);

		return cause;
    }

    /**
     * Add or update a track in the catalog.
     * This modifies the root catalog and creates appropriate patches for streaming.
     * 
     * @param track - The track descriptor to add or update
     */
    setTrack(track: TrackDescriptor): void {
		const old = this.#root.tracks.get(track.name);
		if (!old) {
			this.#patches.push({
				op: "add",
				path: `/tracks/${track.name}`,
				value: track
			});
		} else if (!isEqualTrack(old, track)) {
			this.#patches.push({
				op: "replace",
				path: `/tracks/${track.name}`,
				value: track
			});
		}

		this.#root.tracks.set(track.name, track);
    }

    /**
     * Remove a track from the catalog.
     * This operation only edits the root catalog and creates remove patches for streaming.
     * Unlike setTrack operations that may trigger nextTrack() for new tracks, 
     * remove operations do not trigger any notifications in the decoder.
     * 
     * @param name - The name of the track to remove
     */
    removeTrack(name: string): void {
		this.#patches.push({
			op: "remove",
			path: `/tracks/${name}`
		});

		this.#root.tracks.delete(name);
	}

	hasTrack(name: string): boolean {
		return this.#root.tracks.has(name);
	}

    sync(): void {
        if (this.#patches.length === 0) {
            return;
        }
        this.#encoder.encode(this.#patches);
        this.#patches = [];
    }

	async root(): Promise<CatalogRoot> {
		return this.#root;
	}

	async close(cause?: Error): Promise<void> {
		await Promise.allSettled(Array.from(this.#tracks, async ([track, group]) => {
			await group?.close();
		}));

		this.#encoder.close();

		this.#tracks.clear();
	}

	get encoding(): boolean {
		return this.#tracks.size > 0;
	}
}

export interface CatalogTrackDecoderInit {
	version?: string;
}

/**
 * CatalogTrackDecoder receives and processes catalog track updates.
 * 
 * API Design:
 * - nextTrack(): Only resolves when NEW tracks are added (via "add" patches)
 * - root(): Provides access to the complete catalog state
 * - hasTrack(): Checks if a track exists in the current catalog
 * 
 * This decoder implements the design where nextTrack() exposes only new tracks,
 * while remove/update operations only modify the root catalog without triggering
 * nextTrack() notifications. This maintains consistency with the encoder's 
 * root catalog editing approach.
 */
export class CatalogTrackDecoder implements TrackDecoder {
    #source?: TrackReader;

	#decoder: JsonDecoder;

    #frameCount: number = 0;

	#mutex: Mutex = new Mutex();

	readonly version: string;
	readonly description?: string;

	#root: Promise<CatalogRoot>;
	#resolveRoot?: (root: CatalogRoot) => void;
	#currentRoot?: CatalogRoot;
	#newTracks: Set<string> = new Set();
	#resolveNewTrack?: (track: TrackDescriptor) => void;

    constructor(init: CatalogTrackDecoderInit) {
		this.version = init.version ?? DEFAULT_CATALOG_VERSION;

		this.#root = new Promise<CatalogRoot>((resolve) => {
			this.#resolveRoot = resolve;
		});

        this.#decoder = new JsonDecoder({
            output: async (chunk: JsonValue | JsonPatch) => {
				await this.#mutex.lock();

				try {
					// Check if chunk is a JsonPatch (array of patch objects)
					if (Array.isArray(chunk)) {
						// Patch
						if (!this.#currentRoot) {
							console.warn("Received patch before full catalog");
							throw new Error("Received patch before full catalog");
						}

						let parsed: TrackPatch;
						for (const op of chunk) {
							parsed = TrackPatchSchema.parse(op)
							const trackName = parsed.path.split("/").pop()!;
							
							if (parsed.op === "add") {
								this.#currentRoot.tracks.set(trackName, parsed.value);
								// Notify new track - only "add" operations trigger nextTrack()
								this.#newTracks.add(trackName);
								if (this.#resolveNewTrack) {
									this.#resolveNewTrack(parsed.value);
									this.#resolveNewTrack = undefined;
								}
							} else if (parsed.op === "remove") {
								this.#currentRoot.tracks.delete(trackName);
								// Remove operations only update the root catalog, don't trigger nextTrack()
							} else if (parsed.op === "replace") {
								this.#currentRoot.tracks.set(trackName, parsed.value);
								// Replace operations only update the root catalog, don't trigger nextTrack()
								// (This maintains consistency with the design where nextTrack only exposes new tracks)
							}
						}
					} else {
						// Full catalog
						const root = RootSchema.parse(chunk);
						if (root.version !== this.version) {
							console.warn(`Unsupported catalog version: ${root.version}`);
							throw new Error(`Unsupported catalog version: ${root.version}`);
						}

						this.#currentRoot = root;
						this.#resolveRoot?.(root);
						this.#resolveRoot = undefined;
					}
				} catch (error) {
					console.error("Error writing chunk to destination:", error);
					this.close(error instanceof Error ? error : new Error(String(error)));
				} finally {
					this.#mutex.unlock();
				}
            },
            error: (error) => {
                console.error("Video decoding error (no auto-close):", error);
            }
        });
    }

    get decoding(): boolean {
		return this.#source !== undefined;
	}

	#next(ctx: Promise<void>): void {
		if (this.#source === undefined) {
			return;
		}

		this.#source.acceptGroup(ctx).then(async (result) => {
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

				const chunk = new EncodedJsonChunk({
					type: this.#frameCount < 2 ? "key" : "delta",
					timestamp,
					data: frame.bytes.subarray(headerSize),
				});

				this.#decoder.decode(chunk);
			}

			queueMicrotask(() => this.#next(ctx));
		}).catch(err => {
			console.error("Video decode group error:", err);
		});
	}

	/**
	 * Wait for the next new track to be added to the catalog.
	 * This method only resolves when a new track is added (via "add" operation),
	 * not when tracks are removed or updated. This design exposes only new tracks
	 * while keeping remove/update operations as root catalog editing only.
	 * 
	 * @returns Promise that resolves with the newly added track descriptor
	 * @throws Error if the decoder is cancelled
	 */
	async nextTrack(): Promise<[TrackDescriptor, undefined] | [undefined, Error]> {
		// Wait for the next new track to be added
		const newTrack = new Promise<TrackDescriptor>((resolve) => {
			this.#resolveNewTrack = resolve;
		});

		return newTrack.then(track => [track, undefined] as const);
	}

	hasTrack(name: string): boolean {
		return this.#currentRoot?.tracks.has(name) ?? false;
	}

	async root(): Promise<CatalogRoot> {
		return await this.#root;
	}

    async configure(config: JsonDecoderConfig): Promise<void> {
		// Reset source on configure
		if (this.#source !== undefined) {
			console.warn("[JsonTrackDecoder] source already set. cancelling...");
			await this.#source.closeWithError(InternalSubscribeErrorCode, "codec changed");
			this.#source = undefined;
		}

		// Reset root state when reconfiguring
		this.#currentRoot = undefined;
		this.#newTracks.clear();
		this.#root = new Promise<CatalogRoot>((resolve) => {
			this.#resolveRoot = resolve;
		});

        this.#decoder.configure(config);
    }

    async decodeFrom(ctx: Promise<void>, source: TrackReader): Promise<Error | undefined> {
        if (this.#source !== undefined) {
            console.warn("[JsonTrackDecoder] source already set. replacing...");
            await this.#source.closeWithError(InternalSubscribeErrorCode, "source already set");
        }

		this.#source = source;

        queueMicrotask(() => this.#next(ctx));

		await Promise.race([
            source.context.done(),
            ctx,
        ]);

        return source.context.err() || ContextCancelledError;
    }

	async close(cause?: Error): Promise<void> {
		// Reset state first to prevent #next() from accessing #source
		this.#source = undefined;
		this.#currentRoot = undefined;
		this.#newTracks.clear();
		this.#resolveRoot = undefined;
		this.#resolveNewTrack = undefined;

		this.#decoder.close();
	}
}
