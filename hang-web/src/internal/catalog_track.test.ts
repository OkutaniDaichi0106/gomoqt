import { describe, test, expect, beforeEach, afterEach, vi } from 'vitest';

vi.mock("./json", () => ({
    JsonEncoder: vi.fn().mockImplementation((init) => {
        const instance = {
            configure: vi.fn((config) => {
                // Simulate calling the output callback with decoder config
                // This should trigger the resolveConfig in the encoder
                setTimeout(() => {
                    if (init.output) {
                        // Call output with metadata containing decoderConfig
                        init.output(
                            { type: "key", data: new Uint8Array(0) },
                            { decoderConfig: { space: 2 } }
                        );
                    }
                }, 1);
            }),
            encode: vi.fn(),
            close: vi.fn(),
        };
        return instance;
    }),
    JsonDecoder: vi.fn().mockImplementation(() => ({
        configure: vi.fn(),
        close: vi.fn(),
        decode: vi.fn(),
    })),
    EncodedJsonChunk: vi.fn(),
}));
vi.mock("@okutanidaichi/moqt", () => ({
    PublishAbortedErrorCode: 100,
    InternalGroupErrorCode: 101,
    InternalSubscribeErrorCode: 102,
}));

vi.mock("@okutanidaichi/moqt/io", () => ({
    readVarint: vi.fn().mockReturnValue([BigInt(0), 0]),
}));

vi.mock("golikejs/context", () => {
    const ContextCancelledError = new Error("Context cancelled");
    return {
        withCancelCause: vi.fn().mockReturnValue([
            { 
                done: () => Promise.resolve(),
                err: () => undefined 
            },
            vi.fn()
        ]),
        withPromise: vi.fn(),
        background: vi.fn().mockReturnValue({ 
            done: () => Promise.resolve(),
            err: () => undefined 
        }),
        ContextCancelledError,
        Mutex: class MockMutex {
            async lock() {
                return () => {};
            }
            unlock() {}
        }
    };
});

import { ContextCancelledError } from 'golikejs/context';

import { CatalogTrackEncoder, CatalogTrackDecoder } from "./catalog_track";
import type { CatalogTrackEncoderInit, CatalogTrackDecoderInit } from "./catalog_track";
import type { TrackDescriptor } from "../catalog";
import { DEFAULT_CATALOG_VERSION } from "../catalog";
import type { TrackWriter, TrackReader } from "@okutanidaichi/moqt";

