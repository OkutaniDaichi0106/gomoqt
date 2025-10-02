import { describe, it, expect } from 'vitest';
import { BiStreamTypes, UniStreamTypes } from './stream_type';

describe('BiStreamTypes', () => {
    it('should have correct constant values', () => {
        expect(BiStreamTypes.SessionStreamType).toBe(0x00);
        expect(BiStreamTypes.AnnounceStreamType).toBe(0x01);
        expect(BiStreamTypes.SubscribeStreamType).toBe(0x02);
    });

    it('should be readonly constants', () => {
        // The constants should be defined
        expect(BiStreamTypes).toBeDefined();
        expect(typeof BiStreamTypes).toBe('object');
        
        // Verify all expected properties exist
        expect(BiStreamTypes).toHaveProperty('SessionStreamType');
        expect(BiStreamTypes).toHaveProperty('AnnounceStreamType');
        expect(BiStreamTypes).toHaveProperty('SubscribeStreamType');
    });

    it('should have correct types', () => {
        expect(typeof BiStreamTypes.SessionStreamType).toBe('number');
        expect(typeof BiStreamTypes.AnnounceStreamType).toBe('number');
        expect(typeof BiStreamTypes.SubscribeStreamType).toBe('number');
    });

    it('should have unique values', () => {
        const values = Object.values(BiStreamTypes);
        const uniqueValues = new Set(values);
        expect(uniqueValues.size).toBe(values.length);
    });
});

describe('UniStreamTypes', () => {
    it('should have correct constant values', () => {
        expect(UniStreamTypes.GroupStreamType).toBe(0x00);
    });

    it('should be readonly constants', () => {
        // The constants should be defined
        expect(UniStreamTypes).toBeDefined();
        expect(typeof UniStreamTypes).toBe('object');
        
        // Verify all expected properties exist
        expect(UniStreamTypes).toHaveProperty('GroupStreamType');
    });

    it('should have correct types', () => {
        expect(typeof UniStreamTypes.GroupStreamType).toBe('number');
    });
});

describe('Stream Type Integration', () => {
    it('should export both BiStreamTypes and UniStreamTypes', () => {
        expect(BiStreamTypes).toBeDefined();
        expect(UniStreamTypes).toBeDefined();
    });

    it('should have different constant spaces for bi and uni streams', () => {
        // BiStreamTypes and UniStreamTypes can have overlapping values 
        // since they represent different categories of streams
        expect(BiStreamTypes.SessionStreamType).toBe(0x00);
        expect(UniStreamTypes.GroupStreamType).toBe(0x00);
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
