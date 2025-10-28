import { describe, it, beforeEach, afterEach, assertEquals, assertExists, assertThrows } from "../../deps.ts";
import { BufferPool } from './';

describe('BytesPool', () => {
  it('should acquire and release bytes', () => {
    const pool = new BufferPool({ min: 1, middle: 10, max: 100 });
    const bytes = pool.acquire(10);
    assertEquals(bytes.byteLength, 10);
    pool.release(bytes);
  });

  it('should reuse bytes from the pool', () => {
    const pool = new BufferPool({ min: 1, middle: 10, max: 100 });
    const bytes1 = pool.acquire(10);
    pool.release(bytes1);
    const bytes2 = pool.acquire(10);
    assertEquals(bytes2, bytes1);
  });

  it('should not reuse bytes if capacity is too small', () => {
    const pool = new BufferPool({ min: 1, middle: 10, max: 100 });
    const bytes1 = pool.acquire(10);
    pool.release(bytes1);
    const bytes2 = pool.acquire(20);
    assertNotEquals(bytes2, bytes1);
  });

  it('should clean up old bytes', () => {
    return new Promise<void>((resolve) => {
      const pool = new BufferPool({ min: 1, middle: 10, max: 100, options: { maxPerBucket: 10, maxTotalBytes: 10 } });
      const bytes1 = pool.acquire(10);
      pool.release(bytes1);
      setTimeout(() => {
        pool.cleanup();
        const bytes2 = pool.acquire(10);
        assertNotEquals(bytes2, bytes1);
        resolve();
      }, 20);
    });
  });

  it('should create new buffer when capacity exceeds max', () => {
    const pool = new BufferPool({ min: 1, middle: 10, max: 100 });
    const bytes = pool.acquire(200);
    assertEquals(bytes.byteLength, 200);
    // Cannot release since size doesn't match
  });

  it('should handle empty bucket', () => {
    const pool = new BufferPool({ min: 1, middle: 10, max: 100 });
    // Acquire all from bucket, but since maxPerBucket=5, and we acquire more than that
    for (let i = 0; i < 6; i++) {
      const bytes = pool.acquire(10);
      if (i < 5) pool.release(bytes); // First 5 go to bucket
    }
    // Now bucket has 5, acquire again should get from bucket
    const bytes = pool.acquire(10);
    assertEquals(bytes.byteLength, 10);
  });

  it('should not release when maxTotalBytes exceeded', () => {
    const pool = new BufferPool({ min: 1, middle: 10, max: 100, options: { maxTotalBytes: 10 } });
    const bytes1 = pool.acquire(10);
    pool.release(bytes1); // currentBytes = 10
    const bytes2 = pool.acquire(10);
    pool.release(bytes2); // currentBytes = 10
    const bytes3 = new ArrayBuffer(10); // external buffer
    pool.release(bytes3); // 10 + 10 = 20 > 10, not released
    const bytes4 = pool.acquire(10);
    assertEquals(bytes4, bytes2); // bucket has bytes2
    assertNotEquals(bytes4, bytes3); // bytes3 not in pool
  });

  it('should not release buffers with non-matching sizes', () => {
    const pool = new BufferPool({ min: 1, middle: 10, max: 100 });
    const bytes = new ArrayBuffer(50); // Not matching any size
    pool.release(bytes); // Should not be added to any bucket
    const acquired = pool.acquire(10);
    assertEquals(acquired.byteLength, 10);
    // Since no buffers in bucket, it's new
  });

  it('should cleanup buckets', () => {
    const pool = new BufferPool({ min: 1, middle: 10, max: 100 });
    const bytes = pool.acquire(10);
    pool.release(bytes);
    pool.cleanup();
    const bytes2 = pool.acquire(10);
    assertNotEquals(bytes2, bytes);
  });
});