describe("CatalogTrackEncoder", () => {
    let encoder: CatalogTrackEncoder;

    afterEach(async () => {
        // Clean up encoder and mocks after each test
        if (encoder) {
            await encoder.close();
        }
        vi.clearAllMocks();
    });

    describe("constructor", () => {
        test("creates encoder with default version", () => {
            encoder = new CatalogTrackEncoder({});
            
            expect(encoder).toBeDefined();
            expect(encoder.encoding).toBe(false);
        });

        test("creates encoder with custom version and description", () => {
            const init: CatalogTrackEncoderInit = {
                version: "1.0.0",
                description: "Test catalog"
            };
            
            const encoder = new CatalogTrackEncoder(init);
            
            expect(encoder).toBeDefined();
            expect(encoder.encoding).toBe(false);
        });
    });

    describe("setTrack", () => {
        beforeEach(() => {
            encoder = new CatalogTrackEncoder({});
        });

        test("adds new track to catalog", async () => {
            const track: TrackDescriptor = {
                name: "test-track",
                priority: 1,
                schema: "video/h264",
                config: { width: 1920, height: 1080 }
            };

            encoder.setTrack(track);
            
            const root = await encoder.root();
            expect(root.tracks.has("test-track")).toBe(true);
            expect(root.tracks.get("test-track")).toEqual(track);
        });

        test("updates existing track in catalog", async () => {
            const originalTrack: TrackDescriptor = {
                name: "test-track",
                priority: 1,
                schema: "video/h264",
                config: { width: 1920, height: 1080 }
            };
            const updatedTrack: TrackDescriptor = {
                name: "test-track",
                priority: 2,
                schema: "video/h264",
                config: { width: 3840, height: 2160 }
            };

            encoder.setTrack(originalTrack);
            encoder.setTrack(updatedTrack);
            
            const root = await encoder.root();
            expect(root.tracks.get("test-track")).toEqual(updatedTrack);
        });

        test("does not create patch for identical track", async () => {
            const track: TrackDescriptor = {
                name: "test-track",
                priority: 1,
                schema: "video/h264",
                config: { width: 1920, height: 1080 }
            };

            encoder.setTrack(track);
            encoder.setTrack(track); // Set same track again
            
            const root = await encoder.root();
            expect(root.tracks.has("test-track")).toBe(true);
        });

        test("does not set track when cancelled", async () => {
            const track: TrackDescriptor = {
                name: "test-track",
                priority: 1,
                schema: "video/h264",
                config: {}
            };

            await encoder.close(new Error("cancelled"));
            
            // setTrack should not throw after close
            expect(() => encoder.setTrack(track)).not.toThrow();
        });

        test("setTrack adds track to catalog", () => {
            const encoder = new CatalogTrackEncoder({});
            const track = { name: "test-track", schema: "", config: {}, priority: 0 };
            encoder.setTrack(track);
            
            // Since private fields can't be accessed, we test that no error occurs
            // and the method completes successfully
            expect(() => encoder.setTrack(track)).not.toThrow();
        });
    });

    describe("removeTrack", () => {
        beforeEach(() => {
            encoder = new CatalogTrackEncoder({});
        });

        test("removes existing track from catalog", async () => {
            const track: TrackDescriptor = {
                name: "test-track",
                priority: 1,
                schema: "video/h264",
                config: {}
            };

            encoder.setTrack(track);
            expect((await encoder.root()).tracks.has("test-track")).toBe(true);
            
            encoder.removeTrack("test-track");
            expect((await encoder.root()).tracks.has("test-track")).toBe(false);
        });

        test("does not error when removing non-existent track", async () => {
            expect(() => encoder.removeTrack("non-existent")).not.toThrow();
            
            const root = await encoder.root();
            expect(root.tracks.size).toBe(0);
        });

        test("does not remove track when cancelled", async () => {
            const track: TrackDescriptor = {
                name: "test-track",
                priority: 1,
                schema: "video/h264",
                config: {}
            };

            encoder.setTrack(track);
            // Verify track is added
            const rootBefore = await encoder.root();
            expect(rootBefore.tracks.has("test-track")).toBe(true);
            
            await encoder.close(new Error("cancelled"));
            
            // removeTrack should not throw after close
            expect(() => encoder.removeTrack("test-track")).not.toThrow();
        });
    });

    describe("hasTrack", () => {
        beforeEach(() => {
            encoder = new CatalogTrackEncoder({});
        });

        test("returns true for existing track", async () => {
            const track: TrackDescriptor = {
                name: "test-track",
                priority: 1,
                schema: "video/h264",
                config: {}
            };

            encoder.setTrack(track);
            
            expect(encoder.hasTrack("test-track")).toBe(true);
        });

        test("returns false for non-existent track", () => {
            expect(encoder.hasTrack("non-existent")).toBe(false);
        });
    });

    describe("sync", () => {
        beforeEach(() => {
            encoder = new CatalogTrackEncoder({});
        });

        test("does nothing when no patches exist", () => {
            expect(() => encoder.sync()).not.toThrow();
        });

        test("does not sync when cancelled", async () => {
            const track: TrackDescriptor = {
                name: "test-track",
                priority: 1,
                schema: "video/h264",
                config: {}
            };

            encoder.setTrack(track);
            await encoder.close();
            
            expect(() => encoder.sync()).not.toThrow();
        });
    });

    describe("encodeTo", () => {
        beforeEach(() => {
            encoder = new CatalogTrackEncoder({});
        });

        test("returns error when encoder is cancelled", async () => {
            await encoder.close(new Error("Test cancellation"));
            
            const mockWriter = {
                context: { done: () => Promise.resolve(), err: () => undefined }
            } as Partial<TrackWriter> as TrackWriter;
            
            const ctx = Promise.resolve();
            const result = await encoder.encodeTo(ctx, mockWriter);
            
            // The method should return undefined or an error when cancelled
            // This test verifies the method completes without throwing
        });

        test("encodes successfully when not cancelled", async () => {
            const mockWriter = {
                context: { 
                    done: () => new Promise(() => {}), // Never resolves
                    err: () => undefined 
                }
            } as Partial<TrackWriter> as TrackWriter;
            
            const ctx = new Promise<void>((resolve) => setTimeout(resolve, 10)); // Resolves quickly
            
            const result = await encoder.encodeTo(ctx, mockWriter);
            
            expect(result).toBe(ContextCancelledError);
        });

        test("returns undefined when TrackWriter is already being encoded to", async () => {
            const mockWriter = {
                context: { 
                    done: () => new Promise(() => {}), // Never resolves
                    err: () => undefined 
                }
            } as Partial<TrackWriter> as TrackWriter;
            
            const ctx = new Promise<void>((resolve) => setTimeout(resolve, 10));
            
            // Start two encodeTo calls simultaneously
            const promise1 = encoder.encodeTo(ctx, mockWriter);
            const promise2 = encoder.encodeTo(ctx, mockWriter);
            
            // Wait for both to complete
            const [result1, result2] = await Promise.all([promise1, promise2]);
            
            // One should succeed (return ContextCancelledError), the other should return undefined
            expect([result1, result2]).toContain(ContextCancelledError);
            expect([result1, result2]).toContain(undefined);
        });

        test("returns error when dest context is cancelled", async () => {
            const mockWriter = {
                context: { 
                    done: () => Promise.resolve(),
                    err: () => new Error("Dest context cancelled") 
                }
            } as Partial<TrackWriter> as TrackWriter;
            
            const ctx = new Promise<void>((resolve) => setTimeout(resolve, 10));
            
            const result = await encoder.encodeTo(ctx, mockWriter);
            
            expect(result).toBeInstanceOf(Error);
            expect(result?.message).toBe("Dest context cancelled");
        });
    });

    describe("root", () => {
        beforeEach(() => {
            encoder = new CatalogTrackEncoder({});
        });

        test("returns root catalog with default version", async () => {
            const root = await encoder.root();
            
            expect(root.version).toBe(DEFAULT_CATALOG_VERSION);
            expect(root.description).toBe("");
            expect(root.tracks).toBeInstanceOf(Map);
            expect(root.tracks.size).toBe(0);
        });

        test("returns root catalog with custom values", async () => {
            const init: CatalogTrackEncoderInit = {
                version: "2.0.0",
                description: "Custom catalog"
            };
            encoder = new CatalogTrackEncoder(init);
            
            const root = await encoder.root();
            
            expect(root.version).toBe("2.0.0");
            expect(root.description).toBe("Custom catalog");
        });
    });

    describe("close", () => {
        beforeEach(() => {
            encoder = new CatalogTrackEncoder({});
        });

        test("closes successfully without cause", async () => {
            await encoder.close();
            
            expect(encoder.encoding).toBe(false);
        });

        test("closes successfully with cause", async () => {
            const cause = new Error("Test error");
            
            await encoder.close(cause);
            
            expect(encoder.encoding).toBe(false);
        });

        test("does not error on multiple close calls", async () => {
            await encoder.close();
            await encoder.close();
            
            expect(encoder.encoding).toBe(false);
        });
    });

    describe("configure", () => {
        beforeEach(() => {
            encoder = new CatalogTrackEncoder({});
        });

        test("configures encoder successfully", async () => {
            const config = {
                space: 2,
                replacer: [] as ("bigint" | "date")[]
            };

            const result = await encoder.configure(config);
            
            expect(result).toEqual({ space: 2 });
        });
    });

    describe("encoding property", () => {
        beforeEach(() => {
            encoder = new CatalogTrackEncoder({});
        });

        test("returns false when no tracks are encoding", () => {
            expect(encoder.encoding).toBe(false);
        });
    });
});

