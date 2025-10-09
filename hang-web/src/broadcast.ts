import type { TrackWriter, Session, TrackReader, TrackHandler } from "@okutanidaichi/moqt";
import { TrackNotFoundErrorCode, SubscribeCanceledErrorCode } from "@okutanidaichi/moqt";
import type { BroadcastPath, TrackName } from "@okutanidaichi/moqt";
import {
    JsonEncoder,
    GroupCache,
    type EncodedChunk,
} from "./internal";
import {
    CATALOG_TRACK_NAME,
    RootSchema,
DEFAULT_CATALOG_VERSION
} from "./catalog";
import type { TrackDescriptor, CatalogRoot } from "./catalog";
import type { Context, CancelCauseFunc } from "golikejs/context";
import { withCancelCause, background } from "golikejs/context";
import { participantName } from "./room";
import { CatalogEncodeNode, CatalogDecodeNode} from "./internal/catalog_track";

type EncodeCallback = (chunk: EncodedChunk) => Promise<void>;

interface EncodeNode {
    encodeTo(ctx: Promise<void>, dest: EncodeCallback): Promise<void>;
}

export class BroadcastPublisher implements TrackHandler {
    readonly name: string;
    #encoders: Map<string, EncodeNode> = new Map();
    #catalog: CatalogEncodeNode;
    #ctx: Context;
    #cancelCtx: CancelCauseFunc;

    constructor(name: string, description?: string, catalog?: CatalogEncodeNode) {
        this.name = name;
        [this.#ctx, this.#cancelCtx] = withCancelCause(background());
        this.#catalog = catalog ?? new CatalogEncodeNode({ description });

        // Set up catalog track
        const self = this;

        this.#encoders.set(CATALOG_TRACK_NAME, this.#catalog);
    }

    syncCatalog(): void {
        this.#catalog.sync();
    }

    hasTrack(name: string): boolean {
        return this.#encoders.has(name);
    }

    getTrack(name: string): EncodeNode | undefined {
        return this.#encoders.get(name);
    }

    setTrack(track: TrackDescriptor, encoder: EncodeNode): void {
        if (track.name === CATALOG_TRACK_NAME) {
            throw new Error("Cannot add catalog track");
        }

        this.#encoders.set(track.name, encoder);

        // Update catalog
        this.#catalog.setTrack(track);
    }

    removeTrack(name: string): void {
        if (name === CATALOG_TRACK_NAME) {
            throw new Error("Cannot remove catalog track");
        }

        this.#encoders.delete(name);

        // Update catalog
        this.#catalog.removeTrack(name);
    }

    async serveTrack(ctx: Promise<void>, track: TrackWriter): Promise<void> {
        const encoder = this.#encoders.get(track.trackName);
        if (!encoder) {
            track.closeWithError(TrackNotFoundErrorCode, `track not found: ${track.trackName}`);
            return;
        }

        await encoder.encodeTo(ctx, async (chunk: EncodedChunk) => {

        });

        await track.close(); // Ensure the track is closed after serving; Is this necessary?
    }

    async close(cause?: Error): Promise<void> {
        const catalogEncoder = this.#encoders.get(CATALOG_TRACK_NAME);
        if (catalogEncoder) {
            await catalogEncoder.close(cause);
        }
        this.#catalog.close();
        this.#encoders.clear();
    }
}

// export interface CatalogCallbacks {
//     onroot?: (desc: CatalogRoot) => void;
//     onpatch?: (patch: unknown[]) => void;
// }

export class BroadcastSubscriber {
    #path: BroadcastPath;
    readonly roomID: string;
    readonly session: Session;
    #decoders: Map<string, TrackDecoder> = new Map();
    #catalog: CatalogDecodeNode;

    #ctx: Context;
    #cancelCtx: CancelCauseFunc;

    // oncatalog?: CatalogCallbacks

    constructor(path: BroadcastPath, roomID: string, session: Session, catalog?: CatalogDecodeNode) {
        this.#path = path;
        this.roomID = roomID;
        this.session = session;
        const [ctx, cancelCtx] = withCancelCause(background());
        this.#ctx = ctx;
        this.#cancelCtx = (cause?: Error) => {
            cancelCtx(cause);
        };
        this.#catalog = catalog ?? new CatalogDecodeNode({});

        this.subscribeTrack(CATALOG_TRACK_NAME, this.#catalog).then((err) => {
            // Ignore errors
            if (err) {
                console.error("Failed to subscribe to catalog track:", err);
            }
        });
    }

    async nextTrack(): Promise<[TrackDescriptor, undefined] | [undefined, Error]> {
        return await this.#catalog.nextTrack();
    }

    hasTrack(name: string): boolean {
        return this.#decoders.has(name);
    }

    get name(): string {
        return participantName(this.roomID, this.#path);
    }

    async syncCatalog(): Promise<CatalogRoot> {
        return this.#catalog.root();
    }

    async subscribeTrack(name: TrackName, decoder: TrackDecoder): Promise<Error | undefined> {
        // Make a new subscription
        const [track, err] = await this.session.subscribe(this.#path, name);
        if (err) {
            console.debug("Failed to subscribe to track:", name);
            return err;
        }

        await decoder.decodeFrom(this.#ctx.done(), track);

        // When the decoder is done, ensure to close the track
        await track.closeWithError(SubscribeCanceledErrorCode, "decoder closed");
    }

    async close(cause?: Error): Promise<void> {
        this.#decoders.clear();

        // Cancel context to stop all decoders
        // This will also close all active subscriptions
        this.#cancelCtx(undefined);
    }
}