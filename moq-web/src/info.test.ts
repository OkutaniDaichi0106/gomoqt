import { describe, it, expect } from '@jest/globals';
import type { Info } from './info';

describe('Info', () => {
    it('should be defined as a type', () => {
        // This test ensures the Info interface is properly exported and can be used
        const info: Info = {};
        
        // Since Info is currently an empty interface, we can only verify it exists
        expect(typeof info).toBe('object');
    });

    it('should allow creating empty Info objects', () => {
        // Test that we can create an empty Info object since it's currently an empty interface
        const info: Info = {};
        
        expect(info).toBeDefined();
        expect(info).toEqual({});
    });

    it('should be assignable to object type', () => {
        // Verify that Info objects are valid objects
        const info: Info = {};
        const obj: object = info;
        
        expect(obj).toBe(info);
    });
});