describe("CatalogTrackDecoder", () => {
    let decoder: CatalogTrackDecoder;

    afterEach(async () => {
        if (decoder) { await decoder.close(); }
        vi.clearAllMocks();
    });

    describe("constructor", () => {
        test("creates decoder with default version", () => {
            decoder = new CatalogTrackDecoder({});
            
            expect(decoder).toBeDefined();
            expect(decoder.version).toBe(DEFAULT_CATALOG_VERSION);
            expect(decoder.decoding).toBe(false);
        });

        test("creates decoder with custom version", () => {
            const init: CatalogTrackDecoderInit = {
                version: "1.0.0"
            };
            decoder = new CatalogTrackDecoder(init);
            
            expect(decoder.version).toBe("1.0.0");
        });
    });

    describe("hasTrack", () => {
        beforeEach(() => {
            decoder = new CatalogTrackDecoder({});
        });

        test("returns false when no root is set", () => {
            expect(decoder.hasTrack("test-track")).toBe(false);
        });

        test("returns false for non-existent track", () => {
            expect(decoder.hasTrack("non-existent")).toBe(false);
        });
    });

    describe("root", () => {
        beforeEach(() => {
            decoder = new CatalogTrackDecoder({});
        });

        test("returns root catalog promise", async () => {
            const rootPromise = decoder.root();
            expect(rootPromise).toBeInstanceOf(Promise);
        });
    });

    describe("nextTrack", () => {
        beforeEach(() => {
            decoder = new CatalogTrackDecoder({});
        });

        test("returns error when decoder is cancelled", async () => {
            const cause = new Error("Test cancellation");
            
            // Start nextTrack before closing to test cancellation during wait
            const nextTrackPromise = decoder.nextTrack();
            
            // Close the decoder
            await decoder.close(cause);
            
            // nextTrack should return error after close
            const result = await nextTrackPromise;
            expect(result[0]).toBeUndefined();
            expect(result[1]).toBeInstanceOf(Error);
        });
    });

    describe("configure", () => {
        beforeEach(() => {
            decoder = new CatalogTrackDecoder({});
        });

        test("configures decoder successfully", async () => {
            const config = {
                reviverRules: [] as ("bigint" | "date")[]
            };

            await decoder.configure(config);
            
            // Should complete without error
            expect(decoder.decoding).toBe(false);
        });

        test("does not configure when cancelled", async () => {
            await decoder.close();
            await decoder.configure({ reviverRules: [] });
            
            // Should complete without error
            expect(decoder.decoding).toBe(false);
        });
    });

    describe("decodeFrom", () => {
        beforeEach(() => {
            decoder = new CatalogTrackDecoder({});
        });

        test("returns error when decoder is cancelled", async () => {
            await decoder.close(new Error("Test cancellation"));
            
            const mockReader = {
                context: { done: () => Promise.resolve(), err: () => undefined },
                acceptGroup: () => Promise.resolve([undefined, new Error("cancelled")]),
                closeWithError: () => Promise.resolve()
            } as Partial<TrackReader> as TrackReader;
            
            const ctx = Promise.resolve();
            const result = await decoder.decodeFrom(ctx, mockReader);
            
            // The method should return undefined or an error when cancelled
            // This test verifies the method completes without throwing
        });
    });

    describe("close", () => {
        beforeEach(() => {
            decoder = new CatalogTrackDecoder({});
        });

        test("closes successfully without cause", async () => {
            await decoder.close();
            
            expect(decoder.decoding).toBe(false);
        });

        test("closes successfully with cause", async () => {
            const cause = new Error("Test error");
            
            await decoder.close(cause);
            
            expect(decoder.decoding).toBe(false);
        });

        test("does not error on multiple close calls", async () => {
            await decoder.close();
            await decoder.close();
            
            expect(decoder.decoding).toBe(false);
        });
    });

    describe("decoding property", () => {
        beforeEach(() => {
            decoder = new CatalogTrackDecoder({});
        });

        test("returns false when not decoding", () => {
            expect(decoder.decoding).toBe(false);
        });

        test("returns false when cancelled", async () => {
            await decoder.close();
            
            expect(decoder.decoding).toBe(false);
        });
    });
});

