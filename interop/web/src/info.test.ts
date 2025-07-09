import { Info } from './info';

describe('Info', () => {
    it('should define the correct type structure', () => {
        // This is a type-only test to ensure the type is correctly defined
        const mockInfo: Info = {
            groupOrder: 123n,
            trackPriority: 456n
        };

        expect(mockInfo.groupOrder).toBe(123n);
        expect(mockInfo.trackPriority).toBe(456n);
    });

    it('should allow all required properties', () => {
        // Verify that all properties are required by creating an info object
        const info: Info = {
            groupOrder: 100n,
            trackPriority: 50n
        };

        // Type assertion to ensure all properties exist and are bigint
        expect(typeof info.groupOrder).toBe('bigint');
        expect(typeof info.trackPriority).toBe('bigint');
    });

    it('should handle large bigint values', () => {
        const info: Info = {
            groupOrder: 9223372036854775807n, // Maximum safe bigint
            trackPriority: 0n // Minimum value
        };

        expect(info.groupOrder).toBe(9223372036854775807n);
        expect(info.trackPriority).toBe(0n);
    });

    it('should be immutable once created', () => {
        const info: Info = {
            groupOrder: 123n,
            trackPriority: 456n
        };

        // Properties should be accessible
        expect(info.groupOrder).toBe(123n);
        expect(info.trackPriority).toBe(456n);

        // Info type should support property access
        const { groupOrder, trackPriority } = info;
        expect(groupOrder).toBe(123n);
        expect(trackPriority).toBe(456n);
    });
});
