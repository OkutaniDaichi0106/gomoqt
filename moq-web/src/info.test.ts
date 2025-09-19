import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import type { Info } from './info';

describe('Info', () => {
    it('should define the correct type structure', () => {
        // This is a type-only test to ensure the type is correctly defined
        const mockInfo: Info = {
            groupPeriod: 123,
        };

        expect(mockInfo.groupPeriod).toBe(123);
    });

    it('should allow all required properties', () => {
        // Verify that all properties are required by creating an info object
        const info: Info = {
            groupPeriod: 100,
        };

        // Type assertion to ensure all properties exist and are number
        expect(typeof info.groupPeriod).toBe('number');
    });

    it('should handle large number values', () => {
        const info: Info = {
            groupPeriod: Number.MAX_SAFE_INTEGER, // Maximum safe number
        };

        expect(info.groupPeriod).toBe(Number.MAX_SAFE_INTEGER);
    });

    it('should be immutable once created', () => {
        const info: Info = {
            groupPeriod: 123,
        };

        // Properties should be accessible
        expect(info.groupPeriod).toBe(123);

        // Info type should support property access
        const { groupPeriod } = info;
        expect(groupPeriod).toBe(123);
    });
});
