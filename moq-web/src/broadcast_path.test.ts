import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { BroadcastPath, isValidBroadcastPath, validateBroadcastPath, createBroadcastPath, broadcastPath, getExtension } from './broadcast_path';

describe('BroadcastPath', () => {
    describe('isValidBroadcastPath', () => {
        it('should return true for valid paths', () => {
            expect(isValidBroadcastPath('/')).toBe(true);
            expect(isValidBroadcastPath('/test')).toBe(true);
            expect(isValidBroadcastPath('/test/path')).toBe(true);
            expect(isValidBroadcastPath('/alice.json')).toBe(true);
            expect(isValidBroadcastPath('/video/stream')).toBe(true);
            expect(isValidBroadcastPath('/path/with/multiple/segments')).toBe(true);
        });

        it('should return false for invalid paths', () => {
            expect(isValidBroadcastPath('')).toBe(false);
            expect(isValidBroadcastPath('test')).toBe(false);
            expect(isValidBroadcastPath('test/path')).toBe(false);
            expect(isValidBroadcastPath('alice.json')).toBe(false);
        });
    });

    describe('validateBroadcastPath', () => {
        it('should return path for valid paths', () => {
            expect(validateBroadcastPath('/')).toBe('/');
            expect(validateBroadcastPath('/test')).toBe('/test');
            expect(validateBroadcastPath('/test/path')).toBe('/test/path');
            expect(validateBroadcastPath('/alice.json')).toBe('/alice.json');
        });

        it('should throw error for invalid paths', () => {
            expect(() => validateBroadcastPath('')).toThrow('Invalid broadcast path: "". Must start with "/"');
            expect(() => validateBroadcastPath('test')).toThrow('Invalid broadcast path: "test". Must start with "/"');
            expect(() => validateBroadcastPath('test/path')).toThrow('Invalid broadcast path: "test/path". Must start with "/"');
            expect(() => validateBroadcastPath('alice.json')).toThrow('Invalid broadcast path: "alice.json". Must start with "/"');
        });
    });

    describe('createBroadcastPath', () => {
        it('should create valid BroadcastPath for valid strings', () => {
            expect(createBroadcastPath('/')).toBe('/');
            expect(createBroadcastPath('/test')).toBe('/test');
            expect(createBroadcastPath('/test/path')).toBe('/test/path');
            expect(createBroadcastPath('/alice.json')).toBe('/alice.json');
        });

        it('should throw error for invalid strings', () => {
            expect(() => createBroadcastPath('')).toThrow('Invalid broadcast path: "". Must start with "/"');
            expect(() => createBroadcastPath('test')).toThrow('Invalid broadcast path: "test". Must start with "/"');
            expect(() => createBroadcastPath('test/path')).toThrow('Invalid broadcast path: "test/path". Must start with "/"');
        });
    });

    describe('broadcastPath', () => {
        it('should create BroadcastPath from valid template literals', () => {
            // These should compile and work correctly
            expect(broadcastPath('/')).toBe('/');
            expect(broadcastPath('/test')).toBe('/test');
            expect(broadcastPath('/test/path')).toBe('/test/path');
            expect(broadcastPath('/alice.json')).toBe('/alice.json');
        });

        // Note: Compile-time validation cannot be tested at runtime,
        // but TypeScript will catch invalid templates at compile time
    });

    describe('getExtension', () => {
        it('should return correct extension for paths with extensions', () => {
            expect(getExtension('/alice.json' as BroadcastPath)).toBe('.json');
            expect(getExtension('/video/stream.mp4' as BroadcastPath)).toBe('.mp4');
            expect(getExtension('/file.min.js' as BroadcastPath)).toBe('.js');
            expect(getExtension('/test/path.backup.mp4' as BroadcastPath)).toBe('.mp4');
            expect(getExtension('/test/.hidden.txt' as BroadcastPath)).toBe('.txt');
            expect(getExtension('/test/path.' as BroadcastPath)).toBe('.');
            expect(getExtension('file.txt' as BroadcastPath)).toBe('.txt');
        });

        it('should return empty string for paths without extensions', () => {
            expect(getExtension('/test/path' as BroadcastPath)).toBe('');
            expect(getExtension('/video/stream' as BroadcastPath)).toBe('');
            expect(getExtension('/' as BroadcastPath)).toBe('');
            expect(getExtension('' as BroadcastPath)).toBe('');
        });

        it('should handle edge cases correctly', () => {
            // Extension in directory name but not file
            expect(getExtension('/test.dir/file' as BroadcastPath)).toBe('');
            // Multiple dots in directory and file
            expect(getExtension('/test.dir/file.ext' as BroadcastPath)).toBe('.ext');
            // Hidden files
            expect(getExtension('/.hidden' as BroadcastPath)).toBe('');
        });
    });

    describe('type safety', () => {
        it('should allow BroadcastPath to be used as string', () => {
            const path: BroadcastPath = validateBroadcastPath('/test/path');
            expect(typeof path).toBe('string');
            expect(path.startsWith('/')).toBe(true);
            expect(path.length).toBeGreaterThan(0);
        });

        it('should allow assignment from validated paths', () => {
            const validatedPath = validateBroadcastPath('/test/path');
            const path: BroadcastPath = validatedPath;
            expect(path).toBe('/test/path');
        });
    });
});
