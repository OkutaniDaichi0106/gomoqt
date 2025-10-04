import { VideoConfigSchema, VideoTrackSchema } from "./video";
import { describe, expect, test } from "vitest";

const baseConfig = {
    codec: "avc1.640028",
    container: "cmaf" as const,
};

describe("Video descriptors", () => {
    describe("VideoConfigSchema", () => {
        test("parses valid config with hex string description and defaults", () => {
            const config = VideoConfigSchema.parse({
                ...baseConfig,
                description: "00010203",
                codedWidth: 1920,
                codedHeight: 1080,
                framerate: 60,
                bitrate: 4_000_000,
            });

            expect(config.description).toBeInstanceOf(Uint8Array);
            expect(Array.from(config.description!)).toEqual([0, 1, 2, 3]);
            expect(config.optimizeForLatency).toBe(true); // default applied
            expect(config.rotation).toBe(0); // default applied
            expect(config.flip).toBe(false); // default applied
        });

        test("keeps Uint8Array description untouched", () => {
            const description = new Uint8Array([255, 0, 127]);
            const config = VideoConfigSchema.parse({
                ...baseConfig,
                description,
            });

            expect(config.description).toBe(description);
        });

        test("rejects invalid hex string in description", () => {
            expect(() =>
                VideoConfigSchema.parse({
                    ...baseConfig,
                    description: "not-hex",
                })
            ).toThrow("Invalid hex string");
        });

        test("rejects invalid container values", () => {
            const result = VideoConfigSchema.safeParse({
                ...baseConfig,
                container: "mp4",
            });

            expect(result.success).toBe(false);
            if (!result.success) {
                expect(result.error.issues[0]?.message).toContain("Invalid option");
                expect(result.error.issues[0]?.path).toEqual(["container"]);
            }
        });
    });

    describe("VideoTrackSchema", () => {
        test("parses complete video track descriptor", () => {
            const descriptor = VideoTrackSchema.parse({
                name: "video-main",
                priority: 0,
                schema: "video",
                config: {
                    ...baseConfig,
                    description: "cafebabe",
                    codedWidth: 1280,
                    codedHeight: 720,
                    displayAspectWidth: 16,
                    displayAspectHeight: 9,
                    framerate: 30,
                    optimizeForLatency: false,
                    rotation: 90,
                    flip: true,
                },
            });

            expect(descriptor.schema).toBe("video");
            expect(Array.from(descriptor.config.description!)).toEqual([202, 254, 186, 190]);
            expect(descriptor.config.optimizeForLatency).toBe(false);
            expect(descriptor.config.rotation).toBe(90);
            expect(descriptor.config.flip).toBe(true);
        });

        test("enforces video schema literal", () => {
            const result = VideoTrackSchema.safeParse({
                name: "video-secondary",
                priority: 5,
                schema: "audio",
                config: {
                    ...baseConfig,
                },
            });

            expect(result.success).toBe(false);
            if (!result.success) {
                expect(result.error.issues[0]?.path).toEqual(["schema"]);
            }
        });
    });
});
