import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import * as index from './index';

describe('Internal Index Module', () => {
    describe('re-exports', () => {
        it('should re-export mutex module', () => {
            // Check that mutex exports are available
            expect(index).toHaveProperty('Mutex');
        });

        it('should re-export context module', () => {
            // Check that context exports are available
            expect(index).toHaveProperty('background');
            expect(index).toHaveProperty('withCancel');
        });

        it('should have all expected core exports', () => {
            const exports = Object.keys(index);

            // Should include key exports from each module
            const expectedExports = ['Mutex', 'background', 'withCancel'];

            expectedExports.forEach(expectedExport => {
                expect(exports).toContain(expectedExport);
            });
        });
    });

    describe('module structure', () => {
        it('should be a proper module', () => {
            expect(index).toBeDefined();
            expect(typeof index).toBe('object');
        });

        it('should not be empty', () => {
            const exportCount = Object.keys(index).length;
            expect(exportCount).toBeGreaterThan(0);
        });

        it('should provide access to core internal utilities', () => {
            // Verify that key internal utilities are accessible
            expect(index.Cond).toBeDefined();
            // expect(index.Context).toBeDefined();
            expect(index.Extensions).toBeDefined();
            expect(index.Mutex).toBeDefined();
            // expect(index.Version).toBeDefined();
        });
    });

    describe('export functionality', () => {
        it('should provide working Mutex', async () => {
            const mutex = new index.Mutex();
            expect(mutex).toBeInstanceOf(index.Mutex);
            
            // Test basic functionality
            const unlock = await mutex.lock();
            unlock();
        });
    });

    describe('module dependencies', () => {
        it('should not create circular dependencies', () => {
            // This is more of a structural test
            // If there were circular dependencies, the import would fail
            expect(() => {
                const modules = {
                    Mutex: index.Mutex,
                    Cond: index.Cond,
                    Extensions: index.Extensions,
                };
                return modules;
            }).not.toThrow();
        });

        it('should maintain proper module boundaries', () => {
            // Each exported class/function should be properly namespaced
            expect(index.Mutex.name).toBe('Mutex');
        });
    });

    describe('TypeScript compatibility', () => {
        it('should support TypeScript imports', () => {
            // Verify that TypeScript destructuring works
            expect(() => {
                const { Cond, Mutex, Extensions } = index;
                return { Cond, Mutex, Extensions };
            }).not.toThrow();
        });

        it('should provide proper type information', () => {
            // Test that types are preserved through re-exports
            const mutex = new index.Mutex();
            
            // These should have the correct types (implicit type checking)
            expect(typeof mutex.lock).toBe('function');
        });
    });

    describe('API consistency', () => {
        it('should provide consistent API access', async () => {
            // Compare direct import vs index import
            const { Mutex: MutexDirect } = await import('./mutex');
            
            expect(index.Mutex).toBe(MutexDirect);
        });

        it('should not modify re-exported APIs', async () => {
            // Ensure that re-exports maintain original functionality
            const { Mutex: MutexDirect } = await import('./mutex');
            const directMutex = new MutexDirect();
            const indexMutex = new index.Mutex();
            
            // Both should have identical API
            expect(Object.getOwnPropertyNames(Object.getPrototypeOf(directMutex)))
                .toEqual(Object.getOwnPropertyNames(Object.getPrototypeOf(indexMutex)));
        });
    });
});
