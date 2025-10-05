import * as room from "./room";
import { withCancelCause } from "golikejs/context";
import { describe, it, expect, beforeEach, vi } from "vitest";

import { BroadcastPublisher, BroadcastSubscriber } from "./broadcast";

// Mock external dependencies
vi.mock("./room", () => ({
    participantName: vi.fn((roomID: string, path: string) => "participant"),
}));

const catalogEncoderInstances: Array<{ sync: ReturnType<typeof vi.fn>; setTrack: ReturnType<typeof vi.fn>; removeTrack: ReturnType<typeof vi.fn>; close: ReturnType<typeof vi.fn>; }> = [];
const catalogDecoderInstances: Array<{ decodeFrom: ReturnType<typeof vi.fn>; nextTrack: ReturnType<typeof vi.fn>; root: ReturnType<typeof vi.fn>; close: ReturnType<typeof vi.fn>; }> = [];

vi.mock("./internal/catalog_track", () => {
    const CatalogTrackEncoder = vi.fn().mockImplementation(() => {
        const instance = {
            sync: vi.fn(),
            setTrack: vi.fn(),
            removeTrack: vi.fn(),
            close: vi.fn(),
        };
        catalogEncoderInstances.push(instance);
        return instance;
    });

    const CatalogTrackDecoder = vi.fn().mockImplementation(() => {
        const instance = {
            decodeFrom: vi.fn(async () => undefined),
            nextTrack: vi.fn(async () => [{ name: "catalog" }, undefined] as any),
            root: vi.fn(async () => ({ version: "1", tracks: [] })),
            close: vi.fn(),
        };
        catalogDecoderInstances.push(instance);
        return instance;
    });

    return {
        CatalogTrackEncoder,
        CatalogTrackDecoder,
    };
});

vi.mock("./internal", () => ({
    JsonEncoder: vi.fn(),
    GroupCache: vi.fn(),
}));

vi.mock("./catalog", () => ({
    CATALOG_TRACK_NAME: "catalog",
    RootSchema: {},
    DEFAULT_CATALOG_VERSION: "1",
}));

vi.mock("golikejs/context", () => {
    return {
        withCancelCause: vi.fn(() => {
            const ctx = {
                done: vi.fn(() => Promise.resolve()),
            };
            const cancel = vi.fn();
            return [ctx, cancel] as const;
        }),
        background: vi.fn(() => ({})),
    };
});

beforeEach(() => {
    catalogEncoderInstances.length = 0;
    catalogDecoderInstances.length = 0;
    vi.clearAllMocks();
});

describe("BroadcastPublisher", () => {
    let publisher: BroadcastPublisher;

    beforeEach(() => {
        publisher = new BroadcastPublisher("test-publisher");
    });

    describe("constructor", () => {
        it("should create an instance with name", () => {
            expect(publisher.name).toBe("test-publisher");
        });

        it("should have catalog track", () => {
            expect(publisher.hasTrack("catalog")).toBe(true);
        });
    });

    describe("hasTrack", () => {
        it("should return true for existing track", () => {
            expect(publisher.hasTrack("catalog")).toBe(true);
        });

        it("should return false for non-existing track", () => {
            expect(publisher.hasTrack("non-existing")).toBe(false);
        });
    });

    describe("getTrack", () => {
        it("should return track encoder for existing track", () => {
            const track = publisher.getTrack("catalog");
            expect(track).toBeDefined();
        });

        it("should return undefined for non-existing track", () => {
            const track = publisher.getTrack("non-existing");
            expect(track).toBeUndefined();
        });
    });

    describe("syncCatalog", () => {
        it("should sync catalog", () => {
            expect(() => publisher.syncCatalog()).not.toThrow();
        });
    });

    test("setTrack calls catalog encoder setTrack", () => {
        const mockCatalog = {
            sync: vi.fn(),
            setTrack: vi.fn(),
            removeTrack: vi.fn(),
            close: vi.fn(),
        };
        const publisher = new BroadcastPublisher("room", "path", mockCatalog as any);
        const track = { name: "video" } as any;
        const encoder = {} as any;
        publisher.setTrack(track, encoder);
        expect(mockCatalog.setTrack).toHaveBeenCalledWith(track);
    });

    test("removeTrack calls catalog encoder removeTrack", () => {
        const mockCatalog = {
            sync: vi.fn(),
            setTrack: vi.fn(),
            removeTrack: vi.fn(),
            close: vi.fn(),
        };
        const publisher = new BroadcastPublisher("room", "path", mockCatalog as any);
        publisher.removeTrack("video");
        expect(mockCatalog.removeTrack).toHaveBeenCalledWith("video");
    });

    test("serveTrack calls encoder encodeTo", async () => {
        const mockCatalog = {
            sync: vi.fn(),
            setTrack: vi.fn(),
            removeTrack: vi.fn(),
            close: vi.fn(),
        };
        const publisher = new BroadcastPublisher("room", "path", mockCatalog as any);
        const ctx = Promise.resolve();
        const track = { trackName: "video", closeWithError: vi.fn(), close: vi.fn() } as any;
        const encoder = { encodeTo: vi.fn().mockResolvedValue(undefined), close: vi.fn(), encoding: "mock" } as any;
        publisher.setTrack({ name: "video", priority: 0, schema: "", config: {} }, encoder);
        await publisher.serveTrack(ctx, track);
        expect(encoder.encodeTo).toHaveBeenCalledWith(ctx, track);
    });

    test("close calls catalog encoder close", async () => {
        const mockCatalog = {
            sync: vi.fn(),
            setTrack: vi.fn(),
            removeTrack: vi.fn(),
            close: vi.fn(),
        };
        const publisher = new BroadcastPublisher("room", "path", mockCatalog as any);
        await publisher.close();
        expect(mockCatalog.close).toHaveBeenCalled();
    });
});

