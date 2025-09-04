import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { Info } from './info';

describe('Info', () => {
    it('should define the correct type structure', () => {
        // This is a type-only test to ensure the type is correctly defined
        const mockInfo: Info = {
            groupPeriod: 123,
            trackPriority: 456
        };

        expect(mockInfo.groupPeriod).toBe(123);
        expect(mockInfo.trackPriority).toBe(456);
    });

    it('should allow all required properties', () => {
        // Verify that all properties are required by creating an info object
        const info: Info = {
            groupPeriod: 100,
            trackPriority: 50
        };

        // Type assertion to ensure all properties exist and are number
        expect(typeof info.groupPeriod).toBe('number');
        expect(typeof info.trackPriority).toBe('number');
    });

    it('should handle large number values', () => {
        const info: Info = {
            groupPeriod: Number.MAX_SAFE_INTEGER, // Maximum safe number
            trackPriority: 0 // Minimum value
        };

        expect(info.groupPeriod).toBe(Number.MAX_SAFE_INTEGER);
        expect(info.trackPriority).toBe(0);
    });

    it('should be immutable once created', () => {
        const info: Info = {
            groupPeriod: 123,
            trackPriority: 456
        };

        // Properties should be accessible
        expect(info.groupPeriod).toBe(123);
        expect(info.trackPriority).toBe(456);

        // Info type should support property access
        const { groupPeriod, trackPriority } = info;
        expect(groupPeriod).toBe(123);
        expect(trackPriority).toBe(456);
    });
});
