import { describe, test, expect, beforeEach, afterEach, vi } from 'vitest';
import {
    JsonEncoder,
    JsonDecoder,
    EncodedJsonChunk,
    replaceBigInt,
    reviveBigInt,
    replaceDate,
    reviveDate,
    JSON_RULES
} from "./json";
import type {
    JsonEncoderInit,
    JsonEncoderConfig,
    JsonDecoderInit,
    JsonDecoderConfig,
    EncodedJsonChunkInit,
    JsonRuleName
} from "./json";
import type { JsonValue, JsonPatch } from "./json_patch";

describe("JsonEncoder", () => {
    describe("constructor", () => {
        test("creates encoder with output and error handlers", () => {
            const outputSpy = vi.fn();
            const errorSpy = vi.fn();
            const init: JsonEncoderInit = {
                output: outputSpy,
                error: errorSpy
            };

            const encoder = new JsonEncoder(init);

            expect(encoder).toBeDefined();
        });
    });

    describe("configure", () => {
        test("configures encoder with space setting", () => {
            const outputSpy = vi.fn();
            const errorSpy = vi.fn();
            const encoder = new JsonEncoder({ output: outputSpy, error: errorSpy });
            const config: JsonEncoderConfig = {
                space: 2
            };

            expect(() => encoder.configure(config)).not.toThrow();
        });

        test("configures encoder with replacer rules", () => {
            const outputSpy = vi.fn();
            const errorSpy = vi.fn();
            const encoder = new JsonEncoder({ output: outputSpy, error: errorSpy });
            const config: JsonEncoderConfig = {
                replacer: ["bigint", "date"]
            };

            expect(() => encoder.configure(config)).not.toThrow();
        });

        test("configures encoder with both space and replacer", () => {
            const outputSpy = vi.fn();
            const errorSpy = vi.fn();
            const encoder = new JsonEncoder({ output: outputSpy, error: errorSpy });
            const config: JsonEncoderConfig = {
                space: 4,
                replacer: ["bigint"]
            };

            expect(() => encoder.configure(config)).not.toThrow();
        });
    });

    describe("encode", () => {
        test("encodes simple JSON value", () => {
            const outputSpy = vi.fn();
            const errorSpy = vi.fn();
            const encoder = new JsonEncoder({ output: outputSpy, error: errorSpy });

            const value: JsonValue = { test: "value" };
            encoder.encode(value);

            expect(outputSpy).toHaveBeenCalledTimes(1);
            const [chunk, metadata] = outputSpy.mock.calls[0];
            expect(chunk).toBeInstanceOf(EncodedJsonChunk);
            expect(chunk.type).toBe("key");
            expect(chunk.data).toBeInstanceOf(Uint8Array);
            expect(chunk.timestamp).toBeGreaterThan(0);
        });

        test("encodes JSON patch array", () => {
            const outputSpy = vi.fn();
            const errorSpy = vi.fn();
            const encoder = new JsonEncoder({ output: outputSpy, error: errorSpy });

            const patch: JsonPatch = [
                { op: "add", path: "/test", value: "value" }
            ];
            encoder.encode(patch);

            expect(outputSpy).toHaveBeenCalledTimes(1);
            const [chunk] = outputSpy.mock.calls[0];
            expect(chunk.type).toBe("delta");
        });

        test("encodes with bigint replacer", () => {
            const outputSpy = vi.fn();
            const errorSpy = vi.fn();
            const encoder = new JsonEncoder({ output: outputSpy, error: errorSpy });
            encoder.configure({ replacer: ["bigint"] });

            const value = { bigNumber: BigInt(123456789) };
            encoder.encode(value);

            expect(outputSpy).toHaveBeenCalledTimes(1);
            const [chunk] = outputSpy.mock.calls[0];
            
            // Decode the chunk to verify bigint was converted to string
            const decoded = new TextDecoder().decode(chunk.data);
            const parsed = JSON.parse(decoded);
            expect(parsed.bigNumber).toBe("123456789");
        });

        test("encodes with date replacer", () => {
            const outputSpy = vi.fn();
            const errorSpy = vi.fn();
            const encoder = new JsonEncoder({ output: outputSpy, error: errorSpy });
            encoder.configure({ replacer: ["date"] });

            const testDate = new Date("2023-01-01T00:00:00.000Z");
            const value = { createdAt: testDate };
            encoder.encode(value);

            expect(outputSpy).toHaveBeenCalledTimes(1);
            const [chunk] = outputSpy.mock.calls[0];
            
            // Decode the chunk to verify date was converted to ISO string
            const decoded = new TextDecoder().decode(chunk.data);
            const parsed = JSON.parse(decoded);
            expect(parsed.createdAt).toBe("2023-01-01T00:00:00.000Z");
        });

        test("includes metadata with decoder config on first encode", () => {
            const outputSpy = vi.fn();
            const errorSpy = vi.fn();
            const encoder = new JsonEncoder({ output: outputSpy, error: errorSpy });
            encoder.configure({ replacer: ["bigint"] });

            encoder.encode({ test: "value" });

            expect(outputSpy).toHaveBeenCalledTimes(1);
            const [, metadata] = outputSpy.mock.calls[0];
            expect(metadata).toBeDefined();
            expect(metadata.decoderConfig).toBeDefined();
            expect(metadata.decoderConfig.reviverRules).toEqual(["bigint"]);
        });

        test("does not include metadata on subsequent encodes", () => {
            const outputSpy = vi.fn();
            const errorSpy = vi.fn();
            const encoder = new JsonEncoder({ output: outputSpy, error: errorSpy });
            encoder.configure({ replacer: ["bigint"] });

            encoder.encode({ test: "value1" });
            encoder.encode({ test: "value2" });

            expect(outputSpy).toHaveBeenCalledTimes(2);
            const [, metadata2] = outputSpy.mock.calls[1];
            expect(metadata2).toBeUndefined();
        });
    });

    describe("close", () => {
        test("closes successfully", () => {
            const outputSpy = vi.fn();
            const errorSpy = vi.fn();
            const encoder = new JsonEncoder({ output: outputSpy, error: errorSpy });

            expect(() => encoder.close()).not.toThrow();
        });

        test("can be called multiple times", () => {
            const outputSpy = vi.fn();
            const errorSpy = vi.fn();
            const encoder = new JsonEncoder({ output: outputSpy, error: errorSpy });

            encoder.close();
            expect(() => encoder.close()).not.toThrow();
        });
    });
});

