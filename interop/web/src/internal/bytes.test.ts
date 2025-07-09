import { BytesBuffer } from './bytes';
import { BytesPool } from './bytes_pool';

describe('BytesBuffer', () => {
    it('should write and read data', () => {
        const buffer = new BytesBuffer();
        const data = new Uint8Array([1, 2, 3]);
        buffer.write(data);
        expect(buffer.size()).toBe(3);
        const readBuf = new Uint8Array(3);
        const bytesRead = buffer.read(readBuf);
        expect(bytesRead).toBe(3);
        expect(readBuf).toEqual(data);
        expect(buffer.size()).toBe(0);
    });

    it('should expand the buffer when needed', () => {
        const buffer = new BytesBuffer();
        const data1 = new Uint8Array([1, 2]);
        buffer.write(data1);
        expect(buffer.capacity()).toBeGreaterThanOrEqual(2);
        const data2 = new Uint8Array([3, 4, 5]);
        buffer.write(data2);
        expect(buffer.capacity()).toBeGreaterThanOrEqual(5);
        const readBuf = new Uint8Array(5);
        const bytesRead = buffer.read(readBuf);
        expect(bytesRead).toBe(5);
        expect(readBuf).toEqual(new Uint8Array([1, 2, 3, 4, 5]));
    });

    it('should reset the buffer', () => {
        const buffer = new BytesBuffer();
        const data = new Uint8Array([1, 2, 3]);
        buffer.write(data);
        buffer.reset();
        expect(buffer.size()).toBe(0);
        expect(buffer.capacity()).toBeGreaterThanOrEqual(3);
    });

    it('should read all available bytes when provided buffer is larger', () => {
        const buffer = new BytesBuffer();
        const data = new Uint8Array([1, 2, 3]);
        buffer.write(data);
        const readBuf = new Uint8Array(4);
        const bytesRead = buffer.read(readBuf);
        expect(bytesRead).toBe(3); // It should read all available bytes
        expect(readBuf.subarray(0, 3)).toEqual(data);
        expect(buffer.size()).toBe(0);
    });

    it('should handle multiple writes', () => {
        const buffer = new BytesBuffer();
        const data1 = new Uint8Array([1, 2]);
        const data2 = new Uint8Array([3, 4, 5]);
        buffer.write(data1);
        buffer.write(data2);
        expect(buffer.size()).toBe(5);
        const readBuf = new Uint8Array(5);
        const bytesRead = buffer.read(readBuf);
        expect(bytesRead).toBe(5);
        expect(readBuf).toEqual(new Uint8Array([1, 2, 3, 4, 5]));
    });

    it('should slide the buffer when reading', () => {
        const buffer = new BytesBuffer();
        const data1 = new Uint8Array([1, 2, 3, 4, 5]);
        buffer.write(data1);
        const readBuf1 = new Uint8Array(2);
        const bytesRead1 = buffer.read(readBuf1);
        expect(bytesRead1).toBe(2);
        expect(readBuf1).toEqual(new Uint8Array([1, 2]));
        expect(buffer.size()).toBe(3);
        const data2 = new Uint8Array([6, 7, 8, 9, 10, 11]);
        buffer.write(data2);
        expect(buffer.capacity()).toBeGreaterThanOrEqual(9);
        const readBuf2 = new Uint8Array(9);
        const bytesRead2 = buffer.read(readBuf2);
        expect(bytesRead2).toBe(9);
        expect(readBuf2).toEqual(new Uint8Array([3, 4, 5, 6, 7, 8, 9, 10, 11]));
    });

    it('should be initialized with a Uint8Array', () => {
        const initialData = new Uint8Array([1, 2, 3]);
        const buffer = new BytesBuffer(initialData);
        expect(buffer.size()).toBe(3);
        expect(buffer.capacity()).toBeGreaterThanOrEqual(3);
        const readBuf = new Uint8Array(3);
        const bytesRead = buffer.read(readBuf);
        expect(bytesRead).toBe(3);
        expect(readBuf).toEqual(initialData);
    });

    it('should return the unread portion as a new Uint8Array with toUint8Array()', () => {
        const buffer = new BytesBuffer();
        buffer.write(new Uint8Array([1, 2, 3, 4, 5]));
        const tempReadBuf = new Uint8Array(2);
        buffer.read(tempReadBuf);
        const unread = buffer.toUint8Array();
        expect(unread).toEqual(new Uint8Array([3, 4, 5]));
        // Ensure it's a copy, not a view
        unread[0] = 99;
        expect(buffer.toUint8Array()).toEqual(new Uint8Array([3, 4, 5]));
    });

    it('should release the buffer (default pool)', () => {
        const buffer = new BytesBuffer();
        const data = new Uint8Array([1, 2, 3]);
        buffer.write(data);
        expect(buffer.size()).toBe(3);
        expect(buffer.capacity()).toBeGreaterThan(0);
        buffer.release();
        expect(buffer.size()).toBe(0);
        expect(buffer.capacity()).toBe(0);
    });

    it('should allow writing after release', () => {
        const buffer = new BytesBuffer();
        const data1 = new Uint8Array([1, 2, 3]);
        buffer.write(data1);
        buffer.release();
        expect(buffer.size()).toBe(0);
        expect(buffer.capacity()).toBe(0);

        const data2 = new Uint8Array([4, 5, 6, 7]);
        buffer.write(data2);
        expect(buffer.size()).toBe(4);
        expect(buffer.capacity()).toBeGreaterThanOrEqual(4);
        expect(buffer.toUint8Array()).toEqual(data2);
    });

    it('should release the buffer to a provided pool', () => {
        class MockBytesPool extends BytesPool {
            release = jest.fn();
        }
        const mockPool = new MockBytesPool();

        const buffer = new BytesBuffer();
        const data = new Uint8Array([10, 20, 30]);
        buffer.write(data);
        // Get a reference to the internal buffer before release
        const internalBufferBeforeRelease = buffer.bytes(); 

        buffer.release(mockPool);

        expect(mockPool.release).toHaveBeenCalledWith(internalBufferBeforeRelease.buffer);
        expect(buffer.size()).toBe(0);
        expect(buffer.capacity()).toBe(0);
    });
});