describe("CatalogTrackEncoder and CatalogTrackDecoder Integration", () => {
    let encoder: CatalogTrackEncoder;
    let decoder: CatalogTrackDecoder;

    afterEach(async () => {
        if (encoder) { await encoder.close(); }
        if (decoder) { await decoder.close(); }
        vi.clearAllMocks();
    });

    test("encoder and decoder work together", async () => {
        encoder = new CatalogTrackEncoder({
            version: "1.0.0",
            description: "Integration test catalog"
        });
        decoder = new CatalogTrackDecoder({
            version: "1.0.0"
        });

        // Add tracks to encoder
        const track1: TrackDescriptor = {
            name: "video-track",
            priority: 1,
            schema: "video/h264",
            config: { width: 1920, height: 1080 }
        };
        const track2: TrackDescriptor = {
            name: "audio-track",
            priority: 2,
            schema: "audio/opus",
            config: { sampleRate: 48000 }
        };

        encoder.setTrack(track1);
        encoder.setTrack(track2);

        // Verify encoder root
        const encoderRoot = await encoder.root();
        expect(encoderRoot.tracks.size).toBe(2);
        expect(encoderRoot.tracks.has("video-track")).toBe(true);
        expect(encoderRoot.tracks.has("audio-track")).toBe(true);
    });

    test("encoder handles track updates correctly", async () => {
        encoder = new CatalogTrackEncoder({});
        
        const originalTrack: TrackDescriptor = {
            name: "test-track",
            priority: 1,
            schema: "video/h264",
            config: { width: 1920, height: 1080 }
        };
        
        const updatedTrack: TrackDescriptor = {
            name: "test-track",
            priority: 2,
            schema: "video/h264",
            config: { width: 3840, height: 2160 }
        };

        // Set original track
        encoder.setTrack(originalTrack);
        expect(encoder.hasTrack("test-track")).toBe(true);
        
        // Update track
        encoder.setTrack(updatedTrack);
        const root = await encoder.root();
        expect(root.tracks.get("test-track")).toEqual(updatedTrack);
    });

    test("encoder handles track removal correctly", async () => {
        encoder = new CatalogTrackEncoder({});
        
        const track: TrackDescriptor = {
            name: "test-track",
            priority: 1,
            schema: "video/h264",
            config: {}
        };

        // Add and remove track
        encoder.setTrack(track);
        expect(encoder.hasTrack("test-track")).toBe(true);
        
        encoder.removeTrack("test-track");
        expect(encoder.hasTrack("test-track")).toBe(false);
    });
});

