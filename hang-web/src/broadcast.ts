import type { TrackWriter, Session, TrackReader } from "@okutanidaichi/moqt";
import { TrackNotFoundErrorCode,type BroadcastPath, type TrackName } from "@okutanidaichi/moqt";
import {
    JsonEncoder,
    JsonEncodeStream,
    GroupCache,
    JsonDecodeStream
} from "./internal";
import type {
    TrackEncoder,
    TrackDecoder,
} from "./internal";
import  { CATALOG_TRACK_NAME, type Track, CatalogController,type Root,RootSchema } from "./catalog";

export class BroadcastPublisher implements BroadcastViewer {
    readonly path: BroadcastPath;
    readonly tracks: Map<string, TrackEncoder<any>> = new Map();
    #enqueueCatalog!: () => void;
    #catalog: CatalogController;

    constructor(path: BroadcastPath, description?: string) {
        this.path = path;
        this.#catalog = new CatalogController({ description });

        // Set up catalog track
        const self = this;
        const catalogReader = new ReadableStream({
            start(controller) {
                function enqueueCatalog() {
                    // Enqueue the current full catalog JSON
                    controller.enqueue(self.#catalog.root);
                    // TODO: Enqueue deltas if implemented
                }

                self.#enqueueCatalog = enqueueCatalog;
            }
        }).getReader();
        const catalogEncoder = new JsonEncodeStream({
            source: catalogReader,
            cache: GroupCache,
        });
        this.tracks.set(CATALOG_TRACK_NAME, catalogEncoder);
    }

    catalog(): Root {
        return this.#catalog.root;
    }

    addTrack(encoder: TrackEncoder<any>, track: Track): void {
        if (track.name === CATALOG_TRACK_NAME) {
            throw new Error("Cannot add catalog track");
        }
        this.tracks.set(track.name, encoder);

        this.#catalog.updateTrack(track);
        this.#enqueueCatalog();
    }

    removeTrack(name: string): void {
        if (name === CATALOG_TRACK_NAME) {
            throw new Error("Cannot remove catalog track");
        }
        this.tracks.delete(name);

        this.#catalog.removeTrack(name);
        this.#enqueueCatalog();
    }

    async serveTrack(ctx: Promise<void>, track: TrackWriter): Promise<void> {
        const encoder = this.tracks.get(track.trackName);
        if (!encoder) {
            track.closeWithError(TrackNotFoundErrorCode, `track not found: ${track.trackName}`);
            return;
        }

        encoder.encodeTo(track);

        await ctx;

        track.close();
    }

    view(name: TrackName, dest: WritableStreamDefaultWriter<any>): void {
        const encoder = this.tracks.get(name);
        if (!encoder) {
            throw new Error(`Track not found: ${name}`);
        }

        encoder.preview(dest);
    }

    close(): void {
        for (const [,encoder] of this.tracks) {
            encoder.close();
        }
        this.tracks.clear();
    }
}

export class BroadcastSubscriber implements BroadcastViewer {
    readonly path: BroadcastPath;
    readonly session: Session;
    readonly tracks: Map<string, TrackDecoder<any>> = new Map();
    #catalog: CatalogController;

    constructor(path: BroadcastPath, session: Session) {
        this.path = path;
        this.session = session;
        const controller = new CatalogController();
        this.#catalog = controller;
        this.session.subscribe(this.path, CATALOG_TRACK_NAME).then(async (track)=>{
            const json = new JsonDecodeStream(track);
            const writer = new WritableStream({
                write: (value) => {
                    const root = RootSchema.parse(value);
                    controller.reset(root);
                }
            }).getWriter();
            json.decodeTo(writer);
        });
    }

    catalog(): Root {
        return this.#catalog.root;
    }

    async subscribe(name: TrackName, decoder: new(src: TrackReader) => TrackDecoder<any>): Promise<void> {
        if (!this.#catalog.root.tracks.has(name)) {
            throw new Error(`Track not found: ${name}`);
        }
        if (this.tracks.has(name)) {
            return;
        }

        // Make a new subscription
        const source = await this.session.subscribe(this.path, name);
        const trackDecoder = new decoder(source);

        this.tracks.set(name, trackDecoder);
    }

    view(name: TrackName, dest: WritableStreamDefaultWriter<any>): void {
        this.tracks.get(name)?.decodeTo(dest);
    }

    close(): void {
        for (const [,decoder] of this.tracks) {
            decoder.close();
        }
        this.tracks.clear();
    }
}

export interface BroadcastViewer {
    view(name: TrackName, dest: WritableStreamDefaultWriter<any>): void;
    catalog(): Root;
}