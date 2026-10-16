import { Broadcast as TrackBroadcast } from "../broadcast.ts";
import { GroupErrorCode, SubscribeErrorCode } from "../error.ts";
import type { TrackHandler } from "../track_mux.ts";
import type { TrackWriter } from "../track_writer.ts";
import {
	type Catalog,
	cloneCatalog,
	cloneTrack,
	stringifyCatalog,
	type Track,
	trackId,
	validateCatalog,
} from "./catalog.ts";

/** Default track name used for the catalog in an MSF broadcast. */
export const DefaultCatalogTrackName = "catalog";

/**
 * Catalog-aware broadcast that manages an MSF {@link Catalog}
 * and delegates track serving to registered {@link TrackHandler} instances.
 *
 * A catalog subscription stays open: it gets the catalog at subscribe time as
 * its first group, and one group per change after that, so a subscriber sees
 * tracks registered or removed while it is subscribed. It ends when the
 * subscriber leaves, the track writer is closed, or {@link Broadcast.close}
 * is called, which sends any pending change first.
 *
 * @example
 * ```ts
 * import { Broadcast } from "@qumo/moq/msf";
 * import type { Catalog } from "@qumo/moq/msf";
 *
 * const catalog: Catalog = { tracks: [{ name: "audio", packaging: "cmaf" }] };
 * const broadcast = new Broadcast(catalog);
 * ```
 */
export class Broadcast implements TrackHandler {
	#catalogTrackName: string;
	#catalog: Catalog;
	/** The encoded catalog, built on first use after each change. */
	#payload: Uint8Array | undefined;
	/** Resolved and replaced on each change, and resolved on close. */
	#changed = newSignal();
	#closed = false;
	#tracks = new TrackBroadcast();

	constructor(catalog: Catalog, catalogTrackName = DefaultCatalogTrackName) {
		this.#catalogTrackName = catalogTrackName;
		this.#catalog = prepareCatalog(catalog, this.#catalogTrackName);
	}

	get catalogTrackName(): string {
		return this.#catalogTrackName;
	}

