import { describe, it, expect } from '@jest/globals';
import { Frame } from './frame';

describe('Frame', () => {
  it('reports byteLength correctly', () => {
    const data = new Uint8Array([1, 2, 3]);
    const f = new Frame(data);
    expect(f.byteLength).toBe(3);
  });

  it('copyTo copies into Uint8Array', () => {
    const data = new Uint8Array([10, 20, 30]);
    const f = new Frame(data);
    const dest = new Uint8Array(3);
    f.copyTo(dest);
    expect(dest).toEqual(data);
  });

  it('copyTo copies into ArrayBuffer', () => {
    const data = new Uint8Array([7, 8, 9]);
    const f = new Frame(data);
    const destBuf = new ArrayBuffer(3);
    f.copyTo(destBuf);
    expect(new Uint8Array(destBuf)).toEqual(data);
  });

  it('copyTo throws on unsupported dest type', () => {
    const data = new Uint8Array([1]);
    const f = new Frame(data);
    // @ts-ignore - intentionally passing unsupported type
    expect(() => f.copyTo(123)).toThrow('Unsupported destination type');
  });

  it('copyFrom copies from another Source', () => {
    const srcData = new Uint8Array([5,6,7]);
    const src = new Frame(srcData);
    const dest = new Frame(new Uint8Array(3));
    dest.copyFrom(src);
    expect(dest.bytes).toEqual(srcData);
  });
});