describe("JsonDecoder", () => {
    describe("constructor", () => {
        test("creates decoder with output and error handlers", () => {
            const outputSpy = vi.fn();
            const errorSpy = vi.fn();
            const init: JsonDecoderInit = {
                output: outputSpy,
                error: errorSpy
            };

            const decoder = new JsonDecoder(init);

            expect(decoder).toBeDefined();
        });
    });

    describe("configure", () => {
        test("configures decoder with reviver rules", () => {
            const outputSpy = vi.fn();
            const errorSpy = vi.fn();
            const decoder = new JsonDecoder({ output: outputSpy, error: errorSpy });
            const config: JsonDecoderConfig = {
                reviverRules: ["bigint", "date"]
            };

            expect(() => decoder.configure(config)).not.toThrow();
        });

        test("configures decoder with empty reviver rules", () => {
            const outputSpy = vi.fn();
            const errorSpy = vi.fn();
            const decoder = new JsonDecoder({ output: outputSpy, error: errorSpy });
            const config: JsonDecoderConfig = {
                reviverRules: []
            };

            expect(() => decoder.configure(config)).not.toThrow();
        });
    });

    describe("decode", () => {
        test("decodes simple JSON chunk", () => {
            const outputSpy = vi.fn();
            const errorSpy = vi.fn();
            const decoder = new JsonDecoder({ output: outputSpy, error: errorSpy });

            const jsonString = JSON.stringify({ test: "value" });
            const data = new TextEncoder().encode(jsonString);
            const chunk = new EncodedJsonChunk({
                type: "key",
                data,
                timestamp: Date.now()
            });

            decoder.decode(chunk);

            expect(outputSpy).toHaveBeenCalledTimes(1);
            expect(outputSpy).toHaveBeenCalledWith({ test: "value" });
            expect(errorSpy).not.toHaveBeenCalled();
        });

        test("decodes JSON patch array", () => {
            const outputSpy = vi.fn();
            const errorSpy = vi.fn();
            const decoder = new JsonDecoder({ output: outputSpy, error: errorSpy });

            const patch = [{ op: "add", path: "/test", value: "value" }];
            const jsonString = JSON.stringify(patch);
            const data = new TextEncoder().encode(jsonString);
            const chunk = new EncodedJsonChunk({
                type: "delta",
                data,
                timestamp: Date.now()
            });

            decoder.decode(chunk);

            expect(outputSpy).toHaveBeenCalledTimes(1);
            expect(outputSpy).toHaveBeenCalledWith(patch);
        });

        test("decodes with bigint reviver", () => {
            const outputSpy = vi.fn();
            const errorSpy = vi.fn();
            const decoder = new JsonDecoder({ output: outputSpy, error: errorSpy });
            decoder.configure({ reviverRules: ["bigint"] });

            const jsonString = JSON.stringify({ bigNumber: "123456789" });
            const data = new TextEncoder().encode(jsonString);
            const chunk = new EncodedJsonChunk({
                type: "key",
                data,
                timestamp: Date.now()
            });

            decoder.decode(chunk);

            expect(outputSpy).toHaveBeenCalledTimes(1);
            const [result] = outputSpy.mock.calls[0];
            expect(result.bigNumber).toBe(BigInt(123456789));
        });

        test("decodes with date reviver", () => {
            const outputSpy = vi.fn();
            const errorSpy = vi.fn();
            const decoder = new JsonDecoder({ output: outputSpy, error: errorSpy });
            decoder.configure({ reviverRules: ["date"] });

            const jsonString = JSON.stringify({ createdAt: "2023-01-01T00:00:00.000Z" });
            const data = new TextEncoder().encode(jsonString);
            const chunk = new EncodedJsonChunk({
                type: "key",
                data,
                timestamp: Date.now()
            });

            decoder.decode(chunk);

            expect(outputSpy).toHaveBeenCalledTimes(1);
            const [result] = outputSpy.mock.calls[0];
            expect(result.createdAt).toBeInstanceOf(Date);
            expect(result.createdAt.toISOString()).toBe("2023-01-01T00:00:00.000Z");
        });

        test("handles invalid JSON gracefully", () => {
            const outputSpy = vi.fn();
            const errorSpy = vi.fn();
            const decoder = new JsonDecoder({ output: outputSpy, error: errorSpy });

            const invalidJson = "{ invalid json";
            const data = new TextEncoder().encode(invalidJson);
            const chunk = new EncodedJsonChunk({
                type: "key",
                data,
                timestamp: Date.now()
            });

            decoder.decode(chunk);

            expect(outputSpy).not.toHaveBeenCalled();
            expect(errorSpy).toHaveBeenCalledTimes(1);
            expect(errorSpy.mock.calls[0][0]).toBeInstanceOf(Error);
        });

        test("handles non-Error exceptions", () => {
            const outputSpy = vi.fn();
            const errorSpy = vi.fn();
            
            // Mock JSON.parse to throw a string instead of Error
            const originalParse = JSON.parse;
            JSON.parse = vi.fn().mockImplementation(() => {
                throw "String error";
            });

            const decoder = new JsonDecoder({ output: outputSpy, error: errorSpy });

            const data = new TextEncoder().encode("{}");
            const chunk = new EncodedJsonChunk({
                type: "key",
                data,
                timestamp: Date.now()
            });

            decoder.decode(chunk);

            expect(errorSpy).toHaveBeenCalledTimes(1);
            expect(errorSpy.mock.calls[0][0]).toBeInstanceOf(Error);
            expect(errorSpy.mock.calls[0][0].message).toBe("String error");

            // Restore original JSON.parse
            JSON.parse = originalParse;
        });
    });

    describe("close", () => {
        test("closes successfully", () => {
            const outputSpy = vi.fn();
            const errorSpy = vi.fn();
            const decoder = new JsonDecoder({ output: outputSpy, error: errorSpy });

            expect(() => decoder.close()).not.toThrow();
        });
    });
});

