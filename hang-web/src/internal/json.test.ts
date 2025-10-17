import { describe, test, expect, beforeEach, afterEach, vi } from 'vitest';
import { z } from 'zod';
import {
    JsonEncoder,
    JsonDecoder,
    JsonLineEncoder,
    JsonLineDecoder,
    EncodedJsonChunk,
    replaceBigInt,
    reviveBigInt,
    replaceDate,
    reviveDate,
    JSON_RULES,
    JsonValueSchema
} from "./json";
import type {
    JsonEncoderConfig,
    JsonDecoderConfig,
    EncodedJsonChunkInit,
    JsonRuleName
} from "./json";
import type { JsonValue } from "./json";

describe("JsonValueSchema", () => {
    describe("valid values", () => {
        test("accepts string values", () => {
            const result = JsonValueSchema.safeParse("hello");
            expect(result.success).toBe(true);
        });

        test("accepts number values", () => {
            const result = JsonValueSchema.safeParse(42);
            expect(result.success).toBe(true);
        });

        test("accepts boolean values", () => {
            expect(JsonValueSchema.safeParse(true).success).toBe(true);
            expect(JsonValueSchema.safeParse(false).success).toBe(true);
        });

        test("accepts null values", () => {
            const result = JsonValueSchema.safeParse(null);
            expect(result.success).toBe(true);
        });

        test("accepts object values", () => {
            const obj = { name: "test", value: 123 };
            const result = JsonValueSchema.safeParse(obj);
            expect(result.success).toBe(true);
        });

        test("accepts array values", () => {
            const arr = [1, "hello", true, null];
            const result = JsonValueSchema.safeParse(arr);
            expect(result.success).toBe(true);
        });

        test("accepts nested objects", () => {
            const nested = {
                user: {
                    name: "John",
                    settings: {
                        theme: "dark",
                        notifications: true
                    }
                }
            };
            const result = JsonValueSchema.safeParse(nested);
            expect(result.success).toBe(true);
        });

        test("accepts nested arrays", () => {
            const nested = [[1, 2], ["a", "b"], [true, false]];
            const result = JsonValueSchema.safeParse(nested);
            expect(result.success).toBe(true);
        });
    });

    describe("invalid values", () => {
        test("rejects undefined", () => {
            const result = JsonValueSchema.safeParse(undefined);
            expect(result.success).toBe(false);
        });

        test("rejects functions", () => {
            const result = JsonValueSchema.safeParse(() => {});
            expect(result.success).toBe(false);
        });

        test("rejects symbols", () => {
            const result = JsonValueSchema.safeParse(Symbol("test"));
            expect(result.success).toBe(false);
        });
    });
});

describe("Type Definitions", () => {
    test("JsonValue type accepts all JSON-compatible values", () => {
        // These should compile without TypeScript errors
        const stringValue: JsonValue = "hello";
        const numberValue: JsonValue = 42;
        const booleanValue: JsonValue = true;
        const nullValue: JsonValue = null;
        const objectValue: JsonValue = { key: "value" };
        const arrayValue: JsonValue = [1, 2, 3];

        expect(typeof stringValue).toBe("string");
        expect(typeof numberValue).toBe("number");
        expect(typeof booleanValue).toBe("boolean");
        expect(nullValue).toBe(null);
        expect(typeof objectValue).toBe("object");
        expect(Array.isArray(arrayValue)).toBe(true);
    });
});

describe("Boundary Value Tests", () => {
    test("handles deeply nested objects", () => {
        const createNestedObject = (depth: number): any => {
            if (depth === 0) return "value";
            return { nested: createNestedObject(depth - 1) };
        };

        const deepObject = createNestedObject(100);
        const result = JsonValueSchema.safeParse(deepObject);
        expect(result.success).toBe(true);
    });

    test("handles large arrays", () => {
        const largeArray = new Array(1000).fill(0).map((_, i) => i);
        const result = JsonValueSchema.safeParse(largeArray);
        expect(result.success).toBe(true);
    });

    test("handles special characters in values", () => {
        const specialChars = {
            unicode: "🎉🚀⭐",
            newlines: "line1\nline2\r\nline3",
            tabs: "col1\tcol2\tcol3",
            quotes: 'He said "Hello" to me',
            backslashes: "C:\\Users\\test\\file.txt"
        };
        const result = JsonValueSchema.safeParse(specialChars);
        expect(result.success).toBe(true);
    });

    test("handles valid numeric edge cases", () => {
        const numbers = {
            zero: 0,
            negative: -123,
            decimal: 123.456,
            scientific: 1.23e-10,
            maxSafe: Number.MAX_SAFE_INTEGER,
            minSafe: Number.MIN_SAFE_INTEGER
        };
        const result = JsonValueSchema.safeParse(numbers);
        expect(result.success).toBe(true);
    });

    test("rejects invalid numeric values (infinity, NaN)", () => {
        // JSON doesn't support Infinity or NaN
        const invalidNumbers = [
            Number.POSITIVE_INFINITY,
            Number.NEGATIVE_INFINITY,
            NaN
        ];
        
        invalidNumbers.forEach(num => {
            const result = JsonValueSchema.safeParse(num);
            expect(result.success).toBe(false);
        });
    });
});

