
import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { BufferPool } from './';

describe('BytesPool', () => {
  it('should acquire and release bytes', () => {
    const pool = new BufferPool({ min: 1, middle: 10, max: 100 });
    const bytes = pool.acquire(10);
    expect(bytes.byteLength).toBe(10);
    pool.release(bytes);
  });

  it('should reuse bytes from the pool', () => {
    const pool = new BufferPool({ min: 1, middle: 10, max: 100 });
    const bytes1 = pool.acquire(10);
    pool.release(bytes1);
    const bytes2 = pool.acquire(10);
    expect(bytes2).toBe(bytes1);
  });

  it('should not reuse bytes if capacity is too small', () => {
    const pool = new BufferPool({ min: 1, middle: 10, max: 100 });
    const bytes1 = pool.acquire(10);
    pool.release(bytes1);
    const bytes2 = pool.acquire(20);
    expect(bytes2).not.toBe(bytes1);
  });

  it('should clean up old bytes', (done) => {
    const pool = new BufferPool({ min: 1, middle: 10, max: 100, options: { maxPerBucket: 10, maxTotalBytes: 10 } });
    const bytes1 = pool.acquire(10);
    pool.release(bytes1);
    setTimeout(() => {
      pool.cleanup();
      const bytes2 = pool.acquire(10);
      expect(bytes2).not.toBe(bytes1);
      done();
    }, 20);
  });

  it('should acquire and release bytes', () => {
    const pool = new BufferPool({ min: 1, middle: 10, max: 100 });
    const bytes = pool.acquire(10);
    expect(bytes.byteLength).toBe(10);
    pool.release(bytes);
  });
});