	catalog(): Catalog {
		return cloneCatalog(this.#catalog);
	}

	catalogBytes(): Uint8Array {
		return this.#catalogPayload().slice();
	}

	async setCatalog(catalog: Catalog): Promise<void> {
		const clone = prepareCatalog(catalog, this.#catalogTrackName);
		const staleTrackNames = staleTrackNamesFrom(this.#catalog, clone);
		this.#catalog = clone;
		this.#catalogChanged();
		await Promise.allSettled(staleTrackNames.map((name) => this.#tracks.remove(name)));
	}

	async registerTrack(track: Track, handler: TrackHandler): Promise<void> {
		if (handler === undefined || handler === null) {
			throw new Error("msf: track handler cannot be nil");
		}
		if (!track.name) {
			throw new Error("msf: track name is required");
		}
		if (track.name === this.#catalogTrackName) {
			throw new Error(`msf: \"${track.name}\" is reserved for the catalog track`);
		}
		const trackName = track.name;

		const trackClone = cloneTrack(track);
		const updated = cloneCatalog(this.#catalog);
		const nextId = trackId(trackClone, updated.defaultNamespace);

		let replaced = false;
		for (let i = 0; i < updated.tracks.length; i++) {
			const current = updated.tracks[i]!;
			if (current.name !== trackClone.name) {
				continue;
			}
			if (trackId(current, updated.defaultNamespace) !== nextId) {
				throw new Error(
					`msf: broadcast requires unique track names across namespaces; duplicate name ${
						JSON.stringify(trackClone.name)
					} found`,
				);
			}
			updated.tracks[i] = trackClone;
			replaced = true;
			break;
		}
		if (!replaced) {
			updated.tracks.push(trackClone);
		}
		validateCatalog(updated);
		validateCatalogForBroadcast(updated, this.#catalogTrackName);

		await this.#tracks.register(trackName, handler);
		this.#catalog = updated;
		this.#catalogChanged();
	}

	async removeTrack(name: string): Promise<boolean> {
		if (name === "" || name === this.#catalogTrackName) {
			return false;
		}

		const updated = cloneCatalog(this.#catalog);
		const beforeLength = updated.tracks.length;
		updated.tracks = updated.tracks.filter((track) => track.name !== name);
		const removedFromCatalog = updated.tracks.length !== beforeLength;
		if (removedFromCatalog) {
			this.#catalog = updated;
			this.#catalogChanged();
		}
		const removedFromTracks = await this.#tracks.remove(name);
		return removedFromCatalog || removedFromTracks;
	}

	handler(name: string): TrackHandler {
		if (name === this.#catalogTrackName) {
			return { serveTrack: this.#serveCatalogTrack.bind(this) };
		}
		return this.#tracks.handler(name);
	}

	async serveTrack(trackWriter: TrackWriter): Promise<void> {
		await this.handler(trackWriter.trackName).serveTrack(trackWriter);
	}

	/**
	 * Closes the registered tracks and ends every catalog subscription after
	 * it has the current catalog. A catalog subscription made after close gets
	 * the catalog as one group and ends. Calling close again does nothing more.
	 */
	async close(): Promise<void> {
		if (!this.#closed) {
			this.#closed = true;
			this.#changed.resolve();
		}
		await this.#tracks.close();
	}

	#catalogChanged(): void {
		this.#payload = undefined;
		if (this.#closed) {
			return;
		}
		this.#changed.resolve();
		this.#changed = newSignal();
	}

	#catalogPayload(): Uint8Array {
		this.#payload ??= new TextEncoder().encode(stringifyCatalog(this.#catalog));
		return this.#payload;
	}

	async #serveCatalogTrack(trackWriter: TrackWriter): Promise<void> {
		let left = false;
		const gone = subscriberGone(trackWriter).then(() => {
			left = true;
		});
		let sent: Uint8Array | undefined;
		for (;;) {
			// Read together, with no await between, so a change after this
			// point resolves the signal being waited on.
			const payload = this.#catalogPayload();
			const changed = this.#changed.promise;
			const closed = this.#closed;

			if (sent === undefined || !bytesEqual(payload, sent)) {
				const err = await writeCatalogGroup(trackWriter, payload);
				if (err) {
					// A subscriber that left mid-write is not an error.
					if (!left && trackWriter.context.err() === undefined) {
						await trackWriter.closeWithError(SubscribeErrorCode.InternalError);
					}
					return;
				}
				sent = payload;
			}
			if (closed) {
				await trackWriter.close();
				return;
			}
			if (await Promise.race([changed.then(() => false), gone.then(() => true)])) {
				await trackWriter.close();
				return;
			}
		}
	}
}

/** Writes payload as one catalog group. */
async function writeCatalogGroup(
	trackWriter: TrackWriter,
	payload: Uint8Array,
): Promise<Error | undefined> {
	const [group, openErr] = await trackWriter.openGroup();
	if (openErr || !group) {
		return new Error(`msf: open catalog group: ${openErr}`);
	}
	const writeErr = await group.writeFrame(payload);
	if (writeErr) {
		await group.cancel(GroupErrorCode.InternalError);
		return new Error(`msf: write catalog group: ${writeErr}`);
	}
	await group.close();
	return undefined;
}

/**
 * Resolves when the subscription ends: the track writer's context is done
 * (closed here, or the session ended), or the subscribe stream ended, which
 * readUpdate reports as an error (or throws, for a reset stream). A
 * subscriber that leaves ends the stream without ending the publisher's
 * context.
 */
function subscriberGone(trackWriter: TrackWriter): Promise<void> {
	const streamEnded = (async () => {
		try {
			for (;;) {
				const [, err] = await trackWriter.readUpdate();
				if (err) {
					return;
				}
			}
		} catch {
			// A reset stream has ended too.
		}
	})();
	return Promise.race([trackWriter.context.done(), streamEnded]).then(() => {});
}

interface Signal {
	promise: Promise<void>;
	resolve: () => void;
}

function newSignal(): Signal {
	let resolve!: () => void;
	const promise = new Promise<void>((r) => {
		resolve = r;
	});
	return { promise, resolve };
}

function bytesEqual(a: Uint8Array, b: Uint8Array): boolean {
	if (a.length !== b.length) {
		return false;
	}
	for (let i = 0; i < a.length; i++) {
		if (a[i] !== b[i]) {
			return false;
		}
	}
	return true;
}

function staleTrackNamesFrom(previous: Catalog, next: Catalog): string[] {
	if (previous.tracks.length === 0) {
		return [];
	}

	const activeNames = new Set(
		next.tracks.map((track) => track.name).filter((name): name is string => !!name),
	);
	return previous.tracks
		.map((track) => track.name)
		.filter((name): name is string => !!name && !activeNames.has(name));
}

function prepareCatalog(catalog: Catalog, catalogTrackName: string): Catalog {
	const clone = cloneCatalog(catalog);
	validateCatalog(clone);
	validateCatalogForBroadcast(clone, catalogTrackName);
	return clone;
}

function validateCatalogForBroadcast(catalog: Catalog, catalogTrackName: string): void {
	const seen = new Map<string, string>();
	for (const track of catalog.tracks) {
		if (!track.name) {
			continue;
		}
		if (track.name === catalogTrackName) {
			throw new Error(`msf: catalog contains reserved track name \"${catalogTrackName}\"`);
		}
		const id = trackId(track, catalog.defaultNamespace);
		const previous = seen.get(track.name);
		if (previous !== undefined && previous !== id) {
			throw new Error(
				`msf: broadcast requires unique track names across namespaces; duplicate name ${
					JSON.stringify(track.name)
				} found`,
			);
		}
		seen.set(track.name, id);
	}
}
