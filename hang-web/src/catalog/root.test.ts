import { RootSchema, DEFAULT_CATALOG_VERSION } from "./root";
import { TrackSchema } from "./track";
import { describe, test, expect, beforeEach, it } from 'vitest';

describe("RootSchema", () => {
    describe("valid cases", () => {
        it("should parse minimal valid input", () => {
            const input = {
                tracks: new Map(),
            };

            const result = RootSchema.parse(input);

            expect(result.version).toBe(DEFAULT_CATALOG_VERSION);
            expect(result.description).toBeUndefined();
            expect(result.tracks).toEqual(new Map());
        });

        it("should parse full valid input", () => {
            const track = TrackSchema.parse({
                name: "test-track",
                priority: 1,
                schema: "test-schema",
                config: {},
            });

            const input = {
                version: "2",
                description: "Test catalog",
                tracks: new Map([["test-track", track]]),
            };

            const result = RootSchema.parse(input);

            expect(result.version).toBe("2");
            expect(result.description).toBe("Test catalog");
            expect(result.tracks).toEqual(new Map([["test-track", track]]));
        });

        it("should apply default version when not provided", () => {
            const input = {
                description: "Test",
                tracks: new Map(),
            };

            const result = RootSchema.parse(input);

            expect(result.version).toBe(DEFAULT_CATALOG_VERSION);
        });
    });

    describe("invalid cases", () => {
        it("should reject missing tracks", () => {
            const input = {};

            expect(() => RootSchema.parse(input)).toThrow();
        });

        it("should reject invalid tracks type", () => {
            const input = {
                tracks: "invalid",
            };

            expect(() => RootSchema.parse(input)).toThrow();
        });

        it("should reject description too long", () => {
            const input = {
                tracks: new Map(),
                description: "a".repeat(501),
            };

            expect(() => RootSchema.parse(input)).toThrow();
        });
    });
});
