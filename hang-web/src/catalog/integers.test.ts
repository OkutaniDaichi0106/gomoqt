import {
    uint8Schema,
    uint8,
    uint53Schema,
    uint53,
    uint62Schema,
    uint62
} from './integers';
import { z } from 'zod';
import { describe, test, expect, beforeEach, afterEach, it, vi } from 'vitest';

describe('catalog integers', () => {
  it('uint8 accepts 0..255', () => {
    expect(uint8(0)).toBe(0);
    expect(uint8(255)).toBe(255);
    expect(() => uint8(256)).toThrow();
  });

  it('uint53 accepts safe integers', () => {
    expect(uint53(0)).toBe(0);
    expect(uint53(Number.MAX_SAFE_INTEGER)).toBe(Number.MAX_SAFE_INTEGER);
    expect(() => uint53(Number.MAX_SAFE_INTEGER + 1)).toThrow();
  });

  it('uint62 accepts bigint up to 62 bits', () => {
    expect(uint62(0)).toBe(0);
    expect(uint62(123n)).toBe(123n);
    expect(() => uint62(-1 as any)).toThrow();
  });
});

describe("Integer Types", () => {
    describe("uint8", () => {
        describe("uint8Schema", () => {
            test("accepts valid uint8 values", () => {
                expect(uint8Schema.parse(0)).toBe(0);
                expect(uint8Schema.parse(1)).toBe(1);
                expect(uint8Schema.parse(127)).toBe(127);
                expect(uint8Schema.parse(255)).toBe(255);
            });

            test("rejects values below 0", () => {
                expect(() => uint8Schema.parse(-1)).toThrow();
                expect(() => uint8Schema.parse(-100)).toThrow();
            });

            test("rejects values above 255", () => {
                expect(() => uint8Schema.parse(256)).toThrow();
                expect(() => uint8Schema.parse(1000)).toThrow();
            });

            test("rejects non-integer values", () => {
                expect(() => uint8Schema.parse(1.5)).toThrow();
                expect(() => uint8Schema.parse(255.1)).toThrow();
                expect(() => uint8Schema.parse(Math.PI)).toThrow();
            });

            test("rejects non-numeric types", () => {
                expect(() => uint8Schema.parse("255")).toThrow();
                expect(() => uint8Schema.parse(null)).toThrow();
                expect(() => uint8Schema.parse(undefined)).toThrow();
                expect(() => uint8Schema.parse({})).toThrow();
                expect(() => uint8Schema.parse([])).toThrow();
                expect(() => uint8Schema.parse(true)).toThrow();
            });

            test("handles edge cases", () => {
                expect(uint8Schema.parse(0)).toBe(0);
                expect(uint8Schema.parse(255)).toBe(255);
                expect(() => uint8Schema.parse(NaN)).toThrow();
                expect(() => uint8Schema.parse(Infinity)).toThrow();
                expect(() => uint8Schema.parse(-Infinity)).toThrow();
            });
        });

        describe("uint8 function", () => {
            test("validates and returns uint8 values", () => {
                expect(uint8(0)).toBe(0);
                expect(uint8(128)).toBe(128);
                expect(uint8(255)).toBe(255);
            });

            test("throws on invalid values", () => {
                expect(() => uint8(-1)).toThrow();
                expect(() => uint8(256)).toThrow();
                expect(() => uint8(1.5)).toThrow();
            });

            test("type safety at runtime", () => {
                const validValue = uint8(42);
                expect(typeof validValue).toBe("number");
                expect(Number.isInteger(validValue)).toBe(true);
                expect(validValue >= 0 && validValue <= 255).toBe(true);
            });
        });
    });

    describe("uint53", () => {
        describe("uint53Schema", () => {
            test("accepts valid uint53 values", () => {
                expect(uint53Schema.parse(0)).toBe(0);
                expect(uint53Schema.parse(1)).toBe(1);
                expect(uint53Schema.parse(1000000)).toBe(1000000);
                expect(uint53Schema.parse(Number.MAX_SAFE_INTEGER)).toBe(Number.MAX_SAFE_INTEGER);
            });

            test("rejects values below 0", () => {
                expect(() => uint53Schema.parse(-1)).toThrow();
                expect(() => uint53Schema.parse(-Number.MAX_SAFE_INTEGER)).toThrow();
            });

            test("rejects values above MAX_SAFE_INTEGER", () => {
                expect(() => uint53Schema.parse(Number.MAX_SAFE_INTEGER + 1)).toThrow();
                expect(() => uint53Schema.parse(Number.MAX_VALUE)).toThrow();
            });

            test("rejects non-integer values", () => {
                expect(() => uint53Schema.parse(1.5)).toThrow();
                expect(() => uint53Schema.parse(0.1)).toThrow();
                expect(() => uint53Schema.parse(Math.E)).toThrow();
            });

            test("handles edge cases", () => {
                expect(uint53Schema.parse(0)).toBe(0);
                expect(uint53Schema.parse(Number.MAX_SAFE_INTEGER)).toBe(Number.MAX_SAFE_INTEGER);
                expect(() => uint53Schema.parse(NaN)).toThrow();
                expect(() => uint53Schema.parse(Infinity)).toThrow();
                expect(() => uint53Schema.parse(-Infinity)).toThrow();
            });

            test("aligns with Number.isSafeInteger", () => {
                const testValues = [0, 1, 1000, Number.MAX_SAFE_INTEGER];
                
                testValues.forEach(value => {
                    expect(Number.isSafeInteger(value)).toBe(true);
                    expect(() => uint53Schema.parse(value)).not.toThrow();
                });

                const unsafeValues = [Number.MAX_SAFE_INTEGER + 1, Number.MAX_SAFE_INTEGER + 2];
                unsafeValues.forEach(value => {
                    expect(Number.isSafeInteger(value)).toBe(false);
                    expect(() => uint53Schema.parse(value)).toThrow();
                });
            });
        });

        describe("uint53 function", () => {
            test("validates and returns uint53 values", () => {
                expect(uint53(0)).toBe(0);
                expect(uint53(42)).toBe(42);
                expect(uint53(Number.MAX_SAFE_INTEGER)).toBe(Number.MAX_SAFE_INTEGER);
            });

            test("throws on invalid values", () => {
                expect(() => uint53(-1)).toThrow();
                expect(() => uint53(Number.MAX_SAFE_INTEGER + 1)).toThrow();
                expect(() => uint53(1.5)).toThrow();
            });

            test("maintains type safety", () => {
                const value = uint53(12345);
                expect(typeof value).toBe("number");
                expect(Number.isInteger(value)).toBe(true);
                expect(Number.isSafeInteger(value)).toBe(true);
            });
        });
    });

    describe("uint62", () => {
        describe("uint62Schema", () => {
            test("accepts valid number values (uint53 range)", () => {
                expect(uint62Schema.parse(0)).toBe(0);
                expect(uint62Schema.parse(1)).toBe(1);
                expect(uint62Schema.parse(Number.MAX_SAFE_INTEGER)).toBe(Number.MAX_SAFE_INTEGER);
            });

            test("accepts valid bigint values", () => {
                expect(uint62Schema.parse(0n)).toBe(0n);
                expect(uint62Schema.parse(1n)).toBe(1n);
                expect(uint62Schema.parse(2n ** 53n)).toBe(2n ** 53n);
                expect(uint62Schema.parse(2n ** 62n - 1n)).toBe(2n ** 62n - 1n);
            });

            test("rejects negative numbers", () => {
                expect(() => uint62Schema.parse(-1)).toThrow();
                expect(() => uint62Schema.parse(-100)).toThrow();
            });

            test("rejects negative bigints", () => {
                expect(() => uint62Schema.parse(-1n)).toThrow();
                expect(() => uint62Schema.parse(-100n)).toThrow();
            });

            test("rejects numbers above MAX_SAFE_INTEGER", () => {
                expect(() => uint62Schema.parse(Number.MAX_SAFE_INTEGER + 1)).toThrow();
                expect(() => uint62Schema.parse(Number.MAX_VALUE)).toThrow();
            });

            test("rejects bigints above 2^62-1", () => {
                expect(() => uint62Schema.parse(2n ** 62n)).toThrow();
                expect(() => uint62Schema.parse(2n ** 63n)).toThrow();
                expect(() => uint62Schema.parse(2n ** 100n)).toThrow();
            });

            test("rejects non-integer numbers", () => {
                expect(() => uint62Schema.parse(1.5)).toThrow();
                expect(() => uint62Schema.parse(0.1)).toThrow();
            });

            test("handles boundary values correctly", () => {
                // Maximum safe integer as number
                expect(uint62Schema.parse(Number.MAX_SAFE_INTEGER)).toBe(Number.MAX_SAFE_INTEGER);
                
                // Maximum uint62 as bigint
                const maxUint62 = 2n ** 62n - 1n;
                expect(uint62Schema.parse(maxUint62)).toBe(maxUint62);

                // Just over the limit should fail
                expect(() => uint62Schema.parse(2n ** 62n)).toThrow();
            });

            test("preserves type distinction between number and bigint", () => {
                const numberResult = uint62Schema.parse(42);
                const bigintResult = uint62Schema.parse(42n);
                
                expect(typeof numberResult).toBe("number");
                expect(typeof bigintResult).toBe("bigint");
                expect(numberResult).toBe(42);
                expect(bigintResult).toBe(42n);
            });
        });

        describe("uint62 function", () => {
            test("validates and returns number values", () => {
                expect(uint62(0)).toBe(0);
                expect(uint62(42)).toBe(42);
                expect(uint62(Number.MAX_SAFE_INTEGER)).toBe(Number.MAX_SAFE_INTEGER);
            });

            test("validates and returns bigint values", () => {
                expect(uint62(0n)).toBe(0n);
                expect(uint62(42n)).toBe(42n);
                expect(uint62(2n ** 53n)).toBe(2n ** 53n);
                expect(uint62(2n ** 62n - 1n)).toBe(2n ** 62n - 1n);
            });

            test("throws on invalid number values", () => {
                expect(() => uint62(-1)).toThrow();
                expect(() => uint62(Number.MAX_SAFE_INTEGER + 1)).toThrow();
                expect(() => uint62(1.5)).toThrow();
            });

            test("throws on invalid bigint values", () => {
                expect(() => uint62(-1n)).toThrow();
                expect(() => uint62(2n ** 62n)).toThrow();
            });

            test("maintains type identity", () => {
                const numberInput = 1000;
                const bigintInput = 1000n;
                
                const numberResult = uint62(numberInput);
                const bigintResult = uint62(bigintInput);
                
                expect(typeof numberResult).toBe("number");
                expect(typeof bigintResult).toBe("bigint");
            });
        });
    });

    describe("Integration and Cross-Type Tests", () => {
        test("uint8 is subset of uint53", () => {
            const uint8Values = [0, 1, 127, 255];
            
            uint8Values.forEach(value => {
                expect(() => uint8(value)).not.toThrow();
                expect(() => uint53(value)).not.toThrow();
                expect(() => uint62(value)).not.toThrow();
            });
        });

        test("uint53 is subset of uint62 (number range)", () => {
            const uint53Values = [0, 1, 1000000, Number.MAX_SAFE_INTEGER];
            
            uint53Values.forEach(value => {
                expect(() => uint53(value)).not.toThrow();
                expect(() => uint62(value)).not.toThrow();
            });
        });

        test("type hierarchy consistency", () => {
            // Test that smaller types fit into larger ones
            const testValue = 100;
            
            const u8 = uint8(testValue);
            const u53 = uint53(u8);
            const u62 = uint62(u53);
            
            expect(u8).toBe(testValue);
            expect(u53).toBe(testValue);
            expect(u62).toBe(testValue);
        });

        test("error messages are descriptive", () => {
            try {
                uint8(256);
                expect.fail("Expected error to be thrown");
            } catch (error) {
                expect(error).toBeInstanceOf(z.ZodError);
                expect((error as z.ZodError).issues).toHaveLength(1);
            }

            try {
                uint53(-1);
                expect.fail("Expected error to be thrown");
            } catch (error) {
                expect(error).toBeInstanceOf(z.ZodError);
                expect((error as z.ZodError).issues).toHaveLength(1);
            }

            try {
                uint62(2n ** 62n);
                expect.fail("Expected error to be thrown");
            } catch (error) {
                expect(error).toBeInstanceOf(z.ZodError);
                expect((error as z.ZodError).issues).toHaveLength(1);
            }
        });
    });

    describe("Performance and Edge Cases", () => {
        test("handles rapid successive validations", () => {
            const iterations = 1000;
            
            for (let i = 0; i < iterations; i++) {
                expect(uint8(Math.floor(Math.random() * 256))).toBeGreaterThanOrEqual(0);
                expect(uint53(Math.floor(Math.random() * 1000000))).toBeGreaterThanOrEqual(0);
            }
        });

        test("memory usage is predictable", () => {
            const values = Array.from({ length: 100 }, (_, i) => i);
            
            values.forEach(value => {
                const u8Result = uint8(value % 256);
                const u53Result = uint53(value);
                const u62Result = uint62(value);
                
                expect(typeof u8Result).toBe("number");
                expect(typeof u53Result).toBe("number");
                expect(typeof u62Result).toBe("number");
            });
        });

        test("bigint conversion edge cases", () => {
            // Test conversion between number and bigint representations
            const numberValue = 42;
            const bigintValue = 42n;
            
            expect(uint62(numberValue)).toBe(numberValue);
            expect(uint62(bigintValue)).toBe(bigintValue);
            expect(uint62(numberValue)).not.toBe(bigintValue); // Different types
        });

        test("handles special numeric values", () => {
            const specialValues = [NaN, Infinity, -Infinity];
            
            specialValues.forEach(value => {
                expect(() => uint8(value)).toThrow();
                expect(() => uint53(value)).toThrow();
                expect(() => uint62(value)).toThrow();
            });
        });
    });

    describe("SafeParse API Tests", () => {
        describe("uint8Schema safeParse", () => {
            test("should return success result for valid values", () => {
                const result = uint8Schema.safeParse(0);
                expect(result.success).toBe(true);
                if (result.success) {
                    expect(result.data).toBe(0);
                }

                const result2 = uint8Schema.safeParse(255);
                expect(result2.success).toBe(true);
                if (result2.success) {
                    expect(result2.data).toBe(255);
                }
            });

            test("should return failure result for invalid values", () => {
                const result = uint8Schema.safeParse(-1);
                expect(result.success).toBe(false);
                if (!result.success) {
                    expect(result.error).toBeInstanceOf(z.ZodError);
                }

                const result2 = uint8Schema.safeParse(256);
                expect(result2.success).toBe(false);
                if (!result2.success) {
                    expect(result2.error).toBeInstanceOf(z.ZodError);
                }
            });

            test("should provide error details", () => {
                const result = uint8Schema.safeParse('invalid');
                expect(result.success).toBe(false);
                if (!result.success) {
                    expect(result.error.issues.length).toBeGreaterThan(0);
                }
            });
        });

        describe("uint53Schema safeParse", () => {
            test("should return success result for valid values", () => {
                const result = uint53Schema.safeParse(0);
                expect(result.success).toBe(true);
                if (result.success) {
                    expect(result.data).toBe(0);
                }

                const result2 = uint53Schema.safeParse(Number.MAX_SAFE_INTEGER);
                expect(result2.success).toBe(true);
                if (result2.success) {
                    expect(result2.data).toBe(Number.MAX_SAFE_INTEGER);
                }
            });

            test("should return failure result for invalid values", () => {
                const result = uint53Schema.safeParse(Number.MAX_SAFE_INTEGER + 1);
                expect(result.success).toBe(false);

                const result2 = uint53Schema.safeParse(-1);
                expect(result2.success).toBe(false);
            });
        });

        describe("uint62Schema safeParse", () => {
            test("should return success result for valid number values", () => {
                const result = uint62Schema.safeParse(0);
                expect(result.success).toBe(true);
                if (result.success) {
                    expect(result.data).toBe(0);
                }

                const result2 = uint62Schema.safeParse(Number.MAX_SAFE_INTEGER);
                expect(result2.success).toBe(true);
                if (result2.success) {
                    expect(result2.data).toBe(Number.MAX_SAFE_INTEGER);
                }
            });

            test("should return success result for valid bigint values", () => {
                const result = uint62Schema.safeParse(0n);
                expect(result.success).toBe(true);
                if (result.success) {
                    expect(result.data).toBe(0n);
                }

                const result2 = uint62Schema.safeParse(2n ** 62n - 1n);
                expect(result2.success).toBe(true);
                if (result2.success) {
                    expect(result2.data).toBe(2n ** 62n - 1n);
                }
            });

            test("should return failure result for values exceeding 2^62-1", () => {
                const result = uint62Schema.safeParse(2n ** 62n);
                expect(result.success).toBe(false);

                const result2 = uint62Schema.safeParse(2n ** 63n);
                expect(result2.success).toBe(false);
            });
        });
    });

    describe("Boundary and Corner Cases", () => {
        describe("uint8 boundaries", () => {
            test("should handle all valid boundaries", () => {
                // Test all critical boundaries
                expect(uint8(0)).toBe(0);           // Minimum
                expect(uint8(1)).toBe(1);           // Just above minimum
                expect(uint8(127)).toBe(127);       // Middle value
                expect(uint8(254)).toBe(254);       // Just below maximum
                expect(uint8(255)).toBe(255);       // Maximum

                // Test boundary violations
                expect(() => uint8(-1)).toThrow();
                expect(() => uint8(0.5)).toThrow();
                expect(() => uint8(255.5)).toThrow();
                expect(() => uint8(256)).toThrow();
            });

            test("should maintain consistency with schema", () => {
                for (let i = 0; i <= 255; i++) {
                    expect(uint8(i)).toBe(uint8Schema.parse(i));
                }
            });
        });

        describe("uint53 boundaries", () => {
            test("should handle critical boundary values", () => {
                // Test around Number.MAX_SAFE_INTEGER
                const maxSafe = Number.MAX_SAFE_INTEGER;
                expect(uint53(maxSafe - 1)).toBe(maxSafe - 1);
                expect(uint53(maxSafe)).toBe(maxSafe);
                expect(() => uint53(maxSafe + 1)).toThrow();
            });

            test("should handle powers of 2 correctly", () => {
                // Powers of 2 are important boundary values
                for (let i = 0; i <= 52; i++) {
                    const value = Math.pow(2, i);
                    if (value <= Number.MAX_SAFE_INTEGER) {
                        expect(uint53(value)).toBe(value);
                    }
                }
            });

            test("should recognize safe vs unsafe integers", () => {
                // Test that all passed values are recognized as safe
                const testValues = [0, 1, 100, 1000000, Number.MAX_SAFE_INTEGER];
                testValues.forEach(value => {
                    expect(Number.isSafeInteger(value)).toBe(true);
                    expect(uint53(value)).toBe(value);
                });

                // Test unsafe values
                const unsafeValues = [Number.MAX_SAFE_INTEGER + 1, Number.MAX_VALUE];
                unsafeValues.forEach(value => {
                    expect(Number.isSafeInteger(value)).toBe(false);
                    expect(() => uint53(value)).toThrow();
                });
            });
        });

        describe("uint62 boundaries with bigint", () => {
            test("should handle uint62 boundary values precisely", () => {
                const max62 = 2n ** 62n - 1n;
                const over62 = 2n ** 62n;

                // Maximum valid value
                expect(uint62(max62)).toBe(max62);

                // Just over the limit
                expect(() => uint62(over62)).toThrow();
            });

            test("should handle 2^53 transition boundary", () => {
                // 2^53 is the boundary where JavaScript numbers lose precision
                const value = 2n ** 53n;
                expect(uint62(value)).toBe(value);

                // Just below and above
                expect(uint62(2n ** 53n - 1n)).toBe(2n ** 53n - 1n);
                expect(uint62(2n ** 53n + 1n)).toBe(2n ** 53n + 1n);
            });

            test("should preserve exact values at boundaries", () => {
                const testValues = [
                    0n,
                    1n,
                    2n ** 8n - 1n,      // 255
                    2n ** 16n - 1n,     // 65535
                    2n ** 32n - 1n,     // 4294967295
                    2n ** 53n - 1n,     // Number.MAX_SAFE_INTEGER as bigint
                    2n ** 53n,
                    2n ** 53n + 1n,
                    2n ** 60n,
                    2n ** 62n - 1n,     // Maximum uint62
                ];

                testValues.forEach(value => {
                    expect(uint62(value)).toBe(value);
                });
            });
        });
    });

    describe("Type System and Runtime Validation", () => {
        describe("Type preservation", () => {
            test("uint8 always returns number", () => {
                const values = [0, 1, 127, 255];
                values.forEach(value => {
                    const result = uint8(value);
                    expect(typeof result).toBe('number');
                    expect(Number.isInteger(result)).toBe(true);
                });
            });

            test("uint53 always returns number", () => {
                const values = [0, 1, 1000000, Number.MAX_SAFE_INTEGER];
                values.forEach(value => {
                    const result = uint53(value);
                    expect(typeof result).toBe('number');
                    expect(Number.isInteger(result)).toBe(true);
                    expect(Number.isSafeInteger(result)).toBe(true);
                });
            });

            test("uint62 preserves input type", () => {
                // Numbers remain numbers
                const numResult = uint62(42);
                expect(typeof numResult).toBe('number');
                expect(numResult).toBe(42);

                // Bigints remain bigints
                const bigintResult = uint62(42n);
                expect(typeof bigintResult).toBe('bigint');
                expect(bigintResult).toBe(42n);

                // They are not equal due to type difference
                expect(numResult).not.toBe(bigintResult);
                expect(Number(bigintResult)).toBe(numResult);
            });
        });

        describe("Union type handling for uint62", () => {
            test("should accept both number and bigint in union", () => {
                // Number path
                const numValue = uint62Schema.parse(100);
                expect(numValue).toBe(100);
                expect(typeof numValue).toBe('number');

                // Bigint path
                const bigintValue = uint62Schema.parse(100n);
                expect(bigintValue).toBe(100n);
                expect(typeof bigintValue).toBe('bigint');
            });

            test("should correctly validate both branches independently", () => {
                // Valid as number
                expect(() => uint62(Number.MAX_SAFE_INTEGER)).not.toThrow();

                // Valid as bigint beyond Number.MAX_SAFE_INTEGER
                expect(() => uint62(BigInt(Number.MAX_SAFE_INTEGER) + 1n)).not.toThrow();

                // Invalid in both branches
                expect(() => uint62(-1)).toThrow();
                expect(() => uint62(-1n)).toThrow();
            });
        });
    });

    describe("Schema Composition and Reusability", () => {
        describe("Schema independence", () => {
            test("schemas should not affect each other", () => {
                // Modifying one schema validation should not affect others
                expect(uint8Schema.parse(255)).toBe(255);
                expect(uint53Schema.parse(255)).toBe(255);
                expect(uint62Schema.parse(255)).toBe(255);

                expect(() => uint8Schema.parse(256)).toThrow();
                expect(uint53Schema.parse(256)).toBe(256);
                expect(uint62Schema.parse(256)).toBe(256);
            });

            test("should parse consistently across multiple calls", () => {
                const value = 42;
                for (let i = 0; i < 10; i++) {
                    expect(uint8Schema.parse(value)).toBe(value);
                    expect(uint53Schema.parse(value)).toBe(value);
                    expect(uint62Schema.parse(value)).toBe(value);
                }
            });
        });

        describe("Error consistency", () => {
            test("all schemas throw ZodError on invalid input", () => {
                expect(() => uint8Schema.parse(-1)).toThrow(z.ZodError);
                expect(() => uint53Schema.parse(-1)).toThrow(z.ZodError);
                expect(() => uint62Schema.parse(-1n)).toThrow(z.ZodError);
            });

            test("error structure is consistent", () => {
                try {
                    uint8Schema.parse('invalid');
                } catch (error) {
                    expect(error).toBeInstanceOf(z.ZodError);
                    if (error instanceof z.ZodError) {
                        expect(error.issues).toBeDefined();
                        expect(Array.isArray(error.issues)).toBe(true);
                    }
                }
            });
        });
    });

    describe("Integration with Math Operations", () => {
        test("uint8 values work with standard arithmetic", () => {
            const a = uint8(10);
            const b = uint8(20);
            expect(a + b).toBe(30);
            expect(a * b).toBe(200);
        });

        test("uint53 values work with large number operations", () => {
            const a = uint53(Number.MAX_SAFE_INTEGER - 1);
            const result = a + 1;
            expect(result).toBe(Number.MAX_SAFE_INTEGER);
        });

        test("uint62 bigint values work with bitwise operations", () => {
            const a = uint62(0b1010n) as bigint;
            const b = uint62(0b1100n) as bigint;
            expect(a | b).toBe(0b1110n);
            expect(a & b).toBe(0b1000n);
        });
    });

    describe("Practical Use Cases", () => {
        test("uint8 for byte values", () => {
            // Typical byte range validation
            const bytes = [0, 127, 255];
            bytes.forEach(byte => {
                expect(uint8(byte)).toBe(byte);
            });

            // Out of byte range
            expect(() => uint8(256)).toThrow();
            expect(() => uint8(-1)).toThrow();
        });

        test("uint53 for array indices and counts", () => {
            // Valid array operations
            const count = uint53(1000000);
            const index = uint53(999999);
            expect(count).toBeGreaterThan(index);

            // Array length cannot exceed MAX_SAFE_INTEGER
            expect(() => uint53(Number.MAX_SAFE_INTEGER + 1)).toThrow();
        });

        test("uint62 for handling extended ranges", () => {
            // Can handle numbers beyond safe integer range
            const largeValue = uint62(2n ** 50n);
            expect(largeValue).toBe(2n ** 50n);

            // Useful for protocol/format encoding
            const encoding = uint62(0x3FFFFFFFFFFFFFFFn);  // Max 62-bit value
            expect(encoding).toBe(0x3FFFFFFFFFFFFFFFn);
        });
    });

    describe("Negative Cases and Error Paths", () => {
        test("should reject all invalid inputs uniformly", () => {
            const invalidInputs = [
                -1,
                -0.1,
                0.5,
                1.5,
                Number.MAX_VALUE,
                Number.MIN_VALUE,
                NaN,
                Infinity,
                -Infinity,
                'string',
                {},
                [],
                null,
                undefined,
                true,
                false,
            ];

            invalidInputs.forEach(input => {
                expect(() => uint8(input as any)).toThrow();
                expect(() => uint53(input as any)).toThrow();
            });
        });

        test("negative bigints should always fail", () => {
            const negativeBigints = [-1n, -100n, -(2n ** 50n)];
            negativeBigints.forEach(value => {
                expect(() => uint62(value)).toThrow();
            });
        });

        test("should reject bigints exceeding uint62 range", () => {
            const tooLarge = [
                2n ** 62n,
                2n ** 63n,
                2n ** 100n,
                2n ** 1000n,
            ];
            tooLarge.forEach(value => {
                expect(() => uint62(value)).toThrow();
            });
        });
    });
});
