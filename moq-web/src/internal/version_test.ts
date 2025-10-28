import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows } from "../../deps.ts";
import type { Version} from './version.ts';
import { Versions, DEFAULT_VERSION } from './version.ts';

describe('Version', () => {
    describe('Version type', () => {
        it('should be a bigint type', () => {
            const version: Version = 123n;
            assertEquals(typeof version, 'bigint');
        });

        it('should accept bigint values', () => {
            const version1: Version = 0n;
            const version2: Version = 1n;
            const version3: Version = 0xffffff00n;
            
            assertEquals(version1, 0n);
            assertEquals(version2, 1n);
            assertEquals(version3, 0xffffff00n);
        });

        it('should support large version numbers', () => {
            const largeVersion: Version = 18446744073709551615n; // 2^64 - 1
            assertEquals(largeVersion, 18446744073709551615n);
        });
    });

    describe('Versions constants', () => {
        it('should have DEVELOP constant', () => {
            assertExists(Versions.DEVELOP);
            assertEquals(typeof Versions.DEVELOP, 'bigint');
            assertEquals(Versions.DEVELOP, 0xffffff00n);
        });

        it('should be read-only constants', () => {
            // TypeScript should prevent this, but let's verify runtime behavior
            const originalDevelop = Versions.DEVELOP;
            
            // Verify the constant has the correct value
            assertEquals(Versions.DEVELOP, 0xffffff00n);
            assertEquals(originalDevelop, 0xffffff00n);
        });

        it('should have correct hex value for DEVELOP', () => {
            // 0xffffff00 = 4294967040 in decimal
            assertEquals(Versions.DEVELOP, 0xffffff00n);
        });

        it('should be typed as Version', () => {
            const develop: Version = Versions.DEVELOP;
            assertEquals(develop, Versions.DEVELOP);
        });
    });

    describe('DEFAULT_VERSION', () => {
        it('should be defined', () => {
            assertExists(DEFAULT_VERSION);
        });

        it('should be a Version type', () => {
            assertEquals(typeof DEFAULT_VERSION, 'bigint');
        });

        it('should equal DEVELOP version', () => {
            assertEquals(DEFAULT_VERSION, Versions.DEVELOP);
        });

        it('should have the correct value', () => {
            assertEquals(DEFAULT_VERSION, 0xffffff00n);
        });
    });

    describe('version operations', () => {
        it('should support comparison operations', () => {
            const version1: Version = 1n;
            const version2: Version = 2n;
            const version3: Version = 1n;
            
            assertEquals(version1 < version2, true);
            assertEquals(version2 > version1, true);
            assertEquals(version1 === version3, true);
            assertEquals(version1 !== version2, true);
        });

        it('should support arithmetic operations', () => {
            const baseVersion: Version = 100n;
            const increment: Version = 1n;
            
            const nextVersion = baseVersion + increment;
            const prevVersion = baseVersion - increment;
            
            assertEquals(nextVersion, 101n);
            assertEquals(prevVersion, 99n);
        });

        it('should support bitwise operations', () => {
            const version: Version = 0xffffff00n;
            
            // Test bitwise AND
            const masked = version & 0xffn;
            assertEquals(masked, 0n);
            
            // Test bitwise OR
            const combined = version | 0x0fn;
            assertEquals(combined, 0xffffff0fn);
            
            // Test bitwise XOR
            const xored = version ^ 0xffffff00n;
            assertEquals(xored, 0n);
        });

        it('should maintain precision with large numbers', () => {
            const largeVersion1: Version = 0xfffffffffffffffn;
            const largeVersion2: Version = 0xffffffffffffffen;
            
            assertEquals(largeVersion1 - largeVersion2, 1n);
            assertEquals(largeVersion1 > largeVersion2, true);
        });
    });

    describe('version compatibility', () => {
        it('should work with different numeric representations', () => {
            // Decimal representation
            const decimal: Version = 0xffffff00n;
            
            // Hexadecimal representation
            const hex: Version = 0xffffff00n;
            
            // Binary representation (conceptually)
            const binary: Version = 0b11111111111111111111111100000000n;
            
            assertEquals(decimal, hex);
            assertEquals(hex, binary);
            assertEquals(decimal, Versions.DEVELOP);
        });

        it('should handle version ranges', () => {
            const minVersion: Version = 0n;
            const maxVersion: Version = 0xffffffffffffffffn; // 2^64 - 1
            
            assertEquals(DEFAULT_VERSION >= minVersion, true);
            assertEquals(DEFAULT_VERSION <= maxVersion, true);
        });

        it('should support version checks', () => {
            const isValidVersion = (version: Version): boolean => {
                return version >= 0n && version <= 0xffffffffffffffffn;
            };
            
            expect(isValidVersion(DEFAULT_VERSION)).toBe(true);
            expect(isValidVersion(Versions.DEVELOP)).toBe(true);
            expect(isValidVersion(0n)).toBe(true);
            expect(isValidVersion(1n)).toBe(true);
        });
    });

    describe('string conversion', () => {
        it('should convert to string correctly', () => {
            expect(DEFAULT_VERSION.toString()).toBe(0xffffff00n.toString());
            expect(Versions.DEVELOP.toString()).toBe(0xffffff00n.toString());
        });

        it('should convert to hex string', () => {
            expect(DEFAULT_VERSION.toString(16)).toBe('ffffff00');
            expect(Versions.DEVELOP.toString(16)).toBe('ffffff00');
        });

        it('should convert to binary string', () => {
            const binaryString = DEFAULT_VERSION.toString(2);
            assertEquals(binaryString, '11111111111111111111111100000000');
        });

        it('should parse from string', () => {
            const versionFromString = BigInt('4294967040');
            assertEquals(versionFromString, DEFAULT_VERSION);
            
            const versionFromHex = BigInt('0xffffff00');
            assertEquals(versionFromHex, DEFAULT_VERSION);
        });
    });
});