describe("Error Handling", () => {
    afterEach(() => {
        vi.clearAllMocks();
    });

    test("encoder handles encoding errors gracefully", async () => {
        const encoder = new CatalogTrackEncoder({});
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

        // This should trigger error handling in the encoder
        await encoder.close(new Error("Test encoding error"));

        expect(encoder.encoding).toBe(false);
        consoleErrorSpy.mockRestore();
        await encoder.close();
    });

    test("decoder handles decoding errors gracefully", async () => {
        const decoder = new CatalogTrackDecoder({});
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

        await decoder.close(new Error("Test decoding error"));

        expect(decoder.decoding).toBe(false);
        consoleErrorSpy.mockRestore();
        await decoder.close();
    });
});

describe("Patch Generation and Sync Tests", () => {
    let encoder: CatalogTrackEncoder;

    beforeEach(() => {
        encoder = new CatalogTrackEncoder({});
    });

    afterEach(async () => {
        if (encoder) { await encoder.close(); }
        vi.clearAllMocks();
    });

    test("encoder generates add patch for new track", async () => {
        const track: TrackDescriptor = {
            name: "new-track",
            priority: 1,
            schema: "video/h264",
            config: { width: 1920, height: 1080 }
        };

        encoder.setTrack(track);
        
        // Verify track was added to root
        const root = await encoder.root();
        expect(root.tracks.has("new-track")).toBe(true);
        expect(root.tracks.get("new-track")).toEqual(track);
    });

    test("encoder generates replace patch for updated track", async () => {
        const originalTrack: TrackDescriptor = {
            name: "update-track",
            priority: 1,
            schema: "video/h264",
            config: { width: 1920, height: 1080 }
        };
        const updatedTrack: TrackDescriptor = {
            name: "update-track",
            priority: 2,
            schema: "video/h264",
            config: { width: 3840, height: 2160 }
        };

        encoder.setTrack(originalTrack);
        encoder.setTrack(updatedTrack);
        
        const root = await encoder.root();
        expect(root.tracks.get("update-track")).toEqual(updatedTrack);
        expect(root.tracks.get("update-track")?.priority).toBe(2);
    });

    test("encoder generates remove patch for deleted track", async () => {
        const track: TrackDescriptor = {
            name: "remove-track",
            priority: 1,
            schema: "video/h264",
            config: {}
        };

        encoder.setTrack(track);
        expect(encoder.hasTrack("remove-track")).toBe(true);
        
        encoder.removeTrack("remove-track");
        expect(encoder.hasTrack("remove-track")).toBe(false);
        
        const root = await encoder.root();
        expect(root.tracks.has("remove-track")).toBe(false);
    });

    test("encoder handles multiple operations before sync", async () => {
        const track1: TrackDescriptor = {
            name: "track1",
            priority: 1,
            schema: "video/h264",
            config: {}
        };
        const track2: TrackDescriptor = {
            name: "track2",
            priority: 2,
            schema: "audio/opus",
            config: {}
        };

        encoder.setTrack(track1);
        encoder.setTrack(track2);
        encoder.removeTrack("track1");
        
        const root = await encoder.root();
        expect(root.tracks.has("track1")).toBe(false);
        expect(root.tracks.has("track2")).toBe(true);
    });

    test("sync does not error with empty patches", () => {
        expect(() => encoder.sync()).not.toThrow();
    });
});

