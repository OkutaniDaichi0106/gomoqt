import * as index from './index';

describe('Internal Index Module', () => {
    describe('re-exports', () => {
        it('should re-export bytes module', () => {
            // Check that bytes exports are available
            expect(index).toHaveProperty('BytesBuffer');
        });

        it('should re-export bytes_pool module', () => {
            // Check that bytes_pool exports are available
            expect(index).toHaveProperty('BytesPool');
        });

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
            const expectedExports = ['BytesBuffer', 'BytesPool', 'Mutex', 'background', 'withCancel'];
            
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
            expect(index.BytesBuffer).toBeDefined();
            expect(index.BytesPool).toBeDefined();
            expect(index.Cond).toBeDefined();
            // expect(index.Context).toBeDefined();
            expect(index.Extensions).toBeDefined();
            expect(index.Mutex).toBeDefined();
            // expect(index.Version).toBeDefined();
        });
    });

    describe('export functionality', () => {
        it('should provide working BytesBuffer', () => {
            const buffer = index.BytesBuffer.make(1024);
            expect(buffer).toBeInstanceOf(index.BytesBuffer);
            
            // Test basic functionality
            const data = new Uint8Array([1, 2, 3]);
            buffer.write(data);
            expect(buffer.size).toBe(3);
        });

        it('should provide working BytesPool', () => {
            const pool = new index.BytesPool();
            expect(pool).toBeInstanceOf(index.BytesPool);
            
            // Test basic functionality
            const buffer = pool.acquire(10);
            expect(buffer).toBeInstanceOf(ArrayBuffer);
        });

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
                    BytesBuffer: index.BytesBuffer,
                    BytesPool: index.BytesPool,
                    Mutex: index.Mutex,
                    Cond: index.Cond,
                    Extensions: index.Extensions,
                };
                return modules;
            }).not.toThrow();
        });

        it('should maintain proper module boundaries', () => {
            // Each exported class/function should be properly namespaced
            expect(index.BytesBuffer.name).toBe('BytesBuffer');
            expect(index.BytesPool.name).toBe('BytesPool');
            expect(index.Mutex.name).toBe('Mutex');
        });
    });

    describe('TypeScript compatibility', () => {
        it('should support TypeScript imports', () => {
            // Verify that TypeScript destructuring works
            expect(() => {
                const { BytesBuffer, BytesPool, Cond, Mutex, Extensions } = index;
                return { BytesBuffer, BytesPool, Cond, Mutex, Extensions };
            }).not.toThrow();
        });

        it('should provide proper type information', () => {
            // Test that types are preserved through re-exports
            const buffer = index.BytesBuffer.make(1024);
            const pool = new index.BytesPool();
            const mutex = new index.Mutex();
            
            // These should have the correct types (implicit type checking)
            expect(typeof buffer.write).toBe('function');
            expect(typeof pool.acquire).toBe('function');
            expect(typeof mutex.lock).toBe('function');
        });
    });

    describe('API consistency', () => {
        it('should provide consistent API access', () => {
            // Compare direct import vs index import
            const BytesBufferDirect = require('./bytes').BytesBuffer;
            const BytesPoolDirect = require('./bytes_pool').BytesPool;
            const MutexDirect = require('./mutex').Mutex;
            
            expect(index.BytesBuffer).toBe(BytesBufferDirect);
            expect(index.BytesPool).toBe(BytesPoolDirect);
            expect(index.Mutex).toBe(MutexDirect);
        });

        it('should not modify re-exported APIs', () => {
            // Ensure that re-exports maintain original functionality
            const directBuffer = require('./bytes').BytesBuffer.make(1024);
            const indexBuffer = index.BytesBuffer.make(1024);
            
            // Both should have identical API
            expect(Object.getOwnPropertyNames(Object.getPrototypeOf(directBuffer)))
                .toEqual(Object.getOwnPropertyNames(Object.getPrototypeOf(indexBuffer)));
        });
    });
});
