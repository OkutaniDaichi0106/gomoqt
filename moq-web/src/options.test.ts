import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { MOQOptions } from './options';
import { Extensions } from './internal/extensions';

describe('MOQOptions', () => {
    it('should define the correct interface structure', () => {
        // This is a type-only test to ensure the interface is correctly defined
        const mockOptions: MOQOptions = {
            extensions: new Extensions()
        };

        expect(mockOptions.extensions).toBeDefined();
        expect(mockOptions.extensions).toBeInstanceOf(Extensions);
    });

    it('should allow empty options', () => {
        // Extensions should be optional
        const emptyOptions: MOQOptions = {};

        expect(emptyOptions.extensions).toBeUndefined();
    });

    it('should allow options with extensions', () => {
        const extensions = new Extensions();
        extensions.addString(1, 'test');

        const options: MOQOptions = {
            extensions: extensions
        };

        expect(options.extensions).toBe(extensions);
        expect(options.extensions?.getString(1)).toBe('test');
    });

    it('should support partial assignment', () => {
        // Should be able to create options incrementally
        const options: MOQOptions = {};

        // Initially no extensions
        expect(options.extensions).toBeUndefined();

        // Can add extensions later
        options.extensions = new Extensions();
        expect(options.extensions).toBeInstanceOf(Extensions);
    });

    it('should be compatible with different extension configurations', () => {
        const extensions1 = new Extensions();
        extensions1.addBytes(1, new Uint8Array([1, 2, 3]));

        const extensions2 = new Extensions();
        extensions2.addString(2, 'test');
        extensions2.addNumber(3, 42n);

        const options1: MOQOptions = { extensions: extensions1 };
        const options2: MOQOptions = { extensions: extensions2 };

        expect(options1.extensions?.getBytes(1)).toEqual(new Uint8Array([1, 2, 3]));
        expect(options2.extensions?.getString(2)).toBe('test');
        expect(options2.extensions?.getNumber(3)).toBe(42n);
    });
});
