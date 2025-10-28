
import { describe, it, expect } from 'vitest';
import { isValidPrefix, validateTrackPrefix } from './track_prefix';

describe('TrackPrefix', () => {
    describe('isValidPrefix', () => {
        it('should return true for valid prefixes', () => {
            expect(isValidPrefix('/foo/')).toBe(true);
            expect(isValidPrefix('//')).toBe(true);
            expect(isValidPrefix('/')).toBe(true);
        });

        it('should return false for invalid prefixes', () => {
            expect(isValidPrefix('foo/')).toBe(false);
            expect(isValidPrefix('/foo')).toBe(false);
            expect(isValidPrefix('foo')).toBe(false);
            expect(isValidPrefix('')).toBe(false);
        });
    });

    describe('validateTrackPrefix', () => {
        it('should return the prefix for valid prefixes', () => {
            expect(validateTrackPrefix('/foo/')).toBe('/foo/');
            expect(validateTrackPrefix('//')).toBe('//');
            expect(validateTrackPrefix('/')).toBe('/');
        });

        it('should throw an error for invalid prefixes', () => {
            expect(() => validateTrackPrefix('foo/')).toThrow();
            expect(() => validateTrackPrefix('/foo')).toThrow();
            expect(() => validateTrackPrefix('foo')).toThrow();
            expect(() => validateTrackPrefix('')).toThrow();
        });
    });
});