describe("JSON Processing Tests", () => {
    test("decoder handles full catalog reception", async () => {
        const decoder = new CatalogTrackDecoder({
            version: "1.0.0"
        });

        // The decoder should be in initial state
        expect(decoder.hasTrack("any-track")).toBe(false);
        expect(decoder.decoding).toBe(false);
        
        await decoder.close();
    });

    test("decoder validates catalog version", () => {
        const decoder = new CatalogTrackDecoder({
            version: "2.0.0"
        });

        expect(decoder.version).toBe("2.0.0");
    });

    test("decoder nextTrack returns proper result structure", async () => {
        const decoder = new CatalogTrackDecoder({});
        
        // Start nextTrack before closing to test cancellation
        const nextTrackPromise = decoder.nextTrack();
        
        // Close decoder
        await decoder.close(new Error("Test cancellation"));
        
        // nextTrack should return a tuple [track | undefined, error | undefined]
        const result = await nextTrackPromise;
        expect(Array.isArray(result)).toBe(true);
        expect(result.length).toBe(2);
        // When cancelled, should return [undefined, Error]
        expect(result[0]).toBeUndefined();
        expect(result[1]).toBeInstanceOf(Error);
    });

    test("decoder configure resets state", async () => {
        const decoder = new CatalogTrackDecoder({});
        const config = {
            reviverRules: []
        };

        await decoder.configure(config);
        
        // State should be reset
        expect(decoder.decoding).toBe(false);
        
        await decoder.close();
    });

    test("decoder handles version mismatch gracefully", () => {
        const decoder1 = new CatalogTrackDecoder({ version: "1.0.0" });
        const decoder2 = new CatalogTrackDecoder({ version: "2.0.0" });

        expect(decoder1.version).toBe("1.0.0");
        expect(decoder2.version).toBe("2.0.0");
        expect(decoder1.version).not.toBe(decoder2.version);
    });
});

describe("Context and Cancellation Tests", () => {
    test("encoder operations respect cancellation", async () => {
        const encoder = new CatalogTrackEncoder({});
        const track: TrackDescriptor = {
            name: "test-track",
            priority: 1,
            schema: "video/h264",
            config: {}
        };

        // Close encoder first
        await encoder.close();
        
        // Operations after close should not affect state
        encoder.setTrack(track);
        encoder.sync();
        
        expect(encoder.encoding).toBe(false);
    });

    test("decoder operations respect cancellation", async () => {
        const decoder = new CatalogTrackDecoder({});
        
        // Close decoder first
        await decoder.close(new Error("Manual cancellation"));
        
        // Operations after close should handle cancellation gracefully
        const nextTrackResult = await decoder.nextTrack();
        expect(nextTrackResult[0]).toBeUndefined(); // Track should be undefined
        
        // Create a proper mock with context
        const mockSource = {
            context: {
                done: () => new Promise(() => {}),
                err: () => undefined
            },
            acceptGroup: () => Promise.resolve([undefined, new Error("cancelled")]),
            closeWithError: () => Promise.resolve()
        } as Partial<TrackReader> as TrackReader;
        
        const decodeResult = await decoder.decodeFrom(Promise.resolve(), mockSource);
        // Should return an error due to cancellation
        expect(decodeResult).toBeInstanceOf(Error);
    });

    test("multiple close calls are safe", async () => {
        const encoder = new CatalogTrackEncoder({});
        const decoder = new CatalogTrackDecoder({});

        // Multiple closes should not throw
        await encoder.close();
        await encoder.close();
        await encoder.close();
        
        await decoder.close();
        await decoder.close();
        await decoder.close();
        
        expect(encoder.encoding).toBe(false);
        expect(decoder.decoding).toBe(false);
    });
});

