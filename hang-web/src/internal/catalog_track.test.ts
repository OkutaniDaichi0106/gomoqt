import { describe, test, expect, vi } from 'vitest';

vi.mock("./json", () => ({
    JsonEncoder: vi.fn().mockImplementation((init) => ({
        configure: vi.fn((config) => { setTimeout(() => { init.output?.({ type: "key", data: new Uint8Array(0) }, { decoderConfig: { space: 2 } }); }, 1); }),
        encode: vi.fn(),
        close: vi.fn(),
    })),
    JsonDecoder: vi.fn().mockImplementation((init) => ({ configure: vi.fn(), close: vi.fn(), decode: vi.fn(), output: init?.output })),
    EncodedJsonChunk: vi.fn(),
}));

vi.mock("@okutanidaichi/moqt", () => ({ PublishAbortedErrorCode: 100, InternalGroupErrorCode: 101, InternalSubscribeErrorCode: 102 }));
vi.mock("@okutanidaichi/moqt/io", () => ({ readVarint: vi.fn().mockReturnValue([BigInt(0), 0]) }));
vi.mock("golikejs/context", () => ({
    withCancelCause: vi.fn().mockReturnValue([{ done: () => Promise.resolve(), err: () => undefined }, vi.fn()]),
    background: vi.fn().mockReturnValue({ done: () => Promise.resolve(), err: () => undefined }),
    ContextCancelledError: new Error("Context cancelled"),
    Mutex: class MockMutex { async lock() { return () => {}; } unlock() {} }
}));

import { CatalogEncodeNode, CatalogDecodeNode } from './catalog_track';
import { DEFAULT_CATALOG_VERSION } from '../catalog';

describe('CatalogTrack (minimal)', () => {
    test('encoder root default version', async () => {
        const encoder = new CatalogEncodeNode({});
        const root = await encoder.root();
        expect(root.version).toBe(DEFAULT_CATALOG_VERSION);
        await encoder.close();
    });

    test('decoder default version', () => {
        const decoder = new CatalogDecodeNode({});
        expect(decoder.version).toBe(DEFAULT_CATALOG_VERSION);
    });

    test('set and remove track (encoder)', async () => {
        const encoder = new CatalogEncodeNode({});
        const t = { name: 't', priority: 1, schema: 'video/h264', config: {} } as any;
        encoder.setTrack(t);
        expect(encoder.hasTrack('t')).toBe(true);
        encoder.removeTrack('t');
        expect(encoder.hasTrack('t')).toBe(false);
        await encoder.close();
    });
});
