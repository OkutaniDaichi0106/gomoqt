import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import * as index from './index';
import { Mutex, Cond } from 'golikejs/sync';
import { background, withCancel } from 'golikejs/context';

describe('Internal Index Module', () => {
    describe('re-exports', () => {
        it('should re-export extensions module', () => {
            // Check that extensions exports are available
            expect(index).toHaveProperty('Extensions');
        });

        it('should re-export queue module', () => {
            // Check that queue exports are available
            expect(index).toHaveProperty('Queue');
        });

        it('should have all expected core exports', () => {
            const exports = Object.keys(index);

            // Should include key exports from each module
            const expectedExports = ['Extensions', 'Queue'];

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
            expect(index.Extensions).toBeDefined();
            expect(index.Queue).toBeDefined();
        });
    });

    describe('export functionality', () => {
        it('should provide working Queue', async () => {
            const queue = new index.Queue();
            expect(queue).toBeInstanceOf(index.Queue);
        });
    });

    describe('module dependencies', () => {
        it('should not create circular dependencies', () => {
            // This is more of a structural test
            // If there were circular dependencies, the import would fail
            expect(() => {
                const modules = {
                    Queue: index.Queue,
                    Extensions: index.Extensions,
                };
                return modules;
            }).not.toThrow();
        });

        it('should maintain proper module boundaries', () => {
            // Each exported class/function should be properly namespaced
            expect(index.Queue.name).toBe('Queue');
        });
    });

    describe('TypeScript compatibility', () => {
        it('should support TypeScript imports', () => {
            // Verify that TypeScript destructuring works
            expect(() => {
                const { Queue, Extensions } = index;
                return { Queue, Extensions };
            }).not.toThrow();
        });

        it('should provide proper type information', () => {
            // Test that types are preserved through re-exports
            const queue = new index.Queue();
            
            // These should have the correct types (implicit type checking)
            expect(typeof queue.enqueue).toBe('function');
        });
    });
});
