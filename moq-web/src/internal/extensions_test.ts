import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows } from "../../deps.ts";
import { Extensions } from './extensions.ts';

describe('Extensions', () => {
    let extensions: Extensions;

    beforeEach(() => {
        extensions = new Extensions();
    });

    describe('constructor', () => {
        it('should create an empty extensions map', () => {
            assertInstanceOf(extensions.entries, Map);
            assertEquals(extensions.entries.size, 0);
        });
    });

    describe('has', () => {
        it('should return false for non-existent keys', () => {
            expect(extensions.has(1)).toBe(false);
            expect(extensions.has(999)).toBe(false);
        });

        it('should return true for existing keys', () => {
            extensions.addString(1, 'test');
            expect(extensions.has(1)).toBe(true);
        });
    });

    describe('delete', () => {
        it('should return false for non-existent keys', () => {
            expect(extensions.delete(1)).toBe(false);
        });

        it('should delete existing entries and return true', () => {
            extensions.addString(1, 'test');
            expect(extensions.has(1)).toBe(true);
            expect(extensions.delete(1)).toBe(true);
            expect(extensions.has(1)).toBe(false);
        });

        it('should not affect other entries', () => {
            extensions.addString(1, 'test1');
            extensions.addString(2, 'test2');

            extensions.delete(1);

            expect(extensions.has(1)).toBe(false);
            expect(extensions.has(2)).toBe(true);
            expect(extensions.getString(2)).toBe('test2');
        });
    });

    describe('addBytes and getBytes', () => {
        it('should store and retrieve byte arrays', () => {
            const data = new Uint8Array([1, 2, 3, 4, 5]);
            extensions.addBytes(1, data);

            const retrieved = extensions.getBytes(1);
            assertEquals(retrieved, data);
        });

        it('should return undefined for non-existent keys', () => {
            expect(extensions.getBytes(999)).toBeUndefined();
        });

        it('should handle empty byte arrays', () => {
            const data = new Uint8Array([]);
            extensions.addBytes(1, data);

            const retrieved = extensions.getBytes(1);
            assertEquals(retrieved, data);
            assertEquals(retrieved?.length, 0);
        });

        it('should handle large byte arrays', () => {
            const data = new Uint8Array(1000).fill(42);
            extensions.addBytes(1, data);

            const retrieved = extensions.getBytes(1);
            assertEquals(retrieved, data);
            assertEquals(retrieved?.length, 1000);
        });
    });

    describe('addString and getString', () => {
        it('should store and retrieve strings', () => {
            const testString = 'Hello, World!';
            extensions.addString(1, testString);

            const retrieved = extensions.getString(1);
            assertEquals(retrieved, testString);
        });

        it('should return undefined for non-existent keys', () => {
            expect(extensions.getString(999)).toBeUndefined();
        });

        it('should handle empty strings', () => {
            extensions.addString(1, '');

            const retrieved = extensions.getString(1);
            assertEquals(retrieved, '');
        });

        it('should handle Unicode strings', () => {
            const unicodeString = '🚀 Hello, 世界! 🌍';
            extensions.addString(1, unicodeString);

            const retrieved = extensions.getString(1);
            assertEquals(retrieved, unicodeString);
        });

        it('should handle multi-line strings', () => {
            const multilineString = 'Line 1\nLine 2\nLine 3';
            extensions.addString(1, multilineString);

            const retrieved = extensions.getString(1);
            assertEquals(retrieved, multilineString);
        });

        it('should encode and decode correctly', () => {
            const testString = 'Test String';
            extensions.addString(1, testString);

            // Verify internal storage as bytes
            const bytes = extensions.getBytes(1);
            assertExists(bytes);

            // Verify decoder works correctly
            const decoder = new TextDecoder();
            const decodedString = decoder.decode(bytes);
            assertEquals(decodedString, testString);
        });
    });

    describe('addNumber and getNumber', () => {
        it('should store and retrieve bigint numbers', () => {
            const testNumber = 12345678901234567890n;
            extensions.addNumber(1, testNumber);

            const retrieved = extensions.getNumber(1);
            assertEquals(retrieved, testNumber);
        });

        it('should return undefined for non-existent keys', () => {
            expect(extensions.getNumber(999)).toBeUndefined();
        });

        it('should handle zero', () => {
            extensions.addNumber(1, 0n);

            const retrieved = extensions.getNumber(1);
            assertEquals(retrieved, 0n);
        });

        it('should handle maximum safe bigint values', () => {
            const maxValue = 18446744073709551615n; // 2^64 - 1
            extensions.addNumber(1, maxValue);

            const retrieved = extensions.getNumber(1);
            assertEquals(retrieved, maxValue);
        });

        it('should handle small numbers', () => {
            extensions.addNumber(1, 42n);

            const retrieved = extensions.getNumber(1);
            assertEquals(retrieved, 42n);
        });

        it('should return undefined for incorrectly sized byte arrays', () => {
            // Manually add bytes that are not 8 bytes long
            extensions.addBytes(1, new Uint8Array([1, 2, 3])); // 3 bytes instead of 8

            const retrieved = extensions.getNumber(1);
            assertEquals(retrieved, undefined);
        });

        it('should store numbers as 8-byte arrays', () => {
            extensions.addNumber(1, 42n);

            const bytes = extensions.getBytes(1);
            assertEquals(bytes?.length, 8);
        });
    });

    describe('addBoolean and getBoolean', () => {
        it('should store and retrieve true', () => {
            extensions.addBoolean(1, true);

            const retrieved = extensions.getBoolean(1);
            assertEquals(retrieved, true);
        });

        it('should store and retrieve false', () => {
            extensions.addBoolean(1, false);

            const retrieved = extensions.getBoolean(1);
            assertEquals(retrieved, false);
        });

        it('should return undefined for non-existent keys', () => {
            expect(extensions.getBoolean(999)).toBeUndefined();
        });

        it('should return undefined for incorrectly sized byte arrays', () => {
            // Manually add bytes that are not 1 byte long
            extensions.addBytes(1, new Uint8Array([1, 2])); // 2 bytes instead of 1

            const retrieved = extensions.getBoolean(1);
            assertEquals(retrieved, undefined);
        });

        it('should store booleans as single bytes', () => {
            extensions.addBoolean(1, true);

            const bytes = extensions.getBytes(1);
            assertEquals(bytes?.length, 1);
            assertEquals(bytes?.[0], 1);
        });

        it('should store false as zero byte', () => {
            extensions.addBoolean(1, false);

            const bytes = extensions.getBytes(1);
            assertEquals(bytes?.length, 1);
            assertEquals(bytes?.[0], 0);
        });
    });

    describe('mixed operations', () => {
        it('should handle multiple different data types', () => {
            extensions.addString(1, 'Hello');
            extensions.addNumber(2, 42n);
            extensions.addBoolean(3, true);
            extensions.addBytes(4, new Uint8Array([1, 2, 3]));

            expect(extensions.getString(1)).toBe('Hello');
            expect(extensions.getNumber(2)).toBe(42n);
            expect(extensions.getBoolean(3)).toBe(true);
            expect(extensions.getBytes(4)).toEqual(new Uint8Array([1, 2, 3]));
        });

        it('should overwrite existing entries', () => {
            extensions.addString(1, 'First');
            expect(extensions.getString(1)).toBe('First');

            extensions.addString(1, 'Second');
            expect(extensions.getString(1)).toBe('Second');
        });

        it('should handle max key values', () => {
            const largeKey = Number.MAX_SAFE_INTEGER; // 2^53 - 1
            extensions.addString(largeKey, 'Large key test');

            expect(extensions.getString(largeKey)).toBe('Large key test');
            expect(extensions.has(largeKey)).toBe(true);
        });

        it('should maintain separate entries for different keys', () => {
            for (let i = 0; i < 100; i++) {
                extensions.addString(i, `Value ${i}`);
            }

            for (let i = 0; i < 100; i++) {
                expect(extensions.getString(i)).toBe(`Value ${i}`);
            }
        });
    });
});
