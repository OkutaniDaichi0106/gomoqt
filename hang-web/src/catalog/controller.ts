import { DEFAULT_CATALOG_VERSION } from "./root";
import type { Root } from "./root";
import type { Track } from "./track"

export class CatalogController {
	#root: Root;
	// #deltas?: any[]; // TODO: Implement delta tracking

	constructor(init: CatalogControllerInit = {}) {
		this.#root = {
			version: init.version ?? DEFAULT_CATALOG_VERSION,
			description: init.description ?? "",
			tracks: init.tracks ?? new Map()
		};
	}

	reset(init: CatalogControllerInit): void {
		this.#root = {
			version: init.version ?? DEFAULT_CATALOG_VERSION,
			description: init.description ?? "",
			tracks: init.tracks ?? new Map()
		};
	}

	get root(): Root {
		return this.#root;
	}

	// get delta(): any[] | undefined {
	// 	const deltas = this.#deltas;

	// 	// Clear deltas after reading
	// 	this.#deltas = undefined;
		
	// 	return deltas;
	// }

	updateTrack(track: Track): void {
		this.#root.tracks.set(track.name, track);
	}
	
	removeTrack(trackName: string): void {
		this.#root.tracks.delete(trackName);
	}
}

export interface CatalogControllerInit {
	version?: string;
	description?: string;
	tracks?: Map<string, Track>;
}

// TODO: Support JSON patching