describe("EncodedJsonChunk", () => {
    describe("constructor", () => {
        test("creates chunk with required properties", () => {
            const init: EncodedJsonChunkInit = {
                type: "key",
                data: new Uint8Array([1, 2, 3]),
                timestamp: 123456789
            };

            const chunk = new EncodedJsonChunk(init);

            expect(chunk.type).toBe("key");
            expect(chunk.data).toEqual(new Uint8Array([1, 2, 3]));
            expect(chunk.timestamp).toBe(123456789);
        });

        test("creates delta chunk", () => {
            const init: EncodedJsonChunkInit = {
                type: "delta",
                data: new Uint8Array([4, 5, 6]),
                timestamp: 987654321
            };

            const chunk = new EncodedJsonChunk(init);

            expect(chunk.type).toBe("delta");
        });
    });

    describe("byteLength", () => {
        test("returns correct byte length", () => {
            const data = new Uint8Array([1, 2, 3, 4, 5]);
            const chunk = new EncodedJsonChunk({
                type: "key",
                data,
                timestamp: Date.now()
            });

            expect(chunk.byteLength).toBe(5);
        });

        test("returns zero for empty data", () => {
            const chunk = new EncodedJsonChunk({
                type: "key",
                data: new Uint8Array(0),
                timestamp: Date.now()
            });

            expect(chunk.byteLength).toBe(0);
        });
    });

    describe("copyTo", () => {
        test("copies data to target array", () => {
            const sourceData = new Uint8Array([1, 2, 3]);
            const chunk = new EncodedJsonChunk({
                type: "key",
                data: sourceData,
                timestamp: Date.now()
            });

            const target = new Uint8Array(5);
            chunk.copyTo(target);

            expect(target.subarray(0, 3)).toEqual(sourceData);
        });

        test("copies data to exact size target", () => {
            const sourceData = new Uint8Array([1, 2, 3]);
            const chunk = new EncodedJsonChunk({
                type: "key",
                data: sourceData,
                timestamp: Date.now()
            });

            const target = new Uint8Array(3);
            chunk.copyTo(target);

            expect(target).toEqual(sourceData);
        });
    });
});

