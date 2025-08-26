import { describe, it, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { BytesBuffer } from './bytes';

describe('BytesBuffer', () => {
    it('should write and read data', () => {
        const buffer = new BytesBuffer(new ArrayBuffer(1024));
        const data = new Uint8Array([1, 2, 3]);
        buffer.write(data);
        expect(buffer.size).toBe(3);
        const readBuf = new Uint8Array(3);
        const bytesRead = buffer.read(readBuf);
        expect(bytesRead).toBe(3);
        expect(readBuf).toEqual(data);
        expect(buffer.size).toBe(0);
    });

    it('should grow when capacity is exceeded', () => {
        const buffer = new BytesBuffer(new ArrayBuffer(2));
        const data = new Uint8Array([1, 2]);
        buffer.write(data);
        expect(buffer.capacity).toBeGreaterThanOrEqual(2);
        const moreData = new Uint8Array([3, 4, 5]);
        buffer.write(moreData);
        expect(buffer.capacity).toBeGreaterThanOrEqual(5);
    });

    describe('readUint8', () => {
        it('should read a single byte', () => {
            const buffer = new BytesBuffer(new ArrayBuffer(10));
            const data = new Uint8Array([1, 2, 3]);
            buffer.write(data);
            expect(buffer.size).toBe(3);
            const byte = buffer.readUint8();
            expect(byte).toBe(1);
            expect(buffer.size).toBe(2);
        });
    });

    describe('writeUint8', () => {
        it('should write a single byte', () => {
            const buffer = new BytesBuffer(new ArrayBuffer(10));
            buffer.writeUint8(42);
            buffer.writeUint8(43);
            buffer.writeUint8(44);
            buffer.writeUint8(45);
            buffer.writeUint8(46);
            expect(buffer.size).toBe(5);
        });
    });

    describe('reserve', () => {
        it('should return a writable buffer with sufficient capacity', () => {
            const buffer = new BytesBuffer(new ArrayBuffer(10));
            const data1 = new Uint8Array([1, 2, 3]);
            buffer.write(data1);
            expect(buffer.size).toBe(3);
            
            const reservedBuffer = buffer.reserve(6);
            expect(reservedBuffer.length).toBeGreaterThanOrEqual(6);
            expect(buffer.capacity).toBeGreaterThanOrEqual(9);
        });
    });

    describe('construction with initial data', () => {
        it('should initialize with existing data', () => {
            const initialData = new Uint8Array([1, 2, 3]);
            const buffer = new BytesBuffer(initialData.buffer);
            // Initially buffer is empty for writing, data needs to be written first
            expect(buffer.size).toBe(0);
            expect(buffer.capacity).toBeGreaterThanOrEqual(3);
        });
    });

    describe('bytes method', () => {
        it('should return current buffer content without consuming it', () => {
            const buffer = new BytesBuffer(new ArrayBuffer(10));
            const data = new Uint8Array([1, 2, 3, 4, 5]);
            buffer.write(data);
            
            const currentBytes = buffer.bytes();
            expect(currentBytes).toEqual(data);
            // bytes() doesn't consume the data, it's still there
            expect(buffer.size).toBe(5);
        });
    });

    describe('write method with large data', () => {
        it('should handle writing data larger than initial capacity', () => {
            const buffer = new BytesBuffer(new ArrayBuffer(2));
            const largeData = new Uint8Array([1, 2, 3, 4]);
            buffer.write(largeData);
            expect(buffer.size).toBe(4);
            expect(buffer.capacity).toBeGreaterThanOrEqual(4);
            expect(buffer.bytes()).toEqual(largeData);
        });
    });

    describe('edge cases', () => {
        it('should handle empty writes and reads', () => {
            const buffer = new BytesBuffer(new ArrayBuffer(10));
            const emptyData = new Uint8Array(0);
            buffer.write(emptyData);
            expect(buffer.size).toBe(0);

            const readBuf = new Uint8Array(5);
            const bytesRead = buffer.read(readBuf);
            expect(bytesRead).toBe(0);
            expect(buffer.size).toBe(0);
        });

        it('should handle reset correctly', () => {
            const buffer = new BytesBuffer(new ArrayBuffer(10));
            const data = new Uint8Array([1, 2, 3]);
            buffer.write(data);
            expect(buffer.size).toBe(3);
            buffer.reset();
            expect(buffer.size).toBe(0);
        });
    });
});
