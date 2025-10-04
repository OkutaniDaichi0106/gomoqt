import { AudioConfigSchema, AudioTrackSchema } from "./audio";
import { describe, expect, test } from "vitest";

const baseConfig = {
    codec: "opus",
    sampleRate: 48_000,
    numberOfChannels: 2,
    container: "loc" as const,
};

describe("Audio descriptors", () => {
    describe("AudioConfigSchema", () => {
        test("parses valid config with hex string description", () => {
            const result = AudioConfigSchema.parse({
                ...baseConfig,
                description: "48656c6c6f", // "Hello" in hex
                bitrate: 96_000,
            });

            expect(result.codec).toBe("opus");
            expect(result.description).toBeInstanceOf(Uint8Array);
            expect(Array.from(result.description!)).toEqual([72, 101, 108, 108, 111]);
            expect(result.bitrate).toBe(96_000);
        });

        test("parses config with Uint8Array description", () => {
            const description = new Uint8Array([1, 2, 3]);
            const result = AudioConfigSchema.parse({
                ...baseConfig,
                description,
            });

            expect(result.description).toBe(description);
        });

        test("rejects invalid hex string description", () => {
            expect(() =>
                AudioConfigSchema.parse({
                    ...baseConfig,
                    description: "invalid-hex",
                })
            ).toThrow("Invalid hex string");
        });

        test("rejects invalid container values", () => {
            const result = AudioConfigSchema.safeParse({
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

    describe("AudioTrackSchema", () => {
        test("parses complete audio track descriptor", () => {
            const descriptor = AudioTrackSchema.parse({
                name: "audio-main",
                priority: 1,
                schema: "audio",
                config: {
                    ...baseConfig,
                    description: "001122",
                },
                dependencies: ["catalog"],
            });

            expect(descriptor.schema).toBe("audio");
            expect(descriptor.config.description).toBeInstanceOf(Uint8Array);
            expect(Array.from(descriptor.config.description!)).toEqual([0, 17, 34]);
            expect(descriptor.dependencies).toEqual(["catalog"]);
        });

        test("enforces audio schema literal", () => {
            const result = AudioTrackSchema.safeParse({
                name: "audio-secondary",
                priority: 2,
                schema: "video",
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