describe("JSON Rule Functions", () => {
    describe("replaceBigInt", () => {
        test("converts bigint to string", () => {
            const result = replaceBigInt("key", BigInt(123456789));
            expect(result).toBe("123456789");
        });

        test("leaves non-bigint values unchanged", () => {
            expect(replaceBigInt("key", 123)).toBe(123);
            expect(replaceBigInt("key", "string")).toBe("string");
            expect(replaceBigInt("key", null)).toBe(null);
            expect(replaceBigInt("key", undefined)).toBe(undefined);
        });

        test("handles zero bigint", () => {
            const result = replaceBigInt("key", BigInt(0));
            expect(result).toBe("0");
        });

        test("handles negative bigint", () => {
            const result = replaceBigInt("key", BigInt(-123));
            expect(result).toBe("-123");
        });
    });

    describe("reviveBigInt", () => {
        test("converts numeric string to bigint", () => {
            const result = reviveBigInt("key", "123456789");
            expect(result).toBe(BigInt(123456789));
        });

        test("leaves non-numeric strings unchanged", () => {
            expect(reviveBigInt("key", "abc")).toBe("abc");
            expect(reviveBigInt("key", "123abc")).toBe("123abc");
            expect(reviveBigInt("key", "")).toBe("");
        });

        test("leaves non-string values unchanged", () => {
            expect(reviveBigInt("key", 123)).toBe(123);
            expect(reviveBigInt("key", null)).toBe(null);
            expect(reviveBigInt("key", undefined)).toBe(undefined);
        });

        test("handles zero string", () => {
            const result = reviveBigInt("key", "0");
            expect(result).toBe(BigInt(0));
        });

        test("handles invalid bigint gracefully", () => {
            // Mock BigInt to throw an error
            const originalBigInt = global.BigInt;
            (global as any).BigInt = vi.fn().mockImplementation(() => {
                throw new Error("Invalid BigInt");
            });

            const result = reviveBigInt("key", "123");
            expect(result).toBe("123");

            // Restore original BigInt
            global.BigInt = originalBigInt;
        });
    });

    describe("replaceDate", () => {
        test("converts date to ISO string", () => {
            const date = new Date("2023-01-01T00:00:00.000Z");
            const result = replaceDate("key", date);
            expect(result).toBe("2023-01-01T00:00:00.000Z");
        });

        test("leaves non-date values unchanged", () => {
            expect(replaceDate("key", "string")).toBe("string");
            expect(replaceDate("key", 123)).toBe(123);
            expect(replaceDate("key", null)).toBe(null);
        });
    });

    describe("reviveDate", () => {
        test("converts ISO string to date", () => {
            const result = reviveDate("key", "2023-01-01T00:00:00.000Z");
            expect(result).toBeInstanceOf(Date);
            expect(result.toISOString()).toBe("2023-01-01T00:00:00.000Z");
        });

        test("leaves invalid date strings unchanged", () => {
            expect(reviveDate("key", "invalid")).toBe("invalid");
            expect(reviveDate("key", "2023-01-01")).toBe("2023-01-01");
        });

        test("leaves non-string values unchanged", () => {
            expect(reviveDate("key", 123)).toBe(123);
            expect(reviveDate("key", null)).toBe(null);
        });

        test("handles invalid date construction gracefully", () => {
            // Mock Date constructor to throw
            const originalDate = global.Date;
            global.Date = vi.fn().mockImplementation(() => {
                throw new Error("Invalid Date");
            }) as any;

            const result = reviveDate("key", "2023-01-01T00:00:00.000Z");
            expect(result).toBe("2023-01-01T00:00:00.000Z");

            // Restore original Date
            global.Date = originalDate;
        });
    });
});

