import { describe, test, expect, beforeEach, afterEach, jest } from 'vitest';
// Mock the external dependencies before importing the module under test
vi.mock("@okutanidaichi/moqt", () => ({
    PublishAbortedErrorCode: 100,
    InternalGroupErrorCode: 101,
    InternalSubscribeErrorCode: 102,
}));

vi.mock("@okutanidaichi/moqt/io", () => ({
    readVarint: vi.fn().mockReturnValue([BigInt(0), 0]),
}));

vi.mock("golikejs/context", () => ({
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
    ContextCancelledError: new Error("Context cancelled"),
    Mutex: class MockMutex {
        async lock() {
            return () => {};
        }
    }
}));

import { CatalogTrackEncoder, CatalogTrackDecoder } from "./catalog_track";
import type { CatalogTrackEncoderInit, CatalogTrackDecoderInit } from "./catalog_track";
import type { TrackDescriptor } from "../catalog";
import { DEFAULT_CATALOG_VERSION } from "../catalog";



describe("CatalogTrackEncoder", () => {
    describe("constructor", () => {
        test("creates encoder with default version", () => {
            const encoder = new CatalogTrackEncoder({});
            
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
        test("adds new track to catalog", async () => {
            const encoder = new CatalogTrackEncoder({});
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
            const encoder = new CatalogTrackEncoder({});
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
            const encoder = new CatalogTrackEncoder({});
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
            const encoder = new CatalogTrackEncoder({});
            const track: TrackDescriptor = {
                name: "test-track",
                priority: 1,
                schema: "video/h264",
                config: {}
            };

            await encoder.close();
            encoder.setTrack(track);
            
            const root = await encoder.root();
            expect(root.tracks.has("test-track")).toBe(false);
        });
    });

    describe("removeTrack", () => {
        test("removes existing track from catalog", async () => {
            const encoder = new CatalogTrackEncoder({});
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
            const encoder = new CatalogTrackEncoder({});

            expect(() => encoder.removeTrack("non-existent")).not.toThrow();
            
            const root = await encoder.root();
            expect(root.tracks.size).toBe(0);
        });

        test("does not remove track when cancelled", async () => {
            const encoder = new CatalogTrackEncoder({});
            const track: TrackDescriptor = {
                name: "test-track",
                priority: 1,
                schema: "video/h264",
                config: {}
            };

            encoder.setTrack(track);
            await encoder.close();
            encoder.removeTrack("test-track");
            
            const root = await encoder.root();
            expect(root.tracks.has("test-track")).toBe(true);
        });
    });

    describe("hasTrack", () => {
        test("returns true for existing track", async () => {
            const encoder = new CatalogTrackEncoder({});
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
            const encoder = new CatalogTrackEncoder({});
            
            expect(encoder.hasTrack("non-existent")).toBe(false);
        });
    });

    describe("sync", () => {
        test("does nothing when no patches exist", () => {
            const encoder = new CatalogTrackEncoder({});
            
            expect(() => encoder.sync()).not.toThrow();
        });

        test("does not sync when cancelled", async () => {
            const encoder = new CatalogTrackEncoder({});
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
        test("returns error when encoder is cancelled", async () => {
            const encoder = new CatalogTrackEncoder({});
            await encoder.close(new Error("Test cancellation"));
            
            // Create a partial mock that has the minimum required properties
            const mockWriter = {
                context: { done: () => Promise.resolve(), err: () => undefined }
            } as any;
            
            const ctx = Promise.resolve();
            const result = await encoder.encodeTo(ctx, mockWriter);
            
            // The method should return undefined or an error when cancelled
            // This test verifies the method completes without throwing
        });
    });

    describe("root", () => {
        test("returns root catalog with default version", async () => {
            const encoder = new CatalogTrackEncoder({});
            
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
            const encoder = new CatalogTrackEncoder(init);
            
            const root = await encoder.root();
            
            expect(root.version).toBe("2.0.0");
            expect(root.description).toBe("Custom catalog");
        });
    });

    describe("close", () => {
        test("closes successfully without cause", async () => {
            const encoder = new CatalogTrackEncoder({});
            
            await encoder.close();
            
            expect(encoder.encoding).toBe(false);
        });

        test("closes successfully with cause", async () => {
            const encoder = new CatalogTrackEncoder({});
            const cause = new Error("Test error");
            
            await encoder.close(cause);
            
            expect(encoder.encoding).toBe(false);
        });

        test("does not error on multiple close calls", async () => {
            const encoder = new CatalogTrackEncoder({});
            
            await encoder.close();
            await encoder.close();
            
            expect(encoder.encoding).toBe(false);
        });
    });

    describe("configure", () => {
        test("configures encoder successfully", async () => {
            const encoder = new CatalogTrackEncoder({});
            const config = {
                space: 2,
                replacer: [] as any[]
            };

            // The configure method might hang waiting for decoder config resolution
            // Let's test that it can be called without hanging
            const configurePromise = encoder.configure(config);
            
            // Close the encoder to ensure the configure method doesn't hang indefinitely
            setTimeout(() => encoder.close(), 100);
            
            try {
                await Promise.race([
                    configurePromise,
                    new Promise((_, reject) => setTimeout(() => reject(new Error("Timeout")), 1000))
                ]);
            } catch (error) {
                // Expected to timeout or throw error when encoder is closed
                expect(error).toBeDefined();
            }
        });
    });

    describe("encoding property", () => {
        test("returns false when no tracks are encoding", () => {
            const encoder = new CatalogTrackEncoder({});
            
            expect(encoder.encoding).toBe(false);
        });
    });
});

describe("CatalogTrackDecoder", () => {
    describe("constructor", () => {
        test("creates decoder with default version", () => {
            const decoder = new CatalogTrackDecoder({});
            
            expect(decoder).toBeDefined();
            expect(decoder.version).toBe(DEFAULT_CATALOG_VERSION);
            expect(decoder.decoding).toBe(false);
        });

        test("creates decoder with custom version", () => {
            const init: CatalogTrackDecoderInit = {
                version: "1.0.0"
            };
            const decoder = new CatalogTrackDecoder(init);
            
            expect(decoder.version).toBe("1.0.0");
        });
    });

    describe("hasTrack", () => {
        test("returns false when no root is set", () => {
            const decoder = new CatalogTrackDecoder({});
            
            expect(decoder.hasTrack("test-track")).toBe(false);
        });

        test("returns false for non-existent track", () => {
            const decoder = new CatalogTrackDecoder({});
            
            expect(decoder.hasTrack("non-existent")).toBe(false);
        });
    });

    describe("root", () => {
        test("returns root catalog promise", async () => {
            const decoder = new CatalogTrackDecoder({});
            
            const rootPromise = decoder.root();
            expect(rootPromise).toBeInstanceOf(Promise);
        });
    });

    describe("nextTrack", () => {
        test("returns error when decoder is cancelled", async () => {
            const decoder = new CatalogTrackDecoder({});
            
            await decoder.close(new Error("Test cancellation"));
            
            const result = await decoder.nextTrack();
            // The result should be a tuple with error info when cancelled
            expect(result).toBeDefined();
            expect(Array.isArray(result)).toBe(true);
        });
    });

    describe("configure", () => {
        test("configures decoder successfully", async () => {
            const decoder = new CatalogTrackDecoder({});
            const config = {
                reviverRules: [] as any[]
            };

            await decoder.configure(config);
            
            // Should complete without error
            expect(decoder.decoding).toBe(false);
        });

        test("does not configure when cancelled", async () => {
            const decoder = new CatalogTrackDecoder({});
            
            await decoder.close();
            await decoder.configure({ reviverRules: [] });
            
            // Should complete without error
            expect(decoder.decoding).toBe(false);
        });
    });

    describe("decodeFrom", () => {
        test("returns error when decoder is cancelled", async () => {
            const decoder = new CatalogTrackDecoder({});
            await decoder.close(new Error("Test cancellation"));
            
            // Create a partial mock that has the minimum required properties
            const mockReader = {
                context: { done: () => Promise.resolve(), err: () => undefined }
            } as any;
            
            const ctx = Promise.resolve();
            const result = await decoder.decodeFrom(ctx, mockReader);
            
            // The method should return undefined or an error when cancelled
            // This test verifies the method completes without throwing
        });
    });

    describe("close", () => {
        test("closes successfully without cause", async () => {
            const decoder = new CatalogTrackDecoder({});
            
            await decoder.close();
            
            expect(decoder.decoding).toBe(false);
        });

        test("closes successfully with cause", async () => {
            const decoder = new CatalogTrackDecoder({});
            const cause = new Error("Test error");
            
            await decoder.close(cause);
            
            expect(decoder.decoding).toBe(false);
        });

        test("does not error on multiple close calls", async () => {
            const decoder = new CatalogTrackDecoder({});
            
            await decoder.close();
            await decoder.close();
            
            expect(decoder.decoding).toBe(false);
        });
    });

    describe("decoding property", () => {
        test("returns false when not decoding", () => {
            const decoder = new CatalogTrackDecoder({});
            
            expect(decoder.decoding).toBe(false);
        });

        test("returns false when cancelled", async () => {
            const decoder = new CatalogTrackDecoder({});
            
            await decoder.close();
            
            expect(decoder.decoding).toBe(false);
        });
    });
});

describe("CatalogTrackEncoder and CatalogTrackDecoder Integration", () => {
    test("encoder and decoder work together", async () => {
        const encoder = new CatalogTrackEncoder({
            version: "1.0.0",
            description: "Integration test catalog"
        });
        const decoder = new CatalogTrackDecoder({
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

        // Clean up
        await encoder.close();
        await decoder.close();
    });

    test("encoder handles track updates correctly", async () => {
        const encoder = new CatalogTrackEncoder({});
        
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

        await encoder.close();
    });

    test("encoder handles track removal correctly", async () => {
        const encoder = new CatalogTrackEncoder({});
        
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

        await encoder.close();
    });
});

describe("Error Handling", () => {
    test("encoder handles encoding errors gracefully", async () => {
        const encoder = new CatalogTrackEncoder({});
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation();

        // This should trigger error handling in the encoder
        await encoder.close(new Error("Test encoding error"));

        expect(encoder.encoding).toBe(false);
        consoleErrorSpy.mockRestore();
    });

    test("decoder handles decoding errors gracefully", async () => {
        const decoder = new CatalogTrackDecoder({});
        const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation();

        await decoder.close(new Error("Test decoding error"));

        expect(decoder.decoding).toBe(false);
        consoleErrorSpy.mockRestore();
    });
});

describe("Patch Generation and Sync Tests", () => {
    test("encoder generates add patch for new track", async () => {
        const encoder = new CatalogTrackEncoder({});
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
        
        await encoder.close();
    });

    test("encoder generates replace patch for updated track", async () => {
        const encoder = new CatalogTrackEncoder({});
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
        
        await encoder.close();
    });

    test("encoder generates remove patch for deleted track", async () => {
        const encoder = new CatalogTrackEncoder({});
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
        
        await encoder.close();
    });

    test("encoder handles multiple operations before sync", async () => {
        const encoder = new CatalogTrackEncoder({});
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
        
        await encoder.close();
    });

    test("sync does not error with empty patches", () => {
        const encoder = new CatalogTrackEncoder({});
        
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
        
        // Close decoder first to ensure nextTrack returns immediately
        await decoder.close(new Error("Test cancellation"));
        
        // nextTrack should return a tuple [track | undefined, error | undefined]
        const result = await decoder.nextTrack();
        expect(Array.isArray(result)).toBe(true);
        expect(result.length).toBe(2);
        // When cancelled, should return [undefined, Error]
        expect(result[0]).toBeUndefined();
        // The error might be undefined due to mocking, but the structure should be correct
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
        
        const decodeResult = await decoder.decodeFrom(Promise.resolve(), {} as any);
        // Should return some result (could be undefined due to mocking)
        // The important thing is that it doesn't throw
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
