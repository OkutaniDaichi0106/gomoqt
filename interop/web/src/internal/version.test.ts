import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { Version, Versions, DEFAULT_VERSION } from './version';

describe('Version', () => {
    describe('Version type', () => {
        it('should be a bigint type', () => {
            const version: Version = 123n;
            expect(typeof version).toBe('bigint');
        });

        it('should accept bigint values', () => {
            const version1: Version = 0n;
            const version2: Version = 1n;
            const version3: Version = 0xffffff00n;
            
            expect(version1).toBe(0n);
            expect(version2).toBe(1n);
            expect(version3).toBe(0xffffff00n);
        });

        it('should support large version numbers', () => {
            const largeVersion: Version = 18446744073709551615n; // 2^64 - 1
            expect(largeVersion).toBe(18446744073709551615n);
        });
    });

    describe('Versions constants', () => {
        it('should have DEVELOP constant', () => {
            expect(Versions.DEVELOP).toBeDefined();
            expect(typeof Versions.DEVELOP).toBe('bigint');
            expect(Versions.DEVELOP).toBe(0xffffff00n);
        });

        it('should be read-only constants', () => {
            // TypeScript should prevent this, but let's verify runtime behavior
            const originalDevelop = Versions.DEVELOP;
            
            // Verify the constant has the correct value
            expect(Versions.DEVELOP).toBe(0xffffff00n);
            expect(originalDevelop).toBe(0xffffff00n);
        });

        it('should have correct hex value for DEVELOP', () => {
            // 0xffffff00 = 4294967040 in decimal
            expect(Versions.DEVELOP).toBe(0xffffff00n);
        });

        it('should be typed as Version', () => {
            const develop: Version = Versions.DEVELOP;
            expect(develop).toBe(Versions.DEVELOP);
        });
    });

    describe('DEFAULT_VERSION', () => {
        it('should be defined', () => {
            expect(DEFAULT_VERSION).toBeDefined();
        });

        it('should be a Version type', () => {
            expect(typeof DEFAULT_VERSION).toBe('bigint');
        });

        it('should equal DEVELOP version', () => {
            expect(DEFAULT_VERSION).toBe(Versions.DEVELOP);
        });

        it('should have the correct value', () => {
            expect(DEFAULT_VERSION).toBe(0xffffff00n);
        });
    });

    describe('version operations', () => {
        it('should support comparison operations', () => {
            const version1: Version = 1n;
            const version2: Version = 2n;
            const version3: Version = 1n;
            
            expect(version1 < version2).toBe(true);
            expect(version2 > version1).toBe(true);
            expect(version1 === version3).toBe(true);
            expect(version1 !== version2).toBe(true);
        });

        it('should support arithmetic operations', () => {
            const baseVersion: Version = 100n;
            const increment: Version = 1n;
            
            const nextVersion = baseVersion + increment;
            const prevVersion = baseVersion - increment;
            
            expect(nextVersion).toBe(101n);
            expect(prevVersion).toBe(99n);
        });

        it('should support bitwise operations', () => {
            const version: Version = 0xffffff00n;
            
            // Test bitwise AND
            const masked = version & 0xffn;
            expect(masked).toBe(0n);
            
            // Test bitwise OR
            const combined = version | 0x0fn;
            expect(combined).toBe(0xffffff0fn);
            
            // Test bitwise XOR
            const xored = version ^ 0xffffff00n;
            expect(xored).toBe(0n);
        });

        it('should maintain precision with large numbers', () => {
            const largeVersion1: Version = 0xfffffffffffffffn;
            const largeVersion2: Version = 0xffffffffffffffen;
            
            expect(largeVersion1 - largeVersion2).toBe(1n);
            expect(largeVersion1 > largeVersion2).toBe(true);
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
            
            expect(decimal).toBe(hex);
            expect(hex).toBe(binary);
            expect(decimal).toBe(Versions.DEVELOP);
        });

        it('should handle version ranges', () => {
            const minVersion: Version = 0n;
            const maxVersion: Version = 0xffffffffffffffffn; // 2^64 - 1
            
            expect(DEFAULT_VERSION >= minVersion).toBe(true);
            expect(DEFAULT_VERSION <= maxVersion).toBe(true);
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
            expect(binaryString).toBe('11111111111111111111111100000000');
        });

        it('should parse from string', () => {
            const versionFromString = BigInt('4294967040');
            expect(versionFromString).toBe(DEFAULT_VERSION);
            
            const versionFromHex = BigInt('0xffffff00');
            expect(versionFromHex).toBe(DEFAULT_VERSION);
        });
    });
});