describe("BroadcastSubscriber", () => {
    const flushPromises = () => Promise.resolve();
    let mockSession: {
        subscribe: ReturnType<typeof vi.fn>;
    };
    let mockTrack: {
        trackName: string;
        closeWithError: ReturnType<typeof vi.fn>;
    };

    beforeEach(() => {
        mockTrack = {
            trackName: "catalog",
            closeWithError: vi.fn(async () => undefined),
        };

        mockSession = {
            subscribe: vi.fn(async () => [mockTrack, undefined]),
        };
    });

    it("computes participant name and subscribes to catalog track", async () => {
        const mockCatalog = {
            decodeFrom: vi.fn(async () => undefined),
            nextTrack: vi.fn(async () => [{ name: "catalog" }, undefined] as any),
            root: vi.fn(async () => ({ version: "1", tracks: [] })),
            close: vi.fn(),
        };
        const subscriber = new BroadcastSubscriber("/path/to/broadcast", "room-1", mockSession as any, mockCatalog as any);

        await flushPromises();

        expect(mockSession.subscribe).toHaveBeenCalledWith("/path/to/broadcast", "catalog");
        expect(subscriber.name).toBe("participant");

    // participant name should be set on the subscriber (do not assert internal helper calls)
    expect(subscriber.name).toBe("participant");

        expect(mockCatalog.decodeFrom).toHaveBeenCalled();
    });

    it("returns error when subscribeTrack fails", async () => {
        const subscriptionError = new Error("subscribe failed");

        const mockCatalog = {
            decodeFrom: vi.fn(async () => undefined),
            nextTrack: vi.fn(async () => [{ name: "catalog" }, undefined] as any),
            root: vi.fn(async () => ({ version: "1", tracks: [] })),
            close: vi.fn(),
        };
        const subscriber = new BroadcastSubscriber("/path", "room", mockSession as any, mockCatalog as any);
        await flushPromises();

        mockSession.subscribe.mockImplementationOnce(async () => [undefined, subscriptionError]);

        const decoder = { decodeFrom: vi.fn() };
        const result = await subscriber.subscribeTrack("video", decoder as any);

        expect(result).toBe(subscriptionError);
        expect(decoder.decodeFrom).not.toHaveBeenCalled();
    });

    it("cancels context on close", async () => {
        const mockCatalog = {
            decodeFrom: vi.fn(async () => undefined),
            nextTrack: vi.fn(async () => [{ name: "catalog" }, undefined] as any),
            root: vi.fn(async () => ({ version: "1", tracks: [] })),
            close: vi.fn(),
        };
        const subscriber = new BroadcastSubscriber("/path", "room", mockSession as any, mockCatalog as any);
        await flushPromises();

        // withCancelCause のモックを取得
        const callResult = vi.mocked(withCancelCause).mock.results[0]?.value as [unknown, ReturnType<typeof vi.fn>];
        const cancel = callResult ? callResult[1] : undefined;

        subscriber.close();

        expect(cancel).toBeDefined();
        if (cancel) {
            expect(cancel).toHaveBeenCalled();
        }
    });
});