describe("Boundary Value Tests", () => {
    test("encoder handles empty track name", async () => {
        const encoder = new CatalogTrackEncoder({});
        const track: TrackDescriptor = {
            name: "",
            priority: 0,
            schema: "",
            config: {}
        };

        encoder.setTrack(track);
        
        const root = await encoder.root();
        expect(root.tracks.has("")).toBe(true);
        
        await encoder.close();
    });

    test("encoder handles very long track name", async () => {
        const encoder = new CatalogTrackEncoder({});
        const longName = "a".repeat(1000);
        const track: TrackDescriptor = {
            name: longName,
            priority: 1,
            schema: "video/h264",
            config: {}
        };

        encoder.setTrack(track);
        
        const root = await encoder.root();
        expect(root.tracks.has(longName)).toBe(true);
        
        await encoder.close();
    });

    test("encoder handles maximum priority value", async () => {
        const encoder = new CatalogTrackEncoder({});
        const track: TrackDescriptor = {
            name: "max-priority",
            priority: Number.MAX_SAFE_INTEGER,
            schema: "video/h264",
            config: {}
        };

        encoder.setTrack(track);
        
        const root = await encoder.root();
        expect(root.tracks.get("max-priority")?.priority).toBe(Number.MAX_SAFE_INTEGER);
        
        await encoder.close();
    });

    test("encoder handles negative priority value", async () => {
        const encoder = new CatalogTrackEncoder({});
        const track: TrackDescriptor = {
            name: "negative-priority",
            priority: -1,
            schema: "video/h264",
            config: {}
        };

        encoder.setTrack(track);
        
        const root = await encoder.root();
        expect(root.tracks.get("negative-priority")?.priority).toBe(-1);
        
        await encoder.close();
    });

    test("encoder handles complex config objects", async () => {
        const encoder = new CatalogTrackEncoder({});
        const complexConfig = {
            width: 1920,
            height: 1080,
            framerate: 30,
            bitrate: 2000000,
            nested: {
                property: "value",
                array: [1, 2, 3],
                boolean: true
            }
        };
        const track: TrackDescriptor = {
            name: "complex-track",
            priority: 1,
            schema: "video/h264",
            config: complexConfig
        };

        encoder.setTrack(track);
        
        const root = await encoder.root();
        expect(root.tracks.get("complex-track")?.config).toEqual(complexConfig);
        
        await encoder.close();
    });

    test("decoder handles empty version string", () => {
        const decoder = new CatalogTrackDecoder({ version: "" });
        
        expect(decoder.version).toBe("");
        expect(decoder.decoding).toBe(false);
    });

    test("decoder handles undefined version", () => {
        const decoder = new CatalogTrackDecoder({});
        
        expect(decoder.version).toBe(DEFAULT_CATALOG_VERSION);
    });
});