describe("JSON_RULES", () => {
    test("contains bigint rule", () => {
        expect(JSON_RULES.bigint).toBeDefined();
        expect(JSON_RULES.bigint.replacer).toBe(replaceBigInt);
        expect(JSON_RULES.bigint.reviver).toBe(reviveBigInt);
    });

    test("contains date rule", () => {
        expect(JSON_RULES.date).toBeDefined();
        expect(JSON_RULES.date.replacer).toBe(replaceDate);
        expect(JSON_RULES.date.reviver).toBe(reviveDate);
    });

    test("has correct rule names", () => {
        const ruleNames: JsonRuleName[] = ["bigint", "date"];
        ruleNames.forEach(name => {
            expect(JSON_RULES[name]).toBeDefined();
        });
    });
});

describe("Buffer Management Tests", () => {
    test("encoder efficiently reuses buffer for small data", () => {
        const outputSpy = vi.fn();
        const errorSpy = vi.fn();
        const encoder = new JsonEncoder({ output: outputSpy, error: errorSpy });

        // Encode multiple small objects
        encoder.encode({ test: 1 });
        encoder.encode({ test: 2 });
        encoder.encode({ test: 3 });

        expect(outputSpy).toHaveBeenCalledTimes(3);
        // All should succeed without errors
        expect(errorSpy).not.toHaveBeenCalled();
    });

    test("encoder expands buffer for large data", () => {
        const outputSpy = vi.fn();
        const errorSpy = vi.fn();
        const encoder = new JsonEncoder({ output: outputSpy, error: errorSpy });

        // Create a large object that would exceed initial buffer
        const largeObject = {
            data: "x".repeat(2000),
            array: Array(100).fill("large string content")
        };

        encoder.encode(largeObject);

        expect(outputSpy).toHaveBeenCalledTimes(1);
        const [chunk] = outputSpy.mock.calls[0];
        expect(chunk.byteLength).toBeGreaterThan(1024);
        expect(errorSpy).not.toHaveBeenCalled();
    });

    test("encoder releases buffer on close", () => {
        const outputSpy = vi.fn();
        const errorSpy = vi.fn();
        const encoder = new JsonEncoder({ output: outputSpy, error: errorSpy });

        encoder.encode({ test: "data" });
        expect(outputSpy).toHaveBeenCalledTimes(1);

        encoder.close();
        
        // Should still be able to encode after close (creates new buffer)
        encoder.encode({ test: "after close" });
        expect(outputSpy).toHaveBeenCalledTimes(2);
    });
});

