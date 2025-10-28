import { describe, it, assertEquals, assertExists } from "../deps.ts";
import { BiStreamTypes, UniStreamTypes } from './stream_type.ts';

describe('BiStreamTypes', () => {
    it('should have correct constant values', () => {
        assertEquals(BiStreamTypes.SessionStreamType, 0x00);
        assertEquals(BiStreamTypes.AnnounceStreamType, 0x01);
        assertEquals(BiStreamTypes.SubscribeStreamType, 0x02);
    });

    it('should be readonly constants', () => {
        // The constants should be defined
        assertExists(BiStreamTypes);
        assertEquals(typeof BiStreamTypes, 'object');
        
        // Verify all expected properties exist
        expect(BiStreamTypes).toHaveProperty('SessionStreamType');
        expect(BiStreamTypes).toHaveProperty('AnnounceStreamType');
        expect(BiStreamTypes).toHaveProperty('SubscribeStreamType');
    });

    it('should have correct types', () => {
        assertEquals(typeof BiStreamTypes.SessionStreamType, 'number');
        assertEquals(typeof BiStreamTypes.AnnounceStreamType, 'number');
        assertEquals(typeof BiStreamTypes.SubscribeStreamType, 'number');
    });

    it('should have unique values', () => {
        const values = Object.values(BiStreamTypes);
        const uniqueValues = new Set(values);
        assertEquals(uniqueValues.size, values.length);
    });
});

describe('UniStreamTypes', () => {
    it('should have correct constant values', () => {
        assertEquals(UniStreamTypes.GroupStreamType, 0x00);
    });

    it('should be readonly constants', () => {
        // The constants should be defined
        assertExists(UniStreamTypes);
        assertEquals(typeof UniStreamTypes, 'object');
        
        // Verify all expected properties exist
        expect(UniStreamTypes).toHaveProperty('GroupStreamType');
    });

    it('should have correct types', () => {
        assertEquals(typeof UniStreamTypes.GroupStreamType, 'number');
    });
});

describe('Stream Type Integration', () => {
    it('should export both BiStreamTypes and UniStreamTypes', () => {
        assertExists(BiStreamTypes);
        assertExists(UniStreamTypes);
    });

    it('should have different constant spaces for bi and uni streams', () => {
        // BiStreamTypes and UniStreamTypes can have overlapping values 
        // since they represent different categories of streams
        assertEquals(BiStreamTypes.SessionStreamType, 0x00);
        assertEquals(UniStreamTypes.GroupStreamType, 0x00);
        // This is expected and correct - they are in different namespaces
    });

    it('should be usable in switch statements', () => {
        // Test that the constants can be used in switch statements
        const testBiStreamType = (type: number): string => {
            switch (type) {
                case BiStreamTypes.SessionStreamType:
                    return 'session';
                case BiStreamTypes.AnnounceStreamType:
                    return 'announce';
                case BiStreamTypes.SubscribeStreamType:
                    return 'subscribe';
                default:
                    return 'unknown';
            }
        };

        const testUniStreamType = (type: number): string => {
            switch (type) {
                case UniStreamTypes.GroupStreamType:
                    return 'group';
                default:
                    return 'unknown';
            }
        };

        expect(testBiStreamType(BiStreamTypes.SessionStreamType)).toBe('session');
        expect(testBiStreamType(BiStreamTypes.AnnounceStreamType)).toBe('announce');
        expect(testBiStreamType(BiStreamTypes.SubscribeStreamType)).toBe('subscribe');
        expect(testBiStreamType(999)).toBe('unknown');

        expect(testUniStreamType(UniStreamTypes.GroupStreamType)).toBe('group');
        expect(testUniStreamType(999)).toBe('unknown');
    });
});