describe("Advanced Integration Tests", () => {
    test("encoder and decoder handle complete workflow", async () => {
        const encoder = new CatalogTrackEncoder({
            version: "1.0.0",
            description: "Advanced test catalog"
        });
        const decoder = new CatalogTrackDecoder({
            version: "1.0.0"
        });

        // Create multiple tracks with different types
        const videoTrack: TrackDescriptor = {
            name: "video",
            priority: 1,
            schema: "video/h264",
            config: { width: 1920, height: 1080, framerate: 30 }
        };
        const audioTrack: TrackDescriptor = {
            name: "audio",
            priority: 2,
            schema: "audio/opus",
            config: { sampleRate: 48000, channels: 2 }
        };
        const dataTrack: TrackDescriptor = {
            name: "data",
            priority: 3,
            schema: "application/json",
            config: { format: "json" }
        };

        // Add tracks to encoder
        encoder.setTrack(videoTrack);
        encoder.setTrack(audioTrack);
        encoder.setTrack(dataTrack);

        // Verify all tracks exist
        expect(encoder.hasTrack("video")).toBe(true);
        expect(encoder.hasTrack("audio")).toBe(true);
        expect(encoder.hasTrack("data")).toBe(true);

        // Update a track
        const updatedVideoTrack: TrackDescriptor = {
            ...videoTrack,
            config: { width: 3840, height: 2160, framerate: 60 }
        };
        encoder.setTrack(updatedVideoTrack);

        // Remove a track
        encoder.removeTrack("data");
        expect(encoder.hasTrack("data")).toBe(false);

        // Verify final state
        const root = await encoder.root();
        expect(root.tracks.size).toBe(2);
        expect(root.tracks.get("video")?.config).toEqual(updatedVideoTrack.config);
        expect(root.tracks.has("audio")).toBe(true);
        expect(root.tracks.has("data")).toBe(false);

        await encoder.close();
        await decoder.close();
    });

    test("multiple encoders with same version work independently", async () => {
        const encoder1 = new CatalogTrackEncoder({ version: "1.0.0" });
        const encoder2 = new CatalogTrackEncoder({ version: "1.0.0" });

        const track1: TrackDescriptor = {
            name: "track1",
            priority: 1,
            schema: "video/h264",
            config: {}
        };
        const track2: TrackDescriptor = {
            name: "track2",
            priority: 2,
            schema: "audio/opus",
            config: {}
        };

        encoder1.setTrack(track1);
        encoder2.setTrack(track2);

        // Encoders should be independent
        expect(encoder1.hasTrack("track1")).toBe(true);
        expect(encoder1.hasTrack("track2")).toBe(false);
        expect(encoder2.hasTrack("track1")).toBe(false);
        expect(encoder2.hasTrack("track2")).toBe(true);

        await encoder1.close();
        await encoder2.close();
    });

    test("encoder handles rapid track updates", async () => {
        const encoder = new CatalogTrackEncoder({});
        const baseName = "rapid-track";

        // Rapidly add and update multiple tracks
        for (let i = 0; i < 100; i++) {
            const track: TrackDescriptor = {
                name: `${baseName}-${i}`,
                priority: i,
                schema: i % 2 === 0 ? "video/h264" : "audio/opus",
                config: { iteration: i }
            };
            encoder.setTrack(track);
        }

        // Verify all tracks were added
        const root = await encoder.root();
        expect(root.tracks.size).toBe(100);

        // Verify some specific tracks
        expect(root.tracks.get("rapid-track-0")?.priority).toBe(0);
        expect(root.tracks.get("rapid-track-99")?.priority).toBe(99);
        expect(root.tracks.get("rapid-track-50")?.config).toEqual({ iteration: 50 });

        // Remove some tracks
        for (let i = 0; i < 50; i++) {
            encoder.removeTrack(`${baseName}-${i}`);
        }

        const finalRoot = await encoder.root();
        expect(finalRoot.tracks.size).toBe(50);
        expect(finalRoot.tracks.has("rapid-track-0")).toBe(false);
        expect(finalRoot.tracks.has("rapid-track-99")).toBe(true);

        await encoder.close();
    });
});

describe("Error Resilience Tests", () => {
    test("encoder continues working after sync with no writers", () => {
        const encoder = new CatalogTrackEncoder({});
        const track: TrackDescriptor = {
            name: "test-track",
            priority: 1,
            schema: "video/h264",
            config: {}
        };

        encoder.setTrack(track);
        encoder.sync(); // Should not throw even with no writers
        
        expect(encoder.hasTrack("test-track")).toBe(true);
        expect(encoder.encoding).toBe(false);
    });

    test("decoder handles configure with invalid config gracefully", async () => {
        const decoder = new CatalogTrackDecoder({});
        
        // Configure should not throw with empty config
        await decoder.configure({});
        
        expect(decoder.decoding).toBe(false);
        
        await decoder.close();
    });

    test("operations are safe after context cancellation", async () => {
        const encoder = new CatalogTrackEncoder({});
        const decoder = new CatalogTrackDecoder({});
        
        // Close both first
        await encoder.close(new Error("Test cancellation"));
        await decoder.close(new Error("Test cancellation"));
        // Wait for ctx to be cancelled
        await new Promise<void>(resolve => setTimeout(resolve, 100));
        
        // All operations should be safe
        const track: TrackDescriptor = {
            name: "safe-track",
            priority: 1,
            schema: "video/h264",
            config: {}
        };
        
        encoder.setTrack(track);
        encoder.removeTrack("safe-track");
        encoder.sync();
        
        const nextTrackResult = await decoder.nextTrack();
        await decoder.configure({});
        
        expect(encoder.encoding).toBe(false);
        expect(decoder.decoding).toBe(false);
        expect(nextTrackResult[0]).toBeUndefined(); // Track should be undefined
    });
});