describe("Edge Case Tests", () => {
    test("encoder handles empty object", () => {
        const outputSpy = vi.fn();
        const errorSpy = vi.fn();
        const encoder = new JsonEncoder({ output: outputSpy, error: errorSpy });

        encoder.encode({});

        expect(outputSpy).toHaveBeenCalledTimes(1);
        const [chunk] = outputSpy.mock.calls[0];
        expect(chunk.type).toBe("key");
        
        const decoded = new TextDecoder().decode(chunk.data);
        expect(JSON.parse(decoded)).toEqual({});
    });

    test("encoder handles null and primitive values", () => {
        const outputSpy = vi.fn();
        const errorSpy = vi.fn();
        const encoder = new JsonEncoder({ output: outputSpy, error: errorSpy });

        encoder.encode(null);
        encoder.encode("string");
        encoder.encode(42);
        encoder.encode(true);

        expect(outputSpy).toHaveBeenCalledTimes(4);
        const chunks = outputSpy.mock.calls.map(call => call[0]);
        
        chunks.forEach(chunk => {
            expect(chunk.type).toBe("key"); // Primitives are not arrays, so they're "key" type
        });
    });

    test("decoder handles empty JSON patch array", () => {
        const outputSpy = vi.fn();
        const errorSpy = vi.fn();
        const decoder = new JsonDecoder({ output: outputSpy, error: errorSpy });

        const emptyPatch: JsonPatch = [];
        const jsonString = JSON.stringify(emptyPatch);
        const data = new TextEncoder().encode(jsonString);
        const chunk = new EncodedJsonChunk({
            type: "delta",
            data,
            timestamp: Date.now()
        });

        decoder.decode(chunk);

        expect(outputSpy).toHaveBeenCalledTimes(1);
        expect(outputSpy).toHaveBeenCalledWith([]);
    });

    test("encoder handles nested objects with mixed types", () => {
        const outputSpy = vi.fn();
        const errorSpy = vi.fn();
        const encoder = new JsonEncoder({ output: outputSpy, error: errorSpy });
        encoder.configure({ replacer: ["bigint", "date"] });

        const complexObject = {
            id: BigInt(123),
            createdAt: new Date("2023-01-01T00:00:00.000Z"),
            nested: {
                array: [1, 2, { deep: BigInt(456) }],
                nullValue: null,
                boolValue: true
            }
        };

        encoder.encode(complexObject);

        expect(outputSpy).toHaveBeenCalledTimes(1);
        const [chunk] = outputSpy.mock.calls[0];
        
        const decoded = new TextDecoder().decode(chunk.data);
        const parsed = JSON.parse(decoded);
        expect(parsed.id).toBe("123");
        expect(parsed.createdAt).toBe("2023-01-01T00:00:00.000Z");
        expect(parsed.nested.array[2].deep).toBe("456");
    });
});

