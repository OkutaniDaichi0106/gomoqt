import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows } from "../deps.ts";
import type { BroadcastPath} from './broadcast_path.ts';
import { isValidBroadcastPath, validateBroadcastPath, extension } from './broadcast_path.ts';

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


    describe('extension', () => {
        it('should return correct extension for paths with extensions', () => {
            expect(extension('/alice.json' as BroadcastPath)).toBe('.json');
            expect(extension('/video/stream.mp4' as BroadcastPath)).toBe('.mp4');
            expect(extension('/file.min.js' as BroadcastPath)).toBe('.js');
            expect(extension('/test/path.backup.mp4' as BroadcastPath)).toBe('.mp4');
            expect(extension('/test/.hidden.txt' as BroadcastPath)).toBe('.txt');
            expect(extension('/test/path.' as BroadcastPath)).toBe('.');
            expect(extension('file.txt' as BroadcastPath)).toBe('.txt');
        });

        it('should return empty string for paths without extensions', () => {
            expect(extension('/test/path' as BroadcastPath)).toBe('');
            expect(extension('/video/stream' as BroadcastPath)).toBe('');
            expect(extension('/' as BroadcastPath)).toBe('');
            expect(extension('' as BroadcastPath)).toBe('');
        });

        it('should handle edge cases correctly', () => {
            // Extension in directory name but not file
            expect(extension('/test.dir/file' as BroadcastPath)).toBe('');
            // Multiple dots in directory and file
            expect(extension('/test.dir/file.ext' as BroadcastPath)).toBe('.ext');
            // Hidden files
            expect(extension('/.hidden' as BroadcastPath)).toBe('');
        });
    });

    describe('type safety', () => {
        it('should allow BroadcastPath to be used as string', () => {
            const path: BroadcastPath = validateBroadcastPath('/test/path');
            assertEquals(typeof path, 'string');
            expect(path.startsWith('/')).toBe(true);
            expect(path.length).toBeGreaterThan(0);
        });

        it('should allow assignment from validated paths', () => {
            const validatedPath = validateBroadcastPath('/test/path');
            const path: BroadcastPath = validatedPath;
            assertEquals(path, '/test/path');
        });
    });
});


describe('broadcast path utilities', () => {
  it('isValidBroadcastPath works', () => {
    expect(isValidBroadcastPath('/')).toBe(true);
    expect(isValidBroadcastPath('/a')).toBe(true);
    expect(isValidBroadcastPath('a')).toBe(false);
  });

  it('validateBroadcastPath throws on invalid', () => {
    expect(() => validateBroadcastPath('no-slash')).toThrow();
    expect(validateBroadcastPath('/ok')).toBe('/ok');
  });

  it('extension extraction', () => {
    expect(extension('/alice.hang')).toBe('.hang');
    expect(extension('/path/to/file')).toBe('');
    expect(extension('/.hidden')).toBe('');
    expect(extension('/dir.name/file.txt')).toBe('.txt');
  });
});