describe("JsonEncoder", () => {
    describe("constructor", () => {
        test("creates encoder", () => {
            const encoder = new JsonEncoder();
            expect(encoder).toBeDefined();
        });
    });

    describe("configure", () => {
        test("configures encoder with space setting", () => {
            const encoder = new JsonEncoder();
            const config: JsonEncoderConfig = {
                space: 2
            };

            expect(() => encoder.configure(config)).not.toThrow();
        });

        test("configures encoder with replacer rules", () => {
            const encoder = new JsonEncoder();
            const config: JsonEncoderConfig = {
                replacer: ["bigint", "date"]
            };

            expect(() => encoder.configure(config)).not.toThrow();
        });

        test("configures encoder with both space and replacer", () => {
            const encoder = new JsonEncoder();
            const config: JsonEncoderConfig = {
                space: 4,
                replacer: ["bigint"]
            };

            expect(() => encoder.configure(config)).not.toThrow();
        });
    });

    describe("encode", () => {
        test("encodes simple JSON value", () => {
            const encoder = new JsonEncoder();

            const value: JsonValue = { test: "value" };
            const chunk = encoder.encode([value]);

            expect(chunk).toBeInstanceOf(EncodedJsonChunk);
            expect(chunk.data.constructor.name).toBe('Uint8Array');
        });

        test("encodes with bigint replacer", () => {
            const encoder = new JsonEncoder();
            encoder.configure({ replacer: ["bigint"] });

            const value = { bigNumber: BigInt(123456789) };
            const chunk = encoder.encode([value]);

            expect(chunk).toBeInstanceOf(EncodedJsonChunk);
            
            // Decode the chunk to verify bigint was converted to string
            const decoded = new TextDecoder().decode(chunk.data);
            const parsed = JSON.parse(decoded);
            expect(parsed[0].bigNumber).toBe("123456789");
        });

        test("encodes with space formatting", () => {
            const encoder = new JsonEncoder();
            encoder.configure({ space: 2 });

            const values = [{ name: "test", nested: { value: 123 } }];
            const chunk = encoder.encode(values);

            const decoded = new TextDecoder().decode(chunk.data);
            const parsed = JSON.parse(decoded);

            // Verify the structure is correct
            expect(parsed[0].name).toBe("test");
            expect(parsed[0].nested.value).toBe(123);

            // Verify formatting was applied (contains newlines and spaces)
            expect(decoded).toContain('\n');
            expect(decoded).toContain('  '); // 2-space indentation
        });

        test("encodes with tab formatting", () => {
            const encoder = new JsonEncoder();
            encoder.configure({ space: '\t' });

            const values = [{ level1: { level2: "deep" } }];
            const chunk = encoder.encode(values);

            const decoded = new TextDecoder().decode(chunk.data);
            const parsed = JSON.parse(decoded);

            expect(parsed[0].level1.level2).toBe("deep");
            expect(decoded).toContain('\n\t'); // tab indentation
        });

describe("JsonDecoder", () => {
    describe("constructor", () => {
        test("creates decoder", () => {
            const decoder = new JsonDecoder();
            expect(decoder).toBeDefined();
        });
    });

    describe("configure", () => {
        test("configures decoder with reviver rules", () => {
            const decoder = new JsonDecoder();
            const config: JsonDecoderConfig = {
                reviverRules: ["bigint", "date"]
            };

            expect(() => decoder.configure(config)).not.toThrow();
        });

        test("configures decoder with empty reviver rules", () => {
            const decoder = new JsonDecoder();
            const config: JsonDecoderConfig = {
                reviverRules: []
            };

            expect(() => decoder.configure(config)).not.toThrow();
        });
    });

    describe("decode", () => {
        test("decodes simple JSON chunk", () => {
            const decoder = new JsonDecoder();

            const jsonString = JSON.stringify([{ test: "value" }]);
            const data = new TextEncoder().encode(jsonString);
            const chunk = new EncodedJsonChunk({ data, type: "json" });

            const result = decoder.decode(chunk);

            expect(result).toEqual([{ test: "value" }]);
        });

        test("decodes JSON patch array", () => {
            const decoder = new JsonDecoder();

            const patch = [[{ op: "add", path: "/test", value: "value" }]];
            const jsonString = JSON.stringify(patch);
            const data = new TextEncoder().encode(jsonString);
            const chunk = new EncodedJsonChunk({ data, type: "json" });

            const result = decoder.decode(chunk);

            expect(result).toEqual(patch);
        });

        test("decodes with bigint reviver", () => {
            const decoder = new JsonDecoder();
            decoder.configure({ reviverRules: ["bigint"] });

            const jsonString = JSON.stringify([{ bigNumber: "123456789" }]);
            const data = new TextEncoder().encode(jsonString);
            const chunk = new EncodedJsonChunk({ data, type: "json" });

            const result = decoder.decode(chunk);

            expect(result).toEqual([{ bigNumber: 123456789n }]);
        });
        //         timestamp: Date.now()
        //     });

        //     decoder.decode(chunk);

        //     expect(outputSpy).toHaveBeenCalledTimes(1);
        //     const [result] = outputSpy.mock.calls[0];
        //     expect(result.bigNumber).toBe(123456789n);
        // });

        test("decodes with date reviver", () => {
            const decoder = new JsonDecoder();
            decoder.configure({ reviverRules: ["date"] });

            const jsonString = JSON.stringify([{ createdAt: "2023-01-01T00:00:00.000Z" }]);
            const data = new TextEncoder().encode(jsonString);
            const chunk = new EncodedJsonChunk({ type: "json", data });

            const result = decoder.decode(chunk);

            expect(result).not.toBeNull();
            if (result && Array.isArray(result) && result[0] && typeof result[0] === "object" && "createdAt" in result[0]) {
                expect(result[0].createdAt).toBeInstanceOf(Date);
            } else {
                throw new Error("Decoded result does not have a createdAt property");
            }
            expect(result && result[0] && (result[0].createdAt as Date).toISOString()).toBe("2023-01-01T00:00:00.000Z");
        });

        test("handles invalid JSON gracefully", () => {
            const decoder = new JsonDecoder();

            const invalidJson = "{ invalid json";
            const data = new TextEncoder().encode(invalidJson);
            const chunk = new EncodedJsonChunk({ type: "json", data });

            expect(() => decoder.decode(chunk)).toThrow();
        });

        test("handles non-Error exceptions", () => {
            // Mock JSON.parse to throw a string instead of Error
            const originalParse = JSON.parse;
            JSON.parse = vi.fn().mockImplementation(() => {
                throw "String error";
            });

            const decoder = new JsonDecoder();

            const data = new TextEncoder().encode("{}");
            const chunk = new EncodedJsonChunk({ type: "json", data });

            expect(() => decoder.decode(chunk)).toThrow();

            // Restore original JSON.parse
            JSON.parse = originalParse;
        });

    });
});

describe("EncodedJsonChunk", () => {
    describe("constructor", () => {
        test("creates chunk with json type", () => {
            const data = new Uint8Array([1, 2, 3]);
            const chunk = new EncodedJsonChunk({ type: "json", data });

            expect(chunk.type).toBe("json");
            expect(chunk.data).toEqual(new Uint8Array([1, 2, 3]));
        });

        test("creates chunk with jsonl type", () => {
            const data = new Uint8Array([4, 5, 6]);
            const chunk = new EncodedJsonChunk({ type: "jsonl", data });

            expect(chunk.type).toBe("jsonl");
            expect(chunk.data).toEqual(new Uint8Array([4, 5, 6]));
        });

        test("creates chunk with empty data", () => {
            const data = new Uint8Array(0);
            const chunk = new EncodedJsonChunk({ type: "json", data });

            expect(chunk.type).toBe("json");
            expect(chunk.data).toEqual(new Uint8Array(0));
        });
    });

    describe("byteLength", () => {
        test("returns correct byte length", () => {
            const data = new Uint8Array([1, 2, 3, 4, 5]);
            const chunk = new EncodedJsonChunk({ type: "json", data });

            expect(chunk.byteLength).toBe(5);
        });

        test("returns zero for empty data", () => {
            const chunk = new EncodedJsonChunk({
                type: "json",
                data: new Uint8Array(0)
            });

            expect(chunk.byteLength).toBe(0);
        });
    });

    describe("copyTo", () => {
        test("copies data to target array", () => {
            const sourceData = new Uint8Array([1, 2, 3]);
            const chunk = new EncodedJsonChunk({ type: "json", data: sourceData });

            const target = new Uint8Array(5);
            chunk.copyTo(target);

            expect(target.subarray(0, 3)).toEqual(sourceData);
        });

        test("copies data to exact size target", () => {
            const sourceData = new Uint8Array([1, 2, 3]);
            const chunk = new EncodedJsonChunk({ type: "json", data: sourceData });

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
        const encoder = new JsonEncoder();

        // Encode multiple small objects
        const result1 = encoder.encode([{ test: 1 }]);
        const result2 = encoder.encode([{ test: 2 }]);
        const result3 = encoder.encode([{ test: 3 }]);

        expect(result1.data).toBeDefined();
        expect(result2.data).toBeDefined();
        expect(result3.data).toBeDefined();
    });

    test("encoder expands buffer for large data", () => {
        const encoder = new JsonEncoder();

        // Create a large object that would exceed initial buffer
        const largeObject = {
            data: "x".repeat(2000),
            array: Array(100).fill("large string content")
        };

        const result = encoder.encode([largeObject]);

        expect(result.data).toBeDefined();
        expect(result.data.byteLength).toBeGreaterThan(1024);
    });
});

describe("Edge Case Tests", () => {
    test("encoder handles empty object", () => {
        const encoder = new JsonEncoder();

        const result = encoder.encode([{}]);

        expect(result.data).toBeDefined();
        
        const decoded = new TextDecoder().decode(result.data);
        expect(JSON.parse(decoded)).toEqual([{}]);
    });

    test("encoder handles null and primitive values", () => {
        const encoder = new JsonEncoder();

        const result1 = encoder.encode([null]);
        const result2 = encoder.encode(["string"]);
        const result3 = encoder.encode([42]);
        const result4 = encoder.encode([true]);

        expect(result1.data).toBeDefined();
        expect(result2.data).toBeDefined();
        expect(result3.data).toBeDefined();
        expect(result4.data).toBeDefined();
    });
});

    test("encoder handles nested objects with mixed types", () => {
        const encoder = new JsonEncoder();
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

        const result = encoder.encode([complexObject]);
        
        const decoded = new TextDecoder().decode(result.data);
        const parsed = JSON.parse(decoded);
        expect(parsed[0].id).toBe("123");
        expect(parsed[0].createdAt).toBe("2023-01-01T00:00:00.000Z");
        expect(parsed[0].nested.array[2].deep).toBe("456");
    });

describe("Error Handling Edge Cases", () => {
    test("decoder handles corrupted UTF-8 data", () => {
        const decoder = new JsonDecoder();

        // Create invalid UTF-8 sequence
        const invalidData = new Uint8Array([0xFF, 0xFE, 0xFD]);
        const chunk = new EncodedJsonChunk({ type: "json", data: invalidData });

        expect(() => decoder.decode(chunk)).toThrow();
    });

    test("encoder handles circular references gracefully", () => {
        const encoder = new JsonEncoder();

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
        const encoder = new JsonEncoder();

        // Rapidly encode 1000 objects
        const results: any[] = [];
        for (let i = 0; i < 1000; i++) {
            const result = encoder.encode([{ index: i, data: `test-${i}` }]);
            results.push(result);
        }

        expect(results).toHaveLength(1000);
        results.forEach(result => {
            expect(result.data).toBeDefined();
        });
    });

    test("decoder handles many rapid decodes", () => {
        const decoder = new JsonDecoder();

        // Create 100 chunks to decode rapidly
        const results: any[] = [];
        for (let i = 0; i < 100; i++) {
            const jsonString = JSON.stringify([{ index: i }]);
            const data = new TextEncoder().encode(jsonString);
            const chunk = new EncodedJsonChunk({ type: "json", data });
            const result = decoder.decode(chunk);
            results.push(result);
        }

        expect(results).toHaveLength(100);
        results.forEach((result, i) => {
            expect(result[0].index).toBe(i);
        });
    });

    test("encoder configuration is persistent", () => {
        const encoder = new JsonEncoder();

        encoder.configure({ space: 2, replacer: ["bigint"] });

        // Single encode to test configuration persistence
        const result = encoder.encode([{ value1: BigInt(123), value2: "test" }]);

        const decoded = new TextDecoder().decode(result.data);
        
        // Should have indentation (space=2)
        expect(decoded).toContain('\n');
        expect(decoded).toContain('  '); // Should have 2-space indentation
        
        // Verify the actual structure - bigint should be converted to string
        const parsed = JSON.parse(decoded);
        expect(parsed[0].value1).toBe("123");
        expect(parsed[0].value2).toBe("test");
        expect(typeof parsed[0].value1).toBe("string");
        
        // Test that configuration persists for subsequent encodes
        const result2 = encoder.encode([{ number: BigInt(789) }]);
        const decoded2 = new TextDecoder().decode(result2.data);
        const parsed2 = JSON.parse(decoded2);
        
        expect(parsed2[0].number).toBe("789");
        expect(typeof parsed2[0].number).toBe("string");
        expect(decoded2).toContain('\n'); // Still formatted
    });
});

describe("Configuration Combination Tests", () => {
    test("encoder handles multiple replacer rules", () => {
        const encoder = new JsonEncoder();
        encoder.configure({ replacer: ["bigint", "date"] });

        const testData = [
            {
                id: BigInt(123456789),
                createdAt: new Date("2023-01-01T00:00:00.000Z"),
                name: "test",
                count: 42
            }
        ];

        const chunk = encoder.encode(testData);
        const decoded = new TextDecoder().decode(chunk.data);
        const parsed = JSON.parse(decoded);

        expect(parsed[0].id).toBe("123456789");
        expect(parsed[0].createdAt).toBe("2023-01-01T00:00:00.000Z");
        expect(parsed[0].name).toBe("test");
        expect(parsed[0].count).toBe(42);
    });

    test("decoder handles multiple reviver rules", () => {
        const encoder = new JsonEncoder();
        const decoder = new JsonDecoder();

        encoder.configure({ replacer: ["bigint", "date"] });
        decoder.configure({ reviverRules: ["bigint", "date"] });

        const originalData = [
            {
                id: BigInt(987654321),
                timestamp: new Date("2023-12-31T23:59:59.999Z"),
                active: true
            }
        ];

        const chunk = encoder.encode(originalData);
        const result = decoder.decode(chunk);

        expect(result[0].id).toBe(BigInt(987654321));
        expect(result[0].timestamp).toBeInstanceOf(Date);
        expect((result[0].timestamp as Date).toISOString()).toBe("2023-12-31T23:59:59.999Z");
        expect(result[0].active).toBe(true);
    });

    test("line encoder handles multiple replacer rules", () => {
        const encoder = new JsonLineEncoder();
        encoder.configure({ replacer: ["bigint", "date"], space: 2 });

        const testData = [
            { id: BigInt(111), time: new Date("2023-06-15T12:00:00.000Z") },
            { id: BigInt(222), time: new Date("2023-06-16T13:00:00.000Z") }
        ];

        const chunk = encoder.encode(testData);
        const decoded = new TextDecoder().decode(chunk.data);
        const lines = decoded.trim().split('\n');

        expect(lines).toHaveLength(2);
        const first = JSON.parse(lines[0]);
        const second = JSON.parse(lines[1]);

        expect(first.id).toBe("111");
        expect(first.time).toBe("2023-06-15T12:00:00.000Z");
        expect(second.id).toBe("222");
        expect(second.time).toBe("2023-06-16T13:00:00.000Z");
    });

    test("line decoder handles multiple reviver rules", () => {
        const encoder = new JsonLineEncoder();
        const decoder = new JsonLineDecoder();

        encoder.configure({ replacer: ["bigint", "date"] });
        decoder.configure({ reviverRules: ["bigint", "date"] });

        const originalData = [
            { big: BigInt(333), date: new Date("2023-07-01T00:00:00.000Z"), flag: false },
            { big: BigInt(444), date: new Date("2023-07-02T00:00:00.000Z"), flag: true }
        ];

        const chunk = encoder.encode(originalData);
        const result = decoder.decode(chunk);

        expect(result).toHaveLength(2);
        const firstResult = result[0] as any;
        const secondResult = result[1] as any;
        expect(firstResult.big).toBe(BigInt(333));
        expect(firstResult.date).toBeInstanceOf(Date);
        expect(firstResult.date.toISOString()).toBe("2023-07-01T00:00:00.000Z");
        expect(firstResult.flag).toBe(false);
        expect(secondResult.big).toBe(BigInt(444));
        expect(secondResult.date).toBeInstanceOf(Date);
        expect(secondResult.date.toISOString()).toBe("2023-07-02T00:00:00.000Z");
        expect(secondResult.flag).toBe(true);
    });
});

describe("Integration Tests", () => {
        const encoder = new JsonEncoder();
        const decoder = new JsonDecoder();

        const testData = { message: "hello", number: 42 };
        const chunk = encoder.encode([testData]);
        const result = decoder.decode(chunk);

        expect(result).toEqual([testData]);
    });

    test("encoder and decoder work with bigint rules", () => {
        const encoder = new JsonEncoder();
        const decoder = new JsonDecoder();

        encoder.configure({ replacer: ["bigint"] });
        decoder.configure({ reviverRules: ["bigint"] });

        const testData = { id: BigInt(123456789), name: "test" };
        const chunk = encoder.encode([testData]);
        const result = decoder.decode(chunk);

        expect(result[0].id).toBe(BigInt(123456789));
        expect(result[0].name).toBe("test");
    });

    test("encoder and decoder work with date rules", () => {
        const encoder = new JsonEncoder();
        const decoder = new JsonDecoder();

        encoder.configure({ replacer: ["date"] });
        decoder.configure({ reviverRules: ["date"] });

        const testDate = new Date("2023-01-01T00:00:00.000Z");
        const testData = { createdAt: testDate, title: "test" };
        const chunk = encoder.encode([testData]);
        const result = decoder.decode(chunk);

        expect(result[0].createdAt).toBeInstanceOf(Date);
        expect((result[0].createdAt as Date).toISOString()).toBe("2023-01-01T00:00:00.000Z");
        expect(result[0].title).toBe("test");
    });

    describe("Error handling", () => {
        test("throws error when encoding circular reference", () => {
            const encoder = new JsonEncoder();

            const circular: any = { self: null };
            circular.self = circular;

            expect(() => encoder.encode([circular])).toThrow(TypeError);
        });

        test("throws error when encoding BigInt without replacer", () => {
            const encoder = new JsonEncoder();

            const value = { big: BigInt(123) };

            expect(() => encoder.encode([value])).toThrow(TypeError);
        });
    });
});

describe("Schema Validation Error Tests", () => {
    test("decoder rejects invalid JSON structure", () => {
        const decoder = new JsonDecoder();

        // Create JSON that parses but doesn't match JsonArray schema
        const invalidJson = JSON.stringify("not an array");
        const data = new TextEncoder().encode(invalidJson);
        const chunk = new EncodedJsonChunk({ type: "json", data });

        expect(() => decoder.decode(chunk)).toThrow("Decoded JSON is not a valid JsonArray");
    });

    test("line decoder rejects invalid JSON line", () => {
        const decoder = new JsonLineDecoder();

        const jsonLines = [
            JSON.stringify({ valid: "object" }),
            '{"invalid": json}',  // Invalid JSON
            JSON.stringify({ another: "object" })
        ].join('\n');
        const data = new TextEncoder().encode(jsonLines);
        const chunk = new EncodedJsonChunk({ data, type: "jsonl" });

        expect(() => decoder.decode(chunk)).toThrow("Decoded JSON line is not a valid JsonValue");
    });

    test("line decoder handles mixed valid/invalid lines", () => {
        const decoder = new JsonLineDecoder();

        const jsonLines = [
            JSON.stringify({ first: "valid" }),
            '{"incomplete": ',  // Invalid JSON
            JSON.stringify({ third: "valid" })
        ].join('\n');
        const data = new TextEncoder().encode(jsonLines);
        const chunk = new EncodedJsonChunk({ data, type: "jsonl" });

        expect(() => decoder.decode(chunk)).toThrow("Decoded JSON line is not a valid JsonValue");
    });
});

describe("Dynamic Configuration Tests", () => {
    test("encoder configuration can be changed multiple times", () => {
        const encoder = new JsonEncoder();

        // First configuration
        encoder.configure({ replacer: ["bigint"] });
        const result1 = encoder.encode([{ value: BigInt(123) }]);
        const decoded1 = new TextDecoder().decode(result1.data);
        expect(JSON.parse(decoded1)[0].value).toBe("123");

        // Change configuration
        encoder.configure({ replacer: [] }); // No replacers
        const result2 = encoder.encode([{ value: BigInt(456) }]);
        const decoded2 = new TextDecoder().decode(result2.data);
        expect(() => JSON.parse(decoded2)).toThrow(); // BigInt should cause error

        // Change to different replacer
        encoder.configure({ replacer: ["date"] });
        const result3 = encoder.encode([{ time: new Date("2023-01-01T00:00:00.000Z") }]);
        const decoded3 = new TextDecoder().decode(result3.data);
        expect(JSON.parse(decoded3)[0].time).toBe("2023-01-01T00:00:00.000Z");
    });

    test("decoder configuration can be changed multiple times", () => {
        const encoder = new JsonEncoder();
        const decoder = new JsonDecoder();

        // Encode with both rules
        encoder.configure({ replacer: ["bigint", "date"] });
        const original = [{ id: BigInt(789), time: new Date("2023-01-01T00:00:00.000Z") }];
        const chunk = encoder.encode(original);

        // First decode with both revivers
        decoder.configure({ reviverRules: ["bigint", "date"] });
        const result1 = decoder.decode(chunk);
        expect(result1[0].id).toBe(BigInt(789));
        expect(result1[0].time).toBeInstanceOf(Date);

        // Change to only bigint reviver
        decoder.configure({ reviverRules: ["bigint"] });
        const result2 = decoder.decode(chunk);
        expect(result2[0].id).toBe(BigInt(789));
        expect(result2[0].time).toBe("2023-01-01T00:00:00.000Z"); // String, not Date

        // Change to only date reviver
        decoder.configure({ reviverRules: ["date"] });
        const result3 = decoder.decode(chunk);
        expect(typeof result3[0].id).toBe("string"); // "789", not BigInt
        expect(result3[0].time).toBeInstanceOf(Date);
    });
});

describe("Complex Data Structure Tests", () => {
    test("handles deeply nested structures", () => {
        const encoder = new JsonEncoder();
        const decoder = new JsonDecoder();

        const nestedData = [{
            level1: {
                level2: {
                    level3: {
                        value: "deep",
                        number: 42,
                        array: [1, 2, { nested: true }]
                    }
                },
                list: [
                    { id: 1, data: [1, 2, 3] },
                    { id: 2, data: [4, 5, 6] }
                ]
            }
        }];

        const chunk = encoder.encode(nestedData);
        const result = decoder.decode(chunk);

        expect(result[0].level1.level2.level3.value).toBe("deep");
        expect(result[0].level1.level2.level3.number).toBe(42);
        expect(result[0].level1.level2.level3.array).toEqual([1, 2, { nested: true }]);
        expect(result[0].level1.list).toHaveLength(2);
        expect(result[0].level1.list[0].data).toEqual([1, 2, 3]);
    });

    test("handles arrays with mixed types", () => {
        const encoder = new JsonEncoder();
        const decoder = new JsonDecoder();

        const mixedArray = [{
            mixed: [
                "string",
                42,
                true,
                null,
                { object: "in array" },
                [1, 2, 3]
            ]
        }];

        const chunk = encoder.encode(mixedArray);
        const result = decoder.decode(chunk);

        expect(result[0].mixed).toHaveLength(6);
        expect(result[0].mixed[0]).toBe("string");
        expect(result[0].mixed[1]).toBe(42);
        expect(result[0].mixed[2]).toBe(true);
        expect(result[0].mixed[3]).toBe(null);
        expect(result[0].mixed[4]).toEqual({ object: "in array" });
        expect(result[0].mixed[5]).toEqual([1, 2, 3]);
    });

    test("handles empty arrays and objects", () => {
        const encoder = new JsonEncoder();
        const decoder = new JsonDecoder();

        const emptyStructures = [{
            emptyArray: [],
            emptyObject: {},
            nestedEmpty: {
                arr: [],
                obj: {}
            }
        }];

        const chunk = encoder.encode(emptyStructures);
        const result = decoder.decode(chunk);

        expect(result[0].emptyArray).toEqual([]);
        expect(result[0].emptyObject).toEqual({});
        expect(result[0].nestedEmpty.arr).toEqual([]);
        expect(result[0].nestedEmpty.obj).toEqual({});
    });
});

describe("Unicode and Special Character Tests", () => {
    test("handles unicode characters", () => {
        const encoder = new JsonEncoder();
        const decoder = new JsonDecoder();

        const unicodeData = [{
            emoji: "🚀⭐🎉",
            japanese: "こんにちは世界",
            arabic: "مرحبا بالعالم",
            mixed: "Hello 世界 🌍"
        }];

        const chunk = encoder.encode(unicodeData);
        const result = decoder.decode(chunk);

        expect(result[0].emoji).toBe("🚀⭐🎉");
        expect(result[0].japanese).toBe("こんにちは世界");
        expect(result[0].arabic).toBe("مرحبا بالعالم");
        expect(result[0].mixed).toBe("Hello 世界 🌍");
    });

    test("handles special characters and escape sequences", () => {
        const encoder = new JsonEncoder();
        const decoder = new JsonDecoder();

        const specialData = [{
            quotes: 'She said "Hello" to me',
            backslash: "Path: C:\\Users\\file.txt",
            newline: "Line 1\nLine 2",
            tab: "Col1\tCol2\tCol3",
            unicodeEscape: "Unicode: \u0041\u0042\u0043"
        }];

        const chunk = encoder.encode(specialData);
        const result = decoder.decode(chunk);

        expect(result[0].quotes).toBe('She said "Hello" to me');
        expect(result[0].backslash).toBe("Path: C:\\Users\\file.txt");
        expect(result[0].newline).toBe("Line 1\nLine 2");
        expect(result[0].tab).toBe("Col1\tCol2\tCol3");
        expect(result[0].unicodeEscape).toBe("Unicode: ABC");
    });
});
  test('replaceBigInt converts bigint to string', () => {
    expect(replaceBigInt('k', 123n)).toBe('123');
    expect(replaceBigInt('k', 1)).toBe(1);
  });

  test('reviveBigInt converts numeric strings to BigInt', () => {
    expect(reviveBigInt('k', '456')).toBe(456n);
    // non-numeric strings remain as-is
    expect(reviveBigInt('k', '12a')).toBe('12a');
  });

  test('replaceDate converts Date to ISO string', () => {
    const d = new Date('2020-01-02T03:04:05.678Z');
    expect(replaceDate('k', d)).toBe('2020-01-02T03:04:05.678Z');
    expect(replaceDate('k', 'str')).toBe('str');
  });

  test('reviveDate converts ISO string to Date', () => {
    const iso = '2020-01-02T03:04:05.678Z';
    const res = reviveDate('k', iso);
    expect(res).toBeInstanceOf(Date);
    expect((res as Date).toISOString()).toBe(iso);

    expect(reviveDate('k', 'not-a-date')).toBe('not-a-date');
  });
});

describe("JsonLineEncoder", () => {
    describe("constructor", () => {
        test("creates encoder", () => {
            const encoder = new JsonLineEncoder();
            expect(encoder).toBeDefined();
        });
    });

    describe("configure", () => {
        test("configures encoder with space setting", () => {
            const encoder = new JsonLineEncoder();
            const config: JsonEncoderConfig = {
                space: 2
            };

            expect(() => encoder.configure(config)).not.toThrow();
        });

        test("configures encoder with replacer rules", () => {
            const encoder = new JsonLineEncoder();
            const config: JsonEncoderConfig = {
                replacer: ["bigint", "date"]
            };

            expect(() => encoder.configure(config)).not.toThrow();
        });

        test("configures encoder with both space and replacer", () => {
            const encoder = new JsonLineEncoder();
            const config: JsonEncoderConfig = {
                space: 4,
                replacer: ["bigint"]
            };

            expect(() => encoder.configure(config)).not.toThrow();
        });
    });

    describe("encode", () => {
        test("encodes multiple JSON values as lines", () => {
            const encoder = new JsonLineEncoder();

            const values: JsonValue[] = [
                { test: "value1" },
                { test: "value2" },
                { number: 42 }
            ];
            const chunk = encoder.encode(values);

            expect(chunk).toBeInstanceOf(EncodedJsonChunk);
            expect(chunk.data.constructor.name).toBe('Uint8Array');
            
            // Decode and verify it's line-separated JSON
            const decoded = new TextDecoder().decode(chunk.data);
            const lines = decoded.trim().split('\n');
            expect(lines).toHaveLength(3);
            expect(JSON.parse(lines[0])).toEqual({ test: "value1" });
            expect(JSON.parse(lines[1])).toEqual({ test: "value2" });
            expect(JSON.parse(lines[2])).toEqual({ number: 42 });
        });

        test("encodes with bigint replacer", () => {
            const encoder = new JsonLineEncoder();
            encoder.configure({ replacer: ["bigint"] });

            const values = [
                { bigNumber: BigInt(123456789) },
                { another: BigInt(987654321) }
            ];
            const chunk = encoder.encode(values);

            expect(chunk).toBeInstanceOf(EncodedJsonChunk);
            
            // Decode the chunk to verify bigint was converted to string
            const decoded = new TextDecoder().decode(chunk.data);
            const lines = decoded.trim().split('\n');
            expect(lines).toHaveLength(2);
            expect(JSON.parse(lines[0]).bigNumber).toBe("123456789");
            expect(JSON.parse(lines[1]).another).toBe("987654321");
        });

        test("encodes with date replacer", () => {
            const encoder = new JsonLineEncoder();
            encoder.configure({ replacer: ["date"] });

            const testDate = new Date("2023-01-01T00:00:00.000Z");
            const values = [
                { createdAt: testDate },
                { updatedAt: testDate }
            ];
            const chunk = encoder.encode(values);

            expect(chunk).toBeInstanceOf(EncodedJsonChunk);
            
            // Decode the chunk to verify date was converted to ISO string
            const decoded = new TextDecoder().decode(chunk.data);
            const lines = decoded.trim().split('\n');
            expect(lines).toHaveLength(2);
            expect(JSON.parse(lines[0]).createdAt).toBe("2023-01-01T00:00:00.000Z");
            expect(JSON.parse(lines[1]).updatedAt).toBe("2023-01-01T00:00:00.000Z");
        });

        test("encodes with space formatting", () => {
            const encoder = new JsonLineEncoder();
            encoder.configure({ space: 2 });

            const values = [
                { name: "test", nested: { value: 123 } }
            ];

            const chunk = encoder.encode(values);
            const decoded = new TextDecoder().decode(chunk.data);
            const lines = decoded.trim().split('\n');

            expect(lines).toHaveLength(1);
            // Should contain indentation
            expect(lines[0]).toContain('\n  ');
            expect(lines[0]).toContain('    '); // nested indentation
        });

        test("encodes with tab formatting", () => {
            const encoder = new JsonLineEncoder();
            encoder.configure({ space: '\t' });

            const values = [
                { level1: { level2: "deep" } }
            ];

            const chunk = encoder.encode(values);
            const decoded = new TextDecoder().decode(chunk.data);
            const lines = decoded.trim().split('\n');

            expect(lines).toHaveLength(1);
            expect(lines[0]).toContain('\n\t');
            expect(lines[0]).toContain('\t\t'); // nested tabs
        });
    });
});
describe("JsonLineDecoder", () => {
    describe("constructor", () => {
        test("creates decoder", () => {
            const decoder = new JsonLineDecoder();
            expect(decoder).toBeDefined();
        });
    });

    describe("configure", () => {
        test("configures decoder with reviver rules", () => {
            const decoder = new JsonLineDecoder();
            const config: JsonDecoderConfig = {
                reviverRules: ["bigint", "date"]
            };

            expect(() => decoder.configure(config)).not.toThrow();
        });

        test("configures decoder with empty reviver rules", () => {
            const decoder = new JsonLineDecoder();
            const config: JsonDecoderConfig = {
                reviverRules: []
            };

            expect(() => decoder.configure(config)).not.toThrow();
        });
    });

    describe("decode", () => {
        test("decodes line-separated JSON values", () => {
            const decoder = new JsonLineDecoder();

            const jsonLines = [
                JSON.stringify({ test: "value1" }),
                JSON.stringify({ test: "value2" }),
                JSON.stringify({ number: 42 })
            ].join('\n');
            const data = new TextEncoder().encode(jsonLines);
            const chunk = new EncodedJsonChunk({ type: "jsonl", data });

            const result = decoder.decode(chunk);

            expect(result).toHaveLength(3);
            expect(result[0]).toEqual({ test: "value1" });
            expect(result[1]).toEqual({ test: "value2" });
            expect(result[2]).toEqual({ number: 42 });
        });

        test("decodes with bigint reviver", () => {
            const decoder = new JsonLineDecoder();
            decoder.configure({ reviverRules: ["bigint"] });

            const jsonLines = [
                JSON.stringify({ bigNumber: "123456789" }),
                JSON.stringify({ another: "987654321" })
            ].join('\n');
            const data = new TextEncoder().encode(jsonLines);
            const chunk = new EncodedJsonChunk({ type: "jsonl", data });

            const result = decoder.decode(chunk);

            expect(result).toHaveLength(2);
            expect(result[0]).toEqual({ bigNumber: 123456789n });
            expect(result[1]).toEqual({ another: 987654321n });
        });

        test("decodes with date reviver", () => {
            const decoder = new JsonLineDecoder();
            decoder.configure({ reviverRules: ["date"] });

            const jsonLines = [
                JSON.stringify({ createdAt: "2023-01-01T00:00:00.000Z" }),
                JSON.stringify({ updatedAt: "2023-12-31T23:59:59.999Z" })
            ].join('\n');
            const data = new TextEncoder().encode(jsonLines);
            const chunk = new EncodedJsonChunk({ type: "jsonl", data });

            const result = decoder.decode(chunk);

            expect(result).toHaveLength(2);
            const firstResult = result[0] as any;
            const secondResult = result[1] as any;
            expect(firstResult.createdAt).toBeInstanceOf(Date);
            expect(firstResult.createdAt.toISOString()).toBe("2023-01-01T00:00:00.000Z");
            expect(secondResult.updatedAt).toBeInstanceOf(Date);
            expect(secondResult.updatedAt.toISOString()).toBe("2023-12-31T23:59:59.999Z");
        });

        test("handles empty input", () => {
            const decoder = new JsonLineDecoder();

            const data = new TextEncoder().encode("");
            const chunk = new EncodedJsonChunk({ type: "jsonl", data });

            expect(() => decoder.decode(chunk)).toThrow("No JSON lines found");
        });

        test("handles whitespace-only lines", () => {
            const decoder = new JsonLineDecoder();

            const jsonLines = [
                JSON.stringify({ test: "value1" }),
                "   \t   ",
                JSON.stringify({ test: "value2" })
            ].join('\n');
            const data = new TextEncoder().encode(jsonLines);
            const chunk = new EncodedJsonChunk({ type: "jsonl", data });

            const result = decoder.decode(chunk);

            expect(result).toHaveLength(2);
            expect(result[0]).toEqual({ test: "value1" });
            expect(result[1]).toEqual({ test: "value2" });
        });

        test("throws error for invalid JSON line", () => {
            const decoder = new JsonLineDecoder();

            const jsonLines = [
                JSON.stringify({ test: "value1" }),
                "{ invalid json",
                JSON.stringify({ test: "value2" })
            ].join('\n');
            const data = new TextEncoder().encode(jsonLines);
            const chunk = new EncodedJsonChunk({ type: "jsonl", data });

            expect(() => decoder.decode(chunk)).toThrow("Decoded JSON line is not a valid JsonValue");
        });

        test("handles formatted JSON lines", () => {
            const decoder = new JsonLineDecoder();

            const formattedJson = `{
  "name": "formatted",
  "nested": {
    "value": 42,
    "array": [1, 2, 3]
  }
}`;
            const data = new TextEncoder().encode(formattedJson);
            const chunk = new EncodedJsonChunk({ type: "jsonl", data });

            const result = decoder.decode(chunk);

            expect(result).toHaveLength(1);
            expect((result[0] as any).name).toBe("formatted");
            expect((result[0] as any).nested.value).toBe(42);
            expect((result[0] as any).nested.array).toEqual([1, 2, 3]);
        });

        test("handles multiple formatted lines", () => {
            const decoder = new JsonLineDecoder();

            const formattedLines = `{
                "first": "object",
                "count": 1
                }
                {
                "second": "object",
                "count": 2
                }`;
            const data = new TextEncoder().encode(formattedLines);
            const chunk = new EncodedJsonChunk({ type: "jsonl", data });

            const result = decoder.decode(chunk);

            expect(result).toHaveLength(2);
            expect((result[0] as any).first).toBe("object");
            expect((result[0] as any).count).toBe(1);
            expect((result[1] as any).second).toBe("object");
            expect((result[1] as any).count).toBe(2);
        });
    });
});
