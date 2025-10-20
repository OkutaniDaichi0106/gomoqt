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
    CatalogInitSchema,
    DEFAULT_CATALOG_VERSION
} from "./catalog";
import type { TrackDescriptor, CatalogInit } from "./catalog";
import type { Context, CancelCauseFunc } from "golikejs/context";
import { withCancelCause, background } from "golikejs/context";
import { participantName } from "./room";
import { CatalogEncoder,TrackCatalog,CatalogDecoder } from "./internal/catalog_stream";
import type { EncodeDestination } from "./internal/container";

type EncodeCallback = (chunk: EncodedChunk) => Promise<void>;

interface Track {
    encodeTo(dest: EncodeDestination): Promise<void>;
}

export class BroadcastPublisher implements TrackHandler {
    readonly name: string;
    #encoders: Map<string, TrackEncoder> = new Map();
    #ctx: Context;
    #cancelCtx: CancelCauseFunc;

    #tracks: Map<string, TrackDescriptor> = new Map();

    #catalog: CatalogEncoder;

    constructor(name: string, description?: string) {
        this.name = name;
        [this.#ctx, this.#cancelCtx] = withCancelCause(background());

        this.#catalog = new CatalogEncoder({
            version: DEFAULT_CATALOG_VERSION,
        });
    }

    setTrack(tracks: {catalog: TrackCatalog}): void {
        if (track.descriptor.name === CATALOG_TRACK_NAME) {
            throw new Error("Cannot add catalog track");
        }

        this.#encoders.set(track.descriptor.name, track.encoder);

        this.#catalog.set()
    }

    async serveTrack(ctx: Promise<void>, track: TrackWriter): Promise<void> {
        if (track.trackName === CATALOG_TRACK_NAME) {
            await this.#catalog.encodeTo()
            return;
        }

        const encoder = this.#encoders.get(track.trackName);
        if (!encoder) {
            track.closeWithError(TrackNotFoundErrorCode, `track not found: ${track.trackName}`);
            return;
        }

        await encoder.encodeTo({
            output: async (chunk: EncodedChunk): Promise<Error | undefined> => {
                return await track.writeFrame(chunk);
            },
            done: ctx,
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

interface DecodeNode {
    decodeFrom(ctx: Promise<void>, reader: TrackReader): Promise<Error | undefined>;
}

export class BroadcastSubscriber {
    #path: BroadcastPath;
    readonly roomID: string;
    readonly session: Session;
    #decoders: Map<string, DecodeNode> = new Map();
    #catalog?: CatalogDecoder;

    #ctx: Context;
    #cancelCtx: CancelCauseFunc;

    // oncatalog?: CatalogCallbacks

    constructor(path: BroadcastPath, roomID: string, session: Session) {
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

    async catalog(): Promise<CatalogDecoder | Error> {
        if (this.#catalog) {
            return this.#catalog;
        }

        const [track, err] = await this.session.subscribe(this.#path, CATALOG_TRACK_NAME);
        if (err) {
            return err;
        }

        this.#catalog = new CatalogDecoder({
            version: DEFAULT_CATALOG_VERSION,
            reader: track,
        });

        return this.#catalog;
    }

    hasTrack(name: string): boolean {
        return this.#decoders.has(name);
    }

    get name(): string {
        return participantName(this.roomID, this.#path);
    }

    async subscribeTrack(name: TrackName, decoder: DecodeNode): Promise<Error | undefined> {
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
        this.#cancelCtx(cause);
    }
}