describe("Error Handling Edge Cases", () => {
    test("decoder handles corrupted UTF-8 data", () => {
        const outputSpy = vi.fn();
        const errorSpy = vi.fn();
        const decoder = new JsonDecoder({ output: outputSpy, error: errorSpy });

        // Create invalid UTF-8 sequence
        const invalidData = new Uint8Array([0xFF, 0xFE, 0xFD]);
        const chunk = new EncodedJsonChunk({
            type: "key",
            data: invalidData,
            timestamp: Date.now()
        });

        decoder.decode(chunk);

        expect(outputSpy).not.toHaveBeenCalled();
        expect(errorSpy).toHaveBeenCalledTimes(1);
    });

    test("encoder handles circular references gracefully", () => {
        const outputSpy = vi.fn();
        const errorSpy = vi.fn();
        const encoder = new JsonEncoder({ output: outputSpy, error: errorSpy });

        const circular: any = { name: "test" };
        circular.self = circular;

        // JSON.stringify should throw for circular references
        // The error will propagate up to the caller
        expect(() => encoder.encode(circular)).toThrow(/Converting circular structure to JSON/);
    });

    test("reviver functions handle edge cases", () => {
        // Test bigint reviver with edge cases
        expect(reviveBigInt("key", "999999999999999999999")).toBe(BigInt("999999999999999999999"));
        expect(reviveBigInt("key", "0")).toBe(BigInt(0));
        expect(reviveBigInt("key", "123.45")).toBe("123.45"); // Not pure digits
        
        // Test date reviver with edge cases  
        expect(reviveDate("key", "2023-12-31T23:59:59.999Z")).toBeInstanceOf(Date);
        expect(reviveDate("key", "2023-01-01T00:00:00.000")).toBe("2023-01-01T00:00:00.000"); // Missing Z
    });
});

describe("Performance and Reliability Tests", () => {
    test("encoder handles many rapid encodes", () => {
        const outputSpy = vi.fn();
        const errorSpy = vi.fn();
        const encoder = new JsonEncoder({ output: outputSpy, error: errorSpy });

        // Rapidly encode 1000 objects
        for (let i = 0; i < 1000; i++) {
            encoder.encode({ index: i, data: `test-${i}` });
        }

        expect(outputSpy).toHaveBeenCalledTimes(1000);
        expect(errorSpy).not.toHaveBeenCalled();
        
        // Verify timestamps are reasonable
        const timestamps = outputSpy.mock.calls.map(call => call[0].timestamp);
        const minTime = Math.min(...timestamps);
        const maxTime = Math.max(...timestamps);
        expect(maxTime - minTime).toBeLessThan(1000); // Should complete within 1 second
    });

    test("decoder handles many rapid decodes", () => {
        const outputSpy = vi.fn();
        const errorSpy = vi.fn();
        const decoder = new JsonDecoder({ output: outputSpy, error: errorSpy });

        // Create 100 chunks to decode rapidly
        for (let i = 0; i < 100; i++) {
            const jsonString = JSON.stringify({ index: i });
            const data = new TextEncoder().encode(jsonString);
            const chunk = new EncodedJsonChunk({
                type: "key",
                data,
                timestamp: Date.now()
            });
            decoder.decode(chunk);
        }

        expect(outputSpy).toHaveBeenCalledTimes(100);
        expect(errorSpy).not.toHaveBeenCalled();
    });

    test("encoder configuration is persistent", () => {
        const outputSpy = vi.fn();
        const errorSpy = vi.fn();
        const encoder = new JsonEncoder({ output: outputSpy, error: errorSpy });

        encoder.configure({ space: 2, replacer: ["bigint"] });

        // Single encode to test configuration persistence
        encoder.encode({ value1: BigInt(123), value2: "test" });

        expect(outputSpy).toHaveBeenCalledTimes(1);
        
        const [chunk] = outputSpy.mock.calls[0];
        const decoded = new TextDecoder().decode(chunk.data);
        
        // Should have indentation (space=2)
        expect(decoded).toContain('\n');
        expect(decoded).toContain('  '); // Should have 2-space indentation
        
        // Verify the actual structure - bigint should be converted to string
        const parsed = JSON.parse(decoded);
        expect(parsed.value1).toBe("123");
        expect(parsed.value2).toBe("test");
        expect(typeof parsed.value1).toBe("string");
        
        // Test that configuration persists for subsequent encodes
        outputSpy.mockClear();
        encoder.encode({ number: BigInt(789) });
        
        expect(outputSpy).toHaveBeenCalledTimes(1);
        const [chunk2] = outputSpy.mock.calls[0];
        const decoded2 = new TextDecoder().decode(chunk2.data);
        const parsed2 = JSON.parse(decoded2);
        
        expect(parsed2.number).toBe("789");
        expect(typeof parsed2.number).toBe("string");
        expect(decoded2).toContain('\n'); // Still formatted
    });
});

