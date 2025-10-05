import { 
    uint8Schema, 
    uint8, 
    uint53Schema, 
    uint53, 
    uint62Schema, 
    uint62 
} from "./integers";
import { z } from "zod";
import { describe, test, expect, beforeEach, it } from 'vitest';

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
                fail("Expected error to be thrown");
            } catch (error) {
                expect(error).toBeInstanceOf(z.ZodError);
                expect((error as z.ZodError).issues).toHaveLength(1);
            }

            try {
                uint53(-1);
                fail("Expected error to be thrown");
            } catch (error) {
                expect(error).toBeInstanceOf(z.ZodError);
                expect((error as z.ZodError).issues).toHaveLength(1);
            }

            try {
                uint62(2n ** 62n);
                fail("Expected error to be thrown");
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
});
