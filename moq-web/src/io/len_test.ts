import { describe, expect, it } from "../../deps.ts";
import { MAX_VARINT1, MAX_VARINT2, MAX_VARINT4, MAX_VARINT8, varintLen, stringLen, bytesLen } from "./len.js";

describe("len", () => {
    describe("varintLen", () => {
        it("should return 1 for values <= MAX_VARINT1", () => {
            expect(varintLen(0)).toBe(1);
            expect(varintLen(63)).toBe(1);
            expect(varintLen(MAX_VARINT1)).toBe(1);
        });

        it("should return 2 for values <= MAX_VARINT2", () => {
            expect(varintLen(64)).toBe(2);
            expect(varintLen(16383)).toBe(2);
            expect(varintLen(MAX_VARINT2)).toBe(2);
        });

        it("should return 4 for values <= MAX_VARINT4", () => {
            expect(varintLen(16384)).toBe(4);
            expect(varintLen(1073741823)).toBe(4);
            expect(varintLen(MAX_VARINT4)).toBe(4);
        });

        it("should return 8 for values <= MAX_VARINT8", () => {
            expect(varintLen(1073741824)).toBe(8);
            expect(varintLen(BigInt("4611686018427387903"))).toBe(8);
            expect(varintLen(MAX_VARINT8)).toBe(8);
        });

        it("should handle negative numbers", () => {
            expect(() => varintLen(-1)).toThrow(RangeError);
        });

        it("should throw for values > MAX_VARINT8", () => {
            expect(() => varintLen(BigInt("4611686018427387904"))).toThrow(RangeError);
        });
    });

    describe("stringLen", () => {
        it("should calculate length for empty string", () => {
            expect(stringLen("")).toBe(1); // varintLen(0) = 1
        });

        it("should calculate length for short string", () => {
            expect(stringLen("a")).toBe(2); // 1 (varint) + 1 (char)
        });

        it("should calculate length for longer string", () => {
            const str = "hello world";
            expect(stringLen(str)).toBe(1 + str.length); // varintLen(11) = 1
        });
    });

    describe("bytesLen", () => {
        it("should calculate length for empty bytes", () => {
            expect(bytesLen(new Uint8Array(0))).toBe(1); // varintLen(0) = 1
        });

        it("should calculate length for bytes", () => {
            const bytes = new Uint8Array([1, 2, 3]);
            expect(bytesLen(bytes)).toBe(1 + 3); // varintLen(3) = 1 + 3
        });
    });
});


describe('io len utilities', () => {
  it('varintLen small values', () => {
    expect(varintLen(0)).toBe(1);
    expect(varintLen(63)).toBe(1);
    expect(varintLen(64)).toBe(2);
    expect(varintLen(16383)).toBe(2);
    expect(varintLen(16384)).toBe(4);
  });

  it('stringLen includes varint header', () => {
    const s = 'abcd';
    expect(stringLen(s)).toBe(1 + s.length);
  });

  it('bytesLen includes varint header', () => {
    const b = new Uint8Array(10);
    expect(bytesLen(b)).toBe(1 + b.length);
  });
});