describe("Integration Tests", () => {
    test("encoder and decoder work together", () => {
        const results: any[] = [];
        const errors: Error[] = [];
        
        const encoder = new JsonEncoder({
            output: (chunk) => {
                decoder.decode(chunk);
            },
            error: (error) => errors.push(error)
        });
        
        const decoder = new JsonDecoder({
            output: (data) => results.push(data),
            error: (error) => errors.push(error)
        });

        const testData = { message: "hello", number: 42 };
        encoder.encode(testData);

        expect(results).toHaveLength(1);
        expect(results[0]).toEqual(testData);
        expect(errors).toHaveLength(0);
    });

    test("encoder and decoder work with bigint rules", () => {
        const results: any[] = [];
        const errors: Error[] = [];
        
        const encoder = new JsonEncoder({
            output: (chunk, metadata) => {
                if (metadata?.decoderConfig) {
                    decoder.configure(metadata.decoderConfig);
                }
                decoder.decode(chunk);
            },
            error: (error) => errors.push(error)
        });
        
        const decoder = new JsonDecoder({
            output: (data) => results.push(data),
            error: (error) => errors.push(error)
        });

        encoder.configure({ replacer: ["bigint"] });

        const testData = { id: BigInt(123456789), name: "test" };
        encoder.encode(testData);

        expect(results).toHaveLength(1);
        expect(results[0].id).toBe(BigInt(123456789));
        expect(results[0].name).toBe("test");
        expect(errors).toHaveLength(0);
    });

    test("encoder and decoder work with date rules", () => {
        const results: any[] = [];
        const errors: Error[] = [];
        
        const encoder = new JsonEncoder({
            output: (chunk, metadata) => {
                if (metadata?.decoderConfig) {
                    decoder.configure(metadata.decoderConfig);
                }
                decoder.decode(chunk);
            },
            error: (error) => errors.push(error)
        });
        
        const decoder = new JsonDecoder({
            output: (data) => results.push(data),
            error: (error) => errors.push(error)
        });

        encoder.configure({ replacer: ["date"] });

        const testDate = new Date("2023-01-01T00:00:00.000Z");
        const testData = { createdAt: testDate, title: "test" };
        encoder.encode(testData);

        expect(results).toHaveLength(1);
        expect(results[0].createdAt).toBeInstanceOf(Date);
        expect(results[0].createdAt.toISOString()).toBe("2023-01-01T00:00:00.000Z");
        expect(results[0].title).toBe("test");
        expect(errors).toHaveLength(0);
    });
});
