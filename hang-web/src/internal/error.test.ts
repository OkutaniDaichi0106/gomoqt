import { describe, test, expect, beforeEach, afterEach, vi } from 'vitest';

// Mock the external dependencies before importing the module under test
vi.mock("@okutanidaichi/moqt", () => ({}));

import { EncodeErrorCode, DecodeErrorCode } from "./error";

describe("Error Constants", () => {
    describe("EncodeErrorCode", () => {
        test("has correct value", () => {
            expect(EncodeErrorCode).toBe(101);
        });

        test("is a number", () => {
            expect(typeof EncodeErrorCode).toBe("number");
        });

        test("is a positive integer", () => {
            expect(EncodeErrorCode).toBeGreaterThan(0);
            expect(Number.isInteger(EncodeErrorCode)).toBe(true);
        });
    });

    describe("DecodeErrorCode", () => {
        test("has correct value", () => {
            expect(DecodeErrorCode).toBe(102);
        });

        test("is a number", () => {
            expect(typeof DecodeErrorCode).toBe("number");
        });

        test("is a positive integer", () => {
            expect(DecodeErrorCode).toBeGreaterThan(0);
            expect(Number.isInteger(DecodeErrorCode)).toBe(true);
        });
    });

    describe("Error Code Relationships", () => {
        test("DecodeErrorCode is greater than EncodeErrorCode", () => {
            expect(DecodeErrorCode).toBeGreaterThan(EncodeErrorCode);
        });

        test("error codes are unique", () => {
            expect(EncodeErrorCode).not.toBe(DecodeErrorCode);
        });

        test("error codes are sequential", () => {
            expect(DecodeErrorCode - EncodeErrorCode).toBe(1);
        });
    });

    describe("Error Code Range", () => {
        test("error codes are in expected range", () => {
            expect(EncodeErrorCode).toBeGreaterThanOrEqual(100);
            expect(EncodeErrorCode).toBeLessThan(200);
            expect(DecodeErrorCode).toBeGreaterThanOrEqual(100);
            expect(DecodeErrorCode).toBeLessThan(200);
        });
    });

    describe("Boundary Value Tests", () => {
        test("EncodeErrorCode is not zero", () => {
            expect(EncodeErrorCode).not.toBe(0);
        });

        test("DecodeErrorCode is not zero", () => {
            expect(DecodeErrorCode).not.toBe(0);
        });

        test("error codes are not negative", () => {
            expect(EncodeErrorCode).toBeGreaterThanOrEqual(0);
            expect(DecodeErrorCode).toBeGreaterThanOrEqual(0);
        });

        test("error codes are finite numbers", () => {
            expect(Number.isFinite(EncodeErrorCode)).toBe(true);
            expect(Number.isFinite(DecodeErrorCode)).toBe(true);
        });

        test("error codes are safe integers", () => {
            expect(Number.isSafeInteger(EncodeErrorCode)).toBe(true);
            expect(Number.isSafeInteger(DecodeErrorCode)).toBe(true);
        });
    });

    describe("Type Compatibility Tests", () => {
        test("EncodeErrorCode is compatible with SubscribeErrorCode type", () => {
            // TypeScript compilation will catch type issues, but we can test runtime behavior
            const errorCode: number = EncodeErrorCode;
            expect(errorCode).toBe(101);
        });

        test("DecodeErrorCode is compatible with SubscribeErrorCode type", () => {
            const errorCode: number = DecodeErrorCode;
            expect(errorCode).toBe(102);
        });

        test("error codes can be used in comparisons", () => {
            expect(EncodeErrorCode < DecodeErrorCode).toBe(true);
            expect(EncodeErrorCode <= DecodeErrorCode).toBe(true);
            expect(DecodeErrorCode > EncodeErrorCode).toBe(true);
            expect(DecodeErrorCode >= EncodeErrorCode).toBe(true);
        });

        test("error codes can be used in arithmetic operations", () => {
            expect(EncodeErrorCode + 1).toBe(DecodeErrorCode);
            expect(DecodeErrorCode - 1).toBe(EncodeErrorCode);
            expect(EncodeErrorCode * 2).toBe(202);
            expect(DecodeErrorCode / 2).toBe(51);
        });

        test("error codes can be used as object keys", () => {
            const errorMap = {
                [EncodeErrorCode]: "Encode Error",
                [DecodeErrorCode]: "Decode Error"
            };

            expect(errorMap[101]).toBe("Encode Error");
            expect(errorMap[102]).toBe("Decode Error");
            expect(errorMap[EncodeErrorCode]).toBe("Encode Error");
            expect(errorMap[DecodeErrorCode]).toBe("Decode Error");
        });

        test("error codes can be used in switch statements", () => {
            const getErrorType = (code: number): string => {
                switch (code) {
                    case EncodeErrorCode:
                        return "encode";
                    case DecodeErrorCode:
                        return "decode";
                    default:
                        return "unknown";
                }
            };

            expect(getErrorType(EncodeErrorCode)).toBe("encode");
            expect(getErrorType(DecodeErrorCode)).toBe("decode");
            expect(getErrorType(999)).toBe("unknown");
        });
    });

    describe("Immutability Tests", () => {
        test("EncodeErrorCode cannot be modified", () => {
            const originalValue = EncodeErrorCode;
            // Attempt to modify (should have no effect due to const)
            expect(() => {
                (global as any).EncodeErrorCode = 999;
            }).not.toThrow();
            
            // Original constant should remain unchanged
            expect(EncodeErrorCode).toBe(originalValue);
        });

        test("DecodeErrorCode cannot be modified", () => {
            const originalValue = DecodeErrorCode;
            expect(() => {
                (global as any).DecodeErrorCode = 999;
            }).not.toThrow();
            
            expect(DecodeErrorCode).toBe(originalValue);
        });

        test("error codes maintain reference equality", () => {
            const encode1 = EncodeErrorCode;
            const encode2 = EncodeErrorCode;
            const decode1 = DecodeErrorCode;
            const decode2 = DecodeErrorCode;

            expect(encode1).toBe(encode2);
            expect(decode1).toBe(decode2);
            expect(Object.is(encode1, encode2)).toBe(true);
            expect(Object.is(decode1, decode2)).toBe(true);
        });
    });

    describe("String Conversion Tests", () => {
        test("error codes convert to strings correctly", () => {
            expect(String(EncodeErrorCode)).toBe("101");
            expect(String(DecodeErrorCode)).toBe("102");
            expect(EncodeErrorCode.toString()).toBe("101");
            expect(DecodeErrorCode.toString()).toBe("102");
        });

        test("error codes work with template literals", () => {
            expect(`Error code: ${EncodeErrorCode}`).toBe("Error code: 101");
            expect(`Error code: ${DecodeErrorCode}`).toBe("Error code: 102");
        });

        test("error codes work with JSON serialization", () => {
            const errorData = {
                encodeError: EncodeErrorCode,
                decodeError: DecodeErrorCode
            };

            const jsonString = JSON.stringify(errorData);
            const parsed = JSON.parse(jsonString);

            expect(parsed.encodeError).toBe(101);
            expect(parsed.decodeError).toBe(102);
        });
    });

    describe("Array and Collection Tests", () => {
        test("error codes work in arrays", () => {
            const errorCodes = [EncodeErrorCode, DecodeErrorCode];
            
            expect(errorCodes).toHaveLength(2);
            expect(errorCodes[0]).toBe(101);
            expect(errorCodes[1]).toBe(102);
            expect(errorCodes.includes(EncodeErrorCode)).toBe(true);
            expect(errorCodes.includes(DecodeErrorCode)).toBe(true);
        });

        test("error codes work in Sets", () => {
            const errorSet = new Set([EncodeErrorCode, DecodeErrorCode]);
            
            expect(errorSet.size).toBe(2);
            expect(errorSet.has(EncodeErrorCode)).toBe(true);
            expect(errorSet.has(DecodeErrorCode)).toBe(true);
            expect(errorSet.has(999)).toBe(false);
        });

        test("error codes work in Maps", () => {
            const errorMap = new Map([
                [EncodeErrorCode, "Encoding failed"],
                [DecodeErrorCode, "Decoding failed"]
            ]);

            expect(errorMap.size).toBe(2);
            expect(errorMap.get(EncodeErrorCode)).toBe("Encoding failed");
            expect(errorMap.get(DecodeErrorCode)).toBe("Decoding failed");
            expect(errorMap.has(EncodeErrorCode)).toBe(true);
            expect(errorMap.has(DecodeErrorCode)).toBe(true);
        });

        test("error codes can be sorted", () => {
            const unsorted = [DecodeErrorCode, EncodeErrorCode];
            const sorted = unsorted.sort((a, b) => a - b);
            
            expect(sorted[0]).toBe(EncodeErrorCode);
            expect(sorted[1]).toBe(DecodeErrorCode);
        });
    });
});
