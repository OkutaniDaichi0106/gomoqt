import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { Extensions } from './extensions';

describe('Extensions', () => {
    let extensions: Extensions;

    beforeEach(() => {
        extensions = new Extensions();
    });

    describe('constructor', () => {
        it('should create an empty extensions map', () => {
            expect(extensions.entries).toBeInstanceOf(Map);
            expect(extensions.entries.size).toBe(0);
        });
    });

    describe('has', () => {
        it('should return false for non-existent keys', () => {
            expect(extensions.has(1n)).toBe(false);
            expect(extensions.has(999n)).toBe(false);
        });

        it('should return true for existing keys', () => {
            extensions.addString(1n, 'test');
            expect(extensions.has(1n)).toBe(true);
        });
    });

    describe('delete', () => {
        it('should return false for non-existent keys', () => {
            expect(extensions.delete(1n)).toBe(false);
        });

        it('should delete existing entries and return true', () => {
            extensions.addString(1n, 'test');
            expect(extensions.has(1n)).toBe(true);
            expect(extensions.delete(1n)).toBe(true);
            expect(extensions.has(1n)).toBe(false);
        });

        it('should not affect other entries', () => {
            extensions.addString(1n, 'test1');
            extensions.addString(2n, 'test2');
            
            extensions.delete(1n);
            
            expect(extensions.has(1n)).toBe(false);
            expect(extensions.has(2n)).toBe(true);
            expect(extensions.getString(2n)).toBe('test2');
        });
    });

    describe('addBytes and getBytes', () => {
        it('should store and retrieve byte arrays', () => {
            const data = new Uint8Array([1, 2, 3, 4, 5]);
            extensions.addBytes(1n, data);
            
            const retrieved = extensions.getBytes(1n);
            expect(retrieved).toEqual(data);
        });

        it('should return undefined for non-existent keys', () => {
            expect(extensions.getBytes(999n)).toBeUndefined();
        });

        it('should handle empty byte arrays', () => {
            const data = new Uint8Array([]);
            extensions.addBytes(1n, data);
            
            const retrieved = extensions.getBytes(1n);
            expect(retrieved).toEqual(data);
            expect(retrieved?.length).toBe(0);
        });

        it('should handle large byte arrays', () => {
            const data = new Uint8Array(1000).fill(42);
            extensions.addBytes(1n, data);
            
            const retrieved = extensions.getBytes(1n);
            expect(retrieved).toEqual(data);
            expect(retrieved?.length).toBe(1000);
        });
    });

    describe('addString and getString', () => {
        it('should store and retrieve strings', () => {
            const testString = 'Hello, World!';
            extensions.addString(1n, testString);
            
            const retrieved = extensions.getString(1n);
            expect(retrieved).toBe(testString);
        });

        it('should return undefined for non-existent keys', () => {
            expect(extensions.getString(999n)).toBeUndefined();
        });

        it('should handle empty strings', () => {
            extensions.addString(1n, '');
            
            const retrieved = extensions.getString(1n);
            expect(retrieved).toBe('');
        });

        it('should handle Unicode strings', () => {
            const unicodeString = '🚀 Hello, 世界! 🌍';
            extensions.addString(1n, unicodeString);
            
            const retrieved = extensions.getString(1n);
            expect(retrieved).toBe(unicodeString);
        });

        it('should handle multi-line strings', () => {
            const multilineString = 'Line 1\nLine 2\nLine 3';
            extensions.addString(1n, multilineString);
            
            const retrieved = extensions.getString(1n);
            expect(retrieved).toBe(multilineString);
        });

        it('should encode and decode correctly', () => {
            const testString = 'Test String';
            extensions.addString(1n, testString);
            
            // Verify internal storage as bytes
            const bytes = extensions.getBytes(1n);
            expect(bytes).toBeDefined();
            
            // Verify decoder works correctly
            const decoder = new TextDecoder();
            const decodedString = decoder.decode(bytes);
            expect(decodedString).toBe(testString);
        });
    });

    describe('addNumber and getNumber', () => {
        it('should store and retrieve bigint numbers', () => {
            const testNumber = 12345678901234567890n;
            extensions.addNumber(1n, testNumber);
            
            const retrieved = extensions.getNumber(1n);
            expect(retrieved).toBe(testNumber);
        });

        it('should return undefined for non-existent keys', () => {
            expect(extensions.getNumber(999n)).toBeUndefined();
        });

        it('should handle zero', () => {
            extensions.addNumber(1n, 0n);
            
            const retrieved = extensions.getNumber(1n);
            expect(retrieved).toBe(0n);
        });

        it('should handle maximum safe bigint values', () => {
            const maxValue = 18446744073709551615n; // 2^64 - 1
            extensions.addNumber(1n, maxValue);
            
            const retrieved = extensions.getNumber(1n);
            expect(retrieved).toBe(maxValue);
        });

        it('should handle small numbers', () => {
            extensions.addNumber(1n, 42n);
            
            const retrieved = extensions.getNumber(1n);
            expect(retrieved).toBe(42n);
        });

        it('should return undefined for incorrectly sized byte arrays', () => {
            // Manually add bytes that are not 8 bytes long
            extensions.addBytes(1n, new Uint8Array([1, 2, 3])); // 3 bytes instead of 8
            
            const retrieved = extensions.getNumber(1n);
            expect(retrieved).toBeUndefined();
        });

        it('should store numbers as 8-byte arrays', () => {
            extensions.addNumber(1n, 42n);
            
            const bytes = extensions.getBytes(1n);
            expect(bytes?.length).toBe(8);
        });
    });

    describe('addBoolean and getBoolean', () => {
        it('should store and retrieve true', () => {
            extensions.addBoolean(1n, true);
            
            const retrieved = extensions.getBoolean(1n);
            expect(retrieved).toBe(true);
        });

        it('should store and retrieve false', () => {
            extensions.addBoolean(1n, false);
            
            const retrieved = extensions.getBoolean(1n);
            expect(retrieved).toBe(false);
        });

        it('should return undefined for non-existent keys', () => {
            expect(extensions.getBoolean(999n)).toBeUndefined();
        });

        it('should return undefined for incorrectly sized byte arrays', () => {
            // Manually add bytes that are not 1 byte long
            extensions.addBytes(1n, new Uint8Array([1, 2])); // 2 bytes instead of 1
            
            const retrieved = extensions.getBoolean(1n);
            expect(retrieved).toBeUndefined();
        });

        it('should store booleans as single bytes', () => {
            extensions.addBoolean(1n, true);
            
            const bytes = extensions.getBytes(1n);
            expect(bytes?.length).toBe(1);
            expect(bytes?.[0]).toBe(1);
        });

        it('should store false as zero byte', () => {
            extensions.addBoolean(1n, false);
            
            const bytes = extensions.getBytes(1n);
            expect(bytes?.length).toBe(1);
            expect(bytes?.[0]).toBe(0);
        });
    });

    describe('mixed operations', () => {
        it('should handle multiple different data types', () => {
            extensions.addString(1n, 'Hello');
            extensions.addNumber(2n, 42n);
            extensions.addBoolean(3n, true);
            extensions.addBytes(4n, new Uint8Array([1, 2, 3]));
            
            expect(extensions.getString(1n)).toBe('Hello');
            expect(extensions.getNumber(2n)).toBe(42n);
            expect(extensions.getBoolean(3n)).toBe(true);
            expect(extensions.getBytes(4n)).toEqual(new Uint8Array([1, 2, 3]));
        });

        it('should overwrite existing entries', () => {
            extensions.addString(1n, 'First');
            expect(extensions.getString(1n)).toBe('First');
            
            extensions.addString(1n, 'Second');
            expect(extensions.getString(1n)).toBe('Second');
        });

        it('should handle large key values', () => {
            const largeKey = 9223372036854775807n; // 2^63 - 1
            extensions.addString(largeKey, 'Large key test');
            
            expect(extensions.getString(largeKey)).toBe('Large key test');
            expect(extensions.has(largeKey)).toBe(true);
        });

        it('should maintain separate entries for different keys', () => {
            for (let i = 0; i < 100; i++) {
                extensions.addString(BigInt(i), `Value ${i}`);
            }
            
            for (let i = 0; i < 100; i++) {
                expect(extensions.getString(BigInt(i))).toBe(`Value ${i}`);
            }
        });
    });
});
